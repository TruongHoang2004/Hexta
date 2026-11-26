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
- **Atlas AI dev database**: default `docker+postgres://_/pgvector/pgvector:0.8.1-pg17/dev` (pinned to `-bookworm` in the follow-up below). Atlas Community
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
## Review follow-up — 2026-11-26 (independent review: OWNER_REVIEW_REQUIRED)

The remaining-work list above is superseded by this section; the earlier text is kept as history.

Owner decisions 2026-11-26 (DECIDED): production AI is OFF until fully configured (default config never calls a
paid model; bot uses exact scripts only, everything else goes to staff); enabling turns on Claude + Jev + Gemini
embeddings together; partial config is fail-safe (AI disabled with a log line, never a crash that rolls back a
deploy; `AI_JEV_ENABLED=true` without `TYPESAFE_API_KEY` is logged as an error and Jev stays off). Stock changes sync
to AI knowledge immediately. This replaces the earlier production default `anthropic`/`claude-sonnet-5-5`.

IMPLEMENTED (local/unmerged), per finding:

- **B1 collation guard.** `compose.prod.yml` one-shot `postgres-collation-guard` (pgvector image, root, no
  network, volume read-only); `postgres` depends on it (`service_completed_successfully`). Passes only for an empty
  volume or `$PGDATA/HEXTA_GLIBC_COLLATION_OK`; `prod/init-db.sh` writes the marker on a fresh glibc initdb; the
  failure message names section 10a. Compose v5.6 replaces the old postgres container before running the guard,
  so a blocked deploy means downtime until the agent's automatic rollback (rehearsed), never a glibc start.
  Inverse check: an Alpine-era compose cannot be guarded by Compose, so §10a/§9 document it and
  `scripts/deploy-agent.sh` (`musl_downgrade`) refuses to deploy or auto-roll back to a commit without the guard
  while the volume carries the marker (effective from the first agent run after this release is checked out).
- **B2 runbook (§10a).** Explicit `X`/`HEXTA_IMAGE_TAG=sha-$X` and checkout of commit X; preflight `mkdir -p
  ~/hexta-backups`, `docker pull`, TCP superuser login check (+ password realignment), disk check; empty-target
  volume copy; REINDEX DATABASE + SYSTEM of every database (template0 opened temporarily) in an isolated
  `--network none` container; `REFRESH COLLATION VERSION` only where `datcollversion IS NOT NULL`; amcheck
  `bt_index_check(…, heapallindexed)` over every valid B-tree index then `DROP EXTENSION amcheck`; marker; manual
  deploy and `state/deployed` before resuming the timer. Rollback: dump, `down`, `docker volume rm`, previous
  release `up --no-start postgres` recreates the empty volume, copy into it (refuses non-empty), start. Owner told
  to edit both `AI_PROVIDER` and `AI_MODEL`. README.demo.md and infrastructure/README.md: REINDEX loop for old
  Alpine demo/legacy volumes (snippet run in a disposable container).
- **I5.** `pgvector/pgvector:0.8.1-pg17-bookworm` (verified: `docker manifest inspect` multi-arch index; pulled
  image is Debian 12.13, glibc 2.36, PostgreSQL 17.8) in compose prod/demo/legacy/connector fixture, backend CI,
  atlas.hcl and `migrations/schema/ai.sql`.
- **I1 redaction.** House number + 1–4 street words of any case/diacritics before an administrative unit
  (`12 Nguyễn Huệ, Q1`, `123 nguyễn trãi quận 5`, `88 Đồng Khởi Quận 1`, `25 le loi q1`), unless a quantity/verb word
  is among them (`mua 2 cái về quận 7` keeps `2 cái`); `số 5, đường số 7`; phones glued to a phone keyword
  (`sđt0912345678`) or followed by letters (`0912345678nha`); 9-digit mobiles without 0 (`912345678`) unless
  dotted, after an order/code context or followed by a money unit. Negative tests: prices, quantities, order
  codes (`DH0912345678`, `mã đơn 912345678`, `#ORD-2026-912345678`, `SP0912345678`).
