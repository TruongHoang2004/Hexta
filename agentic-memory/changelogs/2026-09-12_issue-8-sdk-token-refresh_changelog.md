# Development Changelog & Walkthrough: Silent Token Refresh Interceptor & SDK Rebranding

- **Issue**: [#8](https://github.com/TruongHoang2004/Hexta/issues/8)
- **Date**: 2026-09-12
- **Branch**: `task/issue-8-feat-sdk-add-silent-token-refresh-interc`

---

## 1. Change Summary
This change modernizes the client SDK package and implements silent token refresh with concurrency queueing:
- Modernized package branding from `@ubi/sdk` to `@hexta/sdk` and renamed main SDK class to `HextaSDK` (with `EcommerceHubSDK` exported as an alias for full backward compatibility).
- Enhanced token lifecycle management in `IdentityClient` and `StorageAdapter` to persist both `access_token` and `refresh_token`.
- Built an Axios response interceptor for automated `401 Unauthorized` recovery that queues concurrent requests during refresh, triggers `/api/v1/auth/refresh`, updates stored credentials, and replays all pending requests.
- Added comprehensive unit testing with `vitest` and `axios-mock-adapter` (14 passing tests).
- Updated consumers in `apps/web` (`lib/sdk.ts`, `tenant/page.tsx`, `auth/callback/page.tsx`) to adopt `@hexta/sdk` and wrapped search parameters in `<Suspense>`.

---

## 2. Impacted Components & Files

| Component / File | Action | Description |
| :--- | :--- | :--- |
| [`packages/sdk/package.json`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/packages/sdk/package.json) | `[MODIFY]` | Rebranded name to `@hexta/sdk`, added `test` script (`vitest run`), added `vitest` and `axios-mock-adapter` |
| [`packages/sdk/src/index.ts`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/packages/sdk/src/index.ts) | `[MODIFY]` | Added `TokenPair`, `MemoryStorage`, refreshed token handling in `IdentityClient`, 401 response interceptor with concurrency queueing in `HextaSDK`, and `EcommerceHubSDK` alias |
| [`packages/sdk/src/index.test.ts`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/packages/sdk/src/index.test.ts) | `[NEW]` | Unit tests for storage, tokens, 401 refresh, concurrency queueing, and failure modes |
| [`apps/web/package.json`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/package.json) | `[MODIFY]` | Replaced dependency `@ubi/sdk` with `@hexta/sdk` |
| [`apps/web/lib/sdk.ts`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/lib/sdk.ts) | `[MODIFY]` | Imported `HextaSDK` from `@hexta/sdk` and hooked `onAuthExpired` to redirect to `/login` |
| [`apps/web/app/(dashboard)/tenant/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/app/(dashboard)/tenant/page.tsx) | `[MODIFY]` | Updated import from `@ubi/sdk` to `@hexta/sdk` |
| [`apps/web/app/auth/callback/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/app/auth/callback/page.tsx) | `[MODIFY]` | Wrapped search params component in `<Suspense>` boundary and used `DefaultBrowserStorage` |
| [`.gitignore`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/.gitignore) | `[MODIFY]` | Ignored `**/node_modules/` and `.vite/` |
| [`Makefile`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/Makefile) | `[MODIFY]` | Updated example comments to `@hexta/sdk` |

---

## 3. Key Technical Decisions
1. **Request Queueing & Mutual Exclusion during Refresh**:
   When an access token expires, multiple network requests might be in flight simultaneously. Instead of making redundant calls to `/api/v1/auth/refresh` (which could cause refresh token invalidation in token-rotation schemes), only the first 401 triggers the refresh request. An internal `failedQueue` stores `{ resolve, reject }` callbacks for subsequent requests until the renewal finishes, at which point all queued requests receive the new token and execute.
2. **Loop Prevention Flagging**:
   Requests are marked with `_retry = true` to ensure that if a request fails even after token renewal, it will not enter an infinite refresh loop. Similarly, the refresh call itself is tagged with `_skipAuthRefresh: true` and strips expired `Authorization` headers.
3. **Backward Compatibility**:
   Exporting `EcommerceHubSDK as HextaSDK` guarantees that legacy codebases importing the previous class name will continue operating without disruption. Default token keys remain `auth_token` and `refresh_token`.

---

## 4. Step-by-Step Walkthrough

### 4.1 Token Storage & Refresh in `HextaSDK`
```ts
// Response interceptor snippet
if (this.isRefreshing) {
  return new Promise<string>((resolve, reject) => {
    this.failedQueue.push({ resolve, reject });
  })
    .then((token) => {
      originalRequest._retry = true;
      setHeader(originalRequest, "Authorization", `Bearer ${token}`);
      return this.client(originalRequest);
    });
}

originalRequest._retry = true;
this.isRefreshing = true;

try {
  const refreshToken = await this.storage.get(this.refreshTokenKey);
  const refreshResponse = await this.client.post("/api/v1/auth/refresh", { refresh_token: refreshToken }, { _skipAuthRefresh: true });
  const { access_token, refresh_token: newRefreshToken } = refreshResponse.data.data;

  await this.storage.set(this.accessTokenKey, access_token);
  if (newRefreshToken) await this.storage.set(this.refreshTokenKey, newRefreshToken);

  this.isRefreshing = false;
  this.processQueue(null, access_token);

  setHeader(originalRequest, "Authorization", `Bearer ${newAccessToken}`);
  return this.client(originalRequest);
} catch (refreshErr) {
  this.isRefreshing = false;
  await this.clearTokens();
  this.processQueue(refreshErr, null);
  this.onAuthExpiredCallback?.();
  throw refreshErr;
}
```

---

## 5. Verification & Testing Guide

### Automated Tests
1. **SDK Unit Tests**:
   ```bash
   pnpm --filter @hexta/sdk test
   ```
   *Result*: 14/14 unit tests passing in Vitest covering storage adapters, login/register token persistence, single 401 retry, concurrent 401 queueing, and refresh failure token purging.

2. **Recursive Workspace Build**:
   ```bash
   pnpm run --recursive build
   ```
   *Result*: All 6 workspace packages (`@hexta/sdk`, `web`, `landing`, `admin`, `docs`, `@hexta/ui`) build cleanly without errors.

3. **Backend API Tests**:
   ```bash
   go test ./services/api/...
   ```
   *Result*: Passes with 0 regressions.
