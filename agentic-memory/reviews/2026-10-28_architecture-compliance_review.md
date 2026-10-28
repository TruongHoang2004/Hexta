# Current architecture and engineering compliance review

Date: 2026-10-28. Baseline: `codex/demo-work-units`, `b52a8ba`. Primary epic: EPIC-001. Scope: current Hexta Go API, web/SDK, Python AI boundary, migration and CI source. Existing unrelated documentation/evidence changes preserved. This is a bounded source audit, not exhaustive security certification or production acceptance. No live GitHub/CI/merge state is asserted.

## Assessment

The repository follows a recognizable layered architecture, especially in current sales/workflow/overview paths. It is not a fully dependency-inverted Clean Architecture or an enforced DDD modular monolith. Most business code shares one Go API process/database; Python AI is a separately deployable technical service. Domain folders, aggregates and bounded-context isolation are not consistently enforced. Kafka/outbox, real Python model providers and ecosystem SSO remain targets, not current patterns.

## Current layers

| Area | Layer | Evidence/responsibility |
| --- | --- | --- |
| Go | Entry/composition | `services/api/cmd`, `internal/bootstrap`: configuration, Fx constructors, lifecycle and wiring |
| Go | Presentation | `internal/present/http`: Gin routes/middleware/controllers, request validation, DTOs and response mapping |
| Go | Application/business | `internal/core/service`: authorization, sales rules, workflow transitions and orchestration |
| Go | Persistence | `internal/repository`: GORM queries, transactions, locks, cache decorators |
| Go | Infrastructure | `internal/infrastructure`: PostgreSQL, Redis, Gemini and generated RPC adapters |
| Go | Shared data types | `internal/core/model`, `internal/core/types`: currently mix domain data, ORM types and input/contract metadata; not another execution tier |
| Web | Routing/view | `apps/web/app`, `components`: App Router, forms and rendering |
| Web | Context/state | `components/sales/workspace.tsx`, `store/useAuthStore.ts`: workspace/auth state |
| Web | API access | `lib/services`, `lib/sdk.ts`, `packages/sdk`: typed API facade, auth lifecycle and HTTP interceptors |
| Python | Transport/composition | `hexta_ai/app.py`, `rpc.py`: FastAPI/gRPC/auth and lifecycle |
| Python | Execution | `graph.py`, `durable.py`: deterministic fake graph and durable worker orchestration |
| Python | Persistence | `store.py`: PostgreSQL acceptance, claim, lease, cancellation and fenced result publication |

These five Go layers describe responsibilities, not a requirement that every request calls all five in sequence. Infrastructure supplies resources through composition.

## Patterns actually present

| Pattern | Evidence | Assessment |
| --- | --- | --- |
| Dependency injection/composition root | `bootstrap/service.go`, `repository.go`, Fx lifecycle | Established, but constructor/interface usage varies |
| Repository | `ISalesRepository`, identity/session/tenant/order repositories | Established; several newer services use concrete repositories |
| Decorator + cache-aside | `session_wrapper.go`, `tenant_wrapper.go` | Tenant decorator is injected; auth constructor bypasses session decorator (R6) |
| Ports/adapters, selectively | `service/ai_agent_service.go`: `IAIGenerator`, `IDraftCache`; Gemini Fx binding | Good bounded example; not consistently applied across core |
| Transaction callback / unit-of-work style | Sales `Transaction`, workflow `Within`, overview `WithinSnapshot` | Good current paths; inconsistent legacy transaction ownership |
| Finite state machine | `WorkflowService.Command`, legacy `OrderService.TransitionStatus` | Explicit transition logic; this is not the object-oriented State pattern |
| Optimistic concurrency + row locking | Expected versions and conditional updates in sales/workflow repositories | Strong protection against stale business writes |
| Idempotent command handling | Workflow creation keys/hashes; AI tenant/run digest | Durable replay/conflict handling present |
| Durable leased worker + fencing | Python store `SKIP LOCKED`, lease token/expiry predicates | Present; whole-run replay, not node checkpoint recovery |
| Middleware/interceptor pipeline | Gin auth/recovery/logging; SDK Axios refresh handling | Established cross-cutting behavior |
| Query/read-model separation | Owner overview aggregation in repeatable-read snapshot | Useful separation; not a full CQRS/event-sourced system |

## Findings

### R1 — P2: Raw persistence errors cross the public API boundary

`services/api/internal/repository/base_repository.go:18-19` copies `err.Error()` into a public error detail. Active identity/session/tenant repositories call this helper. `controller/base_controller.go:33-35` uses `ConvertErrorToResponse`, which preserves Detail (`packages/shared/pkg/errors/errors.go`). A driver/constraint failure can therefore disclose database/schema details or offending values in an HTTP response. This is source-confirmed propagation; no production fault was injected.

Remediation: log the original failure server-side with trace correlation and return a stable safe message/code. Reuse the newer `salesDBError` approach, with explicit conflict/not-found mapping. Example replacement return: `errors.ErrSystemError(ctx, "persistence operation failed")`; retain diagnostics only in appropriately redacted server logs.

### R2 — P2: Authentication repositories accept context but do not attach it to SQL

`repository/identify_repository.go:26-52` and `repository/session_repository.go:28-61` use `r.db.Where/Create/Save` without `WithContext(ctx)`. Request cancellation/deadlines do not reach these DB operations. `core/service/health_service.go:49` similarly calls `Ping()` instead of `PingContext(ctx)`. Under a slow/unavailable DB, auth requests and readiness probes can outlive their intended request budget.

