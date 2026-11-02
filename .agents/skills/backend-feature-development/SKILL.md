---
name: backend-feature-development
description: "Implement Go API use cases or endpoint behavior in Hexta with existing layer, authorization, transaction and contract patterns."
---

# Backend feature development

Required context: expected behavior, affected API use case, relevant rule IDs and neighboring source/tests. Read [API guidance](../../../services/api/AGENTS.md) and select modules through the root router.

1. Trace route/controller/validator → service → repository interface/implementation and bootstrap wiring. Identify actual actor/tenant authority and affected consumers.
2. Define validation, transaction scope, conflict/replay behavior and safe error outcomes before changing the use case. Preserve current interface placement; draft architecture is not a refactor mandate.
3. Implement in owning layers, reusing existing helpers/mappers. Include failure-path coverage proportional to changed behavior, especially cross-tenant access, rollback, stale versions and idempotency.
4. For HTTP/RPC/schema changes use [cross-service change](../cross-service-change/SKILL.md); regenerate only affected owned artifacts with documented tools.
5. Run applicable formatting, vet, focused/race and isolated integration checks. Do not convert missing DB prerequisites into PASS or apply schema changes to the application DB during validation.

Output: implementation, affected rule/contract decisions, exact validation and unresolved acceptance. Record observed deviations separately from introduced changes; preserve E/D/O/P provenance.
