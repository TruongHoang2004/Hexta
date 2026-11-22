# Connector service: Facebook and Telegram ingress, connection and outbound text

Local implementation, unmerged. Telegram Bot connections (Bot Token onboarding/rotation, private 1-1 chats, outbound text) follow [the Telegram design](../../agentic-memory/designs/2026-11-16_telegram-bot-connection_design.md); see "Telegram bots" below. Inbound: signed Page webhook → connector PostgreSQL receipt/event outbox → Kafka → API-owned ingress consumer → tenant inbox (text, plus a media placeholder on the same message event). Connection: owner OAuth in the API → connector code exchange and sealed Page tokens → `PUT /internal/connections` with webhook subscription. Outbound: API intent outbox → `hexta.delivery.send-requested.v1` → one Send API call per attempt → `hexta.delivery.send-completed.v1` / `hexta.connection.health-changed.v1`. Inbound messages publish `hexta.messaging.message-received.v1`. Contracts: `packages/shared/events`. Python is not involved. See [the design](../../agentic-memory/designs/2026-11-14_facebook-platform-connection_design.md). This is at-least-once delivery; a successful provider HTTP response follows the connector transaction, never an in-memory queue or Kafka ACK.

## Ownership and rollout

`packages/shared/ingress` owns the handwritten additive v1 event contract; no generator exists for JSON events. Connector produces it, `services/api/cmd/ingress` consumes it, and existing API inbox DTOs add `source=facebook` and message `status=received`. Deploy API migrations and compatible consumer before enabling the producer; stop ingestion before rollback. Unknown envelope versions are quarantined by hash, without payload. Future event schemas must preserve old readers or use a new versioned topic.

Connector uses its own PostgreSQL schema/history/revision schema. Provision a dedicated login with access only to `connector`; API's login must have no access to that schema. API authority lives in public `messaging_connections`, not connector tables. Runtime logins do not own migrations. The initial implementation can use separate databases, as the tests do. Do not grant the connector access to API tables.

Connector management uses a service credential, never a user JWT. The API authorizes the tenant owner, then calls the machine endpoints:

