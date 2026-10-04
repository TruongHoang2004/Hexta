# Plan: PR Review Agent with Auto-Merge & Conflict Escalation

**Date**: 2026-09-13  
**Issue**: [#26 - feat(automation): implement PR review agent with auto-merge for simple PRs and conflict escalation](https://github.com/TruongHoang2004/Hexta/issues/26)  
**Branch**: `task/issue-26-feat-automation-implement-pr-review-agen`  

---

## 1. Overview & Goal

Currently, the Hexta monorepo has multiple open Pull Requests in `in-review` status without an automated merge pipeline. As PRs linger, divergence from `main` leads to conflicts (e.g. PR #23, #15, #14, #13, #9). The objective is to build an autonomous PR Review & Merge Agent that:
1. Categorizes open pull requests by risk level (Low, Medium, High).
2. Autonomously merges low-risk, clean pull requests using squash merge.
3. Attempts intelligent rebase and automated conflict resolution for trivial lockfile/generated file conflicts.
4. Escalates complex conflicts and high-risk changes (security, auth, migrations, large blast radius) to human maintainers via PR comments and the `needs-human-review` label.
5. Provides dry-run previews, CLI entry points, and an automated GitHub Actions workflow.

---

## 2. Current State Analysis

- **Open PRs**: 8 open PRs (#24, #23, #19, #15, #14, #13, #12, #9).
  - Clean & Mergeable: PR #24 (Automation/Lifecycle), PR #19 (Local dev guide / docs), PR #12 (Security auth fix).
  - Conflicting / Dirty: PR #23 (Go module migration - 62 files), PR #15 (SDK token refresh), PR #14 (Multi-tenant DB & API), PR #13 (Web auth tokens), PR #9 (CI workflows).
- **Existing Automation**:
  - `manage_issue_lifecycle.py` and `get_next_issue.py` handle issue transitions, but stop at creating PRs in `in-review`.
  - There is no automated PR processing or merging daemon.
  - Branch protection rules are disabled on `main`, enabling programmatic squash merging.
- **Critical File Paths & Escalation Triggers**:
  - Auth/JWT/Security paths (`services/api/internal/core/service/auth_service.go`, `**/middleware/auth*`, `**/security*`, `**/oauth*`).
  - Database migrations (`migrations/api/*.sql`).
  - Module roots and path definitions (`services/api/go.mod`, `packages/shared/go.mod`).
  - High blast radius (`changedFiles > 50` or core services modified without accompanying unit tests).

---

## 3. Task Breakdown (Step-by-Step)

### Phase 1: Technical Design
- Define risk matrix heuristics and path classification regexes.
- Define git rebase and conflict resolution algorithm (detecting ours/theirs lockfile regeneration vs complex overlapping AST conflicts).
- Define GitHub API interactions via `gh` CLI for label assignment, comment posting, and squash merging.
- Save to `agentic-memory/designs/2026-09-13_issue-26-pr-review-agent_design.md`.

### Phase 2: Core Script Implementation
- Create `.agents/scripts/pr_review_agent.py`:
  - `RiskClassifier`: Analyzes PR files, diff stats, labels, and title/body to assign `LOW`, `MEDIUM`, or `HIGH`.
  - `ConflictResolver`: Evaluates `CONFLICTING` branches, attempts git rebase against `main`, detects if conflicts only touch lockfiles/generated files (`go.sum`, `go.work.sum`, docs), resolves trivial conflicts, or aborts cleanly.
  - `ReviewEngine`: Evaluates PR list ordered by priority/age, executes `--review-and-merge`, `--resolve-conflicts`, and `--dry-run`.
  - `Escalator`: Adds `needs-human-review` label and leaves structured markdown comments with file breakdowns and conflict details.
- Create `.agents/scripts/pr_review_agent.sh` executable wrapper.
- Create automated unit tests in `.agents/scripts/test_pr_review_agent.py` to test risk classification, path matching, and conflict handling with high test coverage.

### Phase 3: Workflow, Skill & Rule Updates
- Create GitHub Actions workflow `.github/workflows/pr-review-agent.yml` with `workflow_dispatch` and scheduled cron triggers.
- Create skill `.agents/skills/pr-review-agent/SKILL.md` documenting usage, flags, risk classifications, and manual operator overrides.
- Update `agentic-memory/rules/github-workflow.md` documenting the PR review agent lifecycle state machine and handoff from task runner.

### Phase 4: Verification & Validation
- Run test suite: `python3 .agents/scripts/test_pr_review_agent.py`.
- Run dry-run execution: `python3 .agents/scripts/pr_review_agent.py --dry-run`.
- Verify risk classification accuracy across all 8 existing open PRs.
- Validate workflow YAML syntax and file permissions.

### Phase 5: Documentation & Pull Request
- Create Code Review report in `agentic-memory/reviews/2026-09-13_issue-26-pr-review-agent_review.md`.
- Create Changelog in `agentic-memory/changelogs/2026-09-13_issue-26-pr-review-agent_changelog.md`.
- Commit, push, and open Pull Request with Conventional Commits linking Issue #26.
- Transition issue to `in-review` and post completion comment.

---

## 4. Risk Assessment & Edge Cases

| Risk / Edge Case | Impact | Mitigation Strategy |
|---|---|---|
| Auto-merging a breaking change | Regressions on `main` | Strict risk engine: any security, migration, core service without tests, or >50 files is classified `HIGH` and escalated. |
| Rebase failure leaves working directory dirty | Corrupted agent worktree | Ensure rebase operations run in isolated temp worktrees or cleanly `git rebase --abort` upon failure. |
| GitHub API rate limiting or network glitches | Partial merge or lost state | Standardize timeout and error handling around `gh` CLI calls; fail gracefully without partial state. |
| False positive auto-resolutions | Silent syntax error in merged PR | Only auto-resolve deterministic files (e.g., `go.sum`, `go.work.sum` re-generation via `go mod tidy` / `go work sync`). Overlapping source code conflicts must always escalate. |

---

## 5. Definition of Done (DoD)

- [x] Plan documented in `agentic-memory/plans/2026-09-13_issue-26-pr-review-agent_plan.md`.
- [ ] Technical design documented in `agentic-memory/designs/2026-09-13_issue-26-pr-review-agent_design.md`.
- [ ] `.agents/scripts/pr_review_agent.py` and `.agents/scripts/pr_review_agent.sh` fully implemented with `--review-and-merge`, `--resolve-conflicts`, `--dry-run`.
- [ ] Unit tests in `.agents/scripts/test_pr_review_agent.py` passing 100%.
- [ ] Label `needs-human-review` confirmed active.
- [ ] Skill `.agents/skills/pr-review-agent/SKILL.md` created.
- [ ] Workflow `.github/workflows/pr-review-agent.yml` created.
- [ ] Rules updated in `agentic-memory/rules/github-workflow.md`.
- [ ] Code review completed in `agentic-memory/reviews/2026-09-13_issue-26-pr-review-agent_review.md`.
- [ ] Changelog written to `agentic-memory/changelogs/2026-09-13_issue-26-pr-review-agent_changelog.md`.
- [ ] Branch pushed, PR opened with `Closes #26`, issue moved to `in-review`.
