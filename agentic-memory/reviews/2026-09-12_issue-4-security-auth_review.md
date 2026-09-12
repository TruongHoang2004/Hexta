# Code Review: Authentication Security Hardening (Issue #4)

- **Date**: 2026-09-12
- **Issue**: [#4: security(auth): patch hardcoded JWT secret, OAuth2 CSRF vulnerability, and enforce session ownership](https://github.com/TruongHoang2004/Hexta/issues/4)
- **Reviewer**: AI Pair Programmer (Antigravity)
- **Status**: Passed / Approved

---

## 1. Overview
This review covers the security hardening and architectural audit for the authentication domain in `services/api` and web callback handling in `apps/web`:
- **Backend Components**:
  - `services/api/config/config.go`, `services/api/config/config.yaml`: Injected JWT secrets, expirations, and frontend host URLs.
  - `services/api/internal/core/service/auth_service.go`: Injected token signing/expiration configurations, dual-key isolation (`access` vs `refresh`), cryptographically secure OAuth2 state generation/verification, and ownership checks on logout.
  - `services/api/internal/present/http/controller/auth_controller.go`: Updated logout endpoint to require authentication, enforce session ownership, and deliver tokens via secure cookies instead of URL query strings.
  - `services/api/internal/present/http/middleware/auth.go`: Created `AuthMiddleware` to parse Bearer tokens, validate session status, and inject `types.AuthInfo` into Gin context (`common.SetAuthInfo`).
  - `services/api/internal/present/http/router/router.go`: Applied `AuthMiddleware` to all private routes (`CreatePrivateRouterGroup`) and secured `/api/v1/auth/logout`.
  - `services/api/internal/bootstrap/`: Wired configuration and middleware into Uber Fx container.
  - `services/api/utils/jwt.go`: Corrected expiration duration unit to seconds.
  - `services/api/docs/`: Regenerated Swagger OpenAPI specifications.
- **Frontend Components**:
  - `apps/web/app/auth/callback/page.tsx`: Added support for extracting `auth_token` from secure cookies and wrapped page logic with React `<Suspense>`.

---

## 2. The Good
- **Strict 5-Layer Go Architecture**:
  - Controller layer validates HTTP parameters and maps errors to status codes.
  - Service layer encapsulates core business rules (token generation, expiration, state verification, session ownership).
  - Persistence queries and caching are kept within repository wrappers.
- **Defense-in-Depth Security**:
  - **Cryptographic Randomness**: OAuth state utilizes 32 bytes from `crypto/rand` encoded via `base64.RawURLEncoding`.
  - **Single-Use CSRF Tokens**: OAuth state is stored with a 5-minute TTL and atomically consumed/deleted upon first validation (`LoadAndDelete` / Redis `DEL`), preventing replay attacks.
  - **Key & Token Type Separation**: Access and refresh tokens utilize distinct secret keys and distinct `token_type` claim payloads. Refresh tokens cannot be presented as access tokens.
  - **Session Revocation Validation**: `ValidateAccessToken` validates active session status against the repository/cache, ensuring immediate revocation across distributed nodes.
  - **Protected Logout**: Session revocation verifies caller ownership (`session.UserID == caller.UserID`), returning `403 Forbidden` if an attacker attempts to terminate another user's session.
  - **No Token Leakage in URLs**: `GoogleCallback` sets access tokens in standard cookies and refresh tokens in `HttpOnly`, `SameSite=Lax` cookies, eliminating exposure in browser history and HTTP `Referer` headers.
- **Comprehensive Unit Testing**:
  - 100% pass rate across service tests, middleware tests, and utility tests.
  - Verified token signing, expiration timing, invalid signature rejection, CSRF state replay rejection, and session ownership enforcement.

---

## 3. Security & Bug Resolutions Verified

### 3.1 Hardcoded JWT Secrets (Resolved)
- **Before**: `jwtSecret: []byte("super-secret-key-replace-me-later")` hardcoded in `auth_service.go`.
- **After**: `accessTokenSecret`, `accessTokenExpire`, `refreshTokenSecret`, and `refreshTokenExpire` are dynamically loaded from `config.AppConfig.JWT` with safe development fallbacks.

### 3.2 OAuth2 CSRF Vulnerability (Resolved)
- **Before**: Static state `"state"` in `GoogleLogin` and no state validation in `GoogleCallback`.
- **After**: Unique random state generated with `crypto/rand`, stored in Redis with 5-minute TTL, validated and immediately deleted on callback.

### 3.3 Insecure Session Revocation / Arbitrary Logout (Resolved)
- **Before**: `POST /api/v1/auth/logout` accepted arbitrary `session_id` in request body without authentication.
- **After**: Endpoint is protected by `AuthMiddleware`. Default behavior revokes caller's current session (`authInfo.SessionID`). If explicit `session_id` is supplied, caller ownership is validated (`session.UserID == authInfo.UserID`).

### 3.4 Missing Private Route Middleware (Resolved)
- **Before**: `CreatePrivateRouterGroup` returned an unauthenticated Gin group.
- **After**: `CreatePrivateRouterGroup` requires `*middleware.AuthMiddleware` injected via Fx, applying `private.Use(authMiddleware.Authenticate())` across all private routes.

### 3.5 Token Exposure in Query Parameters (Resolved)
- **Before**: `GoogleCallback` redirected to `http://localhost:3000/auth/callback?token=<access_token>`.
- **After**: Redirects to `${frontend_url}/auth/callback` without tokens in query parameters. Tokens are delivered via secure cookies (`auth_token` and `HttpOnly` `refresh_token`).

---

## 4. Suggestions & Future Improvements
- **Cookie Domain Hardening**: When deploying to production across custom subdomains (e.g. `api.commercehub.com` and `app.commercehub.com`), consider configuring an explicit top-level cookie domain (e.g. `.commercehub.com`) in `config.AppConfig`.
- **Key Rotation**: Consider introducing key identifiers (`kid`) in JWT headers if multi-key rotation strategies are adopted in the future.

---

## 5. Verification Summary
- `go build ./services/api/...`: Pass (Exit code 0)
- `go test -count=1 ./services/api/...`: Pass (Exit code 0, 100% tests green)
- `make -C services/api swagger`: Pass (OpenAPI spec regenerated)
- Security audit: All items in Issue #4 resolved and validated.
