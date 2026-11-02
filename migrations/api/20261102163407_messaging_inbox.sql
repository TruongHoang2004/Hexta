-- Create "messaging_threads" table
CREATE TABLE "messaging_threads" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "external_key" character varying(120) NOT NULL,
  "display_name" character varying(120) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messaging_threads_tenant_id_external_key_key" UNIQUE ("tenant_id", "external_key"),
  CONSTRAINT "messaging_threads_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "messaging_threads_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_messaging_threads_list" to table: "messaging_threads"
CREATE INDEX "idx_messaging_threads_list" ON "messaging_threads" ("tenant_id", "id" DESC);
-- Create "messaging_turns" table
CREATE TABLE "messaging_turns" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "thread_id" uuid NOT NULL,
  "lifecycle" character varying(16) NOT NULL,
  "assignee_id" character varying(255) NULL,
  "control" character varying(16) NOT NULL,
  "needs_human" boolean NOT NULL DEFAULT false,
  "context_revision" bigint NOT NULL,
  "control_generation" bigint NOT NULL,
  "version" bigint NOT NULL,
  "last_activity_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messaging_turns_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "messaging_turns_tenant_id_thread_id_id_key" UNIQUE ("tenant_id", "thread_id", "id"),
  CONSTRAINT "messaging_turns_tenant_id_thread_id_fkey" FOREIGN KEY ("tenant_id", "thread_id") REFERENCES "messaging_threads" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_turns_context_revision_check" CHECK (context_revision > 0),
  CONSTRAINT "messaging_turns_control_check" CHECK ((control)::text = ANY ((ARRAY['system'::character varying, 'sales'::character varying])::text[])),
  CONSTRAINT "messaging_turns_control_generation_check" CHECK (control_generation > 0),
  CONSTRAINT "messaging_turns_lifecycle_check" CHECK ((lifecycle)::text = ANY ((ARRAY['open'::character varying, 'closed'::character varying])::text[])),
  CONSTRAINT "messaging_turns_version_check" CHECK (version > 0)
);
-- Create index "idx_messaging_one_open_turn" to table: "messaging_turns"
CREATE UNIQUE INDEX "idx_messaging_one_open_turn" ON "messaging_turns" ("tenant_id", "thread_id") WHERE ((lifecycle)::text = 'open'::text);
-- Create index "idx_messaging_turn_assignment" to table: "messaging_turns"
CREATE INDEX "idx_messaging_turn_assignment" ON "messaging_turns" ("tenant_id", "assignee_id", "thread_id") WHERE ((lifecycle)::text = 'open'::text);
-- Create "messaging_messages" table
CREATE TABLE "messaging_messages" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "thread_id" uuid NOT NULL,
  "turn_id" uuid NOT NULL,
  "direction" character varying(16) NOT NULL,
  "actor_id" character varying(255) NOT NULL,
  "body" text NOT NULL,
  "request_key" uuid NOT NULL,
  "request_hash" character varying(64) NOT NULL,
  "status" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messaging_messages_tenant_id_request_key_key" UNIQUE ("tenant_id", "request_key"),
  CONSTRAINT "messaging_messages_tenant_id_thread_id_turn_id_fkey" FOREIGN KEY ("tenant_id", "thread_id", "turn_id") REFERENCES "messaging_turns" ("tenant_id", "thread_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_messages_body_check" CHECK ((char_length(body) >= 1) AND (char_length(body) <= 4000)),
  CONSTRAINT "messaging_messages_direction_check" CHECK ((direction)::text = ANY ((ARRAY['inbound'::character varying, 'outbound'::character varying])::text[])),
  CONSTRAINT "messaging_messages_status_check" CHECK ((status)::text = 'simulated'::text)
);
-- Create index "idx_messaging_messages_history" to table: "messaging_messages"
CREATE INDEX "idx_messaging_messages_history" ON "messaging_messages" ("tenant_id", "thread_id", "id" DESC);
