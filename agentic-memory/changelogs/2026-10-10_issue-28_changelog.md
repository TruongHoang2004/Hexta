# Changelog: Resolve 5-Layer Architecture Boundary Violations in API Controllers

- **Issue**: [#28](https://github.com/TruongHoang2004/Hexta/issues/28)
- **Date**: 2026-10-10
- **Domain**: `domain:backend-api`
- **Component**: `services/api`
- **Author**: Autonomous Task Runner

---

## Summary of Changes

Enforced strict 5-layer Clean Architecture boundaries across API presentation controllers by eliminating direct database and GORM references from controller structs and handlers. Standardized health check responses using dedicated DTOs and `response.Response[T]`, updated OpenAPI/Swagger documentation, and added controller unit tests.

### Key Changes:
1. **`services/api/internal/present/http/controller/health_controller.go`**:
   - Ensured `HealthController` only depends on `service.IHealthService` and validator, removing direct `*gorm.DB` references.
   - Refactored `HealthCheck` handler to return standard `response.Response[dto.HealthCheckResponse]`.
   - Updated Swagger annotations with `@Success 200` and `@Failure 503` wrapping `dto.HealthCheckResponse`.
2. **`services/api/internal/present/http/dto/health_dto.go`**:
   - Added `HealthCheckResponse` DTO for structured health check status and component details.
3. **`services/api/internal/present/http/controller/health_controller_test.go`**:
   - Added unit test suite verifying `HealthCheck` responses for healthy (200 OK) and degraded (503 Service Unavailable) states.
4. **`services/api/docs/`**:
   - Re-generated Swagger documentation (`docs.go`, `swagger.json`, `swagger.yaml`) reflecting the standardized response DTO.

---

## Verification
- `go vet ./services/api/... ./packages/shared/...`: Passed with zero warnings.
- `go test -v -race ./services/api/internal/present/http/controller/...`: Passed all tests with race detection.
- `go test -race ./services/api/... ./packages/shared/...`: All unit tests passed cleanly.
- `make swagger`: Generated valid Swagger OpenAPI specs.
- `python3 .agents/scripts/audit_codebase.py`: Architecture scanner reported zero boundary violations.
