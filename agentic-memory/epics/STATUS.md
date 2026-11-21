# Epic delivery status log

Moved verbatim from the [epic index](README.md) on 2026-11-21 so the index stays a small navigation table. Dated, point-in-time evidence: recheck code, PRs and external state before trusting it. Append new delivery notes to the owning epic or session first; add a line here only for cross-epic status.

Current delivery: [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59) verified OPEN/draft on 2026-10-28; review/merge/manual gates remain. PRs #52–55/#57/#58 verified merged into main. None of these individual deliveries completes a parent epic automatically. No CI-success assertion from this cleanup.

Current owner priority: Hexta thesis milestone 1Dec2026, one product at a time; see [roadmap](../plans/2026-10-27_thesis-first-ecosystem_plan.md). Mapper/cache policy and the root rule router remain separately labeled owner direction versus proposed mechanism.

Latest memory maintenance: [cleanup session](EPIC-001-system-design/sessions/2026-10-28_memory-pruning.md). Removed-file lookup: [archive ledger](../ARCHIVE_INDEX.md), optional historical context only. State/completion semantics and session protocol live in [project memory](../README.md), referenced by AGENTS; new sessions use [the template](_SESSION_TEMPLATE.md).

Rules/memory delivery: [PR #60](https://github.com/TruongHoang2004/Hexta/pull/60), OPEN/non-draft against main; [publication session](EPIC-001-system-design/sessions/2026-10-28_rules-memory-pr.md) in-review/completed false. Original overview PR remains separate.

Active messaging slice: [2026-11-02 simulator/manual inbox](EPIC-004-messaging/sessions/2026-11-02_messaging-inbox.md), local/unmerged; provider, AI, Kafka and full turn lifecycle remain open. Older delivery snapshots above retain their dated verification limits.

Active provider slice: [2026-11-03 AI adapters](EPIC-002-ai-runtime/sessions/2026-11-03_ai-providers.md), local/in-review; offline and isolated PostgreSQL verification passed, live provider acceptance and merge remain open.

Provider follow-up: [shared LangChain integration](EPIC-002-ai-runtime/sessions/2026-11-03_langchain-providers.md) supersedes direct SDK calls; 73 isolated/offline tests passed, local/in-review with live access and merge open.

Shopping consultant slice: [Go SSE + web widget/checkout](EPIC-002-ai-runtime/sessions/2026-11-03_shopping-consultant.md), [PR #62](https://github.com/TruongHoang2004/Hexta/pull/62) open/in-review; live Python consult endpoint, browser acceptance and merge remain open.
Active connector slice: [2026-11-03 direct Facebook ingress](EPIC-008-integrations-deployment/sessions/2026-11-03_connector-ingress.md), [PR #66](https://github.com/TruongHoang2004/Hexta/pull/66), OPEN/non-draft/in-review; isolated HTTP → PostgreSQL → Kafka → API member inbox verified. Live Meta, outbound/AI and merge remain open.

Facebook platform connection: [2026-11-15 session](EPIC-008-integrations-deployment/sessions/2026-11-15_facebook-platform-connection.md), [PR #69](https://github.com/TruongHoang2004/Hexta/pull/69) open, not merged; owner OAuth/Page management, media placeholders and single-attempt sends verified against mock Meta in isolation. Live Meta and merge open.

Telegram Bot connection: [2026-11-16 session](EPIC-008-integrations-deployment/sessions/2026-11-16_telegram-bot-connection.md), [PR #71](https://github.com/TruongHoang2004/Hexta/pull/71) merged into `main` (`cec936c`); owner Bot Token connect/rotate, private-chat ingress and single-attempt sends verified against a mock Bot API in isolation. Live Telegram webhook acceptance open.

Active Desk slice: [2026-11-10 Chatbot / Desk control](EPIC-004-messaging/sessions/2026-11-10_chatbot-desk-control.md), local/unmerged; durable queue and explicit control transitions implemented. AI routing/dispatch, live channels and granular Desk entitlements remain open.

Audit hardening: [2026-11-14 session](EPIC-001-system-design/sessions/2026-11-14_audit-hardening.md), [PR #68](https://github.com/TruongHoang2004/Hexta/pull/68) OPEN/in-review; eight fixes with isolated DB/Kafka verification, merge open.

Real consultation: [2026-11-11 runtime connection](EPIC-002-ai-runtime/sessions/2026-11-11_real-consultant.md), local/unmerged; API key/live inference acceptance pending.
