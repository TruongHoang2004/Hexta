# API service engineering rules — review draft

Status: DRAFT, pending owner review. Version: 0.3, 2026-10-28. Baseline: `b52a8ba` on `codex/demo-work-units`.

This folder specifies the intended rules for `services/api` and its shared contracts/migrations. It does not activate a hook, change runtime code, or replace approved [AGENTS.md](../../AGENTS.md) instructions. Existing approved requirements and explicit owner-directed rules remain effective; the remaining MUST/MUST NOT statements describe the proposed contract pending review. Framework conventions below are repository policy, not claims about upstream framework requirements.

## Selection

Use the single [root rule router](../README.md) to choose modules. This file is optional catalog metadata for rule authors/reviewers, not a mandatory preflight read. Do not load all sibling files or the inventory by default. Detailed rule bodies are split by responsibility; IDs and provenance remain stable.

## Interpreting a rule

Every rule has a stable ID, origin, requirement and verification method. IDs must survive rewording. New IDs must not reuse removed meanings.

- **E — Existing**: restates an explicit current instruction in AGENTS.md or development rules. Its presence in this draft does not make that instruction optional.
- **O — Observed**: formalizes a pattern found in code. Observation alone is not owner approval or proof of complete compliance.
- **D — Owner-directed**: explicitly requested direction in the mapper/cache/root-location follow-up; detailed package/protocol choices remain proposals unless separately approved.
- **P — Proposed**: adds precision or a stricter boundary beyond explicit existing instructions; requires review.
- **S — Static**: import/AST/config/generated-artifact check can establish a bounded fact.
- **T — Test**: behavioral/unit/integration evidence is necessary.
- **R — Review**: reasoning about business semantics is necessary. A grep match or green test suite cannot establish full compliance.

MUST is required; MUST NOT is forbidden; MAY is permitted within the stated boundary. This draft avoids implicit SHOULD exceptions. Each listed verification method is planned coverage, not an already installed check.

Rule assessment states: PASS, FAIL, NOT_RUN (including missing prerequisites), NOT_APPLICABLE (with reason). A rule may have several evidence items; all applicable required items must pass. Existing deviations remain FAIL, even if migration work is separately scheduled. No silent whitelist or false whole-codebase PASS.

## Authority and adoption

The root `rules/` location and aggregate mapper ownership are owner decisions; replacing per-repository cache wrappers is the owner-selected direction. Cache mechanics and mapper subpackage layout remain proposed.

The owner-requested rule folder is a modular elaboration of the canonical AGENTS.md, not a second independent authority. Before activating proposed refinements, review this draft and keep AGENTS entry instructions consistent with the selected modules. Legacy memory rule files are compatibility redirects, not duplicate policy. User instructions take precedence; source code demonstrates implementation, not permission to weaken a rule.

Changes to policy require an explicit rationale, affected rule IDs, compatibility/deviation assessment and owner review. A hook must not rewrite rules or baselines merely to make its own output pass. No hook implementation is included in this deliverable.
