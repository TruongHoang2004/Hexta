---
epic: EPIC-007
session: 2026-11-26_order-exception-cases
status: in-review
completed: false
date: 2026-11-26
---

# Cancellation and return cases on confirmed orders (MVP 6, S-10 additions)

## Scope

FR-6.1–6.3 of [MVP 6](../../../../docs/specs/prd/mvp-6-exceptions-overview.md): open a cancellation or return case on a confirmed order, bound return quantities by the original line, let the assigned handler record progress and only the owner send back or finalize, and keep history. Owner decisions: [Orders and flexible workflows](../../../designs/system/01-product-decisions.md). Overview (FR-6.4/6.5) is out of scope for this session.

## Starting evidence

Branch `worktree-agent-a52b151d3fb81a3c9` from `main` at `b2d886b`. Implemented in parallel with other MVP flows; `atlas.sum` is expected to need re-hashing at integration.

## Outcome and decisions

IMPLEMENTED (not merged):

- Schema `20261126400000_order_exception_cases`: `order_cases`, `order_case_lines`, append-only `order_case_events`, plus a unique index on `order_items (tenant_id, order_id, id, quantity)` so case lines reference the original line and quantity. A check keeps `0 < quantity <= original_quantity`.
- API: `GET|POST|PUT /api/v1/tenants/{tenantID}/orders/{id}/case`, `POST .../case/send-back`, `POST .../case/finalize` (`OrderCaseService`, `OrderCaseStore`, `mapper/ordercase/{httpmap,persistencemap}`, explicit DTOs).
- Web: `modules/sales/components/order-case/*` panel in the order editor for confirmed orders; vi/en copy.

DECIDED in this session (implementation choices the owner should confirm):

- One case per order for its lifetime (`UNIQUE (tenant_id, order_id)`); a finalized case is terminal, matching the PRD state diagram. Multiple or follow-up cases are OPEN.
- The order row is not modified: the extended order states ExceptionOpen/ExceptionFinalized are derived from the case status, so order lists, filters and workflow code are unchanged.
- Statuses `open → sent_back → open … → finalized`. Any current member may open a case and becomes its handler; only the owner may assign or change another handler. Handler or owner records progress; owner only sends back (reason required) or finalizes. A return case needs at least one returned line before finalizing; a cancellation case has no lines.
- History is bounded at 500 versions per case. Each change also writes an `audit_logs` row. No domain event is published.
- No payment, refund, stock or inventory reservation behavior is reused or added.

## Verification and evidence

Run against a disposable PostgreSQL 17 container only (Atlas-migrated per test):

- `sh scripts/check-gofmt.sh`, `go vet ./services/api/... ./packages/shared/...`: PASS.
- `API_TEST_DATABASE_URL=<disposable> go test -race -count=1 ./services/api/... ./packages/shared/...`: PASS (service, repository, controller and contract tests for the new endpoints included).
- `atlas migrate validate`, `atlas migrate diff` (no drift), upgrade from the previous head with existing orders/items: PASS.
- `make swagger` regenerated; `pnpm --filter web lint`, `pnpm --filter web build`, `tsc --noEmit`, `pnpm --filter @hexta/sdk lint/test`: PASS.
- Browser acceptance: NOT_RUN.

## Remaining work and handoff

Integrate with other parallel flows (re-hash `atlas.sum`), browser acceptance of the panel, and owner confirmation of the decisions above. The wireframe's "order confirmed" timeline entry is not shown because orders store no confirmation timestamp in this model. FR-6.4/6.5 overview remains open.
