"""Tenant knowledge (RAG) for grounded answers: ingest, embeddings and pgvector retrieval."""

from hexta_ai.knowledge.embeddings import DIMENSIONS, EmbeddingError, GeminiEmbeddings, HashingEmbeddings
from hexta_ai.knowledge.models import KnowledgeInput, KnowledgeItem
from hexta_ai.knowledge.service import KnowledgeService, RagConfig, knowledge_from_env

__all__ = ["DIMENSIONS", "EmbeddingError", "GeminiEmbeddings", "HashingEmbeddings", "KnowledgeInput", "KnowledgeItem",
           "KnowledgeService", "RagConfig", "knowledge_from_env"]
