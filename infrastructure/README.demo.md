# Local demo environment bootstrap

VPS capacity and domain selection are deferred until a runnable slice can be measured. This standalone Compose file starts isolated PostgreSQL, Redis and the simulation-only Python AI service. It does not replace the existing development Compose stack and does not load the repository root `.env`.

## Start

From the repository root with a running Docker engine and Compose v2:

```sh
export DEMO_AI_INTERNAL_TOKEN="$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')"
docker compose -f infrastructure/compose.demo.yml config --quiet
docker compose -f infrastructure/compose.demo.yml up --build -d --wait
curl --fail http://127.0.0.1:58000/health/ready
python3 services/ai/smoke.py
(cd services/api && go run ./cmd/ai-smoke)
docker compose -f infrastructure/compose.demo.yml ps
```

Health must report `provider=fake` and `simulated=true`. See `services/ai/README.md` for an authenticated preview request and tests. Ports are localhost-only: PostgreSQL 55433, Redis 56379, AI HTTP 58000 and gRPC 55051. The PostgreSQL password is intentionally local-demo-only and must not be reused on a public host. The database is initially empty; no migrations run automatically.

Stop without deleting data:

```sh
docker compose -f infrastructure/compose.demo.yml down
```

The named database volume remains. Do not use `down -v` unless you intentionally want to delete demo data.

## Verification boundaries

This bootstrap is not the full HEX-10 clean-install acceptance: Go/web, migrations, durable AI runs, actual provider adapters, webhook/send simulator and the self-hosted Langfuse stack still need integration and runtime verification. Do not point the existing Go config at this empty DB and call the setup complete. No root environment file, live connector or provider key is used by this stack.

Do not deploy this local profile unchanged to a VPS. The later VPS profile needs HTTPS, explicit image pins/resource budgets, backup/restore and restricted access to observability. Langfuse's documented storage dependencies must be capacity-planned separately; no hosted tracing is enabled here.
