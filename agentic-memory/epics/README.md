# Epic index

Latest EPIC-001 discovery: [DDD domain map](EPIC-001-system-design/sessions/2026-10-20_ddd-domain-map.md), local/unmerged recommendation; parent status unchanged.

Updated: 2026-10-20. This is navigation/status only. Read the selected epic README and its sessions for context; shared instructions are in [AGENTS.md](../../AGENTS.md) and architecture in the [system-design index](../designs/2026-10-18_system-design_design.md).

| Epic | Scope | Status | Completed |
| --- | --- | --- | --- |
| [EPIC-001](EPIC-001-system-design/README.md) | System design, tenant-first framework, demo-first backlog, branding/thesis and post-merge standards audit | in-progress | false |
| [EPIC-002](EPIC-002-ai-runtime/README.md) | Python AI runtime and durable gRPC jobs | in-progress | false |
| [EPIC-003](EPIC-003-local-docker/README.md) | Complete current local Docker runtime | in-review | false |
| [EPIC-004](EPIC-004-messaging/README.md) | Conversation domain, scripts and reply/send policy | not-started | false |
| [EPIC-005](EPIC-005-kafka/README.md) | Kafka event backbone | not-started | false |
| [EPIC-006](EPIC-006-provider-observability/README.md) | Real AI providers, checkpoints and self-host traces | not-started | false |
| [EPIC-007](EPIC-007-orders-workflows/README.md) | Orders, workspace-first demo and flexible business workflows | in-progress | false |
| [EPIC-008](EPIC-008-integrations-deployment/README.md) | Real Pancake connector and eventual VPS deployment | not-started | false |

## Status rules

- `not-started`: scoped work has not begun; design decisions may exist.
- `in-progress`: active implementation/documentation remains.
- `in-review`: scoped artifacts/checks are ready, delivery review/merge remains.
- `blocked`: identify the actual missing prerequisite in the epic/session; still not complete.
- `done`: all scoped acceptance criteria satisfied with evidence and merged delivery confirmed. Only this state permits `completed: true`.

Epic and session completion are independent. PR state is last known, not continuously synchronized; verify GitHub/CI before updating. No epic is fully done merely because one component or design session is done. These local epic IDs organize repository context and do not fabricate GitHub/Notion issue IDs or labels.

New work: create a new session from [the template](_SESSION_TEMPLATE.md) in the primary epic. Cross-epic work links that record instead of duplicating it. Preserve earlier session history and update the epic summary/index as progress changes. Keep shared architecture in its modular documents rather than copying it into every session.

Latest cross-epic audit: [2026-10-19 ticket reconciliation](EPIC-001-system-design/sessions/2026-10-19_ticket-reconciliation.md). PR #54/#55 merges verified; parent epic completion is unchanged.

Current sprint scheduling: [Sprint 2 linkage repair](EPIC-001-system-design/sessions/2026-10-19_sprint-2-repair.md), verified on Notion; epic states unchanged.

Latest bounded sales verification: [2026-10-20 browser session](EPIC-007-orders-workflows/sessions/2026-10-20_manual-browser-testing.md), nine cases PASS with screenshot evidence; remaining acceptance and epic states unchanged.

Current D4 implementation: [2026-10-20 workflow session](EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md). Local workflow/work vertical slice implemented and automated checks/build/upgrade passed; browser/review/merge remain. Epic states unchanged.

D4 browser attempt ended NOT RUN (IAB unavailable/native Chrome did not return UI state); local implementation and automated evidence are retained, manual/review/merge gates remain open.

Latest schema visualization: [2026-10-20 current-code ERD session](EPIC-001-system-design/sessions/2026-10-20_current-code-erd.md). Code snapshot af06cf4; documentation only, no epic completion.

Current migration ownership: [Atlas-only API/AI session](EPIC-003-local-docker/sessions/2026-10-20_atlas-all-services.md), local implementation verified; EPIC-003 in-progress, no automatic schema updates or parent completion.

Workflow/Atlas delivery: [PR #57](https://github.com/TruongHoang2004/Hexta/pull/57) is open against main; initial CI in progress. EPIC-003 is in-review/completed false; D4 browser acceptance and merge remain open.

Latest code/rule audit: [2026-10-20 session](EPIC-001-system-design/sessions/2026-10-20_code-rules-audit.md). PR #57 merge verified; six existing gaps OPEN. DB-dependent local checks skipped; D4 browser acceptance unchanged. Parent epic states remain unchanged.

Latest compliance implementation: [delegated fixes](EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md), local/uncommitted on codex/rules-compliance. R1–R6 implemented and verified within documented bounds; independent integration review passed after corrections. Epic states/completion unchanged.

Latest delivery follow-up: [compliance PR publication](EPIC-001-system-design/sessions/2026-10-21_compliance-pr.md), 2026-10-21. Fresh checks passed within documented limits; parent epic states unchanged.

2026-10-22 publication: [PR #58](https://github.com/TruongHoang2004/Hexta/pull/58) is open/non-draft against main; initial backend/frontend/Docker CI in progress at ab053d5, lifecycle check passed. [Publication session]( EPIC-001-system-design/sessions/2026-10-21_compliance-pr.md) is in-review/completed false. Screenshots excluded and kept local; review/merge and parent acceptance remain pending.

2026-10-23: [next demo work units](../plans/2026-10-23_demo-work-units_plan.md), [primary EPIC-007 session](EPIC-007-orders-workflows/sessions/2026-10-23_demo-work-units.md). U01–U06 cover remaining Sprint 2 acceptance; U07–U09 cover conditional D5. PR #58 verified merged into main as 656a6c4; earlier open/unmerged notes are historical. Plan remains in-review/unmerged; parent epic table states unchanged.

2026-10-23 delegated execution: [U01–U09 evidence](EPIC-007-orders-workflows/sessions/2026-10-23_work-unit-execution.md). Owner overview API/UI implemented locally on codex/demo-work-units; Go race/vet, web lint/types/date tests and final Docker build/start pass. Bounded root browser verification recovered; remaining dual-identity/operator/mobile/other manual gaps stay explicit. Review/merge pending; EPIC-007 remains in-progress/completed false.
