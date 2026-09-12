# Changelog: Standardize Issue & Task Lifecycle Rules

**Date**: 2026-09-12  
**Type**: `docs` / `chore`

---

## 1. Modified Files
- `.agents/scripts/get_next_issue.py`: Added mandatory `ready` label check; excluded `in-review`, `review`, and `pr-opened`.
- `.agents/skills/github-task-runner/SKILL.md`: Added label transitions for `ready` -> `in-progress` and `in-progress` -> `in-review`; added Definition of Done.
- `agentic-memory/rules/github-workflow.md`: Documented status labels and lifecycle state machine.
- `GEMINI.md`: Added Section 7 detailing lifecycle rules for all automated runners and human developers.
- `agentic-memory/plans/2026-09-12_standardize-issue-lifecycle-rules_plan.md`: Implementation plan.
- `agentic-memory/reviews/2026-09-12_standardize-issue-lifecycle-rules_review.md`: Quality and architecture review.

---

## 2. Technical Rationale
- Prevents premature processing of untriaged or underspecified GitHub issues by requiring the `ready` label.
- Improves visibility by ensuring tickets with active Pull Requests transition from `in-progress` to `in-review`.
- Formally ties task completion ("Done") to Pull Request merge into `main`.

---

## 3. Verification & Testing
- Ran `.agents/scripts/get_next_issue.py` locally and in `.worktrees/task-runner` to verify filter logic.
- Created `ready` label on GitHub.
- Verified transition commands with `gh issue edit`.
