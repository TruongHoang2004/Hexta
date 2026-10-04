# Implementation Plan: Multi-Tenant Database Models, Service Domain, and REST API

- **Issue Reference**: [#7](https://github.com/TruongHoang2004/Hexta/issues/7)
- **Branch**: `task/issue-7-feat-api-implement-multi-tenant-database`
- **Date**: 2026-09-12
- **Author**: Autonomous Task Runner Agent

---

## 1. Overview & Goal
The objective of this task is to implement the end-to-end multi-tenant domain architecture in `services/api` and connect it with `@ubi/sdk` and `apps/web`.
Currently, the tenant dashboard in `apps/web/app/(dashboard)/tenant/page.tsx` and the client SDK rely on mock values. In the backend, there are no database entities, migrations, repositories, service domain logic, or REST controllers for managing workspaces and memberships.

### Key Deliverables:
1. GORM data models for `tenants` and `tenant_members`.
2. Atlas SQL database migrations with correct constraints and indexes.
3. Redis caching layer for tenant resolution by ID and Slug.
4. Repository layer with database persistence and cache wrapper.
5. Service domain layer enforcing business rules, RBAC, and membership isolation.
6. HTTP REST Controller, request/response DTOs, and JWT Authentication Middleware.
7. Swagger OpenAPI documentation updates.
8. `@ubi/sdk` updates for tenant management.
9. Dynamic workspace data integration in Next.js web application.

---

## 2. Current State Analysis
- **Backend (`services/api`)**:
  - Contains authentication modules (`AuthIdentities`, `Sessions`).
  - Architecture follows a strict 5-layer Go pattern with Uber Fx DI.
  - Currently lacks a tenant domain and an HTTP authentication middleware to validate JWT Bearer tokens and extract user identity into context.
- **Frontend & SDK**:
  - `@ubi/sdk` defines basic `Tenant` and `User` interfaces with `getTenant(id)` and `getUsers(tenantId)` calling `/api/v1/tenants/:id` endpoints that currently 404.
  - `apps/web/app/(dashboard)/tenant/page.tsx` hardcodes `"mock-tenant-id"` with fallback demo data.
  - `apps/web/app/auth/callback/page.tsx` needs a React `<Suspense>` wrapper around `useSearchParams()` to satisfy Next.js build prerendering requirements.

---

## 3. Architecture & Technical Design

### 3.1 Database Schema & Migration
- **Table: `tenants`**:
  - `id`: VARCHAR(36) PRIMARY KEY (UUID format)
  - `name`: VARCHAR(255) NOT NULL
  - `slug`: VARCHAR(255) NOT NULL UNIQUE
  - `plan`: VARCHAR(50) NOT NULL DEFAULT 'free' (values: `free`, `pro`, `enterprise`)
  - `status`: VARCHAR(50) NOT NULL DEFAULT 'active' (values: `active`, `suspended`)
  - `owner_id`: VARCHAR(255) NOT NULL (INDEX: `idx_tenants_owner_id`)
  - `created_at`: TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
  - `updated_at`: TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
- **Table: `tenant_members`**:
  - `id`: BIGSERIAL PRIMARY KEY
  - `tenant_id`: VARCHAR(36) NOT NULL (FOREIGN KEY to `tenants(id)` ON DELETE CASCADE)
  - `user_id`: VARCHAR(255) NOT NULL
  - `role`: VARCHAR(50) NOT NULL DEFAULT 'member' (values: `owner`, `admin`, `member`)
  - `created_at`: TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
  - `updated_at`: TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
  - UNIQUE INDEX on `(tenant_id, user_id)`: `idx_tenant_members_tenant_user`
  - INDEX on `user_id`: `idx_tenant_members_user_id`

### 3.2 5-Layer Go Architecture
1. **Core Models (`internal/core/model/tenants.go`)**:
   - `Tenant` and `TenantMember` GORM structs with JSON and GORM tags.
2. **Infrastructure Cache (`internal/infrastructure/cache/tenant_cache.go`)**:
   - `TenantCache` using `RedisClient` with cache keys `tenant:id:%s`, `tenant:slug:%s`.
3. **Repository Layer (`internal/repository/tenant_repository.go`, `tenant_wrapper.go`)**:
   - `ITenantRepository` interface covering transactional creation, retrieval by ID/slug, update, user workspaces, and member management.
   - `TenantRepository` implementing GORM queries.
   - `TenantCacheWrapper` implementing `ITenantRepository` with transparent Redis read-through and cache invalidation.
4. **Core Service Layer (`internal/core/service/tenant_service.go`)**:
   - `TenantService` implementing:
     - `CreateTenant(ctx, userID, req)`: Generates UUID, checks slug uniqueness, saves tenant and assigns creator as `owner`.
     - `GetTenant(ctx, userID, tenantID)`: Verifies caller's membership and returns workspace details.
     - `GetUserTenants(ctx, userID)`: Lists all workspaces the user belongs to.
     - `ListMembers(ctx, userID, tenantID)`: Verifies membership and lists all workspace members.
     - `InviteMember(ctx, currentUserID, tenantID, targetUserID, role)`: Verifies caller is `owner` or `admin`, ensures user isn't already a member, and creates membership record.
     - `UpdateTenant(ctx, currentUserID, tenantID, req)`: Verifies caller is `owner` or `admin` and updates settings.
5. **Presentation Layer**:
   - **Middleware (`internal/present/http/middleware/auth.go`)**:
     - Validates `Authorization: Bearer <token>`, extracts `userID` and `sessionID`, populates context via `common.SetAuthInfo`.
   - **DTOs (`internal/present/http/dto/tenant.go`)**:
     - `CreateTenantRequest`, `UpdateTenantRequest`, `InviteMemberRequest`, `TenantResponse`, `TenantMemberResponse`.
   - **Controller (`internal/present/http/controller/tenant_controller.go`)**:
     - `POST /api/v1/tenants`: Create workspace.
     - `GET /api/v1/tenants`: List current user's workspaces.
     - `GET /api/v1/tenants/:id`: Get workspace details.
     - `PUT /api/v1/tenants/:id`: Update workspace settings.
     - `GET /api/v1/tenants/:id/users`: List workspace members.
     - `POST /api/v1/tenants/:id/invites`: Invite member.
   - **Router (`internal/present/http/router/router.go`)**:
     - Register routes under authenticated router group.

### 3.3 Uber Fx Dependency Injection
- Register cache in `internal/bootstrap/cache.go`.
- Register repositories in `internal/bootstrap/repository.go`.
- Register service in `internal/bootstrap/service.go`.
- Register controller and middleware in `internal/bootstrap/controller.go` and `internal/bootstrap/middleware.go`.
- Wire routes in `internal/bootstrap/router.go`.

### 3.4 SDK & Frontend Integration
- **`@ubi/sdk`**:
  - Export types `Tenant`, `TenantMember`, `CreateTenantInput`, `InviteMemberInput`.
  - Add client methods `createTenant`, `getUserTenants`, `inviteMember`, and ensure `getTenant` & `getUsers` adhere to API response format.
- **`apps/web`**:
  - Fix `<Suspense>` boundary in `apps/web/app/auth/callback/page.tsx`.
  - Update `apps/web/app/(dashboard)/tenant/page.tsx` to dynamically query user workspaces via SDK or API, display details, show members list, and handle empty states.

---

## 4. Step-by-Step Task Breakdown

1. **Database & Migration**:
   - Create `services/api/internal/core/model/tenants.go`.
   - Create migration SQL file `migrations/api/20260912160000_create_tenants.sql`.
   - Update Atlas migration checksum hash (`atlas migrate hash`).
2. **Infrastructure & Repository**:
   - Create `services/api/internal/infrastructure/cache/tenant_cache.go`.
   - Create `services/api/internal/repository/tenant_repository.go`.
   - Create `services/api/internal/repository/tenant_wrapper.go`.
3. **Authentication Middleware & Core Service**:
   - Create `services/api/internal/present/http/middleware/auth.go`.
   - Create `services/api/internal/core/service/tenant_service.go`.
   - Add unit tests for service layer in `services/api/internal/core/service/tenant_service_test.go`.
4. **HTTP Presentation (DTOs, Controller, Router)**:
   - Create `services/api/internal/present/http/dto/tenant.go`.
   - Create `services/api/internal/present/http/controller/tenant_controller.go`.
   - Wire dependencies in `services/api/internal/bootstrap/`.
   - Configure routes in `services/api/internal/present/http/router/router.go`.
   - Regenerate Swagger documentation (`make -C services/api swagger`).
5. **SDK & Web Integration**:
   - Update `packages/sdk/src/index.ts` and compile (`pnpm --filter @ubi/sdk run build`).
   - Fix Suspense boundary in `apps/web/app/auth/callback/page.tsx`.
   - Update `apps/web/app/(dashboard)/tenant/page.tsx` for dynamic multi-tenant display.
6. **Verification & Testing**:
   - Run `go build ./services/api/...`
   - Run `go test -v ./services/api/...`
   - Run `pnpm run --recursive build` to ensure all TypeScript packages and Next.js apps build successfully.
7. **Quality Audit & Documentation**:
   - Formulate code review report in `agentic-memory/reviews/2026-09-12_issue-7-multi-tenant_review.md`.
   - Formulate changelog in `agentic-memory/changelogs/2026-09-12_issue-7-multi-tenant_changelog.md`.
8. **Git Commit, Push & PR**:
   - Commit changes referencing Issue #7.
   - Push branch and create Pull Request via `gh pr create`.
   - Comment on Issue #7 via `gh issue comment`.

---

## 5. Risk Assessment & Mitigations
- **Tenant Data Isolation**: Users could access workspaces they do not belong to.
  - *Mitigation*: The service layer checks caller membership in every operation (`GetTenant`, `ListMembers`, `InviteMember`, `UpdateTenant`).
- **Concurrent Slug Clashing**:
  - *Mitigation*: Unique database constraint on `slug` column and early duplicate validation in `CreateTenant`.
- **Cache Drift**:
  - *Mitigation*: Cache wrapper invalidates both `tenant:id:%s` and `tenant:slug:%s` upon mutation.

---

## 6. Definition of Done (DoD)
- [x] All Go packages in `services/api` compile cleanly without errors or warnings.
- [x] Service and repository tests pass.
- [x] Swagger docs updated and valid.
- [x] SDK builds without errors.
- [x] Next.js web application builds without build/prerender errors.
- [x] Agentic memory artifacts (`plan`, `review`, `changelog`) created and committed.
- [x] Pull Request opened against `main` and linked to Issue #7.
