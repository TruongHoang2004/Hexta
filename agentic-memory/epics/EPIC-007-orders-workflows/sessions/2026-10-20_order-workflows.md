---
epic: EPIC-007
session: 2026-10-20_order-workflows
status: in-progress
completed: false
date: 2026-10-20
---

# D4 configurable sequential work

## Scope

Owner directed moving straight to the next work after bounded manual sales testing. Implement the approved D4 workflow/work-item vertical slice from the active backlog. Preserve remaining D1–D3 manual acceptance gaps. [Technical design](../../../designs/2026-10-20_order-workflows_design.md).

## Starting evidence

Started at 3b9d182 on codex/thesis-content; created codex/order-workflows without resetting existing uncommitted reconciliation and screenshot records. Inspected actual routes, sales service/repository, browser checklist, relevant architecture modules and frontend instructions. No pre-existing workflow routes/tables/UI. No fresh remote PR/CI state asserted; this session is local implementation.

## Outcome and decisions

Added repository-layer models in repository/workflowmodel (with root aliases): definitions, immutable versions/ordered steps, linked work and transactional events. Owner publishes 1–30 user-named steps with direct/acceptance policy; a new publication does not change existing work. Owner explicitly creates work from a confirmed order with an initial current member. Current responsible actor hands to another current member; acceptance leaves old responsibility/current step intact until nominated recipient accepts. Owner reassigns/unassigns without advancing, cancelling pending nominations. Current responsible actor completes only the final step. No stock/payment/message/order mutations.

New Gin DTOs/controllers, Fx wiring and ten authenticated tenant routes expose the slice. Expected versions and row locks fence transitions; creation keys/hashes prevent repeated creates with the same request. Work detail returns latest 100 events; earlier records retained. Browser CORS allows Idempotency-Key for the configured origin. Typed web services reuse SDK transport; Work/Workflows pages and sidebar support configuration, work creation/detail, pending acceptance, reassignment and history. Confirmed orders link to Work. Actor/tenant/member checks remain server-side.

Atlas generated 20261020075908_order_workflows.sql with additive tables, tenant/reference composite FKs, state checks and transition uniqueness. Existing order identity gets an additional tenant+id unique index; business data retained. The canonical make migrate-diff command cannot reach its configured Docker hostname postgres from the host. Used Atlas diff against the exact replayed existing SQL plus reviewed additive target on a disposable postgres:17.7-alpine3.23 database; no unrelated GORM baseline drift accepted. First generation corrected an existing index-name collision before producing the final migration. A dependency-free repository model subpackage avoids an existing common -> DTO -> repository logger import cycle.

## Verification and evidence

- GOCACHE=/private/tmp/hexta-go-build go test -race ./services/api/... ./packages/shared/... passed for implementation code.
- Go vet for API/shared passed.
- Seven workflow scenarios passed with real PostgreSQL and race detection using isolated schemas: pinning/acceptance/direct/final completion, membership/reassignment, create replay/boundaries, publication validation, accept-versus-assign race, concurrent same-key create, and actual Atlas schema/FK checks. The migrated-schema test rolls back its test data.
- Actual migrated-schema check exercised publish -> create -> nominate -> accept -> complete, rejected cross-tenant references, rejected a nonexistent pinned step and duplicate event version.
- Isolated HTTP test passed malformed key/reference/policy/version rejection, owner/recipient authorization, handoff/accept/complete and no internal request-key/digest exposure. CORS regression passed.
- pnpm --filter web lint and pnpm --filter web exec tsc --noEmit passed. Initial lint violations were fixed using the established async effect loading pattern.
- GOCACHE=/private/tmp/hexta-go-build make swagger succeeded. Regenerated JSON preserves every existing path/definition semantically and adds workflow contracts. Initial unconfigured host cache failure was resolved with the temporary GOCACHE.
- atlas migrate validate --dir file://migrations/api passed. Seven migrations applied to an empty disposable PostgreSQL database. scripts/local.sh up --build -d --wait api web succeeded, including API/Next.js production builds and existing demo upgrade; API/web/PostgreSQL/Redis healthy and init/migrate exit 0.
- Disposable workflow PostgreSQL test container stopped/automatically removed after passing checks; demo stack/volume retained.
- D4 browser cases NOT RUN: delegated tester received "Browser is not available: iab" for tab listing/creation; browser inventory was empty. Native Chrome was listed, but getApp returned no usable state and was interrupted after approximately 302 seconds. No D4 UI mutation, account/invitation/privilege change or screenshot occurred. Earlier sales evidence predates D4 and is not reused as workflow proof.

## Remaining work and handoff

Browser verification/evidence remains blocked by unavailable computer-use browser surfaces in this session. Use the running local app and D4 checklist when browser access is restored. Review local diff and deliver/merge separately; no commit/PR/merge or issue lifecycle changes in this session. Parent epic remains in-progress, completed false. Earlier approval/invitation/isolation/stale-edit manual gaps are not completed by D4 tests. Cancellation/returns, overview D5, messaging, AI/providers and integrations remain separate.

## Commit preparation follow-up

Owner explicitly requested committing the D4 implementation. Prepare a scoped feat(workflows) commit including code, tests, migration, generated Swagger and D4-only documentation. Preserve earlier reconciliation and sales screenshot changes outside the commit. Existing automated results remain the verification evidence; no new application code changes or test rerun during commit preparation. Browser verification, PR and merge remain open; status stays in-progress/completed false.
