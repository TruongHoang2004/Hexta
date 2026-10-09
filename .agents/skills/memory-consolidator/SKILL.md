---
name: memory-consolidator
description: Use this skill to synthesize and consolidate agentic-memory documents, update the master CHANGELOG.md and ARCHIVE_INDEX.md, and prune stale plans, reviews, changelogs, and superseded audits to keep repository storage lean.
---

# Agentic Memory Consolidator & Pruning Skill (`memory-consolidator`)

This skill manages and compacts the [`agentic-memory/`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/) directory to prevent historical artifact bloat, reduce token consumption in agent workflows, and maintain an organized, high-density repository memory.

---

## 1. When to Use This Skill

- When `agentic-memory/` accumulates numerous past plans, code reviews, or weekly audit runs.
- After merging multiple PRs or completing sprint milestones.
- When cleaning up individual task changelogs in `changelogs/` that have already been integrated into [`agentic-memory/CHANGELOG.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/CHANGELOG.md).
- Prior to creating new releases or when prompt context footprint needs optimization.

---

## 2. Consolidation & Pruning Architecture

The memory consolidator operates under a **zero-information-loss** principle:

```
┌──────────────────────────────────────────────────────────────┐
│                  agentic-memory/ Scanning                     │
└──────────────────────────────┬───────────────────────────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       ▼                       ▼                       ▼
 ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
 │ changelogs/ │         │ audits/     │         │ plans/      │
 └──────┬──────┘         └──────┬──────┘         │ reviews/    │
        │                       │                └──────┬──────┘
        ▼                       ▼                       ▼
Check & Synthesize into   Keep latest N runs,     Verify closed status;
CHANGELOG.md, then prune  prune superseded runs   Record summary in
redundant task files                              ARCHIVE_INDEX.md;
                                                  Prune if past retention
```

### Safety & Protection Invariants:
1. **Protected Dirs & Files**:
   - [`agentic-memory/rules/`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/rules/): **NEVER** modified or deleted (living architectural standards).
   - [`agentic-memory/README.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/README.md): **NEVER** deleted.
   - [`agentic-memory/CHANGELOG.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/CHANGELOG.md): Master historical log, only appended to.
2. **Active Work Protection**:
   - Artifacts tied to currently `OPEN`, `in-progress`, or `in-review` GitHub issues are **strictly preserved**.
3. **Traceability**:
   - Pruned plans and reviews have their metadata (date, category, issue #, objective) permanently logged in [`agentic-memory/ARCHIVE_INDEX.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/ARCHIVE_INDEX.md).

---

## 3. Workflow Steps

### Step 1: Pre-Flight Assessment & Preview (`--dry-run`)
Always run a dry-run first to preview what will be consolidated, archived, and pruned:

```bash
.agents/scripts/consolidate_memory.sh --dry-run
```

Review the output report:
- Total files scanned and byte footprint.
- Changelogs identified for synthesis.
- Stale audit reports exceeding the retention limit.
- Closed plans and reviews slated for archival.

### Step 2: Execute Consolidation & Pruning
Run the script to perform consolidation and cleanup:

```bash
# Standard run (default: 14 days retention for completed tasks, keep 2 latest audits)
.agents/scripts/consolidate_memory.sh

# Aggressive cleanup of all closed issue artifacts regardless of retention days:
.agents/scripts/consolidate_memory.sh --all-closed
```

### Step 3: Verify Memory Integrity
1. Verify that [`agentic-memory/CHANGELOG.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/CHANGELOG.md) contains all consolidated release entries.
2. Verify that [`agentic-memory/ARCHIVE_INDEX.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/ARCHIVE_INDEX.md) records the pruned files.
3. Check `git status` to review deleted and modified documents.

---

## 4. CLI Tool Reference: `consolidate_memory.sh` / `consolidate_memory.py`

Path: `.agents/scripts/consolidate_memory.sh` (or `python3 .agents/scripts/consolidate_memory.py`)

### Options & Flags:

| Flag | Type | Default | Description |
|---|---|---|---|
| `--dry-run` | Flag | `false` | Simulates actions without mutating files. |
| `--retention-days <N>` | Integer | `14` | Retention period in days for completed plans/reviews. |
| `--keep-audits <N>` | Integer | `2` | Number of most recent audit scans to preserve in `audits/`. |
| `--all-closed` | Flag | `false` | Prunes all artifacts for closed issues regardless of retention days. |
| `--prune-designs` | Flag | `false` | Also prunes stale designs for closed issues (default keeps designs). |
| `--json` | Flag | `false` | Outputs machine-readable JSON metrics. |

---

## 5. Examples

### Dry-run preview:
```bash
.agents/scripts/consolidate_memory.sh --dry-run
```

### Clean up after a major release:
```bash
.agents/scripts/consolidate_memory.sh --retention-days 7 --keep-audits 1
```

### Inspect as JSON:
```bash
.agents/scripts/consolidate_memory.sh --dry-run --json
```
