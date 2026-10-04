# Changelog: Multi-Tenant Database Models, Service Domain, and REST API

- **Issue Reference**: [#7](https://github.com/TruongHoang2004/Hexta/issues/7)
- **Branch**: `task/issue-7-feat-api-implement-multi-tenant-database`
- **Date**: 2026-09-12
- **Author**: Autonomous Task Runner Agent

---

## 1. Change Summary
Implemented the core multi-tenant domain architecture for Hexta / CommerceHub, replacing static mock values across the frontend and SDK with a production-grade backend service, Atlas database schema, Redis caching, RBAC authorization, and dynamic Next.js UI integration.

---

## 2. Impacted Components & Files

| Component | File Path | Action | Description |
|---|---|---|---|
| **Database** | `migrations/api/20260912160000_create_tenants.sql` | `[NEW]` | Atlas migration for `tenants` and `tenant_members` tables |
| **Database** | `migrations/api/atlas.sum` | `[MODIFY]` | Atlas migration checksum file updated |
| **Core Models** | `services/api/internal/core/model/tenants.go` | `[NEW]` | `Tenant` and `TenantMember` GORM models and role constants |
| **Cache** | `services/api/internal/infrastructure/cache/tenant_cache.go` | `[NEW]` | Redis cache for tenants with ID and slug lookups |
| **Repository** | `services/api/internal/repository/tenant_repository.go` | `[NEW]` | `ITenantRepository` interface and GORM database repository |
| **Repository** | `services/api/internal/repository/tenant_wrapper.go` | `[NEW]` | `TenantCacheWrapper` implementing Redis cache-through |
| **Service** | `services/api/internal/core/service/tenant_service.go` | `[NEW]` | `TenantService` business logic and RBAC membership validation |
| **Service** | `services/api/internal/core/service/tenant_service_test.go` | `[NEW]` | Unit tests for tenant creation, slug conflict, and invite permissions |
| **Service** | `services/api/internal/core/service/auth_service.go` | `[MODIFY]` | Exported `ParseToken` on `IAuthService` and `AuthService` |
| **Middleware** | `services/api/internal/present/http/middleware/auth.go` | `[NEW]` | `AuthMiddleware.RequireAuth()` validating Bearer JWT tokens |
| **DTOs** | `services/api/internal/present/http/dto/tenant.go` | `[NEW]` | Request and response DTOs for tenant endpoints |
| **Controller** | `services/api/internal/present/http/controller/tenant_controller.go` | `[NEW]` | REST controller handling workspace and membership routes |
| **Router** | `services/api/internal/present/http/router/router.go` | `[MODIFY]` | Registered protected `/tenants` routes in Gin engine |
| **Bootstrap** | `services/api/internal/bootstrap/*.go` | `[MODIFY]` | Registered tenant cache, repo, service, controller, and middleware in Uber Fx |
| **API Docs** | `services/api/docs/*` | `[MODIFY]` | Regenerated Swagger OpenAPI specifications |
| **SDK** | `packages/sdk/src/index.ts` | `[MODIFY]` | Extended `IdentityClient` with tenant CRUD, user listing, and invitations |
| **SDK** | `packages/sdk/dist/*` | `[MODIFY]` | Compiled ESM, CJS, and DTS bundles for `@ubi/sdk` |
| **Web Frontend** | `apps/web/app/auth/callback/page.tsx` | `[MODIFY]` | Wrapped `useSearchParams()` in React `<Suspense>` boundary |
| **Web Frontend** | `apps/web/app/(dashboard)/tenant/page.tsx` | `[MODIFY]` | Dynamic workspace dashboard with creation, switcher, and member list |
| **Agentic Memory** | `agentic-memory/plans/2026-09-12_issue-7-multi-tenant_plan.md` | `[NEW]` | Development implementation plan |
| **Agentic Memory** | `agentic-memory/reviews/2026-09-12_issue-7-multi-tenant_review.md` | `[NEW]` | Code review audit report |
| **Agentic Memory** | `agentic-memory/changelogs/2026-09-12_issue-7-multi-tenant_changelog.md` | `[NEW]` | This changelog record |

---

## 3. Key Technical Decisions
- **Multi-Tenant Membership Isolation**:
  - Implemented strict checks in `TenantService` verifying the caller belongs to the target workspace before allowing data reads.
  - Restricted invitation and settings updates to workspace `owner` and `admin` roles.
- **Transactional Tenant Creation**:
  - Tenant creation writes the `Tenant` entity and immediately creates a `TenantMember` record with role `owner` in a single transaction, preventing orphaned workspaces.
- **Dual Cache Lookups with Automatic Invalidation**:
  - Redis cache stores tenants by both `id` and `slug`. Any workspace update immediately purges both keys to guarantee consistency.
- **Next.js App Router Suspense Fix**:
  - Wrapped `useSearchParams()` in `apps/web/app/auth/callback/page.tsx` inside `<Suspense>`, resolving the static prerender build failure during `next build`.

---

## 4. Verification & Testing Guide

### Automated Tests
1. Backend builds and tests:
   ```bash
   go build ./services/api/...
   go test -v ./services/api/...
   ```
2. Swagger documentation generation:
   ```bash
   make -C services/api swagger
   ```
3. SDK and frontend builds:
   ```bash
   pnpm --filter @ubi/sdk run build
   pnpm run --recursive build
   ```

### Manual Verification
1. **API Endpoints**:
   - `POST /api/v1/tenants`: Create a workspace with Authorization header.
   - `GET /api/v1/tenants`: List workspaces for authenticated user.
   - `GET /api/v1/tenants/:id`: Retrieve workspace details.
   - `GET /api/v1/tenants/:id/users`: List team members in the workspace.
   - `POST /api/v1/tenants/:id/invites`: Invite a user to the workspace.
2. **Web Application**:
   - Navigate to `/tenant` in browser.
   - If user has no workspaces, verify the "Create Workspace" form is displayed.
   - Create a workspace; verify workspace details (Name, Slug, Plan, Status) and team members are rendered dynamically.
