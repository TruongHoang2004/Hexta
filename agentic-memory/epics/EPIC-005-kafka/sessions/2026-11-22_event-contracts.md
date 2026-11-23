---
epic: EPIC-005
session: 2026-11-22_event-contracts
status: in-review
completed: false
date: 2026-11-22
---

# Event contracts in code

## Scope

Owner request: the chat pipeline must accept any number of connections of any type per tenant, with normalized events. Owner decisions (2026-11-22): contracts in `packages/shared`, one file per event, clear schema names; each topic carries one event type; messages are text-only, so one message event. Owner correction the same day: keep the existing topic names and the order/conversation envelopes — only the code-level schema organization was requested. Related: EPIC-004 (messaging), EPIC-008 (connector). The connection-management API/UI unification (one `connections` resource) is a separate, planned follow-up.

## Starting evidence

Worktree `.claude/worktrees/event-contracts`, branch `refactor/event-contracts` from `origin/main` `0455153` (after #91 CQRS and #92 mock platform). Before: `packages/shared/ingress.Event` (one struct for message/echo/media), `packages/shared/outbound` (intent, result and connection health in one file), API-local `DomainEventEnvelope` + `OrderConfirmed`/`ConversationChanged`; `hexta.orders.events.v1` and `hexta.messaging.conversations.v1` used envelopes with `event_type`.

## Outcome and decisions

- First revision (commit `e96d592`) renamed all topics to `hexta.<domain>.<event>.v1` and flattened the envelope topics; the owner rejected the renames. Follow-up commit restores the six main topic names, the `events.Envelope` wire format on `hexta.orders.events.v1`/`hexta.messaging.conversations.v1`, Compose provisioning and the connector topic check (the added migration was removed).
- Kept: `packages/shared/events` with `meta.go`, `envelope.go`, `platform.go`, `messaging/{message_received,conversation_changed}.go`, `delivery/{send_requested,send_completed}.go`, `connection/health_changed.go`, `order/order_confirmed.go`, `catalog/` (event → single topic, outbox event type, enveloped flag). Each event has `Topic`, `Validate`, `Key` and its own test; a test keeps Compose provisioning equal to the catalog.
- `MessageReceived` replaces `ingress.Event` (text, or no text plus `unsupported_media_kinds`; `direction=outbound` marks echoes). The three outbound/health events embed `events.Meta`. The API outbox relay routes by catalog through one multi-topic writer; enveloped events keep aggregate keys.
- Removed `packages/shared/ingress`, `packages/shared/outbound`, API `types/domain_events.go`, default-topic `kafkatransport` constructors.

## Verification and evidence

- `go vet` clean for shared, api, connector and test modules; gofmt clean; `git diff --check` PASS.
- Isolated fixture (`compose.connector-test.yml`, project `hexta-event-contracts-fixture`, ports 55539/59192, removed after each run; the running demo stack was not touched), after the owner correction: `go test -race -count=1 -p 1 -v ./services/connector/... ./services/api/... ./packages/shared/...` → 176 PASS, 1 SKIP (live Telegram), 1 FAIL caused by a test-setup edit (missing outbox insert) — fixed, then the outbox/projector/sales subset reran PASS including `TestOutboxRelayPublishesCommittedEventToKafka`. Includes connector → Kafka → API end-to-end for Facebook and Telegram.
- `atlas migrate validate --env connector` PASS; infrastructure and migrations identical to main.
- NOT_RUN: `test/e2e` (needs a running API on :9090), live Meta/Telegram.

## Remaining work and handoff

Review and merge. No topic, schema or migration change is needed on deployed stacks; producers and consumers of the four flat topics must be deployed together because their payloads changed. Next: unified connection management (EPIC-008 session 2026-11-22_unified-connections).
