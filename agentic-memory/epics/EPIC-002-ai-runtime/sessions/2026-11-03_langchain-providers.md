---
epic: EPIC-002
session: 2026-11-03_langchain-providers
status: in-review
completed: false
date: 2026-11-03
---

# Use LangChain for all real provider calls

## Scope and starting evidence

Owner explicitly requested LangChain after reviewing the direct SDK implementation.
Follow-up to [provider session](2026-11-03_ai-providers.md) on
codex/messaging-inbox/e466625; preserve its historical test evidence and unrelated
messaging, API config, host setup and newly installed design skills. Cross-epic:
EPIC-006. Reused the relevant provider/configuration/verification rule selection
and amended the [existing design](../../../designs/2026-11-03_ai-providers_design.md).

## Outcome and decisions

- Replaced three direct SDK adapter modules with one LangChainProvider. Factory
  uses init_chat_model, structured calls use with_structured_output(json_schema,
  include_raw=True) and ainvoke, and usage reads normalized AIMessage metadata.
- Added langchain and its Google/OpenAI/Anthropic integration dependencies, removed
  direct SDK requirements; underlying SDKs remain transitive dependencies. Locked
  langchain 1.4.8, google-genai integration 4.4.2, openai integration 1.7.3 and
  anthropic integration 1.7.8.
- Preserved fake behavior, timeout/token limits, output validation, disabled hosted
  tracing, safe error categories and the durable paid-call fence. Google integration
  max_retries maps to total SDK attempts, so one disables repeat requests; other
  integrations use zero retries. HTTP transport tests verify one request on errors.
- LangChain handles provider-specific request/response formatting; application
  code has no direct provider SDK imports or request calls. Client cleanup closes
  initialized integration resources without creating unused lazy clients.
- No graph/persistence/protobuf/schema/business authority change, application DB
  access, live provider calls, commits or remote PR/CI changes.

## Verification and evidence

- `uv run --frozen pytest -q --tb=short`: 67 passed, six database tests skipped
  without an explicit test DSN.
- Full suite with AI_TEST_DATABASE_URL pointing to a task-owned PostgreSQL container
  migrated explicitly with Atlas: 73 passed, no skips. Covers LangChain integration
  transport payloads/usage, malformed/truncated output, quota/timeout/no-retry,
  factory budgets, routing, preview, gRPC durable metadata/digest and paid recovery.
- compileall and scoped git diff --check passed. Verified obsolete adapter imports
  and direct SDK calls absent from application code. No new Go or protobuf change.
- PostgreSQL test container removed afterward; application services/data unchanged.
  Live provider/account/model acceptance remains NOT VERIFIED.

## Remaining work and handoff

Local/unmerged; review/merge and explicit live smoke remain open. Parent epics
remain incomplete. Existing LangGraph checkpoint, quota, production evaluation
and business integration limits from the primary provider session still apply.
