package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/response"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type UserController struct {
	*baseController
	userService *service.UserService
	validate    *validator.Validate
}

func NewUserController(
	validate *validator.Validate,
	userService *service.UserService,
) *UserController {
	return &UserController{
		baseController: NewBaseController(validate),
		userService:    userService,
	}
}

func (c *UserController) GetProfile(ctx *gin.Context) {
	res, err := c.userService.GetProfile(ctx.Request.Context())
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *UserController) ListUsers(ctx *gin.Context) {
	pagination, err := c.GetPaginationParams(ctx)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	res, err := c.userService.ListUsers(ctx.Request.Context(), &dto.ListUsersRequest{
		PaginationRequest: *pagination,
		Filter: dto.ListUsersFilter{
			UserName:    ctx.Query("user_name"),
			FullName:    ctx.Query("full_name"),
			Email:       ctx.Query("email"),
			Gender:      ctx.Query("gender"),
			Phone:       ctx.Query("phone"),
			DateOfBirth: ctx.Query("date_of_birth"),
		},
		Search: ctx.Query("search"),
	})
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *UserController) RegisterRoutes(private *gin.RouterGroup) {
	userGroup := private.Group("/users")
	{
		userGroup.GET("/me", c.GetProfile)
		userGroup.GET("/", c.ListUsers)
	}
}
