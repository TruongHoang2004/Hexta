---
epic: EPIC-002
session: 2026-11-03_ai-providers
status: in-review
completed: false
date: 2026-11-03
---

# Provider adapters and guarded candidate routing

## Scope and starting evidence

Owner supplied the provider/LangGraph task attachment. Started on
`codex/messaging-inbox` at e466625 with unrelated messaging, API configuration,
dependency links and AI host setup changes already present. Preserved those
changes; no branch rewrite, commit, PR/CI claim or merge verification.
Cross-epic provider scope: EPIC-006. [Design](../../../designs/2026-11-03_ai-providers_design.md).

Selected rules by responsibility: protobuf INF-004 and Go workspace INF-005/006
for regenerated bindings; outbound client INF-002/003, configuration INF-001,
authorization ownership, mapper purity, safe errors/logging, and QA-001/006 for
Python equivalents. Go-specific DTO/GORM/Zap conventions do not impose those Go
packages on Python. Atlas ownership/test-isolation rules applied to existing
migration-backed tests; no physical schema changes or new migration required.

## Outcome and decisions

- IMPLEMENTED official async Gemini/OpenAI/Anthropic structured adapters plus fake,
  factory, provider-specific credentials, timeout/output/input budgets, usage and
  sanitized errors. SDK retries disabled; no automatic provider failover on errors.
  As explicitly requested, a missing credential selects fake and health says so.
- IMPLEMENTED validate/prepare/classify → script/generate/clarify/human → validate
  graph. Confidence/active IDs, generation permission, bounded facts/history,
  unique IDs, revisions, templates and heuristic email/number redaction enforced.
- DECIDED conservative extractive grounding: generated text must equal cited
  authorized facts joined verbatim; unsupported claims require human review.
  Streaming port buffers a full structured candidate; no token-streaming RPC.
- IMPLEMENTED backward-compatible protobuf history/facts/constraints/generation
  fields and provider/model/confidence/usage result fields. Regenerated Python/Go
  with documented tools; run_mapper owns pure aggregate boundary conversions.
- IMPLEMENTED durable paid-call fence using existing error_code column, without
  startup DDL: commit LLM_CALL_STARTED before live inference; recovered uncertain
  work fails LLM_RECOVERY_UNCERTAIN, persisted output is reused, model/provider
  mismatch fails safely. Release retains marker; SDK retry remains disabled.
  This is not exactly-once billing or persistent LangGraph node checkpoints.
- Updated ignored local .env by adding only missing settings, preserving credentials;
  replaced example credentials with an explicitly synthetic service token. README
  and smoke document offline defaults and explicit --live. Existing host Makefile
  and __init__.py setup were preserved, not reimplemented.

## Verification and evidence

- `AI_TEST_DATABASE_URL=<isolated target> uv run --frozen pytest -q --tb=short`:
  70 passed, no skips (including real PostgreSQL/Atlas adoption/history tests,
  durable gRPC metadata/digest, paid marker transaction visibility/fencing/recovery,
  SDK HTTP transport mocks and smoke refusing unapproved live inference).
- A subsequent conservative token-estimate refinement passed all 35 affected
  provider/routing/smoke tests; compileall and scoped git diff --check passed.
- Earlier no-DSN `uv run --frozen pytest`: 58 passed/6 expected database skips
  before the final truncation/smoke additions. No skipped DB test is claimed as
  evidence; the later isolated DB suite ran all cases.
- Official SDKs locked: google-genai 2.34.0, openai 2.54.0, anthropic 1.17.0.
  Actual SDK transport tests caught Anthropic's httpx2 requirement and removal of
  temperature from messages.parse; adapter/tests now match the locked API.
- `go test ./services/api/cmd/ai-smoke ./services/api/internal/infrastructure/ai/...`
  and corresponding AI package `go vet`: passed (compile checks, no Go test cases).
- Generators: protoc 34.1, protoc-gen-go 1.36.11, protoc-gen-go-grpc 1.6.1;
  Python uses locked grpcio-tools. No generated binding was hand-edited.
- PostgreSQL test container used a random loopback port and Atlas AI history;
  only disposable test databases were migrated. Application DB/services untouched.
  Test container removed at session end. No paid/live model calls or secret output.

## Remaining work and handoff

Local/unmerged, parent epics incomplete. Live valid-key/model/preview acceptance
is NOT VERIFIED; run `uv run python smoke.py --live` explicitly with privately
configured provider credentials. Offline SDK tests establish payload compatibility,
not external account/model access or response quality. Review/merge required.

Drain old worker binaries before enabling live providers; older code cannot honor
the paid marker. Node checkpoints, reconciliation of uncertain charges, calibrated
routing evaluation, tenant quotas, richer factual generation, Langfuse and Go
messaging integration remain open. Heuristic redaction is not universal PII
protection; an authorized snapshot must exclude secrets before provider export.

## Owner-directed follow-up

The direct SDK implementation described above is superseded by the
[LangChain follow-up](2026-11-03_langchain-providers.md). Earlier verification
remains historical; current shared integration has its own 73-test evidence.