- **AI fail-safe.** `DisabledProvider` (`disabled`/`none`, exact script or human, reason `ai_disabled`);
  `provider_from_env` returns it on any `AIConfigError` (missing key or `AI_MODEL`, model of another family,
  invalid numbers, unknown provider) with an error log; real providers now need explicit `AI_MODEL`.
  `router_from_env(ai_enabled=…)` and `knowledge_from_env(paid_embeddings=…)` keep Jev and Gemini off while AI is
  disabled; Jev without key / invalid Jev settings / Gemini without key are logged and fall back. One startup line
  `AI runtime enabled|DISABLED: provider=… model=… router=… embeddings=…`. Consult endpoint uses a disabled
  consultant graph (static fallback, no model). Prod compose and `prod/.env.example` default `AI_PROVIDER=disabled`,
  empty `AI_MODEL`; README.deploy.md "AI settings" lists the exact enable set; misleading "re-embeds on next sync"
  comment replaced (re-sync command).
- **Stock sync.** `InventoryRepository` create/reserve/release/deduct/adjust enqueue the variant's product in the
  same transaction through the shared coalescing upsert (`knowledgeSyncCoalesce`: one row per product, fence+1);
  revision stays the newest product/variant/inventory `updated_at` (µs).
- **I4 re-sync.** `services/api/cmd/knowledge-resync` (`-tenant`, `-all-tenants`, `-status`; `API_DATABASE_URL`),
  built into the api image; `make knowledge-resync tenant=…|all=1|status=1`; repository
  `EnqueueKnowledgeResync` (active products, active scripts, live documents; failed jobs revived).
- **Minor.** `knowledge-document-form.tsx` uses `@hexta/ui` `Input`. Failed/stuck (≥ 5 attempts) knowledge sync
  counts: `KnowledgeSyncStatus`, logged by the messaging-command worker every 10 min when non-zero, printed by
  `knowledge-resync`. Attempts are not capped (backoff already capped at 600 s). AI migrations renamed
  `20261126950000/950001` (invalid hour 95) → `20261125223600/223601` and rehashed (never applied outside disposable
  DBs); API `20261127100000/100001` kept: valid timestamps already after every existing API version (a valid
  timestamp after the existing pseudo-version 20261126900200 must be on 2026-11-27).

### Rehearsals (real containers, Compose v5.6.0, disposable projects; all removed afterwards)

| Case | Result |
| --- | --- |
| (a) fresh empty volume | guard `empty data volume (fresh install)`, postgres healthy, marker written by init-db.sh, `vector 0.8.1`; restart: guard `marker present` |
| (b) Alpine 3.23 volume (`datcollversion` NULL, 7200-row text-PK probe), new compose `up -d --wait --remove-orphans` | exit 1, `REFUSING TO START POSTGRES … section 10a`; pgvector container only `Created`, never started; old compose `up` (agent rollback) → `postgres:17.7-alpine3.23` healthy, data intact |
| (c) §10a runbook on that volume | before REINDEX amcheck: `item order invariant violated for index "coll_probe_name"`; after REINDEX of 5 databases amcheck OK (hexta 167, others 164 indexes); marker; compose start: guard passes, `postgres-extensions` exit 0, amcheck via compose 167 OK, index lookup finds the row, 7200 rows |
| Rollback of the switch | dump, `down`, volume rm, old `up --no-start`, copy into empty volume (copy into non-empty refused, exit 1), Alpine healthy, musl amcheck 167 OK |
| deploy-agent `musl_downgrade` | old release + marked volume REFUSED; guarded release allowed; unmarked musl volume allowed; no volume allowed and none created |

### Verification (this follow-up)

