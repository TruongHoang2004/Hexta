# Manual tenant-first demo checks

Use this checklist for the local demo PR. Record each case as PASS, FAIL or NOT RUN; this document describes expected behavior, not proof of completion. No provider keys, prepared dataset or email service are required. Use test-only data entered through the UI.

## Start and resume

From the repository root, with Docker Desktop running:

```sh
./scripts/local.sh up --build -d --wait api web
./scripts/local.sh ps
```

API, web, PostgreSQL and Redis should be healthy; init/migrate should exit successfully. Open http://localhost:3000/tenants. The build is required for the latest icon; source edits do not update the Docker deployment automatically. If favicon is cached, hard-refresh or open a new private browser window.

Keep your existing local account credentials privately. Existing approved tenants and orders are preserved. Use two isolated browser profiles/windows for owner and employee; tabs in the same profile share authentication. A normal member account and an optional non-member account make permission tests easier.

To stop for the evening: `./scripts/local.sh down`. To resume unchanged images: `./scripts/local.sh up -d --wait api web`. Do not use `down -v`: that deletes demo data. See [runtime runbook](README.demo.md) for logs and operator setup.

## Priority 1: first working sales demo (about 15 minutes)

| ID | Action | Expected result |
| --- | --- | --- |
| UI-1 | Open home/login, toggle light/dark, open /tenants | Hexta logo remains readable; browser tab uses the blue emblem; no broken image |
| T-1 | Sign in and open /tenants | Account-level chooser and registration requests; business sidebar appears only after entering a tenant |
| T-2 | Enter an existing approved tenant | URL is /tenants/{id}/orders; sidebar shows the tenant and its resources |
| C-1 | Customers: create a named test customer with a phone or email; edit and reload | Data persists; updated values visible |
| P-1 | Products: create a test product/SKU and price; edit and reload | Data persists; product is available in the order editor |
| O-1 | Orders: New order, leave customer/items empty, Save draft; reload and reopen | Incomplete draft persists; no implicit confirmation |
| O-2 | Assign the draft to yourself, choose customer/product, quantity 2 and unit price 10; save | Saved value is 20; draft remains editable |
| O-3 | Confirm as the assigned sale, reload and reopen | State is CONFIRMED and remains read-only; no payment, stock reservation or customer message occurs |
| T-3 | Change tenant; enter another tenant or open two tenant URLs in separate tabs | Each screen uses its URL tenant; no orders/catalog from the other tenant appear |

## Priority 2: validation and stale edits (about 10 minutes)

| ID | Action | Expected result |
| --- | --- | --- |
| V-1 | In a fresh draft, clear unit price and save; attempt confirmation with other fields complete | Missing price remains missing and confirmation fails; it is not treated as zero |
| V-2 | Enter unit price 0 explicitly and save, with all other required data present | Zero is valid and can be confirmed by the assigned sale |
| V-3 | Select delivery, leave address empty, save and try confirming | Confirmation rejects missing address; succeeds after a valid address is saved |
| V-4 | Try zero/negative quantity or negative price | Invalid values are rejected; no invalid confirmed record is created |
| V-5 | Attempt confirmation without customer contact, items or assigned sale | Validation explains missing requirements; record stays draft |
| V-6 | Open one saved draft in two tabs, edit/save in tab A, then edit/save the old version in tab B | Stale save is rejected; refresh/reselect before retrying; A's data is not silently overwritten |
| C-2 | Archive a test product/customer used by an existing order | Active catalog excludes it; historical order remains accessible |

## Priority 3: application and approval

A newly registered account has no tenant membership. It may submit requests immediately. Testing approval requires a separately provisioned Hexta platform operator; tenant ownership does not grant this authority. No operator has been auto-created by this PR. A trusted administrator may use the explicit CLI in [the runbook](README.demo.md#manual-sales-demo-no-provider-keys-or-prepared-dataset) with a privately supplied PLATFORM_DATABASE_URL. If that is not set up, mark approval cases NOT RUN and continue sales tests with existing tenants.

| ID | Action | Expected result |
| --- | --- | --- |
| A-1 | New account: register/sign in; submit a unique name/slug in /tenants | Request is PENDING; no tenant card/access is created |
| A-2 | Ordinary account opens /platform | Access denied; cannot read/review applications |
| A-3 | Authorized operator opens /platform and approves the test request | Request becomes APPROVED; applicant refreshes /tenants and sees the new tenant; applicant is owner |
| A-4 | Submit a second unique request; operator rejects with a reason | Applicant sees REJECTED and reason; no tenant is provisioned |
| A-5 | Owner views Service plan | Plan/status are read-only; approval is not shown as payment collection |

## Priority 4: invitations and tenant boundaries

Use an employee account with a different email in a separate browser profile. Owner-created links are private credentials: never paste them into bug reports or screenshots. No email is sent; manually copy/share only within your test workflow. Opening a link alone must not activate membership.

| ID | Action | Expected result |
| --- | --- | --- |
| I-1 | Owner: Members & invitations, enter employee email, Create invitation link | Pending invitation plus private link; explicit no-email-delivery notice |
| I-2 | Employee: open link while signed out, then sign in with matching email | Invitation survives sign-in; preview/accept screen appears; no membership until acceptance |
| I-3 | Employee accepts; reload /tenants and enter tenant | Tenant access appears; owner member list includes employee once |
| I-4 | Fresh invitation opened under a different account email | Preview/accept rejected; no access added |
| I-5 | Owner revokes a pending invitation, then recipient opens old link | Old link cannot activate membership |
| I-6 | Owner reissues to the same email before acceptance | Earlier pending link is invalid; newest link works |
| I-7 | Recipient declines a fresh invitation | No membership added; invitation shows declined |
| I-8 | Member tries owner invitation controls; uninvited account opens a tenant deep link | Owner actions/access denied; tenant records not exposed |
| I-9 | Non-owner tries another sale's draft edit/confirmation | Cannot edit another sale's assigned draft; only assigned sale confirms (owner may reassign then assigned actor confirms) |

Seven-day expiry and concurrent approval/acceptance are covered by isolated automated checks; do not alter the machine clock or database to simulate expiry during the manual session. Chat, configurable workflows, cancellation/returns, payment, AI/providers and actual email delivery are outside this PR's manual acceptance scope.

## Report an issue

Record case ID, account role (no password/token), tenant name, exact steps, expected/actual behavior and whether reload changes it. Attach a screenshot without invitation links/authentication values. Example: `V-6 FAIL — second tab overwrites first tab after stale save; reproducible twice`. Keep environment/log details separate from business observations.
