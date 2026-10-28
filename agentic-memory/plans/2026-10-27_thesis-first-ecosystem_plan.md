# Thesis-first roadmap for the Hexta ecosystem

Date: 2026-10-27 (Asia/Ho_Chi_Minh). Status: proposed, in-progress, completed: false. Primary epic: EPIC-001. No implementation or lifecycle transitions authorized by this document.

## Overview and goal

Owner deadline: 2026-12-01 for the Hexta thesis milestone. Notion describes that date as an internal MVP/full-draft handoff, with supervisor feedback afterwards; formal submission/defense logistics remain unconfirmed. Hexta is the immediate priority. The long-term product roles are Hexta for orders/business operations, PostPilot for fanpage management, ViralForge for content creation. Develop one major product increment at a time.

This roadmap preserves the approved Hexta thesis scope. It does not silently move AI/conversations to PostPilot or remove them because of the broader product positioning. Long-term conversation ownership requires a separate product decision.

## Sources and current state

Read live Notion on 2026-10-27:

- [HEX-6 requirements](https://app.notion.com/p/2e3c653546b342b3a90a994a760f4340): 13 functional and 5 nonfunctional groups, last edited 2026-10-18. Completed task status refers to specification work, not delivered software.
- [HEX-4 MVP scope](https://app.notion.com/p/558e491fd9044a6baf30b4e61d7b92a8): newer 2026-10-18 routing/suggest/auto-send decisions supersede older draft-only paragraphs.
- [Sprint database](https://app.notion.com/p/3f6b0ea49f098074b033f0aa003f6e68): fetched schema and queried all eight milestones; Sprint 3 is 26Oct–01Nov, Sprint 6 is evaluation 16–22Nov, Sprint 8 is handoff 30Nov–01Dec.
- Queried outstanding backlog rows: 10-plant dataset remains marked not started; thesis writing/evaluation/release tasks remain outstanding. Notion states are planning evidence, not code verification.

Current source and local memory show tenant/onboarding, customers/products/orders, versioned sequential work/handoff and owner overview foundations. Browser/operator/two-identity acceptance gaps remain in the 23Oct work-unit record. Messaging/script/send-control, real AI integration, cancellation/returns and requested dashboard metrics have significant missing scope. The overview is present locally; current PR merge/CI was not refreshed, so no delivery completion is asserted. Thesis files are a template, not proof of a completed manuscript.

## Requirement coverage and near-term work packages

| Package | Requirements preserved | Evidence / work needed |
| --- | --- | --- |
| A: business foundation acceptance | FR-10 customers/products/orders, FR-11 versioned workflows; NFR-01/02/03 | Recheck current source/runtime/PRs; finish two-identity invitations/access, handoff acceptance, immutable versions and concurrency evidence |
| B: conversations and scripts | FR-01/02/06/07/08; relevant NFRs | Tenant/page/thread binding, claim/assignment, turn lifecycle, stale drafts, script versions and independent owner-granted permissions; simulator visibly labelled |
| C: routed assistance and sending | FR-03/04/05/09; NFR-02/03/04/05 | Model router, script/AI/human branches, clarification limit, suggest/auto-send policies, stop queued automation on takeover, explicit return, send outcome/unknown reconciliation |
| D: business exceptions and reporting | FR-12/13 | Cancellation/return requests, assigned handler/owner finalization and history; implement missing asked/bought product metrics with explicit definitions; reconcile overview counts/filters |
| E: evaluation and handoff | All FR/NFR evidence | Label real versus simulated source, evaluate AI versus quick-reply baseline including review/edit time, permission/race/failure tests, reproducible demo, manuscript and release package |

Packages are independently verifiable slices, not commitments to complete entire domains within a few days. No new GitHub/Notion tasks are created by this roadmap; reuse and refine existing tickets before implementation.

## Proposed calendar to 1 December

| Window | Outcome and dependency | Main inspection/change locations |
| --- | --- | --- |
| 27Oct–01Nov | Inventory every FR/NFR as implemented/verified/missing; settle dataset/model/provider access; finish A acceptance; start the smallest B→C vertical slice. Obtain source/PR/runtime baseline before claiming completion. | Existing work-unit plan/session, API router/services/repositories and tenant-qualified web UI; provider/simulator boundaries; thesis requirement/design chapters |
| 02–08Nov | Demonstrable conversation→route→proposal/send-simulation→human-takeover slice, then required B/C branches and script permissions; implement D exceptions as separate slices. Update thesis chapters alongside verified changes. | API conversation/script modules to be designed, services/ai, proto, migration schema/history, SDK, web; exception/reporting services and UI |
| 09–15Nov | Close remaining B/C/D acceptance, source-backed report definitions and failure/race tests; target integrated beta. Decide by 08Nov whether scope/capacity needs an explicit owner/supervisor discussion. | Relevant services/tests/migrations; system-design modules and thesis implementation chapter |
| 16–22Nov | Stabilization and evaluation as in Notion Sprint 6. Target feature freeze on 15Nov, earlier than Notion's final-week freeze, to protect evaluation time. No new product scope; any essential missing approved requirement is explicitly tracked as an exception. | Evaluation dataset/harness and evidence, regression tests; thesis implementation/evaluation/conclusion chapters |
| 23–29Nov | Release candidate, complete manuscript, demo rehearsal, reproducible installation and backup/demo recovery. Fix critical defects. | Deployment runbook, README, thesis, demo instructions; narrow defect fixes |
| 30Nov–01Dec | Final end-to-end verification and delivery package. | Versioned release identity, source/install docs, dataset provenance, test/evaluation evidence, full thesis draft and demonstration materials |

The calendar is a proposed allocation, not a feasibility guarantee. Available hours are unknown and missing domains are substantial. First checkpoint by 30Oct: compare remaining effort with actual weekly capacity and reserve documentation/evaluation time. If it does not fit, surface the exact requirement gap and propose an explicit reduced-depth/scope agreement; preserve the original requirement until approved. Do not make the product roadmap dependent on unapproved scope cuts or fabricate successful acceptance.

Real Pancake access is conditional; the spec permits a clearly labelled simulator fallback. A fake graph is not evidence of real model routing. For AI evaluation, obtain permitted data and model access or explicitly report missing evaluation; do not invent business prices/policies or equate simulated tests with field effectiveness. The 10-plant dataset is an early owner dependency, not a task for the final week.

## Long-term sequence after the thesis milestone

Time ranges below are planning horizons, not promised dates. Supervisor revisions can continue after 1Dec and take precedence. Reassess at each gate; do not start the next large build merely because a calendar month changed.

| Stage | Focus | Exit condition before next major build |
| --- | --- | --- |
| 1: thesis handoff | Hexta only through the agreed milestone | Reproducible MVP, requirement evidence/gaps, complete manuscript and handoff; delivered code/docs require verified main merge |
| 2: Hexta supervised pilot (roughly following 1–3 months) | Stabilize one business's order/workflow use, learn actual pain points, backup/restore and support needs; keep other apps in maintenance | Repeated real use with correct ownership/data and recoverability, understood support burden, evidence of value; pilot interest is not proof of willingness to pay |
| 3: one companion app (next outcome-driven cycle) | Tentative preference PostPilot for a narrow fanpage use case, starting with reliable connection/scheduling/publishing. Validate demand before choosing; fanpage inbox/comments/analytics are not automatically in scope | Authorized real publishing and recovery tests; pilot users repeatedly use the selected workflow; SaaS tenant isolation before multiple businesses |
| 4: ViralForge content MVP | Narrow to one useful format with quality/cost controls; reuse existing pipeline and paid-call consent | Permitted input→reviewable useful content, measured provider/render costs, repeat use, ownership/isolation and retention handled |
| 5: connect proven workflows | First approved content→PostPilot draft import; later Hexta reporting/attribution where source evidence exists | Same-tenant authorization, durable media handoff, explicit review/scheduling, duplicate/recovery tests; shared sign-in only when multiple apps are actually in use |
| 6: commercial ecosystem (6–12+ month horizon, conditional) | Packaging, subscriptions, usage budgets, shared navigation/account experience and additional apps | Demonstrated multi-app demand, support capacity, unit costs and billing/revocation rules; avoid committing to a separate platform product prematurely |

If content demand is stronger than fanpage operations, swap stages 3 and 4. Hexta remains the initial center; independent app usefulness should precede broad integration. Sequential development does not mean abandoning supported users: maintenance and urgent fixes continue while one product has feature focus.

## Work discipline and risks

- Before 1Dec, reserve nearly all planned development effort for Hexta plus thesis; PostPilot/ViralForge receive only necessary maintenance for current use. No shared platform rewrite, SSO migration, marketplace or cross-app billing project on the thesis critical path unless an existing required criterion demands it.
- Keep one major feature stream active. Choose the next task by an unmet acceptance criterion, not architectural novelty. Write thesis/evidence continuously.
- Unknown source/PR status: verify before reuse; preserve unrelated screenshots and working changes. Existing Notion statuses are not percentage-complete metrics.
- External access/cost: settle early, use approved simulation fallbacks, record paid/live versus fixture evidence accurately.
- Deadline pressure: protect evaluation and writing; escalate concrete capacity mismatch early. Do not silently drop auto-send, cancellation/returns or the research evaluation.
- Future SaaS: tenant isolation and ownership are required gates before exposing current single-user apps to multiple businesses.

## Definition of done and verification

Roadmap artifact: source links and dated owner constraints, coverage of all 13 FR/5 NFR groups, explicit proposed versus approved sequencing, capacity checkpoints and post-thesis gates. No changes to source code, Notion statuses, remote tasks or deployments.

Thesis delivery: verified build/release identity and merge evidence; all scoped requirements have actual acceptance results or an explicit approved exception; repeatable demonstration; labelled simulation and data provenance; measured evaluation with limitations; complete manuscript and install/recovery instructions. Requirements not tested remain NOT RUN, not passed. Roadmap/session stays in-progress/completed false until its delivery state is confirmed.

This planning session performed Notion reads/queries, bounded source/memory inspection and local document-link/whitespace checks. No runtime, integration or provider tests were run.
