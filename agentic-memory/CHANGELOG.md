# Consolidated Project Changelog

This document consolidates all historical development changelogs for the Hexta platform into a unified, high-density reference. Individual task changelogs have been synthesized into thematic milestones to optimize readability and context loading.

---

## 2026-10-19 — Resolve PR #53 documentation conflicts

- Merged origin/main (9784695, merged PR #52) into the documentation branch without rewriting published history. Resolved four documentation conflicts by preserving new changelog entries, historical-design notices and the correct API-prefixed Swagger URL alongside the implementation baseline.
- Added a dedicated EPIC-001 conflict-resolution session and refreshed its handoff/architecture base evidence. Documentation PR #53 now targets main; parent dependency is satisfied, but documentation completion remains false until its own confirmed merge.
- Verification: no unmerged paths/markers, Markdown link/fence and metadata consistency, whitespace and remote PR mergeability checks. Runtime code is unchanged; no runtime test claim added.

---

## 2026-10-19 — Canonical agent rules and epic-scoped sessions

- **Summary:** Moved all nine engineering-rule sections from GEMINI.md to root AGENTS.md. GEMINI.md is now only a read-AGENTS.md instruction; AGENTS.md is the canonical automatic entry point and contains the complete shared rules. The canonical rules describe epic/session startup and maintenance.
- **Memory layout:** Added eight epic directories with individual context/acceptance/remaining-work READMEs, separate dated session records, explicit status/completed fields, a status index and session template. Historical PR #51 bootstrap is done based on owner-confirmed merge; PR #52 slices retain last-known review status, and future feature epics remain incomplete. A session's completion does not complete its parent epic.
- **Navigation:** Replaced the global SESSION_CONTEXT.md with a short legacy redirect and updated root/memory README and system-design references. Shared architecture modules and the consolidated changelog remain references, not a combined session context. Preserved earlier design work and unrelated files.
- **Verification:** Documentation-only: check migrated nine-section rule content, epic/index/session metadata and completion consistency, local navigation links, redirect-only GEMINI.md, and git whitespace. No application runtime changes or new runtime test claims. Delivery is being committed on codex/system-design-epic-memory as a documentation-only PR stacked on the verified-open PR #52 branch; review/merge remain pending.

---

## 2026-10-18 — Canonical system design and session handoff

- **Summary:** Consolidated the complete system design into a dated index and eight focused modules: product decisions, component ownership, actual/target DB, implemented/target HTTP, exact Protobuf/AI lifecycle, Kafka events, workflows/concurrency, and operations/security/roadmap. Existing migration CREATE TABLE definitions and actual router/Protobuf code were inspected; approved requirements and unimplemented contracts are explicitly labeled.
- **Navigation:** New root AGENTS.md and agentic-memory/SESSION_CONTEXT.md direct future sessions to the canonical pack; GEMINI.md, root README and memory README link the same entry. Historical dated designs remain intact with a supersession notice. The handoff records branch/PR baseline, actual verification limits and open product/deployment inputs without relying on chat history.
- **Corrections:** Documented current Swagger route under /api/v1/swagger/index.html and corrected the local runbook. Clarified that PostgreSQL runs are whole-graph replay, memory adapter is test-only, Go conversation worker/Kafka/providers/checkpoints/Langfuse remain absent, and legacy order AI drafts are not messaging proposals. Memory files are persistent repository context, not magically auto-loaded Markdown.
- **Verification:** Documentation-only work: compare physical catalog to migration blocks, route inventory to router, wire field tags to Protobuf and runtime semantics to store/worker; check local Markdown links, duplicate headings/fences, secret-file exclusions and git whitespace. No new runtime test outcome is claimed; baseline 22-test/Docker evidence is historical.

---

## 2026-10-18 — Durable AI runs, complete local Docker profile and Kafka draft

- **Summary:** Runtime gRPC now commits accepted simulation runs in PostgreSQL and executes them with renewable fenced leases. Added one-command local web/API/AI/PostgreSQL/Redis startup and a draft Kafka use-case design requested by the owner.
- **Files:** New AI `store.py`, `durable.py`, `migrations/001_runs.sql`, Docker `local_smoke.py` and real PostgreSQL tests. Updated app readiness/RPC factory, locked psycopg runtime, test image target, Compose schema initialization/Atlas/API/web/test services, API local config, web public-URL build arg, root build-context excludes and `scripts/local.sh`. Fixed Go cache bootstrap's missing `service.IDraftCache` provider and added Fx graph regression test. Runbook: `infrastructure/README.demo.md`. Designs/plans: durable AI runs and Kafka event backbone under `agentic-memory/designs/`, corresponding local-stack plan.
- **Decisions/walkthrough:** StartRun transaction inserts tenant/run identity plus immutable request digest before acknowledgement. A worker uses SKIP LOCKED and short transactions; request execution occurs outside DB locks. Renewal/completion require the current unexpired lease token; cancelled/finished work cannot be overwritten. An expired attempt can be recovered up to three attempts; errors are sanitized. Graceful shutdown releases owned work without spending crash-recovery budget. Python role/schema is separate from Atlas-managed public tables; Atlas public scope and revision schema are explicit. The old memory adapter is retained for offline tests only. This is whole-run deterministic replay, not LangGraph node checkpoints or exactly-once provider effects.
- **Kafka draft:** Use cases cover ingress, conversations/takeover/config/scripts, AI requests/results, proposals, outbound attempts/receipts, expiry scheduling, independent orders and projections. Includes topic keys, outbox/inbox dedup, versioning, quotas, retry/DLQ/replay, retention/PII, local versus production topology and staged acceptance tests. Owner request supersedes prior optional-broker direction; no Kafka producer/consumer is implemented in this slice.
- **Verification:** Docker AI test target: 22 tests passed including real PostgreSQL concurrent acceptance/claim, changed-input conflict, cross-tenant lookup, expired-lease recovery and old-token fencing, persisted cancellation, exhaustion and replacement-worker execution. Existing offline suite: 21 passed (DB test intentionally skipped without explicit DSN). Go bootstrap test and generated gRPC/smoke packages passed; `go vet ./internal/bootstrap` passed. Full Docker build/start completed with web/API/AI/PostgreSQL/Redis healthy and init/Atlas jobs exit 0. Docker-only smoke passed web login page, API DB/cache health, register/login and authenticated tenant create/list. Accepted run and idempotent retry survived an actual AI container restart. Web production build/TypeScript passed. Shell syntax and whitespace checks passed.
- **Limits:** Python provider remains fake, legacy Go AI draft URL disabled locally; Go conversation dispatcher/send gate, real Pancake, node checkpoints, real provider adapters, Langfuse and Kafka implementation remain pending. Local secrets are generated in ignored `.env.local`, no paid API calls/customer sends. The stack is a local built-app profile, not hot reload or a public VPS deployment. Application containers remain running; ephemeral integration-test DB is stopped.

---

## 2026-10-18 — AI bootstrap review fixes

- **Summary/files:** Updated `services/ai/hexta_ai/graph.py`, `rpc.py` and HTTP/gRPC regression tests to address the supplied review.
- **Behavior:** Strip whitespace before validating message/script fields; whitespace-only values now fail HTTP 422 or gRPC INVALID_ARGUMENT. Construct a complete protobuf candidate privately and replace the stored run only after successful conversion, so serialization failures cannot expose partial text. Preserve cancellation if an adapter raises after cancellation. Synchronize graph-level and result-level steps. Keep the task done callback as the sole normal cleanup path, including tasks cancelled before their coroutine starts.
- **Verification:** 21 Python tests passed, including parameterized whitespace rejection over both transports, consistent graph steps, and an injected protobuf serialization failure after valid candidate fields. Existing authentication/idempotency/cancellation cases remain green. Rebuilt Docker stack is healthy; HTTP smoke and real Go-to-Python gRPC smoke passed after the fixes. The unrelated pre-existing `.gitkeep` change is preserved.

---

## 2026-10-18 — Go/Python gRPC simulation boundary

- **Summary:** Implemented owner-approved internal gRPC with StartRun/GetRun/CancelRun; HTTP health/preview and frontend HTTP/SSE remain separate. This is a verified simulation boundary, not the business-worker integration.
- **Files:** [NEW] `proto/hexta/ai/v1/service.proto`, `proto/README.md`, generated Python and Go bindings, `services/ai/hexta_ai/rpc.py`, `services/ai/tests/test_rpc.py`, `services/api/cmd/ai-smoke/main.go`; [MODIFY] AI app lifespan, runtime dependencies/uv lock, Dockerfile, Compose gRPC loopback mapping, READMEs, HEX-8/9 designs and HEX-10 plan.
- **Decisions/walkthrough:** StartRun validates the authorized snapshot, hashes the request under `(tenant_id, run_id)` and schedules one async graph task; same ID with changed input returns ALREADY_EXISTS. GetRun/CancelRun authenticate metadata and qualify tenant lookup. Cancellation suppresses even a late candidate from a runner that catches cancellation. Result echoes context/control revisions for the future Go acceptance gate. Errors are sanitized. A bounded 1000-run single-process RAM registry exercises the contract; restart loses results/idempotency and capacity includes finished runs. Durable acceptance, leases/checkpoints, provider adapters, trace exporter and business send policy remain pending. Plaintext gRPC is local/private only, not a public VPS profile.
- **Verification:** 11 Python tests passed (`cd services/ai && .venv/bin/python -m pytest -q`), covering authentication, validation, tenant-qualified lookup, request idempotency/conflict, completion, capacity, sanitized failure, cancellation and late output suppression. `go test ./internal/infrastructure/ai/rpc ./cmd/ai-smoke` compiled both packages. Docker build/start passed with all three containers healthy; actual `go run ./cmd/ai-smoke` passed Go/Python auth/idempotency/candidate interoperability. HTTP smoke still passed. Lock check and Python compile passed. Initial runtime build caught missing protobuf in production dependencies; added it explicitly and rebuilt successfully. No paid inference or customer messages.
- **Verification boundary:** Full-repo whitespace check reports an existing trailing blank line in the user's `graduation thesis/assets/.gitkeep`; that unrelated change was preserved. No full-app test result is claimed. HEX-10 remains in progress.

---

## 2026-10-18 — HEX-10 local AI environment bootstrap

- **Summary:** Added an isolated local Compose stack for PostgreSQL, Redis and a fake-only FastAPI/LangGraph service. Existing Go/web configs and user changes were preserved; no provider keys, hosted tracing or real sending are used.
- **Files:** `services/ai/` (project/uv lockfile, graph, protected preview endpoint, Dockerfile, tests and HTTP smoke script); `infrastructure/compose.demo.yml`; `infrastructure/README.demo.md`; `.gitignore` excludes Python environments/test caches. Plan: `agentic-memory/plans/2026-10-18_hex-10-local-ai-bootstrap_plan.md`.
- **Decisions:** Preview is a synchronous environment probe, not the durable runs API. Only an unambiguous active exact-match script renders; unknown/ambiguous or unresolved-variable cases return human handoff. Explicit fake mode prevents accidental real inference, and tracing is disabled around the simulation graph. Containers bind localhost ports 55433/56379/58000, use a separate Compose project/DB volume and leave the database initially empty.
- **Verification:** Python 3.14.5: `services/ai/.venv/bin/python -m pytest -q` passed 6 tests. `uv lock --check --offline` passed. Compose config validation and `up --build -d --wait` passed; all three containers healthy. `python3 services/ai/smoke.py` passed over real container HTTP for simulated readiness, authenticated script preview and unauthenticated rejection. No paid/provider requests were made.
- **Limits:** Go/web, business migrations, provider adapters, durable AI jobs/checkpoints, Pancake event/send simulation and Langfuse are not integrated in this stack yet. HEX-10 stays in progress. The Compose file is local-only and is not a VPS deployment profile. Docker Desktop was started for verification; the three local demo containers remain running.

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

### [#35] Inventory Repository, Stock Reservation, and Movement Tracking
- **Date**: 2026-10-09 | **Scope**: `services/api` | **PR / Issue**: [#35](https://github.com/TruongHoang2004/Hexta/issues/35)
- **Summary**: Implemented the Inventory Management System (IMS) persistence and service layers with atomic two-phase stock reservations and movement tracking.
- **Key Technical Decisions**:
  - **Atomic Conditional Updates**: Implemented `WHERE tenant_id = ? AND variant_id = ? AND available_qty >= ?` atomic conditional SQL updates to prevent race conditions and overselling.
  - **Immutable Stock Movement Ledger**: Automatically recorded outbound fulfillment and adjustment movements within the same transaction.
  - **Uber Fx Injection**: Provided `IInventoryRepository` and `IInventoryService` in `internal/bootstrap/`.
- **Impacted Files**: `services/api/internal/repository/inventory_repository*.go`, `services/api/internal/core/service/inventory_service*.go`, `services/api/internal/bootstrap/{repository,service}.go`.
- **Verification**: `go test -v ./services/api/internal/core/service/... ./services/api/internal/repository/...`.

### [#36] Order State Machine, Pricing, and REST Endpoints
- **Date**: 2026-10-09 | **Scope**: `services/api` | **PR / Issue**: [#36](https://github.com/TruongHoang2004/Hexta/issues/36)
- **Summary**: Implemented the Order Management System (OMS) service, repository, and REST endpoints enforcing the order lifecycle state machine with stock reservation coordination and decimal pricing.
- **Key Technical Decisions**:
  - **FSM State Invariants**: Enforced lifecycle state machine (`Draft` -> `Confirmed` -> `Processing` -> `Fulfilled` / `Cancelled`) with invalid transitions returning HTTP 400 Bad Request.
  - **IMS Coordination**: Integrated with `IInventoryService` to reserve stock upon order confirmation, release upon cancellation, and permanently deduct with outbound movement records upon fulfillment. Guaranteed atomic rollback if any item fails reservation.
  - **Monetary Precision**: Applied `decimal.Decimal` across line totals, subtotal, discount, and total amount.
  - **REST Endpoints & Swagger**: Added `POST /api/v1/orders`, `GET /api/v1/orders`, `GET /api/v1/orders/:id`, `PATCH /api/v1/orders/:id/status`, and `PATCH /api/v1/orders/:id/payment` with full Swagger annotations and `response.Response[T]` DTO wrappers.
- **Impacted Files**: `services/api/internal/repository/order_repository.go`, `services/api/internal/core/service/order_service*.go`, `services/api/internal/present/http/dto/order_dto.go`, `services/api/internal/present/http/controller/order_controller.go`, `services/api/internal/present/http/router/router.go`, `services/api/internal/bootstrap/{repository,service,controller}.go`, `services/api/docs/*`, `Makefile`.
- **Verification**: `go test -count=1 -race ./services/api/internal/core/service/...`, `go build ./services/api/...`, `make swagger`.

### [#37] AI Agent Engine with Tool Calling Registry and Draft Proposal Flow
- **Date**: 2026-10-09 | **Scope**: `services/api` | **PR / Issue**: [#37](https://github.com/TruongHoang2004/Hexta/issues/37)
- **Summary**: Implemented the conversational AI Agent service and HTTP SSE handlers in Go connecting to Google Gemini API via official SDK, establishing schema-governed tool calling registry, Redis draft caching (30-min TTL), and Server-Sent Events (SSE) streaming.
- **Key Technical Decisions**:
  - **Zero Direct DB Mutations**: Enforced invariant that AI Agent never executes direct database mutations. All purchase intent is captured as an Order Draft proposal for human review.
  - **Tool Calling Registry**: Defined schema-governed functions (`search_products`, `check_stock_availability`, `propose_order_draft`) with dynamic argument extraction and multi-tenant validation.
  - **Redis Draft Caching**: Cached proposed drafts under `hexta:draft:<tenant_id>:<draft_id>` with 30-minute TTL and unique UUIDs.
  - **Real-Time SSE Streaming**: Implemented `POST /api/v1/ai/agent/converse` emitting structured events (`token`, `tool_call`, `draft_proposed`, `done`).
  - **Multi-Tenant Boundary Enforcement**: Injected and verified `tenant_id` from JWT session context into all tool executions and draft lookups.
- **Impacted Files**: `services/api/internal/infrastructure/gemini/client.go`, `services/api/internal/core/service/ai_agent_service*.go`, `services/api/internal/present/http/dto/ai_dto.go`, `services/api/internal/present/http/controller/ai_controller.go`, `services/api/internal/present/http/router/router.go`, `services/api/internal/bootstrap/{service,controller}.go`, `services/api/docs/*`.
- **Verification**: `go test -race ./...` in `services/api` (All PASSED), `go vet ./...`.

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

### [#29] Comprehensive Unit Test Suites for Critical Core Services and SDK
- **Date**: 2026-10-10 | **Scope**: `services/api`, `packages/sdk` | **PR / Issue**: [#29](https://github.com/TruongHoang2004/Hexta/issues/29)
- **Summary**: Implemented unit and integration test suites for uncovered repository operations and TypeScript SDK clients, eliminating codebase coverage gaps.
- **Key Technical Decisions**:
  - **OrderRepository Test Coverage**: Added `TestOrderRepository_CRUD` verifying transactional order creation, item eager-loading, filterable list queries, and status/payment updates.
  - **SDK Test Expansion**: Expanded `@hexta/sdk` test suite to cover tenant lifecycle methods, member invites, token clears, and API error formatting.
- **Impacted Files**: `services/api/internal/repository/order_repository_test.go`, `packages/sdk/src/index.test.ts`.
- **Verification**: `go test -race ./services/api/... ./packages/shared/...` and `pnpm --filter @hexta/sdk test --run` pass with 100% success; `audit_codebase.py` coverage findings reduced to 0.

### [Docs] Vietnamese graduation thesis LaTeX scaffold
- **Date**: 2026-10-12 | **Scope**: `graduation thesis/`
- **Summary**: Added an editable scaffold following the supplied thesis PDF's cover, preliminary pages, five chapters, bibliography and appendices.
- **Files**: [NEW] `graduation thesis/main.tex`, `graduation thesis/README.md`, `graduation thesis/.gitignore`.
- **Decisions**: Single standalone XeLaTeX source; Vietnamese Unicode and TeX Gyre fonts; centralized cover metadata; optional official logo with placeholder fallback; continuous Arabic page numbering from abstract. Content and references remain clearly marked placeholders. Margins approximate the reference rather than claim institutional compliance.
- **Walkthrough**: Edit metadata, replace placeholders, add `assets/logo.png`, and compile twice to resolve generated lists and cross-references. Examples cover figures, tables, equations and numeric citations.
- **Verification**: Static checks passed for balanced environments, unique labels and all reference/citation keys. Built-in LaTeX compiler validation requested; see session outcome for compilation status. Manual review: compile with XeLaTeX using README instructions and inspect cover, lists and chapter breaks.

### [Docs] Thesis compiler compatibility
- **Date**: 2026-10-12 | **Files**: `graduation thesis/main.tex`, `graduation thesis/README.md`.
- **Fix**: Select UTF-8/T5/Latin Modern under pdfLaTeX; load fontspec only for XeLaTeX/LuaLaTeX. Use explicit TeX font filenames to avoid relying on system font discovery.
- **Verification**: Built-in compiler invoked after edits; initial diagnostic identified missing system font lookup and prompted explicit font file configuration. Final result recorded in chat.

### [Docs] Split thesis into editable LaTeX components
- **Date**: 2026-10-12 | **Scope**: `graduation thesis/`
- **Changes**: Main document now inputs 15 files across `config`, `frontmatter`, `chapters`, `backmatter`, and `appendices`. README documents editing locations, root-file directives, image paths and local compilation.
- **Decision**: Use `input` to preserve the original page-break behavior and keep the existing main.tex entry point. No thesis content changed.
- **Verification**: Expanded component files matched the original source exactly before saving. Built-in compiler was invoked but cannot access project files (`config/preamble.tex` not found in its isolated compilation). Local XeLaTeX/MacTeX is not installed yet, so multi-file compilation remains to be verified locally or on Overleaf.
