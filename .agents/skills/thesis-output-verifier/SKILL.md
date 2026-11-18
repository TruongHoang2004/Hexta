---
name: thesis-output-verifier
description: "Verify Hexta graduation thesis and academic documentation outputs across conversations for research depth, scope/ecosystem framing, academic abstraction, epistemic accuracy, LaTeX syntax, xelatex PDF compilation, and Notion task alignment."
---

# Thesis & Academic Output Verifier (`thesis-output-verifier`)

Use this skill to audit, verify, or refine thesis chapters (`graduation thesis/**`) and academic docs before accepting the output.

## 1. Content & Framing Checklist

1. **Problem & Market Depth:**
   - Explain the end-to-end operational problem (marketing acquisition $\rightarrow$ multi-channel consultation $\rightarrow$ persistent draft order $\rightarrow$ multi-mode fulfillment handoff).
   - Ensure market comparisons (Sapo/KiotViet/Knote, Pancake/Botcake, MISA AMIS/1Office/Base.vn) cite documented capabilities and trade-offs objectively, noting public-doc limits without claiming unmeasured superiority.
2. **Scope Boundary & Open Ecosystem Framing:**
   - Frame the architecture openly for marketing operations and a future multi-app platform, while **forbidding out-of-scope product names (`ViralForge`, `PostPilot`)** outside the explicit future-direction section.
   - Separate *Core Thesis Scope*, *Open Architectural Boundary*, and *Out of Scope*.
3. **Abstraction Layer & Academic Style:**
   - **Chapter 1 / Conceptual Sections:** Use concise Vietnamese academic terminology. Reject `\texttt{...}` code paths (`services/api`), table/column names (`tenant_id`, `domain_events`), enums (`BOT_ACTIVE`), or RPC names (`StartRun`).
   - **Chapters 2–4 / Appendix:** Verify every technical table, column, route, and RPC symbol against actual source files (`migrations/`, `proto/`, `services/`).
   - Prefer structured tables (`tabularx`, `booktabs`) and vector diagrams (`TikZ`) over verbose prose to match the ~50-page allocation.
4. **Epistemic & Citation Rules:**
   - Keep `IMPLEMENTED`, `VERIFIED`, `DECIDED`, `TARGET`, and `OPEN` distinct; label simulator and mock-provider boundaries explicitly.
   - Reject fabricated `\cite{...}`, `\bibitem`, or `thebibliography` entries unless external literature was supplied by the author.
   - Never invent personal metadata in `graduation thesis/config/metadata.tex`.

## 2. Automated Structural & LaTeX Verification

Verify on modified `.tex` files:
- Zero trailing whitespace and zero unescaped special characters (`_`, `#`, `%`, or `&` outside tables).
- Balanced `{}` braces, `\begin{...}` / `\end{...}` environments, and paired `$` math delimiters.
- Uniform `&` column counts across all rows in every `tabularx` environment.
- Every table and figure has a `\label{...}` and is referenced via `Bảng~\ref{...}` or `Hình~\ref{...}`.
- All chapter/section labels (`\label{chap:...}`, `\label{app:database}`) remain intact.

## 3. PDF Compilation & Working Tree Check

When TeX Live / MacTeX is available (`/Library/TeX/texbin` or `PATH`):

```sh
eval "$(/usr/libexec/path_helper)"
cd "graduation thesis"
xelatex -interaction=nonstopmode -halt-on-error main.tex
xelatex -interaction=nonstopmode -halt-on-error main.tex
rm -f main.xdv
```

Verify from `graduation thesis/main.log` and `git status --short`:
- `main.pdf` compiles with exit code `0` and zero `Overfull`/`Underfull` overflow warnings.
- Report exact total page count and per-chapter page range against the ~50-page plan.
- Confirm `git status --short` has no untracked build artifacts and preserves unrelated working tree edits.

## 4. Verification Report Format

Return a concise report in the user's language covering:
1. **Pass/Fail Summary** across the 4 gates (Academic Style & Abstraction, Scope & Epistemic Honesty, LaTeX Structure, PDF Build & Page Count).
2. **Concrete Findings** (exact file, line, rule violated, and fix).
3. **Build & Git Evidence** (commands run, page count, warnings, and working-tree cleanliness).
