---
epic: EPIC-001
session: 2026-11-14_audit-hardening
status: in-review
completed: false
date: 2026-11-14
---

# Audit hardening: sessions, tracing, pooling, ingress, tenant isolation, search, legacy AI, outbox

## Scope

Eight owner-requested fixes following the codebase audit, delivered as one commit each in [PR #68](https://github.com/TruongHoang2004/Hexta/pull/68). Primary epic EPIC-001 (audit findings). Also affects [EPIC-005](../../EPIC-005-kafka/README.md) (order outbox), [EPIC-007](../../EPIC-007-orders-workflows/README.md) (tenant isolation, confirmation event) and [EPIC-008](../../EPIC-008-integrations-deployment/README.md) (connector ingress lock).

## Starting evidence

Branch `fix/audit-hardening` from main `bd2d29e`. [Architecture audit](../../../reviews/2026-10-28_architecture-compliance_review.md) R6 (session-cache wiring) open. Unrelated pre-existing `packages/sdk/node_modules` changes excluded.

## Outcome and decisions

IMPLEMENTED (in review, not merged):

1. Auth repositories bound by interface. **DECIDED (owner, in-session):** session validity stays DB-authoritative; `ISessionRepository` binds the PostgreSQL repository and the unused `SessionCacheWrapper` binding is removed. Enabling the wrapper would conflict with SEC-001/CACHE-002/CACHE-009 and its cache-aside fill race can resurrect revoked sessions for up to 7 days. Resolves R6 once merged.
2. `trace_id` restored in error responses: `errors.Init` gets the service name and trace key; `Tracer()` is registered after `otelgin` and also writes the Gin context. Error `source` becomes `API_Service`.
3. Connection pools: GORM 50 open/25 idle/30m/5m; AI `RunStore` uses `psycopg_pool` (2–10, 5s checkout), closed after the durable worker stops and on failed startup. New dependency `psycopg-pool` 3.3.3.
4. Connector acceptance no longer takes `pg_advisory_xact_lock(710831)`. Receipt claim is row-level; events insert in key order. **DECIDED (owner, in-session):** the 100,000 pending-event limit is approximate (overshoot ≤ (pool − 1) × 1,000 items).
5. `RequireTenantMember` on tenant resource routes, delegating to `TenantService.AuthorizeMember` (ARC-002/HTTP-008) with an uncached tenant read. Composite `(tenant_id, id)` FKs for `order_items`/`product_variants`.
6. Elasticsearch/Qdrant moved to the `search-engine` Compose profile. Generated `products.search_vector` + GIN index; prefix full-text search. No live path calls `SearchProducts` after item 7.
7. Legacy `AIAgentService`, `infrastructure/gemini`, generation types, unmounted `GetDraft` and draft-cache binding removed; Swagger regenerated (draft DTO definitions removed).
8. Transactional outbox: `domain_events` written in `SalesService.Confirm`; fenced-lease relay to `hexta.orders.events.v1`, enabled only when `kafka.brokers` is set. At-least-once; no consumer yet.

## Verification and evidence

- `sh scripts/check-gofmt.sh`, `go vet` for api/shared/connector/test: pass.
- `go test -race -count=1 -p 1 ./services/api/... ./packages/shared/... ./services/connector/...` with disposable fixture PostgreSQL (`infrastructure/compose.connector-test.yml`) and its Kafka broker: 19 packages pass.
- `./scripts/local.sh --profile test run --build --rm ai-tests`: 109 passed, 0 skipped.
- `make migrate-validate` and desired/history sync at each schema commit: pass. FK upgrade: applies on consistent data; a mismatched row fails and rolls back.
- Outbox relay published to a real Kafka broker and the consumed envelope matched; event marked published.
- Regression checks confirmed new tests fail on old code (unregistered Tracer, missing Gin-context trace, global connector lock).
- NOT_RUN: `go test ./test/...` (needs running servers), live Meta, live Kafka consumers, browser acceptance.

## Remaining work and handoff

- Review and merge PR #68; then mark this session done and close audit R6.
- Before deploying, run the PR's FK pre-check queries; provision `hexta.orders.events.v1` and set `KAFKA_BROKERS` to enable the relay.
- Follow-ups: wire consultant candidate selection to indexed search; remove unused `config.AI` provider fields, `SessionCacheWrapper` source and `cache.NewSessionCache`; bound the ingress worker's DB pool; sessions are created with `Token: "temp"`.
