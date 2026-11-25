package controller

import (
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type MessagingQueryController struct {
	*baseController
	messaging *service.MessagingQueryService
}

func NewMessagingQueryController(validate *validator.Validate, s *service.MessagingQueryService) *MessagingQueryController {
	return &MessagingQueryController{NewBaseController(validate), s}
}
func (ctrl *MessagingQueryController) query(c *gin.Context) (types.MessagingQuery, bool) {
	input := dto.MessagingQuery{Limit: 20}
	if !ctrl.messagingResourceIDs(c) {
		return types.MessagingQuery{}, false
	}
	if err := c.ShouldBindQuery(&input); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid query"))
		return types.MessagingQuery{}, false
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), input); err != nil {
		ctrl.ErrorData(c, err)
		return types.MessagingQuery{}, false
	}
	return types.MessagingQuery{Cursor: input.Cursor, Filter: input.Filter, Limit: input.Limit}, true
}

// List
// @Summary List tenant conversations with current assignment
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param cursor query string false "UUID cursor"
// @Param limit query int false "Page size (1-100, default 20)"
// @Param filter query string false "all (includes closed turns), mine, unassigned, needs_human, waiting or bot (open turns only)"
// @Success 200 {object} response.Response[dto.MessagingPage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/conversations [get]
func (ctrl *MessagingQueryController) List(c *gin.Context) {
	input, ok := ctrl.query(c)
	if !ok {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Conversations(c.Request.Context(), tenant, actor, input)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Page(result))
}

// Summary
// @Summary Count conversations waiting for staff
// @Description Tenant-wide Desk notification state for polling. Requires current active tenant membership.
// @Tags Messaging
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.MessagingSummary]
// @Failure 400,401,403,404 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/conversations/summary [get]
func (ctrl *MessagingQueryController) Summary(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Summary(c.Request.Context(), tenant, actor)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Summary(result))
}

// Detail
// @Summary Read bounded conversation message history
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Resource ID"
// @Param cursor query string false "UUID cursor"
// @Param limit query int false "Page size (1-100, default 20)"
// @Success 200 {object} response.Response[dto.MessagingDetail]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/conversations/{id}/messages [get]
func (ctrl *MessagingQueryController) Detail(c *gin.Context) {
	input, ok := ctrl.query(c)
	if !ok {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Detail(c.Request.Context(), tenant, actor, c.Param("id"), input)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Detail(result))
}
