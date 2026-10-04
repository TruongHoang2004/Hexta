# Codebase Audit Report: 2026-09-12

- **Inspection Date**: 2026-09-12 09:18:40 UTC
- **Repository Root**: `/Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner`
- **Total Findings**: 39
- **Issue Candidates Synthesized**: 5

---

## 1. Executive Summary & Health Metrics

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| `architecture` | 3 | 3 | 0 | 0 |
| `coverage` | 5 | 0 | 2 | 3 |
| `debt` | 5 | 0 | 2 | 3 |
| `language` | 20 | 0 | 0 | 20 |
| `swagger` | 6 | 0 | 0 | 6 |

### Severity Breakdown
- **High Severity**: 3
- **Medium Severity**: 4
- **Low Severity**: 32

---

## 2. Issue Generation & Deduplication Actions

- PENDING (Dry Run): Candidate 'fix(arch): resolve 5-layer architecture boundary violations in API controllers' with severity `high` and labels ['bug', 'priority:high']
- SKIPPED (Duplicate): 'refactor(tenant): replace temporary mock tenant IDs with dynamic token session context' - Reason: Semantic conflict with issue #7: 'feat(api): implement multi-tenant database models, service domain, and REST API'
- PENDING (Dry Run): Candidate 'test(coverage): implement unit test suites for critical core services' with severity `medium` and labels ['enhancement', 'priority:medium']
- PENDING (Dry Run): Candidate 'refactor(debt): resolve pending TODO and FIXME debt annotations across services' with severity `low` and labels ['enhancement', 'priority:low']
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
| **MEDIUM** | `services/api/internal/core/service/auth_service.go` | Missing unit test suite for service auth_service.go | Core business logic in auth_service.go lacks corresponding test file auth_service_test.go. |
| **LOW** | `services/api/internal/repository/session_repository.go` | Missing repository test coverage for session_repository.go | Data persistence wrapper in session_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/session_wrapper.go` | Missing repository test coverage for session_wrapper.go | Data persistence wrapper in session_wrapper.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/identify_repository.go` | Missing repository test coverage for identify_repository.go | Data persistence wrapper in identify_repository.go lacks unit or integration test verification. |
| **MEDIUM** | `packages/sdk` | Missing test suite for packages/sdk | SDK package contains source code but no unit or integration tests in packages/sdk/src. |

### Category: `debt` (5 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/core/service/auth_service.go`:51 | Technical debt: TODO comment in auth_service.go | TODO comment found: 'Inject JWT secret from config'. |
| **LOW** | `services/api/internal/core/service/auth_service.go`:174 | Technical debt: TODO comment in auth_service.go | TODO comment found: 'add UpdateToken to session repository if needed, or just leave it for now.'. |
| **LOW** | `services/api/internal/core/service/auth_service.go`:235 | Technical debt: TODO comment in auth_service.go | TODO comment found: 'use secure random state'. |
| **MEDIUM** | `apps/web/app/(dashboard)/tenant/page.tsx`:18 | Technical debt: Mock Tenant Comment in page.tsx | Mock Tenant Comment found: '// Mock hardcoded tenant ID for demo purposes.'. |
| **MEDIUM** | `apps/web/app/(dashboard)/tenant/page.tsx`:19 | Technical debt: Hardcoded Mock Tenant in page.tsx | Hardcoded Mock Tenant found: 'const data = await sdk.identity.getTenant("mock-tenant-id");'. |

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
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:31 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Register | Handler 'Register' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:67 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Login | Handler 'Login' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:105 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for RefreshToken | Handler 'RefreshToken' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:135 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Logout | Handler 'Logout' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:157 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleLogin | Handler 'GoogleLogin' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:173 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleCallback | Handler 'GoogleCallback' uses raw DTO in @Success instead of standard response.Response[T]. |
