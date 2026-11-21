---
epic: EPIC-001
session: 2026-11-21_token-diet
status: in-review
completed: false
date: 2026-11-21
---

# Agent context reduction (token diet)

## Scope

Owner request: reduce the context agents load per task. Instruction, skill catalog, repository hygiene and thesis relocation only; no application behavior change except the SDK `prepare` script.

## Starting evidence

Worktree `.claude/worktrees/token-diet`, branch `chore/token-diet` from `origin/main` `30b43c0`. Measured before: ~12.2 MB tracked; 27 skills listed per session; ~30 KB of always-read instructions/memory before code; `ui-ux-pro-max` data ~4 MB; generated swagger/protobuf ~650 KB in search results.

## Outcome and decisions

- Thesis moved to sibling repo `../hexta-thesis` via `git subtree split` (6 commits preserved) with its Makefile targets, AGENTS/CLAUDE and `thesis-output-verifier`; removed here with the Makefile thesis targets.
- Design skills (`design`, `banner-design`, `brand`, `slides`, `design-system`, `ui-ux-pro-max`) moved to sibling repo `../hexta-design-skills` (source `30b43c0`). Review aliases `dev-review`/`basic-code-review` removed; references point to `code-review`. Catalog: 27 → 18.
- Untracked build output: `packages/sdk/dist/` (now built by an SDK `prepare` script on install; CI/Docker already build packages first) and stray `packages/ui/node_modules/` symlinks. Removed stale, unused per-app `pnpm-lock.yaml` in web/docs/admin (root shared lockfile is authoritative).
- Added `.ignore` so ripgrep-based search skips lockfiles, generated code, brand assets, worktrees and CHANGELOG/ARCHIVE_INDEX; files stay readable explicitly.
- AGENTS.md: domain AGENTS is the entry for single-domain work; the router is read for cross-domain, root/unowned, contract or referred work; memory is loaded only when continuing an epic or changing architecture/contracts. Router intro aligned.
- Epic index reduced to its navigation table (5.4 KB → 1.7 KB); dated status prose moved verbatim to [STATUS.md](../../STATUS.md), consistent with the existing navigation-only rule.

## Verification and evidence

- `python3 scripts/check-agent-instructions.py` and `--self-test`: PASS (native-client behavior NOT_RUN).
- `pnpm install --frozen-lockfile --offline`: PASS, `prepare` regenerated `packages/sdk/dist`; `@hexta/sdk` lint PASS, tests 35/35 PASS.
- `make thesis` in `../hexta-thesis`: exit 0, `main.pdf` produced, working tree clean.
- `pnpm --filter web run build`, Docker image builds and fresh-session instruction smoke tests: NOT_RUN.

## Remaining work and handoff

Owner review and commit/PR of `chore/token-diet`; decide remotes for the two new repositories. Optional next steps: split oversized source files (`apps/web/components/messaging/inbox.tsx`, `sales/orders.tsx`, `sales/work-items.tsx`, API `messaging/auth/channel` services) and archive older designs/sessions.
