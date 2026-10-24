# Next demo work units

Date: 2026-10-23. Primary epic: EPIC-007; cross-epic reference: EPIC-001.
Status: proposed execution plan, in-review; completed: false. These are local work-unit IDs, not GitHub/Notion tickets.

## Overview and goal

Finish browser acceptance for the existing sales and workflow demo, then deliver the D5 owner overview in small assignments. U01–U06 are the remaining Sprint 2 acceptance work (19–25 October). U07–U09 are the next feature queue, conditional on acceptance; they are not a promise to finish D5 this week. The user will dispatch subagents; this planning session does not launch them or implement the units.

## Current state analysis

- GitHub verified PR #58 merged into main on 2026-10-23 at 11:07:01Z as `656a6c4040b31e8d28049338e3e094df5a6db88a`. Its head `2504d69` had six returned CI checks successful before merge. Local branch remains `codex/rules-compliance` at `2504d69`; no checkout/reset performed. Recheck main before starting each assignment.
- PR #54 delivered manual sales and PR #57 delivered D4 workflows/Atlas. Existing code must be verified before adding replacements.
- Historical sales browser PASS: T-1, T-2, C-1, P-1, O-1, O-2, O-3 (UI scope), V-1 and V-2. Other cases and all D4 browser cases remain open. Earlier evidence is not proof for the newly merged auth contract.
- D4 list APIs currently paginate work; no employee/time overview query contract or overview page exists in the inspected source. D5 must aggregate all matching records server-side, rather than count one client page.
- Governing sources: [Sprint 2](https://app.notion.com/p/e43dcd9f700d400bb1ac6001056bce1c), [manual checklist](../../infrastructure/TESTING.demo.md), [active demo sequence](2026-10-19_next-work-backlog_plan.md), [workflow session](../epics/EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md).
- Skill preflight: the old `agentic-memory/rules/agentic-memory.md` path does not exist. Canonical AGENTS.md epic/session rules, existing development rules and the skill-execution protocol provide the current guidance; no missing rule file was fabricated.

## Assignment contract

### Execution follow-up — 2026-10-23

The owner subsequently authorized subagents to execute all nine units. Root fetched merged main and created `codex/demo-work-units` at 656a6c4, preserving the planning edits. Auth/SDK and focused sales/invitation/access/workflow tests now have fresh passing evidence on disposable Atlas-migrated PostgreSQL; see the separate unit sessions. Browser availability/identity-isolation gaps remain explicit acceptance prerequisites. Independent U07 design and U08/U09 implementation may be staged and reviewed on this integration branch while those manual gates remain open; this does not waive acceptance, claim a merge, or mark any unit/parent done. U08 requires root review of the settled U07 contract; U09 requires the integrated U08 API. Final delivery still requires appropriate checks, browser evidence where available, review and merge.

Copy the prompt for exactly one unit, together with this paragraph: "Read AGENTS.md and applicable nested instructions. Use the latest agreed main baseline and preserve unrelated work. Work only on this unit. Verify current code before changing it. For verification units, report PASS/FAIL/NOT RUN and fix only a directly reproduced defect in this unit; if defects are independent, report them as separate follow-up units instead of bundling them. Add meaningful regression coverage for a fix. Report commit/image identity, affected files, checks and limits. Use a separate epic session for this unit. Do not mark it done until merged into main."

Each unit should produce one reviewable result. Verification without a defect can deliver evidence/docs only; it does not require a manufactured code change. Missing browser/accounts/operator inputs mean NOT RUN with an explicit prerequisite, not an inferred pass. Do not treat unavailable prerequisites as a successful acceptance gate.

## Task breakdown

### U01 — Verify browser authentication after the merge

- **Dependency:** merged PR #58; rebuild API/web together from the agreed main commit.
- **Scope:** sign in, reload, cookie refresh and server-confirmed logout. Refresh credentials must be absent from browser JSON and JavaScript storage. Check the HttpOnly/SameSite cookie attributes using browser metadata without copying credential values. Check Secure in production configuration with existing tests/configuration; do not demand Secure on local HTTP.
- **Files to inspect:** `packages/sdk/`, `apps/web/lib/sdk*`, auth navigation/services, API auth controller/service and existing auth/SDK tests; `infrastructure/README.demo.md` for matching runtime images.
- **Acceptance:** reload/refresh preserves access; successful logout blocks subsequent authenticated access; failed logout remains visible/retryable. Existing automated failure-path tests may support the failure case; use an isolated harness if another reproduction is necessary.
- **Deliverable:** bounded auth evidence and, only if needed, one auth defect fix with regression coverage.
- **Prompt:** "Execute U01 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Verify the merged cookie-auth browser flow and logout behavior against rebuilt API/web images. Fix only a reproduced auth defect and record exact evidence."

### U02 — Finish order validation and stale-edit acceptance

- **Dependency:** U01 passes for the tested baseline.
- **Scope:** checklist UI-1, V-3–V-6 and C-2; rerun the short O-1–O-3 path as a smoke after any relevant fix. Check required delivery address, invalid quantity/price, incomplete confirmation, competing draft edits and archived catalog history.
- **Files to inspect:** `apps/web/components/sales/orders.tsx`, `apps/web/lib/services/sales.ts`, sales/order service and repository, DTOs and their tests.
- **Acceptance:** invalid confirmation leaves the order draft; stale save cannot overwrite newer data; archive removes active selection while historical orders remain readable; UI branding renders. Record every listed case separately.
- **Deliverable:** sales acceptance record or one narrowly scoped validation/stale-edit fix.
- **Prompt:** "Execute U02 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Run the specified sales validation, archive and stale-edit cases. Reproduce any failure before fixing it; preserve missing-price versus explicit-zero semantics."

### U03 — Verify invitation lifecycle

- **Dependency:** U01; owner and matching employee accounts in separate browser profiles.
- **Scope:** checklist I-1–I-7: create, sign-in continuation, accept once, wrong email, revoke, reissue and decline. Invitation links remain private; no email integration is added.
- **Files to inspect:** member/invitation pages, `apps/web/lib/services/onboarding.ts`, onboarding service/repository/controller and existing tests.
- **Acceptance:** opening a link alone never creates membership; only the intended account can accept; revoked/replaced links fail; decline creates no membership; accepted membership persists and is not duplicated.
- **Deliverable:** invitation evidence or one reproduced invitation fix.
- **Prompt:** "Execute U03 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Verify I-1 through I-7 using isolated owner/employee browser profiles. Keep invitation credentials out of reports and screenshots; fix only a reproduced lifecycle defect."

### U04 — Verify tenant and actor access boundaries

- **Dependency:** U01; two tenants and owner/member/non-member test identities. U03 can supply an employee membership.
- **Scope:** T-3, I-8, I-9, A-1, A-2 and A-5. Cover tenant switching/deep links, owner-only invitation controls, another sale's assigned draft, pending registration without tenant access, ordinary-user platform denial and read-only service plan.
- **Files to inspect:** workspace/member/platform UI and onboarding/sales authorization services, existing HTTP and database tests.
- **Acceptance:** no foreign-tenant records appear; forbidden actions are denied server-side; only the assigned sale confirms. Browser rejection alone does not prove API authorization: inspect supporting tests or verify a denied request without logging secrets.
- **Optional evidence:** A-3/A-4 approval/rejection only when an already authorized operator is available. Otherwise record NOT RUN with that prerequisite. Do not provision privileges as part of this assignment. These cases remain an explicit acceptance gap.
- **Deliverable:** access matrix with results and one reproduced boundary fix if needed.
- **Prompt:** "Execute U04 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Verify tenant isolation and role/assignee denial cases. Run operator approval cases only with an existing authorized operator; otherwise preserve them as NOT RUN."

### U05 — Verify the basic D4 handoff flow

- **Dependency:** U01 and two current employee identities (U03 if invitations are needed).
- **Scope:** D4 checklist steps 1–5: publish configurable steps, create work explicitly from a confirmed order, nominate/accept, direct handoff and final completion. Include reload/history checks.
- **Files to inspect:** `apps/web/components/sales/workflows.tsx`, `work-items.tsx`, `apps/web/lib/services/workflows.ts`, API workflow service/repository/controller and existing tests.
- **Acceptance:** acceptance-required handoff retains old responsibility/step until the nominated employee accepts; wrong actor is denied; direct handoff advances immediately; only current responsible actor completes the final step. Order confirmation alone creates no work; workflow completion leaves the confirmed order unchanged.
- **Deliverable:** a complete browser path with sanitized evidence, or one reproduced workflow/UI fix.
- **Prompt:** "Execute U05 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Verify D4 checklist steps 1–5 through the browser with separate employee profiles. Prove responsibility, persistence and history at each transition."

### U06 — Verify D4 reassignment, version pinning and conflicts

- **Dependency:** U05 supplies a verified workflow and work item.
- **Scope:** D4 checklist steps 6–8: owner reassign/unassign, cancelled pending nomination, immutable workflow version pinning and accept-versus-reassign conflict.
- **Files to inspect:** workflow service/repository and tests; work/workflow UI only when reproducing an interface failure.
- **Acceptance:** reassignment cancels pending acceptance without advancing; former recipient cannot accept; unassigned work can be reassigned; old work retains old steps after new publication. A genuine same-version race has one winner and a stale loser, with consistent final state/history. Use an isolated Atlas-migrated database race test for determinism; label ordinary browser sequential stale checks accurately.
- **Deliverable:** edge-case evidence and one reproduced transition/concurrency fix if needed.
- **Prompt:** "Execute U06 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Verify D4 steps 6–8. Reuse existing meaningful concurrency tests and report browser versus automated evidence separately; never claim a race from sequential clicks."

### U07 — Specify the smallest D5 owner overview

- **Dependency:** U02/U04/U05/U06 reviewed; serious demo failures resolved. U03 evidence is required where used to establish identities. Operator-only NOT RUN cases must stay visible.
- **Scope:** documentation-only contract for a read-only owner overview over existing orders/work. Define authorization, route/query/response, card metrics, paginated work rows and filters before implementation.
- **Files:** add a focused design under `agentic-memory/designs/`; update relevant sections of system modules `04-http-api.md` and `07-workflows.md` as TARGET.
- **Required decisions:** explicit time field and timezone; inclusive/exclusive range boundaries; current responsible employee versus assigned sale meaning; waiting responsibility versus nominated recipient; unassigned behavior; counts versus paginated rows; current-step labels using each item's pinned workflow version. Distinguish unassigned from active/waiting/completed so totals cannot accidentally double-count.
- **Acceptance:** a worked example with exact expected counts and filtered rows; deterministic rules for empty data, invalid filters and foreign-tenant/member access. Resolve metric semantics explicitly rather than assuming asked-product/revenue data exists. No migration by default; document any demonstrated need separately.
- **Deliverable:** one implementation-ready contract with example request/response and test matrix. Owner-only overview is the proposed narrow authorization boundary; record it as a design decision for this feature, not an existing capability.
- **Prompt:** "Execute U07 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md using the dev-design skill. Specify the minimal D5 owner overview and exact filter/count semantics. Deliver documentation only; keep proposed contracts labeled TARGET."

### U08 — Implement the D5 read API

- **Dependency:** U07 contract settled and delivered; recheck current code against it.
- **Scope:** implement only the agreed overview read endpoint(s), including global matching counts and paginated work rows. Keep authorization in Go services and SQL/GORM in repository code. No client-side aggregation of paginated data.
- **Files:** new overview core types/service/repository/DTO/controller and focused tests following existing patterns; Fx/bootstrap/router wiring and generated Swagger. Prefer separate overview files to minimize collisions with workflow commands.
- **Acceptance:** exact U07 example; more rows than one page; empty dataset; unauthorized/foreign-tenant requests; invalid employee/date filters; unassigned work; pinned step names. Counts and rows use the same documented predicates and a consistent read boundary. No N+1 queries. Read API cannot mutate orders/work.
- **Verification:** relevant Go race tests with explicit disposable Atlas-migrated PostgreSQL where SQL behavior matters; vet, formatting, `make swagger`, generated-contract checks. Add no AutoMigrate/startup DDL.
- **Deliverable:** one API PR with implemented contract and evidence. No new indexes/schema changes without a demonstrated need and Atlas review.
- **Prompt:** "Execute U08 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Implement the settled U07 overview read API with tenant/owner authorization and server-side counts. Include focused database/HTTP verification, Fx wiring and Swagger."

### U09 — Implement and verify the D5 overview page

- **Dependency:** U08 delivered and available on the agreed main baseline.
- **Scope:** owner overview route, navigation link, typed web service, cards, filtered/paginated work rows and links to existing details. Reuse established UI/loading/error patterns and read the installed Next.js guide required by `apps/web/AGENTS.md` before editing.
- **Files:** new tenant overview page under `apps/web/app/(dashboard)/tenants/[tenantId]/`, component under `apps/web/components/sales/`, `apps/web/lib/services/overview.ts`, and `workspace.tsx` navigation.
- **Acceptance:** counts match U07/U08 with manually entered data; employee/time filters and reload behave consistently; rows show pinned current step and current responsible member; waiting recipient is separately labeled; empty/error/forbidden states are clear; switching tenants never shows old data. Refresh after a work transition reflects new counts. Non-owner denial is backed by API checks.
- **Verification:** web lint/types/production build, SDK lint when touched, browser walkthrough including small viewport and reload; report exact baseline and results. Preserve existing default landing route unless the contract explicitly changes it.
- **Deliverable:** one UI PR and end-to-end D5 evidence. No fabricated metrics, collected-revenue claims, ad analytics or product-interest metrics.
- **Prompt:** "Execute U09 in agentic-memory/plans/2026-10-23_demo-work-units_plan.md. Build the owner overview page against the delivered U08 API and U07 contract. Verify real saved records, filter/count consistency, tenant switching and empty/error states."

## Dispatch order and risks

1. Start U01.
2. After U01, U02, U03 and U04 can run independently with separate accounts/records. U04 may wait for U03 membership setup. U05 can run alongside them if its identities already exist.
3. Run U06 after U05. Browser agents must not share one profile/tab or mutate the same workflow/records concurrently. Serialize them when only one browser surface is available. Each uses distinct named test records.
4. Resolve reproduced critical failures before advancing to U07. U07 -> U08 -> U09 is sequential, avoiding speculative UI/API contracts.

Use isolated branches/worktrees for simultaneous code changes. Branch independence does not isolate the shared Docker volume, ports or browser. Agree one runtime build for a test wave; do not rebuild the shared stack while other agents verify it. A separate coordinator integrates results and updates shared epic/index/changelog files serially; unit agents should prepare isolated session records and required documentation changes in their own branches to avoid shared-file overwrites.

Out of scope: messaging implementation, AI/provider keys, Kafka, mock-data campaigns, real Pancake, payments/refunds, cancellation/return features, production deployment and unrelated architectural refactors. If a unit reveals a larger independent defect, write a follow-up with reproduction and acceptance instead of silently expanding scope.

## Definition of done

- Each unit reports its own bounded outcome; delivered code/docs require review and verified merge into main before done. Parent epics and umbrella tasks remain incomplete until all their acceptance criteria are satisfied.
- PASS/FAIL/NOT RUN are explicit, with environment/commit identity and no credentials. Historical tests remain historical; a green automated test does not substitute for an unrun browser case.
- Run checks appropriate to changed code; preserve unresolved operator/browser prerequisites. Preserve all existing data; no `down -v`, clock changes or privilege grants for convenience.
- Update affected epic session/README/index, relevant architecture modules and CHANGELOG at integration time. No GitHub/Notion lifecycle transitions or implementation were performed by this planning session.
