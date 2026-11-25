-- pgvector for the knowledge store (history: migration 20261126950000_enable_pgvector). Atlas Community rejects a
-- plain CREATE EXTENSION in desired SQL ("extensions are available to logged-in users only") and drops the
-- extension from its dev database when it cleans up after a diff, so the desired schema recreates it on every
-- run inside a DO block. The AI dev database must therefore be a disposable pgvector image (atlas.hcl default:
-- docker+postgres://_/pgvector/pgvector:0.8.1-pg17/dev) whose login may create extensions.
DO $$ BEGIN CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public; END $$;
CREATE SCHEMA IF NOT EXISTS ai_runtime;
CREATE TABLE ai_runtime.runs (
    tenant_id text NOT NULL,
    run_id text NOT NULL,
    request bytea NOT NULL,
    digest bytea NOT NULL,
    result bytea NOT NULL,
    status smallint NOT NULL CHECK (status BETWEEN 1 AND 5),
    attempts integer NOT NULL DEFAULT 0,
    lease_token uuid,
    lease_until timestamptz,
    error_code text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, run_id),
    CHECK ((status = 2) = (lease_token IS NOT NULL AND lease_until IS NOT NULL))
);
CREATE INDEX runs_queue ON ai_runtime.runs(status, lease_until, created_at);

-- Tenant knowledge for grounded answers (RAG). Requires the pgvector extension (>= 0.8.0) in schema public,
-- created by migration 20261126950000_enable_pgvector (and by the DO block above in the Atlas dev database).
-- One row per source; also the revision fence (deleted rows are tombstones) for out-of-order ingest.
CREATE TABLE ai_runtime.knowledge_sources (
    tenant_id text NOT NULL,
    source_type text NOT NULL,
    source_id text NOT NULL,
    source_revision bigint NOT NULL,
    content_hash bytea NOT NULL,
    title text NOT NULL DEFAULT '',
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    embedding_model text NOT NULL,
    chunk_count integer NOT NULL DEFAULT 0,
    deleted boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (tenant_id, source_type, source_id),
    CONSTRAINT knowledge_sources_type_check CHECK (source_type IN ('product', 'document', 'script')),
    CONSTRAINT knowledge_sources_revision_check CHECK (source_revision >= 0),
    CONSTRAINT knowledge_sources_attributes_check CHECK (jsonb_typeof(attributes) = 'object'),
    CONSTRAINT knowledge_sources_chunks_check CHECK (chunk_count >= 0 AND (NOT deleted OR chunk_count = 0))
);
CREATE TABLE ai_runtime.knowledge_chunks (
    tenant_id text NOT NULL,
    source_type text NOT NULL,
    source_id text NOT NULL,
    chunk_index integer NOT NULL,
    content text NOT NULL,
    token_estimate integer NOT NULL,
    embedding_model text NOT NULL,
    embedding public.vector(768) NOT NULL,
    PRIMARY KEY (tenant_id, source_type, source_id, chunk_index),
    CONSTRAINT knowledge_chunks_source_fkey FOREIGN KEY (tenant_id, source_type, source_id)
        REFERENCES ai_runtime.knowledge_sources (tenant_id, source_type, source_id) ON DELETE CASCADE,
    CONSTRAINT knowledge_chunks_index_check CHECK (chunk_index >= 0),
    CONSTRAINT knowledge_chunks_tokens_check CHECK (token_estimate > 0)
);
-- Cosine similarity search; queries always filter by tenant and embedding model (iterative scans).
CREATE INDEX knowledge_chunks_embedding ON ai_runtime.knowledge_chunks USING hnsw (embedding public.vector_cosine_ops);
