package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sharedErrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

type cookieAuthService struct {
	service.IAuthService
	refreshInput  string
	logoutSession int64
}

func cookieTokens() *service.AuthTokens {
	return &service.AuthTokens{AccessToken: "fixture-access", RefreshToken: "fixture-refresh", SessionID: 42, User: &model.AuthIdentities{UserID: "fixture-user"}}
}
func (*cookieAuthService) Register(context.Context, string, string, string, string, string) (*service.AuthTokens, *sharedErrors.Error) {
	return cookieTokens(), nil
}
func (*cookieAuthService) Login(context.Context, string, string, string, string, string) (*service.AuthTokens, *sharedErrors.Error) {
	return cookieTokens(), nil
}
func (s *cookieAuthService) RefreshToken(_ context.Context, token string) (*service.AuthTokens, *sharedErrors.Error) {
	s.refreshInput = token
	if token != "fixture-refresh" {
		return nil, sharedErrors.ErrUnauthorized(context.Background())
	}
	result := cookieTokens()
	result.RefreshToken = "fixture-refreshed"
	return result, nil
}
func (s *cookieAuthService) Logout(_ context.Context, _ string, session int64) *sharedErrors.Error {
	s.logoutSession = session
	return nil
}
func (*cookieAuthService) GoogleCallback(context.Context, string, string, string, string, string) (*service.AuthTokens, *sharedErrors.Error) {
	return cookieTokens(), nil
}
func (*cookieAuthService) GetAccessTokenExpire() time.Duration  { return 15 * time.Minute }
func (*cookieAuthService) GetRefreshTokenExpire() time.Duration { return 7 * 24 * time.Hour }

func cookieTestRouter(production bool) (*gin.Engine, *cookieAuthService) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Server.Production = production
	cfg.Server.FrontendURL = "https://app.example.test"
	config.AppConfig = cfg
	log.GetLogger()
	svc := &cookieAuthService{}
	ctrl := &AuthController{baseController: NewBaseController(validator.New()), authService: svc, config: cfg}
	r := gin.New()
	r.POST("/register", ctrl.Register)
	r.POST("/login", ctrl.Login)
	r.POST("/refresh", ctrl.RefreshToken)
	r.POST("/logout", func(c *gin.Context) {
		ctx := common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: "fixture-user", SessionID: 42})
		c.Request = c.Request.WithContext(ctx)
		ctrl.Logout(c)
	})
	r.GET("/callback", ctrl.GoogleCallback)
	return r, svc
}

func findRefreshCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			return cookie
		}
	}
	t.Fatal("refresh cookie was not set")
	return nil
}

func TestPasswordAuthUsesHttpOnlyRefreshCookie(t *testing.T) {
	for _, production := range []bool{false, true} {
		r, _ := cookieTestRouter(production)
		for _, path := range []string{"/register", "/login"} {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"user@example.test","password":"password123"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "https://app.example.test")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)
			cookie := findRefreshCookie(t, rec)
			require.True(t, cookie.HttpOnly)
			require.Equal(t, production, cookie.Secure)
			require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
			require.Equal(t, "/", cookie.Path)
			require.Empty(t, cookie.Domain)
			require.Positive(t, cookie.MaxAge)
			require.NotContains(t, rec.Body.String(), "refresh_token")
			require.NotContains(t, rec.Body.String(), cookie.Value)
			var body struct {
				Data struct {
					AccessToken string `json:"access_token"`
					SessionID   int64  `json:"session_id"`
					UserID      string `json:"user_id"`
				}
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, "fixture-access", body.Data.AccessToken)
			require.Equal(t, int64(42), body.Data.SessionID)
			require.Equal(t, "fixture-user", body.Data.UserID)
			require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
		}
	}
}

func TestRefreshReadsCookieAndReplacesIt(t *testing.T) {
	r, svc := cookieTestRouter(true)
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.Header.Set("Origin", "https://app.example.test")
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "fixture-refresh"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "fixture-refresh", svc.refreshInput)
	cookie := findRefreshCookie(t, rec)
	require.Equal(t, "fixture-refreshed", cookie.Value)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	require.NotContains(t, rec.Body.String(), "refresh_token")
	require.NotContains(t, rec.Body.String(), cookie.Value)
	// Supplying the credential in JSON cannot substitute for the protected cookie.
	req = httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"fixture-refresh"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	req = httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid"})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Empty(t, rec.Result().Cookies())
}

func TestCookieAuthenticationRejectsUntrustedBrowserOrigins(t *testing.T) {
	r, svc := cookieTestRouter(true)
	for _, origin := range []string{"https://attacker.example.test", "null", "https://app.example.test.attacker.test"} {
		for _, path := range []string{"/register", "/login", "/refresh", "/logout"} {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"user@example.test","password":"password123"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", origin)
			req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "fixture-refresh"})
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusForbidden, rec.Code)
			require.Empty(t, rec.Result().Cookies())
		}
	}
	for _, fetchSite := range []string{"same-site", "cross-site"} {
		req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		req.Header.Set("Sec-Fetch-Site", fetchSite)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "fixture-refresh"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusForbidden, rec.Code)
	}
	require.Empty(t, svc.refreshInput)
	require.Zero(t, svc.logoutSession)
}

func TestLogoutClearsOnlyCurrentSessionCookies(t *testing.T) {
	r, svc := cookieTestRouter(true)
	for _, body := range []string{`{"session_id":99}`, `{}`} {
		req := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		if strings.Contains(body, "99") {
			require.Equal(t, int64(99), svc.logoutSession)
			require.Empty(t, rec.Result().Cookies())
			continue
		}
		require.Equal(t, int64(42), svc.logoutSession)
		cookie := findRefreshCookie(t, rec)
		require.Empty(t, cookie.Value)
		require.Equal(t, -1, cookie.MaxAge)
		require.Equal(t, "/", cookie.Path)
		require.True(t, cookie.HttpOnly)
		require.True(t, cookie.Secure)
		require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	}
}

func TestOAuthUsesSameProtectedRefreshCookie(t *testing.T) {
	r, _ := cookieTestRouter(true)
	req := httptest.NewRequest(http.MethodGet, "/callback?code=fixture&state=fixture", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://app.example.test/auth/callback", rec.Header().Get("Location"))
	cookie := findRefreshCookie(t, rec)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	require.Equal(t, "/", cookie.Path)
	require.NotContains(t, rec.Header().Get("Location"), "fixture-refresh")
}
