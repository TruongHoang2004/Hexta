---
epic: EPIC-007
session: 2026-10-23_demo-work-units
status: in-review
completed: false
date: 2026-10-23
---

# Small assignments for demo acceptance and owner overview

## Scope

Owner requested simple, clear next work units to dispatch to subagents independently. Planning only; primary EPIC-007, related EPIC-001 for auth/compliance delivery.

## Starting evidence

Local branch codex/rules-compliance at 2504d69; only pre-existing screenshot evidence was untracked at startup and is preserved. Read epic/design indexes, selected epic READMEs/latest sessions, sprint plan, actual manual checklist and workflow code. Live GitHub first showed PR #58 open with six green returned checks. After the owner reported merging, refetched PR #58 and verified merged 2026-10-23T11:07:01Z as 656a6c4040b31e8d28049338e3e094df5a6db88a. No checkout, fetch/reset, runtime rebuild or screenshot publication occurred.

## Outcome and decisions

Created [the nine-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md): U01 auth; U02 sales validation; U03 invitations; U04 access boundaries/operator prerequisites; U05 basic handoff; U06 workflow edge cases; U07 overview contract; U08 overview API; U09 overview UI. Each specifies dependencies, bounded files/scope, acceptance, deliverable and a copyable prompt. U01–U06 close Sprint 2 acceptance; U07–U09 are conditional next-feature work, not a same-week commitment. No subagents launched or application behavior implemented.

## Verification and evidence

Read-only GitHub merge/check metadata and repository code/checklist inspection. Documentation diff/whitespace/link existence checks recorded at session end. No application tests or new browser acceptance; earlier NOT RUN cases remain open. The missing historical agentic-memory/rules/agentic-memory.md path was identified and current canonical repository instructions used instead.

## Remaining work and handoff

Owner can dispatch U01 from the verified merged baseline. Serialize shared browser/runtime mutations and shared memory integration; preserve independent session records. Run U07–U09 only after demo acceptance gates. Plan documentation itself remains unmerged/in-review; parent epic unchanged. PR #58 merged delivery does not complete broader browser acceptance or parent epics.
