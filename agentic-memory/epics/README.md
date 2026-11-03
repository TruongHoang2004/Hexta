# Epic index

Updated 2026-11-03. Navigation/status only: read the relevant epic and latest relevant session, not all epic history. Rules are selected through [root rules](../../rules/README.md); contracts through the [system index](../designs/2026-10-18_system-design_design.md).

| Epic | Scope | Status | Completed |
| --- | --- | --- | --- |
| [EPIC-001](EPIC-001-system-design/README.md) | Architecture, rule catalog, [instruction migration](EPIC-001-system-design/sessions/2026-11-03_ai-instructions.md), thesis direction and code audit | in-progress | false |
| [EPIC-002](EPIC-002-ai-runtime/README.md) | Python provider routing and durable gRPC runs | in-progress | false |
| [EPIC-003](EPIC-003-local-docker/README.md) | Local Docker and Atlas migration ownership | in-review | false |
| [EPIC-004](EPIC-004-messaging/README.md) | Conversation domain, scripts and send policy; modular inbox preview in review | in-progress | false |
| [EPIC-005](EPIC-005-kafka/README.md) | Kafka event backbone | in-progress | false |
| [EPIC-006](EPIC-006-provider-observability/README.md) | Real providers, checkpoints and self-host traces | in-progress | false |
| [EPIC-007](EPIC-007-orders-workflows/README.md) | Sales, configurable work and owner overview | in-progress | false |
| [EPIC-008](EPIC-008-integrations-deployment/README.md) | Direct Facebook connector and eventual VPS deployment | in-progress | false |

Current delivery: [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59) verified OPEN/draft on 2026-10-28; review/merge/manual gates remain. PRs #52–55/#57/#58 verified merged into main. None of these individual deliveries completes a parent epic automatically. No CI-success assertion from this cleanup.

Current owner priority: Hexta thesis milestone 1Dec2026, one product at a time; see [roadmap](../plans/2026-10-27_thesis-first-ecosystem_plan.md). Mapper/cache policy and the root rule router remain separately labeled owner direction versus proposed mechanism.

Latest memory maintenance: [cleanup session](EPIC-001-system-design/sessions/2026-10-28_memory-pruning.md). Removed-file lookup: [archive ledger](../ARCHIVE_INDEX.md), optional historical context only. State/completion semantics and session protocol live in [project memory](../README.md), referenced by AGENTS; new sessions use [the template](_SESSION_TEMPLATE.md).

Rules/memory delivery: [PR #60](https://github.com/TruongHoang2004/Hexta/pull/60), OPEN/non-draft against main; [publication session](EPIC-001-system-design/sessions/2026-10-28_rules-memory-pr.md) in-review/completed false. Original overview PR remains separate.

Active messaging slice: [2026-11-02 simulator/manual inbox](EPIC-004-messaging/sessions/2026-11-02_messaging-inbox.md), local/unmerged; provider, AI, Kafka and full turn lifecycle remain open. Older delivery snapshots above retain their dated verification limits.

Active provider slice: [2026-11-03 AI adapters](EPIC-002-ai-runtime/sessions/2026-11-03_ai-providers.md), local/in-review; offline and isolated PostgreSQL verification passed, live provider acceptance and merge remain open.

Provider follow-up: [shared LangChain integration](EPIC-002-ai-runtime/sessions/2026-11-03_langchain-providers.md) supersedes direct SDK calls; 73 isolated/offline tests passed, local/in-review with live access and merge open.

Active connector slice: [2026-11-03 direct Facebook ingress](EPIC-008-integrations-deployment/sessions/2026-11-03_connector-ingress.md), local/in-review; isolated HTTP → PostgreSQL → Kafka → API member inbox verified. Live Meta, outbound/AI and merge remain open.
