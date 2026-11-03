# Hexta AI instruction architecture

Implemented from the owner-approved read-only proposal on 2026-11-03. This document is maintenance/reference context, not an always-loaded rule pack. Canonical global policy is [AGENTS](../../AGENTS.md); [the router](../../rules/README.md) selects domain guidance and existing detailed rule modules.

## Ownership and hierarchy

Global → domain → meaningful app differences → task workflow skills. Domain AGENTS live in apps, services/api, services/ai, packages/ui, packages/sdk, packages/shared, proto, migrations and infrastructure. Web adds app behavior; admin/docs/landing retain generated Next.js blocks without invented business conventions. No generic services/packages layer is needed.

The router's domain-entry table is the source for Antigravity selectors. Its behavior rows remain an explicit reasoning procedure, not an executable semantic checker. Root-started tasks must select deeper and sibling instruction files before editing. Deduplicate already-read sources. A domain entry may point to an existing detailed rule without copying it; keep stable API rule IDs and E/D/O/P provenance. Approval of this migration does not approve previously proposed API mechanisms.

Memory stays in [agentic-memory](../../agentic-memory/README.md), including architecture decisions, epics/sessions, plans, reviews, audits, changelogs and archive ledger. Read-only requests produce no files. Preserve generated Next.js blocks and their app adapters. Current manifests/source, not copied version lists, define tooling.

## Files and context budgets

| Owner | Content | Soft design target; enforced byte ceiling |
| --- | --- | --- |
| Root AGENTS | Global principles, routing, evidence and memory entry | roughly 600–900 tokens; 5,000 bytes |
| Domain/app AGENTS | Local differences and relevant validation | roughly 150–900 tokens; 5,000 bytes each |
| Router | Path/behavior ownership and adapter source table | selective index; 14,000 bytes |
| Existing API modules | Detailed rule bodies | load only selected modules; unchanged in this migration |
| New workflow SKILL | Inputs, procedure, output, validation | 4,000 bytes each |
| CLAUDE adapter | Neighboring AGENTS import | exact `@AGENTS.md`, no policy copy |
| Antigravity selector | Glob metadata and explicit reading targets | under 1,500 bytes each |
| Claude skill adapter | Canonical name/description and body reference | under 1,500 bytes each |
| This guide/history | Maintenance and evidence | zero startup cost unless selected |

Bytes are enforceable storage budgets; token figures are estimates, not measured context consumption. Existing specialist design bodies remain larger and have narrower triggers; do not load them for ordinary feature work. Expected ordinary guidance is approximately 2,000–4,000 tokens before source/framework/task evidence, subject to client measurement.

## Agent compatibility and limitations

Documentation checked 2026-11-03; native behavior still needs a fresh-session smoke test for the installed surface/version.

| Client | Documented behavior | Repository strategy |
| --- | --- | --- |
| Codex | Root-to-working-directory AGENTS chain; override file takes precedence; default 32 KiB combined instruction cap. Skills in ancestor `.agents/skills`. | Canonical AGENTS plus explicit root router for deeper/sibling tasks. No override files introduced. Do not assume shell cd reloads instructions. |
| Claude Code | Current native AGENTS support is conditional on version/settings and absence of project CLAUDE files; nested instructions are read as work reaches directories. Skills use `.claude/skills`. | Every canonical AGENTS has a neighboring `@AGENTS.md` adapter. Generated skill adapters explicitly direct reading canonical bodies and resolving resources there. |
| Antigravity | Current docs recognize AGENTS/GEMINI and ancestor customizations, but do not scan descendants below the workspace. Rules accumulate; glob/model-decision metadata selects rules. Skills use `.agents/skills`. | Root selectors use router globs and explicit read requests. Root AGENTS also requires semantic selection before editing. GEMINI remains a compatibility pointer, not a second policy body. |

