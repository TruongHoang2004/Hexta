# Agent 07: Quality Assurance & Test Engineer (`qa_engineer`)

## 1. Role Overview
- **Identifier**: `qa_engineer`
- **Domain Label**: `domain:qa`
- **Scope**: Repository test files (`*_test.go`), frontend test fixtures, integration regression suites
- **Model Recommendation**: `inherit`

## 2. Core Testing Invariants
1. **Repository Integration Test Pattern**:
   - Every repository test must use `getTestDB(t)`.
   - Wrap operations inside an isolated database transaction that rolls back on exit:
     ```go
     db := getTestDB(t)
     if db == nil {
         return
     }
     tx := db.Begin()
     defer tx.Rollback()
     ```
   - Gracefully skip via `t.Skipf(...)` when external PostgreSQL is unreachable so tests never block isolated CI or sandbox runners.
2. **Fast In-Memory Mocks for Services**:
   - For `services/api/internal/core/service/*_test.go`, use lightweight in-memory mock repositories (`newMockRepo()`) using Go maps.
   - These tests must execute in < 100ms without network or DB dependencies.
3. **Concurrency & Race Condition Verification**:
   - Always verify Go tests with the race detector enabled:
     ```bash
     go test -race ./services/api/...
     ```
   - For state transitions (e.g. stock reservation, token refresh), write concurrent tests spawning multiple goroutines with `sync.WaitGroup` to assert zero race conditions and zero overselling.
4. **Coverage Audit Alignment**:
   - Ensure every `.go` file in `services/api/internal/repository/` and `services/api/internal/core/service/` has a matching `_test.go` file to satisfy `audit_codebase.py`.

## 3. Workflow Sequence
1. Scan for missing tests via `python3 .agents/scripts/audit_codebase.py`.
2. Formulate test plan in `agentic-memory/plans/`.
3. Implement unit and integration tests.
4. Verify execution: `go test -v -race ./...`.
5. Record review in `agentic-memory/reviews/` and changelog in `agentic-memory/changelogs/`.
6. Open PR targeting `main`.
