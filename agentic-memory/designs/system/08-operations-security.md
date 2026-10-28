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

## Verification evidence and current delivery limits

Historical results below remain bounded to the named sessions, not fresh checks from memory cleanup:

| Slice | Evidence owner | Limit |
| --- | --- | --- |
| Initial durable fake AI/local stack | [AI durable session](../../epics/EPIC-002-ai-runtime/sessions/2026-10-18_durable-runs.md), [local stack](../../epics/EPIC-003-local-docker/sessions/2026-10-18_local-stack.md) | Original 22 Docker AI tests/restart smoke are historical; whole-run replay only |
| Atlas all services | [Atlas session](../../epics/EPIC-003-local-docker/sessions/2026-10-20_atlas-all-services.md) | Explicit histories, disposable tests/fresh/adoption/upgrade; #57 now verified merged |
| Auth/DTO/AI boundary compliance | [Compliance fixes](../../epics/EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md) | Historical Go race 80/no skips, SDK35 and contract checks; browser gaps retained; #58 merged |
| Sales/workflow/overview work units | [Active execution](../../epics/EPIC-007-orders-workflows/sessions/2026-10-23_work-unit-execution.md) | Automated/integrated build and bounded browser evidence; manual/approval/merge gaps remain, #59 OPEN/draft |
| Current code-rule audit | [2026-10-28 review](../../reviews/2026-10-28_architecture-compliance_review.md) | Go85 pass/34 DB skips, SDK35, Python23/4 skips, lint/types/Atlas; no fresh DB/browser/race/build certification |

PRs #52–55/#57/#58 verified merged into main on 2026-10-28. PR #59 remains OPEN/draft. Cleanup performs no runtime checks or live grants/provider calls. Do not use earlier pending-merge or pre-compliance bug statements as current state; six current audit findings remain separately OPEN. Rule texts are selected through root rules, not repeated as chronological operations notes.

## Remaining target gates

Messaging permissions/context model and final Go send gate; Kafka producer/consumer/outbox/inbox delivery and replay safety; real provider adapters and cost budgets; LangGraph node checkpoints/result outbox; self-host trace auth/redaction/export failure; actual Pancake auth/webhook/send capability; measured VPS sizing, HTTPS and tested backup restore remain unimplemented or require explicit evidence. Owner's active demo/thesis roadmap determines sequencing; this is not an automatic instruction to build Kafka next.

For schema/API/RPC/event/policy changes, update the owning module and primary epic/session with actual checks/limits. Operational run commands and volume/backup safety above remain applicable; inspect current Compose/ports before using a historical endpoint snapshot.
