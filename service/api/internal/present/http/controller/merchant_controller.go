package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/response"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type MerchantController struct {
	*baseController
	userService     *service.UserService
	merchantService *service.MerchantService
}

func NewMerchantController(
	validate *validator.Validate,
	merchantService *service.MerchantService,
) *MerchantController {
	return &MerchantController{
		baseController:  NewBaseController(validate),
		merchantService: merchantService,
	}
}

func (c *MerchantController) CreateMerchant(ctx *gin.Context) {
	var req dto.CreateMerchantRequest
	if err := c.BindAndValidateRequest(ctx, &req); err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	res, err := c.merchantService.CreateMerchant(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *MerchantController) ListMerchants(ctx *gin.Context) {
	pagination, err := c.GetPaginationParams(ctx)
	if err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	req := dto.ListMerchantsRequest{
		PaginationRequest: *pagination,
		Search:            ctx.Query("search"),
		Filter: dto.ListMerchantFilter{
			Status:      ctx.Query("status"),
			Name:        ctx.Query("name"),
			Description: ctx.Query("description"),
			Phone:       ctx.Query("phone"),
			Email:       ctx.Query("email"),
		},
	}

	res, err := c.merchantService.ListMerchants(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *MerchantController) UpdateMerchant(ctx *gin.Context) {
	var req dto.UpdateMerchantRequest
	if err := c.BindAndValidateRequest(ctx, &req); err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	res, err := c.merchantService.UpdateMerchant(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(err.HTTPStatus, errors.ConvertErrorToResponse(err))
		return
	}

	c.Success(ctx, response.NewSuccessResponse(res, nil))
}

func (c *MerchantController) RegisterRoutes(private *gin.RouterGroup) {
	merchantGroup := private.Group("/merchants")
	{
		merchantGroup.POST("/", c.CreateMerchant)
		merchantGroup.GET("/", c.ListMerchants)
		merchantGroup.PUT("/", c.UpdateMerchant)
	}
}
