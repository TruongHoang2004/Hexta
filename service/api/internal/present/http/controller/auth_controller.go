package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/common/log"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/response"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type AuthController struct {
	*baseController
	authService *service.AuthService
}

func NewAuthController(
	validate *validator.Validate,
	authService *service.AuthService,
) *AuthController {
	return &AuthController{
		baseController: NewBaseController(validate),
		authService:    authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with the provided details
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 200 {object} response.Response[dto.RegisterResponse]
// @Failure 400 {object} errors.ErrorResponse
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := c.BindAndValidateRequest(ctx, &req); err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	res, err := c.authService.Register(ctx, &req)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return access/refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response[dto.LoginResponse]
// @Failure 401 {object} errors.ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := c.BindAndValidateRequest(ctx, &req); err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	ip := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()

	res, err := c.authService.Login(ctx, &req, ip, userAgent)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

// ValidateToken godoc
// @Summary Validate access token
// @Description Check if the provided access token is valid
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[any]
// @Failure 401 {object} errors.ErrorResponse
// @Router /auth/validate-token [get]
func (c *AuthController) ValidateToken(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, errors.ConvertErrorToResponse(errors.ErrUnauthorized(ctx)))
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	log.Info(ctx, "token: %s", token)

	res, err := c.authService.ValidateToken(ctx, token)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *AuthController) RegisterRoutes(public *gin.RouterGroup, private *gin.RouterGroup) {
	authGroup := public.Group("/auth")
	{
		authGroup.POST("/register", c.Register)
		authGroup.POST("/login", c.Login)
		authGroup.GET("/validate-token", c.ValidateToken)
		authGroup.POST("/refresh", c.RefreshToken)
	}
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Use refresh token to get a new access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response[dto.RefreshTokenResponse]
// @Failure 401 {object} errors.ErrorResponse
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.BindAndValidateRequest(ctx, &req); err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	res, err := c.authService.RefreshToken(ctx, &req)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}
