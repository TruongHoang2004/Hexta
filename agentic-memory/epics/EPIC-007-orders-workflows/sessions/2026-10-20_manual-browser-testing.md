---
epic: EPIC-007
session: 2026-10-20_manual-browser-testing
status: in-review
completed: false
date: 2026-10-20
---

# Manual browser sales verification

## Scope

Owner explicitly requested delegated manual local testing and screenshot evidence. One subagent executed browser checks from [the checklist](../../../../infrastructure/TESTING.demo.md); parent reviewed the confirmation screenshot and preserved evidence. No application code or privileges changed.

## Starting evidence

Branch codex/thesis-content, HEAD 3b9d182. Existing unrelated uncommitted memory changes preserved. Approved Docker inspection showed web/API/AI/PostgreSQL/Redis healthy and init/migrate exit 0; no startup required. Shell localhost access initially failed due to sandbox isolation; browser access succeeded. Existing test-owner session and Sales verification tenant available. No fresh remote merge/CI check; deployed image/source equivalence was not independently established.

## Verification and evidence

| Case | Result | Observation |
| --- | --- | --- |
| T-1 | PASS | Tenant chooser, approved tenants and pending request visible without business sidebar. |
| T-2 | PASS | Entered tenant-qualified orders page with business sidebar. |
| C-1 | PASS | Created customer, edited name, reloaded and verified persisted value. |
| P-1 | PASS | Created product, changed price from 12 to 10, reloaded and selected in order editor. |
| O-1 | PASS | Empty draft and notes persisted after full reload/reopen. |
| O-2 | PASS | Assigned self, selected customer/product, saved quantity 2 at price 10, total 20. |
| O-3 | PASS (UI) | Assigned actor confirmed; full reload/reopen retained CONFIRMED and disabled editor. |
| V-1 | PASS | Blank unit price saved as missing; confirmation rejected, order stayed DRAFT. |
| V-2 | PASS | Explicit zero price saved and confirmed. |

Local screenshot directory (excluded from PR publication): `agentic-memory/epics/EPIC-007-orders-workflows/evidence/2026-10-20_manual-browser-testing/`: 01 tenant chooser, 02 customer persisted, 03 product persisted, 04 empty draft persisted, 05 draft total 20, 06 confirmed/read-only after reload, 07 missing price rejected, 08 zero-price confirmed.

No failures observed in these nine cases. No downstream inventory/message effects independently audited. No automated suites rerun. UI-1, T-3, V-3–V-6, C-2, A-1–A-5 and I-1–I-9 NOT RUN; do not infer complete manual acceptance.

## Outcome and decisions

Retained test-only Manual QA 2026-10-20 customer/product (example.test email, SKU MANUAL-QA-20261020) in Sales verification for reproducibility. Confirmed orders: ORD-FA1010D5-CB0B-477E-B9E2-133A40A69722 (20), ORD-CEFFCA65-BA5A-42E9-AC0E-6EED1C64DAC9 (0). Screenshots contain no passwords, tokens or invitation links. No schema/API/workflow changes; architecture verification note only.

## Remaining work and handoff

Run remaining checklist cases, especially cross-tenant/actor denial, stale edits, delivery validation, approval and invitations. Evidence/session delivery is local and unmerged; parent epic remains in-progress, completed false. No issue/tracker state changed.

## 2026-10-21 publication follow-up

Only this text record is included in the compliance PR. Eight screenshot files remain local and untracked; their presence is historical verification evidence, not an uploaded review artifact. Remaining acceptance and parent status are unchanged.
