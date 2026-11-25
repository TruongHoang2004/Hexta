# Hexta AI candidate runtime

The shared LangGraph supports deterministic fake routing and LangChain Gemini, OpenAI and Anthropic integrations. It returns candidate proposals only. Go retains tenant authorization, current-context acceptance and all customer sending. No channel connectors, arbitrary tools or business mutations run in Python. Hosted graph tracing stays disabled.

## Run without Docker

Install Python 3.12–3.14 and uv. From this directory:

```sh
# Copy environment file (if not already present)
cp -n .env.example .env

# Install dependencies
uv sync --frozen

# Apply Atlas AI migrations from the repository root before database-backed startup.
# Or set AI_GRPC_ENABLED=false in .env for HTTP-only preview.
# Start the dev server (reads .env automatically)
make dev
```


The token is a local service credential, not an LLM API key. Health endpoints are public and contain no secrets. Preview requires `Authorization: Bearer <AI_INTERNAL_TOKEN>` and a JSON body such as:

```json
{
  "message": "price",
  "scripts": [
    {"id": "example", "trigger": "price", "text": "Please choose a product.", "active": true}
  ]
}
```

POST it to `/internal/v1/preview`. Unknown or ambiguous inputs and unresolved template variables return a human result with no fabricated answer. This endpoint is an internal preview, not the gRPC run contract; it trusts the authorized caller's script snapshot and is not exposed to sales browsers.

## Provider configuration and routing

Set provider keys privately in ignored `.env`; never paste them into shell history.
Existing process environment takes precedence. Select one provider:

| AI_PROVIDER | Credential (preferred over AI_API_KEY) | AI_MODEL example |
| --- | --- | --- |
| disabled | None | (none) |
| fake | None | deterministic |
| gemini | GEMINI_API_KEY | gemini-2.5-flash |
| openai | OPENAI_API_KEY | gpt-4o-mini |
| anthropic | ANTHROPIC_API_KEY | claude-sonnet-5-5 |

Set `AI_MODEL` explicitly for a real provider; these are configurable examples, not a
claim that every model supports structured output. `disabled` (the production default) proposes only a unique
exact match of an active script and hands everything else to staff, with no router and no paid call. Startup is
fail-safe (owner decision 2026-11-26): a real provider without its key or `AI_MODEL`, a model of another provider
family (`anthropic` needs `claude-*`, `gemini` needs `gemini-*`, `openai` needs `gpt-*`, `chatgpt-*` or
`o1`/`o3`/`o4*`), invalid numeric settings or an unknown `AI_PROVIDER` select `disabled` and log
`AI disabled: <reason>` instead of failing startup. Jev and Gemini embeddings run only with a real provider: while
AI is disabled or `AI_PROVIDER=fake` (offline, simulated; fake answers only through `AI_FAKE_ROUTER`) they stay
off whatever keys are set. Invalid `AI_RAG_*` values disable knowledge retrieval with an error log;
`AI_JEV_ENABLED=true` without `TYPESAFE_API_KEY` and Gemini embeddings without a key are logged as errors and fall
back to Jev off / offline hashing. One `AI runtime …` line reports
the effective provider, model, router and embedding model; health identifies the effective provider. Invalid keys
fail calls safely without cross-provider failover.
`AI_TIMEOUT` defaults to 20 seconds (1–30); `AI_MAX_TOKENS` to 2048 (64–8192);
`AI_MAX_INPUT_TOKENS` to a conservative UTF-8 byte-based token estimate of 16000 (256–64000).
`AI_TEMPERATURE` defaults to 0.7 for Gemini/OpenAI; the Anthropic integration
uses its default sampling settings. Underlying client retries are disabled (Google integration uses one total attempt). Provider-reported
usage is returned in `usage`; missing usage is estimated and labeled explicitly.

The graph validates and sanitizes input, prepares bounded/redacted context,
classifies intent, branches to an active script / grounded candidate / clarification
/ human, and validates output. `AI_CONFIDENCE_THRESHOLD` defaults to 0.8.
Inactive or nonexistent script IDs and low confidence require human review.
`steps` contains node names only, never hidden model reasoning.

