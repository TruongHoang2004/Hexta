# Code Review Report: Reconcile Auth Token Storage, Synchronize JWT Claims & Enforce English Language Rule

- **Date**: 2026-09-12
- **Issue**: [#6 - fix(web): reconcile auth token storage, synchronize JWT claims, and enforce English language rule](https://github.com/TruongHoang2004/Hexta/issues/6)
- **Branch**: `task/issue-6-fix-web-reconcile-auth-token-storage-syn`
- **Reviewer**: Antigravity Assistant

---

## 1. Overview

This review audits changes implemented to fix token storage desynchronization across client pages, ensure complete JWT claims (`email`, `sub`, `user_id`) from the backend, eliminate Vietnamese UI strings in accordance with `GEMINI.md`, and verify server-side route protection.

### Files Reviewed:
- [services/api/internal/core/service/auth_service.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go)
- [services/api/internal/core/service/auth_service_test.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service_test.go)
- [apps/web/store/useAuthStore.ts](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/store/useAuthStore.ts)
- [apps/web/app/(dashboard)/tenant/page.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28dashboard%29/tenant/page.tsx)
- [apps/web/app/auth/callback/page.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/auth/callback/page.tsx)
- [apps/web/app/(auth)/login/page.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28auth%29/login/page.tsx)
- [apps/web/app/(auth)/register/page.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/%28auth%29/register/page.tsx)
- [apps/web/components/auth-nav.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/components/auth-nav.tsx)
- [apps/web/app/page.tsx](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/app/page.tsx)
- [apps/web/proxy.ts](file:///Users/truonghoang/Documents/dev/personal/Hexta/apps/web/proxy.ts)

---

## 2. The Good

1. **Storage Adapter Standardization**:
   - Both `apps/web/store/useAuthStore.ts`, `apps/web/app/auth/callback/page.tsx`, and `apps/web/app/(dashboard)/tenant/page.tsx` now uniformly utilize `DefaultBrowserStorage` from `@ubi/sdk`.
   - Resolves the bug where OAuth callback saved tokens into cookies while the dashboard page attempted to read `localStorage`, which bounced users back to `/login`.
2. **JWT Claims Completeness**:
   - Backend `JWTClaims` now explicitly serializes `email` alongside `user_id` and standard RFC 7519 `Subject: userID`.
   - `useAuthStore` robustly reads `payload.sub || payload.user_id` and `payload.email`, ensuring accurate hydration of user state.
3. **App Router Compliance**:
   - Wrapped `useSearchParams()` inside a React `<Suspense>` boundary in `apps/web/app/auth/callback/page.tsx`, preventing runtime client bail-out warnings in Next.js 16.
4. **Next.js 16 Proxy Architecture**:
   - Preserved Next.js 16's `proxy.ts` (the modern successor to `middleware.ts`), avoiding conflict errors while ensuring server-side cookie redirection for `/tenant/*` and auth pages.
5. **Strict Language Rule Compliance**:
   - All Vietnamese strings across `apps/web` (login, register, callback, navbar, and landing page) have been replaced with idiomatic English, strictly adhering to `GEMINI.md` Rule 1.

---

## 3. Critical Issues (Bugs & Security)

None identified. All automated tests (`go test`, `pnpm --filter web build`, `make ts-build`) pass with zero errors.

---

## 4. Suggestions & Improvements

1. **Centralized Configuration for OAuth Redirects**:
   - Currently `window.location.href = "http://localhost:8080/api/v1/auth/google/login"` is used in login and register forms. In a future refactor, use `process.env.NEXT_PUBLIC_API_URL` to facilitate deployment to preview/staging environments.

---

## 5. Conclusion & Verdict

**Verdict**: **APPROVED**  
The token storage desynchronization, JWT claims mismatch, and language standard violations are cleanly resolved.
