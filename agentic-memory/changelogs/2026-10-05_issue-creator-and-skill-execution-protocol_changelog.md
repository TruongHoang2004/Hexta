# Changelog: Issue Creator Skill & Skill Execution Protocol

- Date: 2026-10-05
- Scope: `.agents/`, `agentic-memory/`, `GEMINI.md`

## 1. Description
Added the `issue-creator` skill and companion CLI tooling to establish a standardized input flow for creating GitHub issues compatible with the autonomous task runner. Also defined and persisted the cross-skill execution protocol governing prerequisites and artifact tracking.

## 2. Modified & Created Files
- `.agents/skills/issue-creator/SKILL.md`: Standard instructions and workflow for formulating GitHub issues.
- `.agents/scripts/create_issue.py`: CLI script with input validation, markdown builder, and GitHub API execution via `gh`.
- `.agents/scripts/create_issue.sh`: Shell wrapper.
- `.agents/scripts/test_create_issue.py`: Unit test suite.
- `.agents/rules/skill-execution-protocol.md`: Protocol specification for skill lifecycle and prerequisites.
- `agentic-memory/rules/skill-execution-protocol.md`: Persisted repository rule.
- `GEMINI.md`: Added Section 8 documenting the Skill Execution & Documentation Protocol.
- `agentic-memory/plans/2026-10-05_issue-creator-and-skill-execution-protocol_plan.md`: Implementation plan.
- `agentic-memory/reviews/2026-10-05_issue-creator-and-skill-execution-protocol_review.md`: Code review verification.

## 3. Verification Steps
1. Executed `test_create_issue.py` (all tests passed).
2. Executed `create_issue.sh --dry-run` to verify structured formatting.
3. Verified full test suite across lifecycle tools.
