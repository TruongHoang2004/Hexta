---
epic: EPIC-002
session: 2026-11-26_ai-knowledge-integration
status: in-review
completed: false
date: 2026-11-26
---

# Go adapter, knowledge sync, pgvector rollout and offline AI demo

## Scope

Integration phase after [2026-11-26_ai-jev-rag](2026-11-26_ai-jev-rag.md) (Python side): the Go adapter
("Go adapter follow-up" there), durable knowledge ingest, owner-managed knowledge documents, pgvector in every
environment, E2E smoke fixes (AI run address, offline demo answers), the router bot-toggle race and a Vietnamese
clarifying question. Primary epic EPIC-002; affects EPIC-004 (reply router, Desk proposals) and EPIC-008
(compose/prod). Code-only for production: nothing was deployed or connected to a server.

## Starting evidence

Worktree branch `worktree-agent-a7dfa57feac96c0ba`, reset to local main `da94bb1`, merged
`worktree-agent-a95cbb162d3887625` (Jev/RAG, based on `ce9fef2`; no conflicts), later merged local main `7f528fd`.

Owner decisions 2026-11-26 (DECIDED): pgvector image `pgvector/pgvector:0.8.1-pg17` in all environments
including production with superuser-created extension; production default `AI_PROVIDER=anthropic`,
`AI_MODEL=claude-sonnet-5-5`. Reply-mode D1–D3 unchanged ([product decisions](../../../designs/system/01-product-decisions.md)).

## Outcome and decisions

IMPLEMENTED (local/unmerged):

- **Go adapter** (`internal/infrastructure/ai/run_client.go`, `reply_router_service.go`): Run fields 20–28 mapped
  to `types.AIRun`; `AllowKnowledgeAnswer=true`. `generate` → pending answer proposal with `RequiresReview` and
  validated citations (`kind:id`, kinds product/document/script/fact, identifier ids, deduplicated, ≤ 20) stored
  in `messaging_reply_proposals.cited_source_ids` and labelled on the Desk (product name, document title, script
  name; tenant-scoped batch lookups). `clarify` carries `RequiresReview` (also when `candidate_kind=generated`).
  `human` uses `handoff_reason` only from the AI vocabulary (customer_requested_staff, insufficient_information,
  ai_error, business_review), else insufficient_information. `autoSendEligible` returns false for
  `RequiresReview`; D2 already excluded `RouteAnswer` (confirmed by `TestAutoSendNeverSendsAIWrittenAnswers`).
- **Router race** (review finding): when route() saw `off` (nil candidate) but apply() sees the bot on and the job
  is not fenced, the job is requeued (`available_at=now`, version+1, not counted as decided) instead of superseded.
