---
epic: EPIC-007
session: 2026-10-23_u07-owner-overview
status: in-review
completed: false
date: 2026-10-23
---

# U07 owner overview contract

## Scope

Documentation-only D5 overview design from [the nine-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md), using dev-design skill and its four-phase protocol. Coordinator owns shared epic/index/changelog/canonical-module integration; this unit owns only its design and this session.

## Starting evidence

Inspected current `codex/demo-work-units` source at merged baseline 656a6c4, AGENTS.md, skill/protocol, development rules, EPIC-007 index/README/latest execution session, system index and DB/API/workflow modules. Read current workflow types/models/service/repository/controller, order/member models, pagination/error/envelope patterns and canonical Atlas schema. Preserved all unrelated local changes. No git mutation or external PR/CI query performed by this unit; merge identity is coordinator-verified starting evidence.

User authorized all nine units. Coordinator authorized independent U07 design preparation after reported supporting automated auth/sales/invitation/access/workflow checks; **original U01–U06 browser/operator acceptance gates remain OPEN**, not inferred passed and not waived by this session.

## Outcome and decisions

[TARGET owner overview design](../../../designs/2026-10-23_owner-overview_design.md) defines GET `/api/v1/tenants/{tenantID}/overview`: current owner membership only, independent order/work creation cohorts, explicit half-open timestamp bounds and UI timezone, employee/assignment semantics, orthogonal unassigned counts, paginated pinned work projection, structured errors, exact fixture/example and focused test matrix.

Proposes one read-only REPEATABLE READ snapshot including plain authorization reads, two server-side aggregates and page projection. Existing Tenant/Member repository methods use FOR SHARE and must not be reused in that transaction. No migration/cache/revenue/AI metric. Canonical module snippets are proposed in the design for coordinator integration; no shared module or source edits by this unit.

## Verification and evidence

Read actual schema/code rather than relying on historical snapshots. Manually checked example arithmetic and boundary/assignee semantics. Documentation-only: no runtime/browser/unit/integration tests, no privilege changes and no PR/commit/merge. Design remains TARGET and in-review/completed false pending coordinator review and merged delivery.

## Remaining work and handoff

Coordinator reviews exact contract and integrates canonical TARGET links and shared records. U08 implements the settled read API with Atlas-backed focused tests and Swagger; U09 implements UI against that contract. Preserve original manual acceptance gaps and any operator prerequisites. Delivered docs/code require verified merge before this session can become done; parent epic remains in-progress/completed false.

## Root integration follow-up — 2026-10-23

Root reviewed and accepted the stated paired bounds/strict new-endpoint validation/status buckets/independent source-order semantics for U08/U09 staging. Docs remain unmerged/in-review. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
