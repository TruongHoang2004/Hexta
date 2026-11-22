---
epic: EPIC-008
session: 2026-11-22_channel-e2e-mock-platform
status: in-progress
completed: false
date: 2026-11-22
---

# Channel send/receive end to end with a mock provider platform

## Scope

Owner goal for 2026-11-22: finish the Facebook and Telegram send/receive flow and build a mock platform for later performance tests (T1–T5 in [08-operations-security.md](../../../designs/system/08-operations-security.md#load-capacity-targets-target-owner-set-2026-11-18)).

Owner decisions:
- Local demo stack plus a live Telegram test through a cloudflared quick tunnel. The owner runs the tunnel.
- Facebook live is deferred because the Meta app is not ready.
- The mock is a Go service plus a load generator.

## Starting evidence

- Branch `feat/channel-mock-platform` from `main` @ `bdfeb77`. Connector Facebook (#69) and Telegram (#71) were already merged, verified only against in-test mocks.
- The demo stack used the placeholder `facebook-disabled` app and had Telegram disabled: no `TELEGRAM_WEBHOOK_BASE_URL` and no provider to message.
- Every relay hop waited up to about 1 s, because kafka-go's default `BatchTimeout` was in effect and the connector leased one event per 1 s tick.

## Outcome and decisions

IMPLEMENTED (local, unmerged):

- **`test/mockplatform` + `test/cmd/mockplatform`** (stdlib only; `test` module):
  - Simulated Graph API: OAuth dialog, code and long-lived token exchange, `me/accounts`, subscribe/unsubscribe, Send API.
  - Simulated Bot API: `getMe`, `setWebhook`, `deleteWebhook`, `sendMessage`.
  - Customer simulator: signed Facebook webhooks and Telegram updates with the registered secret, plus an echo after each Facebook send.
  - A chat page for manual runs.
  - Send fault injection: latency, either global or per Page/bot, and error kinds mapped to the connector's classifications.
  - Messages and counters.
  - Page tokens are derived from the app secret, so they survive mock restarts.
- **`test/cmd/loadgen`:**
  - Sends Facebook or Telegram webhooks straight to the connector at an open-loop fixed rate, with Telegram IDs unique per run.
  - With a read-only API DB session, polls during the run and reports lost and duplicate messages and webhook → visible latency, precise to the poll interval.
  - The API stores the provider timestamp in `messaging_messages.created_at`, not its commit time, so a commit-time column would be needed for long T1 runs.
  - The existing `pgx/v5 v5.6.0`, already used by the connector, was added to `test/go.mod`.
  - Outbound load (T3) is not implemented.
- **Demo stack:**
  - `infrastructure/compose.mock.yml` overlay, enabled with `scripts/local.sh --mock`.
  - Connector `TELEGRAM_WEBHOOK_BASE_URL` comes from `.env.local` (live mode).
  - A `webhook-proxy` (Caddy 2.8, `127.0.0.1:58092`) exposes only `/webhooks/*` for tunnels.
  - `local/api.yaml` reads `META_APP_ID`/`META_OAUTH_DIALOG_URL`, with the placeholder default kept in compose.
  - Image: `test/Dockerfile`.
- **Latency fix:**
  - `kafkatransport` writers set `BatchTimeout` to 5 ms. Shared consumers affected: connector relays, API outbox relay, ingress result path.
  - The connector ingress relay drains until empty or a publish fails.
  - Batch publishing and per-connection parallel sends remain proposals.
- **Docs:** demo runbook section "Channels: mock providers and live Telegram", `test/README.md`, and the design 08 blocker line.

## Verification and evidence

| Check | Result |
| --- | --- |
| `go vet` + `go test -race ./test/mockplatform/...` | PASS. Covers OAuth single-use code, Page token auth, echo signature, customer webhooks (Meta signature, Telegram secret header), webhook lifecycle, faults → Graph codes / Bot API statuses, per-account latency. |
| `docker compose … config --quiet`, base and `+ compose.mock.yml` (dummy env) | PASS |
| Demo stack `./scripts/local.sh --mock up --build -d --wait` | PASS: all services healthy |
| API E2E, Telegram (scratch script; throwaway account `mock-e2e-*@example.test` and tenant "Mock E2E", approved through a temporary operator grant that was revoked right after) | PASS. Connect bot; text and photo visible in Desk after 0.2–0.65 s (photo → `unsupported_media`); claim + reply → `sent` in ~0.5 s and received by the mock; `blocked` → `failed/bot_blocked`; 5xx → `unknown/transport`. |
| API E2E, Facebook | PASS. OAuth start → mock consent → callback offers 3 Pages → connect; customer text visible after 0.65 s; reply `sent` in ~0.56 s. The echo was delivered and created no duplicate; connector events were all published and the API quarantine count was 0. |
| loadgen Telegram 20 msg/s × 30 s, 50 chats | 600/600 HTTP 200, webhook p95 ~10 ms; 600 stored, 0 lost, 0 duplicates; visible p95 ~1.0 s |
| loadgen Facebook 50 msg/s × 20 s, 100 PSIDs | 1000/1000 HTTP 200, webhook p95 ~6 ms; 1000 stored, 0 lost, 0 duplicates; visible p95 ~0.74 s |
| `sh scripts/check-gofmt.sh`; `go vet` connector/shared/test; `git diff --check` | PASS |
| `go test -race -count=1 -p 1 ./services/connector/... ./services/api/... ./packages/shared/...` against `compose.connector-test.yml` (Kafka remapped to 59192 to avoid the demo stack; fresh migrated `connector_fixture`) | PASS, 22 packages |
| Before/after latency on the same stack | NOT_RUN for "before"; the design records the earlier ~1 s per synchronous write |
| Live Telegram via tunnel | NOT_RUN at the time of writing; needs the owner to run cloudflared and message the bot |
| Live Facebook; T1–T5 on the two VPS hosts | NOT_RUN |

Docker Desktop numbers are not VPS capacity measurements.

## Remaining work and handoff

- Owner: run the live Telegram test (runbook steps), then record PASS/FAIL here.
- Open a PR for this branch; delivery is not done before merge into main.
- Follow-ups:
  - Outbound load scenario (T3; needs multi-tenant seeding).
  - API commit timestamp for long T1 runs.
  - The 1 s idle tick bounds visible latency.
  - Batched relays.
- PR #90 (Google login demo fix) is separate.
