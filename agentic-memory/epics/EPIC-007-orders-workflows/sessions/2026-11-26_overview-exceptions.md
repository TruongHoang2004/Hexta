---
epic: EPIC-007
session: 2026-11-26_overview-exceptions
status: in-review
completed: false
date: 2026-11-26
---

# Owner overview reads exception cases

## Scope

Connect the owner overview ([session](2026-11-26_owner-overview.md)) to cancellation/return cases ([session](2026-11-26_order-exception-cases.md)). The overview counted cancelled/returned from legacy order statuses, but cases never change order status; they live in `order_cases`. The approved S-06 wireframe shows "Đang huỷ / trả" and "Đã chốt huỷ / trả" tiles and lists return cases waiting for the owner in the stuck list.

## Starting evidence

Integration branch `integrate/mvp5-6` at `84d1ca8` (flows 1, 3, 4 and 5 merged). Not on main.

## Outcome and decisions

IMPLEMENTED (not merged), additive JSON on `GET /api/v1/tenants/{tenantID}/overview`:

- `exceptions`: `in_progress` (open + sent_back) and `finalized`, each `{total, cancellation, return}`, plus `cancelled_order_value`. Cases are counted through the order cohort (tenant, creation period, order `employee_id`, not soft-deleted) inside the existing REPEATABLE READ snapshot. Repository `CaseTotals` groups by (order status, case type, case status); the service derives the stages.
- `pending_cases {total, items[]}`: open cases (waiting for the owner to finalize or send back), same order cohort, oldest `updated_at` first, capped at 50 with the full total. Items carry `case_id, order_id, order_number, case_type, handler_id, waiting_since`. Sent-back cases wait for the handler and are not listed.
- Chosen: a separate `pending_cases` list instead of a new item kind inside `stuck.items`. Stuck items are work items (required `work_item_id`, workflow/step fields, reasons tied to the 48 h rule, `stuck.total` used by the empty-period check); mixing kinds would make those fields optional and change the meaning of the existing total for current clients. The web merges both lists into one table.
- Order value: orders with a **finalized cancellation case** are subtracted from `orders.order_value` (only orders in value statuses confirmed/processing/fulfilled). Return cases never reduce order value (no refunds in Hexta); open/sent-back cancellations still count. `orders.by_status` buckets keep stored totals; legacy `cancelled`/`returned` buckets are unchanged. **Implementation decision pending owner confirmation.**
- No migration: `order_cases_tenant_id_order_id_key` (unique on `tenant_id, order_id`) already serves the join, and one case per order means the join never duplicates orders.
- Web: tiles Draft, Confirmed, "Đang huỷ / trả", "Đã chốt huỷ / trả" (with cancellation/return split) and order value; legacy cancelled/returned buckets move to the "other statuses" line when non-zero. Pending cases appear after work rows in the stuck table ("Hồ sơ trả hàng #… · Chờ chủ chốt") linking to `/tenants/{id}/orders?order=<order_id>`; the Orders page gained an `?order=` deep link (Suspense-wrapped, dirty-edit guard, stale-response guard). vi/en strings added.

## Verification and evidence

All run on the integration worktree against the disposable `hexta-integ-pg` PostgreSQL (port 55439):

- `sh scripts/check-gofmt.sh`, `go build ./services/api/...`, `go vet ./services/api/... ./packages/shared/...`: PASS.
- `API_TEST_DATABASE_URL=… go test -race -count=1 ./services/api/... ./packages/shared/...`: PASS. New/updated: `TestOverviewCaseTotalsAndPendingCasesFollowTheOrderCohort` (repository, DB), `TestOverviewDatabase` (service, DB: period/employee/all-time/isolation with cases), `TestOverviewExceptionsSplitStagesAndCancelledValue`, `TestOverviewHTTPContract`, `docs` contract test (exceptions/pending_cases schema, no revenue/refund fields).
- `make swagger`: regenerated docs; diff limited to overview.
- `pnpm --filter web lint`, `pnpm --filter web build`, `npx tsc --noEmit`, `node --test modules/overview/utils.test.mjs`: PASS.
- Browser acceptance: NOT_RUN.

## Remaining work and handoff

Owner to confirm the order-value rule (finalized cancellation excluded; returns not). Browser check of the new tiles, the case rows and the `?order=` link. Not done until merged into main.
