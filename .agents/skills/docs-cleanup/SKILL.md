---
name: docs-cleanup
description: "Audit Hexta agent skills, instruction adapters and Markdown docs; classify each KEEP/MERGE/DELETE/REVIEW with repository evidence and apply only high-confidence, low-risk cleanup. Does not commit, push or touch application behavior."
---

# Skills and Markdown Cleanup

Reduce instruction and documentation maintenance overhead without losing unique knowledge, breaking agent loading or changing application behavior.

## 1. Scope and discovery model

- Candidates: `.agents/skills/**`, `.agents/rules/**`, `.claude/skills/**`, `AGENTS.md`/`CLAUDE.md`/`GEMINI.md` files, `rules/**`, `docs/**` and other `**/*.md`.
- Learn loading first from [instruction maintenance](../../../docs/architecture/ai-instructions.md) and [the router](../../../rules/README.md). Skills and AGENTS files are discovered implicitly by directory; no inbound link does not mean unused.
- `agentic-memory/**` pruning belongs to `memory-consolidator` and the [project memory](../../../agentic-memory/README.md) retention rules; only classify it here unless that workflow is explicitly invoked.

## 2. Gather evidence per candidate

1. Search references (relative/absolute paths, skill names, adapters, scripts, CI, `scripts/check-agent-instructions.py` lists, other skills).
2. Check generators: `.claude/skills/*` and `.agents/rules/scope-*.md` are owned by the validator; `CLAUDE.md` adapters must stay `@AGENTS.md`.
3. Use `git log --follow` when purpose is unclear. Compare content for unique, actionable information versus overlap.

## 3. Classify

| Class | Meaning |
| --- | --- |
| KEEP | Active, required, or uniquely valuable |
| MERGE | Useful but substantially overlaps an owner document |
| DELETE | High-confidence obsolete, redundant or irrelevant, with evidence |
| REVIEW | Usage or value cannot be established confidently |

Never delete by age or missing references alone. Always KEEP or REVIEW: canonical/domain AGENTS and CLAUDE adapters, router and rule modules with stable IDs/provenance, compatibility redirects and review aliases, generated Next.js agent blocks, skill-referenced resources, and anything a validator lists as canonical.

## 4. Apply

- Delete only DELETE items. Merge into the existing owner; keep one authoritative body per rule ID and preserve constraints, links and provenance.
- Removing a canonical skill: delete its adapter in the same change (sync never deletes), update `NEW_SKILLS`, the router, skill protocol and guide references when they name it. Never hand-edit generated adapters; run `--sync-adapters`.
- Repair every link to moved/merged/deleted files. Do not rewrite unrelated docs, create new docs without need, or modify application source.
- Preserve REVIEW items. Do not commit or push.

## 5. Validate

```sh
python3 scripts/check-agent-instructions.py
python3 scripts/check-agent-instructions.py --self-test
```

Also grep for remaining references to removed paths/names. Report native-client loading as NOT_RUN unless observed in a fresh session.

## 6. Report

Sections: Deleted (reason + evidence), Merged (what was preserved where), Preserved (why apparent redundancy stays), Manual Review Required (what blocks a decision), Validation (commands, results, limitations). Record a session only per project memory when authorized; read-only requests produce no files.
