---
epic: EPIC-001
session: 2026-10-19_demo-first-reprioritization
status: in-review
completed: false
date: 2026-10-19
---

# Demo-first reprioritization

## Scope

Revise the proposed backlog following the owner's instruction to defer key/mock-data tickets and prioritize working code for rapid product feedback.

## Starting evidence

Existing uncommitted planning changes preserved. Rechecked web routes/SDK/UI inventory and order service: draft item requirement and auto-confirm stock behavior need alignment before reuse. Previous source/Notion/PR inspection remains cited in the plan; no fresh remote state claim.

## Outcome and decisions

Updated [the active plan](../../../ARCHIVE_INDEX.md#pruned-2026-10-19-next-work-backlog-plan): first deliver tenant workspace, manual customer/product CRUD and explicit order draft/confirmation; then configurable responsibility handoff and a factual overview. Manual conversations follow product feedback. Keys, datasets/mock tools, Kafka integration, AI/checkpoints/traces and external connector work are deferred from this demo critical path. Approved long-term scope remains intact. No fake send or fabricated success is introduced.

## Verification and evidence

Documentation-only; git diff --check and local link existence checks. No application tests rerun. No Notion/ticket writes or implementation.

## Remaining work and handoff

Implement D1/D2/D3 vertically and review the first browser demo before expanding. Planning delivery is unmerged; completed stays false. Future schema/API changes must amend their canonical modules.
