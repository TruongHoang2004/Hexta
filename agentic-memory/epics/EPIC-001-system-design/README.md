---
epic: EPIC-001
title: System design and repository context
status: in-progress
completed: false
updated: 2026-10-28
---

# System design and repository context

Latest delivery: [rules/memory PR session](sessions/2026-10-28_rules-memory-pr.md), in-review/completed false; no parent completion.

## Current scope and decisions

- Canonical architecture/contracts: [system-design index](../../designs/2026-10-18_system-design_design.md). Approved instructions enter through AGENTS; focused engineering rules through [root router](../../../rules/README.md). Do not preload the catalog or old sessions.
- Owner priority: Hexta thesis milestone 1Dec2026; Hexta handles orders/business, PostPilot fanpages and ViralForge content. The rejected separate Platform/Sales product split is not approved. [Roadmap](../../plans/2026-10-27_thesis-first-ecosystem_plan.md) and [ecosystem discussion](../../designs/2026-10-27_saas-app-platform_design.md) retain scope and limits; SSO/billing/app integration remain proposals.
- Owner direction: per-aggregate mapper modules, replacement of repository cache wrappers, root rules and selective context. Current catalog has 91 IDs with provenance labels; mapper subpackage layout and shared-cache/fencing mechanism remain review proposals. [Detailed mapper/cache design](../../designs/2026-10-28_aggregate-mappers-cache_design.md) is loaded only for that work.
- The [DDD map](../../designs/2026-10-20_ddd-domain-map_design.md) is a recommendation, not implemented context isolation. Go owns authorization/sending; Python returns candidates; orders/messaging remain separate.

## Current evidence and open work

- [2026-10-28 architecture audit](../../reviews/2026-10-28_architecture-compliance_review.md): six OPEN findings covering raw DB errors, auth query context, dormant legacy order atomicity, core/ORM coupling, persistence-derived DTOs and session-cache wiring. The audit records bounded fresh test evidence; no production certification.
- PRs #52/#53 (system/AI/docs), #54 (tenant/sales) and #55 (thesis) verified merged into main; #57 (workflow/Atlas) and #58 (compliance) also verified merged on 2026-10-28. Old pending-merge text in historical ledgers describes older snapshots. No new CI or runtime test result in this cleanup.
- [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59) remains OPEN/draft. EPIC-007 owns work-unit/overview acceptance, manual gaps and delivery evidence; do not duplicate its implementation log here.

## Acceptance and next work

- [x] Modular architecture/DB/API/RPC/event/workflow/operations contracts and source inventories exist.
- [x] Earlier system-design/documentation deliveries #52/#53 merged into main.
- [ ] Review/finalize the proposed stricter rule details and mapper/cache design; deliver through review/merge when authorized.
- [ ] Address or explicitly track the six current audit gaps; do not mark fixed because rules were written.
- [ ] Keep thesis/product/manual acceptance and parent completion separate from individual delivered slices.

## Evidence to load only when relevant

- [2026-10-28 memory pruning](sessions/2026-10-28_memory-pruning.md): latest session and cleanup/traceability boundaries.
- [SaaS/thesis discovery](sessions/2026-10-27_saas-app-platform.md): owner corrections and scheduling source.
- [Compliance implementation](sessions/2026-10-20_rule-compliance-fixes.md) and [publication](sessions/2026-10-21_compliance-pr.md): exact prior test/contract evidence; #58 merge now verified.
- [Tenant-first implementation](sessions/2026-10-19_tenant-first-implementation.md): approval/invitation implementation and live privileged-test limits.
- [Thesis source/emblem correction](sessions/2026-10-19_thesis-evidence-logo.md) and [installation removal](sessions/2026-10-19_thesis-remove-installation.md): owner-required scope and compilation evidence.
- Other retained session files are topic-specific history. Removed intermediate admin/rule drafts are indexed in [ARCHIVE_INDEX](../../ARCHIVE_INDEX.md), not recreated as startup context.

Status remains in-progress/completed false. This summary replaces an accumulated navigation timeline; it does not rewrite historical session outcomes or claim new feature completion.
