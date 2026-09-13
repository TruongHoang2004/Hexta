# GitHub Workflow & Issue Automation Rules

This document outlines the standard process for task management, issue labeling, branching, and automated processing.

---

## 1. Issue Management & Labeling

### Priority Labels
- `priority:high`: Critical bugs, blockers, high-priority features. Processed first.
- `priority:medium`: Standard features and improvements. Processed second.
- `priority:low` or unlabelled: Minor adjustments, polish, tech debt. Processed in chronological order (oldest first).

### Status Labels & Issue Lifecycle
- `ready`: Ticket is fully specified and ready to be picked up by an automated agent or developer. Only issues with this label are eligible for autonomous runner pickup.
- `in-progress`: Assigned when work has begun (runner removes `ready` and adds `in-progress`) to prevent duplicate execution.
- `in-review`: Assigned when code implementation is complete, verification passed, and a Pull Request has been opened for human review (runner removes `in-progress` and adds `in-review`).
- `done` (Closed): Ticket is only considered **Done** once the Pull Request is **merged into `main`**. GitHub will automatically close the issue upon merge due to the `Closes #<number>` directive.
- `blocked`: Marks issues that depend on external prerequisites or user clarification. Skipped by automated runners.

#### Lifecycle State Machine:
```
[New / Backlog Issue] 
         │ (Triage / Add 'ready' label)
         ▼
     [ready] ── (Runner picks task: remove 'ready', add 'in-progress')
         │
         ▼
   [in-progress] ── (Runner implements code, opens PR: remove 'in-progress', add 'in-review')
         │
         ▼
    [in-review] ── (Human reviews & merges PR into main)
         │
         ▼
    [done (Closed)]
```

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
- The scheduled cron job runs periodically (`*/5 * * * *`).
- It executes `.agents/scripts/get_next_issue.sh` and activates the `/github-task-runner` skill.

---

## 5. Codebase Scout & Tech Debt Audit Lifecycle

The repository includes an automated Codebase Scout (`.agents/scripts/audit_codebase.py`) to discover technical debt, security issues, architectural violations, and test gaps.

### Audit Schedule & Triggers
- **Automated Cron**: Executes weekly on Mondays at 03:00 UTC via `.github/workflows/codebase-scout.yml`.
- **Manual Trigger**: Supports on-demand dispatch via GitHub Actions `workflow_dispatch` with custom parameters (`dry_run`, `max_issues`).
- **Local Execution**: Developers or agents can invoke `./.agents/scripts/audit_codebase.sh --dry-run` to preview findings without creating issues.

### Modular Audit Coverage
1. **Security & Secrets**: Detects private keys, API tokens, and credential patterns.
2. **5-Layer Architecture Compliance**: Enforces Go layer boundaries in `services/api` (no DB/GORM in controllers, no HTTP frameworks in core services).
3. **Swagger / OpenAPI Documentation**: Audits controller endpoints for required Swagger doc annotations and `response.Response[T]` wrapper usage.
4. **Technical Debt & Mock Scans**: Flags unresolved `TODO:`, `FIXME:`, `HACK:`, and temporary mock tenant strings.
5. **Test Coverage Gaps**: Identifies critical services and SDK packages lacking unit tests.
6. **Atlas Migration Integrity**: Verifies `atlas.sum` checksum tracking for all SQL migrations.
7. **English Language Rule**: Enforces English-only comments and identifiers per `GEMINI.md`.

### Deduplication & Issue Triage Process
- All generated issues embed a unique semantic fingerprint (`<!-- audit-fingerprint: <id> -->`).
- The deduplication engine queries all open and closed issues via `gh issue list --state all` before creating new tasks.
- Conflicting topics or existing tracked issues are skipped automatically.
- New issues are capped at `--max-issues` (default 5) per execution to avoid backlog spam.

### Audit Archiving
Every run records a complete markdown audit report in `agentic-memory/audits/YYYY-MM-DD_codebase-audit.md` to track health trends and audit metrics over time.