For generation, the trusted Go caller must supply `allow_generation: true` and
`facts: [{"id": "hours", "text": "The store is open 9 AM to 5 PM."}]`.
Optional `history` accepts up to 20 user/assistant messages; `constraints` accepts
up to 20 strings. Scripts/facts have unique IDs and bounded lengths. Revisions are
nonnegative and echoed through gRPC. Old callers default to no generation.
Candidates cite facts internally and must reproduce their text exactly, joined
with newlines. Added claims, unknown/duplicate citations and unresolved template
variables yield human handoff. This is intentionally extractive grounding, not
unrestricted conversational generation. Email/number redaction is heuristic;
it cannot establish complete PII detection or consent to export domain data.

Only authenticated internal callers can use preview. Failure categories are
`LLM_PROVIDER_ERROR`, `LLM_TIMEOUT`, `LLM_INVALID_OUTPUT`, `QUOTA_EXCEEDED`,
`INPUT_BUDGET_EXCEEDED`; no provider bodies/credentials are returned or logged.
The provider streaming port buffers one complete structured candidate; the public
HTTP endpoint and RPCs remain non-streaming.

## Jev routing, knowledge answers and RAG (2026-11-26)

Design: [agentic-memory/designs/2026-11-26_ai-jev-rag_design.md](../../agentic-memory/designs/2026-11-26_ai-jev-rag_design.md).

- **Router.** With `AI_JEV_ENABLED=true` and `TYPESAFE_API_KEY`, one TypeSafe System One call (model
  `AI_JEV_MODEL`, default pinned `jev-1.14.0`) chooses `script` / `answer` / `clarify` / `human`. Only
  typed-redacted text (`[PHONE]`, `[EMAIL]`, `[ADDRESS]`, see `hexta_ai/redaction.py`) and redacted script
  excerpts are sent; script IDs become local labels. Thresholds (`AI_JEV_MIN_PROBABILITY`,
  `AI_JEV_SCRIPT_MIN_PROBABILITY`, `AI_JEV_NOUL_YES`, `AI_JEV_NOUL_NO`) are configuration because Vietnamese
  accuracy is not documented by the vendor. Jev errors, low probability or invalid output fall back to the
  deterministic exact matcher and then `human`, never to a paid LLM routing call. Jev disabled (default)
  keeps the existing classifier path. The SDK retries only HTTP 429/529 and its body-logging logger is kept
  at WARNING.
- **Knowledge.** Go pushes tenant products, documents and active scripts with `UpsertKnowledge` /
  `DeleteKnowledge` (idempotent by source revision + content hash; attribute-only changes are not
  re-embedded; deletes leave tombstones). Chunks and 768-dimension embeddings live in `ai_runtime`
  (pgvector >= 0.8.0, extension in schema `public`). `AI_EMBEDDING_PROVIDER=fake` (default, offline
  hashing) or `gemini` (`AI_EMBEDDING_MODEL`, default `gemini-embedding-001`, via langchain-google-genai).
  Retrieval is tenant- and embedding-model-filtered top-k (`AI_RAG_TOP_K`), similarity floor
  (`AI_RAG_MIN_SIMILARITY`) and context budget (`AI_RAG_MAX_CONTEXT_TOKENS`). Chunks stored under another
  embedding model are ignored until Go re-sends the source; after switching models run the API's
  `knowledge-resync` command (`make knowledge-resync all=1`, production: README.deploy.md "AI settings").
- **Answers.** The answer route needs the caller's `allow_knowledge_answer`. Claude (default
  `claude-sonnet-5-5`; any configured provider works) returns structured `kind/text/cited_source_ids/
  confidence/needs_human`, with a stable cache-marked system prompt followed by knowledge, conversation and
  message. Answers must cite retrieved sources, meet `AI_ANSWER_MIN_CONFIDENCE` (0.7) and contain only
  numbers present in the cited records or the question. Refusals and provider errors become `human`
  candidates. Generated text is labelled `candidate_kind=generated`, `requires_review=true`: Go must never
  auto-send it.
