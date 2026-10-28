# System design index

Canonical architecture/contract navigation. Engineering policy is selected through [root rules](../../rules/README.md); current work through [epics](../epics/README.md). Read only the module relevant to the task. Historical snapshots and verification limits belong to retained session evidence, not repeated status timelines here.

| Topic | Module |
| --- | --- |
| Product decisions and open scope | [01-product-decisions](system/01-product-decisions.md) |
| Components, ownership and current/target boundaries | [02-architecture](system/02-architecture.md) |
| API/AI schema and desired logical model | [03-database](system/03-database.md) |
| Implemented HTTP inventory and target contracts | [04-http-api](system/04-http-api.md) |
| Protobuf, fake AI graph and durable execution | [05-grpc-ai](system/05-grpc-ai.md) |
| Kafka delivery target | [06-kafka-events](system/06-kafka-events.md) |
| Business workflows and concurrency | [07-workflows](system/07-workflows.md) |
| Runtime, security, verification limits and operations | [08-operations-security](system/08-operations-security.md) |
| Physical code/migration ERD | [09-current-erd](system/09-current-erd.md) |

Evidence labels: IMPLEMENTED = code exists; VERIFIED = a named bounded check passed; DECIDED = owner direction; TARGET = proposed contract/implementation; OPEN = unresolved evidence or choice. Never promote a target, old green test or local slice into production/parent completion.

Current delivery snapshot (GitHub checked 2026-10-28): #52/#53 system/AI/docs, #54 tenant/sales, #55 thesis, #57 workflow/Atlas and #58 compliance merged into main. Owner overview remains local/draft [PR #59](https://github.com/TruongHoang2004/Hexta/pull/59), review/merge/manual gates open. No new runtime tests or current CI-success claims in this cleanup.

Current Go owns business authority/sending; Python owns fake AI execution/durable run state. PostgreSQL owns business truth; Redis is cache. Conversation/send pipeline, real providers, Kafka application integration, checkpoints and self-host traces remain targets. Owner requires aggregate mappers, shared-cache redesign and root/selective rules; detailed cache/mapper mechanism remains proposed, not refactored code.

After schema/API/RPC/event/policy changes, update its owner module, relevant epic/session and CHANGELOG with actual verification boundaries. Removed superseded sources are listed in [archive ledger](../ARCHIVE_INDEX.md), optional lookup only; their original contracts do not override this pack or newer user decisions.
