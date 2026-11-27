---
epic: EPIC-004
session: 2026-11-26_scripts-permissions
status: in-review
completed: false
date: 2026-11-26
---

# Scripts and delegated script permissions (MVP 5 flow 1)

## Scope

Implement PRD MVP 5 FR-5.4/FR-5.5 (business rules 1–2, script lifecycle, acceptance "Script permission") and screens S-15 Scripts and S-17 Settings › Script permissions, end to end: schema, Go API, web, tests. Parent epic EPIC-004 (conversation domain, scripts and send policy). Out of scope: reply router, proposals, reply mode, turn expiry (parallel flows).

## Starting evidence

- Branch `worktree-agent-ad27480794474227f` from `main` at `b2d886b`.
- Spec `docs/specs/prd/mvp-5-bot-replies.md`; owner decision "Scripts and delegated permissions" in `agentic-memory/designs/system/01-product-decisions.md`; approved low-fi wireframes S-15/S-17 (Vietnamese copy).
- No prior script or script-permission code existed; coarse tenant roles did not provide these grants.

## Outcome and decisions

IMPLEMENTED (not yet merged):

- Schema (`migrations/schema/api.sql`, migration `migrations/api/20261126100000_scripts_and_permissions.sql`, additive only): `scripts` (status inactive/active/archived, `active_version_id` pointer, optimistic `version`, idempotent `create_hash`), `script_versions` (numbered, `published_at` freezes content, at most one unpublished draft per script via partial unique index, composite FK keeps the active pointer on a version of the same script and tenant), `script_member_permissions` (four booleans + optimistic `version`, FK to `tenant_members(tenant_id, user_id)` ON DELETE CASCADE so removed members lose grants).
- Go API: `ScriptService`, `ScriptStore`/`ScriptRepository`, `scriptmodel`, `mapper/script/{persistencemap,httpmap}`, DTOs, `ScriptController`, routes under `/api/v1/tenants/{tenantID}`: `GET/POST /scripts`, `GET /scripts/{id}`, `PUT /scripts/{id}/draft`, `POST /scripts/{id}/{activate,deactivate,archive}`, `GET /script-permissions`, `GET /script-permissions/me`, `POST /script-permissions/{userID}/{grant,revoke}`. Swagger regenerated.
- Internal read `ScriptService.ActiveScriptVersions(ctx, tenant)` for the future reply router: published content of active scripts in an active tenant, bounded at 500, no actor and no business authority.
- Web: new `modules/scripts` (S-15 screen and S-17 screen), sidebar group "Kịch bản / Scripts", Settings tab "Quyền kịch bản / Script permissions", vi/en messages; `apps/web/AGENTS.md` module table row added. No SDK change: domain endpoint wrappers live in module services via `sdk.client` (existing pattern).

DECIDED in this session (implementation choices the owner should confirm):

- A draft version is editable until its first activation; from then its content is immutable (repository predicate `published_at IS NULL`). Repeated "Lưu nháp" rewrites the single open draft instead of adding versions.
- `script.activate` may also re-activate an earlier published version (pointer switch, no content copy). Archived is terminal.
- `script.create` covers only the create call; editing afterwards needs `script.edit`. Reads (list/detail/history) need only active membership; the permission list is owner-only.
- Grants/revokes carry the member's permission-set version; a stale version returns 409 ("Quyền đã được cập nhật ở nơi khác").
- Create uses a UUID `Idempotency-Key` as the script ID, like workflow definitions.

## Verification and evidence

- `make migrate-hash svc=api`, `make migrate-validate svc=api`: PASS; `atlas migrate diff --env api` (ephemeral docker dev DB) reports desired schema and history in sync.
- `sh scripts/check-gofmt.sh`, `go vet ./services/api/... ./packages/shared/...`: PASS.
- `go test -race ./services/api/... ./packages/shared/...` with `API_TEST_DATABASE_URL` pointing at a disposable `postgres:17.7-alpine3.23` container: PASS (service unit tests with an in-memory store, repository DB tests, controller HTTP acceptance test, Swagger contract test).
- `pnpm --filter web lint`, `pnpm --filter web build`: PASS.
- `python3 scripts/check-agent-instructions.py`: PASS.
- NOT_RUN: browser acceptance of S-15/S-17 and a live HTTP smoke test (the API needs Redis/Kafka and its default config targets the local application database); SDK checks (SDK unchanged).

## Remaining work and handoff

- Integrator: re-hash `migrations/api/atlas.sum` after merging the parallel flows; resolve expected conflicts in `api.sql`, `router.go`, bootstrap, nav config and i18n files.
- Reply-router flow: consume `ScriptService.ActiveScriptVersions` and record `VersionID` as the proposal source.
- Owner confirmation of the decisions above; browser acceptance pass on S-15/S-17.
- Not done until merged into `main`.