| Scope | Command | Result |
| --- | --- | --- |
| Go format | `sh scripts/check-gofmt.sh` | PASS |
| Go build / vet | `go build` and `go vet ./services/api/... ./packages/shared/...` | PASS |
| Go tests + DB | `API_TEST_DATABASE_URL=<disposable :55439> go test -race -count=1 ./services/api/... ./packages/shared/...` | PASS (24 packages), incl. new `TestStockChangesEnqueueOneCoalescedProductSync`, `TestKnowledgeResyncAndStatusAgainstPostgres` |
| Re-sync CLI | disposable API DB, all API migrations, `make knowledge-resync status=1|all=1|tenant=…`, no flags | counts printed; 1 source enqueued; usage error exit 1 |
| AI offline | `cd services/ai && uv run pytest -q` | 260 passed, 11 skipped |
| AI + pgvector DB | disposable `pgvector/pgvector:0.8.1-pg17-bookworm` initialised by prod `init-db.sh`, AI migrations applied as non-superuser `hexta_ai`, `AI_TEST_DATABASE_URL=<superuser> uv run pytest -q` | 271 passed |
| AI startup (partial config) | uvicorn with `anthropic` and no model/key, Jev on without key, Gemini without key | ready (`provider=disabled`); `AI disabled: AI_PROVIDER=anthropic needs AI_MODEL…`, `Jev router off: …`, `AI runtime DISABLED …` |
| Atlas | `make migrate-hash svc=ai`, `make migrate-validate`; `atlas migrate diff ci_schema_check --env api` / `--env ai` | PASS; both synced |
| Proto / Swagger | no proto or HTTP contract change in this follow-up; `go test ./services/api/docs/...` inside the Go run | unchanged; PASS |
| Web | `pnpm --filter web lint`, `pnpm --filter web exec tsc --noEmit`, `pnpm --filter web build` | PASS |
| Compose | `config -q` (dummy env) for prod, prod+shared-edge, prod+build, demo, demo+mock, demo test profile, legacy docker-compose.yml, connector fixture | PASS (demo+ai-live unchanged, needs a private token: NOT_RUN) |
| Instructions | `python3 scripts/check-agent-instructions.py` | PASS (0 errors) |
| Live Claude / Jev / Gemini, server rollout, deploy agent on a server | — | NOT_RUN |

### Remaining owner actions and handoff

- Merge the release only with the §10a maintenance window (pause timer, backup off-server, manual switch).
  Existing `.env`: set `AI_PROVIDER=disabled` and empty `AI_MODEL` (or the full enable set), never only one of them.
- To enable AI: the seven settings in README.deploy.md "AI settings", then `knowledge-resync -all-tenants`.
- Browser acceptance of Settings › Tài liệu and Desk citations; demo runtime check of the offline answer path.
- Follow-up: RAG answers without Jev (provider classifier never routes to `answer`).
- Not done until merged into main and verified.

## Second review follow-up — 2026-11-26 (verified findings 1–9)

IMPLEMENTED (local `main`, unpushed; commits a18b325..e3fa81f):

- **Preflight superuser check (§10a).** `-h 127.0.0.1` inside the container hits the `trust` pg_hba line, so a
  wrong password passed. The check now runs `psql -h postgres` from a separate container on `hexta-prod_default`
  with `PGPASSWORD` from `.env` (the scram rule postgres-extensions uses); the realignment reads `.env` in the shell
  and never prints it.
- **Fail-closed window.** Blocks 0–5, each a `set -euo pipefail` subshell ending in `OK <n>`, with a stop table.
  Block 0 refuses an unedited `X=XXXXXXX`, an active timer, a missing commit or a `release.env` on `origin/main`
  that does not name `sha-$X`; blocks refuse without the previous result (verified dump, non-empty copy, pinned
  tag, marker on the volume). Block 4 captures the database list before looping, counts REINDEX and amcheck
  successes, writes the marker only when both equal the total, and its EXIT trap prints `marker NOT written`.
- **Rollback.** `docker rm -f hexta-pg-reindex`; refuses unless `hexta-prod_postgres_data_alpine17` exists, holds
  `PG_VERSION` and `global/pg_control` and has no marker; post-switch dump only when postgres is running, with
  pipefail and `gunzip -t`; previous commit read from the switch record.
- **Deploy agent.** `musl_downgrade` captures `git show` before `grep` (no SIGPIPE false "no guard"); volume
  listing, image presence (`docker image inspect` or pull) and `docker run --pull never` errors return status 2,
  which refuses the deploy and the automatic rollback with a log line.