- `POST /internal/facebook/oauth/exchange` `{code, redirect_uri}` → `{oauth_session_id, expires_at, pages[{page_id,page_name}]}`. Page tokens are sealed in `connector.oauth_sessions` for 10 minutes and never returned.
- `PUT /internal/connections` accepts an immutable ID/tenant/platform/Page/app identity and a strictly increasing config version (gaps allowed). It changes enabled state, and optionally `oauth_session_id` (seal that Page's token on the connection and subscribe its webhook) or `unsubscribe_on_disable` (disconnect: unsubscribe and wipe the credential). Replaying the current version is an idempotent no-op. Returns `200 {id, version, enabled, subscribed}`; 422 expired session/credential, 502 Meta failure, 409 conflict. A disconnected binding can be replaced by another tenant's new connection ID; paused/active ones cannot.

- `POST /internal/telegram/bots/verify` `{bot_token}` → `{bot_id, display_name, username}` (getMe). The token is not stored; 400 malformed, 422 rejected by Telegram, 502 Telegram unavailable, 503 Telegram not configured.
- For `platform: "telegram"`, `PUT /internal/connections` uses `app_id: "telegram"`, a numeric `account_id` (bot ID) and optional `bot_token` (connect/rotate; only with `enabled: true` and only for that bot). See "Telegram bots".

These are trusted machine endpoints: expose only on a private network/reverse-proxy policy. Only `/webhooks/facebook` and `/webhooks/telegram/{connectionId}` belong on the public HTTPS listener. A management token alone must never be delegated to a browser.

## Start locally

From repository root, configure private environment variables (template: [`.env.example`](.env.example)); values below are names, not credentials:

- `CONNECTOR_DATABASE_URL`: connector runtime PostgreSQL URL.
- `API_DATABASE_URL`: API runtime PostgreSQL URL for the separate consumer.
- `KAFKA_BROKERS`: comma-separated broker addresses reachable from the Go processes.
- `META_APP_ID`: the registered Meta app ID.
- `META_APP_SECRET_FILE`: private mounted file containing its app secret.
- `META_VERIFY_TOKEN_FILE`: private mounted file containing your verification token.
- `CONNECTOR_MANAGEMENT_TOKEN_FILE`: private mounted service token file (the API's `CONNECTOR_MANAGEMENT_TOKEN`).
- `CONNECTOR_CREDENTIAL_KEY_FILE`: private file with a 32-byte AES-256-GCM key as 64 hex characters. Rotating it makes stored Page tokens unreadable; owners must reconnect Pages.
- `META_GRAPH_URL` / `META_GRAPH_VERSION`: optional; default `https://graph.facebook.com` and `v21.0`. Override the URL only for isolated tests.
- `CONNECTOR_LISTEN_ADDR`: defaults to `127.0.0.1:8091`.
- `TELEGRAM_WEBHOOK_BASE_URL`: optional public HTTPS origin that reaches this connector (Telegram requires HTTPS). Unset disables Telegram management (503). Webhooks are registered at `<base>/webhooks/telegram/<connection id>`.
- `TELEGRAM_API_URL`: optional; defaults to `https://api.telegram.org`. Override only for isolated tests.

Secret files must have at least 16 non-whitespace bytes. The webhook receiver needs no Page token. Page tokens arrive only through the OAuth exchange and are stored sealed in `connector.connections`, bound to the connection ID; they never appear in connection JSON, Kafka, logs or API responses. `credential_ref` is a legacy opaque field.

Apply migrations explicitly using migration-role URLs in a separate shell, then switch to runtime URLs. This delivery has not migrated your application database:

```sh
make migrate-apply-local svc=connector
make migrate-apply-local svc=api
```

The new connector target is explicit/host-operated until Compose role initialization and private connector URL wiring are implemented. Existing `all` apply/status continues to select API/AI; `all` hash/validate includes connector.

Create the topics listed in `packages/shared/events/catalog` (Compose files provision them) using your broker's admin tooling; producer auto-creation is disabled. Configure retention, access controls, partition count and replication for the environment. Connector writes with `acks=all`; it does not claim producer-idempotence or cross-system exactly-once.

The normal path is the owner flow in the web app (Settings → Channels → Connect Facebook Page), which requires the API to have `CONNECTOR_URL`, `CONNECTOR_MANAGEMENT_TOKEN` and `META_APP_ID`, and Meta's valid OAuth redirect set to `<frontend>/channels/facebook/callback`. Manual operator provisioning (receive-only, no Page token) remains possible:

```sql
INSERT INTO messaging_connections(id,tenant_id,version,platform,account_id,enabled)
VALUES (:connection_id,:tenant_id,1,'facebook',:page_id,true);
```

The tenant must already exist and be active. Use the same identity and version in a non-secret JSON file, then make an authenticated machine `PUT /internal/connections` with fields `id`, `tenant_id`, `version`, `platform`, `account_id`, `app_id`, `enabled`, `credential_ref`. Supply the Authorization header from a private file or secret-aware client. Never paste the token into shell history. No user authorization API for provisioning is claimed.

```sh
go run ./services/connector/cmd/connector
go run ./services/api/cmd/ingress
```

Run these in separate terminals. Run the existing API/frontend as usual, then open `/tenants/<tenantId>/inbox` as a current member and Refresh. Facebook messages are marked as received. Real conversations cannot be mutated through simulator input or simulated reply; Facebook sending remains disabled in the composer and on the backend.

## Connect Meta for an authorized receive test

1. In your Meta app's Messenger/Webhooks setup, set the callback to `https://<your-controlled-host>/webhooks/facebook`. Route it to the connector over HTTPS without changing the raw body. Keep internal management private. Use the verification token stored in `META_VERIFY_TOKEN_FILE`.
2. Register the Page subscription for `messages` with the app/Page configuration and permissions actually available to your app. Confirm the Page ID matches the trusted bindings and the App ID matches the configured signature secret. A successful challenge does not prove Page subscription or permission access.
3. Use an account allowed by the app's current development/live mode to send a text to the Page. Confirm the callback returns 200, connector event reaches published, and the text appears as received in the member inbox. This guide authorizes no outbound Hexta send; do not use a send API for a smoke test.
4. Inspect counts/state/reason codes only. Never print webhook bodies, tokens or secret-bearing URLs. For a failure, verify signature configuration and subscription first, then DB backlog, Kafka connectivity and API binding/version.

Official references: [Meta webhook setup](https://developers.facebook.com/docs/graph-api/webhooks/getting-started), [Messenger webhooks](https://developers.facebook.com/docs/messenger-platform/webhooks), [Meta's Messenger collection](https://www.postman.com/meta/messenger-platform-api/documentation/iyp204x/messenger-platform-api). Direct Meta docs returned HTTP 429 during this session. Meta's collection documents the Page and `pages_messaging` prerequisites; no app-specific permission or live capability was verified. Real Conversations API access, scopes and Graph version must be checked before implementing Facebook fetch. Reference-project code is not evidence of your app's access.

## Telegram bots

Owners paste a @BotFather Bot Token in Settings → Channels. The API checks the owner, the connector calls `getMe`, and the API binds the bot ID to the tenant (1 Bot = 1 tenant via the same partial unique index as Pages). The connector then seals the token and a fresh random webhook `secret_token` (AES-256-GCM, bound to the connection ID) in `connector.connections`, commits, and calls `setWebhook` with `allowed_updates=["message"]`. Rotation must present a token for the same bot ID (else API 409), gets a new secret and re-registers the webhook. Pause calls `deleteWebhook(drop_pending_updates=true)`; resume re-registers with the stored secret; disconnect deletes the webhook and wipes token and secret. `deleteWebhook` is best effort: a revoked token cannot delete it, and a disabled connection quarantines updates anyway.

`POST /webhooks/telegram/{connectionId}` compares `X-Telegram-Bot-Api-Secret-Token` with the decrypted secret in constant time (403 when wrong or the connection/secret is unknown). Only `message` updates in `private` chats are accepted: text → `event_type=message`; photo/video/animation/video note/audio/voice/document/sticker → `unsupported_media` with kinds only (captions dropped, no file IDs/URLs). Group, supergroup and channel messages and other update types are acknowledged with 200 and dropped before the outbox. Other private content (location, contact, …) and text over 4,000 characters are quarantined. Accepted updates return 200 only after receipt + outbox commit (receipt hash = connection + body, so Telegram redelivery deduplicates); storage failure is 503 so Telegram retries; an unparseable verified body is 400.

`sendMessage` outcomes: 200 → `sent`; 403 → `bot_blocked`; 400 "chat not found" → `chat_not_found`; 401/404 → `auth_expired` plus a health event (API marks the bot `auth_expired`; rotate the token); 429 → `rate_limit`; other 4xx → failed `transport`; 5xx/network → `unknown`. None are retried automatically. Tokens appear only in the Bot API request path; errors and logs never include URLs.

Isolated tests use a mock Bot API and generate all connector secrets at runtime. Live acceptance is `TestLiveTelegramWebhookAcceptance` (`services/api/internal/bootstrap`), skipped unless configured through the environment only — never commit or write these values to files:

- `TELEGRAM_LIVE_BOT_TOKEN`: a dedicated test bot's token. Use a separate name from `TELEGRAM_BOT_TOKEN`, which the agent notification scripts read from the git-ignored root `.env`.
- `TELEGRAM_LIVE_WEBHOOK_BASE_URL`: a public HTTPS origin that forwards only `POST /webhooks/telegram/*` to `TELEGRAM_LIVE_CONNECTOR_ADDR` (default `127.0.0.1:18091`).
- `TELEGRAM_LIVE_WAIT`: optional, default `5m`.

With the isolated variables below also set, run `go test -race -count=1 -v -run TestLiveTelegramWebhookAcceptance ./services/api/internal/bootstrap`, then send the bot a text and a photo in a private chat. The test checks inbound text and placeholder, one live reply (`sent`) and webhook removal on disconnect.

## Recovery and operational limits

- One callback supports up to 1 MiB and 1,000 items. All entries are visited. Invalid/malformed/unsupported items are quarantined alongside valid siblings, without raw payload retention. Malformed whole batches receive 503 for retry. Text and echo events are published; image/video/audio/file/sticker attachments are published as `unsupported_media` with kinds only (no URLs or binaries). Other attachment kinds, read/delivery/reaction and other events carry safe quarantine classifications.
- Receipt hash deduplicates identical batches; source message identity deduplicates different batches and webhook/poll overlap. Source revision is empty unless the provider supplies one. No invented revisions or media metadata. Message content and IDs stay private Kafka/DB data; logs contain neither.
- Normalization happens before ACK, in the request; no post-ACK normalization queue exists. Persisting normalized outbox rows is the durable continuation point. A failed DB transaction/100,000 pending-item admission limit returns 503. Acceptance takes only row-level locks, so callbacks with different receipts commit concurrently; the admission limit is therefore approximate and concurrent batches can overshoot it by up to (connector DB pool size − 1) × 1,000 items. Concurrent retries of one receipt wait on its row and are deduplicated, and overlapping batches insert events in key order to avoid deadlocks. Receipt and published/quarantine retention cleanup is not automated; monitor DB space and apply a reviewed retention policy.
- Outbox claims use `SKIP LOCKED`, a 30-second lease, fencing, 10-second publication timeout and exponential backoff capped at 300 seconds; after 12 attempts the event is quarantined. A crash after broker ACK but before marking published can duplicate publication. API dedup plus mutation commit together. A crash after API commit before offset commit replays without duplicating the message.
- Quarantine stores safe reason codes and receipt/app identity; unsupported raw content cannot be reprocessed from these records. Re-ingest from the provider only after implementing the missing mapping. Exhausted publication rows retain their normalized payload and can be deliberately requeued by an operator after diagnosis. Never replay with a new source identity to bypass dedup.
- Disable the API binding first, then update connector enabled=false and increment its version. API rechecks its current binding/tenant inside the transaction. Already committed inbox messages remain. Delayed stale-version events are handled without business mutation. Credential rotation increments both versions and reloads mounted app-secret files by restarting the connector explicitly; in-flight callbacks with the retired secret may fail and require provider retry. There is no dual-secret grace period yet.
- On a binding change, stage the API version first. There is no atomic configuration transaction across services: older delayed events may be ignored during rollout. Never change a connection's tenant in place; provision a new ID after disabling the old binding. Operator tooling must enforce the same invariant in the API mirror.
- API malformed envelopes/untrusted bindings are quarantined by hash. DB failures retry the current record before offset advancement. This single-record consumer trades throughput for simple commit safety; concurrent partition processing/metrics/alerting are follow-ups.

## Sync and outbound boundary

`SyncService` and durable `sync_jobs` implement page fetching, mode, bounded limits, checkpoint-after-durable-data, lease/fence, retry-after, bounded attempts and terminal cursor-expired recovery. A crash between page storage and checkpoint refetches that page and deduplicates by provider identity. Cancellation sets state=cancelled and increments fence; an already fetched page can persist but cannot advance the checkpoint. History and incremental reconciliation never open active turns or schedule AI. Fixture tests cover the framework; no runtime sync scheduler/management API is wired. Facebook advertises Fetch=false and returns unsupported. Empty/expired cursors must not silently restart a full history: create an explicit reconciliation job after provider-specific recovery is defined.

Outbound text: the API checks tenant/assignee/context and an active connection, then commits message + intent + attempt + outbox row together. The connector claims `attempt_id` in `connector.outbound_attempts` before calling Meta, rejects expired intents and version/enabled mismatches without a provider call, and records the outcome plus its result (and `auth_expired` health) event in one transaction. A redelivered intent whose attempt is still `dispatching` (crash during the call) becomes `unknown` and is never resent; network/5xx failures are also `unknown`. Meta codes map to `window_24h` (10/2022), `auth_expired` (190, 200–299), `rate_limit`, `invalid_recipient` or `transport`. Pausing/disconnecting cancels pending API intents, but cannot recall an intent already published to Kafka; such a send still reports its real outcome. Provider accepted/delivered/read are separate states; only acceptance is tracked. AI candidates require API revalidation; nothing here schedules AI.

## Isolated verification

Use only disposable infrastructure. The fixture Compose file is independent from the application stack:

```sh
docker compose -f infrastructure/compose.connector-test.yml up -d
docker compose -f infrastructure/compose.connector-test.yml exec postgres psql -U postgres -c 'CREATE DATABASE connector_fixture;'
# Then:
export CONNECTOR_DATABASE_URL='postgres://postgres:fixture@127.0.0.1:55439/connector_fixture?sslmode=disable'
make migrate-apply-local svc=connector
export CONNECTOR_TEST_DATABASE_URL="$CONNECTOR_DATABASE_URL"
export API_TEST_DATABASE_URL='postgres://postgres:fixture@127.0.0.1:55439/postgres?sslmode=disable'
export KAFKA_TEST_BROKERS=127.0.0.1:59092
go test -race -p 1 ./services/connector/... ./services/api/... ./packages/shared/...
make migrate-validate
pnpm --filter web lint
pnpm --filter web build
```

API tests create/drop random databases with Atlas; connector tests require the specifically named migrated disposable `connector_fixture`. The E2E test builds and starts a real connector subprocess, provisions through machine HTTP, sends signed duplicate callbacks while Kafka is unreachable, kills/restarts the connector against the broker, consumes Kafka into a random API DB and verifies tenant-authorized inbox reads. It creates the test topic. Run tests serially against a fresh isolated broker; the API consumer group is fixed. See the EPIC-008 session for exact results and NOT_RUN fault-injection/live checks.