- **Knowledge sync**: transactional outbox table `knowledge_sync_jobs` (one row per source, `fence` advanced by
  every enqueue and lease, `leased` state with 60 s lease, retry backoff 5 s·2^n capped at 600 s, `done` with the
  accepted revision, `failed` with `last_error`). Enqueued in the same transaction by product save/archive, script
  activate/deactivate/archive and active-script rename, and knowledge document create/update/delete. Worker
  `KnowledgeSyncService` in `messaging-command` (`BuildKnowledgeSync`, shares the router's AI client) reads current
  state after commit: products (title name; body category/description/variants; attributes `sku`, `price` only when
  provided, `stock` = available qty, suffixed per variant; revision = newest of product/variant/inventory
  `updated_at` in µs), active scripts (body triggers + answer, attribute `version_id`, revision = script version;
  inactive/archived → delete), documents (revision = version; tombstone → delete); missing row → delete at
  MaxInt64. UNAVAILABLE/RESOURCE_EXHAUSTED/other/AI-unconfigured → retry with the same revision;
  INVALID_ARGUMENT/FAILED_PRECONDITION → `failed` with reason. Data migration enqueues existing active products and
  scripts once.
- **Knowledge documents**: `knowledge_documents` (title ≤ 200, body ≤ 20000, version, soft-delete tombstone);
  `GET/POST /tenants/{id}/knowledge-documents`, `PUT/DELETE …/{id}` (owner writes, members read, `can_manage`,
  `sync_state`); Swagger regenerated. Web Settings › Tài liệu (`modules/knowledge`) and citations on the Desk card.
- **Python** (offline demo): `AI_FAKE_ROUTER=keyword` deterministic Vietnamese router (Jev wins when active,
  simulated profile kept); fake provider writes a grounded Vietnamese answer from the best record with citation;
  hashing embeddings fold diacritics (`hashing-768-v2`); demo floor `AI_RAG_MIN_SIMILARITY=0.15`; static clarify
  text Vietnamese by default and configurable via `AI_CLARIFY_TEXT` (no review needed, may auto-send under D2).
- **Infrastructure**: Postgres images → `pgvector/pgvector:0.8.1-pg17` (demo, legacy local, connector fixture,
  prod, backend CI). Superuser `CREATE EXTENSION vector` in `local/init.sql`, `prod/init-db.sh` and a new prod
  `postgres-extensions` one-shot before `migrate-ai` (covers existing volumes). `ai.run_addr: ${AI_RUN_ADDR}` in
  local/prod `api.yaml` (config expands `${VAR}` with `os.ExpandEnv`; keys absent from YAML are not read from env);
  `messaging-command` gets `AI_RUN_ADDR=ai:50051` and `AI_INTERNAL_TOKEN` in demo and prod. Prod AI defaults
  anthropic/claude-sonnet-5-5 plus pass-through of Jev and embedding settings. Rollout runbook:
  [README.deploy.md §10a](../../../../infrastructure/README.deploy.md); demo notes in README.demo.md.
- **Atlas AI dev database**: default `docker+postgres://_/pgvector/pgvector:0.8.1-pg17/dev`. Atlas Community
  rejects `CREATE EXTENSION` in desired SQL and drops the extension from the dev database after each diff (a
  persistent dev DB with a pre-created extension works once, then fails with `type "public.vector" does not
  exist`). The desired `ai.sql` therefore starts with `DO $$ BEGIN CREATE EXTENSION IF NOT EXISTS vector … END $$`;
  the dev login must be allowed to create extensions. The Atlas `docker` block with `baseline` needs `atlas login`.

Migrations: `migrations/api/20261127100000_knowledge_documents_and_sync.sql` (`atlas migrate diff --env api`,
renamed from Atlas's auto-version and rehashed) and `20261127100001_backfill_knowledge_sync.sql`
(`atlas migrate new`, custom SQL). No new AI migration. The proposals change adds a NOT NULL column with a constant
default (metadata-only) and a CHECK validated under a short exclusive lock.

PROPOSED / not decided: knowledge documents live in a new `modules/knowledge` web module shown in the existing
Settings sidebar group (as Reply mode and Script permissions are), not a new group; inventory changes from order
reservations do not enqueue a product sync (stock refreshes on the next product change or backfill); with
`AI_PROVIDER=anthropic` but Jev disabled, the provider classifier path never routes to RAG answers (needs Jev).

## Verification and evidence

| Scope | Command | Result |
| --- | --- | --- |
| Go format | `sh scripts/check-gofmt.sh` | PASS |
| Go build / vet | `go build` and `go vet ./services/api/... ./packages/shared/...` | PASS |
| Go tests + DB | `API_TEST_DATABASE_URL=<disposable hexta-integ-pg :55439> go test -race ./services/api/... ./packages/shared/...` | PASS (all packages) |
| AI offline | `cd services/ai && uv run pytest -q` | 222 passed, 11 skipped |
| AI + pgvector DB | disposable `pgvector/pgvector:0.8.1-pg17`, `MIGRATE_TARGET=ai scripts/migrate.sh apply`, `AI_TEST_DATABASE_URL=… uv run pytest -q` | 233 passed |
| Non-superuser AI migrator | role without superuser, extension pre-created by superuser, `scripts/migrate.sh apply` | PASS (3 migrations) |
| Atlas | `make migrate-validate`; `make migrate-hash svc=api`; `atlas migrate diff ci_schema_check --env api` and `--env ai` (default docker dev URLs) | PASS; both synced |
| API upgrade | disposable DB: first 29 API migrations, seed tenant + active/archived product, apply rest | PASS; one pending job for the active product only |
| Proto parity | regenerate Python and Go (protoc 34.1, protoc-gen-go v1.36.11, -go-grpc 1.6.1) to scratch, `diff -r` | identical |
| Swagger | `make swagger`; `go test ./services/api/docs/...` | regenerated; PASS |
| Web | `pnpm --filter web lint`, `pnpm --filter web build`, `pnpm --filter web exec tsc --noEmit` | PASS |
| Compose | `docker compose --env-file <dummy> … config -q` for demo, demo+mock, demo test profile, prod, prod+shared-edge, legacy docker-compose.yml, connector fixture | PASS |
| Instructions | `python3 scripts/check-agent-instructions.py` | PASS |
| Live Claude / Jev / Gemini, browser, compose runtime, prod rollout | — | NOT_RUN (no keys/approval; code-only for production) |

## Remaining work and handoff

- Owner: `ANTHROPIC_API_KEY` before the release that carries this (else the AI service fails closed and the deploy
  rolls back), optional `TYPESAFE_API_KEY` + `AI_JEV_ENABLED`, `GEMINI_API_KEY` + `AI_EMBEDDING_PROVIDER=gemini`;
  run the one-time pgvector switch (§10a: pause deploy timer, backup, volume copy, REINDEX, refresh collation).
- Browser acceptance of Settings › Tài liệu and the Desk citations; demo runtime check of the offline answer path.
- Follow-ups: stock-only changes enqueue; a resync command after switching embedding models; RAG answers without
  Jev (provider classifier never routes to `answer`).
- Not done until merged into main and verified.
