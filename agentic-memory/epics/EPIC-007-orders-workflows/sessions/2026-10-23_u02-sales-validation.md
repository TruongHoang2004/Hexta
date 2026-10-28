---
epic: EPIC-007
session: 2026-10-23_u02-sales-validation
status: in-progress
completed: false
date: 2026-10-23
---

# U02 sales validation and stale-edit acceptance

## Scope

Execute [U02](../../../plans/2026-10-23_demo-work-units_plan.md): UI-1, V-3 through V-6 and C-2. This initial bounded phase prepares browser verification and reruns existing sales integration checks. It does not complete the browser acceptance gate.

## Starting evidence

Shared branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a` (merged PR #58 baseline). Preserved existing unrelated memory/evidence changes. Read canonical and nested instructions, EPIC-007 context, relevant database/API/workflow modules, actual checklist, sales UI/adapter, DTO, service, repository and existing tests. U01 exclusively owns the application runtime/browser until its authentication gate is released.

## Outcome and decisions

- No application code changed and no defect was reproduced. Inspection establishes test targets only, not acceptance PASS.
- Prepared an isolated ephemeral `postgres:16-alpine` test server, container `hexta-work-unit-postgres`, on a randomly assigned loopback port with no application volume. Credentials remain in private temporary files outside the repository; no credentials appear here or in tool output. Coordinator owns cleanup after the other work units finish using it.
- Canonical `testutil.Database` created randomly named databases, applied actual API Atlas history and dropped databases on cleanup. No AutoMigrate, application data mutation, runtime rebuild, privilege provisioning or browser operation occurred in this phase.
- Shared epic/index/system/changelog integration is reserved to the coordinator to avoid concurrent overwrites.

## Verification and evidence

`go test ./services/api/internal/core/service -run '^TestSales' -count=1 -v` initially exited zero with all six tests SKIP because no explicit test URL existed. This was not a validation pass. After isolated server setup, the first sandboxed connection attempt failed due to local networking restrictions; the approved unsandboxed invocation below ran successfully.

With `API_TEST_DATABASE_URL` loaded privately for the disposable server:

```sh
go test -race ./services/api/internal/core/service -run '^TestSales' -count=1 -v
```

Exit 0; six tests PASS, package duration 9.540s, no skips:

| Automated case | Result | Evidence boundary |
| --- | --- | --- |
| TestSalesIncompleteDraftAndConfirmation | PASS | Empty draft, missing-price rejection, delivery-address rejection and explicit-zero confirmation, terminal/stale command denial and one confirmation audit |
| TestSalesTenantAndAssignmentBoundaries | PASS | Tenant and assigned-sale service boundaries; supporting evidence only, U04 owns access acceptance |
| TestSalesValidationRollbackAndStaleEdit | PASS | Failed-reference rollback, stale save rejection, negative-price rejection |
| TestSalesCatalogMissingContactsAndPrices | PASS | Missing-contact confirmation rejection, absent catalog-price preservation, archived product excluded from new draft validation |
| TestSalesConcurrentConfirmationPostgres | PASS | Real PostgreSQL competing confirmations |
| TestSalesConcurrentEditAndConfirmationPostgres | PASS | Real PostgreSQL edit-versus-confirmation conflict |

Browser results remain explicitly separate:

| Checklist case | Result | Remaining prerequisite |
| --- | --- | --- |
| UI-1 branding/light-dark/tab icon | NOT RUN | U01 runtime/auth gate, browser slot and approved tenant identity |
| V-3 delivery-address confirmation | NOT RUN | Same; automated service check passed but browser untested |
| V-4 quantity/price validation | NOT RUN | Same; negative price supported automatically, zero/negative quantity UI untested |
| V-5 incomplete confirmation | NOT RUN | Same; individual missing customer/contact/item/assignment UI cases still required |
| V-6 stale draft in two tabs | NOT RUN | Same; automated stale/race checks do not substitute for browser interaction |
| C-2 archived catalog/history | NOT RUN | Same; historical order readability after customer/product archive still required |

O-1 through O-3 browser smoke was not rerun; no relevant fix was made. Historical browser passes remain historical.

## Remaining work and handoff

After U01 passes on the agreed images, obtain the coordinator's browser slot and an existing authorized tenant identity. Use distinct `U02` test-only customer/product/order records. Run every listed browser case separately; preserve nil-versus-zero pricing. Reproduce a defect before changing code and add meaningful regression coverage only for a fix. Retain exact image/baseline evidence. Parent epic and U02 remain incomplete; no merge or acceptance completion is claimed by this session.

## Browser delegation follow-up

Coordinator released the browser phase after bounded U01 evidence and supplied the retained owner tab on the healthy main-baseline web at `http://localhost:13000` (API port 18080). This child agent's fresh `cua.getState()` reported no browser surfaces; the prescribed `cua.getTab('1', {browser: 'iab'})` returned `Browser is not available: iab`. No browser mutation or session replacement occurred. Coordinator will use the U01 executor, which already has a working IAB connection. A deterministic case-by-case U02 sequence with exact catalog/order field labels was handed off; U02 retains findings/fix evaluation ownership. The browser rows above remain NOT RUN until the executor returns actual evidence.

## Root integration follow-up — 2026-10-23

Root browser follow-up verified missing delivery address rejection/success, numeric invalidity, stale save rejection and explicit draft/confirmation smoke. Full V-5/UI-1 and product archive remain open; customer archive approval event/actual archived state are recorded in the primary execution session. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
