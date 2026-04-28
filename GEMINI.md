# Project Rules & Standards (GEMINI)

This document contains the core rules and standards for the CommerceHub project. All AI assistants should follow these guidelines strictly when modifying or adding code.

## 1. General Principles
- **Clarity over Cleverness**: Write readable, maintainable Go code. Follow "Effective Go" principles.
- **Monorepo Structure**: Keep services in `backend/service/` and shared logic in `backend/lib/`.
- **Consistency**: Match the existing coding style, including variable naming (camelCase), file naming (snake_case), and directory structure.

## 2. Backend (Go)
- **Frameworks**: 
  - **Gin** for HTTP routing.
  - **Uber Fx** for Dependency Injection. Use `fx.Provide` and `fx.Invoke` in `internal/bootstrap/`.
  - **Gorm** for database operations.
  - **Zap** for logging. Use `log.Info`, `log.Error`, etc., from the `common/log` package.
- **Architecture**: Follow the established 5-layer architecture:
  1. `cmd/`: Entry points.
  2. `internal/present/http/controller/`: HTTP handlers. Validate requests here.
  3. `internal/core/service/`: Business logic. No infrastructure details here.
  4. `internal/repository/`: Data persistence (PostgreSQL/Gorm/Cache wrappers).
  5. `internal/infrastructure/`: Low-level clients (DB connections, Redis, etc.).
- **DTOs & Responses**:
  - Define all requests and responses in `internal/present/http/dto/`.
  - Use `response.Response[T]` for all API responses to ensure consistency and proper Swagger documentation.
- **Error Handling**: Use the `errors` package from `gitlab.com/ecommercehub1/lib/pkg/errors`.
  - Return `*errors.Error` from services and repositories.
  - Map errors to appropriate HTTP status codes in controllers.

## 3. API Documentation (Swagger)
- Every new endpoint must have **Swagger annotations**.
- Follow the existing pattern in controllers:
  ```go
  // @Summary Action name
  // @Description Detailed description
  // @Tags TagName
  // @Accept json
  // @Produce json
  // @Security BearerAuth (if applicable)
  // @Param request body dto.YourRequest true "Request data"
  // @Success 200 {object} response.Response[dto.YourResponse]
  // @Router /path [method]
  ```
- Run `make swagger` after adding or updating endpoints.

## 4. Database & Migrations
- Use **Atlas** for migrations.
- Models should be defined in the repository layer using Gorm tags.
- Run `make migrate-diff` to generate new migrations and `make migrate-apply` to apply them.
- Use `make migrate svc=all` to apply every supported database migration through the migration image, or `make migrate svc=user` for a single database.

## 5. Frontend (Next.js)
- Use **TypeScript** with strict types.
- Use **Tailwind CSS** for styling.
- Follow the App Router structure in `frontend/app/`.
- Modularize logic into `frontend/lib/services/` for API calls.

## 6. AI Interaction Rules
- **Plan first**: Always describe the plan before making non-trivial changes.
- **Check Workspaces**: Be aware of the `go.work` file. Do not use `replace` directives in `go.mod` unless absolutely necessary and documented.
- **Update Documentation**: If you change standard patterns, reflect them in this `GEMINI.md`.
- **Validation**: When fixing bugs, look for validation tags in DTOs and ensure the frontend matches the requirements.
