"""Embedding port: Gemini through LangChain, plus a deterministic offline hashing model."""

import asyncio
import hashlib
import math
import os
import re
import unicodedata

from pydantic import BaseModel, Field, SecretStr, ValidationError

# Fixed by the ai_runtime.knowledge_chunks vector column; changing it requires a migration.
DIMENSIONS = 768


class EmbeddingError(Exception):
    def __init__(self, code: str = "EMBEDDING_ERROR"):
        self.code = code if code in {"EMBEDDING_ERROR", "EMBEDDING_TIMEOUT", "EMBEDDING_QUOTA"} else "EMBEDDING_ERROR"
        super().__init__(self.code)


def normalize(vector: list[float]) -> list[float]:
    if len(vector) != DIMENSIONS or not all(isinstance(v, (int, float)) and math.isfinite(v) for v in vector):
        raise EmbeddingError("EMBEDDING_ERROR")
    norm = math.sqrt(sum(v * v for v in vector))
    if norm == 0:
        raise EmbeddingError("EMBEDDING_ERROR")
    return [v / norm for v in vector]


def fold(text: str) -> str:
    """Lowercase Vietnamese text without diacritics ("Đổi trả" -> "doi tra"), for offline matching."""
    decomposed = unicodedata.normalize("NFD", text.casefold().replace("đ", "d"))
    return "".join(c for c in decomposed if unicodedata.category(c) != "Mn")


# Question and filler words that carry no topic; dropping them keeps short questions close to their record.
STOPWORDS = frozenset("""a ah a oi ad admin ban bao nhieu bn cho co con cua da duoc gi giup hay hoi khong ko k la
lam may minh mot nao nay nhe nhi oi roi shop thi toi tu va voi vay ve vs khi the nhu nao""".split())


class HashingEmbeddings:
    """Offline, deterministic hashing of diacritic-folded topic words and bigrams. Similar wording yields similar
    vectors with or without Vietnamese diacritics; no network. Bump `model` whenever the vector function changes so
    stored chunks are re-embedded."""

    name = "fake"
    model = "hashing-768-v2"
    simulated = True

    @staticmethod
    def _vector(text: str) -> list[float]:
        words = [w for w in re.findall(r"\w+", fold(unicodedata.normalize("NFC", text))) if w not in STOPWORDS]
        features = words + [f"{a} {b}" for a, b in zip(words, words[1:])]
        vector = [0.0] * DIMENSIONS
        for feature in features or ["<empty>"]:
            digest = hashlib.blake2b(feature.encode(), digest_size=8).digest()
            index = int.from_bytes(digest[:4], "big") % DIMENSIONS
            vector[index] += 1.0 if digest[4] & 1 else -1.0
        if not any(vector):
            vector[0] = 1.0
        return normalize(vector)

    async def embed_documents(self, texts: list[str]) -> list[list[float]]:
        return [self._vector(t) for t in texts]

    async def embed_query(self, text: str) -> list[float]:
        return self._vector(text)

    async def close(self):
        pass


class EmbeddingConfig(BaseModel):
    provider: str = Field(default="fake", pattern=r"^(fake|gemini)$")
    model: str = Field(default="gemini-embedding-001", pattern=r"^[A-Za-z0-9._/-]{1,100}$")
    api_key: SecretStr = Field(default=SecretStr(""), repr=False)
    timeout: float = Field(default=10, ge=1, le=30)


class GeminiEmbeddings:
    """Gemini embedding model via langchain-google-genai (RETRIEVAL_DOCUMENT / RETRIEVAL_QUERY tasks)."""

    name = "gemini"
    simulated = False

    def __init__(self, config: EmbeddingConfig, client=None):
        self.config = config
        self.model = config.model
        if client is None:
            from langchain_google_genai import GoogleGenerativeAIEmbeddings

            client = GoogleGenerativeAIEmbeddings(model=config.model, google_api_key=config.api_key,
                                                  output_dimensionality=DIMENSIONS, vertexai=False)
        self.client = client

    async def _call(self, coroutine):
        try:
            async with asyncio.timeout(self.config.timeout):
                return await coroutine
        except TimeoutError:
            raise EmbeddingError("EMBEDDING_TIMEOUT") from None
        except EmbeddingError:
            raise
        except Exception as error:
            cause = error.__cause__ or error
            status = getattr(cause, "code", None) or getattr(cause, "status_code", None)
            raise EmbeddingError("EMBEDDING_QUOTA" if status in (429, "RESOURCE_EXHAUSTED") else "EMBEDDING_ERROR") from None

    async def embed_documents(self, texts: list[str]) -> list[list[float]]:
        vectors = await self._call(self.client.aembed_documents(texts, task_type="RETRIEVAL_DOCUMENT",
                                                                output_dimensionality=DIMENSIONS))
        if len(vectors) != len(texts):
            raise EmbeddingError("EMBEDDING_ERROR")
        return [normalize(list(v)) for v in vectors]

    async def embed_query(self, text: str) -> list[float]:
        vector = await self._call(self.client.aembed_query(text, task_type="RETRIEVAL_QUERY",
                                                           output_dimensionality=DIMENSIONS))
        return normalize(list(vector))

    async def close(self):
        client = getattr(self.client, "client", None)
        aio = getattr(client, "aio", None)
        closer = getattr(aio, "aclose", None)
        if closer:
            await closer()


def embeddings_from_env():
    try:
        config = EmbeddingConfig(
            provider=os.getenv("AI_EMBEDDING_PROVIDER", "fake").strip().lower() or "fake",
            model=os.getenv("AI_EMBEDDING_MODEL", "").strip() or "gemini-embedding-001",
            api_key=os.getenv("AI_EMBEDDING_API_KEY") or os.getenv("GEMINI_API_KEY", ""),
            timeout=os.getenv("AI_EMBEDDING_TIMEOUT", "10"),
        )
    except ValidationError:
        raise RuntimeError("Invalid embedding configuration") from None
    if config.provider == "fake":
        return HashingEmbeddings()
    if not config.api_key.get_secret_value().strip():
        raise RuntimeError("Gemini embeddings require GEMINI_API_KEY or AI_EMBEDDING_API_KEY")
    return GeminiEmbeddings(config)
