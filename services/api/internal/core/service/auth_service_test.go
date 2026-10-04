package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/stretchr/testify/assert"
)

type mockIdentityRepo struct {
	identities map[string]*model.AuthIdentities
}

func newMockIdentityRepo() *mockIdentityRepo {
	return &mockIdentityRepo{
		identities: make(map[string]*model.AuthIdentities),
	}
}

func (m *mockIdentityRepo) GetCredentialByIdentifier(ctx context.Context, identifier string, provider model.Provider) (*model.AuthIdentities, *errors.Error) {
	key := string(provider) + ":" + identifier
	if ident, ok := m.identities[key]; ok {
		return ident, nil
	}
	return nil, nil
}

func (m *mockIdentityRepo) GetFirstByIdentifier(ctx context.Context, identifier string) (*model.AuthIdentities, *errors.Error) {
	for _, ident := range m.identities {
		if ident.Identifier == identifier {
			return ident, nil
		}
	}
	return nil, nil
}

func (m *mockIdentityRepo) CreateIdentity(ctx context.Context, authIdentity *model.AuthIdentities) (*model.AuthIdentities, *errors.Error) {
	key := string(authIdentity.Provider) + ":" + authIdentity.Identifier
	m.identities[key] = authIdentity
	return authIdentity, nil
}

type mockSessionRepo struct {
	sessions map[int64]*model.Sessions
	nextID   int64
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[int64]*model.Sessions),
		nextID:   1,
	}
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error) {
	session.ID = m.nextID
	m.nextID++
	m.sessions[session.ID] = session
	return session, nil
}

func (m *mockSessionRepo) GetSessionByID(ctx context.Context, id int64) (*model.Sessions, *errors.Error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
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
	if s, ok := m.sessions[id]; ok {
		s.IsActive = false
		s.RevokedAt = time.Now()
		return nil
	}
	return errors.ErrNotFound(ctx, "Session", "not found")
}

func createTestConfig(accessSecret, refreshSecret string, accessExpireSec, refreshExpireSec int) *config.Config {
	cfg := &config.Config{}
	cfg.JWT.AccessTokenSecret = accessSecret
	cfg.JWT.AccessTokenExpire = accessExpireSec
	cfg.JWT.RefreshTokenSecret = refreshSecret
	cfg.JWT.RefreshTokenExpire = refreshExpireSec
	return cfg
}

func TestJWTTokenSigningAndConfigInjection(t *testing.T) {
	cfg := createTestConfig("custom-access-secret-1234567890", "custom-refresh-secret-0987654321", 900, 86400)
	identityRepo := newMockIdentityRepo()
	sessionRepo := newMockSessionRepo()

	authSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, cfg)

	assert.Equal(t, 900*time.Second, authSvc.GetAccessTokenExpire())
	assert.Equal(t, 86400*time.Second, authSvc.GetRefreshTokenExpire())

	ctx := context.Background()
	tokens, err := authSvc.Register(ctx, "test@example.com", "secretpassword", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, int64(1), tokens.SessionID)

	// Validate access token
	authInfo, err := authSvc.ValidateAccessToken(ctx, tokens.AccessToken)
	assert.Nil(t, err)
	assert.Equal(t, tokens.User.UserID, authInfo.UserID)
	assert.Equal(t, tokens.SessionID, authInfo.SessionID)

	// Attempting to validate refresh token as an access token must fail
	_, err = authSvc.ValidateAccessToken(ctx, tokens.RefreshToken)
	assert.NotNil(t, err)

	// Refresh token flow
	refreshed, err := authSvc.RefreshToken(ctx, tokens.RefreshToken)
	assert.Nil(t, err)
	assert.NotEmpty(t, refreshed.AccessToken)
	assert.NotEmpty(t, refreshed.RefreshToken)

	// Attempting to refresh using an access token must fail
	_, err = authSvc.RefreshToken(ctx, tokens.AccessToken)
	assert.NotNil(t, err)
}

func TestTokenExpiration(t *testing.T) {
	// Configure 1-second expiration
	cfg := createTestConfig("access-secret-exp", "refresh-secret-exp", 1, 1)
	identityRepo := newMockIdentityRepo()
	sessionRepo := newMockSessionRepo()

	authSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, cfg)
	ctx := context.Background()

	tokens, err := authSvc.Register(ctx, "expire@example.com", "password123", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)

	// Immediately valid
	info, err := authSvc.ValidateAccessToken(ctx, tokens.AccessToken)
	assert.Nil(t, err)
	assert.NotNil(t, info)

	// Wait for expiration
	time.Sleep(1200 * time.Millisecond)

	// Now expired
	_, err = authSvc.ValidateAccessToken(ctx, tokens.AccessToken)
	assert.NotNil(t, err)

	// Refresh token should also expire
	_, err = authSvc.RefreshToken(ctx, tokens.RefreshToken)
	assert.NotNil(t, err)
}

