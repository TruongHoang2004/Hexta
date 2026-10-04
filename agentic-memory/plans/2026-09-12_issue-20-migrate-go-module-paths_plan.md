# Implementation Plan: Migrate Go Module Paths and Internal Imports from GitLab to GitHub

- **Issue**: [#20](https://github.com/TruongHoang2004/Hexta/issues/20)
- **Title**: refactor(go): migrate Go module paths and internal imports from gitlab to github
- **Date**: 2026-09-12
- **Branch**: `task/issue-20-refactor-go-migrate-go-module-paths-and-`
- **Scope**: Entire Go codebase (`services/api`, `packages/shared`, `test`, `migrations/api`, protobuf, documentation)

---

## 1. Overview & Goal
Migrate all Go module declarations, protobuf packages, database migration tooling, and internal import paths from the legacy GitLab namespace (`gitlab.com/ecommercehub1/...`) to the current GitHub repository namespace (`github.com/TruongHoang2004/Hexta/...`).

This migration eliminates namespace discrepancies, prevents broken external `go get` / `go install` operations, cleans up generated Swagger schemas, aligns protobuf definitions, and ensures all tools (e.g. Atlas schema loader) point to the canonical module paths.

---

## 2. Current State Analysis
1. **Module Definitions (`go.mod`)**:
   - `packages/shared/go.mod` defines module `gitlab.com/ecommercehub1/shared`.
   - `services/api/go.mod` defines module `gitlab.com/ecommercehub1/api` and requires `gitlab.com/ecommercehub1/shared v0.0.0` with a local `replace` directive `replace gitlab.com/ecommercehub1/shared => ../../packages/shared`.
   - `test/go.mod` defines module `gitlab.com/ecommercehub1/test` and has an obsolete require on `gitlab.com/ecommercehub1/lib`.
   - `go.work` tracks `./packages/shared` and `./services/api`.
2. **Protobuf & Generated Bindings**:
   - `packages/shared/proto/identity.proto` declares `option go_package = "gitlab.com/ecommercehub1/shared/gen/go/identity/v1;identityv1";`.
   - `packages/shared/gen/go/identity/v1/identity.pb.go` contains the compiled descriptor with `gitlab.com/ecommercehub1/shared/gen/go/identity/v1`.
3. **Go Source Code Imports**:
   - ~35 files in `services/api` import `gitlab.com/ecommercehub1/api/...` and `gitlab.com/ecommercehub1/shared/...`.
   - `packages/shared/pkg/validator/validator.go` and `validator_test.go` import `gitlab.com/ecommercehub1/shared/pkg/logger`.
   - `test/cmd/main.go`, `test/mock/user.go`, `test/e2e/auth_e2e_test.go`, and `test/stress/auth_stress.go` import `gitlab.com/ecommercehub1/test/...`.
4. **Tooling & Schema Migration**:
   - `migrations/api/atlas.hcl` executes `"go", "run", "gitlab.com/ecommercehub1/api/cmd/tools"`.
5. **Swagger OpenAPI Docs**:
   - `services/api/docs/` contains generated definitions with prefix `gitlab_com_ecommercehub1_api_internal_present_http_dto.*`.

---

## 3. Task Breakdown (Step-by-Step)

### Step 3.1: Update Go Module Declarations and Workspace
- Update `packages/shared/go.mod`:
  - `module github.com/TruongHoang2004/Hexta/packages/shared`
- Update `services/api/go.mod`:
  - `module github.com/TruongHoang2004/Hexta/services/api`
  - Replace required module: `github.com/TruongHoang2004/Hexta/packages/shared v0.0.0`
  - Update replace directive: `replace github.com/TruongHoang2004/Hexta/packages/shared => ../../packages/shared`
- Update `test/go.mod`:
  - `module github.com/TruongHoang2004/Hexta/test`
  - Reconcile `gitlab.com/ecommercehub1/lib` dependency.
- Verify `go.work` configuration.

### Step 3.2: Update Protobuf Definition and Regenerate Bindings
- Update `packages/shared/proto/identity.proto`:
  - `option go_package = "github.com/TruongHoang2004/Hexta/packages/shared/gen/go/identity/v1;identityv1";`
- Compile protobuf using `protoc` and `protoc-gen-go` / `protoc-gen-go-grpc`.

### Step 3.3: Update Import Statements Across All Go Files
- Replace `gitlab.com/ecommercehub1/shared` with `github.com/TruongHoang2004/Hexta/packages/shared` across `packages/shared/` and `services/api/`.
- Replace `gitlab.com/ecommercehub1/api` with `github.com/TruongHoang2004/Hexta/services/api` across `services/api/`.
- Replace `gitlab.com/ecommercehub1/test` with `github.com/TruongHoang2004/Hexta/test` across `test/`.

### Step 3.4: Update Database Migration Tooling Configuration
- Update `migrations/api/atlas.hcl` line 5 to `"github.com/TruongHoang2004/Hexta/services/api/cmd/tools"`.

### Step 3.5: Regenerate Swagger Documentation
- Execute `make swagger` in `services/api` (`swag init -g cmd/main.go --parseDependency --parseInternal`).
- Verify that Swagger specs in `services/api/docs/` reflect the new GitHub namespace.

### Step 3.6: Update Documentation & Guidelines
- Update `GEMINI.md`, `agentic-memory/rules/development-rules.md`, `.agents/skills/dev-design/SKILL.md`, and README files if referencing the old package import paths.

---

## 4. Risk Assessment & Edge Cases
1. **Unresolved Go Module Checksums or Cache**:
   - *Risk*: `go.sum` contains checksums for old GitLab modules; `go test` might complain about missing modules or checksum mismatches.
   - *Mitigation*: Run `go mod tidy` in all modules (`packages/shared`, `services/api`, `test`) to ensure consistent dependencies and clean sums.
2. **Stale Generated Code**:
   - *Risk*: Pre-existing generated files might retain strings from the previous module paths.
   - *Mitigation*: Clean and regenerate Swagger docs with `swag init`, verify with `grep -rn "gitlab.com/ecommercehub1"`.
3. **Workspace vs Standalone Builds**:
   - *Risk*: Running `go build` outside `go.work` in subdirectories might fail if `replace` directives are misconfigured.
   - *Mitigation*: Ensure `services/api/go.mod` retains `replace github.com/TruongHoang2004/Hexta/packages/shared => ../../packages/shared` so both standalone builds and workspace builds succeed.

---

## 5. Definition of Done (DoD)
- [ ] All `go.mod` files use `github.com/TruongHoang2004/Hexta/...` as their module path.
- [ ] No occurrences of `gitlab.com/ecommercehub1` remain in Go source code (`*.go`) or protobuf (`*.proto`).
- [ ] `go build ./services/api/...`, `go test ./services/api/...` pass without errors.
- [ ] `go build ./packages/shared/...`, `go test ./packages/shared/...` pass without errors.
- [ ] Atlas schema loader path in `migrations/api/atlas.hcl` is updated and valid.
- [ ] Swagger documentation in `services/api/docs/` is regenerated and clean.
- [ ] Memory artifacts (`plan`, `design`, `review`, `changelog`) created and documented.
