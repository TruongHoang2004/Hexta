# Code Review Report: Resolve Unique Identifier Collision on OAuth Login & Enable Multi-Provider Linking

- **Date**: 2026-09-12
- **Issue**: [#5 - fix(db): resolve unique identifier collision on OAuth login and enable multi-provider linking](https://github.com/TruongHoang2004/Hexta/issues/5)
- **Branch**: `task/issue-5-fix-db-resolve-unique-identifier-collisi`
- **Reviewer**: Antigravity Assistant

---

## 1. Overview

This code review audits changes implemented to address database unique constraint collisions during OAuth authentication, enable seamless account linking across multiple identity providers (`local`, `google`), and eliminate plaintext dummy password storage.

### Files Reviewed:
- [services/api/internal/core/model/identities.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/identities.go)
- [services/api/internal/repository/identify_repository.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/repository/identify_repository.go)
- [services/api/internal/core/service/auth_service.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go)
- [services/api/cmd/tools/main.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/cmd/tools/main.go)
- [migrations/api/20260912083721_fix_identities_provider_unique_index.sql](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/20260912083721_fix_identities_provider_unique_index.sql)
- [migrations/api/atlas.hcl](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/atlas.hcl)
- [Dockerfile.migrate](file:///Users/truonghoang/Documents/dev/personal/Hexta/Dockerfile.migrate)
- [services/api/internal/core/service/auth_service_test.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service_test.go)
- [services/api/internal/repository/identify_repository_test.go](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/repository/identify_repository_test.go)

---

## 2. The Good

1. **Schema Integrity & Sound Compound Constraint**:
   - Replaced single-column unique index `idx_identities_identifier` with a compound unique index `idx_identities_provider_identifier` on `(provider, identifier)`.
   - Prevents duplicate registration under the same provider while enabling legitimate multi-provider coexistence for the same email.
2. **Account Linking Architecture**:
   - In `GoogleCallback`, when a Google identity record is not yet present, the service queries `identityRepo.GetFirstByIdentifier(ctx, userInfo.Email)`. If an identity exists (e.g. from prior local registration), the newly created Google identity reuses the existing `UserID`.
   - Eliminates split/orphan user accounts for the same real-world identity.
3. **Password Nullability & Security Hardening**:
   - Switched `Password` to `*string` and dropped `NOT NULL` in PostgreSQL.
   - Removed insecure storage of plaintext `"oauth2-dummy"` mock passwords.
   - Protected `Login` against nil-pointer dereferences when users authenticated exclusively via OAuth attempt email/password logins.
4. **Declarative Atlas Provider Implementation**:
   - Introduced a dedicated Go loader (`services/api/cmd/tools/main.go`) leveraging `ariga.io/atlas-provider-gorm/gormschema`. This cleanly resolves Go's internal package visibility rules without fragile shell scripts or file tampering.
   - Migration file `20260912083721_fix_identities_provider_unique_index.sql` is concise, accurate, and completely idempotent.
5. **Comprehensive Testing**:
   - Unit tests verify `Register`, `Login`, nil password handling, and account linking logic.
   - Integration tests execute against real PostgreSQL within isolated transactional rollbacks, proving both successful multi-provider coexistence and rejection of duplicate tuples.

---

## 3. Critical Issues (Bugs & Security)

No blocking or critical security vulnerabilities were discovered. All objectives specified in Issue #5 have been met without regressions.

---

## 4. Suggestions & Future Improvements

1. **Account Linking Confirmation Flow (Multi-Factor / Explicit Linking)**:
   - For trusted OpenID Connect providers like Google where email ownership is verified by the provider, automatic linking is industry standard.
   - If untrusted OAuth providers are introduced in the future (where email verification might be optional), consider requiring an email verification token or password confirmation before linking.
2. **Swagger Docs Update**:
   - If any new endpoints are introduced in subsequent issues, ensure `@Summary`, `@Tags`, and `@Router` annotations are maintained. Existing endpoints (`/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/google/callback`) maintain backward compatibility.

---

## 5. Conclusion & Verdict

**Verdict**: **APPROVED**  
All acceptance criteria satisfied. Database migration tested and verified. Code is ready for merge.
