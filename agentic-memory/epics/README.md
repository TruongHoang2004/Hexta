# Epic index

Updated 2026-10-28. Navigation/status only: read the relevant epic and latest relevant session, not all epic history. Rules are selected through [root rules](../../rules/README.md); contracts through the [system index](../designs/2026-10-18_system-design_design.md).

| Epic | Scope | Status | Completed |
| --- | --- | --- | --- |
| [EPIC-001](EPIC-001-system-design/README.md) | Architecture, rule catalog, thesis direction and current code audit | in-progress | false |
| [EPIC-002](EPIC-002-ai-runtime/README.md) | Python fake AI and durable gRPC runs | in-progress | false |
| [EPIC-003](EPIC-003-local-docker/README.md) | Local Docker and Atlas migration ownership | in-review | false |
| [EPIC-004](EPIC-004-messaging/README.md) | Conversation domain, scripts and send policy | not-started | false |
| [EPIC-005](EPIC-005-kafka/README.md) | Kafka event backbone | not-started | false |
| [EPIC-006](EPIC-006-provider-observability/README.md) | Real providers, checkpoints and self-host traces | not-started | false |
| [EPIC-007](EPIC-007-orders-workflows/README.md) | Sales, configurable work and owner overview | in-progress | false |
| [EPIC-008](EPIC-008-integrations-deployment/README.md) | Pancake connector and eventual VPS deployment | not-started | false |

Current delivery: [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59) verified OPEN/draft on 2026-10-28; review/merge/manual gates remain. PRs #52–55/#57/#58 verified merged into main. None of these individual deliveries completes a parent epic automatically. No CI-success assertion from this cleanup.

Current owner priority: Hexta thesis milestone 1Dec2026, one product at a time; see [roadmap](../plans/2026-10-27_thesis-first-ecosystem_plan.md). Mapper/cache policy and the root rule router remain separately labeled owner direction versus proposed mechanism.

Latest memory maintenance: [cleanup session](EPIC-001-system-design/sessions/2026-10-28_memory-pruning.md). Removed-file lookup: [archive ledger](../ARCHIVE_INDEX.md), optional historical context only. State/completion semantics and session protocol live in AGENTS; new sessions use [the template](_SESSION_TEMPLATE.md).

Rules/memory delivery: [publication session](EPIC-001-system-design/sessions/2026-10-28_rules-memory-pr.md), local/in-review; original overview PR remains separate.
