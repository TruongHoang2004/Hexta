-- Create "messaging_reply_proposals" table
CREATE TABLE "messaging_reply_proposals" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "thread_id" uuid NOT NULL,
  "turn_id" uuid NOT NULL,
  "trigger_message_id" uuid NOT NULL,
  "control_generation" bigint NOT NULL,
  "context_revision" bigint NOT NULL,
  "status" character varying(16) NOT NULL,
  "route" character varying(16) NULL,
  "source" character varying(16) NULL,
  "script_version_id" character varying(36) NULL,
  "ai_run_id" character varying(100) NULL,
  "text" text NULL,
  "reason" character varying(64) NULL,
  "requested_by" character varying(255) NULL,
  "decided_by" character varying(255) NULL,
  "sent_message_id" uuid NULL,
  "attempts" integer NOT NULL DEFAULT 0,
  "available_at" timestamptz NOT NULL,
  "version" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messaging_reply_proposals_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "messaging_reply_proposals_script_fkey" FOREIGN KEY ("tenant_id", "script_version_id") REFERENCES "script_versions" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_reply_proposals_sent_fkey" FOREIGN KEY ("tenant_id", "sent_message_id") REFERENCES "messaging_messages" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_reply_proposals_trigger_fkey" FOREIGN KEY ("tenant_id", "trigger_message_id") REFERENCES "messaging_messages" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_reply_proposals_turn_fkey" FOREIGN KEY ("tenant_id", "thread_id", "turn_id") REFERENCES "messaging_turns" ("tenant_id", "thread_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "messaging_reply_proposals_candidate_check" CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('sent'::character varying)::text, ('auto_sent'::character varying)::text, ('discarded'::character varying)::text])) <= ((source IS NOT NULL) AND (text IS NOT NULL))),
  CONSTRAINT "messaging_reply_proposals_generation_check" CHECK (control_generation > 0),
  CONSTRAINT "messaging_reply_proposals_revision_check" CHECK (context_revision > 0),
  CONSTRAINT "messaging_reply_proposals_route_check" CHECK ((route)::text = ANY (ARRAY[('script'::character varying)::text, ('answer'::character varying)::text, ('clarify'::character varying)::text, ('handoff'::character varying)::text])),
  CONSTRAINT "messaging_reply_proposals_script_source_check" CHECK (((source)::text = 'script'::text) = (script_version_id IS NOT NULL)),
  CONSTRAINT "messaging_reply_proposals_sent_check" CHECK (((status)::text = ANY (ARRAY[('sent'::character varying)::text, ('auto_sent'::character varying)::text])) = (sent_message_id IS NOT NULL)),
  CONSTRAINT "messaging_reply_proposals_source_check" CHECK ((source)::text = ANY (ARRAY[('script'::character varying)::text, ('ai'::character varying)::text])),
  CONSTRAINT "messaging_reply_proposals_status_check" CHECK ((status)::text = ANY (ARRAY[('routing'::character varying)::text, ('pending'::character varying)::text, ('sent'::character varying)::text, ('auto_sent'::character varying)::text, ('discarded'::character varying)::text, ('superseded'::character varying)::text, ('handed_off'::character varying)::text, ('no_candidate'::character varying)::text])),
  CONSTRAINT "messaging_reply_proposals_text_check" CHECK ((char_length(text) >= 1) AND (char_length(text) <= 2000)),
  CONSTRAINT "messaging_reply_proposals_version_check" CHECK (version > 0)
);
-- Create index "idx_messaging_reply_proposals_one_live" to table: "messaging_reply_proposals"
CREATE UNIQUE INDEX "idx_messaging_reply_proposals_one_live" ON "messaging_reply_proposals" ("tenant_id", "turn_id") WHERE ((status)::text = ANY (ARRAY[('routing'::character varying)::text, ('pending'::character varying)::text]));
-- Create index "idx_messaging_reply_proposals_routing" to table: "messaging_reply_proposals"
CREATE INDEX "idx_messaging_reply_proposals_routing" ON "messaging_reply_proposals" ("available_at") WHERE ((status)::text = 'routing'::text);
-- Create index "idx_messaging_reply_proposals_turn" to table: "messaging_reply_proposals"
CREATE INDEX "idx_messaging_reply_proposals_turn" ON "messaging_reply_proposals" ("tenant_id", "turn_id", "id" DESC);
-- Create "tenant_reply_settings" table
CREATE TABLE "tenant_reply_settings" (
  "tenant_id" character varying(36) NOT NULL,
  "reply_mode" character varying(16) NOT NULL,
  "version" bigint NOT NULL,
  "updated_by" character varying(255) NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("tenant_id"),
  CONSTRAINT "tenant_reply_settings_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "tenant_reply_settings_mode_check" CHECK ((reply_mode)::text = ANY (ARRAY[('suggest'::character varying)::text, ('auto_send'::character varying)::text])),
  CONSTRAINT "tenant_reply_settings_version_check" CHECK (version > 0)
);
