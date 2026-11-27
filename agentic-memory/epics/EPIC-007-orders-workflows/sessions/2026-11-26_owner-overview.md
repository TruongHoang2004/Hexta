---
epic: EPIC-007
session: 2026-11-26_owner-overview
status: in-review
completed: false
date: 2026-11-26
---

# Owner overview (PRD MVP 6, S-06)

## Scope

Implement FR-6.4, FR-6.5 and business rule 3 of [MVP 6](../../../../docs/specs/prd/mvp-6-exceptions-overview.md): an owner-only, read-only overview of orders and work. Cancellation/return cases (FR-6.1–6.3) are a separate parallel flow and are out of scope. The October design ([owner overview](../../../designs/2026-10-23_owner-overview_design.md), unmerged draft PR #59) was reused only where it fit current code.

## Starting evidence

Worktree branch `worktree-agent-a4ea89c0f2db8bc54` from `main` at `b2d886b`. No overview code existed on main.

## Outcome and decisions

IMPLEMENTED (not merged):

- API `GET /api/v1/tenants/{tenantID}/overview?from&to&employee_id`: owner identity (`tenants.owner_id`) only; members, non-members and suspended tenants get 403 before any business data is read. One read-only REPEATABLE READ snapshot. Files: `overview_service.go`, `overview_repository.go`, `overviewmodel`, `mapper/overview/{httpmap,persistencemap}`, `overview_controller.go`, `overview_dto.go`.
- Orders are grouped by stored status (`GROUP BY status`); the service fills a fixed ordered bucket list and folds unknown statuses into `other`. A new status bucket is added by appending to `overviewOrderStatuses` (no SQL or contract change). Order value = sum of `total_amount` of confirmed, processing and fulfilled orders, never labelled revenue.
- Work counts per pinned workflow version and current step (active, waiting, completed, unassigned ⊂ active) plus cohort totals.
- Stuck list: open (active/waiting) work that is unassigned or unchanged for more than `OverviewStuckAfter` = 48 h (based on `work_items.updated_at`), oldest first, capped at 50 rows with the full total.
- Filters: half-open creation interval for both orders and work (presets computed in the browser's local time); employee = order `employee_id` / work `responsible_id`.
- Migration `20261126500000_orders_tenant_created.sql`: index `orders (tenant_id, created_at)`.
- Web: new `modules/overview` and owner-only sidebar group; Work page opens `?item=<id>` for the "Open" link. No SDK change: domain endpoint wrappers live in app modules, as for orders and work.

DECIDED here (owner may revisit): 48 h threshold; value statuses; period applies to the stuck list too; members do not see the sidebar entry and get the access-denied state by URL.

## Verification and evidence

- `go test -race ./services/api/... ./packages/shared/...` with `API_TEST_DATABASE_URL` pointing at a disposable Docker PostgreSQL 17.7: PASS (includes DB overview, snapshot read-only and isolation tests).
- `sh scripts/check-gofmt.sh`, `go vet`, `make migrate-validate svc=api`, Atlas desired/history drift check: PASS.
- `pnpm --filter web lint`, `tsc --noEmit`, `pnpm --filter web build`, `node --test modules/overview/utils.test.mjs`: PASS. SDK lint/test/build: PASS (unchanged).
- Browser acceptance: NOT_RUN.

## Remaining work and handoff

Browser acceptance of S-06 states (empty tenant, empty period, member denied, responsive/dark). Integrator re-hashes `migrations/api/atlas.sum` at merge. When the cancellation/return flow lands, add its status or case bucket to `overviewOrderStatuses` and the tile list. Not done until merged into main.
