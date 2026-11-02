---
epic: EPIC-004
session: 2026-11-02_messaging-inbox
status: in-progress
completed: false
date: 2026-11-02
---

# Messaging inbox: simulator and manual handling

## Scope and starting evidence

Owner requested the earliest possible messaging-management flow. Started from clean main `e466625` and created `codex/messaging-inbox`. EPIC-004 had no implementation session; historical AI order types were not reused as conversation replies. PR #59 remains separate draft/unmerged; PR #60 merge was checked in the preceding status discussion. No GitHub lifecycle mutation or provider access in this session.

## Outcome and decisions

IMPLEMENTED locally: tenant-qualified inbox route `/tenants/{tenantId}/inbox`, sidebar entry, filtered cursor-paginated conversations, bounded message history, explicitly simulated customer input, atomic self-claim/take-control, owner reassignment/unassignment and assigned-sale simulated manual replies. PostgreSQL owns threads/independent turns/immutable messages. Core values, ORM rows and HTTP DTOs are separate, with pure aggregate persistence/HTTP mappers. Bootstrap wires the repository port/service/controller; generated Swagger and route contracts cover all six endpoints.

Transactions recheck active tenant/current membership; owner viewing does not authorize replying unless assigned. Thread locking serializes incoming activity, claim/assignment and simulated reply. Resource/context versions reject stale replies; UUID request keys plus request hashes prevent duplicate retries and changed-key-content reuse. Shared audit writes commit with changes; injected audit failure rolls back thread, turn and message creation. UUIDv7 cursor order is creation order, not latest-activity ranking. No cache/network/model work inside transactions. Source/transport status is explicitly simulator/simulated, without provider accepted/delivered claims.

The UI retains text on failure and retries an uncertain reply using the original key/payload even after refresh. Conversation/tenant component boundaries isolate drafts; switching conversations does not promise persisted unsent drafts. Updates require explicit refresh. All repository/UI artifacts are English as required.

Design and route/schema details: [design](../../../designs/2026-11-02_messaging-inbox_design.md).

## Verification and evidence

- Atlas migration generated through root `make migrate-diff svc=api name=messaging_inbox`; reviewed additive SQL and tenant-qualified thread/turn/message FKs, one-open-turn partial uniqueness and request-key constraints.
- `make migrate-validate`: API and AI histories passed.
- `atlas schema diff --from file://migrations/api --to file://migrations/schema/api.sql --dev-url <disposable DB>`: schemas synced.
- Fresh schema applied by Atlas-backed test fixtures; separate prior-history DB with an existing tenant upgraded to latest and retained the fixture plus new messaging tables. All test databases lived on a separate temporary PostgreSQL container; no application business migration was applied.
- Seven new service tests passed with PostgreSQL and `-race`: authority/stale context/reply replay, isolation/revocation/assignment, concurrent claims, concurrent input replay, filters/history pagination, audit rollback, input bounds.
- HTTP flow/validation test passed with PostgreSQL; DI graph and generated Swagger/real-route contract checks passed. HTTP tests include missing actor, invalid query/path/body/size and no internal request metadata in successful responses.
- Final `go test -race ./services/api/... ./packages/shared/...` with explicit disposable PostgreSQL passed; `go vet ./services/api/... ./packages/shared/...` passed. Earlier full run exposed the newly required controller argument in the docs route fixture; corrected, expanded route contracts, then final full suite passed.
- `pnpm --filter web lint`, `pnpm --filter web exec tsc --noEmit` and SDK strict check passed. Web production build succeeded using `pnpm --filter web exec next build --webpack`; default Turbopack build failed because this execution environment denied its internal port binding (initial attempt also could not fetch Google Fonts). This does not claim default Docker/Turbopack build success.
- No browser/mobile/two-profile manual acceptance, real channel sends, AI evaluation, Kafka integration, deployment or merge verification for this slice.

## Remaining work and handoff

This is local implementation, not delivered/merged completion. Review and merge remain open. Apply the explicit API migration job and rebuild API/web before opening the new inbox against an application DB. Test with two real signed-in profiles and review mobile behavior. Do not reuse simulated messages as proof of live customer sending.

Preserved approved follow-ups: automatic 24-hour expiry and later turns, escalation/resolve, explicit return-to-system, scripts and independent grants, AI candidate/proposal pipeline, durable real outbound intents/attempts/reconciliation, Kafka/SSE, real connector capabilities and thesis evaluation. Needs-human filter is forward-compatible, but this slice does not expose an escalation command. The manual simulator runs independently of Kafka to unblock the first demo; Kafka is not claimed implemented or removed from requirements.

Unrelated configuration, local development tooling and tracked SDK dependency-link changes appeared concurrently after the clean start. They were not authored, reverted, staged or included as messaging work. No broad formatting of those files or commit performed.

Final execution note: host API/web listeners appeared on ports 8080/3000 during concurrent local-tooling work. They were not restarted or modified by this session; no claim that their application DB has this migration. The migration checks used the separate disposable test server.
