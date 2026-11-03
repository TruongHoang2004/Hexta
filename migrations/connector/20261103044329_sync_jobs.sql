-- Create "sync_jobs" table
CREATE TABLE "connector"."sync_jobs" (
  "id" uuid NOT NULL,
  "app_id" text NOT NULL,
  "platform" text NOT NULL,
  "account_id" text NOT NULL,
  "cursor" text NOT NULL DEFAULT '',
  "mode" text NOT NULL,
  "page_limit" integer NOT NULL,
  "state" text NOT NULL,
  "reason" text NOT NULL DEFAULT '',
  "fence" bigint NOT NULL DEFAULT 0,
  "attempts" integer NOT NULL DEFAULT 0,
  "available_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sync_jobs_mode_check" CHECK (mode = ANY (ARRAY['history'::text, 'incremental'::text])),
  CONSTRAINT "sync_jobs_page_limit_check" CHECK ((page_limit >= 1) AND (page_limit <= 100)),
  CONSTRAINT "sync_jobs_state_check" CHECK (state = ANY (ARRAY['pending'::text, 'leased'::text, 'done'::text, 'quarantine'::text, 'cancelled'::text]))
);
