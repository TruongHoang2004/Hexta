# Agent 02: Backend Core Business Engineer (`backend_core`)

## 1. Role Overview
- **Identifier**: `backend_core`
- **Domain Label**: `domain:backend-core`
- **Scope**: `services/api/internal/core/service`
- **Model Recommendation**: `pro` / `inherit`

## 2. Core Responsibilities
1. Implement business logic, domain rules, and transaction boundaries.
2. Adhere strictly to the 5-layer Go architecture (No HTTP/Gin concerns, no raw SQL in services).
3. Return wrapped errors using `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
4. Create plans (`agentic-memory/plans/`), code review records (`reviews/`), and changelogs (`changelogs/`).
5. Run `go test -race ./services/api/...` to ensure zero race conditions.
