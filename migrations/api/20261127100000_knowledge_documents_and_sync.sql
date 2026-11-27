-- Modify "messaging_reply_proposals" table
ALTER TABLE "messaging_reply_proposals" ADD CONSTRAINT "messaging_reply_proposals_citations_check" CHECK ((jsonb_typeof(cited_source_ids) = 'array'::text) AND (jsonb_array_length(cited_source_ids) <= 20)), ADD COLUMN "cited_source_ids" jsonb NOT NULL DEFAULT '[]';
-- Create "knowledge_documents" table
CREATE TABLE "knowledge_documents" (
  "id" character varying(36) NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "title" character varying(200) NOT NULL,
  "body" text NOT NULL,
  "version" bigint NOT NULL,
  "created_by" character varying(255) NOT NULL,
  "updated_by" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "knowledge_documents_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "knowledge_documents_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "knowledge_documents_body_check" CHECK ((char_length(body) >= 1) AND (char_length(body) <= 20000)),
  CONSTRAINT "knowledge_documents_title_check" CHECK ((char_length((title)::text) >= 1) AND (char_length((title)::text) <= 200)),
  CONSTRAINT "knowledge_documents_version_check" CHECK (version > 0)
);
-- Create index "idx_knowledge_documents_tenant_updated" to table: "knowledge_documents"
CREATE INDEX "idx_knowledge_documents_tenant_updated" ON "knowledge_documents" ("tenant_id", "updated_at" DESC, "id") WHERE (deleted_at IS NULL);
-- Create "knowledge_sync_jobs" table
CREATE TABLE "knowledge_sync_jobs" (
  "tenant_id" character varying(36) NOT NULL,
  "source_type" character varying(16) NOT NULL,
  "source_id" character varying(36) NOT NULL,
  "state" character varying(16) NOT NULL,
  "attempts" integer NOT NULL DEFAULT 0,
  "fence" bigint NOT NULL,
  "available_at" timestamptz NOT NULL,
  "last_error" character varying(64) NULL,
  "revision" bigint NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("tenant_id", "source_type", "source_id"),
  CONSTRAINT "knowledge_sync_jobs_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "knowledge_sync_jobs_attempts_check" CHECK (attempts >= 0),
  CONSTRAINT "knowledge_sync_jobs_fence_check" CHECK (fence > 0),
  CONSTRAINT "knowledge_sync_jobs_source_type_check" CHECK ((source_type)::text = ANY (ARRAY[('product'::character varying)::text, ('document'::character varying)::text, ('script'::character varying)::text])),
  CONSTRAINT "knowledge_sync_jobs_state_check" CHECK ((state)::text = ANY (ARRAY[('pending'::character varying)::text, ('leased'::character varying)::text, ('done'::character varying)::text, ('failed'::character varying)::text]))
);
-- Create index "idx_knowledge_sync_jobs_due" to table: "knowledge_sync_jobs"
CREATE INDEX "idx_knowledge_sync_jobs_due" ON "knowledge_sync_jobs" ("available_at") WHERE ((state)::text = ANY (ARRAY[('pending'::character varying)::text, ('leased'::character varying)::text]));
