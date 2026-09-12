# Changelog: AI Service Provider Configuration Support

- **Issue**: [#1](https://github.com/TruongHoang2004/Hexta/issues/1)
- **Date**: 2026-09-12
- **Branch**: `task/issue-1-feat-config-add-ai-service-provider-conf`
- **Author**: AI Pair Programmer (Antigravity)

---

## 1. Summary of Changes
Added strongly-typed configuration support for AI / LLM services in the backend service `services/api`. Enables configuring LLM providers (Gemini, OpenAI, Anthropic), API keys, models, endpoint base URLs, timeouts, token limits, and sampling temperatures via environment variables and YAML config.

---

## 2. Modified Files
| File Path | Action | Description |
| :--- | :--- | :--- |
| `services/api/config/config.go` | `[MODIFY]` | Added `AI` struct with YAML and mapstructure bindings |
| `services/api/config/config.yaml` | `[MODIFY]` | Declared `ai:` section with environment variable expansions |
| `services/api/config/config_test.go` | `[NEW]` | Added unit test `TestAIConfigParsing` to validate unmarshaling |
| `agentic-memory/plans/2026-09-12_issue-1-ai-config_plan.md` | `[NEW]` | Implementation plan for Issue #1 |
| `agentic-memory/reviews/2026-09-12_issue-1-ai-config_review.md` | `[NEW]` | Code review audit for Issue #1 |
| `agentic-memory/changelogs/2026-09-12_issue-1-ai-config_changelog.md` | `[NEW]` | Changelog and verification documentation |

---

## 3. Verification & Testing
Run unit tests in `services/api`:
```bash
go test -v ./config/...
```
Expected output:
```text
=== RUN   TestAIConfigParsing
--- PASS: TestAIConfigParsing (0.00s)
PASS
ok  	gitlab.com/ecommercehub1/api/config	0.557s
```
