# Implementation Plan: Issue Lifecycle Manager & Parallel State Synchronization Workflow

- **Issue**: [#22](https://github.com/TruongHoang2004/Hexta/issues/22)
- **Title**: feat(automation): implement issue lifecycle manager and parallel state synchronization workflow
- **Date**: 2026-09-12
- **Branch**: `task/issue-22-feat-automation-implement-issue-lifecycl`
- **Scope**: Automation tooling (`.agents/scripts/`), GitHub Actions (`.github/workflows/`), Agent Skills (`.agents/skills/`), Agentic Memory (`agentic-memory/`)

---

## 1. Overview & Goal
The objective of this task is to eliminate state desynchronization, zombie task locks, and manual triage friction in the repository's autonomous task lifecycle. 

Currently:
1. When pull requests are created or merged, issues often retain stale status labels (e.g., closed issues #1, #5, and #10 remain tagged with `in-progress` or `in-review`).
2. If an autonomous runner or developer abandons a task after marking it `in-progress`, the issue stays locked indefinitely because other runners skip tasks without the `ready` label.
3. Raw backlog issues require manual inspection before receiving the `ready` label.

This implementation delivers:
- An event-driven GitHub Action (`.github/workflows/issue-lifecycle-sync.yml`) triggered on PR lifecycle events (`opened`, `reopened`, `closed`, `synchronize`).
- A Python reconciliation and triage engine (`.agents/scripts/manage_issue_lifecycle.py`) supporting `--sync`, `--triage`, and `--recover-stale`.
- A dedicated skill document (`.agents/skills/issue-lifecycle-manager/SKILL.md`).
- A pre-flight hook in `.agents/skills/github-task-runner/SKILL.md` executing `--sync --recover-stale` before task pickup.

---

## 2. Current State Analysis
- **Labels & State Protocol**: Defined in `GEMINI.md` and `agentic-memory/rules/github-workflow.md`:
  - `ready` -> `in-progress` -> `in-review` -> `done` (closed upon PR merge via `Closes #<number>`).
- **Discrepancies Observed**:
  - Closed issue #1 (`feat(config): add AI service provider configuration support`) still retains `in-progress`.
  - Closed issues #5 (`fix(db): resolve unique identifier collision`) and #10 (`feat(automation): implement scheduled codebase audit workflow`) still retain `in-review`.
  - Open PRs have active branches (e.g. `task/issue-20-*`, `task/issue-18-*`, `task/issue-8-*`, `task/issue-7-*`, `task/issue-6-*`, `task/issue-4-*`, `task/issue-3-*`).
  - No automated cleanup operates on closed issues or PR event hooks.
- **Tools & Dependencies**:
  - Python 3 standard library (`json`, `subprocess`, `argparse`, `re`, `datetime`).
  - GitHub CLI (`gh`) for API queries and mutations.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Formulate Technical Design (`dev-design`)
- Save design document to `agentic-memory/designs/2026-09-12_issue-22-issue-lifecycle-sync_design.md`.
- Detail architecture, state machine, PR-to-issue matching algorithm, recovery heuristics, and CLI interface.

### Step 3.2: Implement Lifecycle Management Engine (`.agents/scripts/manage_issue_lifecycle.py`)
- Modular architecture with Python standard library:
  - `IssueLifecycleManager` class:
    - **`sync()`**:
      - Fetches all open & closed PRs and issues.
      - Maps PRs to issues by branch naming (`task/issue-<id>-...`), commit message patterns, and PR body keywords (`Closes #<id>`, `Fixes #<id>`, `refs #<id>`).
      - Strips development status labels (`in-progress`, `in-review`, `ready`) from closed issues.
      - Ensures open issues with an active open PR are transitioned to `in-review` (removing `in-progress` / `ready`).
    - **`triage()`**:
      - Scans open backlog issues missing `ready`, `in-progress`, `in-review`, or `blocked`.
      - Evaluates issue specification quality (Summary/Objective, Acceptance Criteria checkboxes, description length).
      - Promotes qualified issues to `ready` and assigns a default priority (`priority:medium` or inferred) if absent.
    - **`recover_stale(stale_hours=2.0)`**:
      - Finds open issues labeled `in-progress`.
      - Checks whether an open PR exists for the issue or recent activity exists within `stale_hours`.
      - If inactive > `stale_hours` with no active PR, strips `in-progress`, restores `ready`, and posts an explanatory comment.
  - CLI Flags:
    - `--sync`: Run state reconciliation.
    - `--triage`: Run specification audit and auto-readiness promotion.
    - `--recover-stale`: Run zombie lock recovery.
    - `--stale-hours <float>`: Inactivity threshold in hours (default: 2.0).
    - `--dry-run`: Simulate all label mutations and comments without executing writes.
    - `--json`: Format output as JSON.
    - `--verbose`: Enable detailed log output.

### Step 3.3: Implement GitHub Actions Workflow (`.github/workflows/issue-lifecycle-sync.yml`)
- Trigger on `pull_request` events: `opened`, `reopened`, `closed`, `synchronize`.
- Also support `workflow_dispatch` (with `dry_run` option) and scheduled cron (e.g. hourly or every 30 minutes) for background stale recovery.
- Configure permissions: `issues: write`, `pull-requests: write`, `contents: read`.
- Execute `python3 .agents/scripts/manage_issue_lifecycle.py --sync`.

### Step 3.4: Create Skill Documentation (`.agents/skills/issue-lifecycle-manager/SKILL.md`)
- Document skill overview, triggers, CLI modes, lifecycle state machine, error handling, and manual overrides.

### Step 3.5: Update Task Runner Skill (`.agents/skills/github-task-runner/SKILL.md`)
- Add pre-flight synchronization hook in Step 1 before `get_next_issue.sh`:
  - Run `python3 .agents/scripts/manage_issue_lifecycle.py --sync --recover-stale`
  - Document rationale: Ensures fresh state and recovers any orphaned in-progress tickets before priority selection.

### Step 3.6: Verification & Automated Tests
- Run unit test / simulation of `manage_issue_lifecycle.py` with `--dry-run` and `--sync`.
- Validate syntax and reconciliation on actual repository issues (#1, #5, #10).
- Validate YAML syntax of `.github/workflows/issue-lifecycle-sync.yml`.
- Execute live sync on #1, #5, #10 to clean up existing stale labels.

### Step 3.7: Code Review & Changelog
- Conduct comprehensive code review and write `agentic-memory/reviews/2026-09-12_issue-22-issue-lifecycle-sync_review.md`.
- Generate detailed changelog in `agentic-memory/changelogs/2026-09-12_issue-22-issue-lifecycle-sync_changelog.md`.

### Step 3.8: Git Commit, Push & PR Creation
- Commit with Conventional Commits referencing #22.
- Push branch `task/issue-22-feat-automation-implement-issue-lifecycl`.
- Open PR with `Closes #22` and links to artifacts.
- Move label on #22 to `in-review` and post comment.

---

## 4. Risk Assessment & Edge Cases
| Risk / Edge Case | Impact | Mitigation Strategy |
|---|---|---|
| **Accidental Overwrite of Active Work** | A developer actively working locally without a PR could get reset by `--recover-stale` | 2-hour grace period default; check issue comment activity; add an explicit `--dry-run` flag; only recover if no active branch/PR. |
| **Complex PR Reference Formats** | PRs mentioning multiple issues or using variations (`fixes #X, and #Y`) | Use regex with word boundaries catching `(?:closes|fixes|resolves|refs)?\s*#(\d+)` and branch pattern `task/issue-(\d+)`. |
| **GitHub API Rate Limits** | Sequential `gh` calls during large batch updates | Batch fetch JSON with `--limit 100`; minimize API write calls only to entities requiring actual state mutation. |
| **CI Permissions in GitHub Actions** | GITHUB_TOKEN may lack `issues: write` permission | Explicitly set `permissions: { issues: write, pull-requests: write, contents: read }` in the workflow. |

---

## 5. Definition of Done (DoD)
- [x] Plan documented in `agentic-memory/plans/2026-09-12_issue-22-issue-lifecycle-sync_plan.md`.
- [ ] Technical design documented in `agentic-memory/designs/2026-09-12_issue-22-issue-lifecycle-sync_design.md`.
- [ ] `.agents/scripts/manage_issue_lifecycle.py` implemented with `--sync`, `--triage`, `--recover-stale`, and `--dry-run`.
- [ ] `.github/workflows/issue-lifecycle-sync.yml` created and validated.
- [ ] `.agents/skills/issue-lifecycle-manager/SKILL.md` created.
- [ ] `.agents/skills/github-task-runner/SKILL.md` updated with pre-flight hook.
- [ ] Existing closed issues (#1, #5, #10) reconciled.
- [ ] Code review and changelog generated in `agentic-memory/`.
- [ ] Branch pushed, PR created with `Closes #22`, and issue #22 moved to `in-review`.
