---
epic: EPIC-006
title: Real AI providers, checkpoints and self-host traces
status: in-progress
completed: false
updated: 2026-11-03
---

# Real AI providers, checkpoints and self-host traces

## Scope

Support three provider adapters and observable safe AI flows without moving business authority into Python.

## Decisions and context

- Owner will provide OpenAI/Anthropic/Gemini server keys later; one provider/model per run, no automatic cross-provider error fallback and no required per-tenant key UI.
- MVP AI stays simple; no multi-agent swarm or mandatory vector store. Router evaluation should precede deciding scripts versus generation.
- Observability is self-hosted; Langfuse is proposed. Self-host traces/application does not require self-host model inference or GPU.

## Acceptance checklist

- [x] Local provider adapters/structured normalization and offline SDK/fake tests; review/merge and live validation remain open.
- [ ] Explicit selected model/budget/secret provisioning before paid verification.
- [ ] Node checkpoint/replay and cost/idempotency policy.
- [ ] Self-host Langfuse dependencies/exporter/viewer auth/redaction and trace-loss tests.

## Design and implementation references

- [05-grpc-ai.md](../../designs/system/05-grpc-ai.md)
- [08-operations-security.md](../../designs/system/08-operations-security.md)

## Evidence and verification boundary

Provider implementation/offline verification is recorded in the shared EPIC-002 session below; live access and merge remain open. Checkpoints/Langfuse are not implemented.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-11-03_langchain-providers](../EPIC-002-ai-runtime/sessions/2026-11-03_langchain-providers.md) — shared primary follow-up; direct adapters replaced by LangChain, 73 tests pass, local/in-review.

- [2026-11-03_ai-providers](../EPIC-002-ai-runtime/sessions/2026-11-03_ai-providers.md) — primary shared session; local/in-review, completed false.

## Remaining work and next handoff

Establish provider/model/budget and observability sizing only when this slice is selected. No keys currently supplied; never ask for keys in repository/chat logs.

Owner task (2026-11-03) explicitly permits fake fallback when a selected provider has no credential; it never retries another paid provider after an error.
