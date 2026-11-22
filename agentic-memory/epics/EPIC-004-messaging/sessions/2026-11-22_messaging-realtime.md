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

- DECIDED: server-sent events carrying only a hint `{thread_id, reason}`; the web refetches through the existing authorized list and detail endpoints, so permissions, paging and `expected_version` stay with those endpoints. No message content travels on the stream.
- IMPLEMENTED: `GET /api/v1/tenants/{tenantID}/conversations/stream` in the public API (`ConversationStreamController`, `ConversationStreamHub`). Events: `ready`, `changed`, `resync`, `reconnect`; 20 s heartbeat; closes after about 10 minutes so session and membership are rechecked on reconnect; at most 5 streams per user and tenant (429).
- IMPLEMENTED: fan-out over Redis pub/sub (`hexta:rt:tenant:{id}`, `cache.RealtimeBus`). Hints are best effort (CACHE-002): when the subscription fails the hub sends `resync` to every stream, and the web resyncs on every reconnect and when the tab becomes visible.
- IMPLEMENTED: hints are published by the projector only after the view row is applied (so a refetch sees the change) and by `DeliveryService` after a committed delivery result (sent/failed). The `ingress` process needs `REDIS_ADDR`; without it projection still works and clients only lose the push.
- IMPLEMENTED (web): `useConversationStream` (fetch-based SSE through the SDK for Bearer auth and token refresh, backoff, 300 ms coalescing), quiet reloads in `useInboxConversations` (first page only) and `useThread` (keeps older pages already loaded), connection status in the Desk header, vi/en strings.
- OPEN: browser notifications/sound, typing and presence, the read-model rebuild command.

## Verification and evidence

- `go vet ./...` and `go test ./...` in `services/api` PASS with `API_TEST_DATABASE_URL` on the local demo PostgreSQL; `API_TEST_REDIS_ADDR` against the demo Redis PASS for the pub/sub test; Swagger regenerated.
- New tests: hub (tenant isolation, non-blocking slow streams, per-user limit, resync on source failure), projector (publish only after apply), delivery (one hint per recorded outcome), SSE controller (ready, tenant-only hints, 401, 429), Redis bus.
- `pnpm --filter web lint` and `build` PASS; `node --test` for the new SSE frame parser PASS.
- NOT_RUN: rebuilt multi-process demo stack with two browsers, Caddy behaviour with `encode` on an event stream, 10-minute reconnect, Redis outage and recovery in the browser.
