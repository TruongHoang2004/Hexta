---
epic: EPIC-001
session: 2026-10-19_documentation-pr
status: in-review
completed: false
date: 2026-10-19
---

# Documentation delivery PR

## Scope

Commit and publish the canonical system design, shared agent instructions and epic-scoped session history requested by the owner.

## Starting evidence

PR #52 was checked on 2026-10-19: OPEN, targeting main, branch codex/durable-ai-local-stack. Code baseline is 01b2e11. Use documentation branch codex/system-design-epic-memory with base codex/durable-ai-local-stack so only the documentation follow-up is reviewed here.

## Outcome and decisions

Publish a separate stacked documentation PR. Merge the implementation PR first; then retarget/rebase the documentation PR to main if necessary. Keep epic/session completion false until delivery merge is confirmed. Before commit, the owner clarified the canonical filename must be AGENTS.md rather than the initially requested singular name. Consolidated all shared rules into AGENTS.md, removed the temporary agent.md, and made GEMINI.md redirect directly to AGENTS.md. No application behavior is changed.

## Verification and evidence

Check preservation of all nine migrated engineering-rule sections, status/completed consistency across epic/session metadata and index, local documentation links, balanced fences and staged whitespace. No runtime tests are rerun for this documentation-only change.

## Remaining work and handoff

Human review and merge remain. Confirm parent PR #52 merge/CI and final documentation PR target before merging. Update epic/session status only after verified delivery.
