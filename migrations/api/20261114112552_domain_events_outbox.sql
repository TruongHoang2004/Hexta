-- Create "domain_events" table
CREATE TABLE "domain_events" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "event_type" character varying(64) NOT NULL,
  "aggregate_id" character varying(36) NOT NULL,
  "payload" jsonb NOT NULL,
  "state" character varying(20) NOT NULL DEFAULT 'pending',
  "attempts" integer NOT NULL DEFAULT 0,
  "fence" bigint NOT NULL DEFAULT 0,
  "available_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  "created_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  "published_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "chk_domain_events_state" CHECK ((state)::text = ANY (ARRAY[('pending'::character varying)::text, ('leased'::character varying)::text, ('published'::character varying)::text, ('failed'::character varying)::text]))
);
-- Create index "idx_domain_events_pending" to table: "domain_events"
CREATE INDEX "idx_domain_events_pending" ON "domain_events" ("available_at") WHERE ((state)::text = ANY (ARRAY[('pending'::character varying)::text, ('leased'::character varying)::text]));
