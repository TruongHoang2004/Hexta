# Operations, security, verification and delivery roadmap

## Local runtime (IMPLEMENTED/VERIFIED baseline)

Primary runnable guide: [Docker local runbook](../../../infrastructure/README.demo.md). From repository root:

```sh
./scripts/local.sh up --build -d --wait
./scripts/local.sh ps -a
./scripts/local.sh exec -T ai /app/.venv/bin/python -m hexta_ai.local_smoke
./scripts/local.sh --profile test run --build --rm ai-tests
./scripts/local.sh --profile test stop test-postgres
```

The launcher generates ignored `infrastructure/.env.local` service/JWT secrets using a container and never sources the root .env. Preserve this file across restarts. Build context excludes host .venv/node_modules/.next/data/environment files. Host Go/Node/Python are not needed for this run sequence. This is a built-app local profile, not hot reload.

| Service | Host endpoint | Runtime dependency |
| --- | --- | --- |
| Web | localhost:3000 | Go API; NEXT_PUBLIC_API_URL compiled into image |
| Go API | localhost:8080/api/v1/health | Atlas public migration completed, PostgreSQL, Redis |
| Python HTTP | localhost:58000/health/ready | Python-owned run schema, DB and worker |
| Python gRPC | 127.0.0.1:55051 | Internal token, mandatory AI_DATABASE_URL |
| PostgreSQL | 127.0.0.1:55433 | Named demo volume |
| Redis | 127.0.0.1:56379 | Cache only; no persistent business authority claim |
| db-init / migrate | No public endpoint | One-shot role/schema provisioning and Atlas jobs, expected exit 0 |
| test-postgres | No published port | Explicit test profile, temporary data, separate credentials |

Published host ports are loopback only. Docker-internal names are postgres/redis/api/web, not host localhost. Local API config mounts `infrastructure/local/api.yaml`; legacy Go Gemini URL is blocked on loopback, OAuth unconfigured. Python only accepts fake mode. No provider key or customer send is used by smoke commands.

`down` preserves the database; `down -v` is explicit destructive reset. Backups/logs may contain personal data. The runbook includes pg_dump, troubleshooting and restart checks. ai_runtime schema migration and Atlas public scope must remain independent. Existing DBs require an upgrade plan, not an assumed clean baseline.

## Self-host observability (DECIDED/TARGET)

Use correlation/run IDs across ingress -> outbox/Kafka -> Go job -> Python graph -> Go acceptance -> send attempt. Structured metadata: tenant/thread/turn/run, context/control/graph/prompt/model/script versions, branch/reason, source references, step start/end/duration/status, safe error code and provider-reported token usage. Cost is an estimate only when known model rates/usage exist.

Langfuse self-host is proposed; not installed in the current profile. Its storage/deployment dependencies and version pins require a verified compatible setup and capacity budget. No LangSmith Cloud/hosted telemetry is configured for graph execution. The browser receives operational state, not raw traces or hidden model reasoning. Trace loss cannot block or authorize business sends.

Trace viewer access is separate from broad tenant conversation visibility. Restrict diagnostic payloads, redact credentials/PII before export, cap payload size, and define retention and deletion policy for DB snapshots/checkpoints/traces/Kafka/backups. Raw chain-of-thought is not required or stored.

## Security boundaries

1. Authenticate end users with existing session/JWT/OAuth machinery; derive actor identity server-side. Authorize current tenant membership, actor action and resource tenant at every command/read.
2. Only owner grants/revokes script permissions and reply mode. Admin role does not automatically grant these capabilities. Current schema lacks fine-grained grants; implement explicitly.
3. Go service token authorizes internal RPC calls, not per-user tenant scope. Keep Python off public internet and use authenticated TLS/mTLS across untrusted hosts. Plaintext is restricted to local/private demo networking.
4. Provider/channel secrets remain server-side references, never frontend/Notion/repo/Protobuf snapshots/trace/checkpoint/error bodies. The three future provider keys are deployment-level configuration for MVP.
5. Real webhook source verification must match actual provider capability. Resolve tenant through connector records, bound payloads and quarantine unsupported revisions. Do not trust user-supplied tenant fields.
6. New tenant FKs, grants, context versions and final dispatch gate enforce isolation despite cache/event lag. Log permission/ownership/config changes transactionally.
7. Model outputs are untrusted. Validate schema, factual sources, script eligibility, variables and allowed tools. Python tools must be narrow read-only Go APIs unless explicitly designed otherwise; no direct channel sending or arbitrary business DB writes.
8. Current broad database-owner credentials are local-only. Future VPS uses separate runtime/migrator roles, network restrictions, encrypted transport, restricted observability and tested backups. No claims of completed security audit or production readiness.

