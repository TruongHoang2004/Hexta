---
epic: EPIC-008
session: 2026-11-03_connector-ingress
status: in-review
completed: false
date: 2026-11-03
---

# Direct Facebook text ingress

## Scope and starting evidence

Branch `TruongHoang2004/connect`, initially clean. Primary EPIC-008 session; also affects EPIC-004/005. Named connector drafts were absent. Existing code had a tenant-authorized simulator inbox and no connector/application Kafka path. Only named Omni Bridge source files were read; no reference secrets/config/private dependencies were read or copied.

Selected root/router, API/shared/migration/infrastructure instructions and relevant Go/layer/service/repository/Atlas/verification/authorization/lifecycle/config/client/error/mapper/HTTP/observability modules. New connector paths use nearest Go API responsibility guidance; native machine webhook protocol responses are documented in its runbook. App/web instructions and installed Next server/client documentation selected for the existing inbox adaptation. Skills: cross-service-change, backend-feature-development, dev-changelog. No delegation, commits, external publication, normal application migration/restart or deployment.

## Outcome and decisions

DECIDED: separate Go I/O service + BasePlugin/platform capabilities + durable receipt/outbox + Kafka. Direct Facebook supersedes Pancake. Rejected historical drafts do not reject durability; direct Kafka ACK is not final authority.

IMPLEMENTED locally:

- Core plugin defaults/registry/contract fixtures before Facebook. Unsupported capabilities return errors. Facebook embeds BasePlugin and declares only Webhook=true. No Fiber/Wire/Piscale dependency. kafka-go 0.4.49 added because no application Kafka adapter existed; root workspace resolves shared without another replace.
- Raw-body HMAC, separate challenge, bounded multi-Page batch parse, trusted app/Page lookup, immutable connection identity and machine service authentication. Receipt + normalized outbox/quarantine commit before ACK. No raw payload retention. Leased/fenced Kafka relay with bounded retries/backoff; no network inside transaction.
- Shared handwritten ingress v1 JSON; API-owned binding mirror, quarantine, current-state/dedup/business transaction and offset commit afterward. Source timestamps, provider identities/reply metadata retained. Inbox marks Facebook/received; real threads cannot be changed through simulator input/reply. Business IDs reuse UUIDv7 for inbox cursor ordering; source timestamps advance activity monotonically. No AI scheduling or real sends.
- Durable sync framework/checkpoint/cancellation/retry-after, cursor-expiry and credential-revoked outcomes with fixtures. Facebook fetch, runtime sync scheduler/management and business history timeline persistence remain OPEN/unsupported. History/reconciliation never opens active turns or schedules AI.
- Trusted machine management is implemented; tenant-authorized API provisioning/control-plane reconciliation is not. Migration/schema ownership is separate. Required production role bootstrap and privilege tests remain OPEN.

See [final design](../../../designs/2026-11-03_connector-service_design.md) and [runbook](../../../../services/connector/README.md) for contracts, startup, Meta setup and failure semantics.

## Verification and evidence

Only new disposable PostgreSQL/Redpanda on localhost:55439/59092 were used. No normal application DB was migrated. Test credentials are intentionally public fixture values. The two task-created containers were removed after verification; no application containers or volumes were touched.

PASS:

- Root `make migrate-diff svc=connector name=connector_ingress`/`sync_jobs`, API `messaging_ingress`/`ingress_quarantine`/`ingress_source_metadata`, using `DOCKER_MIGRATE=atlas -o migration-db` and isolated dev URL overrides. SQL reviewed. `make migrate-validate` checks all three histories. Fresh and sequential migration application verified on fixtures.
- `atlas schema diff --from file://migrations/<api|connector> --to file://migrations/schema/<owner>.sql --dev-url <isolated URL>`: both schemas synced. No generated source manually edited.
- `API_TEST_DATABASE_URL=<isolated admin> CONNECTOR_TEST_DATABASE_URL=<migrated connector_fixture> KAFKA_TEST_BROKERS=127.0.0.1:59092 go test -race ./services/api/internal/bootstrap ./services/api/internal/core/service ./services/api/internal/repository ./services/api/internal/present/http/controller ./services/connector/... ./packages/shared/ingress ./packages/shared/kafkatransport`. API fixtures use random Atlas-managed DBs. Connector fixtures require the explicit disposable name.
- Coverage: challenge/signature, mixed multi-Page batch/malformed items, echo mapping, duplicate/overlap, disabled mapping, immutable tenant, expired/stale fencing, consumer current-state guards, history/echo suppression, DB failure no-ACK/safe errors, sync checkpoint/cancellation/mode/cursor expiry. Relay failure injection checks propagated errors and stable retry key.
- `TestConnectorFacebookToInbox`: builds/starts real connector process, authenticated machine PUT, signed duplicate webhook while Kafka is unreachable, connector kill/restart on the same DB with the broker restored, real PostgreSQL/Kafka, API consumer, tenant/member inbox read and duplicate suppression. Repeated with race detector after source/consumer changes. No real provider I/O.
- Scoped `go vet`, gofmt, `git diff --check`; `pnpm --filter web lint`; `pnpm --filter web build`; `docker compose -f infrastructure/compose.connector-test.yml config --quiet`.

NOT_RUN:

- Live Meta challenge/webhook, subscriptions, app/Page entitlement/scopes/Graph version, real fetch/credential revocation/send. Direct developer docs returned 429; Meta-owned Postman collection supports baseline prerequisites, not app-specific access.
- Remaining process-kill matrix around publication/business commit/offset (post-ACK restart verified); broker stop/restart; physical DB outage/full disk/load; actual broker-call lease expiry/rebalance; production role privilege tests. Injected failure and transactional/fencing tests do not prove all operational crash points.
- Manual browser two-profile/responsive acceptance, production-data upgrade, retention/monitoring, deployment/CI/main merge.

## Remaining work and handoff

Use private secret files, app/Page/tenant/connection mapping and controlled HTTPS callback for receive-only Meta acceptance. Do not send tokens via chat. Add tenant-authorized API provisioning before onboarding. Verify Facebook fetch rights before enabling it; finish sync scheduler/history mapping/quota policy. For chatbot replies implement API context/routing/candidate revalidation, intent outbox, current-state attempt permits and unknown-send reconciliation before connector send. No real messages were sent; main merge and parent completion are not claimed.
