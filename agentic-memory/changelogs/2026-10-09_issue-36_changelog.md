# Changelog: Order State Machine, Pricing, and REST Endpoints

- **Issue**: [#36](https://github.com/TruongHoang2004/Hexta/issues/36)
- **Date**: 2026-10-09
- **Domain**: `domain:backend-api`
- **Component**: `services/api`
- **Author**: `backend_core` agent

---

## Summary of Changes

Implemented the Order Management System (OMS) service, repository, and HTTP controllers in Go supporting the full order lifecycle state machine (`Draft` -> `Confirmed` -> `Processing` -> `Fulfilled` / `Cancelled`), coordinating stock reservation with IMS and recording audit lineage.

### Key Changes:
1. **`services/api/internal/repository/order_repository.go`**:
   - `IOrderRepository` interface and GORM implementation.
   - Transactional order creation (`CreateOrder`) with line items.
   - Eager-loading of order line items on queries (`GetByID`, `GetByOrderNumber`, `ListOrders`).
   - Order status and payment status updating methods.
2. **`services/api/internal/core/service/order_service.go`**:
   - `IOrderService` interface and domain business logic.
   - Enforced Finite State Machine transition guards.
   - Coordinated stock reservations and releases with `IInventoryService`.
   - Guaranteed automatic rollback when reservation fails on multi-item orders.
   - Exact monetary calculations using `decimal.Decimal`.
3. **`services/api/internal/present/http/`**:
   - `dto/order_dto.go`: Request and response DTOs with validation rules.
   - `controller/order_controller.go`: Endpoints with `response.Response[T]` formatting and Swagger annotations.
   - `router/router.go`: Protected `/api/v1/orders` route group.
4. **`services/api/internal/bootstrap/`**:
   - Registered `OrderRepository`, `OrderService`, and `OrderController` in Uber Fx container.
5. **Swagger Documentation & Root Makefile**:
   - Added `make swagger` target to root Makefile.
   - Re-generated OpenAPI/Swagger specifications with order models and endpoints.
6. **Unit Tests**:
   - `services/api/internal/core/service/order_service_test.go`: Comprehensive unit tests verifying FSM state transitions, stock coordination, rollback, and pricing.

---

## Verification
- `go build ./services/api/...`: Exited code 0.
- `go test -count=1 -race ./services/api/internal/core/service/...`: Passed cleanly with race detection.
- `go vet ./services/api/...`: Clean, zero issues.
- `make swagger`: Generated valid Swagger documentation.
