# Changelog: GitHub Actions CI/CD Workflows for Monorepo

- **Issue**: [#3](https://github.com/TruongHoang2004/Hexta/issues/3)
- **Date**: 2026-09-12
- **Branch**: `task/issue-3-ci-implement-github-actions-ci-cd-workfl`
- **Author**: AI Pair Programmer (Antigravity)

---

## 1. Summary of Changes
Implemented comprehensive GitHub Actions CI/CD workflows across the Hexta monorepo to validate backend Go code and migrations, test and build frontend Next.js applications and SDK packages, and verify multi-service Docker container image builds with GHCR publishing. Also resolved a Next.js prerender error on `/auth/callback` by adding a React `<Suspense>` boundary.

---

## 2. Modified & Created Files
| File Path | Action | Description |
| :--- | :--- | :--- |
| `.github/workflows/backend-ci.yml` | `[NEW]` | Go vet, race test, and Atlas migration validation workflow |
| `.github/workflows/frontend-ci.yml` | `[NEW]` | pnpm setup, SDK build, linting, and Next.js build workflow |
| `.github/workflows/docker-ci-cd.yml` | `[NEW]` | Matrix Docker Buildx and GHCR delivery workflow |
| `apps/web/app/auth/callback/page.tsx` | `[MODIFY]` | Wrapped `useSearchParams()` in `<Suspense>` to fix prerendering failure |
| `GEMINI.md` | `[MODIFY]` | Documented CI/CD pipelines and concurrency standards |
| `agentic-memory/plans/2026-09-12_issue-3-ci-cd-workflows_plan.md` | `[NEW]` | Implementation plan for Issue #3 |
| `agentic-memory/reviews/2026-09-12_issue-3-ci-cd-workflows_review.md` | `[NEW]` | Code review audit for Issue #3 |
| `agentic-memory/changelogs/2026-09-12_issue-3-ci-cd-workflows_changelog.md` | `[NEW]` | Changelog and verification report for Issue #3 |

---

## 3. Verification & Validation Steps
1. **Frontend Build Verification**:
   ```bash
   pnpm --filter @ubi/sdk run build
   pnpm --filter web run build
   ```
   Output: `✓ Compiled successfully ... ✓ Generating static pages using 9 workers (8/8) in 382ms`.
2. **Backend Compilation & Test Verification**:
   ```bash
   go vet ./services/api/... ./packages/shared/...
   go test -v ./services/api/config/...
   ```
   Output: `PASS ok gitlab.com/ecommercehub1/api/config`.
