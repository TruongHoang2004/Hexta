# Delivery lifecycle and CI

Existing AGENTS workflow requirements extracted for conditional loading. Scope: issue/PR lifecycle and CI/pipeline work; this file is not required for an ordinary controller or mapper edit.

## Issue and task states

- ready: fully specified and available; automated runners pick only this label.
- in-progress: actively developing; remove ready and add in-progress when picked.
- in-review: implementation/tests/memory complete and PR open for human review; remove in-progress and add in-review.
- done: only after the PR is merged into main. Use Closes #number for the intended linked issue. Do not equate an individual completed slice with parent epic completion.

Verify current PR/CI/merge state before claiming delivery. Do not fabricate issue IDs or change lifecycle labels during documentation-only work without the corresponding lifecycle action.

Retained automation conventions: ready items are prioritized by high/medium/low labels, then the runner's documented ordering; use Conventional Commits and real issue references when applicable. Task runners open review PRs and do not merge their own work or directly commit to main. Apply the current reviewer skill's risk/authorization rules for any merge action; an old low-risk matrix or a closed issue is not blanket merge permission. PR descriptions link the actual relevant plan/review/evidence, without manufacturing redundant documents. Use the current agent branch-prefix instruction rather than copying obsolete task/issue branch examples.

## CI responsibilities

- Backend workflow owns Go formatting, race tests, vet and Atlas history/drift checks; inspect actual YAML before changing its behavior.
- Frontend workflow owns Node/pnpm setup, SDK build/strict types/tests, web lint and production build; current instructions specify pnpm v10/Node v20, but inspect the actual workflow before asserting its configuration.
- Docker workflow builds API/web/migrator images and publishes to GHCR on main. Changes to build/publish triggers require reviewing the actual workflow and relevant service rules.
- Preserve cancellation of superseded PR runs; do not cancel required main delivery work unintentionally.

Select Atlas/auth/config/other rule modules only for the behaviors the workflow actually changes. CI does not remove the need for task-specific checks or permit unsafe application DB migrations during a checker run.
