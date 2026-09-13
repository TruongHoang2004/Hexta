# Implementation Plan: Security Hardening for Authentication (Issue #4)

- **Date**: 2026-09-12
- **Issue**: [#4: security(auth): patch hardcoded JWT secret, OAuth2 CSRF vulnerability, and enforce session ownership](https://github.com/TruongHoang2004/Hexta/issues/4)
- **Status**: In-Progress
- **Author**: Antigravity Task Runner Agent

---

## 1. Overview & Goal
The objective is to eliminate multiple critical security vulnerabilities and architectural gaps in `services/api` authentication:
1. **Hardcoded JWT Secret**: Replace hardcoded `jwtSecret` with injected configuration values (`jwt.access_token_secret`, `jwt.access_token_expire`, `jwt.refresh_token_secret`, `jwt.refresh_token_expire`) from `config.AppConfig`.
2. **OAuth2 CSRF Vulnerability**: Replace static OAuth state (`"state"`) with cryptographically secure random state generated via `crypto/rand`, stored in Redis with a 5-minute TTL, and strictly validated upon Google OAuth callback.
3. **Insecure Logout & Arbitrary Session Revocation**: Enforce session ownership so users can only revoke their own sessions. If no session ID is passed, default to revoking the current authenticated session.
4. **Missing Private Route Auth Middleware**: Implement `AuthMiddleware` in `internal/present/http/middleware/auth.go` to validate Bearer tokens, verify active session status, inject `types.AuthInfo` into Gin context (`common.SetAuthInfo`), and protect private routes and `/api/v1/auth/logout`.
5. **URL Query Token Exposure**: Stop leaking access tokens into browser history and HTTP referrers via query parameters in `GoogleCallback`. Instead, deliver tokens via secure `HttpOnly` and standard cookies, redirecting cleanly to `/auth/callback`.

---

## 2. Current State Analysis
- **`services/api/internal/core/service/auth_service.go`**:
  - Contains `jwtSecret: []byte("super-secret-key-replace-me-later")`. Access and refresh tokens share the exact same key and hardcoded TTLs (`15*time.Minute`, `7*24*time.Hour`).
  - `GoogleLogin` invokes `conf.AuthCodeURL("state")` using hardcoded string `"state"`.
  - `GoogleCallback` does not validate the state parameter, allowing CSRF login attacks.
  - `Logout(ctx, sessionID)` directly revokes `sessionID` without checking if the calling user owns that session.
- **`services/api/internal/present/http/controller/auth_controller.go`**:
  - `Logout` accepts `req.SessionID` from untrusted request body without requiring authentication.
  - `GoogleCallback` redirects to `redirectURL := "http://localhost:3000/auth/callback?token=" + tokens.AccessToken`, leaking secrets into URLs and discarding `tokens.RefreshToken`.
- **`services/api/internal/present/http/router/router.go`**:
  - `CreatePrivateRouterGroup` creates a bare Gin group with no authentication middleware applied.
  - `POST /auth/logout` is bound to `params.Public` instead of `params.Private`.
- **`apps/web/app/auth/callback/page.tsx`**:
  - Currently expects `token` query param and lacks fallback to cookie storage or `<Suspense>` wrapper.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Config & Dependency Injection (`config`, `bootstrap`)
- Update `services/api/config/config.go` & `config.yaml` to include `frontend_url` under `Server`.
- Expose `*config.Config` in `internal/bootstrap/service.go` via `fx.Provide(func() *config.Config { return config.AppConfig })`.
- Update `service.NewAuthService` signature and wiring to accept `redisClient *cache.RedisClient` and `cfg *config.Config`.

### Step 3.2: Secure JWT Token Handling in `AuthService`
- Refactor `AuthService` to hold:
  - `accessTokenSecret []byte`
  - `accessTokenExpire time.Duration`
  - `refreshTokenSecret []byte`
  - `refreshTokenExpire time.Duration`
- Add dedicated `generateAccessToken`, `generateRefreshToken`, `parseAccessToken`, `parseRefreshToken`.
- Validate token type (`"access"` vs `"refresh"`) in `JWTClaims` so access tokens cannot be reused as refresh tokens and vice versa.
- Provide `ValidateAccessToken(ctx context.Context, tokenStr string) (*types.AuthInfo, *errors.Error)` that verifies signature, expiration, and active session status in the session repository.

### Step 3.3: Cryptographically Secure OAuth2 State Handling
- In `AuthService`, implement `generateAndSaveOAuthState(ctx)` using 32 bytes from `crypto/rand` encoded with `base64.RawURLEncoding`.
- Store state in Redis under `oauth_state:<state>` with 5-minute TTL (with an in-memory `sync.Map` fallback for unit tests).
- In `AuthService.ValidateAndConsumeOAuthState(ctx, state)`, atomically verify and delete the state to ensure single-use CSRF protection.
- In `AuthController.GoogleLogin`, invoke secure state generation and set temporary cookie.
- In `AuthController.GoogleCallback`, validate `c.Query("state")`, reject missing or invalid state, and consume state.
- Implement account linking in `GoogleCallback` if an existing local identity matches the OAuth email, avoiding duplicate split accounts and dummy passwords.

### Step 3.4: Implement `AuthMiddleware` & Protect Private Routes
- Create `services/api/internal/present/http/middleware/auth.go`:
  - Extract Bearer token from `Authorization` header (with cookie fallback).
  - Call `authService.ValidateAccessToken`.
  - Set `types.AuthInfo` on Gin context and `common.SetAuthInfo` on request context.
- Provide `middleware.NewAuthMiddleware` in `internal/bootstrap/middleware.go`.
- Update `internal/present/http/router/router.go`:
  - Inject `*middleware.AuthMiddleware` into `CreatePrivateRouterGroup(r *gin.Engine, authMiddleware *middleware.AuthMiddleware)`.
  - Apply `private.Use(authMiddleware.Authenticate())`.
  - Move `POST /auth/logout` into `params.Private`.

### Step 3.5: Enforce Session Ownership on Logout
- Update `AuthService.Logout(ctx context.Context, userID string, sessionID int64) *errors.Error`:
  - Fetch session from repository. Return `404 Not Found` if missing.
  - Assert `session.UserID == userID`. Return `403 Forbidden` if mismatched.
  - Revoke session via repository.
- Update `AuthController.Logout`:
  - Require authenticated user (`common.GetAuthInfo`).
  - If `req.SessionID == 0`, default to `authInfo.SessionID`.
  - Enforce ownership and clear auth cookies.

### Step 3.6: Secure OAuth Token Delivery
- In `AuthController.GoogleCallback`:
  - Set `auth_token` cookie for access token.
  - Set `refresh_token` in an `HttpOnly`, `SameSite=Lax` cookie.
  - Redirect to `${frontend_url}/auth/callback` without `?token=` query parameters.
- Update `apps/web/app/auth/callback/page.tsx`:
  - Check `Cookies.get("auth_token")` alongside query parameters for seamless authentication.
  - Wrap component in React `<Suspense>` to adhere to Next.js App Router requirements.

### Step 3.7: Testing & Documentation
- Write comprehensive unit tests in `services/api/internal/core/service/auth_service_test.go`:
  - Access token signing with configured secret and expiration.
  - Refresh token signing with configured secret and expiration.
  - Inability to use access token with refresh token secret or vice-versa.
  - Token expiration rejection.
  - OAuth state generation, validation, one-time consumption, and rejection of reused/tampered state.
  - Session ownership verification in `Logout`.
- Write unit tests in `services/api/internal/present/http/middleware/auth_test.go` verifying header extraction, 401 handling, context injection.
- Run `make swagger` to update OpenAPI specifications.
- Verify full test suite and build passing.

---

## 4. Risk Assessment & Edge Cases
| Risk / Edge Case | Impact | Mitigation Strategy |
|---|---|---|
| Redis unavailable during OAuth login / callback | User cannot log in via OAuth | Implement robust fallback to in-memory state tracking (`sync.Map`) with expiration when Redis client is nil or down. |
| Clock skew during JWT validation | Valid token rejected immediately upon creation | Standard leeway in `jwt.RegisteredClaims` and ensure accurate server time. |
| Replay attacks using OAuth `state` | Attacker attempts to reuse a intercepted state | State is atomically deleted immediately upon first validation (`LoadAndDelete` / Redis `DEL`). |
| Missing `Authorization` header on Logout | Unauthenticated caller | Blocked at routing level by `AuthMiddleware` returning 401 before hitting controller. |
| Revoking someone else's session | Malicious denial of service on other users | Service layer checks `session.UserID == caller.UserID`, returning 403 Forbidden if mismatched. |

---

## 5. Definition of Done (DoD)
- [ ] No hardcoded JWT secrets remain in `services/api`.
- [ ] OAuth2 state is cryptographically randomized, stored with 5-minute TTL, and verified on callback.
- [ ] Token is no longer exposed in `GoogleCallback` URL query parameters; cookies are set appropriately.
- [ ] `AuthMiddleware` is applied to all private routes in `CreatePrivateRouterGroup`.
- [ ] `/api/v1/auth/logout` requires authentication and enforces session ownership.
- [ ] Comprehensive unit tests pass with 100% green status (`go test ./services/api/...`).
- [ ] Code compiles cleanly (`go build ./services/api/...`).
- [ ] Swagger documentation regenerated (`make swagger`).
- [ ] Review report (`agentic-memory/reviews/`) and changelog (`agentic-memory/changelogs/`) generated and committed.
