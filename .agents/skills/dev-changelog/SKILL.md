---
name: dev-changelog
description: Use this skill after modifying code or completing a task to explain changes, document architectural decisions, provide diff summaries, and outline testing verification steps. It saves the changelog to agentic-memory/changelogs/.
---

# Development Changelog & Walkthrough Skill (`dev-changelog`)

Use this skill whenever you complete a feature, bug fix, or refactoring task to document what changed, why it changed, and how to verify it.

## Workflow

### 1. Collect Changes
- Review `git status` and `git diff` to identify all modified, created, or deleted files.
- Formulate the motivation and rationale behind non-obvious code changes.

### 2. Formulate the Changelog Document
Structure the document into:
1. **Change Summary**: Executive summary of what was implemented or fixed.
2. **Impacted Components & Files**:
   - Table or bullet points listing file paths with clickable links and action (`[NEW]`, `[MODIFY]`, `[DELETE]`).
3. **Key Technical Decisions**:
   - Why a specific approach was chosen over alternatives.
   - Any architectural or schema implications.
4. **Step-by-Step Walkthrough**:
   - Explanation of critical functions, handlers, or UI components modified.
   - Code diffs highlighting the core logic.
5. **Verification & Testing Guide**:
   - Automated test commands (e.g. `go test ./...`, `npm run test`).
   - Manual verification steps (e.g., steps to reproduce/test in browser or via API calls).

### 3. Persist into `agentic-memory/changelogs/`
- Determine `<feature-name>`.
- Get today's date in `YYYY-MM-DD` format.
- Write the changelog to:
  `agentic-memory/changelogs/YYYY-MM-DD_<feature-name>_changelog.md`
- In your conversation response, provide a brief summary and a clickable link to the saved changelog file.
