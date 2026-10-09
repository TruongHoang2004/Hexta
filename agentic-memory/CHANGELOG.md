# Consolidated Project Changelog

This document consolidates all historical development changelogs for the Hexta platform into a unified, high-density reference. Individual task changelogs have been synthesized into thematic milestones to optimize readability and context loading.

---

## Table of Contents
1. [Security, Authentication & Identity](#1-security-authentication--identity)
2. [Multi-Tenancy & Core Domain](#2-multi-tenancy--core-domain)
3. [Monorepo Architecture & Modernization](#3-monorepo-architecture--modernization)
4. [CI/CD & DevOps Automation](#4-cicd--devops-automation)
5. [Autonomous Agents & Repository Governance](#5-autonomous-agents--repository-governance)
6. [Frontend & Design System](#6-frontend--design-system)


---

## 1. Security, Authentication & Identity

### [#4] Authentication Security Hardening
- **Date**: 2026-09-12 | **Scope**: `services/api`, `apps/web` | **PR / Issue**: [#4](https://github.com/TruongHoang2004/Hexta/issues/4)
- **Summary**: Patched hardcoded JWT secrets, resolved OAuth2 CSRF exposure, and enforced strict session ownership.
- **Key Technical Decisions**:
  - **Dynamic JWT Config**: Replaced static secrets with runtime configuration (`config.AppConfig.JWT`) injected via Uber Fx.
  - **Token Type Isolation**: Added `token_type` claims (`"access"` vs `"refresh"`) to block privilege escalation.
  - **OAuth2 CSRF Defense**: Replaced static state with 32-byte cryptographic random state stored in Redis/memory with 5-minute TTL and atomic single-use consumption (`ValidateAndConsumeOAuthState`).
  - **Session Ownership Guard**: Required caller ownership checks on `/api/v1/auth/logout` before revoking sessions.
  - **Secure Token Delivery**: Prevented token exposure in URL queries by delivering auth tokens via secure cookies (`auth_token` and `HttpOnly` `refresh_token`).
- **Impacted Files**: `services/api/internal/present/http/middleware/auth.go`, `services/api/internal/core/service/auth_service.go`, `services/api/utils/jwt.go`, `apps/web/app/auth/callback/page.tsx`.
- **Verification**: `go test ./services/api/internal/core/service/... ./services/api/internal/present/http/middleware/...`.

### [#5] OAuth Account Linking & Compound Unique Index
- **Date**: 2026-09-12 | **Scope**: `services/api`, `migrations/api` | **PR / Issue**: [#5](https://github.com/TruongHoang2004/Hexta/issues/5)
- **Summary**: Resolved unique identifier collision on OAuth login and enabled multi-provider account linking.
- **Key Technical Decisions**:
  - **Compound Unique Index**: Replaced single-column index on `identities.identifier` with compound index `(provider, identifier)`.
  - **Automated Account Linking**: Associated incoming Google OAuth logins with existing `local` accounts sharing verified email addresses.
  - **Nullable Password**: Made `identities.password` nullable (`*string`), eliminating plaintext dummy passwords (`"oauth2-dummy"`) for OAuth identities.
- **Impacted Files**: `services/api/internal/core/model/identities.go`, `services/api/internal/repository/identify_repository.go`, `migrations/api/*fix_identities_provider_unique_index.sql`.
- **Verification**: Atlas migration applied, verified compound index uniqueness in integration tests.

### [#6] Frontend Auth State & Token Storage Synchronization
- **Date**: 2026-09-12 | **Scope**: `apps/web`, `packages/sdk` | **PR / Issue**: [#6](https://github.com/TruongHoang2004/Hexta/issues/6)
- **Summary**: Unified token storage across cookies and localStorage to eliminate client-server hydration mismatch and stale auth states.
- **Key Technical Decisions**:
  - **Dual Storage Persistence**: Standardized SDK `DefaultBrowserStorage` to read and write to both cookies and localStorage.
  - **SSR Hydration Safety**: Replaced client-side state flipping with `useSyncExternalStore` in `AuthNav` component.
  - **Reactive Auth Sync**: Bound Zustand `useAuthStore` to storage changes for immediate header and navigation updates on login/logout.
- **Impacted Files**: `apps/web/store/useAuthStore.ts`, `apps/web/components/auth-nav.tsx`, `apps/web/app/auth/callback/page.tsx`.
- **Verification**: `pnpm --filter web run build` (8/8 static routes prerendered).

### [#8] SDK Silent Token Refresh Interceptor
- **Date**: 2026-09-12 | **Scope**: `packages/sdk` | **PR / Issue**: [#8](https://github.com/TruongHoang2004/Hexta/issues/8)
- **Summary**: Implemented seamless HTTP 401 recovery in `@hexta/sdk` using an Axios response interceptor with token request queuing.
- **Key Technical Decisions**:
  - **Single Flight Refresh**: Implemented an in-flight promise lock (`isRefreshing`) to ensure concurrent 401s trigger only one refresh call.
  - **Pending Request Queue**: Queued subsequent requests during token refresh and replayed them upon new token arrival.
  - **Circuit Breaker**: Cleared tokens and triggered `onAuthFailure` callback if the refresh endpoint itself returned 401.
- **Impacted Files**: `packages/sdk/src/index.ts`, `packages/sdk/src/index.test.ts`.
- **Verification**: Vitest test suite (`packages/sdk/src/index.test.ts`) passed with 100% interceptor coverage.

---

## 2. Multi-Tenancy & Core Domain

### [#7] Multi-Tenant Architecture & REST APIs
- **Date**: 2026-09-12 | **Scope**: `services/api`, `packages/sdk`, `apps/web` | **PR / Issue**: [#7](https://github.com/TruongHoang2004/Hexta/issues/7)
- **Summary**: Built the core multi-tenant foundation: schema models, Redis caching, RBAC authorization, and dynamic workspace switcher UI.
- **Key Technical Decisions**:
  - **Isolated Tenant Domain**: Created `tenants` and `tenant_members` tables with roles (`owner`, `admin`, `member`).
  - **Cache-Aside Architecture**: Implemented `TenantCacheWrapper` for Redis lookups by ID and slug with automatic cache invalidation on updates.
  - **RBAC Enforcement**: Protected tenant operations via `AuthMiddleware` and role assertions in `TenantService`.
  - **SDK & Web Integration**: Exposed tenant CRUD in `@hexta/sdk` and created workspace management dashboard in `apps/web/app/(dashboard)/tenant`.
- **Impacted Files**: `services/api/internal/core/model/tenants.go`, `services/api/internal/repository/tenant_*.go`, `services/api/internal/core/service/tenant_service.go`, `apps/web/app/(dashboard)/tenant/page.tsx`.
- **Verification**: `go test -v ./services/api/internal/core/service/...`.

### [#34] Core Business Schema Migrations & GORM Models
- **Date**: 2026-10-09 | **Scope**: `services/api`, `migrations/api` | **PR / Issue**: [#34](https://github.com/TruongHoang2004/Hexta/issues/34)
- **Summary**: Implemented relational schema migrations and GORM models for Product Catalog, Inventory, Order, Customer, and Audit/AI domains.
- **Key Technical Decisions**:
  - **Monetary Precision**: Applied `decimal.Decimal` (`numeric(15,2)`) across all prices, subtotals, and discounts.
  - **Tenant Scoping**: Embedded indexed `tenant_id` on all tables for defense-in-depth data isolation.
  - **Two-Phase Inventory Tracking**: Defined `inventory_items` with available, reserved, and on-hand quantities to prevent overselling.
  - **Atlas Integration**: Generated migration `20261009180000_create_core_business_schema.sql` and validated checksums via `atlas migrate validate`.
- **Impacted Files**: `services/api/internal/core/model/{products,inventory,orders,customers,audit_logs,ai_drafts}.go`, `services/api/cmd/tools/main.go`, `migrations/api/20261009180000_create_core_business_schema.sql`.
- **Verification**: `go build ./services/api/...` and `atlas migrate validate --dir file://migrations/api`.

---

## 3. Monorepo Architecture & Modernization

### [#1] Strongly-Typed AI Provider Configuration
- **Date**: 2026-09-12 | **Scope**: `services/api` | **PR / Issue**: [#1](https://github.com/TruongHoang2004/Hexta/issues/1)
- **Summary**: Added strongly-typed configuration support for LLM providers (Gemini, OpenAI, Anthropic) via environment variables and YAML config.
- **Impacted Files**: `services/api/config/config.go`, `services/api/config/config.yaml`.
- **Verification**: `go test -v ./services/api/config/...`.

### [#20] Go Module Path Migration from GitLab to GitHub
- **Date**: 2026-09-12 | **Scope**: Monorepo-wide Go modules | **PR / Issue**: [#20](https://github.com/TruongHoang2004/Hexta/issues/20)
- **Summary**: Migrated all Go module declarations, protobuf bindings, Atlas schema configurations, Swagger docs, and internal imports from `gitlab.com/ecommercehub1/...` to `github.com/TruongHoang2004/Hexta/...`.
- **Key Technical Decisions**:
  - **Monorepo Go Workspace**: Preserved `replace` directives and integrated `./test` into `go.work`.
  - **Namespace Alignment**: Updated protobuf `option go_package` and regenerated identity protobuf Go bindings.
- **Impacted Files**: `go.work`, `packages/shared/go.mod`, `services/api/go.mod`, 35+ Go source files in `services/api`.
- **Verification**: `go test ./...` and `go vet ./...` passed monorepo-wide.

### [#21] Documentation Standardization & Legacy Reference Elimination
- **Date**: 2026-09-13 | **Scope**: Documentation, Makefiles | **PR / Issue**: [#21](https://github.com/TruongHoang2004/Hexta/issues/21)
- **Summary**: Removed legacy GitLab templates and references, standardized project naming to Hexta, and created technical READMEs for `packages/shared`, `infrastructure`, and `test`.
- **Impacted Files**: `GEMINI.md`, `Makefile`, `services/api/README.md`, `packages/shared/README.md`, `infrastructure/README.md`, `test/README.md`.
- **Verification**: Verified zero broken markdown references and validated make targets.

### [#18] Local Development Guide & Monorepo Onboarding
- **Date**: 2026-09-12 | **Scope**: Monorepo root | **PR / Issue**: [#18](https://github.com/TruongHoang2004/Hexta/issues/18)
- **Summary**: Authored comprehensive local development and onboarding guides in `README.md` and `docs/local-development-guide.md`.
- **Key Technical Decisions**:
  - **Execution Modes**: Documented Full Docker Compose vs Hybrid execution (local Go/Node + containerized Postgres/Redis).
  - **Port Reference Matrix**: Mapped all service ports (API: 8080, Web: 3000, DB: 5432, Redis: 6379, MinIO: 9000/9001).
  - **Environment Templates**: Added `.env.example` templates across all sub-apps.
- **Impacted Files**: `README.md`, `docs/local-development-guide.md`, `infrastructure/.env.example`, `services/api/.env.example`, `apps/web/.env.example`.

---

## 4. CI/CD & DevOps Automation

### [#3] GitHub Actions CI/CD Pipeline & Toolchain Alignment
- **Date**: 2026-09-12 / 2026-10-05 | **Scope**: `.github/workflows/`, Dockerfiles | **PR / Issue**: [#3](https://github.com/TruongHoang2004/Hexta/issues/3), [#9](https://github.com/TruongHoang2004/Hexta/pull/9)
- **Summary**: Established production CI/CD workflows for Go backend, Next.js frontend, and multi-service Docker builds with GHCR publishing.
- **Key Technical Decisions**:
  - **Backend CI (`backend-ci.yml`)**: Setup Go via `go-version-file: 'go.work'`, automated race detection (`go test -race`), `go vet`, and Atlas directory hash integrity validation (`atlas migrate validate`).
  - **Frontend CI (`frontend-ci.yml`)**: Configured Node 20 with pnpm 10, cached dependencies, built `@hexta/sdk`, checked ESLint, and compiled Next.js standalone output.
  - **Docker CI/CD (`docker-ci-cd.yml`)**: Matrix Docker Buildx for `services/api`, `apps/web`, and `Dockerfile.migrate` with GitHub Actions cache backend (`type=gha`).
  - **Build Context & Container Optimization**: Excluded `node_modules` from `.dockerignore` (reducing context from 1.83GB to 300KB) and configured `pnpm-workspace.yaml` `onlyBuiltDependencies`.
- **Impacted Files**: `.github/workflows/*.yml`, `services/api/Dockerfile`, `apps/web/Dockerfile`, `.dockerignore`, `pnpm-workspace.yaml`.
- **Verification**: All 6 CI checks passing on GitHub Actions.

---

## 5. Autonomous Agents & Repository Governance

### [#10] Automated Codebase Scout & Tech Debt Audit
- **Date**: 2026-09-12 | **Scope**: `.agents/scripts/`, `.github/workflows/` | **PR / Issue**: [#10](https://github.com/TruongHoang2004/Hexta/issues/10)
- **Summary**: Implemented scheduled codebase scanning for security risks, Go 5-layer architectural leaks, Swagger compliance, and debt markers.
- **Key Technical Decisions**:
  - **Zero Dependency Python Engine**: Implemented using pure Python 3 standard library.
  - **Semantic Issue Deduplication**: Generates deterministic SHA-256 fingerprints to avoid duplicate issue generation.
  - **Scheduled Workflow**: Configured `.github/workflows/codebase-scout.yml` to run weekly and via `workflow_dispatch`.
- **Impacted Files**: `.agents/scripts/audit_codebase.py`, `.github/workflows/codebase-scout.yml`, `agentic-memory/audits/2026-09-12_codebase-audit.md`.

### [#22] Issue Lifecycle Manager & Parallel State Synchronization
- **Date**: 2026-09-12 | **Scope**: `.agents/scripts/`, `.github/workflows/` | **PR / Issue**: [#22](https://github.com/TruongHoang2004/Hexta/issues/22)
- **Summary**: Automated issue state reconciliation, task lock management, and PR association.
- **Key Technical Decisions**:
  - **Dual-Layer Sync**: Real-time GitHub Actions trigger on PR events + 30-minute cron fallback.
  - **Zombie Task Recovery**: Automatically strips stale `in-progress` locks after 2 hours of inactivity.
  - **Signal Association**: Maps branches and commits (`task/issue-<id>-...`, `Closes #<id>`) directly to issue lifecycle states.
- **Impacted Files**: `.agents/scripts/manage_issue_lifecycle.py`, `.github/workflows/issue-lifecycle-sync.yml`.

### [#26] Autonomous PR Review Agent with Auto-Merge & Escalation
- **Date**: 2026-09-13 | **Scope**: `.agents/scripts/`, `.github/workflows/` | **PR / Issue**: [#26](https://github.com/TruongHoang2004/Hexta/issues/26)
- **Summary**: Autonomous agent evaluating open PR risk, squash-merging low-risk PRs, resolving lockfile conflicts, and escalating critical PRs.
- **Key Technical Decisions**:
  - **Risk Classification Matrix**: Evaluates blast radius (critical paths: auth, migrations, configs, module definitions).
  - **Ephemeral Rebase Worktrees**: Performs conflict resolution in isolated git worktrees (`.worktrees_tmp*`) to protect primary trees.
  - **Automated Escalation**: Labels high-risk PRs with `needs-human-review` and posts idempotent review comments.
- **Impacted Files**: `.agents/scripts/pr_review_agent.py`, `.github/workflows/pr-review-agent.yml`, `.agents/skills/pr-review-agent/SKILL.md`.

### [Gov] Issue Lifecycle Rules Standardization
- **Date**: 2026-09-12 | **Scope**: `.agents/`, `GEMINI.md` | **Task**: Standardize lifecycle rules
- **Summary**: Enforced deterministic state progression (`ready` -> `in-progress` -> `in-review` -> `done`).
- **Key Technical Decisions**:
  - **Runner Constraint**: Autonomous task runners only pick tickets labeled `ready`.
  - **Definition of Done**: A ticket is strictly considered Done only after its Pull Request is merged into `main`.
- **Impacted Files**: `GEMINI.md`, `.agents/scripts/get_next_issue.py`, `agentic-memory/rules/github-workflow.md`.

### [Gov] Issue Creator Skill & Skill Execution Protocol
- **Date**: 2026-10-05 | **Scope**: `.agents/skills/issue-creator/`, `.agents/rules/`, `GEMINI.md`
- **Summary**: Standardized ticket formulation and 4-phase execution sequence (Pre-flight, Context Loading, Execution, Artifact & State).
- **Key Technical Decisions**:
  - **Issue Formulation CLI**: Validates components, acceptance criteria, and DoD before creating GitHub issues.
  - **Execution Protocol**: Mandates pre-flight validation and context loading before any code modification.
- **Impacted Files**: `.agents/skills/issue-creator/SKILL.md`, `.agents/scripts/create_issue.py`, `.agents/rules/skill-execution-protocol.md`, `GEMINI.md`.

---

## 6. Frontend & Design System

### [#32] Default Light Mode with Seamless Dark Mode Toggle Support
- **Date**: 2026-10-05 | **Scope**: `packages/ui`, `apps/web` | **PR / Issue**: [#32](https://github.com/TruongHoang2004/Hexta/issues/32)
- **Summary**: Switched the primary interface theme to Light Mode as the default experience while preserving dark mode functionality with a responsive theme toggle mechanism.
- **Key Technical Decisions**:
  - **Encapsulated Theme Provider & Toggle (`packages/ui`)**: Implemented `ThemeProvider` wrapping `next-themes` and `ThemeToggle` component with Sun/Moon transition icons, client hydration guards, and accessibility attributes.
  - **Tailwind CSS v4 Dark Variant (`apps/web/app/globals.css`)**: Configured `@custom-variant dark (&:where(.dark, .dark *));` to ensure class-based dark mode toggling works with Tailwind utilities. Defined clean slate light theme on `:root` and dark mode properties under `.dark`.
  - **Default Light Theme Configuration (`apps/web/app/layout.tsx`)**: Configured `ThemeProvider` with `attribute="class"`, `defaultTheme="light"`, and `enableSystem={false}` to guarantee light mode renders on initial load.
  - **Interactive Toggles in Headers**: Integrated `ThemeToggle` into both `AuthNav` and the workspace dashboard header in `tenant/page.tsx`.
- **Impacted Files**: `packages/ui/src/components/theme-provider.tsx`, `packages/ui/src/components/theme-toggle.tsx`, `packages/ui/src/index.ts`, `apps/web/app/globals.css`, `apps/web/app/layout.tsx`, `apps/web/components/auth-nav.tsx`, `apps/web/app/(dashboard)/tenant/page.tsx`.
- **Verification**: `pnpm --filter "./packages/*" run build`, `pnpm --filter web run lint`, and `pnpm --filter web run build` all pass with zero errors.


### [Task] Changelog: Health Controller 5-Layer Architecture Boundary Resolution
- **Date**: 2026-10-09 | **Scope**: `packages/`, `apps/`

### [Task] Changelog: Repository Layer Test Coverage Expansion
- **Date**: 2026-10-09 | **Scope**: `packages/`, `apps/`
