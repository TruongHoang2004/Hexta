# Development Changelog & Verification Guide: Reconcile Auth Token Storage, Synchronize JWT Claims & Enforce English Language Rule (Issue #6)

- **Date**: 2026-09-12
- **Issue**: [#6 - fix(web): reconcile auth token storage, synchronize JWT claims, and enforce English language rule](https://github.com/TruongHoang2004/Hexta/issues/6)
- **Branch**: `task/issue-6-fix-web-reconcile-auth-token-storage-syn`

---

## 1. Change Summary

This update resolves authentication flow inconsistencies between `apps/web` and `services/api`:
1. **Token Storage Reconciled**: Standardized token storage on Cookies via `@ubi/sdk`'s `DefaultBrowserStorage` across the entire web frontend, eliminating the bug where users logging in via Google OAuth were bounced from `/tenant` back to `/login` because `/tenant` looked in `localStorage`.
2. **JWT Claims Synchronized**: Added `email` to backend `JWTClaims` alongside `user_id` and RFC 7519 `Subject: userID`. Updated `useAuthStore` to extract `id` and `email` reliably.
3. **App Router `<Suspense>` Boundary**: Wrapped `useSearchParams()` in `apps/web/app/auth/callback/page.tsx` within React `<Suspense>` to prevent Next.js client de-optimization.
4. **Server-Side Route Protection**: Leveraged Next.js 16's `proxy.ts` to guard `/tenant/*` against unauthenticated requests and redirect logged-in users away from `/login` and `/register`.
5. **English Language Standard Enforced**: Translated all hardcoded Vietnamese strings across `apps/web` to clean, idiomatic English in full compliance with `GEMINI.md`.

---

## 2. Impacted Components & Files

| File | Action | Description |
| :--- | :---: | :--- |
| [`services/api/internal/core/service/auth_service.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go) | `[MODIFY]` | Added `Email` to `JWTClaims`, updated `generateToken` signature and call sites. |
| [`services/api/internal/core/service/auth_service_test.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service_test.go) | `[NEW]` | Added unit test verifying `JWTClaims` has `sub`, `user_id`, and `email` populated. |
| [`apps/web/store/useAuthStore.ts`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/store/useAuthStore.ts) | `[MODIFY]` | Standardized on `DefaultBrowserStorage`, extracted both `sub`/`user_id` and `email`. |
| [`apps/web/app/(dashboard)/tenant/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28dashboard%29/tenant/page.tsx) | `[MODIFY]` | Replaced `localStorage.getItem` with `DefaultBrowserStorage.get("auth_token")`. |
| [`apps/web/app/auth/callback/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/auth/callback/page.tsx) | `[MODIFY]` | Wrapped search params in `<Suspense>`, used `DefaultBrowserStorage`, translated to English. |
| [`apps/web/app/(auth)/login/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28auth%29/login/page.tsx) | `[MODIFY]` | Translated all Vietnamese UI strings to English. |
| [`apps/web/app/(auth)/register/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28auth%29/register/page.tsx) | `[MODIFY]` | Translated all Vietnamese UI strings to English. |
| [`apps/web/components/auth-nav.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/components/auth-nav.tsx) | `[MODIFY]` | Translated all Vietnamese UI strings to English. |
| [`apps/web/app/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/page.tsx) | `[MODIFY]` | Translated landing page hero, features, navigation, and footer to English. |

---

## 3. Key Technical Decisions

1. **Cookie-Based Standard Storage (`DefaultBrowserStorage`)**:
   - Storing tokens in cookies via `DefaultBrowserStorage` allows both client-side JavaScript (Zustand, SDK) and server-side Next.js edge runtime (`proxy.ts`) to read the authentication token seamlessly without flash of unauthenticated content.
2. **Next.js 16 `proxy.ts` Architecture**:
   - In Next.js 16 (Turbopack), Next.js transitioned `middleware.ts` to `proxy.ts`. Attempting to define both simultaneously results in a build failure. We retain and enforce `proxy.ts` for clean edge redirects.
3. **Dual JWT ID Extraction Fallback**:
   - `payload.sub || payload.user_id || ""` provides full compatibility with standard RFC 7519 `sub` claims while retaining backward compatibility with existing tokens containing `user_id`.

---

## 4. Step-by-Step Walkthrough

### 4.1 Backend Claims Enrichment
In [`services/api/internal/core/service/auth_service.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go):
```go
type JWTClaims struct {
	SessionID int64  `json:"session_id"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateToken(sessionID int64, userID string, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		SessionID: sessionID,
		UserID:    userID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
```

### 4.2 Client Auth Store & Page Check
In [`apps/web/store/useAuthStore.ts`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/store/useAuthStore.ts):
```typescript
const storage = new DefaultBrowserStorage();
...
const token = storage.get("auth_token");
if (token) {
  const payload = JSON.parse(atob(token.split(".")[1]));
  const id = payload.sub || payload.user_id || "";
  const email = payload.email || "";
  set({ user: { id, email }, isAuthenticated: true });
}
```

In [`apps/web/app/(dashboard)/tenant/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28dashboard%29/tenant/page.tsx):
```typescript
const storage = new DefaultBrowserStorage();
...
if (!storage.get("auth_token")) {
  router.push("/login");
  return;
}
```

---

## 5. Verification & Testing Guide

### 5.1 Automated Backend Tests
Run Go unit tests:
```bash
go test -v ./services/api/internal/core/service/...
```
Expected output:
```
=== RUN   TestAuthService_JWTClaims_EmailAndSub
--- PASS: TestAuthService_JWTClaims_EmailAndSub (0.00s)
PASS
```

### 5.2 Frontend Build & TypeScript Checks
Run Next.js build:
```bash
pnpm --filter web build
```
Expected output:
```
▲ Next.js 16.3.0 (Turbopack)
✓ Compiled successfully
  Running TypeScript ...
  Finished TypeScript ...
✓ Generating static pages (8/8)
Route (app)
┌ ○ /
├ ○ /_not-found
├ ○ /auth/callback
├ ○ /login
├ ○ /register
└ ○ /tenant
ƒ Proxy (Middleware)
```

Run all TypeScript builds:
```bash
make ts-build
```
Verify that all 6 workspace packages compile cleanly with zero errors.
