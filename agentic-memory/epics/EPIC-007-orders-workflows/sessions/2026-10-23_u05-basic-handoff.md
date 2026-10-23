---
epic: EPIC-007
session: 2026-10-23_u05-basic-handoff
status: in-progress
completed: false
date: 2026-10-23
---

# U05 basic D4 handoff preparation and automated verification

## Scope

Initial automated/preparation phase of U05 in [the work-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md), covering D4 manual checklist steps 1–5. Browser/runtime ownership remains with U01; no browser action or shared runtime rebuild was performed by this unit.

## Starting evidence

Branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a`, the PR #58 main baseline supplied by the coordinator. Existing shared memory edits, sessions and screenshots were preserved. Read canonical instructions, epic context, D4 implementation session, relevant database/HTTP/workflow design modules, actual service/repository/controller tests and web workflow/work components. No new remote merge/CI lookup; this is a bounded local verification record.

## Outcome and decisions

No application defect reproduced; no production code, test, schema or API change. Existing meaningful tests already exercise the main transition path and rejected commands. No replacement workflow implementation or manufactured fix was added.

Code inspection confirms that the work UI offers owner publication and explicit creation from confirmed orders, shows pending acceptance retaining responsibility, limits acceptance to the nominated user and completion to the responsible final-step actor, and provides refresh/history. These are code observations, not browser acceptance. The sales confirmation path contains no work-creation call; the explicit workflow Create path checks confirmed status.

## Verification and evidence

Command: privately source/export the configured test environment, then run `go test -race -count=1 -v ./services/api/internal/core/service ./services/api/internal/present/http/controller -run '^TestWorkflow'`. The environment contains credentials and is deliberately not copied here. The canonical test helper creates separate random disposable PostgreSQL databases, applies the real API Atlas history, and drops only those databases afterward. Local network sandbox restriction required the approved escalated rerun.

Final run: **PASS**, seven service tests and one HTTP controller test, no skipped scenarios, no race report. Service package 6.296s; controller package 2.790s. Earlier non-exported attempt skipped DB tests; exported sandbox attempt could not connect. Neither earlier attempt is counted as evidence of application failure or acceptance.

| Scenario | Automated evidence | Browser result |
| --- | --- | --- |
| Owner publishes ordered acceptance/direct steps | Publication validation/replay and transition fixtures PASS | NOT RUN |
| Explicit confirmed-order work; drafts rejected | Creation replay/boundaries and migrated-schema tests PASS | NOT RUN |
| Nomination retains old responsible actor and current step | Acceptance/pinned-version assertions PASS | NOT RUN |
| Wrong accept actor denied; nominated recipient accepts | Service and HTTP authorization assertions PASS | NOT RUN |
| Direct handoff advances immediately | Acceptance/pinned-version assertions PASS | NOT RUN |
| Responsible actor completes final step; early completion rejected | Service/HTTP transition assertions PASS; wrong final actor denial is code-inspected only | NOT RUN |
| Reload/history; original confirmed order unchanged | Fresh service Item read asserts completed state, five events and acceptance transfer; original order status/version unchanged, PASS | NOT RUN |
| Confirmation alone creates no work | Code inspection only; no isolated explicit count assertion claimed | NOT RUN |

Additional existing service tests for reassignment/membership, real acceptance-versus-reassignment race, same-key creation race and migrated constraints passed in the same focused run. U06 still owns its full edge-case acceptance; these checks do not complete that unit.

HTTP tests use injected authenticated actors with real service/database execution, so they validate DTO/service status behavior rather than browser cookie authentication.

## Remaining work and handoff

U05 remains in-progress/completed false. After U01 authentication passes and the coordinator releases the browser slot, run checklist steps 1–5 with two current employee identities in separate profiles and distinct U05 records. Capture sanitized responsibility/step/history evidence after reload, explicitly count no work after confirmation before creation, and verify wrong actor/final-step behavior. No account links, credentials or tokens belong in screenshots or this record. No browser PASS is inferred from automated tests.

Coordinator integrates this session into epic/index/changelog serially. Browser acceptance, review and merged delivery of this evidence remain open; parent epic completion is unchanged.

## Root integration follow-up — 2026-10-23

Root browser follow-up verified explicit create, direct handoff, owner reassign/final complete, waiting nomination responsibility retention and reload/history. Actual nominated employee browser acceptance remains open. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
