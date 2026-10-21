package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	*baseController
	authService service.IAuthService
	config      *config.Config
}

func NewAuthController(validate *validator.Validate, authService *service.AuthService, cfg *config.Config) *AuthController {
	return &AuthController{
		baseController: NewBaseController(validate),
		authService:    authService,
		config:         cfg,
	}
}

// Register
// @Summary Register user
// @Description Register a new user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register credentials"
// @Success 200 {object} response.Response[dto.RegisterResponse]
// @Router /api/v1/auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	if !ctrl.allowBrowserOrigin(c) {
		return
	}
	var req dto.RegisterRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()
	deviceInfo := c.GetHeader("X-Device-Info")

	tokens, err := ctrl.authService.Register(c.Request.Context(), req.Email, req.Password, deviceInfo, ipAddress, userAgent)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.RegisterResponse{
		AccessToken: tokens.AccessToken,
		SessionID:   tokens.SessionID,
		UserID:      tokens.User.UserID,
	}

	ctrl.setRefreshCookie(c, tokens.RefreshToken, int(ctrl.authService.GetRefreshTokenExpire().Seconds()))
	ctrl.Success(c, res)
}

// Login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response[dto.LoginResponse]
// @Router /api/v1/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	if !ctrl.allowBrowserOrigin(c) {
		return
	}
	var req dto.LoginRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()
	deviceInfo := c.GetHeader("X-Device-Info")

	tokens, err := ctrl.authService.Login(c.Request.Context(), req.Email, req.Password, deviceInfo, ipAddress, userAgent)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.LoginResponse{
		AccessToken: tokens.AccessToken,
		SessionID:   tokens.SessionID,
		UserID:      tokens.User.UserID,
	}

	ctrl.setRefreshCookie(c, tokens.RefreshToken, int(ctrl.authService.GetRefreshTokenExpire().Seconds()))
	ctrl.Success(c, res)
}

// RefreshToken
// @Summary Refresh access token
// @Description Rotate the HttpOnly refresh cookie and return a new access token; no request body is required
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response[dto.RefreshTokenResponse]
// @Router /api/v1/auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	if !ctrl.allowBrowserOrigin(c) {
		return
	}
	refreshToken, cookieErr := c.Cookie("refresh_token")
	if cookieErr != nil || refreshToken == "" {
		ctrl.ErrorData(c, errors.ErrUnauthorized(c.Request.Context()).SetDetail("Refresh cookie required"))
		return
	}
	tokens, err := ctrl.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.RefreshTokenResponse{
		AccessToken: tokens.AccessToken,
	}

	ctrl.setRefreshCookie(c, tokens.RefreshToken, int(ctrl.authService.GetRefreshTokenExpire().Seconds()))
	ctrl.Success(c, res)
}

// Logout
// @Summary Logout user
// @Description Revoke the user's current or specified session (enforcing ownership)
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest false "Logout request"
// @Success 200
// @Router /api/v1/auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	if !ctrl.allowBrowserOrigin(c) {
		return
	}
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	authInfo, ok := common.GetAuthInfo(c.Request.Context())
	if !ok || authInfo == nil {
		ctrl.ErrorData(c, errors.ErrUnauthorized(c.Request.Context()).SetDetail("Unauthenticated"))
		return
	}

	targetSessionID := req.SessionID
	if targetSessionID == 0 {
		targetSessionID = authInfo.SessionID
	}

	err := ctrl.authService.Logout(c.Request.Context(), authInfo.UserID, targetSessionID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	// Clear cookies if current session was revoked
	if targetSessionID == authInfo.SessionID {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("auth_token", "", -1, "/", "", ctrl.isProduction(), false)
		ctrl.setRefreshCookie(c, "", -1)
	}

	ctrl.Success(c, gin.H{"message": "Logged out successfully"})
}

// GoogleLogin
// @Summary Google OAuth login
// @Description Redirects to Google consent screen with secure CSRF state
// @Tags Auth
// @Success 302
// @Router /api/v1/auth/google/login [get]
func (ctrl *AuthController) GoogleLogin(c *gin.Context) {
	url, err := ctrl.authService.GoogleLogin(c.Request.Context())
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

// GoogleCallback
// @Summary Google OAuth callback
// @Description Handles Google OAuth callback, validates state, and redirects to frontend with secure cookies
// @Tags Auth
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth CSRF state"
// @Success 302
// @Router /api/v1/auth/google/callback [get]
func (ctrl *AuthController) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	frontendURL := "http://localhost:3000"
	if ctrl.config != nil && ctrl.config.Server.FrontendURL != "" {
		frontendURL = ctrl.config.Server.FrontendURL
	}

	if code == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=missing_code", frontendURL))
		return
	}
	if state == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=missing_state", frontendURL))
		return
	}

	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()
	deviceInfo := c.GetHeader("X-Device-Info")

	tokens, err := ctrl.authService.GoogleCallback(c.Request.Context(), code, state, deviceInfo, ipAddress, userAgent)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=auth_failed", frontendURL))
		return
	}

	// Deliver tokens securely via cookies to prevent leaking tokens in URL query strings
	isProduction := ctrl.isProduction()
	accessExpireSec := int(ctrl.authService.GetAccessTokenExpire().Seconds())
	refreshExpireSec := int(ctrl.authService.GetRefreshTokenExpire().Seconds())

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", tokens.AccessToken, accessExpireSec, "/", "", isProduction, false)
	ctrl.setRefreshCookie(c, tokens.RefreshToken, refreshExpireSec)

	// Clean redirect without tokens in query parameters
	redirectURL := fmt.Sprintf("%s/auth/callback", frontendURL)
	c.Redirect(http.StatusFound, redirectURL)
}

func (ctrl *AuthController) isProduction() bool {
	return ctrl.config != nil && ctrl.config.Server.Production
}

func (ctrl *AuthController) setRefreshCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("refresh_token", token, maxAge, "/", "", ctrl.isProduction(), true)
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

// SameSite is defense in depth: explicitly reject untrusted browser origins,
// including same-site sibling origins that would otherwise receive cookies.
func (ctrl *AuthController) allowBrowserOrigin(c *gin.Context) bool {
	origin := c.GetHeader("Origin")
	if origin == "" && (c.GetHeader("Sec-Fetch-Site") == "" || c.GetHeader("Sec-Fetch-Site") == "same-origin") {
		return true
	}
	allowed := []string{}
	if ctrl.config != nil {
		if parsed, err := url.Parse(ctrl.config.Server.FrontendURL); err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") {
			allowed = append(allowed, parsed.Scheme+"://"+parsed.Host)
		}
	}
	if !ctrl.isProduction() {
		allowed = append(allowed, "http://localhost:3000", "http://127.0.0.1:3000")
	}
	for _, candidate := range allowed {
		if strings.EqualFold(origin, candidate) {
			return true
		}
	}
	ctrl.ErrorData(c, errors.ErrForbidden(c.Request.Context()).SetDetail("Untrusted browser origin"))
	return false
}
