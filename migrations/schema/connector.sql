CREATE SCHEMA "connector";
CREATE TABLE "connector"."connections" (
 id uuid PRIMARY KEY,
 version bigint NOT NULL CHECK(version>0),
 tenant_id varchar(36) NOT NULL,
 platform text NOT NULL CHECK(platform IN ('facebook','telegram')),
 account_id text NOT NULL,
 app_id text NOT NULL,
 enabled boolean NOT NULL,
 credential_ref text NOT NULL DEFAULT '',
 -- AES-256-GCM encrypted provider credential (Facebook Page token, Telegram Bot Token); wiped on disconnect.
 credential_ciphertext bytea NULL,
 key_version int NULL,
 token_updated_at timestamptz NULL,
 -- AES-256-GCM encrypted Telegram webhook secret_token (same key_version as the credential); wiped on disconnect.
 webhook_secret_ciphertext bytea NULL,
 UNIQUE(platform,account_id),
 CHECK((credential_ciphertext IS NULL) = (key_version IS NULL)),
 CHECK(webhook_secret_ciphertext IS NULL OR credential_ciphertext IS NOT NULL)
);
CREATE TABLE "connector"."receipts" (
 id text NOT NULL,
 app_id text NOT NULL,
 received_at timestamptz NOT NULL,
 PRIMARY KEY(id,app_id)
);
CREATE TABLE "connector"."events" (
 id uuid PRIMARY KEY,
 receipt_id text NOT NULL,
 app_id text NOT NULL,
 event_key text UNIQUE NOT NULL,
 payload bytea,
 state text NOT NULL CHECK(state IN ('pending','leased','published','quarantine')),
 reason text NOT NULL DEFAULT '',
 fence bigint NOT NULL DEFAULT 0,
 attempts int NOT NULL DEFAULT 0,
 available_at timestamptz NOT NULL,
 FOREIGN KEY(receipt_id,app_id) REFERENCES connector.receipts(id,app_id)
);
CREATE INDEX "connector_events_ready" ON "connector"."events"(state,available_at);
CREATE TABLE "connector"."sync_jobs" (
 id uuid PRIMARY KEY,
 app_id text NOT NULL,
 platform text NOT NULL,
 account_id text NOT NULL,
 cursor text NOT NULL DEFAULT '',
 mode text NOT NULL CHECK(mode IN ('history','incremental')),
 page_limit int NOT NULL CHECK(page_limit BETWEEN 1 AND 100),
 state text NOT NULL CHECK(state IN ('pending','leased','done','quarantine','cancelled')),
 reason text NOT NULL DEFAULT '',
 fence bigint NOT NULL DEFAULT 0,
 attempts int NOT NULL DEFAULT 0,
 available_at timestamptz NOT NULL
);
-- Ephemeral encrypted Page token staging between OAuth exchange and connection provisioning.
CREATE TABLE "connector"."oauth_sessions" (
 id uuid PRIMARY KEY,
 app_id text NOT NULL,
 payload_ciphertext bytea NOT NULL,
 key_version int NOT NULL,
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL
);
CREATE INDEX "connector_oauth_sessions_expiry" ON "connector"."oauth_sessions"(expires_at);
-- Connector-side idempotency: one provider call per attempt_id, never retried automatically.
CREATE TABLE "connector"."outbound_attempts" (
 attempt_id uuid PRIMARY KEY,
 intent_id uuid NOT NULL,
 connection_id uuid NOT NULL,
 status text NOT NULL CHECK(status IN ('dispatching','sent','failed','unknown')),
 provider_message_id text NULL,
 error_category text NULL,
 error_code text NULL,
 created_at timestamptz NOT NULL
);
-- Result/health outbox published to Kafka after the durable outcome commits.
CREATE TABLE "connector"."outbound_events" (
 id uuid PRIMARY KEY,
 topic text NOT NULL CHECK(topic IN ('hexta.delivery.send-completed.v1','hexta.connection.health-changed.v1')),
 event_key text NOT NULL,
 payload bytea NOT NULL,
 state text NOT NULL CHECK(state IN ('pending','leased','published','quarantine')),
 fence bigint NOT NULL DEFAULT 0,
 attempts int NOT NULL DEFAULT 0,
 available_at timestamptz NOT NULL
);
CREATE INDEX "connector_outbound_events_ready" ON "connector"."outbound_events"(state,available_at);
