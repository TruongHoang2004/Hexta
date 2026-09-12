# Development Rules & Coding Standards

This document defines the core development rules and engineering standards across the Hexta / CommerceHub codebase.

---

## 1. Language Standard
- **Mandatory English**: All documentation, code comments, commit messages, variable/type naming, API schemas, and agentic memory artifacts (`plans/`, `designs/`, `reviews/`, `changelogs/`, `rules/`) **MUST be written in English**.

---

## 2. Backend Architecture (Go / Gin)
- **5-Layer Architecture**:
  1. `cmd/`: Application entry points.
  2. `internal/present/http/controller/`: HTTP handlers. Validate incoming requests here.
  3. `internal/core/service/`: Business logic. Strictly free of HTTP transport or raw database concerns.
  4. `internal/repository/`: Data persistence (PostgreSQL / GORM / cache wrappers).
  5. `internal/infrastructure/`: Low-level clients and drivers (DB connections, Redis, brokers).
- **Dependency Injection**: Managed via **Uber Fx** (`fx.Provide`, `fx.Invoke` in `internal/bootstrap/`).
- **Error Handling**:
  - Always return `*errors.Error` from service and repository layers.
  - Use `gitlab.com/ecommercehub1/shared/pkg/errors` (or `lib/pkg/errors`).
  - Controllers map errors to corresponding HTTP status codes.
- **DTOs & Standardized Responses**:
  - Define all requests and responses in `internal/present/http/dto/`.
  - Use `response.Response[T]` for all API responses to ensure consistency and OpenAPI compliance.
- **Swagger Documentation**:
  - Every HTTP endpoint must have full Swagger annotations (`@Summary`, `@Description`, `@Tags`, `@Param`, `@Success`, `@Router`).

---

## 3. Database & Migrations
- Use **Atlas** for version-controlled schema migrations.
- Models must be defined in the repository layer using standard GORM tags.
- Foreign keys and indexes must be declared explicitly. Avoid queries inside loops (N+1 query problem).

---

## 4. Frontend (Next.js / TypeScript)
- Use **TypeScript** with strict type checking (avoid `any`).
- Use **Tailwind CSS** for styling.
- Follow Next.js App Router conventions (`apps/web/app/`).
- Wrap client-side hooks such as `useSearchParams()` inside React `<Suspense>` boundaries.
- Modularize API calls into client services; never hardcode backend host URLs.

---

## 5. Security & Authentication
- **OAuth2**:
  - Always generate a cryptographically secure random `state` (`crypto/rand`) and validate it in the callback to prevent CSRF attacks.
  - Never transmit Access Tokens or Refresh Tokens in URL query parameters.
  - Store Refresh Tokens in secure `HttpOnly`, `SameSite=Lax` cookies.
  - Implement proper account linking: if a user logs in via OAuth with an email that already exists under another provider, associate the identity with the existing user account.

---

## 6. Agentic Memory Protocol
- When executing non-trivial development workflows, the agent must persist artifacts to:
  1. **Planning**: `agentic-memory/plans/YYYY-MM-DD_<feature>_plan.md`
  2. **Technical Design**: `agentic-memory/designs/YYYY-MM-DD_<feature>_design.md`
  3. **Code Review**: `agentic-memory/reviews/YYYY-MM-DD_<feature>_review.md`
  4. **Changelogs & Walkthroughs**: `agentic-memory/changelogs/YYYY-MM-DD_<feature>_changelog.md`
