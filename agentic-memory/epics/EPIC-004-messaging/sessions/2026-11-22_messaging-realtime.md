---
epic: EPIC-004
session: 2026-11-22_messaging-realtime
status: in-progress
completed: false
date: 2026-11-22
---

# Realtime inbox updates

## Scope

Owner request: the web inbox must receive new messages and conversation changes without pressing Refresh. Built on the read/write split ([session](2026-11-22_messaging-cqrs-split.md)).

## Outcome and decisions

- DECIDED (owner correction): the stream pushes the message itself. A first version pushed content-free hints and made the UI refetch; the owner rejected that. When a customer writes while staff has that chat open, the message appears in the open chat at once, without a refetch.
- IMPLEMENTED: `GET /api/v1/tenants/{tenantID}/conversations/stream` (`ConversationStreamController`, `ConversationStreamHub`). Events: `message` (whole stored message in the API message shape, including body), `delivery` (message id, state, reason), `changed` (list refresh trigger from the projector), `resync`, `ready`, `reconnect`; 20 s heartbeat; closes after about 10 minutes so session and membership are rechecked; at most 5 streams per user and tenant (429).
- IMPLEMENTED: messages are published after their transaction commits, only for newly stored rows (idempotent replays and rejected commands push nothing): by `IngressService` for real Facebook/Telegram customers (in the `ingress` process) and by `MessagingCommandService` for simulator input and staff replies (in `messaging-command`, so a colleague's UI shows them too). `DeliveryService` pushes delivery updates. Fan-out is Redis pub/sub (`hexta:rt:tenant:{id}`, `cache.RealtimeBus`); message bodies cross Redis, which is internal and non-persistent.
- DECIDED: PostgreSQL stays the source of truth (CACHE-002). A failed publish never affects the command; when the Redis subscription fails the hub sends `resync`; the web also resyncs on every reconnect and when the tab becomes visible.
- IMPLEMENTED (web): `useConversationStream` (fetch-based SSE through the SDK for Bearer auth and token refresh, backoff) with `subscribe`. `useThread` appends a pushed message immediately (deduped by id), patches delivery state, and quietly reloads the turn because its version feeds `expected_version`. The list refreshes quietly (first page only) from a coalesced trigger. `MessageHistory` scrolls new messages into view when the reader is at the bottom, otherwise shows a "new messages" button; connection status in the Desk header; vi/en strings.
- OPEN: unread badges on list rows, browser notifications/sound, typing and presence, the read-model rebuild command.

## Verification and evidence

- `go vet ./...` and `go test ./...` in `services/api` PASS with `API_TEST_DATABASE_URL` on the local demo PostgreSQL; `API_TEST_REDIS_ADDR` against the demo Redis PASS for the pub/sub test; Swagger regenerated.
- New tests: hub (tenant isolation, non-blocking slow streams, per-user limit, resync on source failure), projector (publish only after apply), ingress (only the stored message is pushed), command service (one push per stored message, none for replay, conflict or non-member), delivery (one update per recorded outcome), SSE controller (frames, tenant isolation, 401, 429), Redis bus, and an end-to-end test: command service stores a customer message, Redis and the hub deliver it whole to an open stream.
- `pnpm --filter web lint` and `build` PASS; `node --test` for the new SSE frame parser PASS.
- NOT_RUN: rebuilt multi-process demo stack with two browsers, Caddy behaviour with `encode` on an event stream, 10-minute reconnect, Redis outage and recovery in the browser.
