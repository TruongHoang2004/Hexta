-- Add new schema named "connector"
CREATE SCHEMA "connector";
-- Create "connections" table
CREATE TABLE "connector"."connections" (
  "id" uuid NOT NULL,
  "version" bigint NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "platform" text NOT NULL,
  "account_id" text NOT NULL,
  "app_id" text NOT NULL,
  "enabled" boolean NOT NULL,
  "credential_ref" text NOT NULL DEFAULT '',
  PRIMARY KEY ("id"),
  CONSTRAINT "connections_platform_account_id_key" UNIQUE ("platform", "account_id"),
  CONSTRAINT "connections_platform_check" CHECK (platform = 'facebook'::text),
  CONSTRAINT "connections_version_check" CHECK (version > 0)
);
-- Create "receipts" table
CREATE TABLE "connector"."receipts" (
  "id" text NOT NULL,
  "app_id" text NOT NULL,
  "received_at" timestamptz NOT NULL,
  PRIMARY KEY ("id", "app_id")
);
-- Create "events" table
CREATE TABLE "connector"."events" (
  "id" uuid NOT NULL,
  "receipt_id" text NOT NULL,
  "app_id" text NOT NULL,
  "event_key" text NOT NULL,
  "payload" bytea NULL,
  "state" text NOT NULL,
  "reason" text NOT NULL DEFAULT '',
  "fence" bigint NOT NULL DEFAULT 0,
  "attempts" integer NOT NULL DEFAULT 0,
  "available_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "events_event_key_key" UNIQUE ("event_key"),
  CONSTRAINT "events_receipt_id_app_id_fkey" FOREIGN KEY ("receipt_id", "app_id") REFERENCES "connector"."receipts" ("id", "app_id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "events_state_check" CHECK (state = ANY (ARRAY['pending'::text, 'leased'::text, 'published'::text, 'quarantine'::text]))
);
-- Create index "connector_events_ready" to table: "events"
CREATE INDEX "connector_events_ready" ON "connector"."events" ("state", "available_at");
