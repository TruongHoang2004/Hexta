# Codebase Audit Report: 2026-10-09

- **Inspection Date**: 2026-10-09 13:32:19 UTC
- **Repository Root**: `/Users/truonghoang/Documents/dev/personal/Hexta`
- **Total Findings**: 26
- **Issue Candidates Synthesized**: 1

---

## 1. Executive Summary & Health Metrics

| Category | Count | High | Medium | Low |
|---|---|---|---|---|
| `language` | 20 | 0 | 0 | 20 |
| `swagger` | 6 | 0 | 0 | 6 |

### Severity Breakdown
- **High Severity**: 0
- **Medium Severity**: 0
- **Low Severity**: 26

---

## 2. Issue Generation & Deduplication Actions

- PENDING (Dry Run): Candidate 'chore(i18n): enforce English language rule across source code and comments' with severity `low` and labels ['enhancement', 'priority:low']

---

## 3. Detailed Audit Findings

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
