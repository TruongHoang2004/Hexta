# Codebase Audit Report: 2026-09-14

- **Inspection Date**: 2026-09-14 08:26:42 UTC
- **Repository Root**: `/home/runner/work/Hexta/Hexta`
- **Total Findings**: 34
- **Issue Candidates Synthesized**: 4

---

## 1. Executive Summary & Health Metrics

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| `architecture` | 3 | 3 | 0 | 0 |
| `coverage` | 5 | 0 | 1 | 4 |
| `language` | 20 | 0 | 0 | 20 |
| `swagger` | 6 | 0 | 0 | 6 |

### Severity Breakdown
- **High Severity**: 3
- **Medium Severity**: 1
- **Low Severity**: 30

---

## 2. Issue Generation & Deduplication Actions

- CREATED: [fix(arch): resolve 5-layer architecture boundary violations in API controllers](https://github.com/TruongHoang2004/Hexta/issues/28) with labels `bug,priority:high`
- SKIPPED (Duplicate): 'refactor(tenant): replace temporary mock tenant IDs with dynamic token session context' - Reason: Semantic conflict with issue #7: 'feat(api): implement multi-tenant database models, service domain, and REST API'
- CREATED: [test(coverage): implement unit test suites for critical core services](https://github.com/TruongHoang2004/Hexta/issues/29) with labels `enhancement,priority:medium`
- SKIPPED (Duplicate): 'chore(i18n): enforce English language rule across source code and comments' - Reason: Semantic conflict with issue #6: 'fix(web): reconcile auth token storage, synchronize JWT claims, and enforce English language rule'

---

## 3. Detailed Audit Findings

### Category: `architecture` (3 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **HIGH** | `services/api/internal/present/http/controller/health_controller.go` | Layer boundary violation: controller directly imports gorm.io/gorm in health_controller.go | Direct GORM import in controller violates 5-layer boundary. Controllers must only communicate with core services or DTOs. |
| **HIGH** | `services/api/internal/present/http/controller/health_controller.go`:16 | Controller struct contains direct database client reference in health_controller.go | Controllers must not hold direct database connections or perform DB operations directly. |
| **HIGH** | `services/api/internal/present/http/controller/health_controller.go`:22 | Controller struct contains direct database client reference in health_controller.go | Controllers must not hold direct database connections or perform DB operations directly. |

### Category: `coverage` (5 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/repository/tenant_wrapper.go` | Missing repository test coverage for tenant_wrapper.go | Data persistence wrapper in tenant_wrapper.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/tenant_repository.go` | Missing repository test coverage for tenant_repository.go | Data persistence wrapper in tenant_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/session_wrapper.go` | Missing repository test coverage for session_wrapper.go | Data persistence wrapper in session_wrapper.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/session_repository.go` | Missing repository test coverage for session_repository.go | Data persistence wrapper in session_repository.go lacks unit or integration test verification. |
| **MEDIUM** | `packages/sdk` | Missing test suite for packages/sdk | SDK package contains source code but no unit or integration tests in packages/sdk/src. |

### Category: `language` (20 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/present/http/validator/validator.go`:49 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:50 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:51 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:52 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:53 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:54 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:55 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:56 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:104 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `services/api/internal/present/http/validator/validator.go`:105 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:47 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:48 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:49 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:50 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:51 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:52 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:53 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:54 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:104 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
| **LOW** | `packages/shared/pkg/validator/validator.go`:105 | Non-English comment detected in validator.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |

### Category: `swagger` (6 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:37 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Register | Handler 'Register' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:73 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Login | Handler 'Login' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:109 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for RefreshToken | Handler 'RefreshToken' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:140 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Logout | Handler 'Logout' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:176 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleLogin | Handler 'GoogleLogin' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:193 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleCallback | Handler 'GoogleCallback' uses raw DTO in @Success instead of standard response.Response[T]. |
