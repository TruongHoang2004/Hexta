# Implementation Plan: Comprehensive Local Development Guide

- **Issue**: [#18](https://github.com/TruongHoang2004/Hexta/issues/18)
- **Title**: docs: create comprehensive local development guide ("How to run system in local machine")
- **Date**: 2026-09-12
- **Branch**: `task/issue-18-docs-create-comprehensive-local-developm`
- **Scope**: Root documentation (`README.md`), environment configuration templates (`.env.example`), infrastructure docker compose fixes (`infrastructure/local-all.yml`, `infrastructure/docker-compose.ui.yml`), frontend build fix (`apps/web/app/auth/callback/page.tsx`), and agentic memory records.

---

## 1. Overview & Goal
Establish a production-grade, authoritative local development guide in root `README.md` titled **"How to run system in local machine"** for the Hexta monorepo.
The guide provides a seamless developer onboarding experience by clearly documenting:
- System prerequisites and tooling versions (macOS/Linux/WSL2, Docker, Go, Node.js, pnpm, Atlas CLI, Make).
- Architecture topology and comprehensive service port mapping (Postgres 5433, Redis 6379, Elasticsearch 9200, MinIO 9000/9001, Kafka 9092, Qdrant 6333, RedisInsight 5540, Go API Gateway 8080, Next.js Apps 3000-3003).
- Dual execution pathways:
  - **Mode A (Full Docker Quickstart)**: Run the backend stack in Docker via `make local-up`.
  - **Mode B (Hybrid Developer Workflow)**: Run shared infra in Docker (`make infra-up`), apply Atlas migrations (`make migrate-apply svc=api`), run backend locally with live debugging (`cd services/api && go run cmd/main.go` or `make debug`), and start Next.js frontend applications (`pnpm install && make ts-dev filter=web`).
- Environment configuration templates (`.env.example`) across infrastructure, backend, and frontend.
- Verification procedures (health check endpoints, Swagger UI, unit/build tests).
- Common troubleshooting scenarios and FAQs (port collisions, migration drift, volume resets, OAuth callbacks).

---

## 2. Current State Analysis
- **Missing Root Documentation**: The repository currently lacks a top-level `README.md`. Onboarding instructions are non-existent or fragmented in legacy template files (`infrastructure/README.md`).
- **Environment Gaps**: Developers had no reference `.env.example` templates for `infrastructure/`, `services/api/`, or `apps/web/`.
- **Infrastructure Compose Inconsistencies**: `infrastructure/local-all.yml` referenced a non-existent `docker-compose.infra.yml`, outdated paths (`service/api` instead of `services/api`), and a mismatched network (`ecommerce_network` vs `hexta_network`).
- **Frontend Prerender Boundary**: `apps/web/app/auth/callback/page.tsx` lacked a `<Suspense>` boundary around `useSearchParams()`, causing static build failures during `pnpm run --recursive build`.
- **Language & Standards Compliance**: All documentation and code comments must adhere strictly to English per `GEMINI.md`.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Environment Templates & Infrastructure Fixes
- Create `.env.example` files:
  - `infrastructure/.env.example` with PostgreSQL and MinIO default credentials.
  - `services/api/.env.example` with JWT secrets, OAuth placeholders, and AI configuration.
  - `apps/web/.env.example` with `NEXT_PUBLIC_API_URL`.
- Align `infrastructure/local-all.yml` and `infrastructure/docker-compose.ui.yml` with `docker-compose.yml` (`hexta_network`, `services/api/Dockerfile`).
- Fix `apps/web/app/auth/callback/page.tsx` with `<Suspense>` boundary and English localization to ensure clean workspace builds.

### Step 3.2: Technical Design Document (`dev-design`)
- Author `agentic-memory/designs/2026-09-12_issue-18-local-dev-guide_design.md` covering:
  - Architecture component diagram and request flow.
  - Port assignment and network connectivity matrix.
  - Lifecycle state machine for local runtime modes.
  - Secrets and environment variable hierarchies.
  - Failure recovery workflows.

### Step 3.3: Author Root `README.md`
- Craft the complete local development documentation with:
  1. Monorepo Overview & High-Level Architecture (Go + Gin + Next.js + PostgreSQL + Redis + MinIO + Elasticsearch + Kafka + Qdrant).
  2. System Prerequisites & Installation (Go 1.24+, Node 20+, pnpm 10+, Docker Compose v2+, Atlas CLI).
  3. Service Architecture & Port Reference Matrix.
  4. Mode A: Full Docker Quickstart (`make local-up` / `make local-down`).
  5. Mode B: Hybrid Local Development Workflow (Infra -> Migrations -> Backend -> Frontend).
  6. Environment Variables Configuration (`.env.example` to `.env` walkthrough).
  7. Verification, Health Checks & Swagger Docs (`/api/v1/health`, `/api/v1/swagger/index.html`).
  8. Testing & Quality Commands (`go test ./...`, `pnpm run --recursive build`).
  9. Troubleshooting & Common Gotchas (port 5432 vs 5433, Atlas hash recalculation, Docker clean reset, Google OAuth callback configuration).

### Step 3.4: Verification & Code Review (`dev-review`)
- Validate markdown links and structure.
- Verify syntax of all Docker compose files and Make targets.
- Execute full test suites (`go test ./services/api/...`, `pnpm run --recursive build`).
- Save review report to `agentic-memory/reviews/2026-09-12_issue-18-local-dev-guide_review.md`.

### Step 3.5: Changelog (`dev-changelog`) & Pull Request Creation
- Record detailed changelog in `agentic-memory/changelogs/2026-09-12_issue-18-local-dev-guide_changelog.md`.
- Commit changes using Conventional Commits (`docs: create comprehensive local development guide (refs #18)`).
- Push branch `task/issue-18-docs-create-comprehensive-local-developm`.
- Open PR with `Closes #18` and links to memory artifacts.
- Transition issue to `in-review` and post completion comment.

---

## 4. Risk Assessment & Edge Cases
| Risk / Edge Case | Impact | Mitigation Strategy |
|---|---|---|
| **Port Conflicts (e.g. Postgres 5432 vs 5433)** | Local Postgres on developer machine blocks Docker or local service | Docker Compose maps host `5433` -> container `5432`; document explicit connection string in guide. |
| **Atlas Migration Checksum Mismatch** | Schema changes cause Atlas to reject migration execution | Document `make migrate-hash svc=api` and step-by-step resolution. |
| **Missing Environment Secrets** | API fails to start with empty JWT secrets or OAuth config | Supply complete `.env.example` templates and mention default fallback behavior. |
| **Monorepo Build Breakages** | Client components breaking build on static generation | Verified and resolved `useSearchParams()` `<Suspense>` requirement in `apps/web`. |

---

## 5. Definition of Done (DoD)
- [x] Pre-flight fixes applied (`.env.example` templates, `local-all.yml` validation, `Suspense` fix in `apps/web`).
- [x] All Go tests passing (`go test ./services/api/...`).
- [x] All Next.js/TypeScript builds passing (`pnpm run --recursive build`).
- [ ] Technical design document created in `agentic-memory/designs/2026-09-12_issue-18-local-dev-guide_design.md`.
- [ ] Comprehensive root `README.md` created adhering to English language and project standards.
- [ ] Code review report created in `agentic-memory/reviews/2026-09-12_issue-18-local-dev-guide_review.md`.
- [ ] Changelog created in `agentic-memory/changelogs/2026-09-12_issue-18-local-dev-guide_changelog.md`.
- [ ] Branch committed, pushed, PR created closing #18, and issue labeled `in-review`.
