# Development Changelog & Verification Guide: Resolve Unique Identifier Collision on OAuth Login (Issue #5)

- **Date**: 2026-09-12
- **Issue**: [#5 - fix(db): resolve unique identifier collision on OAuth login and enable multi-provider linking](https://github.com/TruongHoang2004/Hexta/issues/5)
- **Branch**: `task/issue-5-fix-db-resolve-unique-identifier-collisi`

---

## 1. Change Summary

This update resolves the PostgreSQL unique constraint collision (`idx_identities_identifier`) that occurred when a user who had previously registered with email/password attempted to log in using Google OAuth with the same email. 

It accomplishes three primary goals:
1. **Compound Unique Index**: Replaces the single-column unique index on `identities.identifier` with a compound unique index on `(provider, identifier)`.
2. **Seamless Account Linking**: Automatically links OAuth accounts to existing `user_id`s when an identical email address exists in the system.
3. **Password Nullability & Security Cleanup**: Allows `password` in `identities` to be nullable, eliminating plaintext dummy passwords (`"oauth2-dummy"`) for OAuth identities and guarding against nil pointer dereferences during login.

---

## 2. Impacted Components & Files

| File | Action | Description |
| :--- | :---: | :--- |
| [`services/api/internal/core/model/identities.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/identities.go) | `[MODIFY]` | Compound unique index `idx_identities_provider_identifier`; nullable `Password *string`. |
| [`services/api/internal/repository/identify_repository.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/repository/identify_repository.go) | `[MODIFY]` | Added `GetFirstByIdentifier` method to query identities across providers. |
| [`services/api/internal/core/service/auth_service.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go) | `[MODIFY]` | Implemented account linking in `GoogleCallback`, nil password safety in `Login`, pointer assignment in `Register`. |
| [`services/api/cmd/tools/main.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/cmd/tools/main.go) | `[NEW]` | Standalone Go schema loader using `atlas-provider-gorm` for internal package visibility. |
| [`services/api/cmd/tools/tools.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/cmd/tools/tools.go) | `[DELETE]` | Replaced by executable loader in `main.go`. |
| [`migrations/api/atlas.hcl`](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/atlas.hcl) | `[MODIFY]` | Configured external schema program to invoke `cmd/tools` and set migration directory. |
| [`migrations/api/20260912083721_fix_identities_provider_unique_index.sql`](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/20260912083721_fix_identities_provider_unique_index.sql) | `[NEW]` | Atlas migration to drop single index, alter password column, and create compound index. |
| [`Dockerfile.migrate`](file:///Users/truonghoang/Documents/dev/personal/Hexta/Dockerfile.migrate) | `[MODIFY]` | Updated base image to `golang:alpine` with `GOTOOLCHAIN=auto`. |
| [`services/api/internal/core/service/auth_service_test.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service_test.go) | `[NEW]` | Unit tests for registration, login nil check, and multi-provider linking. |
| [`services/api/internal/repository/identify_repository_test.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/repository/identify_repository_test.go) | `[NEW]` | PostgreSQL integration test verifying compound unique index and account linking. |

---

## 3. Key Technical Decisions

1. **Compound Index `(provider, identifier)`**:
   - In authentication systems with multiple identity providers, the same email or username can exist across distinct providers (`local`, `google`, `facebook`). The combination of `provider` and `identifier` must be unique to prevent collisions while preserving single-account uniqueness per provider.
2. **Automated Account Linking via Verified Email**:
   - Google OpenID Connect guarantees that user emails returned from Google's userinfo endpoint are verified. If a user previously registered via `local` email/password, associating the Google identity with the existing `user_id` ensures their profile, orders, and sessions remain unified under one account.
3. **Nullable Password Column (`*string`)**:
   - Storing dummy placeholder strings like `"oauth2-dummy"` poses security risks and pollutes the database with misleading credentials. With `Password *string` and `DROP NOT NULL`, OAuth identities store `nil` passwords cleanly.

---

## 4. Step-by-Step Walkthrough

### 4.1 Model & Migration
In [`services/api/internal/core/model/identities.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/identities.go):
```go
type AuthIdentities struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     string    `gorm:"column:user_id;not null;index" json:"user_id"`
	Provider   Provider  `gorm:"column:provider;type:varchar(50);not null;uniqueIndex:idx_identities_provider_identifier" json:"provider"`
	Identifier string    `gorm:"column:identifier;type:varchar(255);not null;uniqueIndex:idx_identities_provider_identifier" json:"identifier"`
	Password   *string   `gorm:"column:password;type:varchar(255)" json:"password,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
```

Generated Atlas migration [`migrations/api/20260912083721_fix_identities_provider_unique_index.sql`](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/20260912083721_fix_identities_provider_unique_index.sql):
```sql
-- Drop index "idx_identities_identifier" from table: "identities"
DROP INDEX "public"."idx_identities_identifier";
-- Modify "identities" table
ALTER TABLE "public"."identities" ALTER COLUMN "password" DROP NOT NULL;
-- Create index "idx_identities_provider_identifier" to table: "identities"
CREATE UNIQUE INDEX "idx_identities_provider_identifier" ON "public"."identities" ("provider", "identifier");
```

### 4.2 Account Linking Logic
In [`services/api/internal/core/service/auth_service.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/service/auth_service.go):
```go
if identity == nil {
	// Check if an identity with the same verified email already exists under another provider (e.g. local)
	existingIdentity, err := s.identityRepo.GetFirstByIdentifier(ctx, userInfo.Email)
	if err != nil {
		return nil, err
	}

	userID := uuid.New().String()
	if existingIdentity != nil {
		// Seamlessly link OAuth identity to the existing user account
		userID = existingIdentity.UserID
	}

	// Auto-register or link Google identity
	identity = &model.AuthIdentities{
		UserID:     userID,
		Provider:   model.ProviderGoogle,
		Identifier: userInfo.Email,
		Password:   nil, // OAuth identities have no password
	}
	identity, dbErr = s.identityRepo.CreateIdentity(ctx, identity)
	if dbErr != nil {
		return nil, dbErr
	}
}
```

