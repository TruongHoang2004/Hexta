-- Create "knowledge_sources" table
CREATE TABLE "ai_runtime"."knowledge_sources" (
  "tenant_id" text NOT NULL,
  "source_type" text NOT NULL,
  "source_id" text NOT NULL,
  "source_revision" bigint NOT NULL,
  "content_hash" bytea NOT NULL,
  "title" text NOT NULL DEFAULT '',
  "attributes" jsonb NOT NULL DEFAULT '{}',
  "embedding_model" text NOT NULL,
  "chunk_count" integer NOT NULL DEFAULT 0,
  "deleted" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  "updated_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY ("tenant_id", "source_type", "source_id"),
  CONSTRAINT "knowledge_sources_attributes_check" CHECK (jsonb_typeof(attributes) = 'object'::text),
  CONSTRAINT "knowledge_sources_chunks_check" CHECK ((chunk_count >= 0) AND ((NOT deleted) OR (chunk_count = 0))),
  CONSTRAINT "knowledge_sources_revision_check" CHECK (source_revision >= 0),
  CONSTRAINT "knowledge_sources_type_check" CHECK (source_type = ANY (ARRAY['product'::text, 'document'::text, 'script'::text]))
);
-- Create "knowledge_chunks" table
CREATE TABLE "ai_runtime"."knowledge_chunks" (
  "tenant_id" text NOT NULL,
  "source_type" text NOT NULL,
  "source_id" text NOT NULL,
  "chunk_index" integer NOT NULL,
  "content" text NOT NULL,
  "token_estimate" integer NOT NULL,
  "embedding_model" text NOT NULL,
  "embedding" public.vector(768) NOT NULL,
  PRIMARY KEY ("tenant_id", "source_type", "source_id", "chunk_index"),
  CONSTRAINT "knowledge_chunks_source_fkey" FOREIGN KEY ("tenant_id", "source_type", "source_id") REFERENCES "ai_runtime"."knowledge_sources" ("tenant_id", "source_type", "source_id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "knowledge_chunks_index_check" CHECK (chunk_index >= 0),
  CONSTRAINT "knowledge_chunks_tokens_check" CHECK (token_estimate > 0)
);
-- Create index "knowledge_chunks_embedding" to table: "knowledge_chunks"
CREATE INDEX "knowledge_chunks_embedding" ON "ai_runtime"."knowledge_chunks" USING HNSW ("embedding" public.vector_cosine_ops);
