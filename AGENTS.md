# Hexta agent instructions and engineering standards

This is the single canonical file for shared agent instructions. GEMINI.md is an entry-point redirect; do not duplicate shared rules there. Use English for repository artifacts and the user's language in conversation.

## Selective rule loading

Read [rules/README.md](rules/README.md), the single rule router, then only modules selected by task paths and behavior changes. Do not preload an entire rules directory, catalog, inventory or every linked document. Re-evaluate selection when scope expands. Approved existing requirements and explicit owner decisions remain effective; draft refinements do not silently override them. The router also selects frontend/SDK and delivery rules when applicable.

## Session startup and epic memory

1. Inspect git status, current branch and current user instructions. Preserve unrelated work.
2. Read [the epic index](agentic-memory/epics/README.md), select the epic(s) relevant to the task, then read each selected epic's README and latest session record.
3. Read [the canonical system-design index](agentic-memory/designs/2026-10-18_system-design_design.md) and only the DB/API/RPC/event/workflow modules relevant to the task. Apply nested AGENTS.md and applicable skills.
4. Recheck actual code, PR/CI state and external dependencies before treating a historical snapshot as current. Repository memory is read through these instructions; not every Markdown file is automatically loaded.

### Epic and session recording protocol

- Shared engineering instructions enter here; detailed rule modules live in root `rules/`. Cross-domain architecture lives in the modular system design. Task context, progress, blockers and handoffs live under `agentic-memory/epics/<epic-id>/`.
- Each epic has README.md containing status, completed, scope/decisions, acceptance checklist, evidence links, sessions and next work. `epics/README.md` is an index only, not a combined context dump.
- Every meaningful work session gets a separate `sessions/YYYY-MM-DD_<topic>.md` in its primary epic. Use the session template. For cross-epic work, store one primary session and link it from other affected epics; do not duplicate the full log.
- Session records include status, completed, outcome, changes/decisions, verification, unresolved work and next handoff. Never overwrite previous session history to make progress look complete; append a follow-up/correction and update the current epic summary.
- States: not-started, in-progress, in-review, blocked, done. `completed: true` is allowed only with status done and evidence that all scoped acceptance criteria are satisfied. For delivered code/docs, confirm merge into main before marking done. A completed session/slice never automatically completes its parent epic. Use last-known review state until merge is verified.
- At session end update the affected epic README, its session record, the epic index row, relevant system-design modules and CHANGELOG.md. Keep sessions scoped; do not rebuild a global SESSION_CONTEXT.md.
- Legacy SESSION_CONTEXT.md is only a redirect. Retained dated designs/plans/reviews are topic-specific evidence, not startup context; removed superseded files are indexed in ARCHIVE_INDEX.md. Do not create multiple documents that merely restate one documentation change.

Keep IMPLEMENTED, VERIFIED, DECIDED, TARGET and OPEN distinct. Go owns business authorization/sending; Python returns AI candidates. Do not infer provider capabilities, implement every target automatically, or inherit legacy order AI behavior as approved messaging scope. Never expose secrets in documentation/tool output.

## Task execution

- Describe the plan before non-trivial changes. Preserve unrelated work and inspect actual code before treating memory as current.
- Keep repository code/comments/docs/commits in English; converse in the user's language. Use readable code and existing monorepo conventions.
- Every aggregate owns a dedicated pure mapper module. Do not create new per-repository cache wrappers. Detailed mapper/cache mechanisms remain proposals where labeled; read their rules only when the router selects them.
- Use the Go workspace; adding go.mod replace directives requires documented necessity. The selected Go rules contain the precise convention.
- When changing patterns, update the affected rule module and router selection rather than duplicating its full text here. Do not alter rules merely to make code checks pass.
- After code changes, select applicable verification rules/checks and report evidence and skipped prerequisites accurately. Frontend/DTO validation must agree when their contract changes.
- For invoked skills, follow `.agents/rules/skill-execution-protocol.md`: prerequisites → relevant context → execution → artifact/state. Read legacy memory/workflow rules only when applicable; older broad-loading guidance must not trigger a full rule-directory preload.
- For lifecycle/CI work, use the router's delivery module. Preserve issue/epic states and confirm merge into main before claiming done.
