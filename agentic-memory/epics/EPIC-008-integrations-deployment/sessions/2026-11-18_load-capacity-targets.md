---
epic: EPIC-008
session: 2026-11-18_load-capacity-targets
status: in-review
completed: false
date: 2026-11-18
---

# Load capacity targets

## Scope

Record the owner-set load capacity targets (T1–T5) for the single-host stack in the system design and the thesis. Documentation only; no load test or code change.

## Starting evidence

Branch `docs/load-capacity-targets` from `main` (`4255ccd`). Before this session, [08-operations-security.md](../../../designs/system/08-operations-security.md) stated that no throughput, latency SLO or VPS fit had been established, and thesis Chapter 4 listed load testing as `NOT_RUN`.

## Outcome and decisions

- DECIDED (owner, 2026-11-18): adopt an impressive but physically feasible capacity goal on the owner's two existing servers, without adding hardware. A first draft of 10× an estimated pilot (100 msg/s) was superseded the same day by T1 300 msg/s (stretch 500), T2 1,000 msg/s × 60 s, T3 100 sends/s, T4 500 req/s, T5 unchanged. The target sits at ~30–50% of an estimated ~600–1,000 msg/s PostgreSQL ceiling and must be recalibrated with `fio`/`pgbench`/`kafka-producer-perf-test` on the real hosts.
- DECIDED (owner, 2026-11-18): the 2 vCPU / 4 GB server runs infrastructure (PostgreSQL, Kafka, Redis); the 2 vCPU / 16 GB server runs the applications. The agent recommended the reverse because current infrastructure memory limits already sum to ~3.8 GB; the owner kept this split, so re-budgeting memory on the 4 GB host is part of the challenge. The two-host deployment is TARGET; `compose.prod.yml` is still single-host.
- TARGET: scenarios T1–T5 and pass criteria recorded in [08-operations-security.md](../../../designs/system/08-operations-security.md#load-capacity-targets-target-owner-set-2026-11-18).
- DECIDED (owner, 2026-11-18): replace Redpanda with Confluent `confluentinc/cp-kafka:7.6.1` (already the local-stack image) in `compose.prod.yml`, `prod/api.yaml` and `compose.connector-test.yml`. KRaft single node, `KAFKA_AUTO_CREATE_TOPICS_ENABLE=false`, `KAFKA_LOG_FLUSH_INTERVAL_MESSAGES=1`, heap 768m within the existing 1536m limit; the `kafka-topics` job uses `kafka-topics --create --if-not-exists`. Thesis, deploy README and design module updated; historical sessions keep their Redpanda evidence.
- Thesis: Chapter 1 states the capacity objective next to the three research questions; Chapter 3 adds a non-functional requirement, `tab:load-targets` and the code bottlenecks to solve; Chapter 4 marks baseline and post-optimization measurement as `NOT_RUN`.
- Baseline finding (MEASURED locally): `kafkatransport.NewTopicPublisher` leaves kafka-go `BatchTimeout` at its 1 s default, so each synchronous single-message publish takes ~1,013 ms (~1 msg/s); `BatchTimeout = 1ms` gives ~4.2 ms (238 msg/s). Probe: throwaway program in the session scratchpad against a standalone `cp-kafka:7.6.1` container with per-message flush; container stopped afterwards. Code unchanged so the baseline can be measured on the VPS first.
- Bottleneck analysis comes from reading current code: one-event leases in the connector and API outbox relays, one message per transaction in `cmd/ingress`, serial outbound intent consumption, the per-webhook backlog count and 100,000-event cap, and the serial AI worker. Proposed fixes are directions, not approved mechanisms.
- Notion was not updated: no Notion connector is available in this session.

## Verification and evidence

- Thesis: `xelatex` ×2, exit 0, 74 pages, no undefined references; Overfull count 31 (32 before).
- Compose config: `docker compose config -q` passes for `compose.prod.yml` (placeholder env) and `compose.connector-test.yml`.
- Production broker smoke (throwaway project `hexta-kafka-smoke`, volumes removed afterwards): `kafka` healthy; `kafka-topics` exit 0 created the 5 topics with 3 partitions, `flush.messages=1`; broker reports `auto.create.topics.enable=false`, `log.flush.interval.messages=1`; ~440 MiB used of the 1.5 GiB limit at idle; after `restart kafka` the job re-ran with exit 0 and all 5 topics persisted.
- Integration suite on a fresh `compose.connector-test.yml` (cp-kafka) fixture: `go test -race -count=1 -p 1 -v ./services/connector/... ./services/api/... ./packages/shared/...` → 22/22 packages ok, 203 PASS, 0 FAIL, 1 SKIP (`TestLiveTelegramWebhookAcceptance`, needs live credentials). Includes `TestConnectorFacebookToInbox`, `TestOutboxRelayPublishesCommittedEventToKafka`, `TestConnectorFacebookOAuthSendAndDeliveryResults`, `TestConnectorTelegramConnectIngressSendAndResults`. The working tree contained the owner's uncommitted `services/connector/internal/plugin/facebook` changes (`facebook.go`, `payload.go`), so this run covers them too.
- No load test or benchmark was run. All T1–T5 values are targets; real-server deployment of the new broker is NOT_RUN.

## Remaining work and handoff

- Owner to mirror the targets in Notion (`Hexta — Tài liệu dự án` / `Hexta — Tasks & Backlog`).
- Split `compose.prod.yml` into infrastructure and application stacks over a private network, with a 4 GB memory budget for the infrastructure host.
- Calibrate the physical ceiling on the real hosts (`fio`, `pgbench`, `kafka-producer-perf-test`) and adjust T1–T5 if needed.
- Next: build a k6 load harness and simulated Meta/Telegram providers, measure the baseline on current code, then address the bottlenecks above and re-measure.
- Confirm the pilot assumption (~300 messages/day per tenant) with real customer data if available.
