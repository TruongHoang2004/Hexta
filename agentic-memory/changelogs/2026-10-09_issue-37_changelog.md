# Changelog: AI Agent Engine with Tool Calling Registry and Draft Proposal Flow

- **Issue**: [#37](https://github.com/TruongHoang2004/Hexta/issues/37)
- **Date**: 2026-10-09
- **Domain**: `domain:ai-agent`
- **Component**: `services/api`
- **Author**: `ai_architect` agent

---

## Summary of Changes

Implemented the conversational AI Agent service and HTTP SSE handlers in Go connecting to Google Gemini API via official SDK, establishing schema-governed tool calling registry (`search_products`, `check_stock_availability`, `propose_order_draft`), Redis draft caching (30-min TTL), and Server-Sent Events (SSE) streaming endpoint.

### Key Changes:
1. **`services/api/internal/infrastructure/gemini/client.go`**:
   - `IGeminiClient` interface and wrapper around Gemini API.
   - Support for `GenerateContent` and streaming with function calling tool definitions.
2. **`services/api/internal/core/service/ai_agent_service.go`**:
   - `IAIAgentService` interface and business logic.
   - Tool calling declarations and execution dispatch: `search_products`, `check_stock_availability`, `propose_order_draft`.
   - Invariant: Zero direct database mutations executed by AI Agent; order drafts cached in Redis with 30-min TTL.
   - Multi-tenant boundary enforcement: dynamic injection and validation of `tenant_id` on all queries and draft access.
3. **`services/api/internal/present/http/`**:
   - `dto/ai_dto.go`: Request, response, and SSE event schemas.
   - `controller/ai_controller.go`: SSE streaming handler (`POST /api/v1/ai/agent/converse`) and draft lookup (`GET /api/v1/ai/agent/drafts/:id`).
   - `router/router.go`: Protected `/api/v1/ai/agent` route group.
4. **`services/api/internal/bootstrap/`**:
   - Uber Fx bindings in `service.go` and `controller.go`.
5. **Swagger Documentation**:
   - Updated Swagger specs (`docs.go`, `swagger.json`, `swagger.yaml`).
6. **Unit Tests**:
   - `services/api/internal/core/service/ai_agent_service_test.go`: Comprehensive test suite verifying zero DB mutations, multi-tenant queries, draft caching, and tenant isolation.

---

## Verification
- `go test -race ./...` in `services/api`: Passed cleanly with race detection.
- `go vet ./...` in `services/api`: Clean, zero issues.
