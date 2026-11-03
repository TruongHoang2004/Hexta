# Hexta agent instructions

Canonical repository-wide instructions. Use English for repository artifacts and the user's language in conversation. Agent adapters reference these instructions; they do not define independent policy.

## Select context before acting

- Inspect the current request, branch and working tree; preserve unrelated work. Describe the plan before non-trivial changes.
- Read [the rule router](rules/README.md), then only the domain instructions and behavior modules matching intended files and affected consumers. Read applicable ancestor/local `AGENTS.md` files before editing, including for sibling directories and tasks started at the repository root. Re-evaluate when scope expands; do not assume changing shell directory refreshes client instructions.
- Deduplicate already-read files. Links are lookup options, not permission to recursively load catalogs, entire directories or all skills. Domain guidance specializes global requirements; unresolved contradictions must be surfaced, not silently resolved by copying another rule.
- For meaningful project work, use [the epic index](agentic-memory/epics/README.md) to select relevant epics and latest relevant sessions. For architecture/contracts, use [the system-design index](agentic-memory/designs/2026-10-18_system-design_design.md) and only relevant modules. Recheck code and current external state before trusting dated evidence.

## Engineering boundaries

- Reuse existing implementations and public package APIs before creating alternatives. Preserve service/package ownership and cross-service compatibility; trace actual consumers when contracts change.
- Go owns business authorization and sending; Python returns untrusted AI candidates. PostgreSQL owns durable business truth; Redis is cache. Configured infrastructure does not prove implemented integration.
- Keep approved requirements and explicit owner decisions distinct from proposals. Existing rule IDs and E/D/O/P provenance remain authoritative: relocating instructions does not approve draft mechanisms. Keep IMPLEMENTED, VERIFIED, DECIDED, TARGET and OPEN distinct.
- Keep secrets and private payloads out of source, public errors, logs, artifacts and tool output. Preserve tenant isolation and current-state authorization.
- Use existing workspace manifests, lockfiles and tooling. Justify new dependencies; do not introduce a framework or independent lockfile as a shortcut. Generated artifacts must come from their owning sources and documented generators, never manual edits.
- Update the owning instruction or contract when an approved pattern changes; do not duplicate rule bodies or relax rules to make checks pass.

## Verification and memory

- Run checks appropriate to changed behavior and its consumers. Report scope, command, result and missing prerequisites; unavailable integration checks are NOT_RUN, not PASS. No old green run proves new edits correct.
- Behavior changes need meaningful coverage for affected validation, authorization, failure and consistency paths. Instruction-only changes use the [instruction validation guide](docs/architecture/ai-instructions.md); application tests are not a substitute for instruction loading tests.
- Follow [project memory](agentic-memory/README.md) for session/status recording. Explicit read-only requests prohibit artifact writes. Preserve history; one session does not complete a parent epic, and delivered work is not done until merge into main is verified.
- Use relevant workflows from `.agents/skills/` with the [skill protocol](.agents/rules/skill-execution-protocol.md). A skill does not expand task authorization to external messages, issue changes, merging, deployment or pruning. Select [delivery rules](rules/delivery-workflow.md) only for lifecycle/CI work.
