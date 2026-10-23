# Hexta agent instructions and engineering standards

This is the single canonical file for shared agent instructions. GEMINI.md is an entry-point redirect; do not duplicate shared rules there. Use English for repository artifacts and the user's language in conversation.

## Session startup and epic memory

1. Inspect git status, current branch and current user instructions. Preserve unrelated work.
2. Read [the epic index](agentic-memory/epics/README.md), select the epic(s) relevant to the task, then read each selected epic's README and latest session record.
3. Read [the canonical system-design index](agentic-memory/designs/2026-10-18_system-design_design.md) and only the DB/API/RPC/event/workflow modules relevant to the task. Apply nested AGENTS.md and applicable skills.
4. Recheck actual code, PR/CI state and external dependencies before treating a historical snapshot as current. Repository memory is read through these instructions; not every Markdown file is automatically loaded.

### Epic and session recording protocol

- Shared engineering rules live here. Cross-domain architecture lives in the modular system design. Task context, progress, blockers and handoffs live under `agentic-memory/epics/<epic-id>/`.
- Each epic has README.md containing status, completed, scope/decisions, acceptance checklist, evidence links, sessions and next work. `epics/README.md` is an index only, not a combined context dump.
- Every meaningful work session gets a separate `sessions/YYYY-MM-DD_<topic>.md` in its primary epic. Use the session template. For cross-epic work, store one primary session and link it from other affected epics; do not duplicate the full log.
- Session records include status, completed, outcome, changes/decisions, verification, unresolved work and next handoff. Never overwrite previous session history to make progress look complete; append a follow-up/correction and update the current epic summary.
- States: not-started, in-progress, in-review, blocked, done. `completed: true` is allowed only with status done and evidence that all scoped acceptance criteria are satisfied. For delivered code/docs, confirm merge into main before marking done. A completed session/slice never automatically completes its parent epic. Use last-known review state until merge is verified.
- At session end update the affected epic README, its session record, the epic index row, relevant system-design modules and CHANGELOG.md. Keep sessions scoped; do not rebuild a global SESSION_CONTEXT.md.
- Legacy SESSION_CONTEXT.md is only a redirect. Historical dated designs/plans/reviews remain evidence; they are not the current per-epic progress store.

Keep IMPLEMENTED, VERIFIED, DECIDED, TARGET and OPEN distinct. Go owns business authorization/sending; Python returns AI candidates. Do not infer provider capabilities, implement every target automatically, or inherit legacy order AI behavior as approved messaging scope. Never expose secrets in documentation/tool output.

## 1. General Principles
- **Language**: Use English exclusively for all code, comments, documentation, commit messages, and memory records.
- **Clarity over Cleverness**: Write readable, maintainable Go code. Follow "Effective Go" principles.
- **Monorepo Structure**: Keep services in `services/` and shared logic in `packages/`.
- **Consistency**: Match the existing coding style, including variable naming (camelCase), file naming (snake_case), and directory structure.

## 2. Backend (Go)
- **Frameworks**:
  - **Gin** for HTTP routing.
  - **Uber Fx** for Dependency Injection. Use `fx.Provide` and `fx.Invoke` in `internal/bootstrap/`.
  - **Gorm** for database operations.
  - **Zap** for logging. Use `log.Info`, `log.Error`, etc., from the `common/log` package.
- **Architecture**: Follow the established 5-layer architecture:
  1. `cmd/`: Entry points.
  2. `internal/present/http/controller/`: HTTP handlers. Validate requests here.
  3. `internal/core/service/`: Business logic. No infrastructure details here.
  4. `internal/repository/`: Data persistence (PostgreSQL/Gorm/Cache wrappers).
  5. `internal/infrastructure/`: Low-level clients (DB connections, Redis, etc.).
- **DTOs & Responses**:
  - Define all requests and responses in `internal/present/http/dto/`.
  - Use `response.Response[T]` for all API responses to ensure consistency and proper Swagger documentation.
- **Error Handling**: Use the `errors` package from `github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors`.
  - Return `*errors.Error` from services and repositories.
  - Map errors to appropriate HTTP status codes in controllers.

### Authentication configuration and browser credentials
- Require explicitly configured, distinct access and refresh signing keys of at least 32 bytes; fail startup on missing/invalid values. Never fall back to fixed signing keys.
- The API owns browser refresh credentials through HttpOnly, SameSite=Lax cookies, Secure in production. Do not return refresh credentials in browser JSON responses or store/read them in JavaScript.
- Credentialed browser requests require an explicit allowed frontend origin; cookie-changing auth handlers validate browser origin. Server logout must succeed before the UI reports completion.

