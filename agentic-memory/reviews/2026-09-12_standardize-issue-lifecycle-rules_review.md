# Code Review: Standardize Issue & Task Lifecycle Rules

**Date**: 2026-09-12  
**Review Target**: Issue Lifecycle Rules (`ready` -> `in-progress` -> `in-review` -> `done`)

---

## 1. Summary of Changes
- **`.agents/scripts/get_next_issue.py`**: Added check for `ready` / `status:ready` label requirement before an issue can be queued. Added `in-review`, `review`, `pr-opened` to exclusion filters.
- **`.agents/skills/github-task-runner/SKILL.md`**: Updated Step 2 to remove `ready` and add `in-progress`. Updated Step 7 to remove `in-progress` and add `in-review`. Added Step 7.7 Definition of Done.
- **`agentic-memory/rules/github-workflow.md`**: Documented the 4 lifecycle states and ASCII state machine diagram.
- **`GEMINI.md`**: Added Section 7 detailing the project-wide task lifecycle standard.

---

## 2. Quality & Architecture Assessment
- **Reliability**: Eliminates race conditions and duplicate task execution by requiring an explicit human-approved `ready` label before automated processing begins.
- **Observability**: Clearly distinguishes between issues currently being developed (`in-progress`) and issues waiting for human review (`in-review`).
- **Simplicity**: No external database or complex workflow manager needed; relies purely on GitHub issue labels and native PR merge hooks (`Closes #<number>`).
- **Standards Compliance**: All documentation and code comments strictly adhere to the English language requirement in `GEMINI.md`.

---

## 3. Verification Results
- `python3 .agents/scripts/get_next_issue.py` executed cleanly, returning `{"status": "no_tasks"}` when no issues have `ready`.
- GitHub label `ready` created with `#2da44e`.
- Open issues successfully transitioned from `in-progress` to `in-review`.
- All changes reviewed and confirmed clean.
