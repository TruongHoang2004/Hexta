# Codebase Audit Report: 2026-10-19

- **Inspection Date**: 2026-10-19 08:22:14 UTC
- **Repository Root**: `/home/runner/work/Hexta/Hexta`
- **Total Findings**: 30
- **Issue Candidates Synthesized**: 3

---

## 1. Executive Summary & Health Metrics

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| `coverage` | 2 | 0 | 0 | 2 |
| `language` | 20 | 0 | 0 | 20 |
| `swagger` | 8 | 0 | 2 | 6 |

### Severity Breakdown
- **High Severity**: 0
- **Medium Severity**: 2
- **Low Severity**: 28

---

## 2. Issue Generation & Deduplication Actions

- CREATED: [docs(swagger): add missing OpenAPI annotations to API controller endpoints](https://github.com/TruongHoang2004/Hexta/issues/56) with labels `enhancement,priority:medium`
- SKIPPED (Duplicate): 'refactor(tenant): replace temporary mock tenant IDs with dynamic token session context' - Reason: Semantic conflict with issue #7: 'feat(api): implement multi-tenant database models, service domain, and REST API'
- SKIPPED (Duplicate): 'chore(i18n): enforce English language rule across source code and comments' - Reason: Semantic conflict with issue #6: 'fix(web): reconcile auth token storage, synchronize JWT claims, and enforce English language rule'

---

## 3. Detailed Audit Findings

### Category: `coverage` (2 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/repository/sales_repository.go` | Missing repository test coverage for sales_repository.go | Data persistence wrapper in sales_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/tenant_onboarding_repository.go` | Missing repository test coverage for tenant_onboarding_repository.go | Data persistence wrapper in tenant_onboarding_repository.go lacks unit or integration test verification. |

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

### Category: `swagger` (8 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **MEDIUM** | `services/api/internal/present/http/controller/ai_controller.go`:54 | Missing Swagger annotations on controller handler Converse in ai_controller.go | Endpoint handler 'Converse' is missing required @Summary or @Router annotations. |
| **MEDIUM** | `services/api/internal/present/http/controller/ai_controller.go`:109 | Missing Swagger annotations on controller handler GetDraft in ai_controller.go | Endpoint handler 'GetDraft' is missing required @Summary or @Router annotations. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:37 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Register | Handler 'Register' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:73 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Login | Handler 'Login' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:109 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for RefreshToken | Handler 'RefreshToken' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:140 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Logout | Handler 'Logout' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:176 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleLogin | Handler 'GoogleLogin' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:193 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleCallback | Handler 'GoogleCallback' uses raw DTO in @Success instead of standard response.Response[T]. |
