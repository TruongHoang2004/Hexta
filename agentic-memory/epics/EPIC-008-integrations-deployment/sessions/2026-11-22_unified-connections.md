---
epic: EPIC-008
session: 2026-11-22_unified-connections
status: in-review
completed: false
date: 2026-11-22
---

# One connection model across channel types

## Scope

Owner request: treat every Facebook Page or Telegram bot as one connection with a type, with any number per tenant, instead of fixed per-platform sections. Follows the event-contract session ([EPIC-005](../../EPIC-005-kafka/sessions/2026-11-22_event-contracts.md), PR #93); this branch is stacked on it.

## Starting evidence

Branch `feat/unified-connections` from `refactor/event-contracts` (`3858530`, after the owner kept the existing topics). Data model already generic (`messaging_connections.platform`, N rows per tenant; threads/messages/outbound keyed by `connection_id`). Gaps: per-platform list/update routes and service methods, a web Channels page hard-coding Facebook and Telegram sections, type copy scattered across maps.

## Outcome and decisions

- API: `GET /tenants/{id}/connections?platform=` and `PATCH /tenants/{id}/connections/{connectionId}` replace the four per-platform list/update routes; creation routes unchanged. `ChannelService.ListConnections`/`UpdateConnection` dispatch by the connection's own type; `rotate_token` on a type without credential rotation returns 400. `types.ChannelPlatforms` registry + `ChannelConnectionActions`; `ConnectionDTO` (one shape for all types, with `actions`) replaces Facebook/Telegram DTOs. Swagger regenerated with `make swagger`; contract test updated.
- Web (`modules/settings`): one connection list with type filter and per-row actions from `actions`, "Add connection" dialog choosing Facebook (OAuth) or Telegram (Bot Token), shared `useConnections` hook and `ConnectionRow`; per-platform settings/rows/hooks removed. Inbox banners/composer locks load the same list with one request. i18n en/vi updated.
- No schema change; authorization, audit names (`channel.<platform>.<action>`) and connector behavior unchanged.

## Verification and evidence

- `go vet` + gofmt clean. Isolated fixture (`compose.connector-test.yml`, project `hexta-connections-fixture`, removed afterwards): `go test -race -count=1 -p 1 -v ./services/connector/... ./services/api/... ./packages/shared/...` → 178 PASS, 1 SKIP (live Telegram), after updating the Swagger contract test (the first run failed only on its stale route list). New: `TestConnectionsAcrossTypesAreListedAndManagedTogether` (2 Pages + 1 bot in one tenant, filters, other-tenant isolation, owner-only, rotate rejected for Facebook), `TestChannelConnectionActionsFollowStateAndType`.
- Web: `pnpm --filter web run lint` PASS; `pnpm --filter web run build` PASS (type check included).
- NOT_RUN: browser walkthrough of the new Channels screen, live Meta/Telegram.

## Remaining work and handoff

Review after #93. Then: browser acceptance of add/pause/rotate/disconnect on the demo stack with the mock platform.
