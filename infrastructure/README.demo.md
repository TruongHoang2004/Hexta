# Run the current Hexta application locally with Docker

This is the primary runbook for the current runnable system: Next.js web, Go API, Python AI runtime, PostgreSQL, Redis, and database initialization/migration jobs. Docker runs all application runtimes; host Go/Node/Python/uv installations are not required. This profile is separate from the legacy `docker-compose.yml` and never loads the root `.env`.

## Requirements and first start

Install Docker Engine/Desktop with Compose v2, start the engine, and open a POSIX shell (macOS/Linux/WSL) at the repository root. Ensure ports 3000, 8080, 55433, 56379, 58000 and 55051 are free. The first build downloads Go/Node/Python dependencies and needs network access. Resource needs depend on the host; this is not a measured VPS sizing recommendation.

```sh
./scripts/local.sh up --build -d --wait
./scripts/local.sh ps -a
```

The launcher generates three random local service/JWT secrets in `infrastructure/.env.local` on first use using a temporary Python container. It preserves that file on subsequent runs; do not commit or share it. No provider keys or Google OAuth credentials are needed. If launching from PowerShell, use WSL or create the equivalent env file and invoke `docker compose --env-file infrastructure/.env.local -f infrastructure/compose.demo.yml ...`.

Initialization creates a restricted `hexta_ai` role and `ai_runtime` schema. Atlas applies the existing Go business migrations to `public` with explicit revision metadata scope. The AI service applies its own idempotent version-001 SQL migration only to its schema. `db-init` and `migrate` exiting with code 0 is expected; five long-running services should be healthy. Existing `hexta-demo` PostgreSQL volume is retained and initialized in place; no root/legacy database is used.

| Service | Host address | Purpose |
| --- | --- | --- |
| Web | http://localhost:3000 | Register/login and current tenant UI |
| Go API | http://localhost:8080/api/v1/health | Business API and DB/Redis health |
| Swagger | http://localhost:8080/swagger/index.html | Existing HTTP API reference |
| AI HTTP | http://localhost:58000/health/ready | Runtime/DB/worker readiness and offline preview |
| AI gRPC | 127.0.0.1:55051 | Internal StartRun/GetRun/CancelRun |
| PostgreSQL | 127.0.0.1:55433 | Database hexta_demo; local-only user hexta_demo/password local-demo-only |
| Redis | 127.0.0.1:56379 | Local application cache |

Open `/register`, create an account, sign in, and open `/tenant`. No initial owner account or production seed data is injected. `NEXT_PUBLIC_API_URL` is baked into the web image as `http://localhost:8080`; changing it requires a web rebuild. Browser localhost and container names are different: API connects to `postgres:5432` and `redis:6379`, never host loopback. Google login is not configured in this profile. The legacy Go AI draft endpoint has its model URL set to an unreachable loopback endpoint to prevent provider calls; it is not the Python messaging integration.

## Verify the stack using only Docker

```sh
./scripts/local.sh exec -T ai /app/.venv/bin/python -m hexta_ai.local_smoke
./scripts/local.sh exec -T ai /app/.venv/bin/python -m hexta_ai.local_smoke submit local-restart-check
./scripts/local.sh restart ai
./scripts/local.sh up -d --wait ai
./scripts/local.sh exec -T ai /app/.venv/bin/python -m hexta_ai.local_smoke check local-restart-check
```

The first smoke creates a disposable local account and tenant, checks web login, API health, registration/login and authenticated tenant create/list. It prints no tokens/passwords. The next commands verify an accepted completed run survives service restart and StartRun remains idempotent. Reusing this run ID is safe. Unit/integration tests separately exercise recovery of an unfinished leased run and rejection of a stale worker's completion. These checks never call a provider or send a customer message.

Run all AI tests, including an isolated real PostgreSQL integration database:

```sh
./scripts/local.sh --profile test run --build --rm ai-tests
./scripts/local.sh --profile test stop test-postgres
```

The test database uses ephemeral tmpfs and separate credentials; tests never truncate the application database. For optional host development: `cd services/ai && uv sync --frozen && uv run pytest` runs offline tests, skipping the explicit-DSN integration test. See `services/ai/README.md` for host gRPC smoke/codegen.

## Logs, rebuild, stop and data

```sh
./scripts/local.sh logs --tail 100 api ai migrate
./scripts/local.sh up --build -d --wait
./scripts/local.sh down
```

`down` preserves the named database volume and therefore users, tenants, runs and idempotency history. Restarting/rebuilding the AI container does not delete runs. Back up local DB explicitly if needed:

```sh
./scripts/local.sh exec -T postgres pg_dump -U hexta_demo hexta_demo > hexta-local-backup.sql
```

The dump includes local data and should be stored privately. **Destructive reset only when intentionally deleting all local accounts, business data and AI run history:** `./scripts/local.sh --profile test down -v`. This removes this Compose project's volumes, not the legacy stack's bind mounts. Keep `.env.local` unless intentionally rotating credentials; deleting JWT secrets invalidates active sessions.

## Troubleshooting and boundaries

- Port in use: stop the conflicting local process/legacy stack or change the host mapping. If changing API/web ports, update the browser URL, CORS/frontend settings and web build argument together.
- Migration failure: inspect `logs migrate db-init`; do not delete a nonempty DB or bypass migration history. Atlas is scoped to public; Python owns ai_runtime. Old business data requires a planned upgrade, not an automatic baseline.
- Unhealthy AI: inspect readiness and service logs. AI_DATABASE_URL is mandatory at runtime; it fails closed instead of accepting volatile runs. DB downtime yields gRPC UNAVAILABLE. Logs suppress storage exception payloads and credentials.
- Containers healthy but a page fails: test `/login` and the API separately, verify the compiled public API URL, then inspect browser/network errors. This is a built application profile, not hot reload; rebuild after source changes.
- Kafka and Langfuse are **not started by this profile yet**. Kafka use cases and implementation sequence are drafted in `agentic-memory/designs/2026-10-18_kafka-event-backbone_design.md`. The legacy infrastructure's Kafka advertised localhost listener is not suitable for container clients.

This starts all currently runnable application components, not all planned MVP features. Python runs a fake script/human graph, replays whole runs after lease recovery, and does not yet persist LangGraph node checkpoints. Go's future conversation worker, current-context send gate, Pancake webhook/send integration, three real-provider adapters, messaging UI and self-hosted Langfuse remain pending. Orders and messaging remain separate. No live connectors, paid inference, VPS deployment or domain is configured here. All published ports bind to loopback; do not deploy this profile unchanged on a public server.
