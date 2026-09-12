---
name: dev-review
description: Use this skill when performing a comprehensive code review, auditing code quality, assessing security and performance, or checking adherence to project rules. It saves the review report to agentic-memory/reviews/.
---

# Code Review Skill (`dev-review`)

Use this skill when reviewing code, checking pull requests, auditing commits, or verifying changes against project standards.

## Workflow

### 1. Inspection Checklist
1. **Adherence to Project Standards (`agentic-memory/rules/` & `GEMINI.md`)**:
   - 5-layer Go architecture boundaries respected (no DB/SQL in controllers, no HTTP in services).
   - Correct use of `errors` package (`*errors.Error`).
   - Standardized DTOs & `response.Response[T]`.
   - Swagger annotations present on all new/updated handlers.
   - Next.js App Router rules (strict TypeScript, `<Suspense>` on `useSearchParams()`).
2. **Security Checks**:
   - OAuth `state` CSRF validation present.
   - No token exposure in query parameters or client-side storage without protection.
   - User input validated/sanitized (SQL Injection, XSS prevention).
   - Sensitive credentials never hardcoded.
3. **Performance & Reliability**:
   - No N+1 database queries in loops.
   - HTTP response bodies closed (`defer resp.Body.Close()`).
   - Proper timeout contexts passed down (`ctx context.Context`).
   - Nil pointer dereference risks checked.

### 2. Formulate the Review Report
Structure the review into:
1. **Overview**: Summary of reviewed files, commits, or feature scope.
2. **The Good**: Clean patterns, good practices, and commendable implementations.
3. **Critical Issues (Bugs & Security)**: Vulnerabilities, logical flaws, or breaking changes with concrete remediation code snippets.
4. **Suggestions & Improvements**: Readability, maintainability, performance tweaks.

### 3. Persist into `agentic-memory/reviews/`
- Determine `<feature-name>`.
- Get today's date in `YYYY-MM-DD` format.
- Write the review report to:
  `agentic-memory/reviews/YYYY-MM-DD_<feature-name>_review.md`
- In your conversation response, provide a brief summary and a clickable link to the saved review file.
