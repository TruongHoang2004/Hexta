# Implementation Plan: AI Service Provider Configuration Support

- **Issue**: [#1](https://github.com/TruongHoang2004/Hexta/issues/1)
- **Date**: 2026-09-12
- **Branch**: `task/issue-1-feat-config-add-ai-service-provider-conf`
- **Scope**: Backend API service configuration (`services/api/config/`)

---

## 1. Overview & Objectives
Introduce structured configuration support for AI / LLM service integrations in `services/api`. This allows the application to cleanly configure LLM providers (e.g. Gemini, OpenAI, Claude), model identifiers, API keys, endpoints, timeouts, token limits, and temperature.

---

## 2. Impacted Components & Files
1. `services/api/config/config.go`:
   - Extend `Config` struct with an `AI` sub-struct.
   - Include fields: `Provider`, `ApiKey`, `Model`, `BaseURL`, `Timeout`, `MaxTokens`, `Temperature`.
2. `services/api/config/config.yaml`:
   - Add the `ai:` configuration section mapping to environment variables:
     - `AI_PROVIDER`
     - `AI_API_KEY`
     - `AI_MODEL`
     - `AI_BASE_URL`
     - `AI_TIMEOUT`
     - `AI_MAX_TOKENS`
     - `AI_TEMPERATURE`

---

## 3. Implementation Steps
1. Edit `services/api/config/config.go` to add the `AI` struct with YAML and mapstructure tags.
2. Edit `services/api/config/config.yaml` to declare default structure and env variable expansions.
3. Run `go build ./...` in `services/api` to verify successful compilation.
4. Run a unit test to verify that `config.LoadConfig()` loads the new AI fields properly without regressions.

---

## 4. Definition of Done
- Strict Go types for all AI settings.
- Environment variable injection via Viper.
- All code, comments, and documentation in English.
- Code review and changelog documented in `agentic-memory/`.
