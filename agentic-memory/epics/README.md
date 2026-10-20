# Epic index

Updated: 2026-10-19. This is navigation/status only. Read the selected epic README and its sessions for context; shared instructions are in [AGENTS.md](../../AGENTS.md) and architecture in the [system-design index](../designs/2026-10-18_system-design_design.md).

| Epic | Scope | Status | Completed |
| --- | --- | --- | --- |
| [EPIC-001](EPIC-001-system-design/README.md) | System design, tenant-first framework, demo-first backlog, web branding/favicon and thesis documentation (author evidence correction; UET emblem; installation appendix removed) | in-progress | false |
| [EPIC-002](EPIC-002-ai-runtime/README.md) | Python AI runtime and durable gRPC jobs | in-progress | false |
| [EPIC-003](EPIC-003-local-docker/README.md) | Complete current local Docker runtime | in-progress | false |
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

Current D4 implementation: [2026-10-20 workflow session](EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md). Local workflow/work vertical slice implemented and automated checks/build/upgrade passed; browser/review/merge remain. Epic states unchanged.

D4 browser attempt ended NOT RUN (IAB unavailable/native Chrome did not return UI state); local implementation and automated evidence are retained, manual/review/merge gates remain open.

Current migration ownership: [Atlas-only API/AI session](EPIC-003-local-docker/sessions/2026-10-20_atlas-all-services.md), local implementation verified; EPIC-003 in-progress, no automatic schema updates or parent completion.
