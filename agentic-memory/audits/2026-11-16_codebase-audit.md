# Codebase Audit Report: 2026-11-16

- **Inspection Date**: 2026-11-16 10:29:44 UTC
- **Repository Root**: `/home/runner/work/Hexta/Hexta`
- **Total Findings**: 35
- **Issue Candidates Synthesized**: 4

---

## 1. Executive Summary & Health Metrics

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| `coverage` | 9 | 0 | 1 | 8 |
| `language` | 21 | 0 | 0 | 21 |
| `security` | 1 | 0 | 1 | 0 |
| `swagger` | 4 | 0 | 0 | 4 |

### Severity Breakdown
- **High Severity**: 0
- **Medium Severity**: 2
- **Low Severity**: 33

---

## 2. Issue Generation & Deduplication Actions

- SKIPPED (Duplicate): 'refactor(tenant): replace temporary mock tenant IDs with dynamic token session context' - Reason: Semantic conflict with issue #7: 'feat(api): implement multi-tenant database models, service domain, and REST API'
- SKIPPED (Duplicate): 'test(coverage): implement unit test suites for critical core services' - Reason: Fingerprint '975b2fe59ab0aec2' matches existing issue
- CREATED: [security: resolve potential hardcoded credential in init-db.sh](https://github.com/TruongHoang2004/Hexta/issues/77) with labels `bug,priority:high`
- SKIPPED (Duplicate): 'chore(i18n): enforce English language rule across source code and comments' - Reason: Semantic conflict with issue #6: 'fix(web): reconcile auth token storage, synchronize JWT claims, and enforce English language rule'

---

## 3. Detailed Audit Findings

### Category: `coverage` (9 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **MEDIUM** | `services/api/internal/core/service/delivery_service.go` | Missing unit test suite for service delivery_service.go | Core business logic in delivery_service.go lacks corresponding test file delivery_service_test.go. |
| **LOW** | `services/api/internal/repository/messaging_repository.go` | Missing repository test coverage for messaging_repository.go | Data persistence wrapper in messaging_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/channel_repository.go` | Missing repository test coverage for channel_repository.go | Data persistence wrapper in channel_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/delivery_repository.go` | Missing repository test coverage for delivery_repository.go | Data persistence wrapper in delivery_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/ingress_repository.go` | Missing repository test coverage for ingress_repository.go | Data persistence wrapper in ingress_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/sales_repository.go` | Missing repository test coverage for sales_repository.go | Data persistence wrapper in sales_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/workflow_models.go` | Missing repository test coverage for workflow_models.go | Data persistence wrapper in workflow_models.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/tenant_onboarding_repository.go` | Missing repository test coverage for tenant_onboarding_repository.go | Data persistence wrapper in tenant_onboarding_repository.go lacks unit or integration test verification. |
| **LOW** | `services/api/internal/repository/workflow_repository.go` | Missing repository test coverage for workflow_repository.go | Data persistence wrapper in workflow_repository.go lacks unit or integration test verification. |

### Category: `language` (21 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/bootstrap/connector_outbound_integration_test.go`:171 | Non-English comment detected in connector_outbound_integration_test.go | GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs. |
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

### Category: `security` (1 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **MEDIUM** | `infrastructure/prod/init-db.sh`:10 | Potential hardcoded password assignment detected in init-db.sh | File contains pattern matching Hardcoded Password Assignment. Hardcoded credentials should never be committed. |

### Category: `swagger` (4 findings)

| Severity | Location | Title | Details |
|---|---|---|---|
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:152 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for Logout | Handler 'Logout' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:192 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleLogin | Handler 'GoogleLogin' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/auth_controller.go`:209 | Swagger @Success in auth_controller.go should use response.Response[T] wrapper for GoogleCallback | Handler 'GoogleCallback' uses raw DTO in @Success instead of standard response.Response[T]. |
| **LOW** | `services/api/internal/present/http/controller/ai_controller.go`:61 | Swagger @Success in ai_controller.go should use response.Response[T] wrapper for Converse | Handler 'Converse' uses raw DTO in @Success instead of standard response.Response[T]. |
