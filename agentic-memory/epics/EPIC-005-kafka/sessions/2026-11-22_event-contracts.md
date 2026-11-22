---
epic: EPIC-005
session: 2026-11-22_event-contracts
status: in-review
completed: false
date: 2026-11-22
---

# Per-event contracts and topics

## Scope

Owner request: the chat pipeline must accept any number of connections of any type per tenant, with normalized events. Owner decisions (2026-11-22): one Kafka topic per event type; contracts in `packages/shared`, one file per event; messages are text-only, so one message event; pre-production, so no v1 compatibility. Related: EPIC-004 (messaging), EPIC-008 (connector). The connection-management API/UI unification (one `connections` resource) is a separate, planned follow-up.

## Starting evidence

Worktree `.claude/worktrees/event-contracts`, branch `refactor/event-contracts` from `origin/main` `0455153` (after #91 CQRS and #92 mock platform). Before: `packages/shared/ingress.Event` (one struct for message/echo/media), `packages/shared/outbound` (intent, result and connection health in one file), API-local `DomainEventEnvelope` + `OrderConfirmed`/`ConversationChanged`; `hexta.orders.events.v1` and `hexta.messaging.conversations.v1` used envelopes with `event_type`.

## Outcome and decisions

- New `packages/shared/events`: `meta.go` (`Meta`, `Hash`, UUID check), `platform.go`, `messaging/{message_received,conversation_changed}.go`, `delivery/{send_requested,send_completed}.go`, `connection/health_changed.go`, `order/order_confirmed.go`, `catalog/` (topic list + decoders). Each event has `Topic`, `Validate`, `Key` and its own test. Topics and keys: see [06-kafka-events](../../../designs/system/06-kafka-events.md#event-contracts-one-event-one-file-one-topic--2026-11-22-owner-decision-implemented-in-branch-refactorevent-contracts).
- `MessageReceived` keeps text or, with no text, `unsupported_media_kinds` so the existing inbox media placeholder keeps working; `direction=outbound` marks echoes. `event_type`/`attachments` removed. `Validate` now enforces content rules previously checked only in the API.
- API outbox: `domain_events.event_type` is the destination topic; the relay looks it up in the catalog, validates the payload, keys by `Key()` and publishes the raw event through one `kafkatransport.MultiTopicPublisher`. Unknown topics or invalid payloads are never published (row retried). Producers set `Meta.EventID` = outbox row ID.
- Removed `packages/shared/ingress`, `packages/shared/outbound`, API `types/domain_events.go`, the default-topic `kafkatransport.NewPublisher/NewReader`.
- Connector schema: `outbound_events.topic` check uses the new names; Atlas migration `20261122070740_event_topic_names` renames pending rows first. Pending API outbox rows with pre-release event types are not migrated (dev data only; they stay unpublished).
- Compose demo/prod provision the six new topics; a catalog test fails if Compose and catalog diverge. Docs: `packages/shared/AGENTS.md`, connector README, system design 06.

## Verification and evidence

- `go vet` clean for shared, api, connector and test modules; gofmt clean; `git diff --check` PASS.
- Isolated fixture (`compose.connector-test.yml`, project `hexta-event-contracts-fixture`, ports 55539/59192, removed after the run; the running demo stack was not touched): connector DB migrated fresh with Atlas; `go test -race -count=1 -p 1 -v ./services/connector/... ./services/api/... ./packages/shared/...` → exit 0, 176 PASS, 1 SKIP (`TestLiveTelegramWebhookAcceptance`, needs live token). Includes connector → Kafka → API end-to-end for Facebook and Telegram, outbound send/result and the outbox → Kafka relay.
- `atlas migrate validate --env connector` PASS; desired schema and history in sync.
- Compose files: `docker compose config --no-interpolate -q` PASS for demo/prod; interpolated config and a live stack run NOT_RUN.
- NOT_RUN: `test/e2e` (needs a running API on :9090), upgrade migration on a database holding old-topic rows, live Meta/Telegram.

## Remaining work and handoff

Review and merge. After merge, recreate local Kafka topics (`make` demo stack re-runs provisioning) and apply the connector migration. Next: Phase 1 — unify connection management (`/connections` list/update, single Channels list, "add connection → choose type").
