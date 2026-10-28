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
