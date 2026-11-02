# Python AI service

Inspect `pyproject.toml`, `uv.lock`, relevant modules under `hexta_ai/` and neighboring `tests/`. Use uv and existing pytest workflows. FastAPI serves internal preview/health; gRPC owns durable run operations. Python proposes candidates; Go alone authorizes business actions and sending.

## Runtime boundaries

- Extend `providers/base.py` and the provider factory/adapter boundary rather than adding provider SDK calls inside graph, RPC or business logic. The current shared adapter uses LangChain; deterministic fake behavior supports offline tests.
- Keep bounded structured inputs/outputs and validation at provider and graph boundaries. User messages, scripts, history, facts and model output are untrusted data. Preserve active-script eligibility, generation opt-in, grounding, confidence and human/clarify fallback behavior.
- Inspect current prompt ownership in `providers/base.py`; do not create a parallel prompt system. Bound context and output, distinguish estimated from reported usage, and preserve timeout/error categories without leaking provider bodies or private data.
- Preserve async cancellation and resource cleanup. No DB lock may span graph/provider work. Durable acceptance follows committed storage, with tenant/run identity, request fingerprint, leases and terminal-state guards in the existing store.
- Inspect paid-call recovery before changing retries. Do not repeat potentially paid effects merely because a worker lease expired; retain uncertain-recovery handling and provider configuration checks. SDK defaults are not evidence of safe retry semantics.
- Preserve disabled hosted prompt tracing and sanitized observability. A LangSmith dependency does not authorize exporting prompts. No live-provider calls are implied by tests.
- RAG, channel tools, Kafka integration and self-hosted tracing require evidence/approved scope; backing service configuration alone does not establish an implemented workflow.

## Contracts and checks

For RPC changes read [protobuf instructions](../../proto/AGENTS.md) and regenerate affected Go/Python bindings together. For persistent schema changes use [Atlas instructions](../../migrations/AGENTS.md); startup readiness only checks schema and never creates it.

Use `uv run pytest` from this directory for applicable offline tests; inspect fixtures for required isolated DB configuration. The root CI uses the isolated Compose AI test profile. Select provider, graph, RPC, store/recovery tests according to the change; cover invalid output, budgets/timeouts, replay/conflict, cancellation and uncertain recovery where affected. Report offline, isolated DB and live-provider acceptance separately. Read the current AI architecture module only for contract/design changes, not all historical sessions.
