---
epic: EPIC-007
session: 2026-10-23_u01-auth-verification
status: in-progress
completed: false
date: 2026-10-23
---

# U01 — Bounded authentication verification

## Scope

Verify merged PR #58 authentication against rebuilt API/web. Preserve retained browser identity for U02–U06. No application code changes; browser logout and full browser cookie/storage inspection remain open.

## Starting evidence

Started on codex/rules-compliance at 2504d69 with pre-existing planning/memory edits and screenshot evidence. Coordinator fetched main and switched to codex/demo-work-units at merged baseline 656a6c4040b31e8d28049338e3e094df5a6db88a, preserving those files. Application sources are that merged baseline; pre-existing local modifications are documentation/evidence only.

Existing demo containers were stopped. Normal startup built successfully but failed because unrelated omni-e2e-api and omni-bridge-kafka-ui containers own host ports 3000 and 8080 respectively. Initial localhost:3000/login M404_PATH_NOT_FOUND belonged to the unrelated app, not Hexta. Neither unrelated service was stopped.

## Outcome and decisions

Rebuilt API/web together and launched hexta-demo on loopback ports 18080/13000 using a private temporary Compose override and API config. Web build embeds http://localhost:18080; API frontend origin is http://localhost:13000. Existing PostgreSQL volume, signing configuration and application data retained. Atlas migration and db-init jobs exited successfully; API/web healthy. No privilege grants, migrations, production deployment, source fixes or secret changes.

Runtime invocation from repository root:

```sh
./scripts/local.sh -f /private/tmp/hexta-u01-runtime/compose.yaml up --build -d --wait api web
```

The override uses Compose `!override` for port lists, replacing rather than adding occupied defaults. Temporary api.yaml is private and must not be published. Exact running image identity:

- API: sha256:9ae6aa6470f5a17fc990cb35267885559b1ad73ed505646651c3414c2268b1b5, 127.0.0.1:18080 -> 8080.
- Web: sha256:9f06e850d1ea08e5bcbf861b44f7be1bd64453638778ea1ce4ed83d80575c1d5, 127.0.0.1:13000 -> 3000.

A dedicated example.test account was registered through the HTTP API for bounded auth checks. It has no tenant/operator grant. Its ephemeral credentials are retained only in a protected temporary fixture for same-session reuse; credential values are absent from this report.

## Verification and evidence

| Check | Result | Evidence and limits |
| --- | --- | --- |
| API/web production image builds and startup | PASS | Both rebuilt from merged sources; web Next.js compile/type/static build passed; alternate-port services healthy. |
| Register/login response minimizes credentials | PASS (HTTP) | Real API responses contain no refresh_token field/value; sanitized Python harness assertions. |
| HttpOnly/SameSite=Lax/path cookie contract | PASS (HTTP and automated) | Real local HTTP response sets HttpOnly, SameSite=Lax, Path=/; Secure=false is expected locally. Cookie values never printed. |
| Secure production cookie configuration | PASS (automated) | Existing controller tests exercise production=true and assert Secure; no production environment claimed. |
| Cookie-only refresh | PASS (HTTP and SDK automated) | Real API refresh with cookie jar/no credential body returns new access; refresh credential absent from JSON. |
| Server-confirmed logout revocation | PASS (HTTP) | Test-only identity logout 200; subsequent authenticated tenant list with old access 401 and refresh 401. |
| Failed logout remains retryable | PASS (SDK automated), NOT RUN (browser) | SDK tests cover server503/network failure with preserved access and successful retry; UI source displays retryable alert. No browser failure injected. |
| Retained owner browser authentication/reload | PASS (browser) | IAB navigated to http://localhost:13000/login, redirected to /tenants and loaded two approved tenant cards. Explicit reload returned the same authenticated chooser. |
| Fresh browser sign-in and browser logout | NOT RUN | Coordinator explicitly preserved retained owner session for later sales/workflow units; no owner password available. Test-only HTTP login/logout is separate evidence. |
| Automatic expiry-driven browser refresh | NOT RUN | No clock/config changes; explicit HTTP refresh and SDK401 tests passed, not claimed as browser expiry verification. |
| Browser cookie metadata / JS storage inspection | NOT RUN | Documented IAB evaluate exposes a read-only DOM scope; document.cookie was unavailable (undefined). No credential values accessed. Source/HTTP/tests support protection but do not replace browser metadata evidence. |

Commands actually passed:

```sh
pnpm --filter @hexta/sdk test
# 1 file, 35 tests passed
cd services/api
go test -race ./internal/present/http/controller ./internal/core/service -run 'Test(PasswordAuth|RefreshReads|CookieAuthentication|Logout|OAuthUses|JWT|TokenExpiration|InvalidAndTampered|AuthService)' -count=1
# both packages passed
python3 /private/tmp/hexta-u01-runtime/check_auth.py
# sanitized real API assertions passed; no credentials printed
```

No reproduced authentication defect; therefore no manufactured code/test change. These checks do not complete U01 browser acceptance or EPIC-007.

## Remaining work and handoff

- Browser state is retained in IAB tab ID `1`, URL http://localhost:13000/tenants. In a new CUA runtime, call `await cua.getState()` first, then `await cua.getTab('1', {browser:'iab'})` (or bind by exact current URL), inspecting returned documentation/state. Subagents must use visible:false when creating IAB tabs; visible:true is unsupported there. Claiming an existing tab does not require new creation.
- Current test owner label: sales-owner-1792380405@example.test. Existing approved tenant IDs: Sales verification 235a481b-a5bd-417f-b805-14c1c8740e11 and Empty workspace verification 2de35c61-8aba-4509-bd15-e97b01d9d1f9. These are test identifiers, not credentials. Do not log out/replace this identity until downstream browser checks finish.
- New tabs on the same localhost browser profile share cookies; creating another tab is not identity isolation. Documented APIs do not expose a new profile/context operation.
- Later complete browser sign-in/logout using the test-only fixture after owner-dependent checks, and inspect cookie attributes/storage through an available supported developer/browser surface. Keep browser-only cases NOT RUN until observed.
- Root coordinator updates shared epic README/index/CHANGELOG serially. This session is local/unmerged and completed:false; no parent completion or lifecycle label change.

## Browser resumption correction — 2026-10-23

After the coordinator reactivated this agent to execute downstream browser cases, its prior JavaScript binding no longer existed. A fresh `cua.getState()` reported `browsers: []`, including no IAB surface. Thus tab `1` was not available for the next agent turn; the earlier handoff instruction describes the previous live snapshot only. No persistent handoff mark had been set on that temporary tab. The native Google Chrome fallback did not return usable UI state before coordinator interruption (the call was aborted after approximately 156 seconds). No login, logout, password, account, application, runtime or session mutation was performed during this retry. U02–U06 browser procedures were not executed by this agent; all previously unrun browser cases remain NOT RUN. The earlier bounded owner reload PASS remains historical evidence from the live U01 turn, not proof of later browser availability or full acceptance. Coordinator explicitly stopped further computer-use attempts.

## Root integration follow-up — 2026-10-23

Root later recovered IAB in the main chat and retained owner reload across the final integrated rebuild. Fresh sign-in/logout/expiry/metadata remain open. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
