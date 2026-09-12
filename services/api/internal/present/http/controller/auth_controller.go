package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

type AuthController struct {
	*baseController
	authService *service.AuthService
}

func NewAuthController(validate *validator.Validate, authService *service.AuthService) *AuthController {
	return &AuthController{
		baseController: NewBaseController(validate),
		authService:    authService,
	}
}

// Register
// @Summary Register user
// @Description Register a new user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register credentials"
// @Success 200 {object} dto.RegisterResponse
// @Router /api/v1/auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
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
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		SessionID:    tokens.SessionID,
		UserID:       tokens.User.UserID,
	}

	ctrl.Success(c, res)
}

// Login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Router /api/v1/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
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
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		SessionID:    tokens.SessionID,
		UserID:       tokens.User.UserID,
	}

	ctrl.Success(c, res)
}

// RefreshToken
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.RefreshTokenResponse
// @Router /api/v1/auth/refresh [post]
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tokens, err := ctrl.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

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
		c.SetCookie("auth_token", "", -1, "/", "", false, false)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
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
	if config.AppConfig != nil && config.AppConfig.Server.FrontendURL != "" {
		frontendURL = config.AppConfig.Server.FrontendURL
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
	isProduction := config.AppConfig != nil && config.AppConfig.Server.Production
	accessExpireSec := int(ctrl.authService.GetAccessTokenExpire().Seconds())
	refreshExpireSec := int(ctrl.authService.GetRefreshTokenExpire().Seconds())

	c.SetCookie("auth_token", tokens.AccessToken, accessExpireSec, "/", "", isProduction, false)
	c.SetCookie("refresh_token", tokens.RefreshToken, refreshExpireSec, "/", "", isProduction, true)

	// Clean redirect without tokens in query parameters
	redirectURL := fmt.Sprintf("%s/auth/callback", frontendURL)
	c.Redirect(http.StatusFound, redirectURL)
}
