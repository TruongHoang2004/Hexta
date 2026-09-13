# Development Changelog & Walkthrough: Issue Lifecycle Manager & Parallel State Synchronization

- **Issue**: [#22](https://github.com/TruongHoang2004/Hexta/issues/22)
- **Title**: feat(automation): implement issue lifecycle manager and parallel state synchronization workflow
- **Date**: 2026-09-12
- **Author**: Autonomous Task Runner Agent
- **Branch**: `task/issue-22-feat-automation-implement-issue-lifecycl`

---

## 1. Change Summary

This change implements an automated issue lifecycle manager and parallel state synchronization workflow across the Hexta monorepo. It establishes real-time synchronization between Pull Requests and GitHub Issues, strips stale development labels from closed issues (#1, #5, #10), automatically unlocks zombie task locks inactive for > 2 hours, and audits backlog issues for readiness promotion.

---

## 2. Impacted Components & Files

| File | Action | Description |
|---|---|---|
| [`.agents/scripts/manage_issue_lifecycle.py`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.agents/scripts/manage_issue_lifecycle.py) | `[NEW]` | Core reconciliation, triage, and stale recovery engine with CLI interface. |
| [`.agents/scripts/manage_issue_lifecycle.sh`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.agents/scripts/manage_issue_lifecycle.sh) | `[NEW]` | Shell wrapper for convenient CLI execution. |
| [`.agents/scripts/test_manage_issue_lifecycle.py`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.agents/scripts/test_manage_issue_lifecycle.py) | `[NEW]` | Comprehensive unit test suite (6 unit tests). |
| [`.github/workflows/issue-lifecycle-sync.yml`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.github/workflows/issue-lifecycle-sync.yml) | `[NEW]` | GitHub Actions event-driven sync on PR events and 30-minute cron. |
| [`.agents/skills/issue-lifecycle-manager/SKILL.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.agents/skills/issue-lifecycle-manager/SKILL.md) | `[NEW]` | Operational documentation and guidelines for issue lifecycle management. |
| [`.agents/skills/github-task-runner/SKILL.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.agents/skills/github-task-runner/SKILL.md) | `[MODIFY]` | Integrated pre-flight sync hook before task selection. |
| [`agentic-memory/plans/2026-09-12_issue-22-issue-lifecycle-sync_plan.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/plans/2026-09-12_issue-22-issue-lifecycle-sync_plan.md) | `[NEW]` | Implementation plan artifact. |
| [`agentic-memory/designs/2026-09-12_issue-22-issue-lifecycle-sync_design.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/designs/2026-09-12_issue-22-issue-lifecycle-sync_design.md) | `[NEW]` | Technical design artifact with state machine diagram and schema. |
| [`agentic-memory/reviews/2026-09-12_issue-22-issue-lifecycle-sync_review.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/reviews/2026-09-12_issue-22-issue-lifecycle-sync_review.md) | `[NEW]` | Code review report. |
| [`agentic-memory/changelogs/2026-09-12_issue-22-issue-lifecycle-sync_changelog.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/changelogs/2026-09-12_issue-22-issue-lifecycle-sync_changelog.md) | `[NEW]` | This changelog and walkthrough document. |

---

## 3. Key Technical Decisions

1. **Dual-Layer Synchronization (Event-Driven + Periodic Polling)**:
   - Event-driven GitHub Actions execute within seconds of a PR being opened, merged, or closed.
   - Periodic cron (every 30 mins) and pre-flight task runner hooks guarantee reconciliation if GitHub webhooks are delayed or offline actions occur.
2. **Batch Querying & Zero-Mutation Idempotence**:
   - Fetches repository issues and pull requests in bulk via `gh issue list --json` and `gh pr list --json` to prevent API rate exhaustion.
   - Compares required label state with current label state before issuing any `gh issue edit` command.
3. **Multi-Signal PR-to-Issue Association**:
   - Resolves issues from Git branch convention (`task/issue-<id>-...`), closing directives (`Closes #<id>`, `Fixes #<id>`), and issue mentions (`refs #<id>`).
4. **Active PR Guard on Stale Recovery**:
   - Before unlocking an `in-progress` task for inactivity, the engine verifies that no open PR is associated with the issue, ensuring active reviews are never disrupted.

---

## 4. Step-by-Step Walkthrough

### 4.1 Lifecycle Synchronization Engine (`manage_issue_lifecycle.py`)
- **`sync()`**:
  - Traverses closed issues and strips lingering labels (`in-progress`, `in-review`, `ready`).
  - Scans open issues and matches them with open PRs; transitions them to `in-review` and removes `in-progress` and `ready`.
- **`triage()`**:
  - Validates backlog items against minimum specification lengths and structured sections (acceptance criteria / checkboxes).
  - Promotes qualifying tickets to `ready` and assigns inferred priority.
- **`recover_stale()`**:
  - Inspects `in-progress` issues without open PRs.
  - If inactivity exceeds threshold (default 2h), removes `in-progress`, adds `ready`, and posts an explanatory comment.

### 4.2 GitHub Actions Workflow (`issue-lifecycle-sync.yml`)
- Triggers on `pull_request` types `[opened, reopened, closed, synchronize]`, `schedule` (`*/30 * * * *`), and `workflow_dispatch`.
- Sets up Python 3.12 and executes `manage_issue_lifecycle.py --sync --recover-stale`.

### 4.3 Task Runner Pre-Flight Hook (`github-task-runner/SKILL.md`)
- Step 1 updated with pre-flight reconciliation call:
  ```bash
  python3 .agents/scripts/manage_issue_lifecycle.py --sync --recover-stale
  ```

---

## 5. Verification & Testing Guide

### Automated Unit Tests
Run the standalone unit test suite:
```bash
python3 .agents/scripts/test_manage_issue_lifecycle.py
```
*Result: 6/6 tests passing (100% component coverage).*

### Dry-Run Verification
Preview operations across repository issues:
```bash
python3 .agents/scripts/manage_issue_lifecycle.py --sync --recover-stale --triage --dry-run --verbose
```

### Live Closed Issue Reconciliation Verification
Executed live `--sync` on repository issues:
- Issue #10 (`feat(automation): implement scheduled codebase audit workflow`): Removed stale `in-review` label.
- Issue #5 (`fix(db): resolve unique identifier collision`): Removed stale `in-review` label.
- Issue #1 (`feat(config): add AI service provider configuration support`): Removed stale `in-progress` label.
- Subsequent run confirmed 0 remaining mutations (complete idempotence).
