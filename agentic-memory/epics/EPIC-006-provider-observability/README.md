---
epic: EPIC-006
title: Real AI providers, checkpoints and self-host traces
status: not-started
completed: false
updated: 2026-10-19
---

# Real AI providers, checkpoints and self-host traces

## Scope

Support three provider adapters and observable safe AI flows without moving business authority into Python.

## Decisions and context

- Owner will provide OpenAI/Anthropic/Gemini server keys later; one provider/model per run, no automatic fallback and no required per-tenant key UI.
- MVP AI stays simple; no multi-agent swarm or mandatory vector store. Router evaluation should precede deciding scripts versus generation.
- Observability is self-hosted; Langfuse is proposed. Self-host traces/application does not require self-host model inference or GPU.

## Acceptance checklist

- [ ] Provider adapters/capability/schema normalization and fake contract tests.
- [ ] Explicit selected model/budget/secret provisioning before paid verification.
- [ ] Node checkpoint/replay and cost/idempotency policy.
- [ ] Self-host Langfuse dependencies/exporter/viewer auth/redaction and trace-loss tests.

## Design and implementation references

- [05-grpc-ai.md](../../designs/system/05-grpc-ai.md)
- [08-operations-security.md](../../designs/system/08-operations-security.md)

## Evidence and verification boundary

No implementation session for these adapters/checkpoints/Langfuse yet; Python durable whole-run execution belongs to EPIC-002.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

No work session recorded yet. Implementation has not started; do not invent historical progress.

## Remaining work and next handoff

Establish provider/model/budget and observability sizing only when this slice is selected. No keys currently supplied; never ask for keys in repository/chat logs.
