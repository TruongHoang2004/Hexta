---
epic: EPIC-001
session: 2026-11-14_prd-design-skills
status: in-progress
completed: false
date: 2026-11-14
---

# PRD and technical-design skill workflow

## Scope

Implement the owner-approved learning proposal for a two-stage specification workflow (PRD → technical design). Instruction-only change under EPIC-001; no application source changed.

## Starting evidence

Branch `fix/audit-hardening` at `b417ed3`, with unrelated uncommitted work (SDK `node_modules` links, an untracked Facebook connection design) preserved. The proposal's file edits were already present in the working tree when the session started.

## Outcome and decisions

- Added canonical [dev-prd](../../../../.agents/skills/dev-prd/SKILL.md): What/Why-only PRDs, Notion and `designs/system/01-product-decisions.md` cross-reference, 7-step discovery, 13-section structure, Notion status (Bản nháp / Đang rà soát / Đã chốt) and 5 readiness checks before `dev-design`. Workflow stops at technical design unless implementation prompts are requested.
- Updated [dev-design](../../../../.agents/skills/dev-design/SKILL.md): stay within approved PRD scope; enumerate every schema change in `migrations/schema/{api,connector,ai}.sql` with the `make migrate-diff` target; enumerate every added/modified/removed HTTP route, gRPC RPC and Kafka topic with concrete request/response or payload DTOs.
- Generated Claude adapter `.claude/skills/dev-prd/SKILL.md` matches validator output.
- Condensed [code-review](../../../../.agents/skills/code-review/SKILL.md) from 4,468 to 3,987 bytes to clear its pre-existing budget failure. Wording only; every scope, rule-loading, risk, verification, severity, reporting and read-only boundary item is retained. Frontmatter unchanged, so its adapter is unaffected.

## Verification and evidence

- Referenced paths/targets exist: `designs/system/01-product-decisions.md`, `migrations/schema/{api,connector,ai}.sql`, Makefile `migrate-diff`, Kafka usage in `services/api`.
- Initial `python3 scripts/check-agent-instructions.py` run: FAIL, one pre-existing error (code-review skill 4,468 bytes at `HEAD`, over the 4,000-byte budget).
- After condensing code-review: `python3 scripts/check-agent-instructions.py` PASS (0 errors, 0 legacy link warnings, 11 route fixtures); `--self-test` 5 tests OK and PASS. dev-prd 3,305 bytes, dev-design 2,950 bytes.
- Native-client skill loading: NOT_RUN.

## Remaining work and handoff

- Commit/review/merge not done; status stays in-progress until merge into main is verified.
