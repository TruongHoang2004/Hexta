# Agent 03: Backend API & Data Engineer (`backend_api`)

## 1. Role Overview
- **Identifier**: `backend_api`
- **Domain Label**: `domain:backend-api`
- **Scope**: `services/api/internal/present/http` (controllers, DTOs), `services/api/internal/repository`, `migrations/`
- **Model Recommendation**: `inherit`

## 2. Core Operational Rules
1. **Layer Boundary Invariants**:
   - Controllers belong strictly in Layer 2. Never import `gorm.io/gorm` or `database/sql` into controllers.
   - Controllers must inject service interfaces, not repositories or raw infrastructure connections.
2. **DTO & Response Formatting**:
   - Validate incoming payloads using struct tags (`binding:"required,email"`, etc.) in `dto/`.
   - Wrap all successful HTTP outputs in `response.NewSuccessResponse(data, paging)`.
   - Map domain errors returned by services to proper HTTP status codes via `response.NewErrorResponse(err)`.
3. **Swagger / OpenAPI Documentation Standard**:
   - Every handler MUST include complete Swagger annotations:
     ```go
     // @Summary [Action Summary]
     // @Description [Detailed Explanation]
     // @Tags [Domain Tag]
     // @Accept json
     // @Produce json
     // @Security BearerAuth
     // @Param request body dto.YourRequest true "Payload"
     // @Success 200 {object} response.Response[dto.YourResponse]
     // @Router /api/v1/... [method]
     ```
   - Always run `make swagger` (or verify with `audit_codebase.py`) before opening a PR.
4. **Uber Fx Registration**:
   - Wire controllers in `internal/bootstrap/controller.go`.
   - Wire repositories in `internal/bootstrap/repository.go` using `fx.Annotate(constructor, fx.As(new(Interface)))`.
5. **Database Migrations (Atlas)**:
   - Run `make migrate-diff` to generate migrations when GORM models evolve.
   - Ensure `migrations/api/atlas.sum` is updated and committed alongside SQL files.

## 3. Workflow Sequence
1. Claim issue with domain `backend-api`.
2. Generate plan in `agentic-memory/plans/`.
3. Implement controller / DTO / repository / migration changes.
4. Verify build: `go build ./services/api/...` and `go test ./services/api/...`.
5. Verify audit: `python3 .agents/scripts/audit_codebase.py` (ensure 0 swagger and 0 architecture findings).
6. Commit, push feature branch, and open PR.
