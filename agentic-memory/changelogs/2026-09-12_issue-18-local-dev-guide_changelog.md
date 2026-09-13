# Changelog: Comprehensive Local Development Guide & Monorepo Onboarding

- **Issue**: [#18](https://github.com/TruongHoang2004/Hexta/issues/18)
- **Title**: docs: create comprehensive local development guide ("How to run system in local machine")
- **Date**: 2026-09-12
- **Branch**: `task/issue-18-docs-create-comprehensive-local-developm`

---

## 1. Change Summary
Authored a complete, production-grade local development setup guide in root `README.md` and `docs/local-development-guide.md` covering prerequisites, dual execution modes (Full Docker vs Hybrid), service port reference matrix, environment configuration, database migrations with Atlas, and common troubleshooting FAQs.
Additionally added `.env.example` templates, fixed infrastructure compose configuration, and resolved a Next.js prerendering boundary issue in `apps/web`.

---

## 2. Impacted Components & Files

| File | Status | Description |
|---|---|---|
| [`README.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/README.md) | `[NEW]` | Comprehensive root onboarding and local development guide |
| [`docs/local-development-guide.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/docs/local-development-guide.md) | `[NEW]` | Standalone local development reference documentation |
| [`infrastructure/.env.example`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/infrastructure/.env.example) | `[NEW]` | Sample environment template for PostgreSQL and MinIO container parameters |
| [`services/api/.env.example`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/services/api/.env.example) | `[NEW]` | Sample environment template for JWT secrets, Google OAuth, and AI providers |
| [`apps/web/.env.example`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/.env.example) | `[NEW]` | Sample environment template for frontend API URL |
| [`apps/web/.gitignore`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/.gitignore) | `[MODIFY]` | Allow `.env.example` to be tracked while keeping `.env` files ignored |
| [`infrastructure/local-all.yml`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/infrastructure/local-all.yml) | `[MODIFY]` | Fixed broken compose include (`docker-compose.yml`), path (`services/api`), and network |
| [`infrastructure/docker-compose.ui.yml`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/infrastructure/docker-compose.ui.yml) | `[MODIFY]` | Aligned network to `hexta_network` |
| [`apps/web/app/auth/callback/page.tsx`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/apps/web/app/auth/callback/page.tsx) | `[MODIFY]` | Wrapped `useSearchParams()` in `<Suspense>` boundary and translated copy to English |
| [`agentic-memory/plans/2026-09-12_issue-18-local-dev-guide_plan.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/plans/2026-09-12_issue-18-local-dev-guide_plan.md) | `[NEW]` | Implementation plan record |
| [`agentic-memory/designs/2026-09-12_issue-18-local-dev-guide_design.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/designs/2026-09-12_issue-18-local-dev-guide_design.md) | `[NEW]` | System architecture and technical design document |
| [`agentic-memory/reviews/2026-09-12_issue-18-local-dev-guide_review.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/reviews/2026-09-12_issue-18-local-dev-guide_review.md) | `[NEW]` | Code review and verification record |
| [`agentic-memory/changelogs/2026-09-12_issue-18-local-dev-guide_changelog.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.worktrees/task-runner/agentic-memory/changelogs/2026-09-12_issue-18-local-dev-guide_changelog.md) | `[NEW]` | Development changelog record |

---

## 3. Key Technical Decisions
1. **Consolidated Root README vs Separate Docs**: Created a rich root `README.md` containing the full local development guide with badges, diagrams, and commands, alongside a dedicated `docs/local-development-guide.md` to cater to both repository front page viewers and technical documentation readers.
2. **Infrastructure Compose Alignment**: Fixed stale paths in `infrastructure/local-all.yml` (`docker-compose.infra.yml` -> `docker-compose.yml`, `service/api` -> `services/api`) and unified networks onto `hexta_network`.
3. **Next.js Suspense Boundary**: Wrapped the client component using `useSearchParams()` in a `<Suspense>` fallback, satisfying Next.js App Router prerendering requirements and enabling green monorepo builds.
4. **Environment Isolation**: Added `.env.example` templates across all layers without exposing real keys, ensuring reproducible developer setup.

---

## 4. Step-by-Step Walkthrough

### 4.1 Auth Callback Suspense & Localization Fix
Wrapped `AuthCallbackContent` in `<Suspense>` in `apps/web/app/auth/callback/page.tsx`:
```tsx
export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
          <Loader2 className="w-10 h-10 animate-spin text-primary mb-4" />
          <h2 className="text-xl font-medium tracking-tight">Loading...</h2>
        </div>
      }
    >
      <AuthCallbackContent />
    </Suspense>
  );
}
```

### 4.2 Docker Compose local-all.yml Fix
Corrected include path and network:
```yaml
include:
  - path: docker-compose.yml
  - path: docker-compose.ui.yml

services:
  api:
    build:
      context: ..
      dockerfile: services/api/Dockerfile
    container_name: hexta_api
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/api?sslmode=disable
      - REDIS_ADDR=redis:6379
    networks: [ hexta_network ]
```

---

## 5. Verification & Testing Guide

```bash
# 1. Verify Go tests pass
go test ./services/api/...

# 2. Verify all Next.js and TypeScript packages build
pnpm run --recursive build

# 3. Verify Docker compose configurations
docker compose -f infrastructure/docker-compose.yml config --quiet
docker compose -f infrastructure/local-all.yml config --quiet
```
All commands execute with zero errors.
