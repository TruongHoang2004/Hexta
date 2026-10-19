---
epic: EPIC-001
session: 2026-10-19_tenant-first-implementation
status: in-progress
completed: false
date: 2026-10-19
---

# Tenant-first implementation

## Scope and starting evidence

Implemented the owner-approved framework on codex/manual-sales-demo, preserving existing uncommitted sales work and local data. Cross-epic impact: EPIC-007 sales navigation/API and EPIC-003 demo runbook. Earlier planning sessions remain historical. No new PR/merge or external tracker transition is claimed.

## Outcome and decisions

IMPLEMENTED locally: pending tenant applications; independent platform operator review; atomic tenant/owner provisioning; owner email invitations with explicit matching-recipient acceptance; tenant-qualified sales API and web shell; read-only commercial fields. Removed public direct-create/direct-add and legacy global sales/AI route bypasses. Existing tenants/data preserved. Added additive Atlas migration, Swagger, typed SDK changes, operator CLI packaged in API Docker image and onboarding tests. The local smoke source now asserts pending application/no membership instead of direct tenant creation.

Manual private-link delivery explicitly states no email was sent; seven-day expiry is an implementation default. Hash-only token storage, token-free audit, tenant-before-invitation locks, revocation/reissue and idempotent acceptance/review are implemented. No provider/payment/chat/workflow expansion. See [implementation design](../../../designs/2026-10-19_tenant-onboarding_design.md).

## Verification and evidence

- Go full race suite and vet passed. Isolated HTTP workflow covers application/review/invite/accept and wrong actor rejection.
- Eleven real PostgreSQL sales/onboarding scenarios passed, including concurrent approval/accept and order races, using isolated test schemas/database.
- 25 SDK tests passed; web lint and TypeScript checks passed. Swagger regenerated and route inventory checked. Atlas integrity and six fresh migrations/86 statements passed; existing demo database upgrade preserved orders.
- Browser verified the account-level tenant chooser, submitted a pending request without active membership, and opened existing tenant nested order navigation with prior confirmed records intact.
- Final API/web image rebuild and health results are appended below when available. Revised Python smoke syntax checked; its runtime check requires rebuilding the AI image and was not rerun.

## Remaining work and handoff

Automatic approval review rejected the proposed temporary operator grant in the running demo because the live privileged access change lacked explicit user authorization. It was not executed. Live approval/invitation acceptance verification therefore remains pending; isolated HTTP/DB fixtures cover those workflows. Request explicit permission before any live grant. Do not bypass with database writes or another account.

Review/commit/merge remain open; completed is false. Email delivery/mailbox verification, billing policy, messaging and D4 workflows remain separate. Rebuild with scripts/local.sh up --build -d --wait api web; start at /tenants. Trusted operator bootstrap instructions are in the local runbook.

## Final runtime evidence

Final API/web production rebuild completed successfully; both containers, PostgreSQL and Redis are healthy, init/migrate exited successfully. Verified /app/platform-operator is executable without invoking it. Docker Desktop credential helper stalled on public metadata; a temporary empty-auth Docker config using the installed plugins/local socket completed the build without touching normal credentials. Browser reloaded the final image, confirmed pending request/no additional tenant access, opened nested orders with old records and inspected Members & invitations with its explicit no-email-delivery notice. Returned demo to /tenants. Disposable PostgreSQL test container stopped; demo volume retained. Final whitespace, Swagger bypass inventory and changed Markdown link checks passed. No operator grant or invitation activation performed on the running demo.