func TestInvalidAndTamperedTokenRejection(t *testing.T) {
	cfg := createTestConfig("correct-secret-12345", "correct-refresh-secret", 900, 86400)
	identityRepo := newMockIdentityRepo()
	sessionRepo := newMockSessionRepo()

	authSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, cfg)
	ctx := context.Background()

	// 1. Malformed token
	_, err := authSvc.ValidateAccessToken(ctx, "invalid.malformed.token")
	assert.NotNil(t, err)

	// 2. Token signed with different secret
	foreignCfg := createTestConfig("attacker-secret-9999", "attacker-refresh-secret", 900, 86400)
	foreignSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, foreignCfg)

	tokens, err := foreignSvc.Register(ctx, "attacker@example.com", "password123", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)

	// Validation against real service must fail
	_, err = authSvc.ValidateAccessToken(ctx, tokens.AccessToken)
	assert.NotNil(t, err)
}

func TestOAuthStateGenerationAndCSRFValidation(t *testing.T) {
	cfg := createTestConfig("secret", "secret", 900, 86400)
	identityRepo := newMockIdentityRepo()
	sessionRepo := newMockSessionRepo()

	authSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, cfg)
	ctx := context.Background()

	// 1. Test GoogleLogin generates a URL with valid state parameter
	loginURL, err := authSvc.GoogleLogin(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, loginURL)
	assert.Contains(t, loginURL, "state=")

	// 2. Extract state from URL and validate
	// Empty state must fail
	err = authSvc.ValidateAndConsumeOAuthState(ctx, "")
	assert.NotNil(t, err)

	// Non-existent state must fail
	err = authSvc.ValidateAndConsumeOAuthState(ctx, "non-existent-state")
	assert.NotNil(t, err)

	// Extract state from loginURL
	u, parseErr := time.ParseDuration("1s")
	assert.Nil(t, parseErr)
	_ = u

	// Let's test single-use state consumption
	// Re-run GoogleLogin to get a known state
	var capturedState string
	// We can parse state from URL
	for i := 0; i < len(loginURL)-6; i++ {
		if loginURL[i:i+6] == "state=" {
			rest := loginURL[i+6:]
			for j := 0; j < len(rest); j++ {
				if rest[j] == '&' {
					capturedState = rest[:j]
					break
				}
			}
			if capturedState == "" {
				capturedState = rest
			}
			break
		}
	}
	assert.NotEmpty(t, capturedState)

	// First consumption must succeed
	err = authSvc.ValidateAndConsumeOAuthState(ctx, capturedState)
	assert.Nil(t, err)

	// Second consumption (replay attack) must fail
	err = authSvc.ValidateAndConsumeOAuthState(ctx, capturedState)
	assert.NotNil(t, err)
}

func TestLogoutEnforcesSessionOwnership(t *testing.T) {
	cfg := createTestConfig("secret", "secret", 900, 86400)
	identityRepo := newMockIdentityRepo()
	sessionRepo := newMockSessionRepo()

	authSvc := service.NewAuthServiceWithRepos(identityRepo, sessionRepo, nil, cfg)
	ctx := context.Background()

	// Create user 1 and session 1
	user1Tokens, err := authSvc.Register(ctx, "user1@example.com", "password123", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)
	session1ID := user1Tokens.SessionID

	// Create user 2 and session 2
	user2Tokens, err := authSvc.Register(ctx, "user2@example.com", "password123", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)
	session2ID := user2Tokens.SessionID

	// 1. User 1 tries to revoke User 2's session -> Must fail with ErrForbidden
	err = authSvc.Logout(ctx, user1Tokens.User.UserID, session2ID)
	assert.NotNil(t, err)
	assert.Equal(t, 403, err.GetHttpStatus())

	// Session 2 must still be active
	s2, dbErr := sessionRepo.GetSessionByID(ctx, session2ID)
	assert.Nil(t, dbErr)
	assert.True(t, s2.IsActive)

	// 2. User 1 tries to revoke non-existent session -> Must fail with ErrNotFound
	err = authSvc.Logout(ctx, user1Tokens.User.UserID, 99999)
	assert.NotNil(t, err)
	assert.Equal(t, 404, err.GetHttpStatus())

	// 3. User 1 revokes their own session -> Must succeed
	err = authSvc.Logout(ctx, user1Tokens.User.UserID, session1ID)
	assert.Nil(t, err)

	// Session 1 is now inactive
	s1, dbErr := sessionRepo.GetSessionByID(ctx, session1ID)
	assert.Nil(t, dbErr)
	assert.False(t, s1.IsActive)

	// Subsequent ValidateAccessToken using User 1's token must fail since session is revoked
	_, err = authSvc.ValidateAccessToken(ctx, user1Tokens.AccessToken)
	assert.NotNil(t, err)
}
