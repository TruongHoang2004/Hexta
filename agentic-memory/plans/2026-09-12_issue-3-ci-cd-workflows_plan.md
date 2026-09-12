# Implementation Plan: GitHub Actions CI/CD Workflows for Monorepo

- **Issue**: [#3](https://github.com/TruongHoang2004/Hexta/issues/3)
- **Date**: 2026-09-12
- **Branch**: `task/issue-3-ci-implement-github-actions-ci-cd-workfl`
- **Scope**: CI/CD automation across backend, frontend monorepo, database migrations, and Docker builds

---

## 1. Objectives & Overview
Establish a robust, modular, and performant GitHub Actions CI/CD infrastructure for the Hexta monorepo.
This includes:
1. **Backend CI** (`.github/workflows/backend-ci.yml`): Go test, race detection, vetting, Swagger validation, and Atlas migration checks.
2. **Frontend CI** (`.github/workflows/frontend-ci.yml`): pnpm workspace setup, dependency caching, SDK/UI builds, linting, and Next.js production builds.
3. **Docker Delivery** (`.github/workflows/docker-ci-cd.yml`): Buildx matrix testing on PRs and automated pushing to GitHub Container Registry (`ghcr.io`) on `main`.

---

## 2. Impacted Components & Files

### New Files
- `.github/workflows/backend-ci.yml`: Go linting, tests with race detection, migration integrity check.
- `.github/workflows/frontend-ci.yml`: pnpm setup, package builds (`@ubi/sdk`), linting, and Next.js app builds.
- `.github/workflows/docker-ci-cd.yml`: Multi-target matrix Docker builds with Buildx and GHA caching.

### Modified Files
- `GEMINI.md`: Add CI/CD guidelines and workflow explanations.
- `README.md`: Document CI/CD pipeline, status badges, and workflow triggers.

---

## 3. Implementation Steps

### Step 3.1: Backend CI Workflow (`backend-ci.yml`)
- Trigger on push to `main` and PRs affecting Go services (`services/**`, `packages/shared/**`, `go.work*`, `migrations/**`).
- Set concurrency with `cancel-in-progress: true` on PRs.
- Matrix or direct job with `actions/setup-go@v5` (Go 1.24/1.23+ with module caching).
- Run `go vet ./...` and `go test -v -race -cover ./...` across workspace modules.
- Check Atlas migration schema directory hash validity.

### Step 3.2: Frontend CI Workflow (`frontend-ci.yml`)
- Trigger on push to `main` and PRs affecting TypeScript/frontend code (`apps/**`, `packages/**`, `pnpm-*`).
- Set concurrency with `cancel-in-progress: true`.
- Setup Node.js (v20) and pnpm (v10 via `pnpm/action-setup`).
- Cache pnpm store (`pnpm store path`).
- Run `pnpm install --frozen-lockfile`.
- Build shared packages (`pnpm --filter @ubi/sdk run build`).
- Run typecheck and linting.
- Build Next.js applications (`pnpm --filter web run build`).

### Step 3.3: Docker CI/CD Workflow (`docker-ci-cd.yml`)
- Trigger on PR (build only, no push) and push to `main` (build and push to `ghcr.io`).
- Use `docker/setup-buildx-action@v3` and `docker/login-action@v3` for GHCR (`GITHUB_TOKEN`).
- Matrix strategy:
  - `services/api`: context `.`, dockerfile `services/api/Dockerfile`
  - `apps/web`: context `.`, dockerfile `apps/web/Dockerfile`
  - `migrator`: context `.`, dockerfile `Dockerfile.migrate`
- Use GitHub Actions cache backend (`type=gha`).

### Step 3.4: Documentation Update
- Update `GEMINI.md` and `README.md` to document the CI/CD pipeline.

---

## 4. Definition of Done
- All 3 workflow files created in `.github/workflows/`.
- YAML syntax validated and lint-free.
- Concurrency and path filtering configured to optimize GitHub Actions minutes.
- Documented in `GEMINI.md`, `README.md`, and recorded in `agentic-memory/`.
