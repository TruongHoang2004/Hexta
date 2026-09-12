# GitHub Workflow & Issue Automation Rules

This document outlines the standard process for task management, issue labeling, branching, and automated processing.

---

## 1. Issue Management & Labeling

### Priority Labels
- `priority:high`: Critical bugs, blockers, high-priority features. Processed first.
- `priority:medium`: Standard features and improvements. Processed second.
- `priority:low` or unlabelled: Minor adjustments, polish, tech debt. Processed in chronological order (oldest first).

### Status Labels
- `in-progress`: Assigned when work has begun to prevent race conditions or duplicate execution.
- `blocked`: Marks issues that depend on external prerequisites or user clarification. Skipped by automated runners.

---

## 2. Branching & PR Strategy

- **Base Branch**: All tasks branch from up-to-date `main`.
- **Branch Naming**: `task/issue-<number>-<short-slug>`
- **Pull Requests**:
  - Direct commits to `main` by automated agents are prohibited.
  - Pull Requests must specify `Closes #<number>` in the description to automate issue closing upon merge.
  - Pull Request descriptions must link to the generated documents in `agentic-memory/plans/`, `agentic-memory/reviews/`, and `agentic-memory/changelogs/`.

---

## 3. Commit Convention
Use Conventional Commits:
- `feat(<scope>): <description> (refs #<number>)`
- `fix(<scope>): <description> (refs #<number>)`
- `refactor(<scope>): <description> (refs #<number>)`
- `chore(<scope>): <description> (refs #<number>)`

---

## 4. Automated Execution Trigger
- The scheduled cron job runs periodically (`*/15 * * * *`).
- It executes `.agents/scripts/get_next_issue.sh` and activates the `/github-task-runner` skill.
