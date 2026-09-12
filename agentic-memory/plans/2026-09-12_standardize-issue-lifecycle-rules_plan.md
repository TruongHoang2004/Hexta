# Implementation Plan: Standardize Issue & Task Lifecycle Rules

**Date**: 2026-09-12  
**Target Feature**: Issue & Task Lifecycle Rules (`ready` -> `in-progress` -> `in-review` -> `done`)

---

## 1. Context & Objective
The automated issue runner previously lacked clear lifecycle boundary definitions:
- When PRs were opened, issues retained the `in-progress` label, creating ambiguity on whether work was completed or still pending.
- Issues without complete specifications could be prematurely picked up by automated runners.
- The definition of "Done" was not formally established as being tied to PR merge.

The objective is to codify a 4-state lifecycle state machine (`ready` -> `in-progress` -> `in-review` -> `done`), enforce `ready` label validation in the automated selector script, update the task runner skill, and reflect rules in core project documentation (`GEMINI.md` and `github-workflow.md`).

---

## 2. Scope of Work
1. **GitHub Labels**:
   - Create `ready` label (`#2da44e`) on the repository.
   - Update existing open issues with active PRs to `in-review`.
2. **Issue Selector (`.agents/scripts/get_next_issue.py`)**:
   - Require `ready` label for an issue to be eligible for runner pickup.
   - Exclude issues with `in-progress`, `in-review`, `pr-opened`, or `blocked`.
3. **Task Runner Skill (`.agents/skills/github-task-runner/SKILL.md`)**:
   - Step 2: Remove `ready` and add `in-progress`.
   - Step 7: Remove `in-progress` and add `in-review`.
   - Add explicit Definition of Done (closed upon merge into `main`).
4. **Documentation**:
   - Update `agentic-memory/rules/github-workflow.md` with lifecycle state machine diagram.
   - Add Section 7 to `GEMINI.md`.
5. **Worktree Synchronization**:
   - Ensure runner worktree (`.worktrees/task-runner`) is synced.

---

## 3. Verification Plan
- Verify `get_next_issue.py` ignores issues without `ready` or with `in-review`.
- Verify GitHub CLI commands execute cleanly.
- Verify Markdown and YAML formatting.
