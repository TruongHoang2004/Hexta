# Hexta AI bootstrap

This is a simulation-only service scaffold. A small LangGraph selects a unique active script by exact trigger match, or returns a human handoff. It never calls a model provider or Pancake and does not implement the full routing policy. The `steps` response exposes the path taken without sending traces to a hosted platform.

## Run without Docker

Install Python 3.12–3.14 and uv 0.11.11. From this directory:

```sh
uv sync --frozen
uv run pytest
# Apply Atlas AI migrations before starting a database-backed runtime.
export AI_PROVIDER=fake
export AI_DATABASE_URL="postgresql://hexta_ai:local-ai-only@127.0.0.1:55433/hexta_demo?sslmode=disable"
export AI_INTERNAL_TOKEN="$(python3 -c 'import secrets; print(secrets.token_urlsafe(32))')"
PYTHONPATH=generated AI_GRPC_ADDRESS=127.0.0.1:55051 uv run uvicorn hexta_ai.app:create_app --factory --host 127.0.0.1 --port 58000
```

The token is a local service credential, not an LLM API key. Health endpoints are public and contain no secrets. Preview requires `Authorization: Bearer <AI_INTERNAL_TOKEN>` and a JSON body such as:

```json
{
  "message": "price",
  "scripts": [
    {"id": "example", "trigger": "price", "text": "Please choose a product.", "active": true}
  ]
}
```

POST it to `/internal/v1/preview`. Unknown or ambiguous inputs and unresolved template variables return a human result with no fabricated answer. This endpoint is an internal preview, not the gRPC run contract; it trusts the authorized caller's script snapshot and is not exposed to sales browsers.

## Limits and next slices

Durable whole-run jobs are implemented in PostgreSQL. Node checkpoints, provider adapters, Langfuse exporter, business-worker integration, customer UI and send endpoints remain pending. The fake matcher is an environment probe, not a trained router or a product capability claim. Real-provider configuration is deliberately rejected until the adapters and contract tests exist. The approved OpenAI/Anthropic/Gemini support remains implementation work.

Do not put API keys in this README, request body or shell command history. Deployment secrets and self-hosted tracing will be configured separately. No API key is needed for these tests.

## Internal gRPC simulation

`proto/hexta/ai/v1/service.proto` is the shared versioned contract. The Python server starts alongside HTTP (port 50051 in Docker); Go bindings live in `services/api/internal/infrastructure/ai/rpc`. RPCs are unary `StartRun`, `GetRun`, `CancelRun`. Authorization metadata is `authorization: Bearer <service token>` on every call; Go authenticates tenant business access before constructing the snapshot. This shared credential identifies the Go service, not a tenant user.

StartRun returns queued/current status without waiting for graph completion. `(tenant_id, run_id)` identifies a run; identical requests return the existing run, changed input returns ALREADY_EXISTS. Get/Cancel require the tenant. Context revision and control generation are echoed so the future Go acceptance gate can reject stale results. Cancel is best effort; completed results remain completed. It never authorizes an outbound send.

**Runtime storage is PostgreSQL**, configured by mandatory `AI_DATABASE_URL`. StartRun commits before acknowledgement. The Python-owned ai_runtime schema stores requests, digests, results, attempts and fenced renewable leases. Expired jobs can be reclaimed, up to three recovery attempts; exhausted runs fail safely. Cancellation persists and prevents publication by old workers. Graceful shutdown requeues active work. Restart retains successful/cancelled results and idempotency history. The in-memory adapter remains solely for isolated unit tests.

Atlas applies `migrations/ai` before AI starts; `make migrate svc=ai` runs the explicit job. The service only checks storage readiness and starts its worker; it never creates schemas/tables. Root `atlas.hcl` and `migrations/schema/ai.sql` define the migration workflow. Existing compatible pre-Atlas run tables are validated and adopted without deleting runs; incompatible schemas fail migration. No automatic retention deletion is enabled yet: disk usage must be monitored. A run stores the authorized input snapshot and output, so treat DB/backups as sensitive. This is whole-graph replay for a deterministic fake provider, not LangGraph node checkpoint recovery or exactly-once paid inference. Runtime processes one leased run at a time; scale/quotas are future work.

The local profile uses plaintext on loopback/private Docker networking only. Do not publish gRPC publicly; configure authenticated TLS or mTLS if crossing hosts. HTTP remains for health and offline preview, while browser business APIs remain in Go.

With Compose running, from repository root:

```sh
(cd services/api && go run ./cmd/ai-smoke)
```

Use the same exported DEMO_AI_INTERNAL_TOKEN as Compose. This command checks Go/Python interoperability, rejects missing auth, retries the same run ID, polls a candidate, and never sends a customer message. See `proto/README.md` for code generation.

## Migration-backed tests

From repository root, run `./scripts/local.sh --profile test run --build --rm ai-tests`.
An Atlas test job creates the isolated AI schema before pytest. The test image includes Atlas only for disposable-database adoption tests; the production image has no migration command. Host tests with AI_TEST_DATABASE_URL additionally require Atlas on PATH and access to the repository Atlas config. No explicit database URL means database tests skip; offline tests still run. Use a dedicated test server with database-creation rights, never the demo/application URL.
