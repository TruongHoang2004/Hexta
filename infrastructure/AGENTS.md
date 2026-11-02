# Infrastructure and local operations

Use [the current demo guide](README.demo.md), `compose.demo.yml` and `scripts/local.sh` at repository root for the current complete local stack. `README.md` describes separate legacy Compose configurations. Inspect the actual profile and includes before choosing commands; do not mix stacks by assumption.

- Preserve startup ordering, health/readiness, network scope, ports, image/build provenance and persistent volume ownership. Review changed configuration with its actual application consumer.
- Use private environment configuration and sanitized examples. Do not print resolved secret-bearing Compose configuration, DSNs or tokens to logs/artifacts. Validate configuration with quiet output where supported.
- Database initialization is distinct from application schema migration. Select [Atlas instructions](../migrations/AGENTS.md) for migrator jobs, root Atlas config or fixture provisioning. Readiness must not mutate schema.
- Docker, Compose and existing GitHub Actions are the current tools; no Kubernetes/Terraform platform is implied. Kafka/Qdrant/Elasticsearch or telemetry configuration alone does not prove application integration or production acceptance.
- Review operational impact and rollback before changing exposure, health dependencies, observability or storage. Preserve redaction and bounded metric labels; do not enable hosted prompt tracing implicitly.
- Do not use destructive volume cleanup, application DB migration, deployment or live provider requests as routine checks. Keep tests on explicit isolated profiles. A successful Compose syntax check is not a runtime acceptance test.
- Select delivery rules when editing build/publish workflows; infrastructure guidance also applies to root/service Dockerfiles, root Compose lifecycle targets and `scripts/local.sh`.

Validate affected Compose configuration without exposing secrets, then run only authorized scoped runtime checks. Record image/network/tool prerequisites, rollback concerns and any NOT_RUN checks.
