# Code Review: Order State Machine, Pricing, and REST Endpoints

- **Issue**: [#36](https://github.com/TruongHoang2004/Hexta/issues/36)
- **Review Date**: 2026-10-09
- **Reviewer Agent**: `dev-review` (on behalf of `backend_core` & `qa_engineer`)
- **Target Changes**: Implementation of `OrderRepository`, `OrderService`, `OrderController`, DTOs, Swagger documentation, and test suite.
- **Verdict**: **APPROVED**

---

## 1. Scope & Changes Checked

### 1.1 Persistence Layer (`services/api/internal/repository/order_repository.go`)
- **Transactional Atomic Insertion**:
  - `CreateOrder`: Atomically creates `Order` header and `OrderItem` line items in a single database transaction (`tx.Transaction`).
- **Eager Loading**:
  - `GetByID`, `GetByOrderNumber`, and `ListOrders` preload `Items` association for complete order representations.
- **Multi-Tenant Isolation**:
  - Every SQL operation strictly predicates on `tenant_id = ?`.

### 1.2 Service Layer & FSM Transitions (`services/api/internal/core/service/order_service.go`)
- **Finite State Machine Invariants**:
  - `Draft -> Confirmed`: Automatically reserves stock across all line items via `IInventoryService.ReserveStock`. If any line item has insufficient stock, already reserved items are immediately and atomically rolled back.
  - `Confirmed -> Processing`: Valid workflow progression.
  - `Confirmed -> Cancelled`: Atomically releases reserved stock via `IInventoryService.ReleaseStock`.
  - `Processing -> Fulfilled`: Permanently deducts stock via `IInventoryService.DeductStock`, recording outbound stock movement with order reference.
  - `Processing -> Cancelled`: Releases held inventory stock.
  - Invalid state jumps (e.g. `Draft -> Fulfilled`, `Fulfilled -> Draft`, `Cancelled -> Confirmed`) are safely rejected with HTTP 400 Bad Request.
- **Monetary Precision**:
  - Uses `decimal.Decimal` across line totals, subtotal, discount, and total amount.

### 1.3 Presentation Layer & Routing (`services/api/internal/present/http/`)
- **DTOs (`dto/order_dto.go`)**:
  - Standardized request and response structures with validation tags (`validate:"required,min=1"`).
- **Controller (`controller/order_controller.go`)**:
  - Implements `CreateOrder`, `ListOrders`, `GetOrder`, `TransitionStatus`, and `UpdatePaymentStatus`.
  - Standardized `response.Response[T]` wrappers and complete Swagger annotations on all endpoints.
- **Router (`router/router.go`)**:
  - Registered under protected `orderGroup` (`/api/v1/orders`) secured by `AuthMiddleware`.

### 1.4 Dependency Injection & Build Integration (`services/api/internal/bootstrap/`, `Makefile`)
- Uber Fx bindings for `IOrderRepository`, `IOrderService`, and `OrderController`.
- Added `make swagger` target to root Makefile to regenerate OpenAPI specifications.

### 1.5 Test Verification (`services/api/internal/core/service/order_service_test.go`)
- Complete unit test coverage for:
  - Draft order creation with monetary calculations.
  - Auto-confirm creation with stock reservations.
  - Stock rollback on reservation failure.
  - FSM lifecycle progression and cancellation release.
  - Invalid state jump rejections and parameter validation.

---

## 2. Quality & Architecture Compliance

| Gate | Status | Evidence |
|---|---|---|
| 5-Layer Go Architecture | **PASSED** | Strict adherence across Controller, Service, Repository, Model, and Infrastructure layers. |
| Language Standard | **PASSED** | 100% English identifiers, comments, commit messages, and documentation. |
| Test Execution | **PASSED** | `go test -race ./services/api/internal/core/service/...` passed cleanly. |
| Go Vet Check | **PASSED** | `go vet ./services/api/...` completed with zero warnings. |
| Swagger Specs | **PASSED** | `make swagger` generated updated `docs/docs.go`, `swagger.json`, and `swagger.yaml`. |
