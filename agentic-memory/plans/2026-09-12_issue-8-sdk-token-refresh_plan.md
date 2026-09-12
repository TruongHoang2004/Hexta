# Implementation Plan: Silent Token Refresh Interceptor and SDK Package Rebranding

- **Issue**: [#8](https://github.com/TruongHoang2004/Hexta/issues/8)
- **Date**: 2026-09-12
- **Branch**: `task/issue-8-feat-sdk-add-silent-token-refresh-interc`
- **Scope**: TypeScript SDK (`packages/sdk`) and web consumer (`apps/web`)

---

## 1. Overview & Goal
The Hexta SDK currently manages only a short-lived `access_token` stored under `auth_token` with an arbitrary 7-day cookie expiry, while backend access tokens expire in 15 minutes. When the access token expires, client API calls fail with `401 Unauthorized` without any automatic recovery or silent renewal mechanism. Furthermore, the SDK retains legacy branding (`EcommerceHubSDK` and `@ubi/sdk`).

The goal of this task is to:
1. Rebrand the SDK package from `@ubi/sdk` to `@hexta/sdk` and introduce `HextaSDK` (preserving `EcommerceHubSDK` as an alias for backwards compatibility).
2. Upgrade `StorageAdapter` and `IdentityClient` to manage both `access_token` and `refresh_token`.
3. Implement an Axios response interceptor for automatic `401 Unauthorized` token refresh with request queuing, token persistence, error purging, and expiration callbacks.
4. Add comprehensive unit tests using `axios-mock-adapter` and `vitest` covering single and concurrent request recovery, refresh failure handling, and token lifecycle.
5. Ensure `apps/web` compiles cleanly and consumes the modernized `@hexta/sdk`.

---

## 2. Current State Analysis
- **Package name & branding**: `packages/sdk/package.json` is named `@ubi/sdk` and exports `EcommerceHubSDK`.
- **Token management**: `IdentityClient.login` and `IdentityClient.register` extract only `access_token` and write it to `auth_token`. `refresh_token` from the backend `LoginResponse` and `RegisterResponse` is completely discarded.
- **Interceptors**: Currently, `EcommerceHubSDK` only has:
  - Request interceptor reading `auth_token` and attaching `Authorization: Bearer <token>`.
  - Response interceptor transforming API error responses to `Error(apiError.message)`.
- **401 Handling**: When an access token expires (15m duration), all client requests fail with 401. No automatic refresh call to `/api/v1/auth/refresh` exists.
- **Testing**: `packages/sdk` currently lacks test runner configuration and unit tests for interceptors.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Add Test Runner & Mock Dependencies to `packages/sdk`
- Add `vitest` and `axios-mock-adapter` to `packages/sdk/package.json` devDependencies.
- Add `test` script: `"test": "vitest run"`.

### Step 3.2: Rebrand and Modernize `packages/sdk`
- Update `packages/sdk/package.json`:
  - Rename package from `@ubi/sdk` to `@hexta/sdk`.
- Update `packages/sdk/src/index.ts`:
  - Define `TokenPair` interface (`access_token: string; refresh_token?: string`).
  - Extend `StorageAdapter` interface and provide `MemoryStorage` in addition to `DefaultBrowserStorage`.
  - Update `IdentityClient`:
    - Store and manage both `accessTokenKey` (default: `"auth_token"`) and `refreshTokenKey` (default: `"refresh_token"`).
    - In `login()` and `register()`, persist both `access_token` and `refresh_token`.
    - Add `refreshToken()` method calling `/api/v1/auth/refresh`.
    - Add `clearTokens()` helper method.
  - Implement `HextaSDK`:
    - Rename primary class from `EcommerceHubSDK` to `HextaSDK`.
    - Re-export `export { HextaSDK as EcommerceHubSDK }` for backward compatibility.
    - Support configuration options in `SDKConfig`: `apiUrl`, `storage`, `accessTokenKey`, `refreshTokenKey`, `onAuthExpired`.
    - Implement Axios 401 response interceptor:
      - Queue concurrent requests while token refresh is in progress.
      - Handle refresh request to `/api/v1/auth/refresh` without infinite recursion.
      - On refresh success: update storage, set new bearer token, replay queued and original requests.
      - On refresh failure: purge stored tokens, notify `onAuthExpired`, reject all queued and original requests.

### Step 3.3: Write Unit Tests for SDK
- Create `packages/sdk/src/index.test.ts`:
  - Test request interceptor attaches Authorization header.
  - Test `login` persists both access and refresh tokens.
  - Test silent token refresh on 401 response: replays original request with new token.
  - Test concurrent requests during refresh: all queue and replay upon refresh completion.
  - Test failed refresh: purges tokens, triggers `onAuthExpired`, rejects requests.
  - Test that 401 on `/api/v1/auth/refresh` or `/api/v1/auth/login` does not trigger an infinite loop.

### Step 3.4: Update Consumers (`apps/web`)
- Update `apps/web/package.json` dependency from `"@ubi/sdk"` to `"@hexta/sdk"`.
- Update `apps/web/lib/sdk.ts` to import `HextaSDK` from `@hexta/sdk`.
- Update `apps/web/app/(dashboard)/tenant/page.tsx` import to `@hexta/sdk`.
- Ensure `apps/web/app/auth/callback/page.tsx` wraps search parameters in `<Suspense>`.
- Run `pnpm install` and verify `pnpm run --recursive build`.

---

## 4. Risk Assessment & Edge Cases

| Risk / Edge Case | Impact | Mitigation Strategy |
| :--- | :--- | :--- |
| **Infinite 401 refresh loops** | Browser hanging / rapid repeated requests | Mark retried requests with `_retry: true`. Prevent intercepting `/api/v1/auth/refresh` and `/api/v1/auth/login` 401s. |
| **Race conditions with concurrent 401s** | Multiple refresh calls made simultaneously, invalidating refresh tokens | Maintain an `isRefreshing` boolean flag and an array queue of pending promises (`failedQueue`). Only the first 401 triggers refresh; subsequent 401s wait in the queue. |
| **Missing refresh token in storage** | Unhandled promise rejection / hang | Check for refresh token presence prior to calling `/api/v1/auth/refresh`. If absent, immediately purge tokens and notify `onAuthExpired`. |
| **Breaking existing imports** | Existing consumers fail to resolve `EcommerceHubSDK` or `@ubi/sdk` | Export `EcommerceHubSDK` as an alias to `HextaSDK`. Keep default storage keys compatible (`auth_token` for access token). |

---

## 5. Definition of Done (DoD)
- [ ] `@hexta/sdk` builds cleanly with `tsup` producing CJS, ESM, and DTS bundles.
- [ ] All unit tests pass in `packages/sdk` via `pnpm --filter @hexta/sdk test`.
- [ ] Concurrency and error recovery logic validated with `axios-mock-adapter`.
- [ ] `apps/web` builds cleanly with `pnpm --filter web build` using the new `@hexta/sdk`.
- [ ] Code review (`agentic-memory/reviews/`) and changelog (`agentic-memory/changelogs/`) documented.
- [ ] Conventional commit created and PR opened referencing Issue #8.