- **Offline demo router.** `AI_FAKE_ROUTER=keyword` (default unset) routes with a deterministic,
  diacritics-insensitive Vietnamese keyword router (`hexta_ai/routing/keyword.py`, router `deterministic`,
  model `keyword-v1`): unique exact trigger → `script`; staff/complaint/refund/order-problem words → `human`
  (`customer_requested_staff`); product/price/stock/size/shipping/policy words → `answer`; very short or
  vague messages → `clarify`; otherwise the deterministic fallback (`keyword_no_match`) → `human`. It makes
  no network call and keeps runs simulated. Jev (enabled, with a key and a real provider) always wins. With the fake
  provider an `answer` is an extractive Vietnamese reply built from the best retrieved record (cited, price
  and stock from its attributes), so a demo shows "AI answer pending review with citations" without paid
  calls. Offline hashing embeddings (`hashing-768-v2`, diacritic-folded words and bigrams without filler
  words) give about 0.4–0.6 cosine for matching demo questions and below 0.15 for unrelated ones: use
  `AI_RAG_MIN_SIMILARITY=0.15` with `AI_EMBEDDING_PROVIDER=fake`; keep the 0.55 default for Gemini.
- **Clarifying question.** The static clarify text (not model-authored, so no review is required) defaults to
  Vietnamese and is configurable with `AI_CLARIFY_TEXT` (1–500 characters, no `{`/`}`; invalid values fail
  startup).
- **Eval.** `uv run python -m hexta_ai.routing.evaluate` checks the Vietnamese fixture set
  (`evals/jev_routing_vi.json`) offline; `--replay` scores recorded answers and `--sweep` compares
  thresholds; `--live --confirm-paid-calls --record FILE` calls the paid API and needs owner approval.

Database prerequisite: the AI migrator is not a superuser in Compose/production, and pgvector is not a
trusted extension, so a superuser must run `CREATE EXTENSION vector WITH SCHEMA public` on a
pgvector-capable image before `make migrate svc=ai`. Generating AI migrations needs `AI_DEV_URL` pointing
to a disposable dev database with `vector` preinstalled in `public` (Atlas Community cannot manage
extensions and drops it after each diff, so reinstall before every run).

## Offline and opt-in live smoke

```sh
# Tests ignore developer provider keys and never make paid calls.
uv run --frozen pytest

# With a running fake server:
uv run python smoke.py

# With a running real provider configured privately:
# Explicitly allows up to two calls using synthetic opening-hours facts.
uv run python smoke.py --live
```

The default smoke refuses an active real provider before calling preview. Live
smoke refuses fake fallback. Each preview is a new call, not a durable idempotent
run; repeating it can incur charges. No smoke sends a customer message.

## Limits and next slices

Durable runs and one shared LangChain provider path are implemented locally. Persistent LangGraph
node checkpoints, Langfuse exporter, operational Go conversation integration and
real channel sending remain pending. This change does not certify production
readiness or prove live credentials/model access. Offline transport mocks exercise the real LangChain integrations for all
three provider request/response paths; live smoke requires a separate explicit run.

Do not put API keys in this README, request body or shell command history. Deployment secrets and self-hosted tracing will be configured separately. No API key is needed for these tests.

## Internal gRPC runs

`proto/hexta/ai/v1/service.proto` is the shared versioned contract. The Python server starts alongside HTTP (port 50051 in Docker); Go bindings live in `services/api/internal/infrastructure/ai/rpc`. RPCs are unary `StartRun`, `GetRun`, `CancelRun`. Authorization metadata is `authorization: Bearer <service token>` on every call; Go authenticates tenant business access before constructing the snapshot. This shared credential identifies the Go service, not a tenant user.

