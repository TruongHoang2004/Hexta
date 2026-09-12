# Code Review: Google OAuth2 Authentication Flow

- **Date**: 2026-09-12
- **Scope**: Commit `6e2571b05f0bf2cd2f7d784e893218425775223e` ("feat: implement Google OAuth2 authentication flow for login and registration")
- **Reviewer**: AI Pair Programmer (Antigravity)
- **Status**: Completed

---

## 1. Overview
Audit of the end-to-end Google OAuth2 authentication flow across the frontend (Next.js) and backend (Go / Gin):
- **Backend**:
  - `services/api/internal/core/service/auth_service.go`
  - `services/api/internal/present/http/controller/auth_controller.go`
  - `services/api/internal/present/http/router/router.go`
  - `services/api/config/config.go`, `config.yaml`
- **Frontend**:
  - `apps/web/app/(auth)/login/page.tsx`
  - `apps/web/app/(auth)/register/page.tsx`
  - `apps/web/app/auth/callback/page.tsx`

---

## 2. The Good
- **Layered Architecture Compliance**: Clean demarcation between Controller (`AuthController`), Service (`AuthService`), and Repository layers.
- **Standard Library Adoption**: Leverages official `golang.org/x/oauth2` and `golang.org/x/oauth2/google` packages rather than ad-hoc HTTP calls.
- **Centralized Configuration**: OAuth credentials (`client_id`, `client_secret`, `redirect_url`) are organized in `config.yaml` and loaded via Viper.
- **Proper Resource Clean-up**: HTTP response body for userinfo retrieval is safely closed with `defer resp.Body.Close()`.

---

## 3. Issues & Recommended Action Items

### 3.1 [Security Vulnerability] Hardcoded OAuth `state` (Missing CSRF Protection)
- **Issue**: In `auth_service.go`, `conf.AuthCodeURL("state")` hardcodes the state string `"state"`, and `GoogleCallback` in `auth_controller.go` fails to validate the `state` parameter entirely. This leaves users exposed to OAuth CSRF attacks, potentially binding a victim's session to an attacker's identity.
- **Fix**: Generate a cryptographically secure random token via `crypto/rand`, store it temporarily in an `HttpOnly` cookie or Redis cache (5-10 min TTL), and verify that `c.Query("state")` matches upon callback.

### 3.2 [Security & Bug] Token Exposure in Query String & Discarded Refresh Token
- **Issue**:
  1. `auth_controller.go` redirects back to the frontend with the token in the URL: `redirectURL := "http://localhost:3000/auth/callback?token=" + tokens.AccessToken`. This leaks the access token via browser history, proxy/server access logs, and `Referer` headers.
  2. `tokens.RefreshToken` is discarded in `GoogleCallback`. In `callback/page.tsx`, the frontend sets cookie expiration to 7 days for an access token that expires in 15 minutes on the backend. After 15 minutes, API requests fail with 401 Unauthorized and silent token refreshing is impossible.
- **Fix**: Write the Refresh Token directly into a secure `HttpOnly`, `SameSite=Lax` cookie from the backend callback controller. Transmit the Access Token via URL hash fragment (`#token=...`) or a one-time code exchange.

### 3.3 [Maintainability] Hardcoded Host URLs
- **Issue**:
  - Backend hardcodes `http://localhost:3000`.
  - Frontend hardcodes `http://localhost:8080/api/v1/auth/google/login`.
  This breaks staging, preview environments, Docker containers, and production deployments.
- **Fix**: Store the frontend URL in backend `config.yaml` (`config.AppConfig.FrontendURL`), and use `process.env.NEXT_PUBLIC_API_URL` on the frontend.

### 3.4 [Logic & Integrity] Account Linking & Plaintext Dummy Password
- **Issue**:
  - `identityRepo.GetCredentialByIdentifier(..., model.ProviderGoogle)` returning `nil` unconditionally generates a brand new `UserID` (`uuid.New()`), even when the user previously registered with the identical email using email/password (`model.ProviderLocal`). This creates duplicate split accounts for the same person.
  - Plaintext password `"oauth2-dummy"` is stored in the database.
- **Fix**: Check whether an account with that email already exists under any provider. If found, link `ProviderGoogle` to the existing `UserID`. Allow `Password` to be nullable for OAuth identities rather than saving dummy plaintext strings.

### 3.5 [Frontend] Missing `<Suspense>` Boundary for `useSearchParams()`
- **Issue**: In the Next.js App Router, using `useSearchParams()` in `apps/web/app/auth/callback/page.tsx` without wrapping the component in a React `<Suspense>` boundary causes Next.js build warnings and de-optimizes page rendering.
- **Fix**: Extract the callback logic into an inner component and wrap it with `<Suspense fallback={<Loading />}>`.
