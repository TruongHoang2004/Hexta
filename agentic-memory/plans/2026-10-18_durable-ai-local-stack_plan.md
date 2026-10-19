# Durable runs and local Docker plan

1. Add PostgreSQL run repository, schema migration, durable RPC adapter and leased worker; retain offline unit-test adapter.
2. Test persisted idempotency, multi-worker claims, cancellation, restart/recovery and fencing with real PostgreSQL.
3. Add isolated local full application Compose services, API config, existing Atlas migration job, web build configuration and Docker-only smoke checks.
4. Run the documented local startup and test sequence, correct actual failures, document limits and data-safe shutdown/reset.
5. Update changelog, commit on codex/durable-ai-local-stack, push and create a PR to main.

Done: tests and actual Docker startup pass; accepted runs survive service replacement; no secret/provider/customer calls; runnable docs and linked PR. Full messaging/AI provider integration remains outside this persistence slice.

## Completed verification

22 tests passed in the Docker AI test target, including real PostgreSQL recovery/fencing. Five application/dependency services healthy; init/Atlas jobs exit 0. Web/API account/tenant smoke and AI restart durability passed. Fixed the existing Go IDraftCache Fx binding discovered by full-stack startup. Added owner-requested Kafka use-case draft; implementation is a later slice. PR creation follows these checks.
