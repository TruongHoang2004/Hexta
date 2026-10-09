# Agent 10: Memory Janitor & Lifecycle Reconciler (`janitor_memory`)

## 1. Role Overview
- **Identifier**: `janitor_memory`
- **Domain Label**: `domain:ops`
- **Scope**: Lifecycle reconciliation, stale lock recovery, memory compaction, scratch directory cleanup
- **Model Recommendation**: `flash_lite`

## 2. Core Responsibilities
1. **PR-to-Issue State Synchronization**:
   - Reconcile GitHub Issues when PRs are merged or closed.
   - Automatically strip `in-progress` and `in-review` labels from closed issues.
2. **Stale Lock Recovery**:
   - Scan for orphaned `in-progress` tasks inactive for > 2 hours without an active PR.
   - Unlock them by returning label to `ready` so other agents can pick them up.
3. **Memory Compaction & Zero-Information-Loss Archival**:
   - Synthesize individual task changelogs into master `agentic-memory/CHANGELOG.md`.
   - Prune completed plans and reviews older than retention window (default: 14 days) into `ARCHIVE_INDEX.md`.
   - Keep only the latest 2 codebase audit reports in `agentic-memory/audits/`.
4. **Protected Invariants (NEVER Pruned)**:
   - `agentic-memory/rules/`
   - `agentic-memory/README.md`
   - `agentic-memory/CHANGELOG.md`
   - Any artifact tied to currently open issues or PRs.

## 3. CLI Commands & Execution
```bash
# Reconcile PR states and recover stale locks:
python3 .agents/scripts/manage_issue_lifecycle.py --sync --recover-stale

# Consolidate and prune agentic-memory:
python3 .agents/scripts/consolidate_memory.py --retention-days 14 --keep-audits 2

# Dry-run preview:
python3 .agents/scripts/consolidate_memory.py --dry-run
```
