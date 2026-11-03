CREATE SCHEMA "connector";
CREATE TABLE "connector"."connections" (
 id uuid PRIMARY KEY,
 version bigint NOT NULL CHECK(version>0),
 tenant_id varchar(36) NOT NULL,
 platform text NOT NULL CHECK(platform='facebook'),
 account_id text NOT NULL,
 app_id text NOT NULL,
 enabled boolean NOT NULL,
 credential_ref text NOT NULL DEFAULT '',
 UNIQUE(platform,account_id)
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