- **Restart.** Operations table: `restart postgres` skips the guard; use `stop postgres && up -d --wait postgres`.
- **AI gating.** `is_real(provider)`: Jev and paid embeddings only with a real provider (never `fake`/`disabled`,
  whatever keys); fake answers only via `AI_FAKE_ROUTER` and simulated; startup line `AI runtime SIMULATED …`;
  `openai` needs `gpt-`/`chatgpt-`/`o1`/`o3`/`o4`; invalid `AI_RAG_*` → knowledge `None` with an error log.
- **Redaction.** Standalone `ấp/thôn/khu phố/kp/tổ N`; house + street before a curated city/district list
  (accented or folded); `order 0912345678` (explicit order words only) and `912345678 cái` kept.
- **Image provenance.** `compose.prod.yml` and `deploy-agent.sh` pin
  `pgvector/pgvector:0.8.1-pg17-bookworm@sha256:3e8b3adf…5700b21` (index digest from `docker buildx imagetools
  inspect`, equal to the local RepoDigest). Block 4 writes digest, PostgreSQL 17.8, glibc 2.36, Debian 12.13 to
  `~/hexta-backups/pgvector-switch-$X.txt` and into the marker. `datcollversion` stays NULL (`REFRESH COLLATION
  VERSION` on it: `invalid collation version change`, rehearsed). Bump procedure documented ("Image pin").

### Rehearsals (bash 5.3 in a Linux container, host Docker 29.9, Compose v5.8; all removed afterwards)

Throwaway checkout with a bare origin: previous release = `origin/main` compose trimmed to `postgres` (Alpine
17.7), release X = this compose trimmed to guard/postgres/postgres-extensions; writers are sleeping stand-ins.
README blocks were extracted verbatim, with only names (`rh10a_*`), `cd` and `X` substituted and `sudo`/`systemctl`
stubbed.

| Case | Result |
| --- | --- |
| Preflight, `.env` matches | `1` |
| `.env` password differs | `password authentication failed`, exit 2; old `-h 127.0.0.1` check with `PGPASSWORD=definitely-wrong` prints `1` (the bug) |
| Realignment, then check | `1` |
| Run A: all blocks pasted at once, injected `IMMUTABLE` function using `random()` in an index | blocks 0–3 OK, 5 databases reindexed, amcheck `heap tuple (0,1) … lacks matching index tuple … "rh_inj_lying"`, `BLOCK 4 FAILED: marker NOT written`, block 5 `no marker on the data volume`; marker `test -f` exit 1; `deployed` unchanged |
| Rollback, copy missing / empty | refused before any removal; data volume and stack intact |
| Rollback after Run A | leftover reindex container removed, musl copy restored, Alpine healthy, 7200 probe rows, no marker |
| Retry with the old copy present | block 2 `copy volume not empty: STOP` |
| Run B (fault removed, copy removed) | amcheck hexta 167, others 164 OK; marker with digest/versions; guard `marker present`; `vector 0.8.1`; `OK 5` |
| Restart | `prod.sh restart postgres` leaves guard `StartedAt` unchanged; `stop` + `up -d --wait postgres` reruns it |
| `musl_downgrade` (functions sourced, disposable volume) | guarded commit / no volume / unmarked: allowed; pre-guard + marker: refused; image unavailable, `docker run` 125, daemon unreachable: refused (status 2); same on the switched rehearsal volume |

### Verification (second follow-up)

| Scope | Command | Result |
| --- | --- | --- |
| Shell | `shellcheck` (koalaman/shellcheck:v0.10.0 container) on `scripts/deploy-agent.sh`; `bash -n` | PASS |
| Go | `sh scripts/check-gofmt.sh`; `go build` + `go vet ./services/api/... ./packages/shared/...` | PASS; no Go change, so `go test -race` not rerun |
| AI offline | `cd services/ai && uv run pytest -q` | 314 passed, 11 skipped (new matrix, openai prefix, RAG fail-safe and redaction cases fail on the previous code) |
| Compose | `config -q` (dummy env) prod, prod+shared-edge, prod+build | PASS |
| Instructions | `python3 scripts/check-agent-instructions.py` | PASS (0 errors) |
| AI + pgvector DB tests, server rollout, agent on a server, live providers | — | NOT_RUN |

Remaining: unchanged from the previous handoff (owner window per §10a, browser acceptance, RAG without Jev). Not
done until merged into main and verified.
