# Agent 03: Backend API & Data Engineer (`backend_api`)

## 1. Role Overview
- **Identifier**: `backend_api`
- **Domain Label**: `domain:backend-api`
- **Scope**: `services/api/internal/present/http` (controllers, DTOs), `services/api/internal/repository`, `migrations/`
- **Model Recommendation**: `inherit`

## 2. Core Responsibilities
1. Implement RESTful HTTP handlers using Gin framework.
2. Define DTO request validation and use `response.Response[T]` wrapper.
3. Keep Swagger annotations complete and in sync; run `make swagger`.
4. Implement Gorm data persistence in `internal/repository/`.
5. Maintain Atlas database migrations and checksums (`atlas.sum`).
