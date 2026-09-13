---
name: issue-lifecycle-manager
description: Use this skill to manage issue lifecycle states, reconcile Pull Requests and GitHub issues, triage backlog tickets, and recover stale or abandoned task locks.
---

# Issue Lifecycle Manager Skill (`issue-lifecycle-manager`)

This skill coordinates autonomous issue lifecycle transitions, state reconciliation between Pull Requests and GitHub issues, backlog triage, and stale lock recovery across the Hexta monorepo.

---

## 1. Lifecycle State Machine

The repository enforces a strict 5-stage lifecycle state machine:

```
[Backlog Issue]
       │
       ▼ (Triage: meets criteria for Title, Scope, Acceptance Criteria)
    [ready]
       │
       ▼ (Runner picks task: removes 'ready', adds 'in-progress')
 [in-progress] ── (Inactivity > 2h without PR: recovered to 'ready')
       │
       ▼ (PR opened: removes 'in-progress', adds 'in-review')
  [in-review]
       │
       ▼ (PR merged into main: GitHub closes issue automatically)
 [done (Closed)] (Post-merge cleanup: removes 'in-progress' & 'in-review')
```

### Label Definitions
| Label | Description | Eligibility / Behavior |
|---|---|---|
| `ready` | Ticket is fully specified and ready for implementation. | **Only** tickets with this label are eligible for autonomous runner pickup. |
| `in-progress` | Active development underway by an agent or developer. | Prevents other runners from picking up the same task concurrently. |
| `in-review` | Code changes complete, verified, and Pull Request opened. | Awaiting review or CI verification before merge. |
| `done` (Closed) | Code merged into `main`. | Triggered automatically via `Closes #<id>` in PR description. |
| `blocked` | Task cannot proceed due to external dependency. | Skipped by autonomous runners. |

---

## 2. CLI Tooling: `manage_issue_lifecycle.py`

The lifecycle engine is located at `.agents/scripts/manage_issue_lifecycle.py` (with bash helper `.agents/scripts/manage_issue_lifecycle.sh`).

### Commands & Modes

#### A. State Synchronization (`--sync`)
Reconciles all repository Pull Requests against corresponding issue labels and cleans up closed issues:
- **Closed Issues**: Removes lingering `in-progress`, `in-review`, and `ready` labels.
- **Open PRs**: When an open PR links to an issue (via branch `task/issue-<id>-...` or `Closes #<id>`), sets `in-review` and removes `in-progress` / `ready`.
- **Merged PRs**: When all associated PRs are merged, strips remaining development labels.

```bash
# Preview synchronization actions without modifying GitHub state
python3 .agents/scripts/manage_issue_lifecycle.py --sync --dry-run

# Execute synchronization live
python3 .agents/scripts/manage_issue_lifecycle.py --sync
```

#### B. Backlog Triage (`--triage`)
Scans open backlog issues lacking any status labels (`ready`, `in-progress`, `in-review`, `blocked`).
- Evaluates specification completeness:
  - Title length >= 10 characters.
  - Body length >= 50 characters.
  - Presence of structured headings, acceptance criteria checkboxes (`- [ ]`), or specification keywords.
- Promotes qualifying issues to `ready` and assigns an estimated priority (`priority:high`, `priority:medium`, `priority:low`) if not already present.

```bash
python3 .agents/scripts/manage_issue_lifecycle.py --triage
```

#### C. Stale Task Lock Recovery (`--recover-stale`)
Detects orphaned tasks stuck in `in-progress` without an active Pull Request:
- Compares issue `updatedAt` against current time.
- If inactivity exceeds threshold (default: 2.0 hours) and no active open PR exists:
  - Removes `in-progress`.
  - Restores `ready`.
  - Posts an audit comment explaining the task lock recovery.

```bash
# Recover tasks inactive for > 2 hours (default)
python3 .agents/scripts/manage_issue_lifecycle.py --recover-stale

# Custom inactivity threshold (e.g. 1.5 hours)
python3 .agents/scripts/manage_issue_lifecycle.py --recover-stale --stale-hours 1.5
```

---

## 3. Automation Triggers

### A. Event-Driven GitHub Action
File: `.github/workflows/issue-lifecycle-sync.yml`
- **Triggers**: `pull_request` (`opened`, `reopened`, `closed`, `synchronize`).
- **Scheduled Cron**: Runs every 30 minutes (`*/30 * * * *`) with `--recover-stale`.
- **Manual Dispatch**: Can be triggered on-demand with custom flags.

### B. Pre-Flight Hook in Task Runner
The autonomous task runner (`.agents/skills/github-task-runner/SKILL.md`) executes:
```bash
python3 .agents/scripts/manage_issue_lifecycle.py --sync --recover-stale
```
prior to selecting the next issue to ensure the task queue is synchronized and no orphaned locks remain.

---

## 4. Manual Overrides & Edge Cases

1. **Marking a Task as Blocked**:
   ```bash
   gh issue edit <id> --remove-label "ready,in-progress" --add-label "blocked"
   ```
2. **Forcing Immediate Task Lock Release**:
   ```bash
   gh issue edit <id> --remove-label "in-progress" --add-label "ready"
   ```
3. **Manual Re-triage**:
   Add detailed acceptance criteria to the issue description and run:
   ```bash
   python3 .agents/scripts/manage_issue_lifecycle.py --triage
   ```