Remediation: consistently start operations with `r.db.WithContext(ctx)` and use a bounded `PingContext`. Verify cancellation behavior with a focused DB integration scenario.

### R3 — P2, dormant legacy path: Order/inventory effects do not share an atomic boundary

`core/service/order_service.go:117-143` reserves stock before creating an order. If the final order insert fails, no compensation runs. `TransitionStatus` also mutates inventory before a separate status update; release errors are ignored and multi-line deductions can partially succeed. `repository/order_repository.go:121-127` updates status without an expected-status/version predicate, allowing concurrent callers to repeat inventory effects.

The current HTTP router binds order endpoints to `SalesController`, not this legacy service. This finding is a latent hazard if legacy behavior is reused; it is not evidence that current manual sales confirmation deducts stock incorrectly.

Remediation: remove/quarantine obsolete entry points or implement one transaction covering the order state and stock mutations, plus conditional state/version checks. Do not reconnect legacy code to routes without failure/concurrency tests.

### R4 — P3: Core isolation and dependency inversion are incomplete

`core/service/auth_service.go:22,52` imports a concrete Redis adapter; `health_service.go:6-7,21-27` imports GORM/cache directly. Workflow (`workflow_service.go:18-26`), onboarding and overview accept concrete repository types. Repository interfaces also live in the persistence package, so core imports persistence even when using an interface. GORM models remain in `core/model` (`products.go:7,18-26`), contrary to the canonical instruction to define persistence models in the repository layer.

Impact: harder isolated testing/replacement and framework changes propagate inward. Health is operational code rather than a sales use case, but its placement still violates the stated strict core rule.

Remediation: incrementally define narrow consumer-owned ports, move operational probes to an adapter/composition boundary, separate domain types from ORM models where change pressure justifies it. No full rewrite is necessary.

### R5 — P3: Response DTOs remain coupled to persistence models

`present/http/dto/workflow_dto.go:5-20` derives responses and page item types directly from `workflowmodel` ORM structs. `sales_dto.go:45-47` aliases sales responses to `core/model` ORM structs. Adding a JSON-visible persistence field can change the API automatically; API contract and storage shape cannot evolve independently.

Remediation: define explicit response structs and map at the presentation boundary. Keep domain input types separate from transport validation as modules evolve. Current hidden hash/key fields use `json:"-"`; no current secret disclosure from these DTOs is asserted.

### R6 — P3: Fx registers a session cache decorator which auth does not consume

`bootstrap/repository.go` provides `NewSessionCacheWrapper` as `ISessionRepository`, but `core/service/auth_service.go:75-81` requests `*SessionRepository`, selecting the DB implementation. Tenant service correctly requests `ITenantRepository`. Auth therefore bypasses the advertised session cache-aside path, including its invalidation hook.

Remediation: explicitly choose the intended policy. If DB-backed revocation freshness is preferred, remove unused wiring or document that choice. If caching is enabled, inject the interface only after verifying revocation/invalidation behavior; do not blindly introduce cached authorization.

## What is working well

- Sampled controllers contain binding/validation/response responsibilities without direct GORM queries; business services use shared application errors.
- Current sales commands authorize tenant membership/status inside the transaction; workflow commands check actor roles, expected version and immutable published steps.
- Overview reads aggregate counts and rows within a read-only repeatable-read transaction, avoiding an internally inconsistent dashboard snapshot.
- Browser refresh credentials are handled through HttpOnly cookies; explicit JWT configuration validation, origin checks and SDK refresh coordination are present. These observations are not a complete auth security audit.
- API calls are centralized in web services/SDK; strict SDK types and tests are enforced. UI service contracts are manually maintained and can still drift.
- Atlas owns API/AI histories; CI validates history, desired-state drift, Go formatting/race/vet and frontend checks. Local DB tests explicitly skip when the test URL is absent, so a green default test run alone does not establish database correctness.

## Fresh verification

- `go test -json ./services/api/... ./packages/shared/...`: exit 0; 85 test outcomes passed, 34 skipped because explicit PostgreSQL test configuration was absent.
- `go vet ./services/api/... ./packages/shared/...`: passed (absolute Go executable used on retry after a non-login PATH mismatch).
- `pnpm --filter @hexta/sdk lint`: passed.
- `pnpm --filter @hexta/sdk test`: 35 passed.
- `pnpm --filter web lint`: passed.
- `sh scripts/check-gofmt.sh` and `make migrate-validate`: passed for API and AI histories; no fresh apply/drift/database acceptance claimed.
- Python `uv run --offline pytest -q`, temporary writable UV cache: 23 passed, 4 skipped after allowing test loopback binding. Initial restricted run had 10 gRPC failures caused by sandbox socket denial, not assertion failures; DB-dependent scenarios remain skipped.
- `pnpm --filter web exec tsc --noEmit` and `pnpm --filter web test:overview-dates`: passed; 3 date tests passed. A first non-login pnpm invocation aborted an attempted dependency reconciliation; retry in the established login environment passed without dependency changes. No fresh browser E2E, production build, race suite, load test, paid-provider/connector verification or remote CI/merge verification.

## Recommended order

1. Sanitize shared DB error responses and propagate request context in active auth repositories.
2. Explicitly quarantine legacy order/inventory behavior before any reuse.
3. Establish narrow core ports and explicit response DTOs incrementally in touched modules.
4. Decide session-cache consistency policy and add import-boundary checks once that policy is agreed.

This is a review only. Findings are OPEN, no runtime fixes were made, and parent epic completion remains unchanged.
