---
epic: EPIC-004
session: 2026-11-10_chatbot-desk-control
status: in-review
completed: false
date: 2026-11-10
---

# Chatbot and Desk: durable control transitions

## Scope and starting evidence

Owner chose to continue the 2026-11-09 Chatbot / Desk flow, starting with durable handling state and atomic claim/handoff. Actual repository: `/Users/truonghoang/Documents/dev/personal/Hexta`; the chat workspace only contains architecture documents. Started on `codex/shopping-consultant`, HEAD `0451bd919691a1704a7071b18b57fdde37ea28f2`, with pre-existing frontend, instruction and memory edits. Preserved those changes; no reset, checkout, commit, push, PR mutation or deployment.

Read the root router, API/frontend owners, selected service/repository/authorization/errors/HTTP/mapper/Go/verification rules, Atlas fixture policy, installed Next.js client-component guidance and cross-service-change skill. Applied SVC/REP/SEC/HTTP/MAP boundaries; schema ownership unchanged. This is a scoped implementation, not a whole-repository compliance audit.

## Outcome and impact

IMPLEMENTED locally, unmerged:

- Project existing control/assignment/escalation into BOT_ACTIVE, WAITING_STAFF, STAFF_ACTIVE or fail-closed UNKNOWN. No duplicate state column or physical schema change.
- Simulator starts in staff queue because it has no connected AI. Claim clears escalation and takes staff control; owner assignment/unassignment preserves staff authority. Legacy unassigned sales-controlled work remains visible in needs_human filtering.
- Add POST handoff with a bounded reason code and POST return-to-bot with explicit expected_version. Only assigned staff or owner can hand off assigned work/return it; a current member may hand off bot work. Only owner can return queued work without first claiming it.
- Shared tenant-qualified thread lock and optimistic versions serialize inbound/claim/assignment/return. Every successful command advances control_generation/version and atomically records actor/time, previous/new state/assignee and reason in audit. Failure rolls back; a stale retry conflicts rather than duplicating a transition.
- New input preserves the chosen controller. Returning to bot never generates/replays a message. Unknown combinations cannot authorize manual sends. No automatic sender exists yet; changing a counter alone is not evidence that late AI results or real sends are fenced.
- Desk exposes current state, waiting queue, handoff reason and explicit bot return; preserves reply drafts and existing uncertain-save restrictions. Reuses existing native controls and app styles; no shared primitive or dependency added.

Contract impact: Go core/service -> DTO/http mapper/controller/router -> generated Swagger -> handwritten web service -> existing Inbox component. Added handling_state is additive; existing fields remain. Deploy API before the new UI. Old UI can still claim/reply, but bot conversations may be unassigned as before, so staff should use the new Waiting for staff view. No application migration is required for this slice.

Owning HTTP/product/workflow modules and epic navigation updated. Existing dated sessions remain historical.

## Verification and evidence

PASS:

- `make swagger`: regenerated docs.go, swagger.json and swagger.yaml through the documented generator. Generator emitted its existing root-package/runtime-constant warnings, then exited successfully. Registered-route/generated-envelope contract tests passed.
- `go test -race ./services/api/... ./packages/shared/...`: passed against an explicitly isolated PostgreSQL 17 container. Existing Atlas testutil created and dropped random disposable databases using the actual API migration history. No application DB was migrated.
- After the legacy queue compatibility fix and UNKNOWN-state test, `go test -race ./services/api/internal/core/service ./services/api/internal/core/types ./services/api/internal/present/http/controller ./services/api/docs`: passed. Covers durable reload, reason validation, cross-tenant/revoked/unassigned/other-assignee authority, stale replies/commands, queue filtering, audit rollback, concurrent return vs reassignment and the extended HTTP flow. Existing concurrent-claim tests also passed.
- `go vet ./services/api/... ./packages/shared/...`: passed before the final legacy query/test addition; final follow-up vet is recorded below.
- `pnpm --filter web lint` and `pnpm --filter web exec tsc --noEmit`: passed.
- `pnpm --filter web exec next build --webpack`: production build passed, including Inbox route. Default Turbopack build was not run.

Known pre-existing check failure: root `sh scripts/check-gofmt.sh` reports only `services/api/config/config_test.go`, outside this change. Kept it untouched. Changed Go files were formatted with gofmt; final scoped formatting/whitespace evidence is recorded below.

NOT_RUN: signed-in two-profile/mobile/keyboard browser acceptance, actual AI result/dispatch and queued-intent cancellation, notifications, real webhook/Page sends, external timeout/reconciliation and current remote CI/main merge. No paid or live provider calls. PostgreSQL container was created for this session only and removed after verification.

## Remaining work and handoff

Review this bounded simulator/control slice and perform browser acceptance before release. It remains local/in-review/completed false; parent EPIC-004 remains in-progress. Current membership/owner/assignee policy is enforced, but separate Desk/Page entitlements are not yet modeled.

Next slices: configured bot enablement and webhook routing, durable AI candidate/send intents with current-state acceptance/final dispatch gates, cancellation and unknown-result reconciliation, notifications/offline handling, handoff reason/history/summary presentation, live Page acceptance and full turn lifecycle. Do not present simulator states or successful fixtures as a live bot or Facebook integration.

## Final compatibility review

Preserved the pre-Desk simulated-reply fingerprint shape explicitly: adding the handoff-only Reason field to the shared command must not change hashes for existing saved messages. The reply replay test now checks this compatibility shape. This avoids breaking recovery of uncertain pre-upgrade replies. Final recheck evidence follows.

Final vet and changed-Go gofmt checks passed. Full `git diff --check` finds one pre-existing trailing blank line in `.agents/skills/code-review/SKILL.md`; the file was already modified at session start and remains untouched by this work. Scoped whitespace checks exclude that unrelated instruction edit.

Final compatibility recheck: the four-package PostgreSQL/race command passed after preserving the legacy reply hash; focused vet and scoped `git diff --check` passed. The replacement disposable PostgreSQL container was stopped and removed. No application data was changed.
