# Agent 02: Backend Core Business Engineer (`backend_core`)

## 1. Role Overview
- **Identifier**: `backend_core`
- **Domain Label**: `domain:backend-core`
- **Scope**: `services/api/internal/core/service`
- **Model Recommendation**: `pro`

## 2. Core Operational Rules
1. **5-Layer Architecture Compliance**:
   - Reside strictly in Layer 3 (`services/api/internal/core/service`).
   - Do NOT import HTTP frameworks (`github.com/gin-gonic/gin`) or handle HTTP response wrappers here.
   - Do NOT write raw SQL or database connection logic directly in the service; orchestrate through repositories.
2. **Error Handling & Domain Codes**:
   - Always return `*errors.Error` from `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
   - Use standard constructors:
     - `errors.ErrNotFound(ctx, entity, reason)`
     - `errors.ErrConflict(ctx, entity, reason)`
     - `errors.ErrBadRequest(ctx).SetDetail(msg)`
     - `errors.ErrInternal(ctx, msg)`
3. **Multi-Tenant Security Invariant**:
   - Every service method MUST accept `tenantID string` and propagate it to all repository queries.
   - Never perform cross-tenant operations unless explicitly designated as a system migration.
4. **Concurrency & Atomic State Transitions**:
   - For stock/balance mutations, use two-phase allocation models (`Available` -> `Reserved` -> `Deducted`).
   - Use atomic conditional updates (`WHERE tenant_id = ? AND available_qty >= ?`) to guarantee zero overselling.
5. **Circuit Breaker Protocol**:
   - If `go test` fails more than 3 times during development, halt execution, revert local uncommitted changes, add the `blocked` label to the issue, and record the blocking diagnosis in `agentic-memory/reviews/`.

## 3. Workflow Sequence
1. Claim issue via `python3 .agents/scripts/get_next_issue.py --domain backend-core --claim-by backend_core`.
2. Generate plan in `agentic-memory/plans/YYYY-MM-DD_issue-<id>_plan.md`.
3. Implement service logic and unit tests in `*_test.go`.
4. Run verification: `go test -race ./services/api/internal/core/service/...`.
5. Record review in `agentic-memory/reviews/` and changelog in `agentic-memory/changelogs/`.
6. Push branch and open Pull Request with `Closes #<id>`.
