# Connector service: Facebook text ingress

Local implementation, unmerged. The runnable path is signed Page webhook → connector PostgreSQL receipt/event outbox → Kafka → API-owned ingress consumer → existing tenant inbox. Python and outbound sending are not involved. This is at-least-once delivery; a successful provider HTTP response follows the connector transaction, never an in-memory queue or Kafka ACK.

## Ownership and rollout

`packages/shared/ingress` owns the handwritten additive v1 event contract; no generator exists for JSON events. Connector produces it, `services/api/cmd/ingress` consumes it, and existing API inbox DTOs add `source=facebook` and message `status=received`. Deploy API migrations and compatible consumer before enabling the producer; stop ingestion before rollback. Unknown envelope versions are quarantined by hash, without payload. Future event schemas must preserve old readers or use a new versioned topic.

Connector uses its own PostgreSQL schema/history/revision schema. Provision a dedicated login with access only to `connector`; API's login must have no access to that schema. API authority lives in public `messaging_connections`, not connector tables. Runtime logins do not own migrations. The initial implementation can use separate databases, as the tests do. Do not grant the connector access to API tables.

Connector management uses a service credential, never a user JWT. `PUT /internal/connections` accepts an immutable ID/tenant/platform/Page/app identity and a strictly increasing config version (initial provisioning should use version 1). It changes only enabled state and credential reference. This is a trusted machine endpoint: expose only on a private network/reverse-proxy policy. Only `/webhooks/facebook` belongs on the public HTTPS listener. Tenant-authorized API provisioning and a user-facing connection setup flow remain follow-ups; operator provisioning is the current local path. A management token alone must never be delegated to a browser.

## Start locally

From repository root, configure private environment variables; values below are names, not credentials:

- `CONNECTOR_DATABASE_URL`: connector runtime PostgreSQL URL.
- `API_DATABASE_URL`: API runtime PostgreSQL URL for the separate consumer.
- `KAFKA_BROKERS`: comma-separated broker addresses reachable from the Go processes.
- `META_APP_ID`: the registered Meta app ID.
- `META_APP_SECRET_FILE`: private mounted file containing its app secret.
- `META_VERIFY_TOKEN_FILE`: private mounted file containing your verification token.
- `CONNECTOR_MANAGEMENT_TOKEN_FILE`: private mounted service token file.
- `CONNECTOR_LISTEN_ADDR`: defaults to `127.0.0.1:8091`.

Secret files must have at least 16 non-whitespace bytes. No Page token is needed by the webhook receiver. It will be needed for Page subscription or eventual fetch/send; store it in the environment's secret manager, never in a connection JSON, Kafka or chat. `credential_ref` is an opaque reference only; no Page-token resolver or provider API calls are implemented yet.

Apply migrations explicitly using migration-role URLs in a separate shell, then switch to runtime URLs. This delivery has not migrated your application database:

```sh
make migrate-apply-local svc=connector
make migrate-apply-local svc=api
```

The new connector target is explicit/host-operated until Compose role initialization and private connector URL wiring are implemented. Existing `all` apply/status continues to select API/AI; `all` hash/validate includes connector.

Create topic `hexta.ingress.received.v1` using your broker's admin tooling; producer auto-creation is disabled. Configure retention, access controls, partition count and replication for the environment. Connector writes with `acks=all`; it does not claim producer-idempotence or cross-system exactly-once.

Provision the API's trusted binding first (operator SQL through its migration/admin login, not connector code):

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

## Recovery and operational limits

- One callback supports up to 1 MiB and 1,000 items. All entries are visited. Invalid/malformed/unsupported items are quarantined alongside valid siblings, without raw payload retention. Malformed whole batches receive 503 for retry. Only text and echo events are published; attachments, read/delivery/reaction and other events carry safe quarantine classifications.
- Receipt hash deduplicates identical batches; source message identity deduplicates different batches and webhook/poll overlap. Source revision is empty unless the provider supplies one. No invented revisions or media metadata. Message content and IDs stay private Kafka/DB data; logs contain neither.
- Normalization happens before ACK, in the request; no post-ACK normalization queue exists. Persisting normalized outbox rows is the durable continuation point. A failed DB transaction/100,000 pending-item admission limit returns 503. Receipt and published/quarantine retention cleanup is not automated; monitor DB space and apply a reviewed retention policy.
- Outbox claims use `SKIP LOCKED`, a 30-second lease, fencing, 10-second publication timeout and exponential backoff capped at 300 seconds; after 12 attempts the event is quarantined. A crash after broker ACK but before marking published can duplicate publication. API dedup plus mutation commit together. A crash after API commit before offset commit replays without duplicating the message.
- Quarantine stores safe reason codes and receipt/app identity; unsupported raw content cannot be reprocessed from these records. Re-ingest from the provider only after implementing the missing mapping. Exhausted publication rows retain their normalized payload and can be deliberately requeued by an operator after diagnosis. Never replay with a new source identity to bypass dedup.
- Disable the API binding first, then update connector enabled=false and increment its version. API rechecks its current binding/tenant inside the transaction. Already committed inbox messages remain. Delayed stale-version events are handled without business mutation. Credential rotation increments both versions and reloads mounted app-secret files by restarting the connector explicitly; in-flight callbacks with the retired secret may fail and require provider retry. There is no dual-secret grace period yet.
- On a binding change, stage the API version first. There is no atomic configuration transaction across services: older delayed events may be ignored during rollout. Never change a connection's tenant in place; provision a new ID after disabling the old binding. Operator tooling must enforce the same invariant in the API mirror.
- API malformed envelopes/untrusted bindings are quarantined by hash. DB failures retry the current record before offset advancement. This single-record consumer trades throughput for simple commit safety; concurrent partition processing/metrics/alerting are follow-ups.

## Sync and outbound boundary

`SyncService` and durable `sync_jobs` implement page fetching, mode, bounded limits, checkpoint-after-durable-data, lease/fence, retry-after, bounded attempts and terminal cursor-expired recovery. A crash between page storage and checkpoint refetches that page and deduplicates by provider identity. Cancellation sets state=cancelled and increments fence; an already fetched page can persist but cannot advance the checkpoint. History and incremental reconciliation never open active turns or schedule AI. Fixture tests cover the framework; no runtime sync scheduler/management API is wired. Facebook advertises Fetch=false and returns unsupported. Empty/expired cursors must not silently restart a full history: create an explicit reconciliation job after provider-specific recovery is defined.

Outbound is an unsupported capability contract only. Before adding a sender, implement API-owned intents/outbox, current-state single-attempt permits, durable connector attempts and result events. Recheck tenant/control/assignee/context before provider dispatch. Permit revocation cannot undo an HTTP request already handed to the provider. Timeout after possible provider acceptance must be unknown, with reconciliation before resend. Provider accepted/delivered/read are separate states. AI candidates require API revalidation; this ingress slice schedules no AI and sends no real messages.

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
go test -race ./services/connector/... ./services/api/internal/bootstrap ./services/api/internal/core/service
make migrate-validate
pnpm --filter web lint
pnpm --filter web build
```

API tests create/drop random databases with Atlas; connector tests require the specifically named migrated disposable `connector_fixture`. The E2E test builds and starts a real connector subprocess, provisions through machine HTTP, sends signed duplicate callbacks while Kafka is unreachable, kills/restarts the connector against the broker, consumes Kafka into a random API DB and verifies tenant-authorized inbox reads. It creates the test topic. Run tests serially against a fresh isolated broker; the API consumer group is fixed. See the EPIC-008 session for exact results and NOT_RUN fault-injection/live checks.
