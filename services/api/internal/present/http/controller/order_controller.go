package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
)

type OrderController struct {
	*baseController
	orderService service.IOrderService
}

func NewOrderController(validate *validator.Validate, orderService service.IOrderService) *OrderController {
	return &OrderController{
		baseController: NewBaseController(validate),
		orderService:  orderService,
	}
}

func (ctrl *OrderController) getTenantID(c *gin.Context) string {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	return tenantID
}

func (ctrl *OrderController) getAuthUserID(c *gin.Context) string {
	if info, ok := common.GetAuthInfo(c.Request.Context()); ok && info.UserID != "" {
		return info.UserID
	}
	return c.GetString("userID")
}

func toOrderResponse(order *model.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = dto.OrderItemResponse{
			ID:        item.ID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			LineTotal: item.LineTotal,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	return dto.OrderResponse{
		ID:            order.ID,
		TenantID:      order.TenantID,
		OrderNumber:   order.OrderNumber,
		CustomerID:    order.CustomerID,
		EmployeeID:    order.EmployeeID,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		Subtotal:      order.Subtotal,
		Discount:      order.Discount,
		TotalAmount:   order.TotalAmount,
		Source:        order.Source,
		DraftID:       order.DraftID,
		Notes:         order.Notes,
		Items:         items,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
}

// CreateOrder
// @Summary Create a new order
// @Description Create order draft or auto-confirm with stock reservation
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param request body dto.CreateOrderRequest true "Order creation payload"
// @Success 200 {object} response.Response[dto.OrderResponse]
// @Router /api/v1/orders [post]
func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = ctrl.getTenantID(c)
	}
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via request body or X-Tenant-ID header"))
		return
	}

	employeeID := req.EmployeeID
	if employeeID == nil {
		if uid := ctrl.getAuthUserID(c); uid != "" {
			employeeID = &uid
		}
	}

	items := make([]service.CreateOrderItemParam, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.CreateOrderItemParam{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	order, err := ctrl.orderService.CreateOrder(c.Request.Context(), service.CreateOrderParam{
		TenantID:    tenantID,
		CustomerID:  req.CustomerID,
		EmployeeID:  employeeID,
		Items:       items,
		Discount:    req.Discount,
		Source:      req.Source,
		DraftID:     req.DraftID,
		Notes:       req.Notes,
		AutoConfirm: req.AutoConfirm,
	})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, toOrderResponse(order))
}

// ListOrders
// @Summary List orders
// @Description Retrieve paginated orders for a workspace
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param status query string false "Filter by order status"
// @Param page query int false "Page number (default: 1)"
// @Param size query int false "Page size (default: 20)"
// @Success 200 {object} response.Response[dto.OrderListResponse]
// @Router /api/v1/orders [get]
func (ctrl *OrderController) ListOrders(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	var statusPtr *model.OrderStatus
	if statusQuery := c.Query("status"); statusQuery != "" {
		st := model.OrderStatus(statusQuery)
		statusPtr = &st
	}

	page := 1
	if pageQuery := c.Query("page"); pageQuery != "" {
		if parsed, err := strconv.Atoi(pageQuery); err == nil && parsed > 0 {
			page = parsed
		}
	}

	size := 20
	if sizeQuery := c.Query("size"); sizeQuery != "" {
		if parsed, err := strconv.Atoi(sizeQuery); err == nil && parsed > 0 {
			size = parsed
		}
	}

	offset := (page - 1) * size

	orders, total, err := ctrl.orderService.ListOrders(c.Request.Context(), tenantID, statusPtr, size, offset)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		orderResponses[i] = toOrderResponse(ord)
	}

	totalPages := 0
	if total > 0 && size > 0 {
		totalPages = int((total + int64(size) - 1) / int64(size))
	}

	ctrl.Success(c, dto.OrderListResponse{
		Items:      orderResponses,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	})
}

// GetOrder
// @Summary Get order details
// @Description Retrieve a specific order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response[dto.OrderResponse]
// @Router /api/v1/orders/{id} [get]
func (ctrl *OrderController) GetOrder(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("order id is required"))
		return
	}

	order, err := ctrl.orderService.GetOrder(c.Request.Context(), tenantID, orderID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, toOrderResponse(order))
}

// TransitionStatus
// @Summary Transition order lifecycle state
// @Description Transition order status following FSM rules (draft -> confirmed -> processing -> fulfilled/cancelled)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param id path string true "Order ID"
// @Param request body dto.UpdateOrderStatusRequest true "Target status"
// @Success 200 {object} response.Response[dto.OrderResponse]
// @Router /api/v1/orders/{id}/status [patch]
func (ctrl *OrderController) TransitionStatus(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("order id is required"))
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	order, err := ctrl.orderService.TransitionStatus(c.Request.Context(), tenantID, orderID, req.Status)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, toOrderResponse(order))
}

// UpdatePaymentStatus
// @Summary Update order payment status
// @Description Update payment status of an order
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param id path string true "Order ID"
// @Param request body dto.UpdatePaymentStatusRequest true "Payment status payload"
// @Success 200 {object} response.Response[dto.OrderResponse]
// @Router /api/v1/orders/{id}/payment [patch]
func (ctrl *OrderController) UpdatePaymentStatus(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("order id is required"))
		return
	}

	var req dto.UpdatePaymentStatusRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	order, err := ctrl.orderService.UpdatePaymentStatus(c.Request.Context(), tenantID, orderID, req.PaymentStatus)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, toOrderResponse(order))
}