StartRun returns queued/current status without waiting for graph completion. `(tenant_id, run_id)` identifies a run; identical requests return the existing run, changed input returns ALREADY_EXISTS. Get/Cancel require the tenant. Context revision and control generation are echoed so the future Go acceptance gate can reject stale results. Cancel is best effort; completed results remain completed. It never authorizes an outbound send.

**Runtime storage is PostgreSQL**, configured by mandatory `AI_DATABASE_URL`. StartRun commits before acknowledgement. The Python-owned ai_runtime schema stores requests, digests, results, attempts and fenced renewable leases. Expired deterministic jobs can be reclaimed up to three recovery attempts; exhausted runs fail safely. Before a live graph starts, the worker commits a lease-fenced `LLM_CALL_STARTED` marker in the existing run row. Recovery after that marker fails with `LLM_RECOVERY_UNCERTAIN` instead of repeating a paid call. Graceful release retains the marker; provider/model changes fail with `PROVIDER_CONFIG_CHANGED`. Already persisted outputs are reused. Cancellation persists and prevents publication by old workers. Graceful shutdown requeues active work with these safeguards. Restart retains successful/cancelled results and idempotency history. The in-memory adapter remains solely for isolated unit tests.

Atlas applies `migrations/ai` before AI starts; `make migrate svc=ai` runs the explicit job. The service only checks storage readiness and starts its worker; it never creates schemas/tables. Root `atlas.hcl` and `migrations/schema/ai.sql` define the migration workflow. Existing compatible pre-Atlas run tables are validated and adopted without deleting runs; incompatible schemas fail migration. No automatic retention deletion is enabled yet: disk usage must be monitored. A run stores the authorized input snapshot and output, so treat DB/backups as sensitive. Only fake work is replayed as a whole graph; live work deliberately trades automatic recovery for avoiding duplicate paid requests. A provider may charge even if its response was lost; exactly-once billing and node checkpoint recovery are not claimed. Drain old workers before enabling live mode; older worker binaries do not understand paid-call markers. Runtime processes one leased run at a time; scale/quotas are future work.

The local profile uses plaintext on loopback/private Docker networking only. Do not publish gRPC publicly; configure authenticated TLS or mTLS if crossing hosts. HTTP remains for health and offline preview, while browser business APIs remain in Go.

With Compose running, from repository root:

```sh
(cd services/api && go run ./cmd/ai-smoke)
```

Use the same exported DEMO_AI_INTERNAL_TOKEN as Compose. This command checks Go/Python interoperability, rejects missing auth, retries the same run ID, polls a candidate, and never sends a customer message. See `proto/README.md` for code generation.

## Migration-backed tests

From repository root, run `./scripts/local.sh --profile test run --build --rm ai-tests`.
An Atlas test job creates the isolated AI schema before pytest. The test image includes Atlas only for disposable-database adoption tests; the production image has no migration command. Host tests with AI_TEST_DATABASE_URL additionally require Atlas on PATH and access to the repository Atlas config. No explicit database URL means database tests skip; offline tests still run. Use a dedicated test server with database-creation rights, never the demo/application URL.

## LangChain implementation

`providers/langchain.py` initializes all real models via `init_chat_model` using
`langchain-google-genai`, `langchain-openai`, or `langchain-anthropic`. Each call
uses the same `with_structured_output(..., method="json_schema", include_raw=True)`
and `ainvoke` path. LangChain normalizes messages and token usage; the application
checks parsing/completion status and maps safe errors. There are no direct
provider SDK request calls in application code. Provider SDK packages remain
transitive dependencies of the LangChain integrations. LangGraph routing and
the durable paid-call fence continue to operate above this model abstraction.

The authenticated `/internal/v1/consult` endpoint serves the LangGraph shopping
consultant. Its model adapter fails closed when a real provider lacks credentials;
offline use must explicitly select fake. While the runtime is disabled the consultant never creates a model and
returns its static fallback reply. For the
private OpenAI launch configuration, see [real consultation](../../infrastructure/README.demo.md#real-openai-consultation).
