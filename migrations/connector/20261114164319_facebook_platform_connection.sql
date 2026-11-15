-- Modify "connections" table
ALTER TABLE "connector"."connections" ADD CONSTRAINT "connections_check" CHECK ((credential_ciphertext IS NULL) = (key_version IS NULL)), ADD COLUMN "credential_ciphertext" bytea NULL, ADD COLUMN "key_version" integer NULL, ADD COLUMN "token_updated_at" timestamptz NULL;
-- Create "oauth_sessions" table
CREATE TABLE "connector"."oauth_sessions" (
  "id" uuid NOT NULL,
  "app_id" text NOT NULL,
  "payload_ciphertext" bytea NOT NULL,
  "key_version" integer NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "connector_oauth_sessions_expiry" to table: "oauth_sessions"
CREATE INDEX "connector_oauth_sessions_expiry" ON "connector"."oauth_sessions" ("expires_at");
-- Create "outbound_attempts" table
CREATE TABLE "connector"."outbound_attempts" (
  "attempt_id" uuid NOT NULL,
  "intent_id" uuid NOT NULL,
  "connection_id" uuid NOT NULL,
  "status" text NOT NULL,
  "provider_message_id" text NULL,
  "error_category" text NULL,
  "error_code" text NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("attempt_id"),
  CONSTRAINT "outbound_attempts_status_check" CHECK (status = ANY (ARRAY['dispatching'::text, 'sent'::text, 'failed'::text, 'unknown'::text]))
);
-- Create "outbound_events" table
CREATE TABLE "connector"."outbound_events" (
  "id" uuid NOT NULL,
  "topic" text NOT NULL,
  "event_key" text NOT NULL,
  "payload" bytea NOT NULL,
  "state" text NOT NULL,
  "fence" bigint NOT NULL DEFAULT 0,
  "attempts" integer NOT NULL DEFAULT 0,
  "available_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "outbound_events_state_check" CHECK (state = ANY (ARRAY['pending'::text, 'leased'::text, 'published'::text, 'quarantine'::text])),
  CONSTRAINT "outbound_events_topic_check" CHECK (topic = ANY (ARRAY['hexta.outbound.result.v1'::text, 'hexta.connector.health.v1'::text]))
);
-- Create index "connector_outbound_events_ready" to table: "outbound_events"
CREATE INDEX "connector_outbound_events_ready" ON "connector"."outbound_events" ("state", "available_at");
