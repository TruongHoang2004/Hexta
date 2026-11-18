---
epic: EPIC-008
session: 2026-11-15_facebook-platform-connection
status: in-review
completed: false
date: 2026-11-15
---

# Facebook platform connection (MVP v1.0) — implementation

## Scope

Implement [the Facebook platform connection design](../../../designs/2026-11-14_facebook-platform-connection_design.md): owner OAuth/Page management, unsupported-media placeholders, and single-attempt outbound text delivery. Conversation routing, AI and Desk orchestration stay out of scope. Cross-epic: extends the EPIC-004 inbox send path.

## Starting evidence

Branch `feat/facebook-platform-connection` from `fix/audit-hardening` @ `b417ed3` (PR #68 in review). Baseline: connector ingress from PR #66. Unrelated uncommitted skill/memory edits from the 2026-11-14 PRD/design-skills session were preserved, not modified.

## Outcome and decisions

IMPLEMENTED (committed on `feat/facebook-platform-connection`, [PR #69](https://github.com/TruongHoang2004/Hexta/pull/69) open against `main`; not merged):

- **Shared contracts:** `packages/shared/outbound` (intent/result/health events with validation); `ingress` adds `unsupported_media` and media kinds; `kafkatransport.NewTopicReader`.
- **Schemas (Atlas-generated):** API `20261114164249_facebook_platform_connection` + data backfill `20261114164302_backfill_connection_status` (disabled → `paused`); connector `20261114164319_facebook_platform_connection`.
- **Connector:** Facebook implementation moved to `services/connector/internal/plugin/facebook` (owner request; the `plugin` package keeps the generic registry/contracts). Meta Graph client (OAuth code → long-lived user token → Page tokens; subscribe/unsubscribe; Send API) with sanitized error classification; AES-256-GCM sealing bound to row identity (`CONNECTOR_CREDENTIAL_KEY_FILE`); `POST /internal/facebook/oauth/exchange`; `PUT /internal/connections` returns JSON and handles session credentials, disconnect wipe/unsubscribe and idempotent replay; intent consumer with durable attempt claim (redelivered `dispatching` → `unknown`, never resent); result/health outbox relay. Attachments map to placeholder kinds (sticker detected); unknown kinds stay quarantined.
- **API:** `ChannelService` (owner-only start/connect/pause/resume/disconnect, member list), connector client adapter, `SendChannel` (assigned sale, active connection, intent + attempt + outbox row in one transaction), outbox relay routes `outbound.intent` raw to the intent topic, ingress process also consumes result and health topics, ingress stores media placeholders and binds threads to connection/recipient. Swagger regenerated.
- **Web:** Channels settings page (status badges, connect/reconnect, pause/resume/disconnect with confirmation), `/channels/facebook/callback` Page picker with availability, Inbox expired-authorization banner, media placeholder, delivery status/reason, real Facebook send with stable `client_message_id` on uncertain retry.

Deviations from the design. They are folded into the design text, with the log in its §7. The owner accepted them on 2026-11-15.

1. **OAuth state** is a random 256-bit single-use value stored as a SHA-256 hash in new table `channel_oauth_states`, bound to tenant/owner/redirect, instead of a signed stateless token. SEC-004 requires consumption, which a signature alone cannot provide. The same row binds the connector `oauth_session_id` and granted Pages to the owner, so another tenant cannot use a leaked session ID.
2. **Thread channel binding:** `messaging_threads` gains `connection_id`/`recipient_id`, because the existing thread key is a hash and the outbound recipient PSID was otherwise unrecoverable. Pre-existing Facebook threads are bound on their next inbound message.
3. **Inbox routes:** the existing routes are `/conversations` and `/turns/...`, not `/inbox/...`. Real send is `POST /api/v1/tenants/{tenantID}/turns/{id}/messages` and keeps `expected_version`/`context_revision` in addition to `client_message_id`/`text`, preserving the stale-context guard. Message `status` gains `channel`; delivery is tracked in `delivery_status`.
4. **Additional error categories:** `expired` (intent TTL 10 min) and `connection_inactive` (version/enabled mismatch at dispatch).
5. **Connector versions** must be strictly increasing (gaps allowed), not exactly +1, so an API retry after a failed connector sync converges.
6. `FacebookOAuthPagesResponse` adds `tenant_id`, since the fixed callback URL carries no tenant.
7. Text limit is 2,000 characters (Messenger limit), not the simulator's 4,000.

## Verification and evidence

Run on 2026-11-15 against disposable `infrastructure/compose.connector-test.yml` (Postgres :55439, Kafka :59092) and Atlas-migrated random/fixture databases:

- `go test -race -p 1 ./services/connector/... ./services/api/... ./packages/shared/...` — PASS, including `TestConnectorFacebookToInbox` (existing E2E, updated for the new key file and 200 JSON response).
- New `TestConnectorFacebookOAuthSendAndDeliveryResults` — PASS: owner OAuth through the real connector binary against a mock Graph server, subscription, signed webhooks, claim, send, outbox → Kafka → connector → Graph → result → inbox. Verified `sent`, `failed/window_24h`, exactly one provider call per intent, connection stays active.
- Unit/DB tests: OAuth state owner/single-use/expiry/denied, 1-Page-1-tenant conflict and handover after disconnect, pause cancelling pending intents, idempotent repeat-sync after connector failure, send authorization/idempotency/hash conflict, result dedup (no regression after a later `sent`), `auth_expired` surfacing and stale health ignored after reconnect, media placeholder ingress, Graph error classification and secret-free errors, sealed credentials, interrupted attempt → `unknown`.
- `TestOutboxRelayPublishesCommittedEventToKafka` (pre-existing) failed whenever it ran after the new E2E in one process: kafka-go writers share `DefaultTransport`, whose cached topic metadata lacked the test's new topic. Fixed test-only by giving that writer its own `kafka.Transport`; the bootstrap package then passed twice in a row. Production relays publish to pre-provisioned topics.
- `sh scripts/check-gofmt.sh` (after staging the package move), `go vet` (api/connector/shared), `make migrate-validate` (api/ai/connector), `make swagger` + `docs` contract test — PASS.
- `pnpm --filter web lint`, `pnpm --filter web build` — PASS.

NOT_RUN:


- Live Meta acceptance (real OAuth, Page subscription, webhooks, 24h window, token revocation): requires App Review / app access.
- Browser acceptance of the settings, callback and Inbox flows (keyboard, responsive, dark mode).
- Fresh-vs-upgrade migration check on a populated database; `make migrate-diff svc=connector` fails because the migrator container lacks `CONNECTOR_DEV_URL` (pre-existing gap), so the connector migration was generated with host `atlas migrate diff --env connector` from the same root `atlas.hcl`.

## Remaining work and handoff

- Known limits: if a connector sync fails after the API commits a version bump, connector events with the older version are ignored as stale until the owner repeats the action. Outbox rows that exhaust 12 attempts leave intents `pending` and messages `sending` (no sweeper yet). `hexta.connector.health.v1` is emitted only on send-time `auth_expired`; no periodic credential check exists. Rotating `CONNECTOR_CREDENTIAL_KEY_FILE` requires owners to reconnect Pages.
- Operators must provision topics `hexta.outbound.intent.v1`, `hexta.outbound.result.v1` and `hexta.connector.health.v1`, and configure `CONNECTOR_URL`, `CONNECTOR_MANAGEMENT_TOKEN`, `META_APP_ID` for the API.
- Next: browser acceptance; live Meta acceptance once app access allows; merge.
