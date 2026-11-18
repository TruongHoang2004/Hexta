---
epic: EPIC-008
title: Direct platform connector and eventual VPS deployment
status: in-progress
completed: false
updated: 2026-11-18
---

# Direct platform connector and eventual VPS deployment

## Scope

Verify live integration capabilities and deploy only after local representative flow and sizing evidence exist.

## Decisions and context

- Owner superseded Pancake on 2026-11-03: direct Facebook through a separate Go I/O service, durable inbox/outbox and BasePlugin/platform plugins. Owner has a Page and Meta app; actual access remains unverified.
- Demo eventually runs on VPS; no host capacity/domain chosen. Measure runnable components before sizing.
- Owner 2026-11-18: two servers available — 2 vCPU / 4 GB runs infrastructure (PostgreSQL, Kafka (Confluent cp-kafka), Redis), 2 vCPU / 16 GB runs applications. The two-host split is TARGET; current Compose is single-host.
- Verify provider webhook authentication/fetch/send capabilities; do not invent delivery guarantees. The first connector design was rejected; the Omni Bridge-based replacement remains a proposal.

## Acceptance checklist

- [ ] Actual Facebook app/Page entitlement, payload/revision/auth/send/lookup behavior verified.
- [ ] Live Telegram bot (real token, public HTTPS webhook, text/media, block/revoke) verified.
- [ ] Live connector mappings/credentials, quotas and unknown-send reconciliation.
- [ ] Media storage/processing and privacy/retention requirements resolved.
- [ ] VPS/domain/resources, HTTPS/private networks, broker/trace budgets and backup restore tested.
- [ ] Load targets T1–T5 ([08-operations-security.md](../../designs/system/08-operations-security.md#load-capacity-targets-target-owner-set-2026-11-18)) measured on the two-host topology (infrastructure 2 vCPU / 4 GB, applications 2 vCPU / 16 GB): baseline and post-optimization.

## Design and implementation references

- [02-architecture.md](../../designs/system/02-architecture.md)
- [08-operations-security.md](../../designs/system/08-operations-security.md)

## Evidence and verification boundary

Pancake research 3fab0ea49f09813ca29cfec484c2a3ef. Media/retention/backup destination/SLOs remain open; HEX-42 dataset was missing in earlier context and needs a fresh Notion check.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-11-03 connector ingress](sessions/2026-11-03_connector-ingress.md) — [PR #66](https://github.com/TruongHoang2004/Hexta/pull/66), OPEN/non-draft/in-review; signed webhook through isolated Kafka to member inbox verified. No live Meta/send/merge claim.
- [2026-11-15 Facebook platform connection](sessions/2026-11-15_facebook-platform-connection.md) — [PR #69](https://github.com/TruongHoang2004/Hexta/pull/69) open, not merged: owner OAuth + Page picker, pause/resume/disconnect, media placeholders, single-attempt outbound send with result/health events. Isolated mock-Meta E2E verified; live Meta NOT_RUN.
- [2026-11-16 single-host deployment](sessions/2026-11-16_single-host-deployment.md) — production Compose stack + runbook for hexta.org (2 vCPU / 16 GB); verified locally with `DOMAIN=localhost`; real-server deployment NOT_RUN.
- [2026-11-16 Telegram Bot connection](sessions/2026-11-16_telegram-bot-connection.md) — [PR #71](https://github.com/TruongHoang2004/Hexta/pull/71) merged into `main` (`cec936c`): owner Bot Token connect/rotate, pause/resume/disconnect, private-chat webhook with secret header, media placeholders, single-attempt sendMessage with `bot_blocked`/`chat_not_found`. Isolated mock-Bot-API E2E verified; live Telegram NOT_RUN.
- [2026-11-18 load capacity targets](sessions/2026-11-18_load-capacity-targets.md) — owner-set T1–T5 targets (300 msg/s sustained, stretch 500; 1,000 msg/s burst; 100 sends/s; 500 req/s; 5 min Kafka outage; baseline ~1 msg/s) on two hosts (infrastructure 2 vCPU / 4 GB, applications 2 vCPU / 16 GB), recorded in design and thesis; measurement NOT_RUN.
- [2026-11-14 audit hardening](../EPIC-001-system-design/sessions/2026-11-14_audit-hardening.md) — [PR #68](https://github.com/TruongHoang2004/Hexta/pull/68) in-review: connector acceptance no longer uses a global advisory lock; admission limit is now approximate.

## Remaining work and next handoff

Tenant-authorized provisioning and single-attempt outbound are implemented locally (2026-11-15 session); design deviations are folded into the design and were accepted by the owner on 2026-11-15. Then verify live Meta OAuth/receipt/send after app access, add an outbox-exhaustion sweeper and credential health checks, and add provider-specific fetch only after capability confirmation. No live sends or deployment implied.
