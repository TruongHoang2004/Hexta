"""Idempotent knowledge ingest and tenant-isolated retrieval for grounded answers."""

import logging
import os

from pydantic import BaseModel, Field, ValidationError

from hexta_ai.knowledge.models import KnowledgeInput, KnowledgeItem, document_chunks
from hexta_ai.knowledge.store import KnowledgeStore, Retry

log = logging.getLogger(__name__)


def estimate_tokens(text: str) -> int:
    """Same conservative UTF-8 byte estimate as the provider input budget; never billed usage."""
    return max(1, len(text.encode("utf-8")))


class RagConfig(BaseModel):
    top_k: int = Field(default=6, ge=1, le=20)
    min_similarity: float = Field(default=0.55, ge=-1, le=1)
    max_context_tokens: int = Field(default=6000, ge=200, le=32000)


def rag_config_from_env() -> RagConfig:
    try:
        return RagConfig(top_k=os.getenv("AI_RAG_TOP_K", "6"), min_similarity=os.getenv("AI_RAG_MIN_SIMILARITY", "0.55"),
                         max_context_tokens=os.getenv("AI_RAG_MAX_CONTEXT_TOKENS", "6000"))
    except ValidationError:
        raise RuntimeError("Invalid retrieval configuration") from None


def decide(row: dict | None, document: KnowledgeInput, model: str) -> str:
    """Pure ingest decision against the current source row (or tombstone)."""
    if row is None:
        return "created"
    stored = row["source_revision"]
    if document.source_revision < stored:
        return "stale"
    if row["deleted"]:
        return "created" if document.source_revision > stored else "stale"
    same_text = bytes(row["content_hash"]) == document.content_hash()
    same_attributes = (row["attributes"] or {}) == document.attributes
    same_model = row["embedding_model"] == model
    if document.source_revision == stored:
        if same_text and same_attributes:
            return "unchanged" if same_model else "updated"
        return "stale"  # Changed content must carry a newer revision.
    if not same_text or not same_model:
        return "updated"
    return "attributes_updated" if not same_attributes else "unchanged"


class KnowledgeService:
    def __init__(self, store: KnowledgeStore, embeddings, config: RagConfig | None = None):
        self.store = store
        self.embeddings = embeddings
        self.config = config or RagConfig()

    @property
    def simulated(self) -> bool:
        return self.embeddings.simulated

    async def ping(self):
        await self.store.ping()

    async def upsert(self, tenant: str, document: KnowledgeInput) -> tuple[str, int]:
        model = self.embeddings.model
        for _ in range(3):
            row = await self.store.source(tenant, document.source_type, document.source_id)
            action = decide(row, document, model)
            chunks = None
            if action in ("created", "updated"):
                texts = document_chunks(document)
                vectors = await self.embeddings.embed_documents(texts)
                chunks = [(text, estimate_tokens(text), vector) for text, vector in zip(texts, vectors)]
            try:
                outcome = await self.store.write(tenant, document, model=model, chunks=chunks,
                                                 decide=lambda current: decide(current, document, model))
            except Retry:
                continue
            count = len(chunks) if chunks is not None else (row or {}).get("chunk_count", 0)
            return outcome, count if outcome != "stale" else 0
        raise Retry()

    async def delete(self, tenant: str, source_type: str, source_id: str, revision: int) -> str:
        for _ in range(3):
            try:
                return await self.store.delete(tenant, source_type, source_id, revision)
            except Retry:
                continue
        raise Retry()

    async def retrieve(self, tenant: str, query: str, *, active_script_ids: set[str]) -> list[KnowledgeItem]:
        """Top-k tenant chunks above the similarity floor, merged per source, within the context budget."""
        vector = await self.embeddings.embed_query(query)
        rows = await self.store.search(tenant, vector, model=self.embeddings.model, limit=self.config.top_k * 2)
        items: dict[str, KnowledgeItem] = {}
        order: list[str] = []
        budget = self.config.max_context_tokens
        selected = 0
        for row in rows:
            if row["similarity"] < self.config.min_similarity:
                continue
            attributes = row["attributes"] or {}
            # Only scripts whose currently active version Go authorized for this run may ground an answer.
            if row["source_type"] == "script" and attributes.get("version_id") not in active_script_ids:
                continue
            if selected >= self.config.top_k or row["token_estimate"] > budget:
                continue
            budget -= row["token_estimate"]
            selected += 1
            key = f"{row['source_type']}:{row['source_id']}"
            if key in items:
                items[key].text += "\n" + row["content"]
            else:
                items[key] = KnowledgeItem(id=key, title=row["title"], text=row["content"],
                                           attributes=attributes, similarity=float(row["similarity"]))
                order.append(key)
        return [items[key] for key in order]

    async def close(self):
        await self.store.close()
        await self.embeddings.close()


def safe_embeddings_from_env(*, paid_allowed: bool = True):
    """Embeddings for the runtime, failing safe to the offline hashing model: while AI is disabled (no paid call) or
    when the real model is misconfigured (logged as an error). Chunks embedded with another model are ignored by
    retrieval until the knowledge re-sync command re-embeds them."""
    from hexta_ai.knowledge.embeddings import HashingEmbeddings, embeddings_from_env

    requested = os.getenv("AI_EMBEDDING_PROVIDER", "fake").strip().lower() or "fake"
    if requested != "fake" and not paid_allowed:
        log.warning("Knowledge embeddings use the offline hashing model: AI is disabled, so AI_EMBEDDING_PROVIDER "
                    "is ignored")
        return HashingEmbeddings()
    try:
        return embeddings_from_env()
    except RuntimeError as error:
        log.error("Knowledge embeddings use the offline hashing model: %s (misconfiguration)", error)
        return HashingEmbeddings()


def knowledge_from_env(*, paid_embeddings: bool = True) -> KnowledgeService | None:
    """Knowledge needs the AI database; without AI_DATABASE_URL retrieval/ingest are unavailable."""
    dsn = os.getenv("AI_DATABASE_URL", "").strip()
    if not dsn:
        return None
    return KnowledgeService(KnowledgeStore(dsn), safe_embeddings_from_env(paid_allowed=paid_embeddings),
                            rag_config_from_env())
