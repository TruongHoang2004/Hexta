# Provider-backed AI candidate routing

Status: implementation design, local/unmerged. Scope: EPIC-002 with provider
support relevant to EPIC-006. Go retains authorization and all business sending.

## Flow and boundaries

HTTP preview and durable gRPC share a provider-configured LangGraph:
validate input → prepare bounded/redacted context → structured classification →
script, grounded candidate, clarification or human → output validation.
One shared LangChainProvider uses Gemini/OpenAI/Anthropic model integrations;
fake remains deterministic. The owner selected LangChain in the follow-up below. No customer connectors, tools or business mutations.

Classification emits a restricted route, active script ID and confidence. Only
explicit `allow_generation` plus authorized facts enables generation. Generation
must cite fact IDs and use verbatim source text; ungrounded output becomes human
handoff. This deliberately favors verified source extracts over unverifiable
free-form factual claims. Constraints/history are bounded data, never tools or
authority. Heuristic PII redaction is defense in depth, not universal detection.

## Contracts and persistence

Add backward-compatible protobuf request fields for history, facts, constraints
and generation opt-in; result fields expose provider/model/confidence/usage.
An aggregate mapper owns request/result boundary conversion. Existing serialized
protobuf request/result columns retain digest/idempotency and metadata; no DDL.

Before a real graph starts, commit a fenced `LLM_CALL_STARTED` marker in the
existing runs.error_code. Recovery seeing that marker fails safely with
`LLM_RECOVERY_UNCERTAIN` rather than repeating inference. Successful terminal
results are persisted atomically by the existing fenced finish. Graceful release
retains the marker. SDK retries are disabled. This bounds a live run to one
classification and at most one generation call; it does not claim exactly-once
provider billing or persistent LangGraph node checkpoints. Config changes cannot
silently turn an accepted fake run into paid execution.

## Configuration and lifecycle

Default fake; explicit real provider without a key falls back to fake as requested.
Provider-specific key precedes AI_API_KEY. Invalid nonempty configuration fails
startup with safe errors. Explicit call timeout/output and input budgets apply;
no automatic cross-provider failover. Lifespan owns and closes shared SDK clients.
Streaming port buffers a complete structured candidate; no partial customer output
or streaming RPC is introduced. Estimated tokens are labeled, usage comes from SDK.

## Verification

Offline SDK mocks cover all adapters, error/timeout/quota handling and usage;
graph tests cover every branch, injection-shaped input, invalid IDs, grounding,
PII/templates and budgets. Durable tests cover recovery/fencing/cancellation and
no repeated paid calls. Optional PostgreSQL tests use an explicit disposable
Atlas-migrated target, never the application DB. No paid smoke runs automatically.
Official SDK references: [Google](https://googleapis.github.io/python-genai/),
[OpenAI](https://developers.openai.com/api/reference/python),
[Anthropic](https://platform.claude.com/docs/en/build-with-claude/structured-outputs).

## Owner correction — LangChain model abstraction

The owner explicitly selected LangChain instead of application-maintained direct
SDK calls. Supersedes the direct-SDK choice above: one LangChainProvider uses
init_chat_model, with_structured_output(method=json_schema, include_raw=True),
and ainvoke for all three integrations. Provider-specific configuration remains
limited to model selection, SDK retry semantics and resource lifecycle. Keep
LangGraph, fake behavior, contracts, grounding, timeout and paid recovery guards.
