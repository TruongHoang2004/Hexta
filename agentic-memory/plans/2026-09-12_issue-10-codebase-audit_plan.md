# Implementation Plan: Scheduled Codebase Audit Workflow & Automated Issue Generation

- **Issue**: [#10](https://github.com/TruongHoang2004/Hexta/issues/10)
- **Date**: 2026-09-12
- **Branch**: `task/issue-10-feat-automation-implement-scheduled-code`
- **Scope**: Automation tooling (`.agents/scripts/`), GitHub Actions (`.github/workflows/`), agentic memory (`agentic-memory/audits/`, `agentic-memory/rules/`)

---

## 1. Overview & Goal
Establish an automated, scheduled codebase health inspection and technical debt triage system. The workflow periodically audits the Hexta monorepo for:
- Security vulnerabilities and leaked secrets / credentials.
- Architectural non-compliance with the 5-layer Go architecture (`GEMINI.md`).
- Missing Swagger / OpenAPI annotations on HTTP controllers.
- Code debt (`TODO:`, `FIXME:`, `HACK:`, mock tenant implementations).
- Gaps in unit/integration test coverage.
- Atlas database migration schema integrity (`atlas.sum`).

Findings will be synthesized into structured, prioritized GitHub issues with deduplication against existing open and closed issues, and detailed audit reports archived into `agentic-memory/audits/`.

---

## 2. Current State Analysis
- **Existing Automation**:
  - `.agents/scripts/get_next_issue.py` handles queue polling for tasks, but there is no proactive codebase scanner to discover technical debt and populate issues.
  - No automated GitHub Action exists for repository-wide audits.
- **Standards & Conventions**:
  - `GEMINI.md` and `agentic-memory/rules/development-rules.md` mandate 5-layer Go architecture, strict TypeScript, English-only comments/identifiers, Atlas migrations, and Swagger annotations.
  - Technical debt and manual TODOs can accumulate across `services/api`, `apps/`, and `packages/` without visibility.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Create Codebase Audit Script (`.agents/scripts/audit_codebase.py`)
- Implement modular analyzer classes using Python standard library:
  - `SecretScanner`: High-entropy tokens, private keys, API keys, password assignments.
  - `ArchitectureScanner`: Enforces 5-layer boundaries in Go (`controller` importing DB, `service` importing HTTP/Gin, English-only enforcement).
  - `SwaggerScanner`: Inspects Go controllers for `@Summary`, `@Tags`, `@Router`, `@Success` annotations.
  - `DebtScanner`: Scans for `TODO`, `FIXME`, `HACK`, and hardcoded mock tenant strings.
  - `CoverageScanner`: Flags critical packages in Go and TypeScript that lack corresponding test files (`*_test.go`, `*.test.ts`).
  - `MigrationScanner`: Checks Atlas migrations directory and `atlas.sum` integrity.
- Implement deduplication engine:
  - Fetches open & closed issues via `gh issue list --state all --json number,title,body,labels`.
  - Embeds semantic fingerprint comment `<!-- audit-fingerprint: <id> -->` to identify existing issues even if titles evolve.
  - Skips creating issues if fingerprint or title matches.
- Implement CLI flags:
  - `--dry-run`: Scans and generates report without creating GitHub issues (default: false).
  - `--create-issues`: Creates GitHub issues for new findings.
  - `--max-issues`: Maximum issues to create per run (default: 5).
  - `--output`: Path to write markdown audit report.
- Save audit reports into `agentic-memory/audits/YYYY-MM-DD_codebase-audit.md`.

### Step 3.2: Create GitHub Actions Workflow (`.github/workflows/codebase-scout.yml`)
- Trigger on cron schedule (`cron: '0 3 * * 1'`, weekly Mondays at 03:00 UTC) and `workflow_dispatch` (manual run with input options for dry-run).
- Permissions: `issues: write`, `contents: write`.
- Runs audit script, creates issues for new findings, and commits/pushes new audit reports back to repository.

### Step 3.3: Update GitHub Workflow Rules (`agentic-memory/rules/github-workflow.md`)
- Document Codebase Scout audit lifecycle:
  - Scheduled cadence and on-demand trigger.
  - Issue generation structure and deduplication mechanism.
  - Priority mapping (`priority:high`, `priority:medium`, `priority:low`).
  - Audit report archiving in `agentic-memory/audits/`.

---

## 4. Risk Assessment & Edge Cases
| Risk / Edge Case | Impact | Mitigation Strategy |
|---|---|---|
| **GitHub Issue Flooding** | Too many issues created at once cluttering repository backlog | Implement `--max-issues` cap (default 5 per run) and deduplication against all open/closed issues. |
| **False Positives in Secret Scanner** | Mock credentials in test fixtures flagged as real secrets | Skip test files (`*_test.go`, `*.test.ts`), documentation, `.env.example`, and known dummy patterns. |
| **GitHub CLI Authentication** | `gh` command fails in CI or local environments without auth | Gracefully handle missing `GH_TOKEN` / `GITHUB_TOKEN`, fallback to report-only mode with clear warning. |
| **Monorepo Scale / Execution Time** | Scanning entire monorepo could be slow on large trees | Target explicit source directories (`services/api`, `apps/`, `packages/`, `migrations/`) and exclude `node_modules/`, `.git/`, `.worktrees/`, `dist/`. |

---

## 5. Definition of Done (DoD)
- [x] `.agents/scripts/audit_codebase.py` implemented with all modular analyzers, deduplication, CLI options, and reporting.
- [x] Script passes dry-run execution against the monorepo producing valid JSON/markdown output in `agentic-memory/audits/`.
- [x] GitHub Actions workflow `.github/workflows/codebase-scout.yml` created with schedule, manual dispatch, and correct permissions.
- [x] `agentic-memory/rules/github-workflow.md` updated with audit process documentation.
- [x] Review report (`agentic-memory/reviews/2026-09-12_issue-10-codebase-audit_review.md`) and changelog (`agentic-memory/changelogs/2026-09-12_issue-10-codebase-audit_changelog.md`) created.
- [x] Changes committed, pushed to branch, and PR opened referencing #10.
