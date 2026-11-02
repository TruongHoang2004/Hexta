---
epic: EPIC-001
session: 2026-11-03_ai-instructions
status: in-review
completed: false
date: 2026-11-03
---

# AI instruction architecture implementation

Owner approved the read-only proposal in chat. Baseline: main at a1c9541c6152b2871f08cb9888ac1cd07217c1c3, clean working tree. The chat workspace was empty; implementation uses the registered repository at /Users/truonghoang/Documents/dev/personal/Hexta. No application/runtime/dependency or existing automation behavior changes are authorized by this migration.

## Outcome and decisions

- Added domain AGENTS, web-specific guidance, six task workflows and generated thin Claude/Antigravity adapters. Canonical domain routes live in rules/README.md; selectors are derived from that table. Nine selectors separate protobuf from migrations so one does not load the other's instructions unnecessarily.
- Preserved existing API rule bodies/IDs/provenance and Next.js-generated blocks. Root instructions retain global boundaries; detailed session recording now lives in agentic-memory/README.md. Legacy frontend/review entries delegate to current owners.
- Reconciled stale memory/skill protocols, narrowed design skill descriptions and corrected selected portability/registration claims. Historical memory, specialist resources and scripts remain present; no content-pruning or automation activation.
- Added the instruction validator and [maintenance/compatibility guide](../../../../docs/architecture/ai-instructions.md), including A–F acceptance procedures and limitations. This guide serves as the lasting design; no duplicate standalone plan/review artifact was created.

## Verification

- `python3 scripts/check-agent-instructions.py --self-test`: PASS, five tests (metadata parsing, safe sync refusal, idempotence/no deletion, Claude hint translation and missing links), plus eleven declared route fixtures, adapter equality, canonical link and byte-budget checks; zero errors/legacy link warnings.
- Bundled skill-creator `quick_validate.py` over all 24 canonical skills: PASS using the existing AI virtualenv Python/PyYAML. Host python3 lacked PyYAML, so no dependency was installed. Six pre-existing Claude `argument-hint` fields were moved into portable metadata and retained in generated Claude adapters; adapter/rule YAML also parses successfully.
- `git diff --check`: PASS. Existing API catalog unchanged; all four Next.js-generated instruction blocks compared byte-for-byte with HEAD and preserved. Scope check: Markdown plus the instruction validator only; no deleted files.
- Static routing does not establish native-client discovery, task compliance or semantic rule enforcement. No application build/test, generator, migration, provider, deployment, remote automation or live CI validation was run for this instruction-only change.

## Remaining acceptance and handoff

Native fresh-session scenarios A–F across Codex, Claude Code and Antigravity remain NOT_RUN. Claude/Antigravity command-line entry points were not found on the current PATH; the available Codex app executable does not establish all surfaces' discovery behavior. Run the documented read-only smoke prompts in the actual supported clients and record versions/settings/read traces before marking compatibility verified. Check-only validation must not run providers, migrations or remote automation.

Owner subsequently authorized committing and pushing this migration. Prepared for direct delivery on the existing main checkout after verifying HEAD matches origin/main; native-client acceptance remains open. No PR, remote comment, deployment or parent-epic completion is claimed. Git history records the resulting delivery commit. API O/P rules remain drafts unless separately approved. Next work: perform the documented native-client acceptance and record its evidence separately from Git delivery.
