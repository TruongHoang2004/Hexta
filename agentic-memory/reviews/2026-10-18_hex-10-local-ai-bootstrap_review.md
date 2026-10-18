# Code Review: HEX-10 Local AI Bootstrap, Go/Python gRPC Boundary & Review Fixes

**Date**: 2026-10-18  
**Scope**: Working tree changes for `HEX-10` (`services/ai/`, `proto/hexta/ai/v1/service.proto`, `services/api/internal/infrastructure/ai/rpc/`, `services/api/cmd/ai-smoke/main.go`, `infrastructure/compose.demo.yml`, `infrastructure/README.demo.md`, `.gitignore`, `agentic-memory/CHANGELOG.md`) and design/plan documents (`HEX-4`, `HEX-8`, `HEX-9`, `HEX-10`, order cancellation/returns).

---

## 1. Overview

The current working tree includes:
1. **Python AI Service Bootstrap (`services/ai/`)**:
   - Simulation-only FastAPI + LangGraph service (`hexta_ai/app.py`, `hexta_ai/graph.py`) exposing `/health/live`, `/health/ready`, and Bearer-authenticated `/internal/v1/preview`.
   - Async gRPC server (`hexta_ai/rpc.py`) managed by FastAPI's `lifespan` (`AI_GRPC_ENABLED=true`, port `50051`), implementing `AIService` (`StartRun`, `GetRun`, `CancelRun`) with bounded volatile storage (`capacity=1000`), `(tenant_id, run_id)` idempotency digest checking, atomic candidate publication, and best-effort cancellation.
2. **Protobuf Contract & Generated Bindings (`proto/` & `services/api/`)**:
   - `proto/hexta/ai/v1/service.proto` and `proto/README.md` documenting reproducible `protoc` codegen.
   - Checked-in Python bindings (`services/ai/generated/`) and Go bindings (`services/api/internal/infrastructure/ai/rpc/`).
   - Go interoperability smoke test CLI (`services/api/cmd/ai-smoke/main.go`).
3. **Isolated Local Demo Compose (`infrastructure/compose.demo.yml` & `infrastructure/README.demo.md`)**:
   - Localhost-bound PostgreSQL (`127.0.0.1:55433`), Redis (`127.0.0.1:56379`), and hardened `ai` container (`127.0.0.1:58000` HTTP, `127.0.0.1:55051` gRPC).
4. **Architecture & Planning Artifacts (`agentic-memory/`)**:
   - Design and planning documents for `HEX-4`, `HEX-8`, `HEX-9`, `HEX-10`, and order cancellation/returns.

---

## 2. The Good & Verified Fixes

- **All Previously Reported Review Findings Resolved**:
  - **Whitespace Stripping (`services/ai/hexta_ai/graph.py:10, 18`)**: `Script` and `PreviewRequest` now set `ConfigDict(extra="forbid", str_strip_whitespace=True)`, rejecting whitespace-only strings (`" \t\n\u2003"`) with HTTP `422` and gRPC `INVALID_ARGUMENT`. Covered by parameterized tests in `test_preview.py:65-74` and `test_rpc.py:126-136`.
  - **Atomic Protobuf Candidate Publication (`services/ai/hexta_ai/rpc.py:80-89`)**: `AIService.execute` builds a separate `completed = pb.Run()` via `completed.CopyFrom(run)` and replaces `self.runs[key]` only after all candidate fields serialize, preventing partial state exposure on mid-assignment errors. Covered by `test_serialization_failure_does_not_publish_partial_candidate` (`test_rpc.py:139-156`).
  - **Cancellation Preservation Under Exception (`services/ai/hexta_ai/rpc.py:92-94`)**: `except Exception:` preserves `RUN_STATUS_CANCELLED` if an adapter raises a non-cancellation exception after `CancelRun`.
  - **Graph Step State Synchronization (`services/ai/hexta_ai/graph.py:81-83`)**: `validate_output` now updates both top-level `state["steps"]` and `state["result"].steps`, verified by `test_graph_steps_stay_consistent_after_validation` (`test_preview.py:77-83`).
  - **Single Task Cleanup Path (`services/ai/hexta_ai/rpc.py:70`)**: Removed redundant `self.tasks.pop(key, None)` from `execute()`, keeping `add_done_callback` as the single task lifecycle cleanup path.
- **Strong Security & Isolation**:
  - Constant-time Bearer token validation (`secrets.compare_digest`) on both HTTP and gRPC.
  - Strict identifier regex validation (`[A-Za-z0-9_-]{1,100}`), tenant-qualified lookups, sanitized RPC errors (`SIMULATION_RUN_FAILED`), disabled hosted tracing (`tracing_context(enabled=False)`), and hardened container runtime (`uid 10001`, `read_only: true`, `no-new-privileges:true`, loopback port bindings).

---

## 3. Critical Issues (Bugs & Security)

None. All previously identified validation and state-consistency issues have been resolved and covered by regression tests.

---

## 4. Minor Observations & Notes

- **Pre-existing `.gitkeep` Diff**: `graduation thesis/assets/.gitkeep` contains a trailing newline intentionally preserved as an unrelated user working-tree change (documented in `agentic-memory/CHANGELOG.md`).
- **Next Slices (Tracked by Design/Plan)**: Durable PostgreSQL run/checkpoint persistence, real LLM provider adapters (`openai`, `anthropic`, `gemini`), Go worker integration, and self-hosted Langfuse deployment remain tracked as subsequent slices under `HEX-8` / `HEX-10`.
