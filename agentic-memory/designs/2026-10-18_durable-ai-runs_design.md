# Durable AI runs and local application stack

## Summary and boundary

Persist the existing simulation run contract before acknowledging StartRun. Go remains the business authority; this slice does not introduce a customer send path or wire the legacy AI draft API to Python. Replay runs the complete deterministic graph, not a node checkpoint. Provider adapters and LangGraph checkpoint persistence remain future slices.

## Flow and persistence

Go gRPC caller -> Python validation -> committed PostgreSQL run -> leased Python worker -> serialized candidate committed with lease fencing -> GetRun. HTTP/SSE browser APIs remain in Go. The offline memory adapter remains available for unit tests only; Compose requires a database.

Python owns `ai_runtime.runs`, separate from Atlas-managed public business tables. Versioned SQL bootstraps this new table idempotently; schema changes require new migrations. Columns: tenant_id/run_id composite PK, request bytes, SHA256 digest, initial/result protobuf bytes, status, attempts, lease_token UUID, lease_until, created_at/updated_at. Queue index on status/lease expiry. Local AI role owns only its schema; Go uses the existing demo DB owner. No Redis cache of run state.

StartRun uses INSERT ON CONFLICT then compares the immutable digest. Get/Cancel qualify tenant. Worker claims one row with FOR UPDATE SKIP LOCKED, commits immediately, and executes outside the transaction. Lease renewal and completion require matching lease_token, RUNNING state and an unexpired lease. At most three crash-recovery attempts; exhausted runs fail with a safe error. Completed runs are never reclaimed. Cancellation persists before returning and revokes publication; cooperative cancellation is an optimization. A lost lease cannot publish. Graceful shutdown requeues owned work; crash recovery waits for expiry. Provider effects cannot be exactly-once guaranteed by this mechanism.

## Contract and errors

Protobuf v1 unchanged. Invalid input => INVALID_ARGUMENT; changed digest => ALREADY_EXISTS; absent tenant/run => NOT_FOUND; database outage => UNAVAILABLE. No new public Go HTTP routes/Swagger required. Python HTTP readiness checks DB and worker. Internal shared service token is not end-user tenant authorization.

## Local Docker stack

Extend standalone compose.demo.yml with explicit AI role/schema initialization, Atlas business migrations, Go API and Next.js web. Keep public ports loopback; existing data volume retained. Config/secret values are local-only and `.env.local` ignored. Build frontend public API URL for localhost; server API origin is distinct from container DNS. Avoid copying host venv/data/secrets into build contexts.

## Verification and limits

Use real PostgreSQL integration tests for concurrent submissions/claims, restart recovery, expired lease fencing, cancellation, tenant lookup and exhausted retries. Keep 21 existing offline tests. Build and start Docker stack, apply existing Atlas migrations to a new DB, verify health and register/login/tenant path where supported. Document exact runtime capabilities; no claim that unimplemented messaging UI, provider integrations, checkpoints or Langfuse exist.

References: https://www.psycopg.org/psycopg3/docs/advanced/async.html ; https://www.postgresql.org/docs/17/sql-select.html .
