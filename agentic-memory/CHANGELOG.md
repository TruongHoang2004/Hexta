# Consolidated Project Changelog

## 2026-10-28 — Focused rules with a single selection router

- [rules/README.md](../rules/README.md) now selects modules by paths and behavior changes; no mandatory inventory/full-catalog or recursive link loading. Split combined HTTP, Fx/infrastructure and errors/auth/logging modules; preserved 91 API IDs/origins, moved cache implementation detail to design context.
- Reduced AGENTS to universal/startup/session instructions and selective routing; extracted approved frontend/delivery details into conditional rule files. API catalog README is optional metadata. Full-codebase audits retain coverage through sequential module review, not simultaneous full-rule context.
- [Design](ARCHIVE_INDEX.md#pruned-2026-10-28-selective-rule-routing-design) and [session](ARCHIVE_INDEX.md#pruned-2026-10-28-selective-rule-routing) document selection/limits. Verified rule uniqueness/counts, links/fences, router coverage and whitespace; no application tests/build/migration, hook/code or live CI/merge changes. Parent state unchanged.

## 2026-10-28 — Aggregate mappers, cache redesign and root rules

- Moved the single API catalog to [rules/](../rules/README.md), rebased links and added AGENTS navigation. Added mapper ownership/purity/boundary rules and shared-manager/aggregate-policy cache design; current catalog v0.2 has 91 IDs. Preserved earlier counts as historical snapshots with follow-up notes.
- [Design](designs/2026-10-28_aggregate-mappers-cache_design.md) and [session](ARCHIVE_INDEX.md#pruned-2026-10-28-mappers-cache-root-rules) distinguish owner decisions from proposed packages/protocol. Cache proposal covers DB-only authority, bounded stale display reads, after-commit invalidation/old aliases, fill fencing, outage/crash limits and measurement before benefit claims. No new wrappers, runtime code, schema or hook implementation.
- Verification: documentation links/fences, rule ID/count/origin consistency, old catalog removal and whitespace. No application tests, benchmark, provider calls, migration apply, live CI or merge verification. Parent state/completion unchanged; unrelated local work preserved.

## 2026-10-28 — API rule catalog for owner review

- Added [modular API rules](../rules/api-service/README.md), [design](ARCHIVE_INDEX.md#pruned-2026-10-28-api-rule-catalog-design) and [EPIC-001 session](ARCHIVE_INDEX.md#pruned-2026-10-28-api-rule-catalog). 73 unique IDs across 11 catalog files: 40 existing requirements, 18 observed patterns and 15 proposed refinements.
- Derived layer/stack/pattern inventory from current API and selected early Git history; specified persistence-only repository responsibilities, SQL-free services, controller validation, explicit contracts, Atlas-only migrations/no runtime DDL, Fx/config/security/cache/error rules and future full-source assessment evidence. Recorded current deviations and owner-review choices. AGENTS remains canonical; no hook or runtime change.
- Verification: document IDs/counts, relative/source links, balanced fences and whitespace only. No application test/build/schema apply/remote CI rerun or new merge/completion claim. Existing unrelated local work preserved; draft awaits owner review.

## 2026-10-28 — Current architecture compliance audit

- Added [review](reviews/2026-10-28_architecture-compliance_review.md) and [EPIC-001 session](ARCHIVE_INDEX.md#pruned-2026-10-28-architecture-compliance); updated architecture/epic navigation. Preserved unrelated local work. No runtime code changes.
- Recorded actual layers/patterns and six OPEN findings: raw persistence errors, missing auth SQL context, dormant legacy order atomicity, incomplete core isolation, ORM-coupled DTOs and bypassed session-cache wiring.
- Fresh verification: Go 85 pass/34 DB skips and vet; SDK strict check/35 tests; web lint/TypeScript/3 date tests; Go formatting; API/AI Atlas history validation; Python 23 pass/4 skips with test loopback access. No fresh browser/race/build/DB integration/remote CI or merge claims. Review artifacts local/in-review; parent epic state unchanged.

## 2026-10-27 — Thesis-first product roadmap

- Added [roadmap](plans/2026-10-27_thesis-first-ecosystem_plan.md) using the owner's 1Dec2026 milestone and live Notion HEX-4/6, sprint and outstanding-task reads. Preserved all 13 FR/5 NFR groups, simulation limits, early capacity review and evaluation/writing buffer.
- Proposed Hexta supervised use followed by one companion app at a time, demand-driven ordering and integration after standalone value. Updated EPIC-001/session/index and canonical architecture; no Notion writes or completion transition.
- Verification: source-backed document inspection, local links/fences and git whitespace checks only. No application, paid-provider, runtime or deployment tests.

## 2026-10-27 — Product-boundary correction

- Recorded owner clarification: Hexta for orders/business operations, PostPilot for fanpage management, ViralForge for content creation. The separate Hexta Platform/Sales product split remains an unaccepted historical proposal.
- Updated proposal, canonical architecture and EPIC-001 context/session/index. Shared identity/billing and exact conversation scope remain open. Documentation only; whitespace checked, no runtime changes/tests.

## 2026-10-27 — Hexta-centered SaaS ecosystem proposal

- Recorded owner direction and [architecture proposal](designs/2026-10-27_saas-app-platform_design.md): Hexta shared platform and sales/operations, ViralForge content generation, PostPilot publishing, canonical tenants, app-owned data and explicit artifact handoff.
- Linked [EPIC-001 session](epics/EPIC-001-system-design/sessions/2026-10-27_saas-app-platform.md), epic index and canonical architecture. Detailed SSO/entitlements/contracts/events/billing remain TARGET/OPEN.
- Verification: bounded current-source/document inspection, local links and whitespace checks only. No runtime/API/schema change, application tests, deployment, PR/CI refresh or merge/completion claim.

## 2026-10-20 — Initial DDD domain discovery

- Added [domain map](designs/2026-10-20_ddd-domain-map_design.md) and [EPIC-001 session](epics/EPIC-001-system-design/sessions/2026-10-20_ddd-domain-map.md); linked architecture and epic navigation. Existing unrelated changes preserved.
- Proposed nine business contexts, candidate aggregate ownership, core/supporting classification and incremental adoption. Distinguished active foundations, legacy inventory/order behavior and target messaging/scripts. AI execution and provider adapters do not own business authorization.
- Verification: current route/service/schema/model/fake-graph inspection, documentation links/fences and whitespace checks. No runtime tests, source/API/schema changes, live PR/CI update or new delivery-completion claim.

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

## 2026-10-19 — Proposed next-work backlog

- Compared selected live Notion requirements/tasks against router, fake graph, legacy order DTO/AI, frontend inventory and epic context. Added a prioritized plan with dependencies, acceptance criteria, existing HEX task mappings and dataset/evaluation work.
- Added an EPIC-001 session and navigation; appended current #52/#53 merge/check evidence to epic/operations notes without rewriting historical sessions or claiming target implementation.
- Verification: read-only GitHub/Notion/code checks, local Markdown targets and git diff --check. No runtime tests rerun; no external tracker writes. New planning delivery remains in-review until merged.

## 2026-10-19 — Demo-first backlog revision

- Revised the active plan at the owner's direction: defer keys, mock-data work and first-demo infrastructure dependencies; deliver vertical workspace/customer/product/order slices, then configurable responsibility handoff and factual overview. Manual conversation records follow feedback; no fake customer-send action is proposed.
- Preserved earlier scope/evidence, added a separate EPIC-001 session and updated epic/index/operations sequencing. No code or external tracker changes.
- Verification: whitespace and new local links; application tests not rerun for this documentation-only revision.

## 2026-10-19 — First manual sales demo (D1–D3, local/unmerged)

- **Outcome:** Implemented workspace selection/team, customer/product entry/edit/archive and order draft/edit/assigned-sale confirmation UI. Records persist in PostgreSQL; no key, dataset preparation, channel simulator or external sending is required.
- **Files:** Added sales service/repository/controller/DTO/types and web sales components/client; updated Fx/router, existing models, SDK error detail handling and generated Swagger. Removed the obsolete HTTP order controller/status/payment mutation routes. Added migration 20261019030000 and sales/CORS/owner-invitation checks. Updated canonical architecture/DB/API/workflow/operations docs and local runbook; progress is recorded in the EPIC-007 session.
- **Decisions:** Transactional membership/reference/actor checks, batched product locks, order versions, missing-price flags and audited confirmation replace public auto-confirm/inventory behavior. Partial phone uniqueness permits customers without phone. Primary product variant is editable; additional variants/history are retained. CORS includes bearer/workspace headers with explicit origins. Confirmed orders are read-only in this slice.
- **Verification:** Go race suite and vet; 25 SDK tests; web lint/type checks; Swagger regeneration and Atlas integrity; fresh PostgreSQL migrations and existing local upgrade; six PostgreSQL sales scenarios including concurrent confirm/edit; API/web Docker production builds and healthy runtime; local HTTP and browser end-to-end smoke through reload of a confirmed order. The inherited make migrate-diff host configuration failed on Docker-only hostname resolution; the narrow migration was scoped manually and applied/validated through Atlas without reset.
- **Handoff:** Open http://localhost:3000/orders, register/sign in and enter records through the forms. Start/rebuild with scripts/local.sh up --build -d --wait api web. Next scope is D4 responsibility handoff after demo feedback. Delivery is not merged; no broad epic/Notion completion is claimed. New-create request idempotency and full workflow/cases/dashboard remain open.

## 2026-10-19 — Explicit workspace selection before orders

- Added a standalone workspace chooser, workspace cards/create-and-open action, and Change workspace navigation. Authentication and entry links now open /workspaces. Removed automatic first-tenant/localStorage fallback; explicit tab-session selection gates management screens and clears on login/logout.
- Verification: web lint/type checks passed. Local Docker rebuild was not executed: automatic approval review failed twice at model capacity, rather than rejecting the action as unsafe. The previous demo stays running; new browser verification is pending rebuild. No backend/schema change. See the dedicated EPIC-007 session; delivery remains incomplete/unmerged.

## 2026-10-19 — Tenant-first framework plan

- Captured owner-confirmed tenant hierarchy and commercial boundary: customer registration request requires Hexta approval; owners invite employees by email and recipients accept before access. Added a source-grounded plan for distinct platform/applicant/member authority, nested business routes, approval and invitation lifecycle, migration risks and acceptance criteria.
- Preserved existing sales code and all uncommitted work. Current direct create/add-member/plan mutation behavior is explicitly a gap; no payment engine, email provider, production capability or implementation completion is claimed. Updated EPIC-001 primary session, cross-epic handoff/index and canonical product/architecture decisions.
- Verification: read-only code inspection, local links and git whitespace. Planning only; no application tests/build or external tracker changes.

## 2026-10-19 — Tenant-first foundation (local/unmerged)

- **Outcome:** Account hub /tenants precedes tenant-qualified orders/catalog/settings. Customers apply; Hexta operator approval provisions owner access atomically. Owners issue email-addressed private invitations; matching recipients explicitly accept before membership.
- **Files:** [NEW] onboarding model/repository/service/controller/DTO/tests, additive migration 20261019040000, trusted operator CLI, web application/review/invitation/settings pages and client. [MODIFY] router/Fx, tenant commercial guard, SDK/generated dist, Swagger, Dockerfile, shell/auth navigation, Python smoke source and local runbook. See the [primary implementation session](epics/EPIC-001-system-design/sessions/2026-10-19_tenant-first-implementation.md).
- **Decisions/walkthrough:** Apply stores pending state; operator review locks request and provisions tenant/owner in one transaction. Invitations store only token digests, default seven-day expiry, serialize tenant before invitation and require explicit accept. Manual links clearly say no email sent. Path tenant context replaces tab selector state; old direct create/add/global sales and legacy AI routes are unregistered. Existing data remains intact; owner plan/status writes rejected.
- **Verification:** Full Go race suite/vet; isolated HTTP onboarding workflow; eleven PostgreSQL sales/onboarding scenarios including concurrency; 25 SDK tests; web lint/type checks; regenerated Swagger and Atlas integrity; six fresh migrations and existing-volume upgrade. Browser checked pending application/no extra access, existing tenant navigation and confirmed-order persistence. Revised Python smoke syntax passes; runtime not rerun. Final Docker rebuild result is recorded in the session.
- **Limit:** Automatic approval review rejected a temporary live operator grant for missing explicit privilege-change authorization; no grant executed. Live review/accept verification remains pending, separate from passing isolated tests. Email/payment/provider integrations and expanded workflows remain deferred. No commit/PR/merge or external tracker completion claimed.

- **Final runtime:** API/web production rebuild passed; web/API/PostgreSQL/Redis healthy, init/migrate exited successfully. Final browser reload verified chooser, nested orders/data preservation and manual-invitation notice; disposable test DB stopped. Operator binary presence checked without granting access.

## 2026-10-19 — Theme-aware Hexta logo

- [NEW] apps/web/public/brand/hexta-light.png and hexta-dark.png: approved generated transparent PNGs, retained unchanged. [NEW] apps/web/components/brand-logo.tsx: accessible shared logo with CSS theme switching and a fixed frame for transparent margins.
- [MODIFY] landing navigation/footer, login/register, workspace shell and invitation entry replace placeholder branding. Existing sales/onboarding edits preserved. No API/schema changes.
- Verification: web lint and tsc --noEmit passed; git diff --check passed. Browser checked landing light/dark switching and dark login on local dev server. Authenticated workspace not runtime-tested.
- Manual check: open home, toggle theme, then visit login/register and tenant/invitation headers. Local/unmerged; see [session](ARCHIVE_INDEX.md#pruned-2026-10-19-theme-aware-branding).

## 2026-10-19 — Icon-only Hexta favicon

- [NEW] apps/web/public/brand/hexta-icon.png: transparent icon-only master generated with ImageGen from approved logo; no wordmark. [NEW] apps/web/app/icon.png: 192px PNG. [MODIFY] apps/web/app/favicon.ico: 16/32/48px frames replace starter favicon. Sharp handles resize/format packaging.
- Next.js discovers icon files and emits versioned metadata links; no runtime theme hooks. Blue emblem works on light/dark tabs.
- Verification: web lint, tsc --noEmit and git diff --check passed. Browser confirmed both icon links and rendered PNG. ICO frame directory validated. Manual check: refresh/open tab and confirm blue Hexta emblem.
- Local/unmerged; [session](ARCHIVE_INDEX.md#pruned-2026-10-19-brand-favicon).

## 2026-10-19 — Demo PR and manual test handoff

- Package the latest icon/favicon assets with their existing memory and add infrastructure/TESTING.demo.md, linked from the runbook and affected epics. The checklist covers setup/resume, tenant boundaries, incomplete drafts/confirmation, validation/concurrency, approval authority, invitations and issue reporting. No new business behavior or live operator grant.
- Verification: current web lint/TypeScript checks, icon file format inspection and whitespace validation. Prior implementation test evidence remains scoped to its recorded snapshot. Manual scenarios are pending human execution; delivery review/merge remains open. See the primary handoff session for PR evidence.

- **Publication:** dc37ab3 commits latest icon/checklist; 0cf11e0 incorporates main with no application content changes. [PR #54](https://github.com/TruongHoang2004/Hexta/pull/54) opened and attached; mergeable, CI pending at publication. Manual acceptance/merge remain open.

- **Evening-test runtime:** rebuilt current API/web successfully; all required services healthy. HTTP icon/favicon content byte-matches the committed assets; existing demo data retained.


## 2026-10-19 — Thesis continuation from Notion

- Summary: populated the existing Vietnamese LaTeX abstract, five chapters, two appendices and bibliography from Notion rationale/PRD/HEX-3/HEX-5 and inspected repository evidence.
- Components: nine modified files under `graduation thesis/`; [session](ARCHIVE_INDEX.md#pruned-2026-10-19-thesis-continuation), epic summary and index updated.
- Decisions/walkthrough: replaced placeholders within existing sections; retained main/preamble/metadata/cover/contents and all heading order. Separate target messaging/workflows from implemented code, historical checks and unmeasured AI outcomes. No architecture/API/schema change.
- Verification: desktop LaTeX compilation succeeded; Python comparison against HEAD confirmed every existing chapter/section/subsection unchanged and all citation keys resolved. Preview opening queued. No exported PDF/page-image visual QA or fresh application test run; final pagination requires author review.
- Delivery: local/unmerged; personal details, academic literature and empirical evaluation remain.

## 2026-10-19 — Thesis author correction and UET emblem

- Summary: removed thesis bibliography/citations because Notion records the author's own thinking. Left related work, experiment data/scenarios, evaluation, abstract and conclusion/future-work empty; removed unsupported historical-test narrative, proposed methodology and editorial equation. Body placeholder instructions removed.
- Components: existing graduation thesis TeX sources; new assets/logo.png; metadata provenance comments; [session](epics/EPIC-001-system-design/sessions/2026-10-19_thesis-evidence-logo.md) and epic/index updates. No architecture/schema/API change.
- Decision/walkthrough: preserve every existing chapter/section and the references include, but references.tex emits no content. Original SchoolLogo macro loads the official UET-site emblem at its existing size. No replacement document, exported PDF or new tab.
- Verification: desktop compilation of the saved main.tex succeeded; heading comparison against HEAD and absence of citations/bibliography/body placeholders passed. Downloaded 512px emblem visually inspected. Main/preamble/cover/contents unchanged. Full rendered-page review and application tests not run.
- Delivery: local/unmerged; missing author inputs and empirical results intentionally remain blank.

## 2026-10-19 — Remove thesis installation/use appendix

- Removed the installation include from existing main.tex and deleted appendices/installation.tex; removed chapter 4's dangling reference and README inventory entry. Owner expects cloud users to work through the UI; deployment completion is not claimed. Other thesis structure/formatting preserved.
- Verification: built-in compilation of the saved main.tex succeeded. Document-only edit; no application tests or separate PDF export.
- [Session](epics/EPIC-001-system-design/sessions/2026-10-19_thesis-remove-installation.md): local/unmerged, in-review; completed false.

## 2026-10-19 — Thesis delivery preparation

- Consolidated the owner-approved thesis changes on codex/thesis-content from origin/main 5a1e719. PR #54 already merged with successful CI; no application changes included.
- Verification: current thesis project compiled through the open chapter; git diff --check passed. [Delivery session](ARCHIVE_INDEX.md#pruned-2026-10-19-thesis-delivery) records scope and merge authorization. PR merge state is authoritative for completion.

## 2026-10-19 — Live Notion/code reconciliation

- Compared all 42 task rows and acceptance criteria with current routes/services/tests and delivery memory. Added per-ticket assessment/next plan plus primary epic session; proposed HEX-15/16 initial-scope closure, HEX-10 setup review and partial progress for broad DB/API/UI/testing tasks. No Notion status/task writes.
- Reverified PR #54 merge and successful checks, and #52/#53/#55 merge evidence. Updated current epic/design navigation without rewriting historical sessions or claiming full MVP/research completion. Priority remains manual acceptance/fixes, versioned work handoff, factual overview, then manual conversation; keys/mock data remain deferred.
- Verification: live Notion schema/query (42 rows, has_more=false), remote PR facts, actual source inspection, local report links and whitespace. No application test reruns, account grants or implementation changes. Report delivery remains local/unmerged.

## 2026-10-19 — Repair current Notion sprint

- Owner authorized fixing the empty Sprint 2. Linked existing HEX-27/23/20/18 for 19–25 October, set partially implemented tasks/sprint in progress, added weekly scope and current sprint goal/deliverables without rewriting umbrella acceptance or historical completed tasks. No duplicate tasks or key/mock-data commitments.
- Verification: six successful remote updates followed by fetching sprint and four tasks; reciprocal relations, dates, notes and current plan confirmed. No application code/tests or privileges changed. See the primary repair session; local memory remains unmerged.

## 2026-10-20 — Delegated manual browser evidence

- User requested a subagent for local UI testing/screenshots. Nine bounded tenant/catalog/order/price cases passed; eight JPEGs retained under EPIC-007 evidence. Test-only customer/product and two confirmed orders retained.
- Verification: existing Docker stack healthy; browser create/edit/reload, empty draft, quantity 2 × price 10, confirmation/read-only reload, missing-price rejection and explicit-zero confirmation. No application failures in executed cases; remaining checklist NOT RUN. No automated test reruns, grants, code changes or downstream-effect audit.
- [Session](epics/EPIC-007-orders-workflows/sessions/2026-10-20_manual-browser-testing.md), epic/index and operations verification updated; local/unmerged, parent epic unchanged.

## 2026-10-20 — D4 configurable workflows and responsibility handoff

- Summary: implemented local versioned sequential workflow/work-item vertical slice on codex/order-workflows. Owners publish ordered steps and explicitly create work from confirmed orders; responsible employees hand over directly or await nominated acceptance; owners reassign/unassign, and final responsible employees complete work. Earlier versions/history persist.
- Components: [NEW] repository/workflowmodel, workflow repository/service/types/DTO/controller and tests; Atlas migration 20261020075908_order_workflows; typed web workflows service, Work/Workflows forms/pages. [MODIFY] Fx/router/CORS wiring, schema loader/generated Swagger, tenant navigation/order link, manual checklist and architecture/epic records. [Design](designs/2026-10-20_order-workflows_design.md), [session](epics/EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md).
- Decisions/walkthrough: immutable ordered versions pin running work; waiting retains old responsibility until acceptance; owner assignment cancels nomination without advancing; row locks plus expected_version append one event transactionally. Durable creation UUID/hash replays exact requests and rejects changed payloads. Composite tenant FKs and pinned-step FK protect links; no stock/payment/message/order effects. Latest 100 events displayed, older events retained. Model-only repository subpackage avoids existing logging/DTO import cycle.
- Verification: API/shared Go race suite and vet passed; seven real PostgreSQL workflow cases passed, including acceptance/reassignment race, same-key creation concurrency and real Atlas constraints. Isolated HTTP validation/actor/transition test passed; CORS regression passed. Web lint/types and Docker API/Next.js production builds passed. Swagger regenerated with all prior JSON contracts semantically preserved. Atlas integrity, seven fresh migrations and existing demo upgrade passed. Disposable test container stopped/automatically removed; demo volume/services retained. Browser evidence pending at this entry.
- Reproduce: run scripts/local.sh up --build -d --wait api web; owner opens Workflows to publish steps, then Work to select confirmed order/member. Follow the D4 section of infrastructure/TESTING.demo.md. PostgreSQL tests use an isolated SALES_TEST_DATABASE_URL; actual-schema tests additionally require WORKFLOW_MIGRATION_DATABASE_URL with Atlas migrations applied.
- Delivery: owner authorized a scoped local commit; unmerged, parent epic remains in-progress. Existing unrelated reconciliation/evidence edits preserved. No issue/Notion lifecycle updates, cancellation/return/D5 or messaging scope.

- D4 browser follow-up: NOT RUN. IAB unavailable/empty inventory; native Chrome getApp returned no usable UI state and was interrupted after ~302 seconds. No D4 screenshots or browser mutations; prior sales screenshots do not verify D4. Automated/API/PostgreSQL/build evidence above is unchanged. Browser verification and review/merge remain open.

## 2026-10-20 — Current code ERD

- Added system/09-current-erd.md and editable Mermaid overview for af06cf4: 21 public business tables plus Python ai_runtime.runs, complete relationship overview and four all-column domain diagrams. Linked canonical design/database navigation and primary EPIC-001 session.
- Verified 22 tables/215 columns by parsing seven Atlas API migrations plus Python 001 and applying sales ALTER overlays; reviewed FK/unique constraints and GORM column mappings. Solid/dashed lines explicitly distinguish physical/application references; no invented users/future tables. No live row inspection, schema/code changes or application test rerun. Documentation is local/unmerged; epic states unchanged.

## 2026-10-20 — Atlas-only migrations for all current services

- Summary: configured root Atlas environments and desired SQL for API/public and AI/ai_runtime. Preserved existing API history; new guarded AI migration adopts compatible existing runs without losing data. Removed Python startup DDL, old SQL/config paths and all Go test AutoMigrate calls. Models remain persistence mappings, not runtime schema mutators.
- Components: root atlas.hcl, migrations/schema and migrations/ai; Docker migration image/Compose/API+AI+test migration gates; Make/script api/ai/all workflow; Go PostgreSQL/Atlas fixture helper, Python startup/adoption tests, CI and AGENTS/runbooks. Existing raw database-URL startup log removed. [Design](designs/2026-10-20_atlas-all-services_design.md), [session](epics/EPIC-003-local-docker/sessions/2026-10-20_atlas-all-services.md).
- Decisions: independent revision schemas; explicit job credentials; preflight every target/required URL before apply; compatible AI table shape validated rather than blind baseline. API never allows dirty. Test DBs are generated/disposable and migrated through actual histories; no SQLite substitute. SQL desired schemas avoid inherited GORM schema drift.
- Verification: API/shared PostgreSQL race suite and vet passed; 27 host and Docker Python tests passed, including legacy retention/incompatible rejection/independent histories/readiness-only startup. Atlas integrity, fresh and populated upgrade, repeat no-op, public Make apply/status and API/AI zero-drift passed. Local five-service stack healthy, init/API/AI migration jobs exit 0; tenant/order/customer/run aggregate counts unchanged. Shell syntax/YAML/whitespace checks passed. Initial permission/dev-schema/Compose-wait/test-path/SSL issues were corrected and affected checks rerun.
- Delivery: local/uncommitted/unmerged on codex/atlas-all-services. No HTTP/RPC/business schema change, provider action, production provisioning or tracker state update; earlier unrelated ERD/reconciliation/evidence preserved.

## 2026-10-20 — Publish workflow and Atlas PR

- Created [PR #57](https://github.com/TruongHoang2004/Hexta/pull/57) against main for D4 and Atlas-only service migrations. Rebased unpublished commits onto current main to exclude already-merged thesis work; application code unchanged. All 23 unrelated local files restored with hash verification and excluded from commits.
- Verified PR OPEN/non-draft and initial backend/frontend/container/lifecycle checks in progress on 74dd65a; no passing-CI or merge claim. D4 browser acceptance remains NOT RUN. Primary sessions now in-review/completed false; metadata publication follows separately.

## 2026-10-20 — Post-merge syntax and engineering-rule audit

- Verified PR #57 merged as 2abd305 with green CI and tracked source equality to origin/main. Added [review](ARCHIVE_INDEX.md#pruned-2026-10-20-code-rules-audit-review) and [EPIC-001 session](epics/EPIC-001-system-design/sessions/2026-10-20_code-rules-audit.md); appended merge corrections and evidence to epic/design summaries, preserving earlier history and unrelated local work.
- OPEN findings: predictable JWT fallbacks, browser-readable refresh tokens, four Swagger envelope declarations, legacy AI service boundaries, 13 SDK explicit any occurrences and 31/135 Go files needing gofmt. Principal findings predate #57; no application code changed.
- Fresh verification: Go vet/default tests, web lint/tsc, SDK 25 tests, AI 23 tests, 13 Python AST parses, Atlas integrity, shell/YAML syntax and whitespace pass. Four AI DB tests and Go DB fixtures skip without database URLs. Initial gRPC bind restriction resolved by permitted rerun. Merged CI is separate evidence; no fresh DB/race/build/browser acceptance claim.
- Next: coordinated auth hardening first, then contract/type/architecture/format cleanup. Review/session local/unmerged; no parent epic completion or external tracker transition claimed.

## 2026-10-20 — Delegated rule-compliance implementation

- **Summary:** Implemented audit R1–R6 through auth/Swagger, SDK/browser, legacy AI and formatting/logging/CI subagents on codex/rules-compliance; independent cross-module review corrected logout failure/refresh races and duplicated Swagger path prefixes. No commit/PR/deployment or schema change.
- **Components:** config/auth/main and HTTP DTO/controller/generated docs; SDK source/tests/builds and two browser logout callers; legacy AI repository/domain/provider/transport boundaries and Fx graph; formatting guard/backend+frontend CI; canonical AGENTS paths/security rules. Shared old repository test helper now provisions disposable Atlas databases.
- **Decisions:** API/SDK upgrade together to HttpOnly refresh-cookie contract; missing explicit separate signing keys fail closed. Logout failure preserves retry state and displays an error; confirmed logout fences/drains refresh. Legacy AI routes remain unregistered, with no new messaging/sending capability. Session-based prior refresh validity is retained; no per-token revocation added.
- **Verification:** full Go race suite80 top-level PASS, 11 packages, zero skips/failures against isolated PostgreSQL; subsequent Swagger contract race test and vet PASS; SDK35 tests, strict/no-any guard/build, web lint/tsc, all145 Go-file formatting, tracked guard, shell/YAML syntax, Atlas histories and whitespace PASS. Independent review resolved both introduced findings. Temporary test database removed; demo data preserved. No browser E2E or remote CI/production rollout.
- **Walkthrough/handoff:** [Primary session](epics/EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md) documents ownership, checks, compatibility and limits; [audit follow-up](ARCHIVE_INDEX.md#pruned-2026-10-20-code-rules-audit-review) preserves findings/history. Artifacts local/unmerged, completed false; parent epics remain incomplete.

## 2026-10-21 — Compliance PR publication

- Prepare API/SDK cookie auth, explicit JWT configuration, Swagger, legacy AI boundaries and CI guards for review; preserve pending design/audit/browser-test text records in a separate documentation commit; screenshots remain local and excluded from publication. See [publication session](epics/EPIC-001-system-design/sessions/2026-10-21_compliance-pr.md) and [implementation details](epics/EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md).
- Fresh verification: Go race/vet PASS (default PostgreSQL skips), SDK strict guard/35 tests/build PASS, web lint/tsc PASS, Atlas API/AI integrity and formatting/diff checks PASS. Historical database-backed checks remain separately attributed; browser E2E and production build not rerun.
- Compatibility: upgrade cookie-auth API and SDK together; configure distinct signing keys of at least 32 bytes and an explicit same-site frontend origin. Review/CI/merge remain pending; no parent epic is marked done.

2026-10-22 publication: [PR #58](https://github.com/TruongHoang2004/Hexta/pull/58) is open/non-draft against main; initial backend/frontend/Docker CI in progress at ab053d5, lifecycle check passed. [Publication session](epics/EPIC-001-system-design/sessions/2026-10-21_compliance-pr.md) is in-review/completed false. Screenshots excluded and kept local; review/merge and parent acceptance remain pending.

## 2026-10-23 — Small demo work units and compliance merge reconciliation

- **Summary:** [NEW] [nine-unit plan](plans/2026-10-23_demo-work-units_plan.md) and [EPIC-007 session](epics/EPIC-007-orders-workflows/sessions/2026-10-23_demo-work-units.md). Six acceptance units precede three conditional owner-overview units; each has scope, dependencies, acceptance, files and a copyable prompt.
- **Decisions:** verify existing features before adding replacements; preserve operator/browser NOT RUN gates; isolate parallel accounts/records/branches and serialize shared runtime/memory integration. No subagents launched or application changes.
- **Delivery:** GitHub verified owner-reported PR #58 merge into main as 656a6c4 at 2026-10-23T11:07:01Z. Six returned checks on head 2504d69 succeeded. [MODIFY] epic READMEs/index, publication follow-up and operations module; parent epics remain incomplete.
- **Verification:** source/checklist inspection and remote PR/check metadata; documentation whitespace/link checks only. No application tests or fresh browser results. Existing screenshot files preserved locally/untracked. Planning artifacts remain unmerged/in-review.

## 2026-10-23 — Delegated acceptance and owner overview

- **Summary:** [NEW] overview API/repository/service/types/DTO/controller, Fx wiring and generated Swagger; owner overview route/component/typed client/date helpers/tests. [MODIFY] tenant nav, selected order/work query handling, frontend test script/CI, router contract fixture and canonical API/workflow/operations modules. [Contract](designs/2026-10-23_owner-overview_design.md), [nine-unit execution/session index](epics/EPIC-007-orders-workflows/sessions/2026-10-23_work-unit-execution.md).
- **Decisions:** owner/current-membership checks and all matching count/row queries share one read-only REPEATABLE READ snapshot; independent creation cohorts, current responsibility versus nomination and unassigned subset semantics are explicit. No schema/cache/payment/AI/channel changes. UI request generations and persistent initial-error guards protect linked detail context.
- **Review:** independent review found and verified corrections for order-editor initialization and failed-link blank-editor context. Integrated regression found the old RegisterRoutes test signature; corrected and rerun. No unresolved actionable static findings.
- **Verification:** focused auth/SDK, sales6, invitation4, access4, workflow8 and overview6 checks passed (overlap is not a distinct-test total). Integrated Go race suite+vet with explicit disposable Atlas PG passed; web lint/types, date3 tests and final Docker default Next.js/API builds/start passed. Root browser verified bounded sales/order/work/overview paths and tenant switching. Remaining unrun cases are explicit; no parent/merged completion claim.
- **Limits:** existing demo data preserved and unrelated occupied ports avoided via temporary13000/18080 override. Dual-profile invitation/acceptance, operator, full browser auth/storage and mobile checks remain open. Automatic approval review rejected archive confirmation; the UI later showed the new test customer archived despite the cancellation attempt. Further archival stopped; test product remains active. Screenshot preview and private runtime files remain local, excluded from publication.

Draft overview delivery: [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59), OPEN/draft against main at implementation3537e37. Initial backend/frontend/Docker CI in progress, lifecycle passed; [execution record](epics/EPIC-007-orders-workflows/sessions/2026-10-23_work-unit-execution.md) retains exact evidence/manual/approval gaps. Review/merge pending; no completion transition.


## 2026-10-28 — Prune obsolete/noisy project memory

- Owner-authorized source-backed cleanup removed 89 Markdown files (28 plans, 27 reviews, 12 designs, 5 per-task changelogs, 4 older audits, 13 superseded/admin sessions). [Archive ledger](ARCHIVE_INDEX.md#2026-10-28--source-backed-memory-pruning) records exact originals, replacements, Git/consolidated-draft recovery and hashes; this entry does not repeat their obsolete implementation instructions.
- Preserved active PR #59 work/acceptance evidence, open issue #56, current source review, detailed mapper/cache proposal, latest product roadmap and canonical contracts. GitHub verified pruned issue closure and PRs #9/#11/#52–55/#57/#58 merged; #59 remains OPEN/draft. No CI-success, runtime-test or parent-completion inference.
- Compacted memory/epic/system navigation and repeated architecture/operations status; legacy rules are short compatibility redirects. Updated historical links and canonical Kafka runbook reference; root rules retain all 91 IDs and selective loading. [Primary cleanup session](epics/EPIC-001-system-design/sessions/2026-10-28_memory-pruning.md) owns verification/limits. No application, dependency, schema or screenshot change.


## 2026-10-28 — Publish rules and memory cleanup for review

- Prepared isolated docs branch from current main, excluding overview application changes/screenshots while retaining referenced acceptance documents and main's newer audit. [Publication session](epics/EPIC-001-system-design/sessions/2026-10-28_rules-memory-pr.md) records validation and eventual PR identity. No runtime test or parent completion claim.