## Reliability and performance targets

Database transactions stay short, with explicit leases/version predicates. Future Kafka ingest/outbox gives durable fan-out and backpressure. AI concurrency/rate/token quotas must be tenant/provider-aware; coalesce obsolete contexts and cancel stale work. No measured throughput, latency SLO, broker partition count or VPS resource fit has been established.

Monitor run age/backlog/recovery failures, outbox oldest event, Kafka lag age/disk/ISR/DLQ, per-tenant AI budget, error rate, unknown sends and trace export failures. Separate operational event history from retained business/audit records. Redis outage must not erase new business proposals. No automatic retention purge is implemented yet.

## Verified evidence and limits

At baseline 01b2e11:

- Docker AI test target: 22 tests passed, including real PostgreSQL concurrent submit/claim, digest conflict, tenant lookup, expired lease recovery, stale lease fencing, persisted cancellation, exhausted attempts and replacement worker.
- Offline AI suite: 21 passed, one explicit-DSN DB test skipped when no test database configured.
- Go bootstrap provider-graph regression and generated RPC/smoke packages passed; go vet ./internal/bootstrap passed.
- Full Docker builds completed; web production build and TypeScript passed; five long-running services healthy; init/Atlas jobs exited 0.
- Web login page, API DB/cache health, register/login and authenticated tenant create/list smoke passed.
- Accepted AI result and same-ID retry survived an actual AI container restart.

These are historical verification results, not a fresh rerun by every documentation edit. They do not cover the unimplemented message pipeline, actual AI providers, Kafka integration, real Pancake, full order scope, browser interaction end-to-end or production deployment.

## Delivery roadmap and gates

| Slice | Current status | Completion evidence needed |
| --- | --- | --- |
| Fake AI boundary and local stack | IMPLEMENTED/VERIFIED in PR #52 branch | Recheck merge/CI before starting from main |
| Durable run jobs | IMPLEMENTED; whole-run replay | Provider/checkpoint work still separate |
| Messaging foundations and permissions | TARGET | Atlas schema, tenant/claim/grant/version tests |
| Kafka ingress/outbox/inbox | DECIDED/TARGET | Local broker dual listeners, typed events, outages/duplicates/rebalance tests |
| Go conversation dispatcher/result acceptance | TARGET | Durable deterministic run IDs, stale-output/ownership/context tests |
| Proposal/send simulator and expiry | TARGET | Full no-live-effects vertical flow, races and unknown-send handling |
| Three AI provider adapters | DECIDED/TARGET | Fake contract tests first, owner secret provisioning and selected model/budget before paid verification |
| LangGraph checkpoints and AI result outbox | TARGET | Node replay/idempotency/cost/failure tests and schema ownership |
| Self-host Langfuse | DECIDED direction/TARGET implementation | Compatible dependency stack, viewer auth, redaction/export outage tests |
| Real Pancake integration | OPEN external access | Actual entitlement/payload/auth/send/reconciliation verified |
| VPS/domain | DECIDED eventual host; OPEN sizing | Measured app load/storage, selected host/domain, HTTPS, backup restore |

Next recommended work is a messaging/Kafka ingress vertical slice with inbox/outbox and a simulator, keeping the durable Python RPC boundary. Exact task selection should check current GitHub/Notion state, PR merge status and the owner's latest instructions; do not automatically implement every target in this document.

## Definition of architecture documentation maintenance

After schema/API/RPC/event changes, update database/HTTP/gRPC/Kafka modules and the affected epic/session records. After policy changes, update product/workflows and tests. After infrastructure work, update the runbook and actual verification evidence. Mark unfinished slices accurately. Link the PR and relevant file paths; never depend on chat history alone.
