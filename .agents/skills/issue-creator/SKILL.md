---
name: issue-creator
description: Use this skill to formulate, validate, and create standardized GitHub issues for features, bug fixes, refactoring, or technical debt, ensuring full compatibility with autonomous task runners and lifecycle triage rules.
---

# GitHub Issue Creator Skill (`issue-creator`)

This skill coordinates the formulation, validation, and submission of standardized GitHub issues in the Hexta monorepo. It ensures that every ticket is well-specified, properly categorized, and immediately actionable for both human engineers and the autonomous [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md).

---

## 1. Specification & Quality Standards

Issues created in this repository must satisfy the triage rules enforced by [`manage_issue_lifecycle.py`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/scripts/manage_issue_lifecycle.py):

| Element | Requirement | Description / Rule |
|---|---|---|
| **Title** | Length >= 10 chars | Conventional format: `<type>(<scope>): <concise summary>` (e.g., `feat(api): add refresh token endpoint`, `fix(web): resolve tenant header redirect`). |
| **Body** | Length >= 50 chars | Structured markdown with headings (`## Summary & Context`, `## Requirements & Scope`, `## Acceptance Criteria`). |
| **Acceptance Criteria** | Mandatory checkboxes | Must contain at least one concrete checkbox (`- [ ]`) describing measurable verification criteria. |
| **Priority** | Explicit label | `priority:high`, `priority:medium`, or `priority:low`. |
| **Status** | Lifecycle label | `ready` if the issue is fully specified and ready to be picked up immediately by `github-task-runner`. |

---

## 2. Standard Issue Structure

```markdown
## Summary & Context
[Clear problem statement or feature overview explaining the 'why' and context]

## Requirements & Scope
- [Specific requirement 1]
- [Specific requirement 2]
- [Non-goals or out-of-scope boundaries]

## Technical Details & Architecture
[Target architecture, 5-layer Go compliance, impacted packages or DTO patterns]

**Key Affected Files / Packages:**
- `services/api/internal/...`
- `packages/shared/...`

## Acceptance Criteria
- [ ] [Measurable condition 1]
- [ ] [Unit / integration test added or passing]
- [ ] [Swagger / documentation updated if API endpoint]

## Automated Workflow Checklist
- [ ] `github-task-runner` picks up task from `ready` queue
- [ ] Implementation plan recorded in `agentic-memory/plans/`
- [ ] Code review audit recorded in `agentic-memory/reviews/`
- [ ] Changelog recorded in `agentic-memory/changelogs/`
```

---

## 3. Workflow Steps

### Step 1: Analyze & Clarify Requirements
1. Understand the goal: feature request, bug fix, refactor, or architectural task.
2. Check existing codebase conventions (`GEMINI.md`, `agentic-memory/rules/`).
3. Identify affected components (e.g., `services/api`, `apps/web`, `packages/shared`).
4. If critical requirements are ambiguous, clarify with the user before finalizing the ticket.

### Step 2: Formulate & Validate Draft
Draft the issue title, priority, type, requirements, and acceptance criteria.
Run pre-validation or preview using the `--dry-run` flag:

```bash
.agents/scripts/create_issue.sh \
  --title "feat(api): add user session revocation endpoint" \
  --type feat \
  --priority medium \
  --summary "Provide an HTTP endpoint to revoke active user sessions." \
  --requirement "Add DELETE /api/v1/sessions/:id controller" \
  --requirement "Validate caller ownership of session" \
  --criterion "Controller returns 200 on successful revocation" \
  --criterion "Unit tests in session_service_test.go cover invalid session ID" \
  --notes "Follow 5-layer architecture and use response.Response[T] wrapper" \
  --file "services/api/internal/present/http/controller/session_controller.go" \
  --dry-run
```

### Step 3: Create GitHub Issue
Execute the script to publish the issue to GitHub:

```bash
.agents/scripts/create_issue.sh \
  --title "feat(api): add user session revocation endpoint" \
  --type feat \
  --priority medium \
  --summary "Provide an HTTP endpoint to revoke active user sessions." \
  --requirement "Add DELETE /api/v1/sessions/:id controller" \
  --requirement "Validate caller ownership of session" \
  --criterion "Controller returns 200 on successful revocation" \
  --criterion "Unit tests in session_service_test.go cover invalid session ID" \
  --notes "Follow 5-layer architecture and use response.Response[T] wrapper" \
  --file "services/api/internal/present/http/controller/session_controller.go"
```

### Step 4: Present Result to User
Provide the user with:
- The created issue number and clickable URL (`https://github.com/.../issues/<id>`).
- Confirmation of assigned labels (`ready`, priority, type).
- Next steps: the autonomous task runner [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md) will be able to automatically pick up the issue, or you can trigger it on demand.

---

## 4. CLI Tooling Reference: `create_issue.sh` / `create_issue.py`

Tool path: `.agents/scripts/create_issue.sh` (or `.agents/scripts/create_issue.py`)

### Common Options:
- `--title <title>`: Issue title (recommended format: `type(scope): summary`).
- `--type <type>`: `feat` (enhancement), `fix` (bug), `refactor`, `docs`, `chore`.
- `--priority <level>`: `high`, `medium` (default), `low`.
- `--summary <text>`: Summary or problem statement.
- `--requirement <item>`: Requirement bullet point (can be repeated).
- `--criterion <item>`: Acceptance criterion checkbox item (can be repeated).
- `--notes <text>`: Architecture or technical implementation notes.
- `--file <path>`: Impacted file or package path (can be repeated).
- `--body-file <path>`: Markdown file containing the full issue body (for complex descriptions).
- `--label <label>`: Additional GitHub label to attach.
- `--no-ready`: Do NOT add `ready` label immediately (leave in backlog for triage).
- `--dry-run`: Preview without creating on GitHub.
- `--json`: Output result as machine-readable JSON.
