# Implementation Plan: Issue Creator Skill & Skill Execution Protocol

- Date: 2026-10-05
- Scope: `.agents/skills/issue-creator`, `.agents/scripts/create_issue.*`, `.agents/rules/skill-execution-protocol.md`, `agentic-memory/rules/skill-execution-protocol.md`, `GEMINI.md`

## 1. Overview & Goal
Establish the missing input layer of the autonomous lifecycle by creating the `issue-creator` skill and CLI tools. Additionally, formalize the cross-skill execution protocol defining prerequisites, context traversal, and artifact persistence before initiating any task runner.

## 2. Requirements & Key Changes
1. **Issue Creator Skill (`issue-creator`)**:
   - Provide standard skill instructions (`.agents/skills/issue-creator/SKILL.md`) for crafting structured, triage-compliant issues.
   - Enforce minimum length (title >= 10, body >= 50) and acceptance criteria checklist (`- [ ]`).
2. **Issue Creation CLI Engine**:
   - Create `.agents/scripts/create_issue.py` and wrapper `.agents/scripts/create_issue.sh`.
   - Support flags `--title`, `--type`, `--priority`, `--summary`, `--requirement`, `--criterion`, `--notes`, `--file`, `--dry-run`, and automatic `ready` label tagging.
   - Add unit tests in `.agents/scripts/test_create_issue.py`.
3. **Skill Execution & Documentation Protocol**:
   - Document the 4-phase lifecycle (Pre-flight Check -> Context Loading -> Execution -> Artifact & State).
   - Define the Prerequisites & Artifact Matrix for every skill.
   - Persist in `.agents/rules/skill-execution-protocol.md` and `agentic-memory/rules/skill-execution-protocol.md`.
   - Update `GEMINI.md` Section 8.

## 3. Verification & Validation
- Run unit test suite: `PYTHONPATH=.agents/scripts python3 .agents/scripts/test_create_issue.py`.
- Run lifecycle manager tests: `PYTHONPATH=.agents/scripts python3 .agents/scripts/test_manage_issue_lifecycle.py`.
- Verify dry-run output via `.agents/scripts/create_issue.sh --dry-run`.
