package service

import (
	"context"
	"testing"
	"time"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type mockIdentityRepo struct {
	identities []*model.AuthIdentities
}

func (m *mockIdentityRepo) GetCredentialByIdentifier(ctx context.Context, identifier string, provider model.Provider) (*model.AuthIdentities, *errors.Error) {
	for _, id := range m.identities {
		if id.Identifier == identifier && id.Provider == provider {
			return id, nil
		}
	}
	return nil, nil
}

func (m *mockIdentityRepo) GetFirstByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentities, *errors.Error) {
	for _, id := range m.identities {
		if id.Identifier == identifier {
			return id, nil
		}
	}
	return nil, nil
}

func (m *mockIdentityRepo) CreateIdentity(ctx context.Context, authIdentity *model.AuthIdentities) (*model.AuthIdentities, *errors.Error) {
	authIdentity.ID = int64(len(m.identities) + 1)
	m.identities = append(m.identities, authIdentity)
	return authIdentity, nil
}

type mockSessionRepo struct {
	sessions []*model.Sessions
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error) {
	session.ID = int64(len(m.sessions) + 1)
	m.sessions = append(m.sessions, session)
	return session, nil
}

func (m *mockSessionRepo) GetSessionByID(ctx context.Context, id int64) (*model.Sessions, *errors.Error) {
	for _, s := range m.sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockSessionRepo) GetSessionByToken(ctx context.Context, token string) (*model.Sessions, *errors.Error) {
	for _, s := range m.sessions {
		if s.Token == token {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockSessionRepo) RevokeSession(ctx context.Context, id int64) *errors.Error {
	for _, s := range m.sessions {
		if s.ID == id {
			s.IsActive = false
			s.RevokedAt = time.Now()
			return nil
		}
	}
	return nil
}

func newTestAuthService(identities []*model.AuthIdentities) (*AuthService, *mockIdentityRepo, *mockSessionRepo) {
	idRepo := &mockIdentityRepo{identities: identities}
	sessRepo := &mockSessionRepo{}
	svc := &AuthService{
		baseService:  NewBaseService(),
		identityRepo: idRepo,
		sessionRepo:  sessRepo,
		jwtSecret:    []byte("test-secret-key-12345"),
	}
	return svc, idRepo, sessRepo
}

func TestAuthService_Register_Success(t *testing.T) {
	svc, idRepo, _ := newTestAuthService(nil)
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "test@example.com", "password123", "mobile", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
	if len(idRepo.identities) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(idRepo.identities))
	}

	saved := idRepo.identities[0]
	if saved.Password == nil {
		t.Fatal("expected password to be saved for local user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*saved.Password), []byte("password123")); err != nil {
		t.Fatalf("password was not hashed properly: %v", err)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	hStr := string(hashed)
	svc, _, _ := newTestAuthService([]*model.AuthIdentities{
		{
			UserID:     "user-1",
			Provider:   model.ProviderLocal,
			Identifier: "test@example.com",
			Password:   &hStr,
		},
	})
	ctx := context.Background()

	_, err := svc.Register(ctx, "test@example.com", "newpassword", "device", "127.0.0.1", "agent")
	if err == nil {
		t.Fatal("expected conflict error for duplicate email")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	hStr := string(hashed)
	svc, _, _ := newTestAuthService([]*model.AuthIdentities{
		{
			UserID:     "user-1",
			Provider:   model.ProviderLocal,
			Identifier: "login@example.com",
			Password:   &hStr,
		},
	})
	ctx := context.Background()

	tokens, err := svc.Login(ctx, "login@example.com", "secret123", "device", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("unexpected error on login: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	hStr := string(hashed)
	svc, _, _ := newTestAuthService([]*model.AuthIdentities{
		{
			UserID:     "user-1",
			Provider:   model.ProviderLocal,
			Identifier: "login@example.com",
			Password:   &hStr,
		},
	})
	ctx := context.Background()

	_, err := svc.Login(ctx, "login@example.com", "wrongpassword", "device", "127.0.0.1", "agent")
	if err == nil {
		t.Fatal("expected error on invalid password")
	}
}

func TestAuthService_Login_NilPassword_OAuthAccount(t *testing.T) {
	// An account created via Google OAuth with no local password (Password is nil)
	svc, _, _ := newTestAuthService([]*model.AuthIdentities{
		{
			UserID:     "user-oauth-1",
			Provider:   model.ProviderLocal, // or local identity without password
			Identifier: "oauth@example.com",
			Password:   nil,
		},
	})
	ctx := context.Background()

	// Attempting to log in with password should return unauthorized safely without panicking
	_, err := svc.Login(ctx, "oauth@example.com", "any-password", "device", "127.0.0.1", "agent")
	if err == nil {
		t.Fatal("expected unauthorized error for nil password")
	}
}

func TestAuthService_AccountLinking_Logic(t *testing.T) {
	ctx := context.Background()
	existingUserID := "existing-user-uuid-1234"
	localPasswordHash := "some-hashed-password"

	// Pre-existing local identity
	localIdentity := &model.AuthIdentities{
		UserID:     existingUserID,
		Provider:   model.ProviderLocal,
		Identifier: "shared@example.com",
		Password:   &localPasswordHash,
	}

	idRepo := &mockIdentityRepo{identities: []*model.AuthIdentities{localIdentity}}

	// Verify that querying by Google provider returns nil
	googleIdentity, err := idRepo.GetCredentialByIdentifier(ctx, "shared@example.com", model.ProviderGoogle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if googleIdentity != nil {
		t.Fatal("expected google identity to be nil initially")
	}

	// Verify that querying existing identity by identifier across providers returns localIdentity
	foundExisting, err := idRepo.GetFirstByIdentifier(ctx, "shared@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if foundExisting == nil {
		t.Fatal("expected to find existing identity for account linking")
	}
	if foundExisting.UserID != existingUserID {
		t.Fatalf("expected UserID %s, got %s", existingUserID, foundExisting.UserID)
	}

	// Link Google provider to existing UserID with Password: nil
	newGoogleIdentity := &model.AuthIdentities{
		UserID:     foundExisting.UserID,
		Provider:   model.ProviderGoogle,
		Identifier: "shared@example.com",
		Password:   nil,
	}
	createdGoogle, err := idRepo.CreateIdentity(ctx, newGoogleIdentity)
	if err != nil {
		t.Fatalf("failed to create linked google identity: %v", err)
	}
	if createdGoogle.UserID != existingUserID {
		t.Fatalf("expected linked UserID %s, got %s", existingUserID, createdGoogle.UserID)
	}
	if createdGoogle.Password != nil {
		t.Fatalf("expected nil password for OAuth identity, got %v", *createdGoogle.Password)
	}

	// Verify both identities exist under same identifier with distinct providers and identical UserID
	loc, _ := idRepo.GetCredentialByIdentifier(ctx, "shared@example.com", model.ProviderLocal)
	goo, _ := idRepo.GetCredentialByIdentifier(ctx, "shared@example.com", model.ProviderGoogle)

	if loc == nil || goo == nil {
		t.Fatal("both local and google identities must exist")
	}
	if loc.UserID != goo.UserID {
		t.Fatalf("user_id mismatch between linked accounts: %s vs %s", loc.UserID, goo.UserID)
	}
}
