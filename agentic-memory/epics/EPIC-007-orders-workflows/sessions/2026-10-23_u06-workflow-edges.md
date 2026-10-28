---
epic: EPIC-007
session: 2026-10-23_u06-workflow-edges
status: in-progress
completed: false
date: 2026-10-23
---

# U06 D4 edge-case automated verification and browser handoff

## Scope

Initial automated/preparation phase of U06 in [the work-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md): D4 checklist steps 6–8, covering owner reassignment/unassignment, cancelled nominations, immutable pinned workflow steps and acceptance-versus-reassignment conflicts. U01 retains browser/runtime ownership. U05 browser acceptance is still a prerequisite; this session does not infer it from automated tests.

## Starting evidence

Branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a`, the merged PR #58 baseline supplied by the coordinator. Preserved unrelated shared documentation, session and evidence changes. Read AGENTS.md, epic/index/latest U05 session, canonical system-design index, relevant database/API/workflow modules, actual workflow service/repository/tests and relevant UI controls. No remote merge/CI recheck performed by this unit; runtime image identity belongs to U01.

## Outcome and decisions

No application defect reproduced. No production code, test, schema or API change was necessary. Existing focused regression tests were reused rather than creating a replacement implementation or redundant tests. Browser cases remain **NOT RUN** in this phase.

Code inspection: owner assignment cancels the recipient and retains the current step, while unassignment stores a null responsible member. Transition version comparison precedes actor-specific acceptance checks, so a stale old recipient gets conflict; a current-version former recipient gets forbidden. Commands save the new version and append the event in the authorized transaction. These observations supplement, rather than extend, the assertions below.

## Verification and evidence

Privately source and export `/private/tmp/hexta-work-unit-tests.env`, without printing its values, then run:

```sh
go test -race -count=1 -v ./services/api/internal/core/service -run '^TestWorkflow(AcceptanceAndPinnedVersions|OwnerReassignmentAndMembership|ConcurrentAcceptanceAndReassignmentPostgres)$'
```

Final approved local-network run: **PASS**, three tests, no skips or race report, service package 3.731s. The initial sandboxed run could not connect to PostgreSQL and is an environment failure, not an application failure. Canonical testutil creates random disposable databases from the explicit test server, applies the real API Atlas migration history and drops only those databases afterward; no application database or shared runtime was changed.

| Requirement | Fresh automated result and exact boundary | Browser |
| --- | --- | --- |
| Reassign pending acceptance without advancing | PASS: recipient cleared, step remains zero, state active; non-owner assignment denied | NOT RUN |
| Former nominee cannot accept | PASS: current-version former nominee gets 403; removed member also denied without changing waiting version/state | NOT RUN |
| Unassign then assign again | PASS: responsible becomes null; subsequent assignment succeeds and assigned sale can nominate another member | NOT RUN |
| Old work keeps immutable steps after publication | PASS: publishes v2 with one new step, then old item acceptance still returns workflow v1, three old steps and original first-step name | NOT RUN |
| Latest version available for new work | Latest-only definition listing passed in U05's broader run; fresh U06 does not assert creating a v2 work item | NOT RUN |
| Genuine same-version conflict | PASS: concurrent goroutines release from one barrier, accept and owner assign both submit expected version 2; exactly one succeeds and one receives 409 | NOT RUN |
| No duplicated race transition history | PASS: fresh read reports version 3 and exactly three events. Existing test does not explicitly compare every resulting assignment/step field to the winning event; do not claim that stronger assertion | NOT RUN |

The separate Go race detector checks data races; the concurrent database test independently checks optimistic-command conflict behavior. Sequential stale browser commands must not be called a simultaneous race.

## Remaining work and handoff

U06 remains in-progress/completed false. U01/shared browser executor should continue after U05 passes, with distinct U06 records and two authenticated profiles; coordinate rather than rebuilding runtime:

1. Create an acceptance-mode U06 work item on pinned workflow v1 and nominate employee B from responsible employee A. Keep B's pending detail open. As owner, choose another current responsible member under **Owner reassignment**, then **Apply reassignment**. Reload and verify recipient disappears, current step stays unchanged, new responsibility and one assignment event persist. B's previously loaded acceptance must fail without another event; after reload B has no acceptance control.
2. Owner chooses **Unassigned queue** and applies. Reload to verify null responsibility/current step/history, then assign A again and verify A can act. Use an active, non-final item.
3. Keep the old item. Publish renamed/reordered steps as v2 through **Workflows**. Reload the old item and verify its v1 labels/order; explicitly create a separate new item selecting latest v2 and verify its labels. Do not alter the shared U05 workflow definition.
4. On a fresh waiting item, load nominee acceptance and owner reassignment from the same version in two profiles. Apply one, then attempt the other from its stale view. Expect one success and a visible stale error, refresh and inspect responsibility/step/recipient/history. Record this as a **sequential stale browser check**; the genuine concurrent race evidence is the isolated test above. If acceptance wins it advances/transfers; if assignment wins it retains step/cancels recipient. Verify one version increment/event.

Capture sanitized screenshots excluding credentials/private invitation links. Record exact runtime identity and separate PASS/FAIL/NOT RUN per checklist case. Review/merge of this evidence and fuller browser acceptance remain open; no parent epic completion change. Coordinator integrates epic/index/CHANGELOG updates serially.

## Root integration follow-up — 2026-10-23

Root browser follow-up verified old running work pins v1 steps/direct policy after v2 publication, pending nomination cancellation by unassign and assign from queue. Real browser accept/reassign competition remains open; automated race evidence is separate. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
