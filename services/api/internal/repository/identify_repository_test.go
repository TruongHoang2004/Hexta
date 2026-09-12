package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		// Default to local docker postgres
		dsn = "postgres://postgres:postgres@localhost:5433/api?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping repository integration test: failed to connect to database (%v)", err)
		return nil
	}
	return db
}

func TestIdentityRepository_MultiProvider_CompoundUniqueIndex(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	// Run within an isolated transaction that rolls back after test
	tx := db.Begin()
	defer tx.Rollback()

	repo := NewIdentityRepository(tx)
	ctx := context.Background()

	testEmail := "multi_provider_" + uuid.New().String() + "@example.com"
	primaryUserID := uuid.New().String()
	passwordHash := "hashed-pwd"

	// 1. Create local identity
	localIdentity := &model.AuthIdentities{
		UserID:     primaryUserID,
		Provider:   model.ProviderLocal,
		Identifier: testEmail,
		Password:   &passwordHash,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	createdLocal, err := repo.CreateIdentity(ctx, localIdentity)
	if err != nil {
		t.Fatalf("failed to create local identity: %v", err)
	}
	if createdLocal.ID == 0 {
		t.Fatal("expected non-zero ID for local identity")
	}

	// 2. Create Google identity with the same identifier (email) and same user_id (account linking)
	googleIdentity := &model.AuthIdentities{
		UserID:     primaryUserID,
		Provider:   model.ProviderGoogle,
		Identifier: testEmail,
		Password:   nil, // OAuth identities have nullable password
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	createdGoogle, err := repo.CreateIdentity(ctx, googleIdentity)
	if err != nil {
		t.Fatalf("failed to create google identity with same identifier (compound unique index failed): %v", err)
	}
	if createdGoogle.ID == 0 {
		t.Fatal("expected non-zero ID for google identity")
	}
	if createdGoogle.Password != nil {
		t.Fatalf("expected nil password, got %v", *createdGoogle.Password)
	}

	// 3. Test GetFirstByIdentifier
	firstFound, err := repo.GetFirstByIdentifier(ctx, testEmail)
	if err != nil {
		t.Fatalf("GetFirstByIdentifier failed: %v", err)
	}
	if firstFound == nil || firstFound.UserID != primaryUserID {
		t.Fatalf("expected to find identity with UserID %s, got %v", primaryUserID, firstFound)
	}

	// 4. Test GetCredentialByIdentifier for both providers
	foundLocal, err := repo.GetCredentialByIdentifier(ctx, testEmail, model.ProviderLocal)
	if err != nil || foundLocal == nil {
		t.Fatalf("failed to get local credential: %v", err)
	}
	if *foundLocal.Password != passwordHash {
		t.Fatalf("expected password %s, got %s", passwordHash, *foundLocal.Password)
	}

	foundGoogle, err := repo.GetCredentialByIdentifier(ctx, testEmail, model.ProviderGoogle)
	if err != nil || foundGoogle == nil {
		t.Fatalf("failed to get google credential: %v", err)
	}
	if foundGoogle.Password != nil {
		t.Fatalf("expected nil password for google identity, got %v", *foundGoogle.Password)
	}

	// 5. Duplicate test: Attempt to create another identity with same provider AND same identifier
	duplicateLocal := &model.AuthIdentities{
		UserID:     uuid.New().String(),
		Provider:   model.ProviderLocal,
		Identifier: testEmail,
		Password:   &passwordHash,
	}
	_, dupErr := repo.CreateIdentity(ctx, duplicateLocal)
	if dupErr == nil {
		t.Fatal("expected error on duplicate (provider, identifier) insertion, but succeeded")
	}
}
