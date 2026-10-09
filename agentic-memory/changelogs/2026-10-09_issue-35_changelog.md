# Changelog: Inventory Repository, Stock Reservation, and Movement Tracking

- **Issue**: [#35](https://github.com/TruongHoang2004/Hexta/issues/35)
- **Date**: 2026-10-09
- **Domain**: `domain:backend-core`
- **Component**: `services/api`
- **Author**: `backend_core` agent

---

## Summary of Changes

Implemented the Inventory Management System core repository, business service, two-phase reservation engine, and movement audit ledger.

### Key Changes:
1. **`services/api/internal/repository/inventory_repository.go`**:
   - `IInventoryRepository` interface and GORM implementation.
   - Atomic conditional queries for stock reservation (`ReserveStock`), stock release (`ReleaseStock`), fulfillment deduction (`DeductStock`), and physical adjustment (`AdjustStock`).
   - Movement ledger tracking with before/after balance calculation.
2. **`services/api/internal/core/service/inventory_service.go`**:
   - `IInventoryService` interface and domain orchestration logic.
   - Multi-tenant validation on all operations.
3. **`services/api/internal/bootstrap/`**:
   - Registered `InventoryRepository` in `repository.go`.
   - Registered `InventoryService` in `service.go`.
4. **Unit & Integration Tests**:
   - `inventory_service_test.go`: Verified concurrency boundaries and out-of-stock validation.
   - `inventory_repository_test.go`: Verified atomic reservation and movement listing.

---

## Verification
- `go build ./services/api/...`: Exited code 0.
- `go test ./services/api/internal/core/service/... ./services/api/internal/repository/...`: Passed cleanly.
