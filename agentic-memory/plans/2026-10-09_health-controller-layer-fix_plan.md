# Plan: Resolve 5-Layer Architecture Boundary Violations in Health Controller

- **Date**: 2026-10-09
- **Author Agent**: `backend_api` (supervised by `architect_lead`)
- **Domain**: `domain:backend-api`
- **Target Component**: `services/api`
- **Status**: IN_PROGRESS

---

## 1. Overview & Goal

The automated code scout identified a high-severity architecture boundary violation in:
`services/api/internal/present/http/controller/health_controller.go`

`HealthController` currently imports `gorm.io/gorm` and `github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache`, directly holding pointers to `*gorm.DB` and `*cache.RedisClient`.
According to `GEMINI.md` Rule 2 (5-layer Go architecture):
- **Layer 2 (Controller)**: Handles HTTP requests, parses/validates parameters, and calls the Service layer. Controllers MUST NOT hold database handles or execute infrastructure checks directly.
- **Layer 3 (Service)**: Encapsulates business logic and orchestrates domain operations.
- **Layer 5 (Infrastructure)**: Provides low-level database, Redis, and external clients.

The goal is to decouple database and cache dependencies from `HealthController` by introducing a dedicated `HealthService` in Layer 3 (`services/api/internal/core/service/health_service.go`), wiring it through Uber Fx DI, and ensuring zero regressions.

---

## 2. Current State vs Target State

### Current State:
```
[HealthController] ──────> gorm.DB (Ping)
                   ──────> cache.RedisClient (Ping)
(Violates Layer 2 ──> Layer 5 direct coupling)
```

### Target State:
```
[HealthController (Layer 2)]
           │
           ▼
[HealthService (Layer 3)]
      │             │
      ▼             ▼
 [gorm.DB]     [cache.RedisClient]
 (Layer 5)         (Layer 5)
```

---

## 3. Step-by-Step Implementation Tasks

### Task 1: Create `HealthService` in `internal/core/service/health_service.go`
- Define `HealthStatus` and `HealthCheckResult` structs or detail maps.
- Define `HealthService` interface:
  ```go
  type HealthService interface {
      CheckHealth(ctx context.Context) (status string, details map[string]string)
  }
  ```
- Implement `healthService` struct holding `*gorm.DB` and `*cache.RedisClient`.
- Implement `CheckHealth(ctx context.Context)` encapsulating the DB ping and Redis ping.
- Provide constructor `NewHealthService(db *gorm.DB, redis *cache.RedisClient) HealthService`.

### Task 2: Refactor `HealthController` in `internal/present/http/controller/health_controller.go`
- Remove `gorm.io/gorm` and `internal/infrastructure/cache` imports.
- Update `HealthController` struct to hold `healthService service.HealthService`.
- Update `NewHealthController` constructor to accept `healthService service.HealthService`.
- Delegate `HealthCheck` handler to `h.healthService.CheckHealth(c.Request.Context())`.

### Task 3: Update Uber Fx DI in `internal/bootstrap/`
- Register `service.NewHealthService` in `services/api/internal/bootstrap/service.go`.
- Ensure `controller.NewHealthController` receives `service.HealthService` smoothly from Fx container.

### Task 4: Unit Testing & Verification
- Add unit test `services/api/internal/core/service/health_service_test.go` or verify existing service tests.
- Run `go build ./services/api/...` to ensure compilation.
- Run `go test ./services/api/...` to verify test suite passes.
- Run `python3 .agents/scripts/audit_codebase.py` to confirm architecture violation count drops from 3 to 0!

---

## 4. Risk Assessment & Edge Cases
- **Fx DI Cycle / Missing Dependency**: `*gorm.DB` and `*cache.RedisClient` are already provided by `bootstrap/database.go` and `bootstrap/cache.go`, so `NewHealthService` will resolve seamlessly.
- **API Response Contract**: The JSON payload structure (`{"status": "...", "details": {...}}`) must remain 100% backward-compatible for consumers of `/health`.

---

## 5. Definition of Done (DoD)
- [ ] No direct `gorm` or `cache` imports in `health_controller.go`.
- [ ] `health_service.go` cleanly implemented in Layer 3.
- [ ] Uber Fx wiring clean and compilable.
- [ ] `go build` and `go test` pass with 0 errors.
- [ ] `audit_codebase.py` architecture violations: 0.
