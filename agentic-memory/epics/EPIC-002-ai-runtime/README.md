---
epic: EPIC-002
title: Python AI runtime and durable gRPC jobs
status: in-progress
completed: false
updated: 2026-10-19
---

# Python AI runtime and durable gRPC jobs

## Scope

Deliver a separate Python execution service with durable, tenant-qualified jobs and safe interoperability with Go.

## Decisions and context

- gRPC AIService.StartRun/GetRun/CancelRun is chosen; Go authorizes snapshots and eventual result acceptance.
- Fake graph supports unique active exact-match script or human fallback. PostgreSQL commits run/digest before acknowledgement; leases fence stale completion, cancellation persists.
- Current recovery is whole-graph replay, not LangGraph node checkpoints. No direct Pancake send or business DB writes from Python.

## Acceptance checklist

- [x] Shared Protobuf, generated Go/Python bindings, auth/input validation and Go smoke exist; PR #51 merge confirmed by owner.
- [x] PostgreSQL run acceptance, digest conflict, leases/recovery/cancellation and readiness implemented on PR #52 branch.
- [x] 22 Docker AI tests including real PostgreSQL passed at baseline; actual container restart retained run/idempotency.
- [ ] Verify PR #52 merge/CI before closing its durable-run slice.
- [ ] Design/implement persistent node checkpoints and future result outbox with replay/cost safety when selected.

## Design and implementation references

- [05-grpc-ai.md](../../designs/system/05-grpc-ai.md)
- [2026-10-18_durable-ai-runs_design.md](../../designs/2026-10-18_durable-ai-runs_design.md)
- [service.proto](../../../proto/hexta/ai/v1/service.proto)

## Evidence and verification boundary

Code baseline 01b2e11; PR #51 merged (owner confirmation); PR #52 last known open/in-review, remote state not reverified in this organization task.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-10-18_durable-runs](sessions/2026-10-18_durable-runs.md) — in-review; completed: false.
- [2026-10-18_grpc-bootstrap](sessions/2026-10-18_grpc-bootstrap.md) — done; completed: true.

## Remaining work and next handoff

Recheck PR #52, then coordinate with messaging/Kafka for a Go dispatcher and result acceptance; do not substitute a smoke CLI for the business worker.

## Atlas ownership correction — 2026-10-20

Python no longer applies SQL at startup. Explicit Atlas migrations/ai and the migrate-ai job provision/adopt ai_runtime before readiness/worker start. Existing compatible runs retained, incompatible schemas rejected; 27 host/Docker AI tests passed. [Primary EPIC-003 session](../EPIC-003-local-docker/sessions/2026-10-20_atlas-all-services.md) records the implementation once. No checkpoint/provider or parent-epic completion; local/unmerged.

## Legacy Go boundary repair — 2026-10-20, local/unmerged

The dormant Go AI service now uses read-only repository and domain generation/event types; Python runtime behavior, providers and messaging scope are unchanged. Atlas-backed isolation/limits and Fx/transport tests pass. [Shared EPIC-001 session](../EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md) records the implementation; no parent-epic completion claim.
