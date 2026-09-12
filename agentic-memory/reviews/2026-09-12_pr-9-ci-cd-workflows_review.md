# Code Review Report: PR #9 (CI/CD GitHub Actions Workflows)

- **PR**: [#9](https://github.com/TruongHoang2004/Hexta/pull/9)
- **Branch**: `task/issue-3-ci-implement-github-actions-ci-cd-workfl`
- **Reviewed by**: AI Pair Programmer (Antigravity)
- **Date**: 2026-09-12
- **Result**: ❌ **CHANGES REQUESTED** (Build & CI Failures Detected)

---

## 1. Scope of Changes
- `.github/workflows/backend-ci.yml`: Go workspace setup, `go vet`, race tests, Atlas migration directory hash integrity check.
- `.github/workflows/frontend-ci.yml`: pnpm setup, SDK build, ESLint check, and Next.js production build.
- `.github/workflows/docker-ci-cd.yml`: Matrix Docker Buildx builds with GHCR publish.
- `apps/web/app/auth/callback/page.tsx`: Wrapped `useSearchParams()` in React `<Suspense>`.
- `GEMINI.md`: Monorepo CI/CD architecture documentation.

---

## 2. CI Verification Findings (Failures Detected on GitHub Actions)

### 🔴 Issue 1: Frontend CI Failing on ESLint (`run 34683202812`)
`pnpm --filter web run lint` failed with 4 errors:
1. `apps/web/app/(auth)/login/page.tsx:27:19`: Unexpected `any` (`@typescript-eslint/no-explicit-any`).
2. `apps/web/app/(auth)/register/page.tsx:40:19`: Unexpected `any` (`@typescript-eslint/no-explicit-any`).
3. `apps/web/components/auth-nav.tsx:17:5`: Synchronous `setState` in `useEffect` body (`react-hooks/set-state-in-effect`).
4. `apps/web/store/useAuthStore.ts:7:18`: Unexpected `any` (`@typescript-eslint/no-explicit-any`).

### 🔴 Issue 2: Docker Build Context Failures (`run 34683202902`)
1. **`services/api/Dockerfile`**:
   Failed at `RUN go mod download`. In a multi-module workspace with `go.work`, running `go mod download` in `/app` errors because there is no root module.
   *Fix*: Run `go work sync` or download modules within their respective subdirectories (`services/api` and `packages/shared`).
2. **`apps/web/Dockerfile`**:
   Failed at `RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile`. Node 22 alpine requires activating pnpm version explicitly via corepack:
   *Fix*: `RUN corepack enable && corepack prepare pnpm@10.23.0 --activate`.

### 🔴 Issue 3: Backend CI Multi-Module Coverage Toolchain Failure (`run 34683201511`)
Failed with `go: no such tool "covdata"`.
Passing `-coverprofile=coverage.out` across multiple distinct modules in a workspace triggers the `covdata` tool which fails when Go toolchains mismatch between runner Go 1.24 and `go.work` toolchain directive.
*Fix*: Remove `-coverprofile` across multi-module invocation or run per-package module test loops.

---

## 3. Suggestions & Architecture Improvements
- **Monorepo Coverage**: `frontend-ci.yml` currently only lints and builds `apps/web`. The other workspace apps (`apps/admin`, `apps/landing`, `apps/docs`) and packages (`packages/ui`) should be included in the CI matrix or recursive check (`pnpm -r lint`).
- **Language Standard**: In `apps/web/app/auth/callback/page.tsx:46`, the fallback string `"Đang tải..."` should be changed to `"Loading..."` per `GEMINI.md`.
- **GHCR Image Case Sensitivity**: In `docker-ci-cd.yml`, repository paths with uppercase characters (`TruongHoang2004/Hexta`) should be lowercased to prevent Docker tag reference rejection.
