# Plan: Implement Test Coverage for Session and Tenant Repositories and Cache Wrappers

- **Date**: 2026-10-09
- **Author Agent**: `qa_engineer` (Domain: `domain:qa`)
- **Target Package**: `services/api/internal/repository`
- **Status**: IN_PROGRESS

---

## 1. Overview & Objective

The codebase audit scout identified 4 missing test coverage gaps in the data persistence layer:
1. `session_repository.go`
2. `tenant_repository.go`
3. `session_wrapper.go`
4. `tenant_wrapper.go`

The objective is to implement unit and integration test suites:
- `session_repository_test.go`: Database operations for sessions (Create, GetByID, GetByToken, Revoke).
- `tenant_repository_test.go`: Database operations for tenants (CreateTenant with transaction, GetByID, GetBySlug, Memberships).
- `session_wrapper_test.go`: Cache-aside behavior (hit/miss, cache invalidation on revoke).
- `tenant_wrapper_test.go`: Cache-aside behavior for tenants (hit/miss, update invalidation).

---

## 2. Test Architecture & Design

### Database Tests (`*_repository_test.go`):
Follow the established pattern in `identify_repository_test.go`:
- Use `getTestDB(t)` with isolated transaction rollback (`tx := db.Begin(); defer tx.Rollback()`).
- Gracefully skip via `t.Skipf(...)` if PostgreSQL docker container is not reachable, ensuring tests never block environments lacking external services.

### Wrapper Tests (`*_wrapper_test.go`):
- Test caching behavior and delegation to underlying repository.

---

## 3. Step-by-Step Task Breakdown

1. Create `services/api/internal/repository/session_repository_test.go`
2. Create `services/api/internal/repository/tenant_repository_test.go`
3. Create `services/api/internal/repository/session_wrapper_test.go`
4. Create `services/api/internal/repository/tenant_wrapper_test.go`
5. Run `go test -v ./services/api/internal/repository/...`
6. Run `audit_codebase.py` to confirm coverage findings drop by 4.

---

## 4. Definition of Done (DoD)
- [ ] Test files created for all 4 repository modules.
- [ ] All tests compile and pass cleanly without race conditions.
- [ ] Audit scanner confirms repository coverage gap resolved.
