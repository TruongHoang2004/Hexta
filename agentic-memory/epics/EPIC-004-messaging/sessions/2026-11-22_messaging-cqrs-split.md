---
epic: EPIC-004
session: 2026-11-22_messaging-cqrs-split
status: in-progress
completed: false
date: 2026-11-22
---

# Messaging read/write split (CQRS)

## Scope

Owner request: the API must not both read and write conversations. Three stacked phases on `refactor/messaging-cqrs-split`: (1) separate query and command code, (2) separate command process, (3) separate read model fed through the outbox and Kafka. URLs, DTOs, permissions, `expected_version`/409 and `request_key` idempotency are unchanged.

## Outcome and decisions

- IMPLEMENTED (phase 1): `MessagingQueryService/Controller/Store` (read-only snapshot transaction, lock-free `ReadScope`) and `MessagingCommandService/Controller`; shared helpers in `messaging_common.go`.
- IMPLEMENTED (phase 2): `cmd/messaging-command` serves only the seven write routes. The public API forwards them through `router.MessagingCommandGateway` to `messaging.command_url` (503 when unset) so the web keeps one base URL. Compose demo/prod run the new service; Caddy still routes only `api`.
- IMPLEMENTED (phase 3): `messaging_conversation_views` read model (Atlas migrations `20261122052820`, backfill `20261122052827`). Every transaction that changes a thread or its turn writes one `messaging.conversation.changed` outbox row with the final snapshot (`MessagingRepository.flushConversationEvents`, also on the ingress path). The relay publishes to `hexta.messaging.conversations.v1` keyed by thread; the projector (a new consumer group in the existing `ingress` process) upserts guarded by `(turn_version, thread_updated_at)`, so replays and out-of-order deliveries are safe. `GET /conversations` reads only the view.
- DECIDED: `GET /conversations/:id/messages` keeps reading the authoritative thread/turn/message tables. Its `turn.version` feeds `expected_version`, so it must not lag; the list may.
- Consistency: the list is eventually consistent (outbox poll ≈1 s plus Kafka). A conversation can be missing or stale for about a second after a write. `ConversationViewRepository.Rebuild(tenant)` repairs a view from the tables.
- OPEN: no operator command exposes `Rebuild`; web does not yet patch its list from the command response.

## Verification and evidence

- `go vet ./...` and `go test ./...` in `services/api` PASS with `API_TEST_DATABASE_URL` pointing at the local demo PostgreSQL (disposable Atlas-migrated databases), including new tests for event emission (one per thread and transaction, rolled back with the command), projector decode, idempotent/out-of-order projection vs rebuild, relay routing, gateway forwarding and 503, and fx composition of both processes. Phases 1–2 were first run without a database (DB tests skipped) and re-run with it after phase 3.
- `make migrate-validate svc=api` PASS; Swagger regeneration produced no diff.
- `docker compose config` resolves for demo and prod with the new service.
- NOT_RUN: live Kafka relay→projector round trip, browser acceptance, and a running multi-process stack.

## Remaining work and handoff

Run the demo stack with the new service and walk claim/handoff/reply in the web inbox; decide whether to expose a rebuild command and to patch the web list from command responses. Not complete until merged into main.
