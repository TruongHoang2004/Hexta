---
trigger: model_decision
description: "Apply the shared execution protocol when invoking a Hexta repository skill."
---
# Skill execution protocol

1. Establish the requested outcome, scope and inputs. A user task is sufficient for development; an issue is required only for workflows that operate on an issue. Do not invent missing issue IDs, diffs or acceptance evidence.
2. Read the selected canonical `SKILL.md`, then only context required by its workflow and [the root router](../../rules/README.md). Reuse already-read instructions and completed discovery results. A legacy redirect is not a request to load its siblings.
3. Execute within the user's authorization. Skills do not authorize unrelated commits, external messages, issue/PR mutations, merges, deployment, paid provider calls or memory pruning. Operational skills must verify their own prerequisites and authorization before side effects. Missing essential input blocks only dependent work; report the gap and continue independent authorized work.
4. Validate the outcome and report exact evidence and NOT_RUN prerequisites. Follow [project memory](../../agentic-memory/README.md) for authorized recording. Read-only tasks return output in conversation, with no artifact or lifecycle mutation.

## Output ownership

- Component discovery returns a reuse decision, not a mandatory design document.
- Feature/cross-service skills return the scoped implementation, consumer impact and validation evidence; one primary epic session records meaningful work.
- `dev-plan` and `dev-design` persist distinct plans/designs only when useful and writes are allowed.
- `code-review` returns findings; compatibility `dev-review`/`basic-code-review` use that procedure. Review alone does not approve or merge a PR.
- `dev-changelog` updates the consolidated CHANGELOG, not a second mandatory per-task changelog.
- Issue lifecycle, PR automation and memory consolidation remain separate operational workflows. A closed issue is not proof that knowledge can be deleted; a risk label is not merge authorization.

Names and descriptions are discovery metadata. Detailed skill bodies and references load only when selected; adapters point to canonical bodies rather than copying them.
