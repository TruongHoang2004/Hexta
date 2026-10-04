# Code Review: Comprehensive Local Development Guide

- **Issue**: [#18](https://github.com/TruongHoang2004/Hexta/issues/18)
- **Title**: docs: create comprehensive local development guide ("How to run system in local machine")
- **Date**: 2026-09-12
- **Reviewer**: Autonomous Task Runner Agent
- **Review Scope**: Root `README.md`, `docs/local-development-guide.md`, `infrastructure/.env.example`, `services/api/.env.example`, `apps/web/.env.example`, `apps/web/.gitignore`, `infrastructure/local-all.yml`, `infrastructure/docker-compose.ui.yml`, `apps/web/app/auth/callback/page.tsx`.

---

## 1. Overview
This review verifies the implementation of the comprehensive local development guide, environment templates, and required runtime fixes for the Hexta monorepo.
All documentation and code changes were tested against live build systems (`go test ./services/api/...`, `pnpm run --recursive build`, `docker compose config`).

---

## 2. The Good
- **Exhaustive Documentation**: The root `README.md` and `docs/local-development-guide.md` provide a complete reference matrix for all 13 services and containers, including ports, protocols, credentials, and health endpoints.
- **Dual Runtime Workflows**: Clear distinction and step-by-step instructions for Mode A (Full Docker containerized quickstart) and Mode B (Hybrid local development with hot reload and live debugging).
- **Practical Troubleshooting**: Directly addresses real-world developer pitfalls:
  - Host port 5432 vs container 5433 conflict explanation.
  - Atlas migration hash drift resolution (`make migrate-hash svc=api`).
  - Volume wipe / clean slate recovery (`make clean`).
  - Google OAuth callback URI alignment.
- **Environment Safety & Reproducibility**: High-fidelity `.env.example` templates created across infrastructure, backend, and frontend without checking in real secrets.
- **Compose Integrity Fix**: Resolved broken include paths (`docker-compose.infra.yml`), missing service definitions, and network naming discrepancies (`ecommerce_network` -> `hexta_network`) in `infrastructure/local-all.yml` and `docker-compose.ui.yml`. Validated with `docker compose -f infrastructure/local-all.yml config`.
- **Prerender Robustness**: Fixed the unhandled `useSearchParams()` call in `apps/web/app/auth/callback/page.tsx` by wrapping it in a `<Suspense>` boundary and localizing user-facing copy to English per `GEMINI.md`.

---

## 3. Critical Issues (Bugs & Security)
- **None Identified**: No security credentials or private tokens are hardcoded; example secrets are explicitly marked as dummy development keys. No breaking schema or API contract changes were introduced.

---

## 4. Suggestions & Improvements
- **Future Service Migrations**: As services like `catalog` or `merchant` are implemented in Go, add their respective directories to `migrations/` and include them in `services/` alongside `services/api`.
- **Pre-commit Hooks**: In the future, consider adding git hooks or CI checks to verify that all `.env.example` files stay in sync with Viper configuration schemas.

---

## 5. Verification Checklist & Results
| Check | Command | Status | Notes |
|---|---|---|---|
| **Go Tests** | `go test ./services/api/...` | PASSED | All package tests pass cleanly |
| **Workspace Build** | `pnpm run --recursive build` | PASSED | All 6 workspace packages built successfully |
| **Docker Compose Config** | `docker compose -f infrastructure/local-all.yml config --quiet` | PASSED | Syntax and service includes valid |
| **English Language Standards** | Visual & AST inspection | PASSED | All comments, docs, and UI strings in English |
| **Links & Markdown Structure** | Markdown link audit | PASSED | Valid relative paths and anchor references |
