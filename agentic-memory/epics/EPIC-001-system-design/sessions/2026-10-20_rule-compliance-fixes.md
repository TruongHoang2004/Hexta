---
epic: EPIC-001
session: 2026-10-20_rule-compliance-fixes
status: in-progress
completed: false
date: 2026-10-20
---

# Delegated engineering-rule compliance fixes

## Scope

Owner authorized a subagent for each module identified in the [audit](../../../reviews/2026-10-20_code-rules-audit_review.md). Primary EPIC-001, cross-domain auth/API/SDK and dormant legacy AI. Preserve unrelated local memory/screenshots. No commit, PR or deployment requested.

## Starting evidence

Dedicated branch codex/rules-compliance created from fetched main 2abd305 (merged PR #57). Existing dirty documents/evidence retained. Audit identifies R1–R6 plus stale frontend paths and database logging debt.

## Plan and module ownership

- auth_backend: JWT configuration validation, server-owned refresh cookie lifecycle, typed auth/health Swagger and tests.
- sdk_frontend: credentialed cookie refresh contract, strict Axios/error types, SDK enforcement/tests and browser integration.
- legacy_ai: repository/domain abstractions, shared business errors and transport mapping; no route registration or messaging expansion.
- format_logging_ci: remaining formatting, safe shared DB logging, backend formatting guard and canonical frontend paths.
- integration_review: independent read-only cross-module review; found logout failure/race and Swagger prefix defects, both fixed and re-reviewed.
- Root: integration review, combined tests, architecture/session/changelog evidence and preservation of unrelated work.

## Contract decisions

Refresh credentials belong to HttpOnly SameSite=Lax cookies, Secure in production. Browser register/login/refresh responses contain access/session/user data without refresh-token exposure. Refresh reads the server cookie; SDK uses credentialed requests and does not persist refresh credentials. Missing signing secrets must fail before serving requests; tests supply explicit fixture keys. No DB/schema changes or automatic migration.

## Outcome and decisions

R1–R6 implemented locally. No schema/history/provider/messaging expansion. Backend validates explicitly supplied, distinct signing keys before serving. Browser auth JSON carries access data only; HttpOnly SameSite=Lax refresh cookie is Secure in production and browser origin is checked. CORS allows credentialed configured origins. Logout failure preserves retryable state and shows UI errors; success blocks refresh and drains pending writes. SDK strict Axios/error types replace explicit any, with AST-backed type enforcement and regression CI. Swagger envelopes/typed health and canonical basePath / match registered API-qualified routes. Legacy AI core uses domain inputs/events and read-only repository/generation ports, with sanitized shared errors and unchanged unregistered routes.

All formatting findings fixed, DB logging uses initialized common/log, root frontend paths corrected. Shared old repository test helper now uses Atlas disposable databases instead of a hardcoded server; all five previously skipped repository tests ran. No AutoMigrate introduced. Existing documentation/screenshots remain preserved.

## Verification and evidence

- Full Go API/shared `go test -count=1 -race -v` with API_TEST_DATABASE_URL pointing only to a disposable local PostgreSQL17 container: 80 top-level tests PASS, 11 test-bearing packages, zero SKIP/FAIL. New AI catalog regression and sales/workflow concurrency cases ran. Log: /private/tmp/hexta-rules-go-tests-final.log (ephemeral, not a committed artifact).
- Generated Swagger contract race test added afterward: PASS via `go test -race ./services/api/docs`; resolves documented health/auth URLs against actual registered routes and checks envelopes/bodyless refresh.
- `go vet ./services/api/... ./packages/shared/...`: PASS after integration.
- SDK: 35 tests PASS, strict/no-explicit-any guard PASS (negative guard probe also rejected forbidden input), ESM/CJS/declaration builds PASS. Web lint and tsc --noEmit PASS. Built tracked SDK artifacts regenerated.
- All 145 current Go files (including untracked new files) gofmt clean; tracked guard, shell syntax, workflow/Compose YAML parsing, both Atlas histories and git diff --check PASS.
- Independent read-only reviewer found logout failure/race semantics and doubled Swagger prefix during integration; implementation agents fixed both and final review found no additional actionable introduced bug. Initial full Go attempt overlapped active formatting and hit import-build errors; settled fresh run above passed.
- Temporary PostgreSQL container stopped/removed after tests; demo data untouched. No browser E2E, live auth/config change, production image rollout or new remote CI run. Python code unchanged; prior Python tests are separate evidence.

## Compatibility and limits

API/SDK cookie auth is a coordinated contract change: old clients expecting JSON refresh_token or supplying refresh JSON must upgrade. Frontend and API must be same-site for SameSite=Lax cookies; credentialed origin must be configured explicitly. JWT keys must be distinct and at least 32 bytes in all environments, including tests. Refresh JWTs are reissued with unique IDs but existing session-based validity remains; individual-token replay revocation was not added. Pending network refresh drain inherits existing request timeout configuration; no general timeout policy change. D4 browser acceptance and parent epic acceptance remain open.

## Remaining work and handoff

Implementation, targeted tests, combined database/race checks and final review are complete locally. User can review/commit this branch next; no commit/PR/deployment was requested or performed. Session status remains in-progress, completed false until delivered/merged. Parent acceptance unchanged.
