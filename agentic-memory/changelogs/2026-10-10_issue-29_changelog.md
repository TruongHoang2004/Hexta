# Changelog: Unit Test Suites for Critical Core Services and SDK

- **Issue**: [#29](https://github.com/TruongHoang2004/Hexta/issues/29)
- **Date**: 2026-10-10
- **Branch**: `task/issue-29-test-coverage-implement-unit-test-suites`
- **Domain**: `testing-and-qa`

---

## 1. Overview
Implemented unit and integration test suites for uncovered repository operations and TypeScript SDK clients, eliminating codebase coverage gaps.

---

## 2. Modified & Added Files

### `services/api/internal/repository/order_repository_test.go` (Added)
- Added `TestOrderRepository_CRUD` verifying:
  - Atomic order creation with items in transaction
  - Eager-loaded item retrieval by ID and by Order Number
  - Not found error verification on invalid IDs/order numbers
  - Paginated list queries with status filtering
  - State updates for `OrderStatus` and `PaymentStatus`

### `packages/sdk/src/index.test.ts` (Modified)
- Added 11 new tests covering:
  - Tenant management endpoints (`createTenant`, `getTenants`, `getTenant`, `updateTenant`)
  - Team member management endpoints (`getUsers`, `inviteMember`)
  - Storage adapter utilities (`getStorage()`, `clearTokens()`, `MemoryStorage.clear()`)
  - Error translation and unwrap behavior (`handleApiError`)

### Memory Artifacts
- Added `agentic-memory/plans/2026-10-10_issue-29_plan.md`
- Added `agentic-memory/reviews/2026-10-10_issue-29_review.md`
- Added `agentic-memory/changelogs/2026-10-10_issue-29_changelog.md`

---

## 3. Verification Steps
1. Run backend race tests:
   ```bash
   go test -race ./services/api/... ./packages/shared/...
   ```
2. Run SDK vitest suite:
   ```bash
   pnpm --filter @hexta/sdk test --run
   ```
3. Run automated codebase audit:
   ```bash
   python3 .agents/scripts/audit_codebase.py --dry-run
   ```
