---
epic: EPIC-002
session: 2026-11-11_real-consultant
status: in-review
completed: false
date: 2026-11-11
---

# Real consultant service connection

Owner requested real OpenAI integration with the existing AI service as soon as
possible. Continued on codex/shopping-consultant, preserving all pre-existing Desk,
frontend, instruction and memory edits. Applied existing cross-service workflow,
API/AI/infrastructure instructions and OpenAI Docs. Official structured-output
reference: https://developers.openai.com/api/docs/guides/structured-outputs .

Implemented Python consult contract, schema-specific provider prompt, bounded
inference and validated variant selection. Reused the existing LangChain timeout,
budget, sanitized errors, disabled tracing and no-retry configuration. Fake preview
remains explicit; real providers without credentials now fail startup. Added null
history compatibility, capped Go's model catalog to 200 variants, and added an
opt-in Compose overlay/launcher connecting the existing authenticated storefront
widget through Go to Python. Only Python gets the provider key. No dependency,
public endpoint or schema change, no migration, commit, push or deployment.

Verification: `uv run --frozen pytest -q`: 70 passed, 6 database-dependent skipped.
Tests cover internal auth, Go-shaped null history/decimal catalog, catalog identity,
validation bounds, fake rejection and safe provider/quota failures. Missing-key
factory tests updated to assert startup failure. Go and Compose results follow.

OPEN: OPENAI_API_KEY was checked for presence only and remains empty in the private
0600/Git-ignored services/ai/.env. No paid/live inference has run; live accuracy,
billing/model access and browser acceptance remain unverified. User can populate
that file without sharing its contents in chat. Health alone cannot prove inference.

The connected UI is the existing signed-in Storefront assistant. Desk automatic
responses, late-AI send fencing, notifications and live channel dispatch remain
unimplemented; this session does not claim them. No order drafts are returned by
the new consultant endpoint. Local unmerged work does not complete EPIC-002.

Final verification: Go consultant service/client tests passed with `-race`; focused `go vet` passed. Compose base+live overlay `config --quiet` passed using synthetic credentials; launcher `sh -n` and scoped whitespace checks passed. Runtime startup/live inference remains pending a nonempty provider key.
