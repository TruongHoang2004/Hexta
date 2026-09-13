# Implementation Plan: Replace Legacy GitLab References & Template READMEs with Project Standards

**Date**: 2026-09-13  
**Target Issue**: #21 - docs: replace legacy gitlab references and template readmes with project standards  
**Feature Branch**: `task/issue-21-docs-replace-legacy-gitlab-references-an`  

---

## 1. Overview & Goal
The Hexta codebase was migrated from an older GitLab repository (`gitlab.com/ecommercehub1/*`). During migration, several legacy references remained:
- Core governance files (`GEMINI.md`, `agentic-memory/rules/development-rules.md`, `.agents/skills/dev-design/SKILL.md`) still referenced old GitLab package import paths (`gitlab.com/ecommercehub1/shared/pkg/errors` or `lib/pkg/errors`) and deprecated monorepo folder hierarchies (`backend/service/` and `backend/lib/`).
- Multiple subdirectories (`packages/shared`, `infrastructure`, `test`) contained default, auto-generated GitLab README templates with broken links (`docs.gitlab.com`, `gitlab.com/ecommercehub1/*`).
- The API service README (`services/api/README.md`) referenced a legacy clone URL (`gitlab.com/ecommercehub1/api.git`).

The goal is to systematically update project governance rules, documentation, and module paths to reflect current Hexta standards, replace all boilerplate GitLab templates with production-grade documentation, and eliminate legacy GitLab references from project documentation and guidance files.

---

## 2. Current State Analysis
1. **Core AI Guidance & Rules**:
   - `GEMINI.md`:
     - Names the project "CommerceHub" instead of "Hexta".
     - Specifies directory structure as `backend/service/` and `backend/lib/` (actual directories are `services/` and `packages/`).
     - References `gitlab.com/ecommercehub1/lib/pkg/errors` for error handling.
   - `agentic-memory/rules/development-rules.md`:
     - References "Hexta / CommerceHub".
     - References `gitlab.com/ecommercehub1/shared/pkg/errors (or lib/pkg/errors)`.
   - `.agents/skills/dev-design/SKILL.md`:
     - Specifies error code mappings using `gitlab.com/ecommercehub1/shared/pkg/errors`.
2. **Template READMEs**:
   - `packages/shared/README.md`: Generic GitLab boilerplate (94 lines of GitLab markdown).
   - `infrastructure/README.md`: Generic GitLab boilerplate (94 lines of GitLab markdown).
   - `test/README.md`: Generic GitLab boilerplate (94 lines of GitLab markdown).
   - `services/api/README.md`: Contains `git clone https://gitlab.com/ecommercehub1/api.git`.
   - `Makefile`: Line 1 mentions "CommerceHub Backend".

---

## 3. Task Breakdown (Step-by-Step)

### Phase 1: Technical Design (`dev-design`)
- Formulate technical documentation design document:
  - Document taxonomy and structure standards for monorepo packages, infrastructure, and tests.
  - Package breakdown for `packages/shared` (`common`, `errors`, `logger`, `telemetry`, `validator`, `proto`).
  - Infrastructure stack specification (PostgreSQL, Redis, Atlas migrations, Docker Compose environments).
  - Test harness specification (E2E API test suites, stress testing, runner commands).
- Save to `agentic-memory/designs/2026-09-13_issue-21-clean-legacy-gitlab-references_design.md`.

### Phase 2: Core AI Rules & Architecture Standards Update
- Update `GEMINI.md`:
  - Replace "CommerceHub" with "Hexta".
  - Correct monorepo structure: `services/` for microservices and `packages/` for shared libraries.
  - Update error handling path: `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
- Update `agentic-memory/rules/development-rules.md`:
  - Align project naming to Hexta.
  - Update error package import path to `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
- Update `.agents/skills/dev-design/SKILL.md`:
  - Update error package import reference to `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
- Update root `Makefile`:
  - Update title comment to "Main Makefile for Hexta Platform".

### Phase 3: Comprehensive Documentation Overhaul
- Rewrite `packages/shared/README.md`:
  - Package overview, module architecture, and Go module identity (`github.com/TruongHoang2004/Hexta/packages/shared`).
  - Component documentation: `logger` (Zap wrapper), `errors` (structured domain errors), `telemetry` (OpenTelemetry tracing/metrics), `validator` (custom validation tags), `proto` (identity v1 gRPC/Protobuf contracts).
  - Installation and usage examples for consumer services.
- Rewrite `infrastructure/README.md`:
  - Infrastructure architecture: Docker Compose setup, PostgreSQL, Redis, Atlas migration tool, ELK/monitoring.
  - Environment configurations (`docker-compose.yml`, `docker-compose.migrate.yml`, `docker-compose.ui.yml`, `local-all.yml`).
  - Common operational commands (`make up`, `make down`, `make build-all`, `make clean`).
  - Database provisioning and schema migration workflows.
- Rewrite `test/README.md`:
  - Testing suite architecture: E2E HTTP integration tests and stress/load testing.
  - Directory layout: `e2e/`, `stress/`, `mock/`, `config/`.
  - Execution instructions: running end-to-end suites against local or containerized services, running load tests with concurrency parameters.
- Update `services/api/README.md`:
  - Update repository clone URL to `https://github.com/TruongHoang2004/Hexta.git`.
  - Align project naming and directory navigation instructions.

### Phase 4: Verification & Code Review (`dev-review`)
- Grep across all markdown, rules, and documentation files to ensure zero remaining occurrences of `gitlab.com`.
- Validate markdown links and syntax formatting.
- Audit against 5-layer architecture rules and project governance requirements.
- Save review report to `agentic-memory/reviews/2026-09-13_issue-21-clean-legacy-gitlab-references_review.md`.

### Phase 5: Changelog Generation (`dev-changelog`)
- Document all file modifications, rationale, and verification steps in `agentic-memory/changelogs/2026-09-13_issue-21-clean-legacy-gitlab-references_changelog.md`.

### Phase 6: Commit, Push, and PR Lifecycle
- Stage all files (`git add .`).
- Commit using Conventional Commits: `docs: replace legacy gitlab references and template readmes with project standards (refs #21)`.
- Push to origin on branch `task/issue-21-docs-replace-legacy-gitlab-references-an`.
- Create Pull Request closing #21 with links to all memory artifacts.
- Update GitHub issue labels: remove `in-progress`, add `in-review`.
- Add completion comment to GitHub issue #21.

---

## 4. Risk Assessment & Edge Cases
- **Risk**: Over-replacing strings in historical changelogs or Go source files that are scheduled for migration under separate issues (e.g. Issue #20).
  - *Mitigation*: Restrict changes specifically to project READMEs, rules, AI guidance docs, and core documentation as scoped by Issue #21.
- **Risk**: Broken markdown links or invalid code snippets in newly written READMEs.
  - *Mitigation*: Rigorously inspect and cross-verify all paths and code examples against real packages and scripts in the repository.

---

## 5. Definition of Done (DoD)
1. Zero occurrences of `gitlab.com` in project READMEs, rules, and AI guidance documents.
2. `GEMINI.md` and `development-rules.md` accurately reflect Hexta monorepo layout and Go package imports.
3. High quality, production-grade READMEs implemented for `packages/shared`, `infrastructure`, and `test`.
4. All 4 agentic-memory artifacts (`plan`, `design`, `review`, `changelog`) recorded and linked.
5. PR created with `Closes #21` and issue #21 labeled `in-review`.
