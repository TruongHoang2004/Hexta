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
	"github.com/google/uuid"
	"net/http"
)

type MessagingController struct {
	*baseController
	messaging *service.MessagingService
}

func NewMessagingController(validate *validator.Validate, s *service.MessagingService) *MessagingController {
	return &MessagingController{NewBaseController(validate), s}
}
func (ctrl *MessagingController) valid(c *gin.Context) bool {
	for _, id := range []string{c.Param("tenantID"), c.Param("id")} {
		if id != "" {
			if _, err := uuid.Parse(id); err != nil {
				ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid resource ID"))
				return false
			}
		}
	}
	return true
}
func (ctrl *MessagingController) query(c *gin.Context) (types.MessagingQuery, bool) {
	input := dto.MessagingQuery{Limit: 20}
	if !ctrl.valid(c) {
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
func (ctrl *MessagingController) bind(c *gin.Context, input any) bool {
	if !ctrl.valid(c) {
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)
	if err := c.ShouldBindJSON(input); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid JSON request"))
		return false
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), input); err != nil {
		ctrl.ErrorData(c, err)
		return false
	}
	return true
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
// @Param filter query string false "all, mine, unassigned or needs_human"
// @Success 200 {object} response.Response[dto.MessagingPage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/conversations [get]
func (ctrl *MessagingController) List(c *gin.Context) {
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
func (ctrl *MessagingController) Detail(c *gin.Context) {
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

// Inbound
// @Summary Create a clearly simulated customer message
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.MessagingInboundRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingMessage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/simulator/messages [post]
func (ctrl *MessagingController) Inbound(c *gin.Context) {
	var input dto.MessagingInboundRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Inbound(c.Request.Context(), tenant, actor, types.MessagingInbound{ExternalKey: input.ExternalKey, DisplayName: input.DisplayName, Body: input.Body, RequestKey: input.RequestKey})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Message(*result))
}

// Claim
// @Summary Claim an unassigned conversation or take control of own assignment
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Resource ID"
// @Param request body dto.MessagingCommandRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/claim [post]
func (ctrl *MessagingController) Claim(c *gin.Context) {
	var input dto.MessagingCommandRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Command(c.Request.Context(), tenant, actor, c.Param("id"), "claim", types.MessagingCommand{ExpectedVersion: input.ExpectedVersion})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// Assign
// @Summary Reassign or unassign a conversation as tenant owner
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Resource ID"
// @Param request body dto.MessagingCommandRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/assign [post]
func (ctrl *MessagingController) Assign(c *gin.Context) {
	var input dto.MessagingCommandRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Command(c.Request.Context(), tenant, actor, c.Param("id"), "assign", types.MessagingCommand{ExpectedVersion: input.ExpectedVersion, AssigneeID: input.AssigneeID})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// Send
// @Summary Save a simulated reply as the current assigned sale
// @Description Simulator only; no real provider sending. Requires current active tenant membership.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Resource ID"
// @Param request body dto.MessagingSendRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingMessage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/send-simulation [post]
func (ctrl *MessagingController) Send(c *gin.Context) {
	var input dto.MessagingSendRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.SendSimulation(c.Request.Context(), tenant, actor, c.Param("id"), types.MessagingCommand{ExpectedVersion: input.ExpectedVersion, ContextRevision: input.ContextRevision, Body: input.Body, RequestKey: input.RequestKey})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Message(*result))
}

// Handoff
// @Summary Move a conversation to the durable staff queue with a reason
// @Description Current assignee or tenant owner required for staffed conversations. Handoff of a bot conversation requires active tenant membership. No AI or live channel dispatch is performed.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param request body dto.MessagingHandoffRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/handoff [post]
func (ctrl *MessagingController) Handoff(c *gin.Context) {
	var input dto.MessagingHandoffRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Command(c.Request.Context(), tenant, actor, c.Param("id"), "handoff", types.MessagingCommand{ExpectedVersion: input.ExpectedVersion, Reason: input.Reason})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// ReturnToBot
// @Summary Explicitly return a conversation to bot control without replaying messages
// @Description Current assignee or tenant owner required for staffed conversations. Handoff of a bot conversation requires active tenant membership. No AI or live channel dispatch is performed.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param request body dto.MessagingCommandRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/return-to-bot [post]
func (ctrl *MessagingController) ReturnToBot(c *gin.Context) {
	var input dto.MessagingCommandRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Command(c.Request.Context(), tenant, actor, c.Param("id"), "return-to-bot", types.MessagingCommand{ExpectedVersion: input.ExpectedVersion})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}
