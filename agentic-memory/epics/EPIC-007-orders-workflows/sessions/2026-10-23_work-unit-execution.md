---
epic: EPIC-007
session: 2026-10-23_work-unit-execution
status: in-review
completed: false
date: 2026-10-23
---

# Delegated work-unit execution

## Scope

Owner explicitly authorized spawning subagents to solve the nine units in [the plan](../../../plans/2026-10-23_demo-work-units_plan.md). Execute dependencies in order, preserve unrelated documentation and local screenshots, and retain separate unit evidence.

## Starting evidence

Fetched origin/main at 656a6c4 (PR #58 merge). The tracked source tree exactly matches prior head 2504d69. Created codex/demo-work-units from origin/main with all planning edits/untracked screenshots preserved. Existing demo runtime was stopped at startup. U01 owns the first rebuild and browser slot; U02/U03 may prepare read-only while authentication acceptance runs.

## Outcome and decisions

Execution in progress. Root integrates shared epic/index/changelog records. Unit agents write separate sessions and change only their owned scope. Shared browser/runtime testing is serialized; application data is preserved.

## Verification and evidence

Baseline comparison: git diff HEAD origin/main before branch creation returned no differences. Runtime/browser results pending unit reports. No acceptance inferred from previous sessions.

## Remaining work and handoff

U01 -> U02/U03/U04/U05 (subject to browser/account availability) -> U06 -> U07 -> U08 -> U09. Parent epic remains in-progress/completed false; delivered artifacts require reviewed merge before done.

## Integrated execution outcome

All nine units were assigned to named subagents; a separate read-only integration reviewer checked U08/U09. Root reviewed and released the settled U07 contract, staged U08/U09 together without waiving original manual gates, integrated shared records, fixed/reviewed introduced integration issues and ran final combined verification. No original verification unit reproduced a source defect; U08/U09 implement the new owner overview.

| Unit | Delivered evidence / implementation | Still open |
| --- | --- | --- |
| [U01](2026-10-23_u01-auth-verification.md) | Cookie-only HTTP refresh/logout/old-access rejection, SDK35 and focused auth race tests; owner browser reload | Fresh browser login/logout, expiry-driven refresh and cookie/storage metadata |
| [U02](2026-10-23_u02-sales-validation.md) | Sales6 race tests; root V-3 address rejection/success, V-4 numeric validity, V-6 stale save rejection; empty draft and explicit confirmation smoke | Full V-5 matrix, full UI-1, product archive portion of C-2; customer archive chronology below |
| [U03](2026-10-23_u03-invitations.md) | Invitation4 race/HTTP tests | Browser I-1–I-7 require separate employee identity/profile; none inferred passed |
| [U04](2026-10-23_u04-access-boundaries.md) | Access4 race/HTTP tests; root T-3 tenant switch, A-2 platform operator denial, A-5 read-only service plan | Remaining multi-actor browser denial/application cases; A-3/A-4 existing authorized operator |
| [U05](2026-10-23_u05-basic-handoff.md) | Workflow7 service+1 HTTP race tests; root explicit work creation, direct handoff, owner reassignment, responsible final completion, waiting nomination and reload/history | Actual nominated employee browser acceptance/wrong-actor scenarios in a second identity |
| [U06](2026-10-23_u06-workflow-edges.md) | Focused3 race tests; root new publication pins old running steps/policy, unassign cancels nomination without advancement, assign from queue | Browser actual accept/reassign competition and former-recipient attempts; automated race already passed |
| [U07](2026-10-23_u07-owner-overview.md) | Contract authored and root-reviewed; exact semantics accepted for local implementation | Reviewed merge of docs |
| [U08](2026-10-23_u08-overview-api.md) | API/Fx/Swagger and six focused PostgreSQL/HTTP/repository/Fx checks; full Go race/vet integration | Reviewed merge, remaining manual authorization observations |
| [U09](2026-10-23_u09-overview-ui.md) | Owner page/nav/filter/pagination/detail links; lint/types/date3 tests/CI; final Docker default production build; bounded root browser verification below | Full failure-state/multi-profile/mobile acceptance, reviewed merge |

Counts above overlap; they are not a distinct-test total. All delivered artifacts remain completed:false until merged. Partial browser evidence does not complete entire units or parent epic.

## Root verification and corrections

- `GOCACHE=/private/tmp/hexta-go-cache go test -race ./services/api/... ./packages/shared/...` with privately exported API_TEST_DATABASE_URL: final rerun PASS. Initial integrated run found docs/contract_test.go direct router call missing the new overview argument; corrected, focused docs race/vet PASS, full rerun PASS. Other package results were reused from the first successful executions where cached.
- `GOCACHE=/private/tmp/hexta-go-cache go vet ./services/api/... ./packages/shared/...`: PASS.
- U09 web lint (zero warnings) and TypeScript PASS; `pnpm --filter web run test:overview-dates` root PASS3. Tests transpile the real pure helper with existing TypeScript and support Node20. Frontend CI runs the script. Host webpack build passed; final default Docker Turbopack build also passed with current final source, superseding host environment font/worker-port limits.
- `./scripts/local.sh -f /private/tmp/hexta-u01-runtime/compose.yaml up --build -d --wait api web`: PASS. API image config sha256:dfeff9030a0d8c00fab05f2de38a7bb561e4d428539394b5dbf3f560bc42b74e; web sha256:450a3e282f5cc38596a92741941b97d75d9199706c008b29a8012af3d8e5178f. API/web/PostgreSQL/Redis healthy, init/migrate exit0. Data volume retained; no new migration or unrelated-service stop.
- Independent [integration review](../../../reviews/2026-10-23_owner-overview_review.md): no unresolved actionable findings after correcting editor initialization/remount and persistent failed-link context. Focused concurrency test proves one snapshot through a concurrent committed write, rather than merely inspecting transaction options.

## Root browser recovery and bounded evidence

A fresh root CUA inventory exposed the IAB provider even though child resumption inventories were empty. Creating a fresh hidden IAB tab restored the existing owner session; no new owner credentials or privilege changes. User subsequently selected using the existing test owner. Temporary tabs are turn-scoped; main overview tab was marked deliverable. This supersedes the earlier environment-wide browser-unavailable handoff, while retaining unrun cases above.

Created through UI only: named U02 U09 verification customer/product (example.test address, SKU U02-U09-20261023), confirmed order ORD-C0C75359-F8C0-4517-9460-6D1876AC11DC (ID625db8b1-4ff0-4ae6-9463-4d2f3f9cce1d, quantity2, price10, value20), U05 U09 workflow versions1/2 and two linked work items. These are test observations, not business/evaluation data.

Observed browser results:

- Empty draft saved, assigned/customer/product draft total20 persisted. Quantity0 and -1 and price-1 invalidity observed; zero save blocked. Missing delivery address confirmation returned validation and retained draft; supplied test address then explicit confirmation produced read-only version5. A concurrent old-tab save failed stale; confirmed result retains winner notes. This is a sequential stale-edit check, not a browser race.
- Work list remained empty after confirmation until explicit creation. Published v1 Review(direct)/Finish, created work2795bb21-fc28-4030-8006-43a99bf47a19, then published v2 Revised review(acceptance)/Finish. Old running item retained v1 Review/direct; direct transfer advanced to the peer, owner reassigned self without advancing, then completed the final step. Order remained confirmed/value20 afterward.
- Created v2 work275be0b4-480d-4edd-9bdc-97968ea79833. Nomination retained owner responsibility/current step while waiting; unassign cancelled pending nomination without advancement; reassigned owner and nominated again. Reload of deep link retained waiting/version5 and complete five-event history. Actual recipient acceptance is still NOT RUN in browser.
- Overview initial actual baseline orders5/draft1/confirmed4/work0. After UI records: orders6/draft1/confirmed5, work2; observed active1/completed1/unassigned1 intermediate subset, final waiting1/completed1/unassigned0. Count rows and pinned versions match saved work. Unassigned filter/reload persisted, zero state was explicit. Same-local-day23Oct query encoded +07 bounds correctly and returned order1/work2; employee owner reduced intermediate work to its one completed row, excluding unassigned. Restored all-time filters afterward.
- Work and order overview links opened the actual selected tenant-qualified detail, including fields/history; source order remained readable after the observed customer archive below. Switched to Empty workspace tenant: zero order/work counts and empty rows without prior tenant values; returned owner plan read-only; /platform denied operator access. Main final overview reload retained expected saved counts.
- Desktop screenshot visually inspected and retained only locally at `/private/tmp/hexta-owner-overview-2026-10-23.png`. A requested390x844 viewport override returned a default1280-wide screenshot; reset override and do not claim responsive verification. No credentials/invitation links/screenshots published.

## Archive approval event and actual state

While testing C-2 on the newly created U02 U09 verification customer, the Archive click timed out on a browser input command. Automatic approval review rejected a subsequent confirmation call because it required explicit user authorization to archive that record. Root attempted to dismiss the pending confirmation instead. The next UI nevertheless showed "Record archived" and the named test customer absent from the active directory. Root stopped all further archive actions. No product archive was attempted after rejection; the U02 U09 test product remains active. Read-only order inspection confirmed the historical confirmed order, quantity/price/value/address/notes remain accessible with unavailable-customer label. Do not claim the blocked confirmation was authorized or that cancellation preserved the record. Product C-2 still needs explicit approval; no workaround or direct DB/API archive was used.

## Final handoff

Implementation/review/checks are ready for a draft review PR. Original screenshot evidence remains untracked, and private runtime/identity/test fixtures remain outside publication. Local server13000/18080 stays running for the owner; disposable test PostgreSQL cleanup is coordinator-owned after final checks. No task or parent epic is marked done. Remaining approval/manual/merge gates above remain visible.

## Cleanup and publication preparation

Disposable hexta-work-unit-postgres container removed after final checks; both private test-server credential files deleted. Application runtime/volume and private runtime overrides retained. Staged51 reviewed source/text files, excluding all screenshots/private fixtures. Documentation local links and credential-pattern scan passed; git diff --check and Go formatting guard passed. Preparing a draft review PR; no merge/parent completion.
