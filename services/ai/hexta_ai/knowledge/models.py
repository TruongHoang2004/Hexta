"""Validated knowledge ingest/retrieval values and pure chunking. No I/O."""

import hashlib
import json
import re
import unicodedata
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator

SOURCE_TYPES = ("product", "document", "script")
IDENTIFIER = r"^[A-Za-z0-9_-]{1,100}$"
CHUNK_CHARS = 1200
CHUNK_OVERLAP = 150
MAX_CHUNKS = 40


def _clean(value: str) -> str:
    value = unicodedata.normalize("NFC", value)
    return "".join(c for c in value if c in "\n\t" or unicodedata.category(c) not in ("Cc", "Cf", "Cs")).strip()


class KnowledgeInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    source_type: Literal["product", "document", "script"]
    source_id: str = Field(pattern=IDENTIFIER)
    source_revision: int = Field(ge=0)
    title: str = Field(max_length=300)
    body: str = Field(min_length=1, max_length=20000)
    attributes: dict[str, str] = Field(default_factory=dict, max_length=30)

    @field_validator("title", "body", mode="before")
    @classmethod
    def clean_text(cls, value):
        return _clean(value) if isinstance(value, str) else value

    @field_validator("attributes")
    @classmethod
    def clean_attributes(cls, values):
        cleaned = {}
        for key, value in values.items():
            value = _clean(value)
            if not re.fullmatch(r"[a-z][a-z0-9_]{0,63}", key) or len(value) > 500:
                raise ValueError("Invalid knowledge attribute")
            cleaned[key] = value
        return cleaned

    @property
    def key(self) -> str:
        return f"{self.source_type}:{self.source_id}"

    def embedded_text(self) -> str:
        return f"{self.title}\n\n{self.body}" if self.title else self.body

    def content_hash(self) -> bytes:
        return hashlib.sha256(self.embedded_text().encode("utf-8")).digest()

    def attributes_json(self) -> str:
        return json.dumps(self.attributes, ensure_ascii=False, sort_keys=True)


class KnowledgeItem(BaseModel):
    """One retrieved source (its matching chunks merged) as shown to the answer model."""

    id: str
    title: str
    text: str
    attributes: dict[str, str] = Field(default_factory=dict)
    similarity: float = 0


def chunk_text(text: str, *, size: int = CHUNK_CHARS, overlap: int = CHUNK_OVERLAP) -> list[str]:
    """Paragraph/sentence-aware character windows; deterministic so equal input yields equal chunks."""
    text = re.sub(r"[ \t]+", " ", text).strip()
    if len(text) <= size:
        return [text]
    pieces = [p for p in re.split(r"(?<=[.!?。\n])\s+", text) if p]
    chunks, current = [], ""
    for piece in pieces:
        while len(piece) > size:
            head, piece = piece[:size], piece[size - overlap:]
            if current:
                chunks.append(current)
                current = ""
            chunks.append(head)
        if len(current) + len(piece) + 1 > size:
            chunks.append(current)
            current = current[-overlap:] + " " + piece if overlap else piece
        else:
            current = f"{current} {piece}".strip()
    if current:
        chunks.append(current)
    return [c.strip() for c in chunks if c.strip()]


def document_chunks(document: KnowledgeInput) -> list[str]:
    """Chunks to embed; each keeps the title so a chunk is meaningful on its own."""
    chunks = chunk_text(document.body)
    if len(chunks) > MAX_CHUNKS:
        raise ValueError("Knowledge document is too long")
    return [f"{document.title}\n{c}" if document.title else c for c in chunks]
