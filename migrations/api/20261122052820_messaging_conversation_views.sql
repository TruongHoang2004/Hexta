-- Create "messaging_conversation_views" table
CREATE TABLE "messaging_conversation_views" (
  "tenant_id" character varying(36) NOT NULL,
  "thread_id" uuid NOT NULL,
  "external_key" character varying(120) NOT NULL,
  "display_name" character varying(120) NOT NULL,
  "thread_created_at" timestamptz NOT NULL,
  "thread_updated_at" timestamptz NOT NULL,
  "connection_id" uuid NULL,
  "recipient_id" text NULL,
  "turn_id" uuid NOT NULL,
  "lifecycle" character varying(16) NOT NULL,
  "control" character varying(16) NOT NULL,
  "assignee_id" character varying(255) NULL,
  "needs_human" boolean NOT NULL DEFAULT false,
  "context_revision" bigint NOT NULL,
  "control_generation" bigint NOT NULL,
  "turn_version" bigint NOT NULL,
  "last_activity_at" timestamptz NOT NULL,
  PRIMARY KEY ("tenant_id", "thread_id"),
  CONSTRAINT "messaging_conversation_views_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_conversation_views_control_check" CHECK ((control)::text = ANY ((ARRAY['system'::character varying, 'sales'::character varying])::text[])),
  CONSTRAINT "messaging_conversation_views_lifecycle_check" CHECK ((lifecycle)::text = ANY ((ARRAY['open'::character varying, 'closed'::character varying])::text[]))
);
-- Create index "idx_messaging_conversation_views_assignee" to table: "messaging_conversation_views"
CREATE INDEX "idx_messaging_conversation_views_assignee" ON "messaging_conversation_views" ("tenant_id", "assignee_id", "thread_id" DESC);
-- Create index "idx_messaging_conversation_views_list" to table: "messaging_conversation_views"
CREATE INDEX "idx_messaging_conversation_views_list" ON "messaging_conversation_views" ("tenant_id", "thread_id" DESC);
