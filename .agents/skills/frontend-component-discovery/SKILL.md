---
name: frontend-component-discovery
description: "Discover existing Hexta shared and app components before substantial UI additions or primitive changes; reuse prior discovery for the same scope."
---

# Frontend component discovery

Required context: target app/package, intended interaction and applicable frontend instructions. No issue or design artifact is required.

1. Inspect `packages/ui/src/index.ts` and search the target app's components and imports. Include app-local compositions such as sales primitives when relevant.
2. Read candidate implementations, props/variants, accessibility behavior and real call sites. A matching name alone is not an API match.
3. Map each needed interaction to an existing primitive/composition, native semantic element or documented gap. Prefer compatible reuse; do not force a shared component into a materially different interaction.
4. Keep business compositions local. Propose a new shared component only when its reusable responsibility and consumer compatibility justify it; no implicit component-library installation.

Output: a compact reuse decision with source paths, chosen APIs and reasons for any new primitive. Validate that referenced exports/props exist and relevant states can be expressed. Pass the result to feature work without repeating discovery unless scope changes. Read [frontend guidance](../../../apps/AGENTS.md); shared-package edits also select its own instructions. Do not write a standalone artifact unless requested or uniquely useful.
