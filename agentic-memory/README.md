# Project memory

This directory stores current product/architecture context, active work and necessary evidence. Engineering rules live in root [rules/README.md](../rules/README.md); canonical agent entry instructions remain [AGENTS.md](../AGENTS.md).

## Load only the relevant context

| Need | Entry |
| --- | --- |
| Choose engineering rules | [Root rule router](../rules/README.md) |
| Find current work/status | [Epic index](epics/README.md), then the relevant epic and its latest relevant session |
| Understand architecture/contract | [System-design index](designs/2026-10-18_system-design_design.md), then only the relevant module |
| Open review findings | reviews/; current source-backed audit and active delivery review |
| Active roadmap/acceptance | plans/; current roadmap, demo work units and retained ticket reconciliation |
| Historical delivery or removed file | Search an excerpt in [CHANGELOG](CHANGELOG.md) or [archive ledger](ARCHIVE_INDEX.md); do not preload either |

Do not load the whole directory, every linked session, inventories or historical ledgers. Epic READMEs describe current state; session evidence is read only for the task it supports. Runtime code/current remote state must be checked before treating a dated claim as current.

## Retention

- Keep active work, unresolved decisions/findings, unique acceptance/test evidence and canonical contracts. Preserve implemented/verified/decided/target/open distinctions and main-merge completion rules.
- Keep one primary session per meaningful work session. Create a separate plan/design/review only when requested, required by the invoked workflow, or when it adds distinct lasting information; do not create several documents that merely restate the same documentation edit.
- When material is superseded, retain its useful decisions/evidence in the current owner document and add original path/reason/replacement/recovery to ARCHIVE_INDEX before deleting it. Never prune by age alone or turn old approval into proof of current correctness.
- Keep index files navigation-only. Update the current epic summary rather than appending an endless stream of superseded status paragraphs. Past session outcomes must not be rewritten to appear completed.
- Append substantive delivery changes to CHANGELOG. It is a historical ledger, not startup context. Rules/legacy compatibility redirects here are not a second policy authority.
- Named plans/designs/reviews use YYYY-MM-DD_topic_kind.md; sessions use epics/<epic>/sessions/YYYY-MM-DD_topic.md. Use the session template. Screenshots/private runtime state are not cleanup targets merely because surrounding Markdown is old.

2026-10-28 pruning is recorded in [the cleanup session](epics/EPIC-001-system-design/sessions/2026-10-28_memory-pruning.md) and archive ledger. No application behavior was changed.

## Epic and session recording protocol

Explicit read-only tasks return their findings without writing memory. For authorized implementation, follow this protocol; update system-design modules only when their owned architecture/contracts change.

- Shared engineering instructions enter through [root AGENTS](../AGENTS.md); detailed rule modules live in root `rules/`. Cross-domain architecture lives in the modular system design. Task context, progress, blockers and handoffs live under `agentic-memory/epics/<epic-id>/`.
- Each epic has README.md containing status, completed, scope/decisions, acceptance checklist, evidence links, sessions and next work. `epics/README.md` is an index only, not a combined context dump.
- Every meaningful work session gets a separate `sessions/YYYY-MM-DD_<topic>.md` in its primary epic. Use the session template. For cross-epic work, store one primary session and link it from other affected epics; do not duplicate the full log.
- Session records include status, completed, outcome, changes/decisions, verification, unresolved work and next handoff. Never overwrite previous session history to make progress look complete; append a follow-up/correction and update the current epic summary.
- States: not-started, in-progress, in-review, blocked, done. `completed: true` is allowed only with status done and evidence that all scoped acceptance criteria are satisfied. For delivered code/docs, confirm merge into main before marking done. A completed session/slice never automatically completes its parent epic. Use last-known review state until merge is verified.
- At session end update the affected epic README, its session record, the epic index row, relevant system-design modules and CHANGELOG.md. Keep sessions scoped; do not rebuild a global SESSION_CONTEXT.md.
- Legacy SESSION_CONTEXT.md is only a redirect. Retained dated designs/plans/reviews are topic-specific evidence, not startup context; removed superseded files are indexed in ARCHIVE_INDEX.md. Do not create multiple documents that merely restate one documentation change.
