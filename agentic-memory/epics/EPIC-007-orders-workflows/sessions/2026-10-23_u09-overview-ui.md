---
epic: EPIC-007
session: 2026-10-23_u09-overview-ui
status: in-progress
completed: false
date: 2026-10-23
---

# U09 owner overview UI

## Scope

Implement the read-only owner overview against the coordinator-reviewed [U07 contract](../../../designs/2026-10-23_owner-overview_design.md) and concurrently implemented U08 API. Coordinator explicitly released local integration before merge; browser acceptance and delivery gates remain open. Root owns shared epic/index/architecture/changelog integration.

## Starting evidence

Branch codex/demo-work-units, main baseline 656a6c4. Preserved unrelated documentation and U08 backend edits. Read canonical/nested AGENTS, epic/API/workflow context, installed Next.js 16.3 server/client, useSearchParams, useParams and async page searchParams guidance. Existing details use inline selections; member API exposes IDs/roles rather than names.

## Outcome and decisions

IMPLEMENTED locally:

- Tenant overview route, owner-only navigation and page affordance, and typed web service matching exact U08 response. API remains authorization authority.
- Default All time; URL-backed employee, assignment and paired inclusive local-date filters. Convert Asia/Ho_Chi_Minh end date to next midnight exclusive, validate calendar dates, percent-encode offsets, reset pagination when applying filters, retain query on pagination/reload.
- Render server counts without client aggregation: every order status including other, total/active/waiting/completed work, unassigned explicitly included in active. Response labels distinguish applied filters from unsaved form edits and current-state creation cohorts from activity.
- Pinned workflow/version/current-step labels, current responsibility and separately nominated recipient, member IDs/You, unavailable-member annotations only after successful member lookup. Responsive wrapping, loading/empty/error/owner-denial states and retry via refresh.
- Abort obsolete overview loads and suppress ignored member responses; tenant/query scope remounts prevent stale records across tenant/filter changes. Refresh reloads server counts and rows together.
- Minimal validated ?order_id= and ?work_id= support loads tenant-qualified details regardless of first list page. Default landing/list behavior retained. Query pages key by selected ID. Pending order link hides editable blank draft; accepted saved ID remounts editor fields. User selection/save/create invalidates late initial requests. Review found the initial editor initialization risk during implementation; corrected before delivery. No browser reproduction/pass claimed.
- Focused Node20-compatible node:test harness transpiles actual pure date helper with existing TypeScript devdependency; no dependencies added. Added web test:overview-dates script and frontend CI step.

Files: new overview route, components/sales/overview.tsx, services/overview.ts, overview_dates.ts and overview_dates.test.mjs; modified workspace, orders/work components and tenant pages, web package script and frontend-ci.yml. No schema, SDK, refresh credentials, runtime or browser changes.

## Verification and evidence

- pnpm --dir apps/web lint: PASS after implementation; final recheck recorded by coordinator if later edits occur.
- pnpm --dir apps/web exec tsc --noEmit: PASS.
- pnpm --dir apps/web test:overview-dates: PASS 3 tests: exact week UTC/cohort bounds and URL offset encoding; same-day/leap/year end; invalid/missing/reversed range. Harness uses standard ESM/node:test and transpileModule, compatible with CI Node20; executed locally with Node24.15.0.
- git diff --check: PASS.
- pnpm --dir apps/web build: first failed existing Google-font fetch under network restriction; escalated retry fetched fonts but Turbopack CSS worker could not bind local port (Operation not permitted).
- pnpm --dir apps/web exec next build --webpack, with escalated network: PASS compilation, TypeScript, all16 static pages, routes including dynamic tenant overview/orders/work. This passed before final small detail-loading guards/member-label refinement; final lint/types passed, coordinator final Docker build remains required for integrated default-Turbopack evidence.
- No font mock/config changes. New JSX formatted with already cached Prettier.
- Browser tests NOT RUN: no usable browser surface. Real records/cards, filter/reload, refresh after transition, tenant switching, forbidden/error/empty and small viewport remain manual acceptance gaps. Backend tests are separate U08 evidence.

## Remaining work and handoff

Coordinator review, integrated image build, browser acceptance against rebuilt API/web, PR and verified merge into main. No completion/parent-epic promotion. Manually exercise an order deep link with slow detail and fast lookup response: saved fields must initialize, blank editor must remain hidden while pending, selecting New draft/another row must prevent late response overwrite. Verify saved work detail links beyond first page and forbidden tenant requests. Root should integrate this session into shared memory and dev-changelog; no shared-file edits from U09.

### Review follow-up: failed detail links

Independent review additionally identified that a failed/invalid order link must not fall through to an editable blank draft, or lose its error on list refresh. Added dedicated initialError state and a blocking detail placeholder with Retry (valid IDs), explicit New order, or list selection. Only an explicit new/selected order clears that blocked context; ordinary list refresh cannot erase it. Work-link errors now also use dedicated state retained across list refresh and cleared by explicit open/create. No browser pass claimed. Final lint/types/diff rechecked after this correction; coordinator integrated image build remains separate.

## Root integration follow-up — 2026-10-23

Root final default Docker Turbopack build/start passed after final guards. Browser verified actual counts, half-open same-day/employee/assignment filters, reload, pinned/detail links, waiting/unassigned subsets and tenant switch/empty state. Responsive override did not visibly apply; do not claim mobile acceptance. Other manual gaps remain. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