## 3. API Documentation (Swagger)
- Every new endpoint must have **Swagger annotations**.
- Follow the existing pattern in controllers:
  ```go
  // @Summary Action name
  // @Description Detailed description
  // @Tags TagName
  // @Accept json
  // @Produce json
  // @Security BearerAuth (if applicable)
  // @Param request body dto.YourRequest true "Request data"
  // @Success 200 {object} response.Response[dto.YourResponse]
  // @Router /path [method]
  ```
- Run `make swagger` after adding or updating endpoints.

## 4. Database & Migrations
- Use **Atlas versioned SQL histories** for all database-owning services (`api` and `ai`). Applications and tests must never call GORM AutoMigrate or apply startup DDL.
- Models should be defined in the repository layer using Gorm tags.
- Update `migrations/schema/<service>.sql`, run `make migrate-diff svc=<api|ai> name=<change>`, review the generated SQL, and run `make migrate-validate` before applying. Root `atlas.hcl` is the canonical configuration; keep API and AI revision schemas independent.
- Use `make migrate svc=all` (or `make migrate svc=api` / `svc=ai`) for explicit Docker migration jobs. Host apply uses private API_DATABASE_URL/AI_DATABASE_URL with `make migrate-apply-local`. Missing legacy user/catalog targets are unsupported. Go database tests use an explicit API_TEST_DATABASE_URL and disposable databases migrated with Atlas; Python integration tests wait for the Atlas AI test job.

## 5. Frontend (Next.js)
- Use **TypeScript** with strict types; avoid explicit `any`. Run `pnpm --filter @hexta/sdk lint` for SDK type enforcement.
- Use **Tailwind CSS** for styling.
- Follow the App Router structure in `apps/web/app/`.
- Modularize logic into `apps/web/lib/services/` for API calls.

## 6. AI Interaction Rules
- **Plan first**: Always describe the plan before making non-trivial changes.
- **Check Workspaces**: Be aware of the `go.work` file. Do not use `replace` directives in `go.mod` unless absolutely necessary and documented.
- **Update Documentation**: If you change standard patterns, reflect them in this `AGENTS.md`.
- **Validation**: When fixing bugs, look for validation tags in DTOs and ensure the frontend matches the requirements.

## 7. Issue & Task Lifecycle Rules
All tasks, GitHub issues, and automated runners must strictly follow the defined lifecycle states:
- `ready`: The ticket is fully specified and ready to be picked up. Automated runners only pick tickets marked with this label.
- `in-progress`: The ticket is actively being developed. When picked, the runner removes `ready` and adds `in-progress`.
- `in-review`: Code changes, unit tests, and memory artifacts are complete, and a Pull Request has been opened for human review. The runner removes `in-progress` and adds `in-review`.
- `done`: The ticket is only counted as **Done** once the Pull Request has been **merged into `main`**. GitHub automatically closes the issue upon PR merge via `Closes #<number>`.

## 8. Skill Execution & Documentation Protocol
When executing skills (`github-task-runner`, `issue-creator`, `dev-plan`, `dev-design`, `dev-review`, `dev-changelog`, `memory-consolidator`), always adhere to the 4-phase sequence in `.agents/rules/skill-execution-protocol.md`:

1. **Pre-flight Check**: Verify that all required inputs, documents, and labels exist before taking action. Never perform blind execution.
2. **Context Loading**: Consult `AGENTS.md`, `agentic-memory/rules/`, and existing designs.
3. **Execution**: Perform the core implementation or CLI steps.
4. **Artifact & State**: Persist reasoning to `agentic-memory/` and transition lifecycle labels.

## 9. CI/CD & Automation (GitHub Actions)
- **Backend CI** (`.github/workflows/backend-ci.yml`): Checks Go formatting, runs Go tests with race detection, `go vet`, and Atlas migration integrity validation.
- **Frontend CI** (`.github/workflows/frontend-ci.yml`): Uses pnpm v10 and Node.js v20, builds SDK packages, enforces SDK strict types and tests, checks web linting, and validates Next.js production builds.
- **Docker CI/CD** (`.github/workflows/docker-ci-cd.yml`): Builds and caches container images for `services/api`, `apps/web`, and `Dockerfile.migrate` via Docker Buildx; pushes images to GitHub Container Registry (`ghcr.io`) on `main`.
- **Concurrency**: Workflows cancel superseded in-progress runs on pull requests (`cancel-in-progress: true`) to optimize runner resources.
