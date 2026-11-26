"""PostgreSQL/pgvector authority for tenant knowledge. Every statement is tenant-qualified.

Embedding happens before these short transactions; no lock is held across provider work. The
source row is the revision fence (including delete tombstones), locked with FOR UPDATE.
"""

from collections.abc import Callable

from hexta_ai.knowledge.models import KnowledgeInput
from hexta_ai.store import PostgresPool

SOURCE_COLUMNS = "source_revision, content_hash, attributes, embedding_model, deleted, chunk_count"


def vector_literal(vector: list[float]) -> str:
    return "[" + ",".join(f"{v:.8g}" for v in vector) + "]"


class Retry(Exception):
    """The source row changed between the unlocked read and the write; recompute and try again."""


class KnowledgeStore(PostgresPool):
    def __init__(self, dsn: str, *, min_size: int = 1, max_size: int = 4):
        super().__init__(dsn, min_size=min_size, max_size=max_size)

    async def ping(self):
        async with self.connect() as db:
            await db.execute("SELECT 1 FROM ai_runtime.knowledge_sources LIMIT 1")

    async def source(self, tenant: str, source_type: str, source_id: str) -> dict | None:
        async with self.connect() as db:
            return await (await db.execute(
                f"""SELECT {SOURCE_COLUMNS} FROM ai_runtime.knowledge_sources
                WHERE tenant_id=%s AND source_type=%s AND source_id=%s""",
                (tenant, source_type, source_id))).fetchone()

    async def write(self, tenant: str, document: KnowledgeInput, *, model: str,
                    decide: Callable[[dict | None], str], chunks: list[tuple[str, int, list[float]]] | None) -> str:
        """Apply the decision made against the locked row. Raises Retry when embeddings are missing or stale."""
        key = (tenant, document.source_type, document.source_id)
        async with self.connect() as db:
            row = await (await db.execute(
                f"""SELECT {SOURCE_COLUMNS} FROM ai_runtime.knowledge_sources
                WHERE tenant_id=%s AND source_type=%s AND source_id=%s FOR UPDATE""", key)).fetchone()
            action = decide(row)
            if action == "stale":
                return action
            if action == "unchanged":
                await db.execute("""UPDATE ai_runtime.knowledge_sources SET source_revision=%s, updated_at=clock_timestamp()
                    WHERE tenant_id=%s AND source_type=%s AND source_id=%s AND source_revision<%s""",
                                 (document.source_revision, *key, document.source_revision))
                return action
            if action == "attributes_updated":
                await db.execute("""UPDATE ai_runtime.knowledge_sources SET source_revision=%s, attributes=%s::jsonb,
                    updated_at=clock_timestamp() WHERE tenant_id=%s AND source_type=%s AND source_id=%s""",
                                 (document.source_revision, document.attributes_json(), *key))
                return action
            if chunks is None:
                raise Retry()
            values = (document.source_revision, document.content_hash(), document.title, document.attributes_json(),
                      model, len(chunks))
            if row is None:
                cursor = await db.execute("""INSERT INTO ai_runtime.knowledge_sources
                    (tenant_id, source_type, source_id, source_revision, content_hash, title, attributes,
                     embedding_model, chunk_count, deleted)
                    VALUES (%s,%s,%s,%s,%s,%s,%s::jsonb,%s,%s,false) ON CONFLICT DO NOTHING""", (*key, *values))
                if cursor.rowcount != 1:
                    raise Retry()
            else:
                await db.execute("""UPDATE ai_runtime.knowledge_sources SET source_revision=%s, content_hash=%s,
                    title=%s, attributes=%s::jsonb, embedding_model=%s, chunk_count=%s, deleted=false,
                    updated_at=clock_timestamp() WHERE tenant_id=%s AND source_type=%s AND source_id=%s""",
                                 (*values, *key))
                await db.execute("""DELETE FROM ai_runtime.knowledge_chunks
                    WHERE tenant_id=%s AND source_type=%s AND source_id=%s""", key)
            async with db.cursor() as cursor:
                await cursor.executemany("""INSERT INTO ai_runtime.knowledge_chunks
                    (tenant_id, source_type, source_id, chunk_index, content, token_estimate, embedding_model, embedding)
                    VALUES (%s,%s,%s,%s,%s,%s,%s,%s::public.vector)""",
                    [(*key, index, content, tokens, model, vector_literal(vector))
                     for index, (content, tokens, vector) in enumerate(chunks)])
            return action

    async def delete(self, tenant: str, source_type: str, source_id: str, revision: int) -> str:
        key = (tenant, source_type, source_id)
        async with self.connect() as db:
            row = await (await db.execute(
                f"""SELECT {SOURCE_COLUMNS} FROM ai_runtime.knowledge_sources
                WHERE tenant_id=%s AND source_type=%s AND source_id=%s FOR UPDATE""", key)).fetchone()
            if row is None:
                # A tombstone fences an older upsert that is still in flight.
                cursor = await db.execute("""INSERT INTO ai_runtime.knowledge_sources
                    (tenant_id, source_type, source_id, source_revision, content_hash, title, attributes,
                     embedding_model, chunk_count, deleted)
                    VALUES (%s,%s,%s,%s,''::bytea,'','{}'::jsonb,'',0,true) ON CONFLICT DO NOTHING""", (*key, revision))
                if cursor.rowcount != 1:
                    raise Retry()
                return "not_found"
            # A delete at the same revision as the stored content wins; an upsert at a tombstone's revision does not.
            if revision < row["source_revision"]:
                return "stale"
            await db.execute("""DELETE FROM ai_runtime.knowledge_chunks
                WHERE tenant_id=%s AND source_type=%s AND source_id=%s""", key)
            await db.execute("""UPDATE ai_runtime.knowledge_sources SET source_revision=GREATEST(source_revision,%s),
                deleted=true, chunk_count=0, attributes='{}'::jsonb, updated_at=clock_timestamp()
                WHERE tenant_id=%s AND source_type=%s AND source_id=%s""", (revision, *key))
            return "not_found" if row["deleted"] else "deleted"

    async def search(self, tenant: str, embedding: list[float], *, model: str, limit: int) -> list[dict]:
        query = vector_literal(embedding)
        async with self.connect() as db:
            # Filtered HNSW scans keep searching until enough tenant rows qualify (pgvector >= 0.8).
            await db.execute("SELECT set_config('hnsw.iterative_scan', 'strict_order', true)")
            cursor = await db.execute(
                """SELECT c.source_type, c.source_id, c.chunk_index, c.content, c.token_estimate,
                       s.title, s.attributes,
                       1 - (c.embedding OPERATOR(public.<=>) %(q)s::public.vector) AS similarity
                FROM ai_runtime.knowledge_chunks c
                JOIN ai_runtime.knowledge_sources s ON s.tenant_id=c.tenant_id
                     AND s.source_type=c.source_type AND s.source_id=c.source_id
                WHERE c.tenant_id=%(t)s AND s.tenant_id=%(t)s AND c.embedding_model=%(m)s AND NOT s.deleted
                ORDER BY c.embedding OPERATOR(public.<=>) %(q)s::public.vector
                LIMIT %(k)s""", {"q": query, "t": tenant, "m": model, "k": limit})
            return await cursor.fetchall()
