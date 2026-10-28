---
epic: EPIC-007
title: Approved order scope and flexible business workflows
status: in-progress
completed: false
updated: 2026-10-28
---

# Sales, work and owner overview

## Decisions and scope

- Messaging remains separate from orders. Drafts may be incomplete; assigned sale explicitly confirms only with named customer/contact, product, positive quantities, explicit nonnegative price and address when delivery applies.
- Businesses choose optional sequential workflow steps; instances pin immutable published versions. Direct/accepted handoff, owner reassignment and unassigned queue have explicit responsibility/history rules.
- Cancellation/return cases preserve original order/line history: assigned handler progresses, owner finalizes. Payment/refund execution, mandatory packing/shipping and legacy stock/auto_confirm effects are not silently approved.

## Current implementation and delivery

- Tenant-first sales/catalog/onboarding delivered through merged #54; workflow/handoff and Atlas through merged #57; compliance through merged #58. Merge status freshly checked 2026-10-28; historical sessions retain their original test/acceptance limits.
- U01–U09 acceptance/overview follow-up is implemented locally on codex/demo-work-units. [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59) is still OPEN/draft; no merge or current CI-success claim.
- Owner overview has real-data counts, coherent read snapshot, date/employee/assignment filters and pinned work/detail links. This is local delivery, not a completed epic.

## Acceptance remaining

- [x] Incomplete drafts and guarded explicit confirmation implemented/delivered.
- [x] Versioned configurable work, responsibility commands and audit history implemented/delivered.
- [ ] Complete U01–U06 manual acceptance, including separate-recipient/profile invitation/acceptance, live operator approval and remaining edge/mobile cases.
- [ ] Review/merge the local owner overview and acceptance follow-up (#59).
- [ ] Cancellation/return quantity/ownership/finalization scope and tests.

## Evidence and next handoff

- [Nine-unit plan](../../plans/2026-10-23_demo-work-units_plan.md) and [primary execution session](sessions/2026-10-23_work-unit-execution.md) own exact acceptance, U01–U09 links, automated results, bounded browser checks and approval limits. Read only the needed unit record, not all nine.
- [Owner overview integration review](../../reviews/2026-10-23_owner-overview_review.md), [overview design](../../designs/2026-10-23_owner-overview_design.md).
- [Manual browser evidence](sessions/2026-10-20_manual-browser-testing.md), [workflow implementation](sessions/2026-10-20_order-workflows.md), [initial sales](sessions/2026-10-19_manual-sales-demo.md): historical bounded evidence, not a fresh rerun.
- [Demo checklist](../../../infrastructure/TESTING.demo.md), [DB](../../designs/system/03-database.md), [HTTP](../../designs/system/04-http-api.md), [workflow contracts](../../designs/system/07-workflows.md).

The primary execution session retains the archive-approval event and observed test-record outcome; do not reinterpret it as authorized full acceptance. No screenshot/private test evidence was deleted during Markdown pruning. Finish explicit manual/review/merge gates before completion; parent remains in-progress/completed false.
