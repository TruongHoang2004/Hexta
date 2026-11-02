---
name: ai-feature-development
description: "Implement Hexta Python provider, graph, durable-run or AI RPC behavior with offline fixtures and explicit recovery and contract validation."
---

# AI feature development

Required context: intended candidate/runtime behavior, current AI source, relevant contract and tests. Read [AI guidance](../../../services/ai/AGENTS.md).

1. Locate ownership in provider, graph, RPC, durable service, mapper or store. Trace input/output and persisted run identities; distinguish implemented behavior from planned RAG/tools/events.
2. Specify structured-output, grounding/opt-in, budget/timeout and failure behavior. For durable or provider changes, inspect cancellation, replay and paid-call uncertainty before proposing retries.
3. Extend existing abstractions using fake/injected provider fixtures first. Preserve Go business authority, redaction and disabled hosted prompt tracing.
4. RPC/schema changes select [cross-service change](../cross-service-change/SKILL.md). Preserve source/generated binding parity and service-owned Atlas history.
5. Run relevant pytest tests and explicitly isolated DB cases. Report fake/offline, DB, and live-provider evidence separately; live paid calls need task authorization and are not implied by test completion.

Output: scoped runtime change, failure/recovery decisions, contract impact and evidence. Never claim sending, production acceptance or parent-epic completion from candidate generation tests.
