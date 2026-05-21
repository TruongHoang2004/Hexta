package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/order/internal/common/log"
	"gitlab.com/ecommercehub1/order/internal/core/service"
	"gitlab.com/ecommercehub1/order/internal/present/http/dto"
)

type OrderController struct {
	orderService service.IOrderService
}

func NewOrderController(orderService service.IOrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

func (c *OrderController) SetupRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	{
		orders.POST("", c.CreateOrder)
	}
}

// @Summary Create Order
// @Description Create a new order (published to Kafka for high throughput)
// @Tags Order
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "Order data"
// @Success 200 {object} dto.OrderResponse
// @Router /api/v1/orders [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn(ctx.Request.Context(), "invalid request payload", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := c.orderService.CreateOrder(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(err.GetHttpStatus(), gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
