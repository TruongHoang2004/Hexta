---
epic: EPIC-001
session: 2026-10-19_thesis-evidence-logo
status: in-review
completed: false
date: 2026-10-19
---

# Thesis evidence correction and UET emblem

## Scope

Owner clarified that Notion contains their own thinking, not consulted literature. Remove bibliography/citations and leave insufficiently supported sections empty. Find the official UET emblem and insert it without changing thesis layout. Continue editing the existing main.tex project; keep the editor open.

## Starting evidence

Branch codex/manual-sales-demo, HEAD 45e7db1, prior thesis edits uncommitted. Read the prior thesis session and current TeX. Preserve previous session history; this correction supersedes its bibliography and historical-test presentation decisions. No application or merge state changes.

## Outcome and decisions

Removed every citation and bibliography entry; retained the references include as an empty source so no bibliography heading or TOC entry is emitted. Cleared related-work prose, experiment data/scenarios, results/evaluation, conclusion/future-work and abstract. Removed placeholder instructions from acknowledgements/declaration. Removed historical test counts, invented evaluation protocol and editorial send equation. Kept requirements/design based on owner decisions and implementation descriptions supported by code. No replacement prose was added to the empty sections.

Downloaded the university site's emblem into graduation thesis/assets/logo.png (512x512 PNG). Official page https://uet.vnu.edu.vn/ declares cropped-logo2_new-1 icon variants; downloaded the original asset https://2679558515.e.cdneverest.net/uet.edu.vn/2017/02/cropped-logo2_new-1.png. Inspected the downloaded image, matching the UET emblem in the site's current header. Provenance is in metadata source comments only, not the thesis bibliography. Existing SchoolLogo macro automatically inserts it at the original 3.2cm height. No image alteration, separate PDF export or new editor tab.

## Verification and evidence

compile_latex_document with the saved existing main.tex path succeeded. Python comparison against HEAD confirms all chapter/section/subsection titles and ordering unchanged. No cite/bibitem/thebibliography remains and no body placeholders remain. main.tex, preamble, cover and contents have no diff. Downloaded original emblem was visually inspected. No full rendered-page visual QA was available; no application tests were run for these document-only edits.

## Remaining work and handoff

Author supplies personal metadata and final title, acknowledgements/declaration, any actual literature, experiment dataset/protocol/results and final abstract/conclusions when ready. Empty sections should stay empty until evidence exists. Review final pagination in the existing editor. Local/unmerged; completed false. No shared architecture change.
