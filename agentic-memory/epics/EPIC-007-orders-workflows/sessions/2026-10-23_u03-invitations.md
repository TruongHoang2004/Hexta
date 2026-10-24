---
epic: EPIC-007
session: 2026-10-23_u03-invitations
status: in-progress
completed: false
date: 2026-10-23
---

# U03 invitation lifecycle verification

## Scope

Execute U03 from [the work-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md): manual checklist I-1 through I-7 and only directly reproduced invitation defects. Shared epic/index/changelog integration belongs to the coordinating agent.

## Starting evidence

Branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a` (merged PR #58 baseline). Existing unrelated documentation/evidence changes preserved. Read canonical/nested instructions, EPIC-007 context, system DB/API/workflow invitation sections, manual checklist, invitation/member pages, onboarding web service, Go service/repository/controller and existing tests. U01 owns runtime/authentication verification; browser use is serialized by the coordinator.

## Outcome and decisions

Read-only code inspection found the intended continuation path: invitation fragment is stored in session storage and removed from the URL, and successful login/register/callback routes back to the invitation page. Backend preview does not add membership; acceptance checks the matching email identity and is idempotent. Reissue revokes pending invitations; owner mutations and acceptance acquire the tenant before the invitation. These are IMPLEMENTED observations, not browser acceptance evidence. No reproduced application defect and no code changes.

## Verification and evidence

From `services/api`, ran:

```sh
go test -race ./internal/core/service ./internal/present/http/controller -run 'TestOnboarding(InvitationActivationAndRecipient|RevokedExpiredAndReissuedInvitations|ConcurrentAcceptPostgres)$|TestTenantOnboardingHTTPWorkflow$' -count=1 -v
```

Initial attempt without an explicit test URL skipped all four cases. Subsequent sandboxed connection attempt failed before test behavior could run. After permitting access to the dedicated ephemeral PostgreSQL test server, the command passed all four selected tests with race detection. Private test connection configuration was sourced without printing values; canonical `testutil.Database` created independent randomly named databases, applied the real Atlas API history and dropped them on cleanup. The application database and browser were untouched.

- PASS (automated): `TestOnboardingInvitationActivationAndRecipient` — wrong-account denial, preview grants no access, explicit activation, duplicate acceptance protection, consumed token cannot restore removed membership.
- PASS (automated): `TestOnboardingRevokedExpiredAndReissuedInvitations` — member invite denial, old reissued link rejection, cross-tenant revoke rejection, revoked/expired acceptance rejection.
- PASS (automated): `TestOnboardingConcurrentAcceptPostgres` — concurrent same-token acceptance produces exactly one membership.
- PASS (automated HTTP harness): `TestTenantOnboardingHTTPWorkflow` — pending link grants no membership; wrong recipient denied; matching repeated acceptance creates one membership. Harness injects authenticated actors and does not verify live browser auth middleware.

| Browser case | Result | Remaining evidence |
| --- | --- | --- |
| I-1 create | NOT RUN | Owner UI, pending entry, private-link/manual-delivery notice |
| I-2 sign-in continuation | NOT RUN | Signed-out employee opens link, signs in, preview without membership |
| I-3 accept once | NOT RUN | Accept, tenant access after reload, owner member list once |
| I-4 wrong email | NOT RUN | Fresh link rejected under another account; no access |
| I-5 revoke | NOT RUN | Owner revokes, recipient old link fails |
| I-6 reissue | NOT RUN | Prior link fails; newest matching link succeeds |
| I-7 decline | NOT RUN | Decline UI, owner declined status, no membership |

Existing inspected tests do not exercise decline explicitly. No new test was added solely to mirror the implementation; browser decline acceptance remains required. Automated checks do not replace the unrun browser cases.

## Remaining work and handoff

Wait for U01 auth gate and an exclusive browser slot. Use an existing approved tenant owner and a fresh matching employee account in isolated browser contexts; use a different identity for wrong-email rejection. Retain private invitation links only in browser/tool memory, exclude them from screenshots/logs/reports, and use distinct fresh invitations for lifecycle branches. Do not grant operator privileges or change clocks. Reproduce any failure before a narrow fix/regression. All I-1–I-7 browser acceptance remains open; this session and U03 are not complete. Coordinator owns shared memory updates and ephemeral test-server cleanup.
