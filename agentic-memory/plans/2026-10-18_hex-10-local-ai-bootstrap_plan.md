# HEX-10: Local AI service bootstrap

## Goal

Prepare a reproducible local Python AI service and isolated Compose dependencies without a VPS, domain, provider key or live Pancake access. Do not claim the full MVP or clean full-stack setup is complete from a bootstrap.

## Current state

Go/Next.js foundations and development infrastructure exist. There is no Python service. Existing Compose starts unnecessary stores and uses host ports unsuitable for direct VPS reuse. Preserve user changes and existing runtime configs.

## Steps

1. Add a standalone Python project, lockfile, Dockerfile and fake-only LangGraph execution endpoint.
2. Keep the HTTP service internal-token protected and bind published ports to localhost. Never enable real inference or real sending in the bootstrap.
3. Add isolated Compose PostgreSQL/Redis/AI services and documentation for local startup, checks and limitations. Do not load the repository root .env or paid provider credentials.
4. Test graph routing, missing data, validation/authentication and health through FastAPI's test client; validate Compose and container startup if a Docker engine is available.
5. Record exact checks and limitations in the changelog and Notion HEX-10. Do not mark the complete local environment verified until Go/web/migrations and the selected observability stack are exercised.

## Risks and acceptance

The fake model must be visibly simulated, no auto-send side effects, no persistent-run API promises from an in-memory bootstrap, no raw secret outputs. Provider adapters, durable jobs/checkpoints and Langfuse deployment remain subsequent slices. Use meaningful behavioral tests, not dependency version snapshots. A Docker engine or network restriction can limit verification; disclose it rather than installing alternative runtimes without user direction.

## Approved gRPC follow-up

Add shared Protobuf and generated Go/Python bindings, a simulation StartRun/GetRun/CancelRun adapter, token metadata, tenant-qualified lookup, idempotency conflicts, bounded volatile run storage and cancellation. Verify actual Go-to-Python calls in Docker. Keep durable persistence, provider adapters and business worker/send-policy integration explicitly pending. The HTTP preview remains diagnostic only.
