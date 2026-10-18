# Hexta AI bootstrap

This is a simulation-only service scaffold. A small LangGraph selects a unique active script by exact trigger match, or returns a human handoff. It never calls a model provider or Pancake and does not implement the full routing policy. The `steps` response exposes the path taken without sending traces to a hosted platform.

## Run without Docker

Install Python 3.12–3.14 and uv 0.11.11. From this directory:

```sh
uv sync --frozen
uv run pytest
export AI_PROVIDER=fake
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

No durable jobs/checkpoints, provider adapters, Langfuse exporter, business-worker integration, customer UI or send endpoint is implemented here. The fake matcher is an environment probe, not a trained router or a product capability claim. Real-provider configuration is deliberately rejected until the adapters and contract tests exist. The approved OpenAI/Anthropic/Gemini support remains implementation work.

Do not put API keys in this README, request body or shell command history. Deployment secrets and self-hosted tracing will be configured separately. No API key is needed for these tests.

## Internal gRPC simulation

`proto/hexta/ai/v1/service.proto` is the shared versioned contract. The Python server starts alongside HTTP (port 50051 in Docker); Go bindings live in `services/api/internal/infrastructure/ai/rpc`. RPCs are unary `StartRun`, `GetRun`, `CancelRun`. Authorization metadata is `authorization: Bearer <service token>` on every call; Go authenticates tenant business access before constructing the snapshot. This shared credential identifies the Go service, not a tenant user.

StartRun returns queued/current status without waiting for graph completion. `(tenant_id, run_id)` identifies a run; identical requests return the existing run, changed input returns ALREADY_EXISTS. Get/Cancel require the tenant. Context revision and control generation are echoed so the future Go acceptance gate can reject stale results. Cancel is best effort; completed results remain completed. It never authorizes an outbound send.

**Bootstrap storage is in memory, limited to 1000 runs per process including completed runs. Restart clears all runs and idempotency history.** At capacity StartRun returns RESOURCE_EXHAUSTED; restart only the simulation service to clear it. No durable acknowledgement, crash recovery, multi-worker coordination or checkpoint persistence is claimed. Run one process; do not add uvicorn workers. Production persistence must replace this registry before real inference/sending.

The local profile uses plaintext on loopback/private Docker networking only. Do not publish gRPC publicly; configure authenticated TLS or mTLS if crossing hosts. HTTP remains for health and offline preview, while browser business APIs remain in Go.

With Compose running, from repository root:

```sh
(cd services/api && go run ./cmd/ai-smoke)
```

Use the same exported DEMO_AI_INTERNAL_TOKEN as Compose. This command checks Go/Python interoperability, rejects missing auth, retries the same run ID, polls a candidate, and never sends a customer message. See `proto/README.md` for code generation.
