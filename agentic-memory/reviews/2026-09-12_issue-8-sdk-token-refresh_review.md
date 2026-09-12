# Code Review: Silent Token Refresh Interceptor and SDK Rebranding

- **Issue**: [#8](https://github.com/TruongHoang2004/Hexta/issues/8)
- **Date**: 2026-09-12
- **Branch**: `task/issue-8-feat-sdk-add-silent-token-refresh-interc`
- **Scope**: `packages/sdk`, `apps/web`, build & workspace configurations

---

## 1. Overview
This review covers the modernization and silent token renewal mechanism in `@hexta/sdk` (formerly `@ubi/sdk` / `EcommerceHubSDK`):
- Modernized package name to `@hexta/sdk` and primary class to `HextaSDK`, keeping `EcommerceHubSDK` as an export alias for backwards compatibility.
- Upgraded `StorageAdapter` to support optional expiration and path options, plus added in-memory adapter `MemoryStorage` for headless/test environments.
- Updated `IdentityClient` (`login`, `register`, `refreshToken`, `logout`, `clearTokens`) to persist and manage both `auth_token` (access token) and `refresh_token`.
- Built an Axios response interceptor for automatic `401 Unauthorized` handling that prevents infinite loops, avoids refreshing auth endpoints, queues concurrent requests during active refresh, replays queued and original requests upon refresh success, and purges tokens while invoking `onAuthExpired` upon refresh failure.
- Added comprehensive unit tests with `vitest` and `axios-mock-adapter`.
- Replaced references in `apps/web` and wrapped `useSearchParams` in `<Suspense>` to ensure clean builds.

---

## 2. The Good
- **Concurrency Safety**: Maintains an `isRefreshing` boolean and promise queue `failedQueue`. When multiple requests encounter 401s concurrently, only a single refresh call to `/api/v1/auth/refresh` is dispatched. All waiting requests replay smoothly with the new token.
- **Infinite Loop Protection**:
  - Sets `_retry = true` on original and queued requests.
  - Passes `_skipAuthRefresh = true` on refresh calls so the refresh endpoint never triggers its own refresh handler.
  - Explicitly bypasses auth endpoints (`/api/v1/auth/login`, `/api/v1/auth/register`, `/api/v1/auth/refresh`).
- **Backward Compatibility**: `EcommerceHubSDK` is exported as an alias to `HextaSDK`, and default storage keys remain `auth_token` and `refresh_token`, maintaining harmony with `apps/web/proxy.ts` middleware.
- **Robust Header Handling**: Safely manipulates headers regardless of whether Axios 1.x `AxiosHeaders` class or plain objects are used.
- **High Test Coverage**: 14 unit tests validating individual methods, storage adapters, concurrent 401 recovery, and failure modes using `MockAdapter`.

---

## 3. Critical Issues (Bugs & Security)
- **None identified**:
  - Tokens are stored in standard cookie/storage adapters without hardcoded secrets.
  - When token renewal fails or credentials expire, tokens are purged immediately from storage to prevent stale authentication states.
  - Sensitive tokens are stripped from refresh request headers to avoid sending expired Authorization tokens.

---

## 4. Suggestions & Improvements
- **Token Rotation Handling**: Backend `/api/v1/auth/refresh` returns both a new access token and a refreshed rotation token. The implementation correctly writes both to storage if returned.
- **SSR Compatibility**: `DefaultBrowserStorage` guards against `typeof window !== "undefined"`, preventing server-side rendering crashes in Next.js Server Components.
