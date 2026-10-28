---
epic: EPIC-007
session: 2026-10-19_manual-sales-demo
status: in-progress
completed: false
date: 2026-10-19
---

# Manual sales demo: workspace, catalog and guarded orders

## Scope

Implement the owner-approved D1–D3 first demo from [the active plan](../../../ARCHIVE_INDEX.md#pruned-2026-10-19-next-work-backlog-plan), without keys, datasets, simulation tooling or new event infrastructure. [Technical design](../../../designs/2026-10-19_sales-demo_design.md). Planning context lives in EPIC-001; workflow/handoff remains unfinished in this parent epic.

## Starting evidence

Started from 413ee35 on codex/system-design-epic-memory, preserving all previous uncommitted planning records. Created codex/manual-sales-demo. Prior #52/#53 merges were verified during planning, not reinterpreted as new demo delivery. Existing order controller lacked the approved incomplete-draft/explicit-confirmation path; web only had auth/tenant screens.

## Outcome and decisions

Implemented shared persisted workspace selection, team/member access, customer/product create/list/edit/archive, and multi-line draft save/edit/confirmation UI. Added typed web services over the existing SDK. Missing prices remain distinct from explicit zero. A sales service/repository boundary checks current tenant membership, references, owner/assignee and optimistic version inside transactions; confirmation and audit commit atomically. Active product references are batched and locked, avoiding per-line queries. No stock/payment/AI/channel side effect is invoked. Legacy order controller and status/payment HTTP routes were removed, leaving internal legacy service/tests intact. Public draft PUT is full replacement with current version. Additional product variants remain preserved; this first form edits the primary variant only.

Added an additive Atlas migration for order version/delivery fields and price presence, plus a partial phone uniqueness index. Historical rows are retained. Owner-only membership invitation was tightened and tested. Browser CORS now permits Authorization/X-Tenant-ID for the configured frontend origin and explicit local origins; unrelated production origins are rejected.

## Verification and evidence

- GOCACHE=/private/tmp/hexta-go-build go test -race ./services/api/... ./packages/shared/... passed (existing DB-dependent tests retain their configured skip behavior).
- GOCACHE=/private/tmp/hexta-go-build go vet ./services/api/... ./packages/shared/... passed.
- Sales tests run on SQLite by default and on a disposable PostgreSQL database through SALES_TEST_DATABASE_URL; six scenarios cover missing data/zero price/delivery, cross-tenant and actor denial, rollback/stale edits, catalog semantics, concurrent confirmations and edit-versus-confirm. PostgreSQL creates/drops isolated test schemas; no demo data is reset.
- pnpm --filter @hexta/sdk test: 25 passed. SDK build, web ESLint/type check, Docker API and Next.js production builds passed.
- GOCACHE=/private/tmp/hexta-go-build make swagger regenerated current route documentation. atlas migrate validate passed. All five migrations applied to a fresh disposable PostgreSQL database; upgrade of the pre-existing demo volume also passed through the local launcher.
- make migrate-diff svc=api name=sales_demo DOCKER_MIGRATE=atlas was attempted using the installed CLI; the inherited config failed because Docker hostname postgres is not resolvable from the host. The reviewed additive SQL was manually scoped, checksummed, validated and applied with Atlas instead of accepting unrelated model drift or resetting a database.
- Local HTTP smoke covered auth, missing workspace, tenant isolation, customer/product CRUD, incomplete draft, rejected auto_confirm, absent price, missing delivery address, wrong actor, stale update, explicit confirmation, reload and absence of the old mutation bypass.
- Browser smoke: signed into a disposable local verification account, created customer/product, edited numeric product price, saved an empty draft, added customer/self-sale/product/quantity, saved and confirmed, then reloaded and reopened the confirmed read-only order with amount preserved. Workspace creation/switching was also verified: a new workspace showed an empty list and switching back restored only its own records. Test-only records are labeled verification and are not a business dataset or evaluation result.
- Final rebuilt runtime: browser CORS preflight and authenticated read of the persisted confirmed order passed; API/web/PostgreSQL/Redis healthy, init/migrate exit 0. The dedicated disposable verification DB was stopped after testing; the demo stack remains running.
- Source/schema/API/workflow/runbook references updated. No keys, real connectors or customer sends were used. The demo remains on loopback.

## Remaining work and handoff

Review/merge delivery is still pending; status remains in-progress, completed false. No external tracker was modified. The first demo ends at order confirmation. Configurable workflows/handoff, cases, dashboard, conversation workspace, real AI, Kafka and external deployment remain unfinished and are not implied by this session. Catalog permissions remain tenant membership for this demo; fine-grained future grants and broad schema tenant-FK retrofits are not implemented. New order creation uses a random unique number but does not yet implement persisted request-id idempotency; a response lost after commit should be checked in the order list before repeating creation. Do not mark the whole HEX-12/13/14 or parent epic done from this bounded slice.
