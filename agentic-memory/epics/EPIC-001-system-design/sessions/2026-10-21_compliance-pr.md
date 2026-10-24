---
epic: EPIC-001
session: 2026-10-21_compliance-pr
status: in-review
completed: false
date: 2026-10-21
---

# Compliance PR publication

## Scope

Owner requested a new PR for the pending branch. Publish the rule-compliance code and pending audit, domain-map, ERD, reconciliation and browser-test text records in a separate documentation commit. Screenshot files remain local and are excluded from this PR. Primary EPIC-001; related EPIC-002/003/007 context is linked from their existing records.

## Starting evidence

Fetched origin/main matches HEAD at 2abd305. Branch codex/rules-compliance has no existing open PR according to GitHub search. Implementation and prior database-backed verification are recorded in [the implementation session](2026-10-20_rule-compliance-fixes.md).

## Outcome and decisions

Prepare coordinated API/SDK cookie-auth fixes, explicit signing-key validation, Swagger contract corrections, legacy AI boundaries, strict SDK checks and formatting CI for review. Existing documentation and screenshot evidence are retained separately. Parent epic states remain unchanged; merge is required before completion.

## Verification and evidence

Fresh publication checks: Go race suite and go vet PASS; SDK strict type guard, 35 tests and declaration/ESM/CJS build PASS; web lint and tsc --noEmit PASS; Atlas API/AI history validation, formatting guard and git diff --check PASS. PostgreSQL tests require API_TEST_DATABASE_URL and are skipped in this default Go rerun; the prior isolated database run is historical evidence, not a new run. No browser E2E or production build rerun. Initial sandboxed Go setup failed; rerun with build-cache access passed.

## Remaining work and handoff

Publish the branch and open the PR against main, then record its URL and initial CI state here. Review, CI and merge remain pending. Cookie API and SDK must be upgraded together; signing keys must be explicitly configured, distinct and at least 32 bytes, and credentialed browser access needs a configured same-site frontend origin.

## Publication scope correction

Automatic approval review rejected a broad memory commit/push because screenshots could contain sensitive data and the destination had not been verified. GitHub repository metadata now confirms origin is the existing public TruongHoang2004/Hexta repository with push/admin permission. Publish reviewed Markdown/Mermaid text only; retain all eight screenshots locally without staging or uploading them. The historical screenshot directory references describe local artifacts.

## 2026-10-22 delivery follow-up

[PR #58](https://github.com/TruongHoang2004/Hexta/pull/58) is OPEN/non-draft against main at ab053d5. GitHub connector creation returned 403; GitHub CLI successfully created it with the existing account. Initial backend/frontend/Docker checks are in progress; lifecycle reconciliation succeeded. Screenshot files remain local/untracked. Review, final CI and merge are pending; completed remains false and parent states unchanged.

## 2026-10-23 merge follow-up

Owner reported merge; GitHub refetch confirmed merged at 2026-10-23T11:07:01Z into main as 656a6c4040b31e8d28049338e3e094df5a6db88a. Six returned CI checks on head 2504d69 succeeded. Pending-merge statements above describe earlier snapshots. This appended record is itself local/unmerged, so frontmatter remains in-review/completed false. Broader browser/epic acceptance remains open. [Next work session](../../EPIC-007-orders-workflows/sessions/2026-10-23_demo-work-units.md) records the handoff.
