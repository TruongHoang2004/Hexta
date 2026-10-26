---
epic: EPIC-007
session: 2026-10-23_u04-access-boundaries
status: in-progress
completed: false
date: 2026-10-23
---

# U04 tenant and actor access verification

## Scope

Bounded automated/preparation phase of U04 from [the work-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md): T-3, I-8, I-9, A-1, A-2 and A-5. Browser acceptance waits for U01 and an exclusive browser slot. A-3/A-4 require an existing authorized operator; no demo privileges may be granted for this assignment. Shared epic/index/changelog updates belong to the coordinator.

## Starting evidence

Branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a` (merged PR #58 baseline). Preserved unrelated documentation/evidence. Read canonical instructions, EPIC-007 index/README/current execution session, system-design index and relevant DB/API/workflow sections, manual acceptance checklist, existing sales/onboarding tests and corresponding service/controller code, tenant update service, workspace guard, platform page and service-plan page. No nested API instructions exist; web instructions were read and no web source was edited.

## Outcome and decisions

No application defect reproduced; no source changes or manufactured regression tests. Existing meaningful authorization tests passed freshly. Browser/runtime are untouched because U01 owns authentication verification and rebuild. This is supporting server evidence, not completion of browser acceptance.

IMPLEMENTED observations from source inspection: workspace derives tenant from the route, hides children while resolving a different tenant, renders an error on denied lookup and remounts tenant children by tenant ID. Onboarding owner gates apply to invitation list/create/revoke. Tenant update rejects supplied plan/status before mutation. Service-plan page contains read-only fields and explicitly says tenant approval does not mean payment was received. These observations are not runtime verification of rendering or navigation.

## Verification and evidence

From `services/api`, with private test configuration exported without printing values:

```sh
go test -race ./internal/core/service ./internal/present/http/controller -run 'TestSalesTenantAndAssignmentBoundaries$|TestOnboardingApprovalAndAuthority$|TestOnboardingRevokedExpiredAndReissuedInvitations$|TestTenantOnboardingHTTPWorkflow$' -count=1 -v
```

PASS: all four selected tests executed with race detection, no skips in the successful run. Core service package passed in 3.873 seconds; controller package passed in 2.913 seconds. Initial non-exported configuration skipped tests; corrected exported configuration hit the sandbox connection boundary. A permitted rerun against the dedicated ephemeral PostgreSQL server passed. The canonical `testutil.Database` helper created independently named disposable databases, applied real Atlas API migrations and dropped the databases on cleanup. Existing operator rows in these disposable fixtures do not grant privileges in the demo. No application database mutation or runtime image build was performed.

| Case | Fresh supporting result | Browser acceptance |
| --- | --- | --- |
| T-3 | PASS, bounded service checks: non-member foreign-tenant customer list returns 403; foreign customer/product draft references return 404 even for an owner of both tenants | NOT RUN: switch/two-tab/deep-link rendering and absence of stale records remain |
| I-8 | PASS, bounded service checks: ordinary member invite creation returns 403; cross-tenant revoke fails. Owner gates for list/revoke and workspace deep-link handling inspected | NOT RUN: owner controls, uninvited deep link and live denied request remain |
| I-9 | PASS: peer cannot edit another sale's draft (403); owner cannot confirm on behalf of assigned sale (403); owner reassignment succeeds; old assignee then gets 403 and new assignee confirms | NOT RUN: browser read-only/denial path remains |
| A-1 | PASS, bounded: application creates no tenant for requested slug before review; HTTP submission succeeds in harness | NOT RUN: fresh account signup, PENDING display and no tenant card/access remain |
| A-2 | PASS: ordinary tenant owner cannot list platform requests or review an application (403); review denial also verified through HTTP handler | NOT RUN: ordinary user's platform page denial remains |
| A-5 | PASS, bounded: owner's attempted plan update returns 403. Read-only page and non-payment copy inspected | NOT RUN: owner page rendering remains; status denial is inspected, not separately exercised in selected tests |
| A-3/A-4 | NOT RUN for live operator acceptance; disposable test approval is only supporting automated evidence | Requires an already authorized demo operator; rejection browser path remains unverified |

The HTTP harness injects authenticated actors, so it verifies handler/service authorization rather than live authentication middleware. Existing fixture actors/records are synthetic and independent from current browser identities. No claim is made that every list/detail endpoint or every invitation owner mutation was executed by this selection.

## Remaining work and handoff

After U01 passes, serialize U04 browser execution using two approved tenants and owner/member/non-member identities with distinct records. Verify each case separately and inspect denied authenticated API behavior without exposing credentials. Reuse U03 membership where available. Keep A-3/A-4 explicitly NOT RUN without an existing authorized operator; do not grant privileges. Record exact runtime image identity when browser work occurs. U04 remains in-progress/completed false; no merge or parent epic completion is claimed.

## Root integration follow-up — 2026-10-23

Root browser follow-up verified tenant switching to empty counts, ordinary-owner /platform denial and read-only service plan. Remaining multi-actor/operator cases remain open. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
