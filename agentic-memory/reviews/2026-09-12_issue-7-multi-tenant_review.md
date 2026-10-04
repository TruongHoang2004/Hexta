# Code Review: Multi-Tenant Database Models, Service Domain, and REST API

- **Issue Reference**: [#7](https://github.com/TruongHoang2004/Hexta/issues/7)
- **Branch**: `task/issue-7-feat-api-implement-multi-tenant-database`
- **Date**: 2026-09-12
- **Reviewer**: Autonomous Task Runner Agent

---

## 1. Overview
This review covers the complete multi-tenant implementation across the backend Go service (`services/api`), the TypeScript SDK (`packages/sdk`), and the Next.js web application (`apps/web`).

### Scope of Changes:
- Database schema and Atlas migration for `tenants` and `tenant_members`.
- Domain models in `internal/core/model/tenants.go`.
- Redis caching layer in `internal/infrastructure/cache/tenant_cache.go`.
- Data persistence layer in `internal/repository/tenant_repository.go` and `tenant_wrapper.go`.
- Service domain logic in `internal/core/service/tenant_service.go` and unit tests in `tenant_service_test.go`.
- HTTP presentation layer:
  - JWT Bearer Authentication Middleware in `internal/present/http/middleware/auth.go`.
  - REST Controller in `internal/present/http/controller/tenant_controller.go`.
  - DTOs in `internal/present/http/dto/tenant.go`.
  - Route registration in `internal/present/http/router/router.go`.
- Uber Fx module registrations in `internal/bootstrap/`.
- Regenerated Swagger OpenAPI specs in `docs/`.
- TypeScript SDK methods in `packages/sdk/src/index.ts`.
- Next.js web dashboard in `apps/web/app/(dashboard)/tenant/page.tsx` and fix for Suspense boundary in `apps/web/app/auth/callback/page.tsx`.

---

## 2. The Good
1. **5-Layer Architecture Compliance**:
   - HTTP transport logic is strictly confined to `controller/`.
   - Business rules, validation, and RBAC authorization are in `service/`.
   - Persistence and caching are abstracted in `repository/` and `cache/`.
   - Uber Fx handles dependency injection cleanly without global state.
2. **Strict Multi-Tenant Isolation**:
   - Every read and write endpoint (`GetTenant`, `ListMembers`, `InviteMember`, `UpdateTenant`) verifies that the caller belongs to the tenant.
   - Creation of tenants automatically generates an `owner` membership record in a transactional boundary.
3. **Dual Redis Caching with Invalidation**:
   - `TenantCacheWrapper` transparently caches workspaces by both `ID` and `Slug`.
   - Updates invalidate both cache entries to avoid cache drift.
4. **Standardized Responses & Error Handling**:
   - Uses `response.Response[T]` across all endpoints for consistent OpenAPI specs.
   - Returns structured `*errors.Error` from repository and service layers.
5. **Next.js App Router Compliance**:
   - Added React `<Suspense>` around `useSearchParams()` in `auth/callback/page.tsx`, fixing Next.js static prerendering build errors.
   - Dynamic workspace switcher, creation form, and membership management in `tenant/page.tsx`.

---

## 3. Security & Reliability Audit

### Authentication & Authorization
- **JWT Middleware**:
  - `RequireAuth()` middleware validates `Authorization: Bearer <token>`, parses JWT claims, and stores `AuthInfo` (`UserID`, `SessionID`) in request context.
- **RBAC Enforcement**:
  - Only users with role `owner` or `admin` can invite new members or update tenant settings.
  - Member enumeration requires active membership in the target workspace.

### Concurrency & Data Integrity
- **Unique Constraints**:
  - `idx_tenants_slug` prevents duplicate URL slugs.
  - `idx_tenant_user` on `tenant_members` prevents duplicate memberships for the same user.
- **Transactions**:
  - Tenant creation and initial owner assignment occur inside a database transaction (`tx.Transaction`), guaranteeing atomicity.

---

## 4. Verification Results
- `go build ./services/api/...`: **Passed** (exit code 0).
- `go test -v ./services/api/...`: **Passed** (all unit tests passed including new tenant domain tests).
- `make -C services/api swagger`: **Passed** (all OpenAPI schemas generated).
- `pnpm --filter @ubi/sdk run build`: **Passed** (ESM, CJS, and DTS bundles emitted).
- `pnpm run --recursive build`: **Passed** (all 6 projects compiled cleanly).

---

## 5. Summary & Verdict
The changes meet all architectural, security, and functional criteria outlined in Issue [#7](https://github.com/TruongHoang2004/Hexta/issues/7) and `GEMINI.md`. Approved for merge.
