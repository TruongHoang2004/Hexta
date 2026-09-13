package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ecommercehub1/api/common"
	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/present/http/middleware"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type mockIdentityRepo struct{}

func (m *mockIdentityRepo) GetCredentialByIdentifier(ctx context.Context, identifier string, provider model.Provider) (*model.AuthIdentities, *errors.Error) {
	return nil, nil
}

func (m *mockIdentityRepo) CreateIdentity(ctx context.Context, authIdentity *model.AuthIdentities) (*model.AuthIdentities, *errors.Error) {
	return authIdentity, nil
}

type mockSessionRepo struct {
	sessions map[int64]*model.Sessions
	nextID   int64
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error) {
	m.nextID++
	session.ID = m.nextID
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
	return nil, nil
}

func (m *mockSessionRepo) RevokeSession(ctx context.Context, id int64) *errors.Error {
	if s, ok := m.sessions[id]; ok {
		s.IsActive = false
		return nil
	}
	return errors.ErrNotFound(ctx, "Session", "not found")
}

func setupTestRouter(authSvc *service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authMiddleware := middleware.NewAuthMiddleware(authSvc)

	protected := r.Group("/api/v1/protected")
	protected.Use(authMiddleware.Authenticate())
	{
		protected.GET("/profile", func(c *gin.Context) {
			authInfo, ok := common.GetAuthInfo(c.Request.Context())
			if !ok || authInfo == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "missing auth info"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"user_id":    authInfo.UserID,
				"session_id": authInfo.SessionID,
			})
		})
	}

	return r
}

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.AccessTokenSecret = "test-middleware-secret-key-12345"
	cfg.JWT.AccessTokenExpire = 900
	cfg.JWT.RefreshTokenSecret = "test-middleware-refresh-secret-12345"
	cfg.JWT.RefreshTokenExpire = 86400

	sessionRepo := &mockSessionRepo{
		sessions: make(map[int64]*model.Sessions),
	}
	authSvc := service.NewAuthServiceWithRepos(&mockIdentityRepo{}, sessionRepo, nil, cfg)
	router := setupTestRouter(authSvc)

	// Create an active session and valid token
	ctx := context.Background()
	tokens, err := authSvc.Register(ctx, "user@example.com", "password123", "device", "127.0.0.1", "agent")
	assert.Nil(t, err)

	t.Run("Missing Token Returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Bearer Format Returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
		req.Header.Set("Authorization", "Basic invalidtoken")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid Bearer Token Returns 200 and AuthInfo Context", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), tokens.User.UserID)
	})

	t.Run("Valid Cookie Fallback Returns 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: tokens.AccessToken,
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), tokens.User.UserID)
	})

	t.Run("Revoked Session Returns 401", func(t *testing.T) {
		// Revoke session
		_ = sessionRepo.RevokeSession(ctx, tokens.SessionID)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/protected/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
