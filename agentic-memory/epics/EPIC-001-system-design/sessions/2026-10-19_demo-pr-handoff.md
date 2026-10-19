---
epic: EPIC-001
session: 2026-10-19_demo-pr-handoff
status: in-review
completed: false
date: 2026-10-19
---

# Demo pull request and manual test handoff

## Scope and starting evidence

Owner requested committing the latest icon work, opening a PR and providing manual testing instructions for later. Branch codex/manual-sales-demo includes 37b9934 and prior squash-merged foundations. Existing favicon session records user-authored asset changes and prior verification. No live privilege grant is authorized or performed by this delivery step.

## Outcome and decisions

Commit current icon assets and associated memory; add [manual test checklist](../../../../infrastructure/TESTING.demo.md) with roles, setup/resume, expected results, sales/approval/invitation/negative cases and reporting format. Incorporate current main before PR publication so squash-merged foundations are not repeated in review. Preserve history. The full PR includes tenant onboarding, manual sales and branding, not just the latest icon.

## Verification and evidence

Current web lint/TypeScript and whitespace checks run before commit. PNG/ICO formats and dimensions inspected. Previous broader Go/SDK/PostgreSQL/build evidence remains in the implementation sessions; no new backend behavior changed here. Actual PR URL/check state will be appended after publication. Manual checklist cases are expected outcomes, not newly executed results.

## Remaining work and handoff

Human manual testing, PR checks/review and main merge remain pending. Live operator approval still requires separately authorized setup; no email/payment/provider integration is claimed. Parent epic stays in-progress, completed false.

## Publication evidence

Latest icon/checklist commit: dc37ab3. Main synchronization: 0cf11e0; baseline tree was verified identical to the branch parent before resolving duplicate squash-history document conflicts, and merge changed no application content. [PR #54](https://github.com/TruongHoang2004/Hexta/pull/54) is OPEN and MERGEABLE. At creation, backend/frontend/container/lifecycle CI was queued or running; no successful final CI/merge claim is made. Current web lint and TypeScript passed. Review/manual acceptance/merge remain pending.

Final icon demo rebuild passed. API/web/PostgreSQL/Redis are healthy; init/migrate exited successfully. HTTP-served icon.png and favicon.ico byte-match the committed files. Existing volume retained.
