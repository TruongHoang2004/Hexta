-- Create "channel_oauth_states" table
CREATE TABLE "channel_oauth_states" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "actor_id" character varying(255) NOT NULL,
  "platform" text NOT NULL,
  "state_hash" character varying(64) NOT NULL,
  "redirect_uri" text NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "consumed_at" timestamptz NULL,
  "oauth_session_id" text NULL,
  "oauth_session_expires_at" timestamptz NULL,
  "pages" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "channel_oauth_states_state_hash_key" UNIQUE ("state_hash"),
  CONSTRAINT "channel_oauth_states_tenant_id_fkey" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "channel_oauth_states_platform_check" CHECK (platform = 'facebook'::text)
);
-- Create index "idx_channel_oauth_states_session" to table: "channel_oauth_states"
CREATE INDEX "idx_channel_oauth_states_session" ON "channel_oauth_states" ("tenant_id", "oauth_session_id") WHERE (oauth_session_id IS NOT NULL);
-- Modify "messaging_connections" table
ALTER TABLE "messaging_connections" DROP CONSTRAINT "messaging_connections_platform_account_id_key", ADD CONSTRAINT "messaging_connections_status_check" CHECK (status = ANY (ARRAY['active'::text, 'paused'::text, 'disconnected'::text, 'auth_expired'::text])), ADD COLUMN "account_name" text NOT NULL DEFAULT '', ADD COLUMN "status" text NOT NULL DEFAULT 'active', ADD COLUMN "last_error_code" text NULL, ADD COLUMN "last_error_message" text NULL, ADD COLUMN "updated_at" timestamptz NOT NULL DEFAULT now();
-- Create index "idx_messaging_connections_tenant" to table: "messaging_connections"
CREATE INDEX "idx_messaging_connections_tenant" ON "messaging_connections" ("tenant_id", "platform");
-- Create index "uq_messaging_connections_active_page" to table: "messaging_connections"
CREATE UNIQUE INDEX "uq_messaging_connections_active_page" ON "messaging_connections" ("platform", "account_id") WHERE (status <> 'disconnected'::text);
-- Modify "messaging_threads" table
ALTER TABLE "messaging_threads" ADD CONSTRAINT "messaging_threads_check" CHECK ((connection_id IS NULL) = (recipient_id IS NULL)), ADD CONSTRAINT "messaging_threads_recipient_id_check" CHECK ((char_length(recipient_id) >= 1) AND (char_length(recipient_id) <= 255)), ADD COLUMN "connection_id" uuid NULL, ADD COLUMN "recipient_id" text NULL, ADD CONSTRAINT "messaging_threads_connection_fkey" FOREIGN KEY ("tenant_id", "connection_id") REFERENCES "messaging_connections" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "messaging_messages" table
ALTER TABLE "messaging_messages" DROP CONSTRAINT "messaging_messages_body_check", ADD CONSTRAINT "messaging_messages_body_check" CHECK (char_length(body) <= 4000), DROP CONSTRAINT "messaging_messages_status_check", ADD CONSTRAINT "messaging_messages_status_check" CHECK ((status)::text = ANY ((ARRAY['simulated'::character varying, 'received'::character varying, 'channel'::character varying])::text[])), ADD CONSTRAINT "messaging_messages_body_content_check" CHECK ((content_type = 'unsupported_media'::text) OR (char_length(body) >= 1)), ADD CONSTRAINT "messaging_messages_content_type_check" CHECK (content_type = ANY (ARRAY['text'::text, 'unsupported_media'::text])), ADD CONSTRAINT "messaging_messages_delivery_status_check" CHECK (delivery_status = ANY (ARRAY['received'::text, 'sending'::text, 'sent'::text, 'failed'::text, 'unknown'::text])), ADD COLUMN "content_type" text NOT NULL DEFAULT 'text', ADD COLUMN "placeholder_meta" jsonb NULL, ADD COLUMN "delivery_status" text NOT NULL DEFAULT 'received', ADD COLUMN "delivery_error_reason" text NULL, ADD CONSTRAINT "messaging_messages_tenant_id_id_key" UNIQUE ("tenant_id", "id");
-- Create "outbound_intents" table
CREATE TABLE "outbound_intents" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "connection_id" uuid NOT NULL,
  "message_id" uuid NOT NULL,
  "client_message_id" text NOT NULL,
  "request_hash" character varying(64) NOT NULL,
  "recipient_id" text NOT NULL,
  "status" text NOT NULL,
  "error_category" text NULL,
  "error_code" text NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "outbound_intents_tenant_id_client_message_id_key" UNIQUE ("tenant_id", "client_message_id"),
  CONSTRAINT "outbound_intents_tenant_id_id_key" UNIQUE ("tenant_id", "id"),
  CONSTRAINT "outbound_intents_tenant_id_connection_id_fkey" FOREIGN KEY ("tenant_id", "connection_id") REFERENCES "messaging_connections" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "outbound_intents_tenant_id_message_id_fkey" FOREIGN KEY ("tenant_id", "message_id") REFERENCES "messaging_messages" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "outbound_intents_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'sent'::text, 'failed'::text, 'unknown'::text, 'cancelled'::text]))
);
-- Create index "idx_outbound_intents_pending" to table: "outbound_intents"
CREATE INDEX "idx_outbound_intents_pending" ON "outbound_intents" ("connection_id") WHERE (status = 'pending'::text);
-- Create "outbound_attempts" table
CREATE TABLE "outbound_attempts" (
  "id" uuid NOT NULL,
  "tenant_id" character varying(36) NOT NULL,
  "intent_id" uuid NOT NULL,
  "connection_version" bigint NOT NULL,
  "status" text NOT NULL,
  "provider_message_id" text NULL,
  "created_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "outbound_attempts_tenant_id_intent_id_fkey" FOREIGN KEY ("tenant_id", "intent_id") REFERENCES "outbound_intents" ("tenant_id", "id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "outbound_attempts_connection_version_check" CHECK (connection_version > 0),
  CONSTRAINT "outbound_attempts_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'sent'::text, 'failed'::text, 'unknown'::text, 'cancelled'::text]))
);
-- Create index "idx_outbound_attempts_intent" to table: "outbound_attempts"
CREATE INDEX "idx_outbound_attempts_intent" ON "outbound_attempts" ("tenant_id", "intent_id");
