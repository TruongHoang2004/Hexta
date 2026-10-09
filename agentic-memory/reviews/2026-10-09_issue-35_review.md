# Code Review: Inventory Repository, Stock Reservation, and Movement Tracking

- **Issue**: [#35](https://github.com/TruongHoang2004/Hexta/issues/35)
- **Review Date**: 2026-10-09
- **Reviewer Agent**: `dev-review` (on behalf of `backend_core` & `qa_engineer`)
- **Target Changes**: Implementation of `InventoryRepository`, `InventoryService`, movement audit ledger, and tests.
- **Verdict**: **APPROVED**

---

## 1. Scope & Changes Checked

### 1.1 Persistence Layer (`services/api/internal/repository/inventory_repository.go`)
- **Atomic Two-Phase Reservation**:
  - `ReserveStock`: Employs atomic conditional SQL update `WHERE tenant_id = ? AND variant_id = ? AND available_qty >= ?`. Prevents overselling without exclusive row locking.
  - `ReleaseStock`: Safe decrement of `reserved_qty` and replenishment of `available_qty`.
  - `DeductStock`: Permanent deduction of `on_hand_qty` and `reserved_qty` upon order fulfillment with automatic `StockMovement` (OUTBOUND) record.
  - `AdjustStock`: Reconciles physical inventory discrepancy with before/after balance tracking.
- **Multi-Tenant Isolation**: All queries and mutations strictly constrain on `tenant_id`.

### 1.2 Service Layer (`services/api/internal/core/service/inventory_service.go`)
- `IInventoryService` interface provides clean abstraction over inventory queries, reservations, stock adjustments, and movement listings.
- Error wrapping conforms to `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.

### 1.3 Dependency Injection (`services/api/internal/bootstrap/`)
- `repository.NewInventoryDBRepository` and `service.NewInventoryService` cleanly wired via Uber Fx `fx.Annotate` into their respective interfaces.

### 1.4 Test Verification
- `services/api/internal/core/service/inventory_service_test.go`: Unit tests for stock reservation, over-reservation checks, and physical adjustment.
- `services/api/internal/repository/inventory_repository_test.go`: Integration test suite exercising atomic conditional reservation and movement tracking.

---

## 2. Quality & Architecture Compliance

| Gate | Status | Evidence |
|---|---|---|
| 5-Layer Go Architecture | **PASSED** | Clean boundary between Controller, Service, Repository, and Model. |
| Language Standard | **PASSED** | 100% English identifiers, comments, and documentation. |
| Test Execution | **PASSED** | `go test ./services/api/internal/core/service/... ./services/api/internal/repository/...` passed in 2.8s. |
| Build Check | **PASSED** | `go build ./services/api/...` compiles with zero warnings. |
