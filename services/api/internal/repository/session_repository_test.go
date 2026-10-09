package repository

import (
	"context"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
)

func TestSessionRepository_CRUD(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	repo := NewSessionDBRepository(tx)
	ctx := context.Background()

	testToken := "test-session-token-" + uuid.New().String()
	testUserID := uuid.New().String()

	// 1. Create Session
	session := &model.Sessions{
		UserID:     testUserID,
		Token:      testToken,
		Provider:   model.ProviderLocal,
		DeviceInfo: "test-device",
		IpAddress:  "127.0.0.1",
		UserAgent:  "go-test-agent",
		IsActive:   true,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	created, err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero session ID, got 0")
	}

	// 2. Get Session by ID
	fetched, err := repo.GetSessionByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get session by ID: %v", err)
	}
	if fetched == nil || fetched.Token != testToken {
		t.Fatalf("expected session with token %s, got %+v", testToken, fetched)
	}

	// 3. Get Session by Token
	byToken, err := repo.GetSessionByToken(ctx, testToken)
	if err != nil {
		t.Fatalf("failed to get session by Token: %v", err)
	}
	if byToken == nil || byToken.ID != created.ID {
		t.Fatalf("expected session ID %d, got %+v", created.ID, byToken)
	}

	// 4. Revoke Session
	if err := repo.RevokeSession(ctx, created.ID); err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}

	revoked, err := repo.GetSessionByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get revoked session: %v", err)
	}
	if revoked.IsActive {
		t.Fatalf("expected session to be inactive after revoke, got isActive=true")
	}
}
