package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
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

	// Optionally pass device info from header, for now empty string or extract from headers
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
// @Description Revoke the user's current session
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Logout request"
// @Success 200
// @Router /api/v1/auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	err := ctrl.authService.Logout(c.Request.Context(), req.SessionID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, gin.H{"message": "Logged out successfully"})
}