---

## 5. Verification & Testing Guide

### 5.1 Automated Unit & Integration Tests
Run unit tests in `services/api`:
```bash
go test -v ./internal/core/service/...
```
Expected output:
```
=== RUN   TestAuthService_Register_Success
--- PASS: TestAuthService_Register_Success (0.14s)
=== RUN   TestAuthService_Register_DuplicateEmail
--- PASS: TestAuthService_Register_DuplicateEmail (0.06s)
=== RUN   TestAuthService_Login_Success
--- PASS: TestAuthService_Login_Success (0.13s)
=== RUN   TestAuthService_Login_InvalidPassword
--- PASS: TestAuthService_Login_InvalidPassword (0.13s)
=== RUN   TestAuthService_Login_NilPassword_OAuthAccount
--- PASS: TestAuthService_Login_NilPassword_OAuthAccount (0.00s)
=== RUN   TestAuthService_AccountLinking_Logic
--- PASS: TestAuthService_AccountLinking_Logic (0.00s)
PASS
```

Run PostgreSQL integration test:
```bash
go test -v ./internal/repository/...
```
Expected output:
```
=== RUN   TestIdentityRepository_MultiProvider_CompoundUniqueIndex
--- PASS: TestIdentityRepository_MultiProvider_CompoundUniqueIndex (0.03s)
PASS
```

### 5.2 Migration Application & Idempotency
Apply migration:
```bash
make migrate-apply svc=api
```
Expected output:
```
Applying migrations for api
No migration files to execute (or ok if applied for the first time)
```

Inspect table schema in PostgreSQL:
```bash
docker exec hexta_postgres psql -U postgres -d api -c "\d identities"
```
Verify:
- `password` column is nullable (`Nullable` without `not null`).
- `idx_identities_provider_identifier UNIQUE, btree (provider, identifier)`.
- `idx_identities_identifier` is removed.
