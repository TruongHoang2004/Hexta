---
name: frontend-feature-development
description: "Implement a multi-step Hexta frontend feature using existing components, app data/state patterns and scoped validation; not routine copy-only edits."
---

# Frontend feature development

Required context: target app, acceptance behavior, applicable app instructions and neighboring implementation. An issue is optional.

1. Establish affected routes, components and data contracts. Read [frontend guidance](../../../apps/AGENTS.md) and target app instructions.
2. Use [component discovery](../frontend-component-discovery/SKILL.md), or reuse its completed result. Inspect existing state ownership and service wrappers before adding another pattern.
3. Implement the smallest complete user flow, including loading, empty, error, cancellation/conflict and success states as applicable. Preserve accessible interaction, responsive layout and existing tokens.
4. If an API/export contract changes, select [cross-service change](../cross-service-change/SKILL.md); otherwise do not load backend catalogs.
5. Run target-app checks and meaningful behavioral validation. Report browser/manual coverage separately; package scripts and test harnesses must actually exist.

Output: scoped feature, reuse choices, changed contracts and evidence/remaining acceptance gaps. Follow the shared memory protocol once for meaningful work; do not generate separate plan/design/review artifacts that repeat the same change.
