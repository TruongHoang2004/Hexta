---
epic: EPIC-008
session: 2026-11-16_telegram-bot-connection
status: in-review
completed: false
date: 2026-11-16
---

# Telegram Bot platform connection (MVP v1.0) — implementation

## Scope

Implement [the Telegram Bot connection design](../../../designs/2026-11-16_telegram-bot-connection_design.md): owner Bot Token onboarding/rotation, private 1-1 inbound, outbound text transport and its UI. Chatbot, AI (`services/ai`, `proto/`) and Desk routing are out of scope and untouched.

## Starting evidence

Branch `feat/telegram-bot-connection` from `chore/docs-cleanup` @ `fb9e4c0` (Facebook connection PR #69 already on that history). Rules read: root AGENTS, rule router, API/shared/migrations/frontend/web AGENTS, Go conventions, boundaries, services, repositories, Atlas, verification, authorization, controller validation, errors, Fx, DTO, responses, client adapters, routing, Swagger, configuration, observability, mappers, delivery workflow, cross-service skill.

## Outcome and decisions

IMPLEMENTED on `feat/telegram-bot-connection` (PR open against `main`; not merged):

- **Shared:** `outbound.ErrorBotBlocked`, `ErrorChatNotFound`; `outbound.Platforms` lets intents target `telegram`. No new topics.
- **Schemas (Atlas-generated):** API `20261116020138_telegram_bot_connection` (`messaging_connections.account_username`); connector `20261116020151_telegram_bot_connection` (`platform` check adds `telegram`, `webhook_secret_ciphertext`, check that a secret implies a credential). Connector diff ran with host Atlas because the migrator container cannot start Docker dev databases (same as the Facebook delivery).
- **Connector:** `internal/plugin/telegram` (update parsing, constant-time secret check, Bot API client for getMe/setWebhook/deleteWebhook/sendMessage with sanitized classification); `POST /internal/telegram/bots/verify`; `PUT /internal/connections` for `platform=telegram` with `bot_token`; `POST /webhooks/telegram/{connectionId}`; outbound sender chosen per platform. Token and secret sealed AES-256-GCM, bound to the connection ID.
- **API:** owner-only `POST/GET/PATCH /api/v1/tenants/{tenantId}/channels/telegram/connections` (list is member-readable), 1 Bot = 1 tenant via the existing partial index, rotate requires the same bot ID (409), pause/disconnect cancel pending intents (shared update path with Facebook), Telegram-specific auth_expired text, Telegram ingress threads (`telegram:` key, "Telegram" name), simulator guards for all channel threads, thread DTO `connection_id`. Swagger regenerated; contract test covers the new routes and token absence.
- **Web:** Settings → Channels shows Facebook and Telegram sections; Telegram masked token input (password field, cleared on submit), status badges, rotate-token form, pause/resume/disconnect confirmations; generic expired banner (Facebook + Telegram) in Inbox and Settings; platform-aware placeholders and delivery reasons (`bot_blocked`, `chat_not_found`); composer locked when the bound connection is paused, expired or disconnected.

Design amendments are logged in the design's §7 and await owner acceptance.

## Verification

Isolated infrastructure: `infrastructure/compose.connector-test.yml` (Postgres 55439, Redpanda 59092). All Docker containers on the host disappeared mid-session (not caused by these commands); the disposable fixture was recreated, `connector_fixture` created and migrated fresh with `make migrate-apply-local svc=connector`. The earlier fixture also applied the new connector migration as an upgrade.

| Check | Command | Result |
| --- | --- | --- |
| Atlas integrity | `make migrate-validate` | PASS (api, ai, connector) |
| Desired vs history | `atlas migrate diff ci_schema_check --env api` / `--env connector` | PASS (synced, no files) |
| Go race tests | `go test -race -count=1 -p 1 ./services/connector/... ./services/api/... ./packages/shared/...` with `API_TEST_DATABASE_URL`, `CONNECTOR_TEST_DATABASE_URL`, `KAFKA_TEST_BROKERS` | PASS, including `TestConnectorTelegramConnectIngressSendAndResults` (real connector subprocess + mock Bot API: rejected token 400, 1-Bot-1-Tenant 409, rotate mismatch 409, wrong secret 403, group/supergroup dropped, voice placeholder, sent / bot_blocked / chat_not_found / unknown (dropped connection) / auth_expired → connection auth_expired, rotate recovery, disconnect) |
| Vet | `go vet ./services/api/... ./services/connector/... ./packages/shared/...` | PASS |
| Format | `sh scripts/check-gofmt.sh`; `gofmt -l services/connector packages/shared` | PASS |
| Web | `pnpm --filter web lint`; `pnpm --filter web build` | PASS |
| Whitespace | `git diff --check` | PASS |
| Live Telegram | real @BotFather token, public HTTPS webhook, live text/media, block/revoke | NOT_RUN |
| Browser acceptance | manual UI walkthrough | NOT_RUN |

## Unresolved work and handoff

- Owner review of the §7 amendments and the PR; merge is the owner's action.
- Live Telegram acceptance needs a public HTTPS origin for `TELEGRAM_WEBHOOK_BASE_URL` and a real bot.
- No periodic token health check: a revoked token is detected on the next send (401) only.
