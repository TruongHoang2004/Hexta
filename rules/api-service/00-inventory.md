# Observed API service architecture

This is a source inventory, not a declaration that every existing implementation is correct. Paths below are relative to `services/api` unless stated otherwise. Evidence reflects the inspected local checkout, not remote main/CI state.

## Layers and supporting packages

| Area | Existing location | Observed responsibility |
| --- | --- | --- |
| Entry points | `cmd/main.go`, `cmd/platform-operator`, `cmd/ai-smoke`, `cmd/tools` | Start HTTP application or explicit administrative/diagnostic commands |
| Composition root | `internal/bootstrap` | Fx dependency graph, constructors, routes, lifecycle and observability registration |
| Presentation | `internal/present/http/router`, `middleware`, `controller` | Route selection, transport authentication, binding, request validation and response mapping |
| Presentation contracts | `internal/present/http/dto`, `response`, `validator` | Request/response shapes, validation functions and success envelope |
| Business/application | `internal/core/service` | Auth, tenant/onboarding, sales, workflow, overview and legacy inventory/order/AI orchestration |
| Core data | `internal/core/model`, `internal/core/types` | Shared business/input/output types; existing models also contain ORM tags/GORM types |
| Persistence | `internal/repository`, `repository/workflowmodel` | SQL/GORM, transactions, row/version conditions, audit writes and cache decorators |
| Infrastructure | `internal/infrastructure/database`, `cache`, `gemini`, `ai/rpc` | PostgreSQL/Redis clients, external HTTP adapter and generated protobuf bindings |
| Supporting code | `config`, `common`, `common/log`, `utils`, `internal/constant` | Configuration, actor/trace context, logging and helpers |
| Verification/contracts | `internal/testutil`, `docs/contract_test.go`, repository-root `migrations`, `proto` | Explicit Atlas fixtures, Swagger contract tests and external contracts |

The established five-layer model is entry → presentation → service → repository, with infrastructure supplying adapters/resources through bootstrap. Bootstrap and DTOs support the layers; they are not additional business tiers. A request need not traverse every layer. Services can compose other services without issuing SQL.

## Technologies actually referenced

| Technology | Evidence | Current role |
| --- | --- | --- |
| Go/workspaces | `go.mod`, root `go.work` | API/shared modules; module declares Go 1.25.3, workspace declares 1.26.2 at inspection |
| Gin + gin-contrib/cors | `bootstrap/router.go`, `present/http` | HTTP routing, middleware, CORS |
| go-playground/validator v10 | `bootstrap/controller.go`, `present/http/validator` | DTO/registered custom validation |
| Uber Fx | `cmd/main.go`, `bootstrap/*` | Constructor injection and startup/shutdown hooks |
| GORM + PostgreSQL/pgx | `infrastructure/database/connect.go`, repositories | Relational persistence and transactions |
| Atlas | root `atlas.hcl`, `migrations/schema/api.sql`, `migrations/api` | Desired SQL and versioned migrations; GORM provider dependency is not the canonical desired schema |
| Redis/go-redis v9 | `infrastructure/cache` | Cache/session/state primitives and legacy draft cache |
| Zap via common/log | `common/log/global.go` | Context/trace-aware application logging |
| Viper + godotenv | `config/config.go` | Environment/config loading and auth startup validation |
| golang-jwt v5, bcrypt, OAuth2 | `core/service/auth_service.go` | Access/refresh credentials, password hashing and Google login |
| Swaggo | `cmd/main.go`, controllers, root `make swagger` | Generated Swagger artifacts/UI |
| shopspring/decimal | sales/order/model types | Decimal money calculations |
| OpenTelemetry + Prometheus | `bootstrap/router.go`, `telemetry.go`, shared telemetry | Trace middleware and metrics handler wiring; no claim of a verified exporter deployment |
| HTTP Gemini adapter | `infrastructure/gemini/client.go`, `IAIGenerator` binding | Legacy Go generation port; no live provider verification or approved messaging scope implied |
| gRPC/protobuf | generated `infrastructure/ai/rpc`, `cmd/ai-smoke`, root `proto` | Python AI contract and smoke command; business dispatcher is not implemented |
| testing/testify, race/vet | `*_test.go`, backend workflow | Unit/integration/Fx contract verification |

Dependency presence alone does not prove a feature is active. AIController constructors exist but AI routes are not registered in the current business router. `/orders` binds SalesController; legacy OrderService stock effects are not that route's current behavior.

## Patterns with concrete evidence

| Pattern | Evidence | Boundary/qualification |
| --- | --- | --- |
| Layered architecture / service layer | `present/http/controller` → `core/service` → `repository` | Main organizational pattern; boundaries have known deviations |
| Constructor DI / composition root | `bootstrap/service.go`, `repository.go` | Fx, including interface bindings |
| Repository | `ISalesRepository`, `ITenantRepository`, `ISessionRepository` | Interfaces currently live in repository; several services depend on concrete repositories |
| Decorator + cache-aside | `session_wrapper.go`, `tenant_wrapper.go` | Tenant consumer uses wrapper; auth requests concrete DB repository |
| Ports/adapters, partial | `IAIGenerator`, `IDraftCache`, Gemini binding | Existing example, not universal hexagonal architecture |
| Unit-of-work-style transaction callback | sales `Transaction`, workflow `Within`, overview `WithinSnapshot` | Persistence mechanics stay in repository; use case chooses atomic scope |
| Explicit finite state machine | `WorkflowService.Command`, legacy `TransitionStatus` | Switch/transition rules, not class-based State pattern |
| Optimistic concurrency and row locks | expected versions, conditional SQL, GORM locking clauses | Current sales/workflow consistency mechanisms |
| Idempotent command/replay | workflow request keys and payload hashes | Same key/payload replays; changed payload conflicts |
| Middleware pipeline / response mapper | Gin middleware and baseController | Cross-cutting request handling and shared error/response conversion |
| Read model / snapshot query | `overview_repository.go` | Read-only repeatable-read overview; not full CQRS |

DDD context isolation, event sourcing, transactional outbox/Kafka delivery and a full Go AI business dispatcher are not implemented patterns in this baseline.

## Historical foundation

Local Git history `c312cfa` already contains the bootstrap/core/service/core/model/repository/infrastructure/presentation split, shared application errors, constructor injection and session-cache wrapper. Later source adds standardized responses, OAuth, tenant-first authorization and sales/workflows. This supports continuity of the layered design; it does not prove individual authorship or every original intention. Current approved Atlas-only and cookie/security rules supersede insecure or obsolete historical implementations. No historical secret values are reproduced here.