Sources: [Codex instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [Codex skills](https://learn.chatgpt.com/docs/build-skills), [Claude memory](https://code.claude.com/docs/en/memory), [Claude skills](https://code.claude.com/docs/en/skills), [Antigravity discovery](https://antigravity.google/docs/customization-discovery), [Antigravity rules](https://antigravity.google/docs/rules), [Antigravity skills](https://antigravity.google/docs/skills).

A Markdown pointer is not a guaranteed native import. Claude's import syntax differs from Antigravity's inline includes. Our Antigravity selectors intentionally ask the agent to read targets and do not duplicate their bodies. Model compliance, pre-edit selection timing and native duplicate suppression are UNVERIFIED until observed. User/global/managed settings may alter discovery. Resolve conflicting repository policies at their owner; never assume last-loaded text universally wins.

## Skills and adapters

Canonical bodies remain under `.agents/skills/<name>/SKILL.md`. Six workflows cover component discovery, frontend/backend/AI implementation, cross-service changes and code review. Review aliases delegate to code-review; PR automation remains a separate operational skill. Existing design resources are preserved with narrower activation; optional external tool examples must be capability-checked before use. No skill grants unrelated external actions or paid execution.

Claude adapters expose all canonical skill metadata without independent workflow bodies. Names and descriptions are generated; edits belong in canonical skills. Claude-only argument hints are retained as `metadata.claude-argument-hint` in canonical frontmatter and emitted as `argument-hint` only in Claude adapters. Resource paths in an adapter-loaded workflow resolve from its canonical directory. Existing user/plugin same-name skills can still conflict; inspect the client's skill list during smoke tests. Legacy personal paths inside optional specialist resources are not a promise that dependencies are installed.

Run from repository root:

```sh
python3 scripts/check-agent-instructions.py
python3 scripts/check-agent-instructions.py --self-test
```

After an intentional domain-route or canonical skill metadata change, refresh adapters explicitly:

```sh
python3 scripts/check-agent-instructions.py --sync-adapters
```

Default checking is read-only. Sync only creates/updates adapters it owns, refuses handwritten replacements and never deletes stale files. Remove/rename an obsolete adapter only through a reviewed migration. The validator checks canonical links, metadata, adapter equality, route fixtures and budgets. It does not prove application compliance or model adherence. Legacy specialist links outside the new canonical set are reported as warnings rather than silently treated as verified.

## Behavioral acceptance

Run fresh sessions from repository root and the scenario's domain directory for each installed client. Record client version/surface, instruction settings, source revision/diff, loaded instruction paths and read trace. Use a read-only planning prompt so no scenario modifies source, calls providers, migrates DBs or sends external messages. Inspect client context diagnostics where available; self-reported compliance alone is insufficient.

| Scenario and sample prompt | Required observations | Failure examples |
| --- | --- | --- |
| A: Plan a new web customer form, no edits | Frontend + web guidance; inspect shared exports, target components, APIs/usages before proposing primitives | Recreates existing Button/Input without inspecting API; loads backend catalog unnecessarily |
| B: Plan an API messaging validation change | API rules and actual service/repository tests; actor/tenant, transaction/error scope | Frontend/design bodies loaded; no authorization analysis |
| C: Plan AI timeout/recovery behavior | AI provider/graph/durable guidance; fake tests, budget and uncertainty handling | Assumes live calls, unrestricted retries or sending authority |
| D: Plan additive AI RPC metadata | Root proto, both service owners, generated Go/Python outputs and old-caller defaults | Misses a binding or calls handwritten HTTP SDK generated |
| E: Plan a shared Button variant | UI API/export, token source and actual app consumers | App business logic placed in UI; consumer compatibility ignored |
| F: Plan demo Compose readiness change | Correct profile, infrastructure and applicable migration ownership; redacted checks/rollback | Runs destructive cleanup or application migration to validate instructions |

For every scenario test root-start and nested-start separately; expand scope to a sibling mid-task to test re-selection. Also test root atlas.hcl/Makefile/CI ownership, missing DB/tool prerequisites, a read-only memory request, and a normal UI task that should not invoke the design catalog. Re-run affected scenarios after router/adapters/client changes. Static fixtures only establish declared routing coverage, not these behavioral outcomes.

## Migration and status

Preserved: API rule bodies/IDs/provenance, historical memory, Next.js-generated blocks, existing scripts/CI and specialist resources. Reconciled: stale memory/skill protocols, frontend ownership, review aliases, skill metadata, and unverified team-registration claims. Added: domain entries, six workflows, thin adapters and the static validator.

Application source, dependencies, runtime configuration and CI execution behavior are outside this migration. No automation was enabled. No native multi-client smoke results are claimed by static checks. Record actual outcomes in the primary [implementation session](../../agentic-memory/epics/EPIC-001-system-design/sessions/2026-11-03_ai-instructions.md); review/merge and native-client acceptance remain separate gates.
