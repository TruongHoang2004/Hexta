# Code Review: AI Agent Engine with Tool Calling Registry and Draft Proposal Flow

- **Issue**: [#37](https://github.com/TruongHoang2004/Hexta/issues/37)
- **Review Date**: 2026-10-09
- **Reviewer Agent**: `dev-review` (on behalf of `ai_architect` & `qa_engineer`)
- **Target Changes**: Implementation of `gemini.IGeminiClient`, `AIAgentService`, `AIController`, DTOs, Swagger documentation, and test suite.
- **Verdict**: **APPROVED**

---

## 1. Scope & Changes Checked

### 1.1 Infrastructure Layer (`services/api/internal/infrastructure/gemini/`)
- **Client Wrapper (`client.go`)**:
  - Implements `IGeminiClient` interface abstracting Google Gemini API SDK.
  - Supports `GenerateContent` and streaming with function declaration tool registrations.
- **Multi-Tenant Guard**:
  - Encapsulates low-level API operations without exposing provider specifics to the core service.

### 1.2 Core Service Layer (`services/api/internal/core/service/ai_agent_service.go`)
- **Tool Calling Registry**:
  - `search_products`: Queries products and variants matching queries, strictly scoped by `tenant_id`.
  - `check_stock_availability`: Checks SKU existence and available stock quantity, strictly scoped by `tenant_id`.
  - `propose_order_draft`: Calculates line items and totals with `decimal.Decimal`, generates unique draft UUID, and stores draft in Redis with 30m TTL.
- **Zero Direct Database Mutations**:
  - Enforces that no SQL write mutations (`INSERT`/`UPDATE` on orders) occur during AI conversation.
  - Generates structured draft proposals for human review and confirmation.
- **Multi-Tenant Boundary**:
  - Injects `tenant_id` from JWT session context into all tool arguments and validates on draft lookup.

### 1.3 Presentation Layer & Routing (`services/api/internal/present/http/`)
- **DTOs (`dto/ai_dto.go`)**:
  - Defines `AIConverseRequest`, `OrderDraftDTO`, `DraftItemDTO`, and `SSEEvent`.
- **Controller (`controller/ai_controller.go`)**:
  - `POST /api/v1/ai/agent/converse`: Real-time SSE streaming with event flush.
  - `GET /api/v1/ai/agent/drafts/:id`: Retrieves cached draft.
  - Standardized `response.Response[T]` wrappers and OpenAPI Swagger annotations.
- **Router (`router/router.go`)**:
  - Registered under protected `aiGroup` (`/api/v1/ai/agent`) with `AuthMiddleware.RequireAuth()`.

### 1.4 Dependency Injection (`services/api/internal/bootstrap/`)
- Uber Fx bindings for `IGeminiClient`, `IAIAgentService`, and `AIController`.

### 1.5 Test Verification (`services/api/internal/core/service/ai_agent_service_test.go`)
- Zero DB mutations verified against SQLite in-memory DB.
- Multi-tenant product search and stock checking verified.
- Draft proposal and 30-minute Redis cache TTL verified.
- Cross-tenant draft isolation verified.
- Race conditions checked with `go test -race ./...` (All PASSED).

---

## 2. Quality & Architecture Compliance

| Gate | Status | Evidence |
|---|---|---|
| 5-Layer Go Architecture | **PASSED** | Clean boundary separation across Controller, Service, DTO, and Infrastructure. |
| Language Standard | **PASSED** | 100% English identifiers, comments, commit messages, and documentation. |
| Invariant: Zero DB Mutation | **PASSED** | AI Agent emits drafts only; no mutations to `orders` table. |
| Invariant: Multi-Tenant Boundary | **PASSED** | Tenant ID dynamically validated on all queries and draft access. |
| Race Condition Check | **PASSED** | Passed `go test -race ./...` without race warnings. |
