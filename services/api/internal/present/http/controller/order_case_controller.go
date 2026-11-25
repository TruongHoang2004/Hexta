package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/ordercase/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// OrderCaseController serves cancellation/return cases on confirmed orders. Authorization is decided by the service.
type OrderCaseController struct {
	*baseController
	cases *service.OrderCaseService
}

func NewOrderCaseController(validate *validator.Validate, cases *service.OrderCaseService) *OrderCaseController {
	return &OrderCaseController{NewBaseController(validate), cases}
}

// scope validates the route IDs and returns tenant, trusted actor and order IDs.
func (ctrl *OrderCaseController) scope(c *gin.Context) (string, string, string, bool) {
	tenant, order := c.Param("tenantID"), c.Param("id")
	for _, id := range []string{tenant, order} {
		if _, err := uuid.Parse(id); err != nil {
			ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid resource ID"))
			return "", "", "", false
		}
	}
	actor := ""
	if info, ok := common.GetAuthInfo(c.Request.Context()); ok {
		actor = info.UserID
	}
	return tenant, actor, order, true
}

func (ctrl *OrderCaseController) bind(c *gin.Context, req any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
	if err := c.ShouldBindJSON(req); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid JSON request"))
		return false
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), req); err != nil {
		ctrl.ErrorData(c, err)
		return false
	}
	return true
}

// GetOrderCase
// @Summary Get the order's cancellation/return case
// @Description Returns the case with its current return lines and full history, or a null case when none is open. Any current workspace member may read it.
// @Tags Sales
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response[dto.OrderCaseLookupResponse]
// @Failure 400,401,403,404 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/orders/{id}/case [get]
func (ctrl *OrderCaseController) Get(c *gin.Context) {
	tenant, actor, order, ok := ctrl.scope(c)
	if !ok {
		return
	}
	value, err := ctrl.cases.Case(c.Request.Context(), tenant, actor, order)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Lookup(value))
}

// OpenOrderCase
// @Summary Open a cancellation or return case on a confirmed order
// @Description Only confirmed orders accept a case and each order has at most one. The caller becomes the handler unless the owner assigns another current member. The original order is not changed; no refund or stock movement happens.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Order ID"
// @Param request body dto.OpenOrderCaseRequest true "Case type, optional handler and note"
// @Success 200 {object} response.Response[dto.OrderCaseDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/orders/{id}/case [post]
func (ctrl *OrderCaseController) Open(c *gin.Context) {
	tenant, actor, order, ok := ctrl.scope(c)
	if !ok {
		return
	}
	var req dto.OpenOrderCaseRequest
	if !ctrl.bind(c, &req) {
		return
	}
	value, err := ctrl.cases.Open(c.Request.Context(), tenant, actor, order, httpmap.OpenInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Case(*value))
}

// UpdateOrderCase
// @Summary Record case progress and return lines
// @Description The assigned handler or the owner replaces the progress note and the full set of return lines on the current version. Each return quantity must not exceed the original line quantity. A sent-back case returns to open; a finalized case is read-only. Only the owner may change the handler.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Order ID"
// @Param request body dto.UpdateOrderCaseRequest true "Expected version, note, optional handler and lines"
// @Success 200 {object} response.Response[dto.OrderCaseDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/orders/{id}/case [put]
func (ctrl *OrderCaseController) Update(c *gin.Context) {
	tenant, actor, order, ok := ctrl.scope(c)
	if !ok {
		return
	}
	var req dto.UpdateOrderCaseRequest
	if !ctrl.bind(c, &req) {
		return
	}
	value, err := ctrl.cases.Update(c.Request.Context(), tenant, actor, order, httpmap.UpdateInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Case(*value))
}

// SendBackOrderCase
// @Summary Send an open case back to its handler (owner only)
// @Description The workspace owner returns an open case to the handler with a required note. Members are refused.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Order ID"
// @Param request body dto.OrderCaseDecisionRequest true "Expected version and reason"
// @Success 200 {object} response.Response[dto.OrderCaseDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/orders/{id}/case/send-back [post]
func (ctrl *OrderCaseController) SendBack(c *gin.Context) {
	tenant, actor, order, ok := ctrl.scope(c)
	if !ok {
		return
	}
	var req dto.OrderCaseDecisionRequest
	if !ctrl.bind(c, &req) {
		return
	}
	value, err := ctrl.cases.SendBack(c.Request.Context(), tenant, actor, order, httpmap.DecisionInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Case(*value))
}

// FinalizeOrderCase
// @Summary Finalize a case (owner only)
// @Description The workspace owner closes the case; it becomes read-only and its history stays on the order. A return case needs at least one return line. Members are refused. No refund or stock movement happens.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Order ID"
// @Param request body dto.OrderCaseDecisionRequest true "Expected version and optional note"
// @Success 200 {object} response.Response[dto.OrderCaseDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/orders/{id}/case/finalize [post]
func (ctrl *OrderCaseController) Finalize(c *gin.Context) {
	tenant, actor, order, ok := ctrl.scope(c)
	if !ok {
		return
	}
	var req dto.OrderCaseDecisionRequest
	if !ctrl.bind(c, &req) {
		return
	}
	value, err := ctrl.cases.Finalize(c.Request.Context(), tenant, actor, order, httpmap.DecisionInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Case(*value))
}
