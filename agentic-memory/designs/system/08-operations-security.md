# Operations, security, verification and delivery roadmap

## Connector direction update — 2026-11-03

DECIDED: direct Facebook replaces Pancake; a separate machine-facing connector receives webhooks or fetches platform messages and normalizes them into Kafka, with abstraction before plugins. The owner REJECTED the first connector design. The [Omni Bridge-based redesign](../2026-11-03_connector-omni-bridge_design.md) is a discussion proposal, not approved implementation authority. Its base/plugin interfaces and direct Kafka ACK delivery choice require review; the rejected receipt/outbox architecture is not mandatory. Earlier Pancake diagrams are historical targets, not the selected integration. API retains business authority. No connector runtime is implemented or verified.

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

Published host ports are loopback only. Docker-internal names are postgres/redis/api/web, not host localhost. Local API config mounts `infrastructure/local/api.yaml`; legacy Go Gemini URL is blocked on loopback, OAuth unconfigured. Compose remains explicitly fake; host Python also supports configured Gemini/OpenAI/Anthropic adapters (local/unmerged provider extension, 2026-11-03). No provider key or customer send is used by smoke commands.

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

### Load capacity targets (TARGET, owner-set 2026-11-18)

The owner wants an impressive but physically feasible goal on the two existing servers, without adding hardware (owner decision, 2026-11-18):

- Infrastructure host, 2 vCPU / 4 GB: PostgreSQL, Kafka (Confluent `cp-kafka`, single-node KRaft), Redis.
- Application host, 2 vCPU / 16 GB: Caddy, `api`, `cmd/ingress`, `connector`, `ai`, `web`.
- The hosts connect over a private network; PostgreSQL and Kafka are not exposed publicly.

The current `infrastructure/compose.prod.yml` still deploys everything on one host; the split is TARGET.

Broker (owner decision, 2026-11-18): Confluent `confluentinc/cp-kafka:7.6.1` is the only Kafka broker, in production, the connector test fixture and the local stack. Single-node KRaft, topic auto-creation disabled, `log.flush.interval.messages=1` so an acknowledged batch is fsynced before the relay marks it published (a single broker has no replica to recover from). The flush applies per appended batch, so batched producers amortize it.

Baseline and ceiling (estimates unless stated):

- Baseline ~1 msg/s through Kafka, MEASURED locally on 2026-11-18 (Docker Desktop, `cp-kafka:7.6.1` with per-message flush): `kafkatransport.NewTopicPublisher` takes ~1,013 ms per synchronous single-message write because `BatchTimeout` is unset and kafka-go defaults it to 1 s; with `BatchTimeout = 1ms` the same writer takes ~4.2 ms (238 msg/s). Not yet measured on the VPS.
- Estimated physical ceiling is PostgreSQL on the infrastructure host (2 vCPU shared with Kafka): ~5,000–8,000 row writes/s, i.e. ~600–1,000 msg/s at ~9 row writes per message with batching. Kafka and the application host are not expected to bind first; VPS disk fsync latency is the largest unknown.
- Before load testing, calibrate the ceiling on the real hosts with `fio` (fsync latency), `pgbench` and `kafka-producer-perf-test`; keep the target near 50–60% of the measured ceiling and lower it if the hardware is slower than estimated.

Pass criteria for every scenario also include: no container OOM-killed on either host (current limits for PostgreSQL 2g, Kafka 1536m with 768m heap and Redis 256m already sum to ~3.8 GB). Measurement uses AI `fake` mode and simulated Meta/Telegram providers, so provider latency and platform quotas are excluded. Nothing below is measured on the target hosts yet.

| ID | Scenario | Pass criteria |
| --- | --- | --- |
| T1 | Sustained ingress 300 msg/s for 30 min (stretch 500 msg/s) | Webhook → Desk visible p95 < 3 s; zero lost, zero duplicate messages |
| T2 | Burst 1,000 msg/s for 60 s (60,000 events) | Zero webhook 5xx; backlog drained ≤ 5 min after the burst |
| T3 | Outbound 100 sends/s; mock provider 300 ms, one Page hitting the 15 s timeout | Other tenants' send p95 < 3 s (no cross-tenant head-of-line blocking) |
| T4 | Staff API 500 req/s, 80% reads / 20% writes | Read p95 < 300 ms, write p95 < 500 ms; 5xx < 0.1% (`409` version conflicts excluded) |
| T5 | Kafka stopped 5 min during T1 | Webhooks keep returning `200`; after recovery all messages reach Desk, none lost or duplicated |

Known blockers in current code (2026-11-18), each a proposal rather than an approved mechanism:

- `packages/shared/kafkatransport` leaves kafka-go `BatchTimeout` at its 1 s default, capping every synchronous relay at ~1 event/s. Direction: set an explicit small `BatchTimeout` and publish batches (T1–T3, T5).
- Connector ingress relay and outbound-result relay lease one event per transaction (`LIMIT 1`) and publish synchronously with RequireAll; the API outbox relay calls `RelayOne` per iteration. `cmd/ingress` applies one message per transaction per topic. Direction: batched leases/publishes and batched consumer transactions that still commit offsets only after the DB commit (T1, T2, T5).
- Each webhook counts pending events (bounded by the remaining room under `maxPendingEvents = 100000`), so acceptance cost grows with the backlog during a burst; 100,000 events also cover only ~5.5 min of Kafka outage at 300 msg/s. Direction: a maintained counter instead of a scan, and a cap sized from the disk budget (T2, T5).
- Connector `consumeIntents` sends one intent at a time for all tenants. Direction: partition- or connection-parallel dispatch with per-connection concurrency limits (T3).
- Single-host Compose. Direction: split into infrastructure and application stacks, re-budget memory within 4 GB (PostgreSQL buffers, Kafka JVM heap, Redis `maxmemory`), and bind PostgreSQL/Kafka to the private network with Kafka advertised on the private address (T1–T5).
- Authorization stays DB-authoritative per request; T4 is addressed by query/index work, not by caching grants.
- The Python durable worker processes runs serially and is bounded by the LLM provider; AI throughput is excluded from T1–T5 and needs its own queue-age target.

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

Messaging permissions/context model and final Go send gate; Kafka producer/consumer/outbox/inbox delivery and replay safety; operational provider integration and tenant cost budgets; LangGraph node checkpoints/result outbox; self-host trace auth/redaction/export failure; actual Pancake auth/webhook/send capability; measured VPS sizing, HTTPS and tested backup restore remain unimplemented or require explicit evidence. Owner's active demo/thesis roadmap determines sequencing; this is not an automatic instruction to build Kafka next.

For schema/API/RPC/event/policy changes, update the owning module and primary epic/session with actual checks/limits. Operational run commands and volume/backup safety above remain applicable; inspect current Compose/ports before using a historical endpoint snapshot.
