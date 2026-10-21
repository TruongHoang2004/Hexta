# Current-code syntax and engineering-rule audit

Date: 2026-10-20. Scope: current API/shared Go, web/SDK TypeScript, AI Python, migration scripts and CI/Compose configuration. This is a source audit, not a penetration test or exhaustive behavioral certification.

## Overview and baseline

PR [#57](https://github.com/TruongHoang2004/Hexta/pull/57) is verified MERGED at 2026-10-20T08:54:27Z as `2abd30537e6ab781f56519784c19c41c2bf5e595`; all returned backend/frontend/container/lifecycle checks succeeded. The current branch `codex/atlas-all-services` at `6fc291e` has the same tracked tree as fetched `origin/main`. Existing local memory and screenshot changes were preserved.

Syntax checks pass, but the repository does not fully match AGENTS.md and development-rules.md. The principal findings below predate PR #57. They are current-code debt, not evidence that the merged migration/workflow change introduced these problems. No application source was changed by this audit.

## The good

- Atlas histories validate for API and AI. No Go AutoMigrate or Python startup migrate function remains in the source scan; schema jobs and application startup have separate ownership.
- New workflow code follows repository-owned models, typed DTO validation, shared errors, tenant-qualified operations and transactional event recording. Its changed Go files are gofmt clean.
- Web lint/strict type checks, Go vet/tests and SDK tests pass within the limits below. Runtime API responses inspected use the standard envelope; the Swagger problem is their declared schema.
- Existing core/model placement is an explicitly documented legacy exception in system/02-architecture.md, not a reason to move every model during this audit.

## Critical issues: security

### R1 — P1: missing JWT configuration silently selects predictable signing keys

Evidence: `services/api/internal/core/service/auth_service.go:96–108`; `services/api/config/config.go:54–83`.

The constructor starts with fixed source-code fallback signing secrets and replaces each only when configured. Configuration loading expands environment variables but performs no mandatory-secret validation. A production instance missing JWT variables can therefore start with predictable signing keys. This breaks the secret-management rule and weakens token trust. This is a conditional deployment risk; the audit does not claim the running demo has missing secrets or that any account was compromised.

Remediation: reject absent/invalid signing secrets during bootstrap, before listeners start. Keep any development fallback explicitly restricted to a development-only mode; tests should provide explicit fixture secrets. Add a production-config rejection test covering empty variables. Do not include secret values in errors or logs.

### R2 — P1: password authentication exposes refresh credentials to browser JavaScript

Evidence: `services/api/internal/present/http/controller/auth_controller.go:54–58`, `:90–94`, `:122–125`; `packages/sdk/src/index.ts:64–66`, `:114–125`, `:148–166`.

Register/login/refresh return refresh tokens in JSON and the browser SDK stores them using js-cookie. Those cookies cannot be HttpOnly; the default storage also drops the declared secure/sameSite options. An injected same-origin script can read the long-lived credential. This contradicts development-rules.md's secure HttpOnly, SameSite=Lax refresh-cookie requirement. The OAuth callback already sets an HttpOnly refresh cookie, so the flows are inconsistent.

Remediation: make the browser authentication contract server-owned for refresh-cookie creation/rotation/clearing, with HttpOnly, Secure in production and explicit SameSite policy. Update refresh/logout and SDK credential handling together; browser JS must not read or store the refresh token. Validate cookie behavior, credentials/CORS settings and rotation/logout in integration tests. A non-browser token contract, if needed, should be explicit rather than forcing browser storage to remain readable.

## Suggestions and improvements

### R3 — P2: four Swagger success schemas disagree with runtime envelopes

Evidence: `services/api/internal/present/http/controller/auth_controller.go:35`, `:71`, `:107`; `services/api/internal/present/http/controller/health_controller.go:33`.

Auth advertises bare response DTOs while BaseController.Success returns response.Response[T]. Health advertises map[string]string while returning an envelope containing status and dependency details. Generated clients/documentation can infer the wrong JSON shape. Correct the annotations and introduce a health response DTO, then run make swagger. For example:

```go
// @Success 200 {object} response.Response[dto.LoginResponse]
```

43 object-valued success annotations were scanned; four omit the required envelope. This count does not certify every error/route annotation.

### R4 — P2: legacy AI service crosses persistence, HTTP and error boundaries

Evidence: `services/api/internal/core/service/ai_agent_service.go:33`, `:39`, `:207–221`, `:259–272`, `:304`.

AIAgentService accepts HTTP DTOs/emits SSE DTOs, directly holds/queries GORM and exposes raw error returns. This conflicts with the five-layer architecture and shared *errors.Error rule. The legacy conversation route is not registered in the guarded demo router; this is dormant architectural debt, not a newly demonstrated reachable vulnerability.

Remediation: put queries behind repository interfaces, map storage/provider errors to shared errors, and use domain inputs/events at service boundaries with HTTP DTO/SSE conversion in the controller. Do not expand the legacy AI behavior into approved messaging scope.

### R5 — P2: explicit any bypasses SDK type guarantees

Evidence: `packages/sdk/src/index.ts:239`, `:248`, `:286`, `:375`, `:417` (13 explicit annotations/casts overall).

Strict TypeScript permits explicit any, so the passing compiler does not enforce the repository's avoid-any rule here. Request flags, headers and error handling lose type checking. Use an Axios request-config extension for _retry/_skipAuthRefresh, AxiosHeaders, unknown for caught errors and axios.isAxiosError narrowing. Extend lint coverage to the SDK; current web lint does not guard these files.

### R6 — P3: 31 Go files are not gofmt formatted

135 Go files under services, packages and test were scanned. 31 need gofmt; none are files changed by PR #57. A separate formatting-only change and CI formatting guard would prevent recurring drift. The complete list is below.

- `services/api/cmd/main.go`
- `services/api/utils/jwt_test.go`
- `services/api/common/parse_param.go`
- `services/api/common/log/global.go`
- `services/api/internal/repository/session_repository.go`
- `services/api/internal/repository/inventory_repository.go`
- `services/api/internal/repository/session_wrapper.go`
- `services/api/internal/repository/order_repository.go`
- `services/api/internal/repository/identify_repository.go`
- `services/api/internal/bootstrap/telemetry.go`
- `services/api/internal/bootstrap/client.go`
- `services/api/internal/bootstrap/app.go`
- `services/api/internal/infrastructure/cache/redis.go`
- `services/api/internal/core/service/inventory_service.go`
- `services/api/internal/core/service/auth_service.go`
- `services/api/internal/core/service/auth_service_test.go`
- `services/api/internal/core/service/order_service.go`
- `services/api/internal/present/http/middleware/log.go`
- `services/api/internal/present/http/middleware/recovery.go`
- `services/api/internal/present/http/middleware/auth_test.go`
- `services/api/internal/present/http/middleware/trace.go`
- `services/api/internal/present/http/middleware/auth.go`
- `services/api/internal/present/http/validator/validator.go`
- `services/api/internal/present/http/controller/auth_controller.go`
- `services/api/internal/present/http/controller/health_controller.go`
- `services/api/internal/present/http/controller/base_controller.go`
- `packages/shared/pkg/validator/validator.go`
- `packages/shared/pkg/validator/validator_test.go`
- `packages/shared/pkg/logger/logger.go`
- `packages/shared/pkg/telemetry/middleware.go`
- `test/stress/auth_stress.go`

### Additional documentation/logging debt

AGENTS.md still names frontend/app and frontend/lib/services, while the actual frontend is apps/web; reconcile those paths rather than treating the entire frontend location as a new code violation. `services/api/internal/infrastructure/database/connect.go:13` and `:19` still use standard-library logging instead of the mandated common/log facade. Neither is a syntax error.

## Verification and limits

| Check | Result |
| --- | --- |
| Go vet API/shared | PASS |
| Go test API/shared | PASS; PostgreSQL fixture cases skip without API_TEST_DATABASE_URL |
| Web lint and tsc --noEmit | PASS |
| SDK Vitest | 25 PASS |
| AI pytest | 23 PASS, 4 PostgreSQL-dependent cases SKIP without AI_TEST_DATABASE_URL |
| Python ast.parse | 13 non-generated source/test files parse |
| make migrate-validate | Both histories PASS |
| sh -n scripts/local.sh scripts/migrate.sh | PASS |
| CI/Compose YAML parsing | PASS |
| gofmt | 31 of 135 files need formatting |
| git diff --check | PASS |

Commands: GOCACHE=/private/tmp/hexta-go-build go vet ./services/api/... ./packages/shared/...; the equivalent go test command; pnpm --filter web lint; pnpm --filter web exec tsc --noEmit; pnpm --filter @hexta/sdk test; services/ai/.venv/bin/python -m pytest services/ai/tests -q. The first Python attempt failed 10 gRPC cases solely because sandbox socket binding was denied; the permitted rerun passed those cases. No failures were hidden by the combined SDK command's exit status.

Fresh PostgreSQL fixture tests, production builds, race detection, browser acceptance and external-provider behavior were not rerun in this audit. Merged CI is separate evidence and is not a substitute for the still-NOT-RUN D4 browser checklist. No auth exploit or live configuration change was attempted.

## Handoff

Fix R1/R2 first in a coordinated auth change with targeted tests, then Swagger, legacy boundaries, SDK types and formatting. Findings are OPEN; suggested remedies are TARGET, not IMPLEMENTED. This report and its session record are local/unmerged. No new issue, PR, commit or external review comment was created.

## Implementation follow-up — 2026-10-20, local/unmerged

Owner authorized module subagents. [Implementation session](../epics/EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md) now records R1–R6 repaired locally: explicit JWT keys, HttpOnly refresh contract and failure-aware logout, Swagger envelopes/base-path regression, legacy AI domain/repository boundaries, strict SDK types and clean gofmt. Supplemental DB logging/path and old Atlas test-helper gaps are also repaired. Independent review caught two introduced auth/Swagger integration issues; both were fixed and re-reviewed. Go race80/no skips, docs race, vet, SDK35/types/build, web lint/tsc and Atlas/format/shell/YAML checks pass. These are implementation/verification claims, not merged completion; original findings above preserve audit history. Browser E2E and per-token replay revocation remain outside the verification/scope boundary.
