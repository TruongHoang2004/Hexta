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
