# Code Review Report: PR #11 (Unique Identifier Collision & Account Linking)

- **PR**: [#11](https://github.com/TruongHoang2004/Hexta/pull/11)
- **Branch**: `task/issue-5-fix-db-resolve-unique-identifier-collisi`
- **Reviewed by**: AI Pair Programmer (Antigravity)
- **Date**: 2026-09-12
- **Result**: ✅ **APPROVED WITH MINOR RECOMMENDATIONS**

---

## 1. Scope of Changes
- Replaced `idx_identities_identifier` unique index on `(identifier)` with compound unique index `idx_identities_provider_identifier` on `(provider, identifier)`.
- Made `password` nullable (`*string`) across GORM model and database schema.
- Added Atlas migration `migrations/api/20260912083721_fix_identities_provider_unique_index.sql`.
- Updated `cmd/tools/main.go` for Atlas GORM schema loader.
- Implemented account linking in `AuthService.GoogleCallback` associating OAuth identities with existing `user_id`.
- Guarded `AuthService.Login` against nil-pointer dereference for passwordless identities.
- Guarded `AuthService.Register` against duplicate registration across any existing provider.
- Added extensive unit tests and PostgreSQL integration tests.

---

## 2. Verification Results
- `go test -v ./services/api/internal/core/service/...`: **PASS** (6/6 tests passed).
- `go vet ./services/api/... ./packages/shared/...`: **PASS** (0 warnings).
- Atlas migration integrity checksums: **PASS** (`atlas.sum` verified).

---

## 3. The Good
- Clean abstraction of account linking logic in `AuthService`.
- Proper nil-pointer handling for nullable `password` field.
- Excellent unit test coverage mocking `IIdentityRepository` and testing edge cases.
- Migration correctly drops the old unique index and creates the compound index without data loss.

---

## 4. Suggestions & Recommendations
1. **Unpinned Base Image in `Dockerfile.migrate`**:
   `FROM golang:alpine` uses floating `alpine`. It is recommended to pin to a stable Go version matching the monorepo (e.g. `golang:1.24-alpine`).
2. **OAuth Email Verification Verification**:
   In `GoogleCallback`, account linking connects any matching email automatically. Verify that Google returned `verified_email: true` in the userinfo payload before linking to prevent potential account takeover if an unverified Google account is used.
