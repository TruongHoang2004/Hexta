# Code Review: Codebase Audit Workflow & Automated Issue Generation

- **Issue**: [#10](https://github.com/TruongHoang2004/Hexta/issues/10)
- **Date**: 2026-09-12
- **Branch**: `task/issue-10-feat-automation-implement-scheduled-code`
- **Scope**: `.agents/scripts/audit_codebase.py`, `.agents/scripts/audit_codebase.sh`, `.github/workflows/codebase-scout.yml`, `agentic-memory/rules/github-workflow.md`, `agentic-memory/audits/`

---

## 1. Overview
This review evaluates the automated codebase audit tool and scheduled GitHub Actions workflow implemented to identify technical debt, security issues, 5-layer Go architectural violations, Swagger annotation gaps, and missing tests across the Hexta monorepo.

### Reviewed Files
- `.agents/scripts/audit_codebase.py` [NEW]
- `.agents/scripts/audit_codebase.sh` [NEW]
- `.github/workflows/codebase-scout.yml` [NEW]
- `agentic-memory/rules/github-workflow.md` [MODIFIED]
- `agentic-memory/audits/2026-09-12_codebase-audit.md` [NEW]
- `.gitignore` [MODIFIED]

---

## 2. The Good
- **Zero-Dependency Portability**: `.agents/scripts/audit_codebase.py` uses only Python standard libraries (`dataclasses`, `hashlib`, `re`, `subprocess`, `argparse`, `pathlib`). It runs out of the box in CI and local machines without requiring pip packages.
- **Robust Multi-Domain Analyzers**:
  - `SecretScanner`: Identifies credentials, AWS keys, private keys, and JWT assignments while avoiding test/placeholder false positives.
  - `ArchitectureScanner`: Enforces 5-layer Go architecture boundaries (e.g. flagging direct GORM imports or DB references in controllers like `health_controller.go`).
  - `SwaggerScanner`: Verifies presence of OpenAPI comments and checks `response.Response[T]` wrapper compliance.
  - `DebtScanner`: Detects unaddressed `TODO`, `FIXME`, and hardcoded mock tenant IDs.
  - `CoverageScanner`: Detects services and SDK packages lacking unit tests.
  - `MigrationScanner`: Checks `atlas.sum` integrity against SQL migration files.
  - `LanguageScanner`: Enforces GEMINI.md Rule 1 (English-only code and comments).
- **Intelligent Deduplication**:
  - Automatically queries all open and closed issues via `gh issue list --state all`.
  - Employs semantic fingerprints (`<!-- audit-fingerprint: <id> -->`) and title keyword conflict resolution. In dry-run verification, it properly recognized existing issues #6 (English language enforcement) and #7 (multi-tenant database models) and skipped duplicate candidate creation.
- **Automated CI/CD Integration**: `.github/workflows/codebase-scout.yml` supports both scheduled cron execution (weekly on Mondays at 03:00 UTC) and manual on-demand triggers with configurable issue limits.
- **Audit Persistence**: Every run archives a timestamped markdown report to `agentic-memory/audits/` tracking repository code health over time.

---

## 3. Critical Issues (Bugs & Security)
- **None Identified**:
  - No secrets or credentials hardcoded in scripts or workflows.
  - CLI operations gracefully handle error conditions and missing GitHub tokens.
  - Monorepo ignore paths properly exclude heavy build directories (`node_modules`, `.next`, `.turbo`, `.git`, `dist`).

---

## 4. Suggestions & Improvements
1. **CI Issue Rate Limiting**: The default `--max-issues 5` limit prevents backlog spamming if a large refactor introduces multiple warnings. This default is appropriate and configurable via `workflow_dispatch`.
2. **Govulncheck / pnpm audit Extension**: Future enhancements can invoke external security linters (`govulncheck`, `pnpm audit`) if binaries are detected in the environment.
3. **Artifact Archiving**: Workflow commits audit reports directly to the repository with `[skip ci]` to prevent infinite CI loops.

---

## 5. Verification Result
- Script dry-run verified successfully against the Hexta monorepo:
  `python3 .agents/scripts/audit_codebase.py --dry-run` executed without errors, discovered 39 findings, synthesized 5 issue candidates, skipped 2 duplicates, and archived report to `agentic-memory/audits/2026-09-12_codebase-audit.md`.
- Status: **APPROVED**.
