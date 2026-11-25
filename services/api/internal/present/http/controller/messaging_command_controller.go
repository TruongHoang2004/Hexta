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
	"net/http"
)

type MessagingCommandController struct {
	*baseController
	messaging *service.MessagingCommandService
}

func NewMessagingCommandController(validate *validator.Validate, s *service.MessagingCommandService) *MessagingCommandController {
	return &MessagingCommandController{NewBaseController(validate), s}
}
func (ctrl *MessagingCommandController) bind(c *gin.Context, input any) bool {
	if !ctrl.messagingResourceIDs(c) {
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
func (ctrl *MessagingCommandController) Inbound(c *gin.Context) {
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
func (ctrl *MessagingCommandController) Claim(c *gin.Context) {
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
func (ctrl *MessagingCommandController) Assign(c *gin.Context) {
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
func (ctrl *MessagingCommandController) Send(c *gin.Context) {
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
func (ctrl *MessagingCommandController) Handoff(c *gin.Context) {
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
// @Description Current assignee or tenant owner required for staffed conversations. While the tenant's reply mode is off the bot is off and the return is refused with 409. No AI or live channel dispatch is performed.
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
func (ctrl *MessagingCommandController) ReturnToBot(c *gin.Context) {
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

// Resolve
// @Summary Mark a conversation's needs-human flag resolved
// @Description Only the assigned sale under staff control can resolve. Restarts the 24-hour inactivity window and keeps staff control; nothing returns to the bot. 409 when the turn changed, is closed or does not need a person.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param request body dto.MessagingCommandRequest true "Command"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/resolve [post]
func (ctrl *MessagingCommandController) Resolve(c *gin.Context) {
	var input dto.MessagingCommandRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.Command(c.Request.Context(), tenant, actor, c.Param("id"), "resolve", types.MessagingCommand{ExpectedVersion: input.ExpectedVersion})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// CloseSession
// @Summary Close the current conversation session before the 24-hour expiry
// @Description The assigned sale under staff control or the tenant owner closes the open session (turn) with a required outcome and an optional note (at most 1000 characters). Closing works like expiry: the version and control generation are bumped (pending proposals become stale and queued bot work is dropped), unstarted sends are cancelled, needs-human is cleared and history is kept. The next new customer message opens a new unassigned session. 403 for other members (including an unassigned or bot-handled session unless owner); 409 when the session changed or is already closed.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn (session) ID"
// @Param request body dto.MessagingSessionRequest true "Outcome, note and expected version"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/close [post]
func (ctrl *MessagingCommandController) CloseSession(c *gin.Context) {
	var input dto.MessagingSessionRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.CloseSession(c.Request.Context(), tenant, actor, c.Param("id"), types.SessionInput{ExpectedVersion: input.ExpectedVersion, Outcome: input.Outcome, Note: input.Note})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// UpdateSessionOutcome
// @Summary Record or edit a closed session's outcome and note
// @Description The tenant owner or the session's last handler sets the outcome (required) and note (at most 1000 characters) of a closed session, including an expired one. The session stays closed. 403 for other members; 409 when the session is open or changed.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn (session) ID"
// @Param request body dto.MessagingSessionRequest true "Outcome, note and expected version"
// @Success 200 {object} response.Response[dto.MessagingTurn]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/outcome [post]
func (ctrl *MessagingCommandController) UpdateSessionOutcome(c *gin.Context) {
	var input dto.MessagingSessionRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.UpdateSessionOutcome(c.Request.Context(), tenant, actor, c.Param("id"), types.SessionInput{ExpectedVersion: input.ExpectedVersion, Outcome: input.Outcome, Note: input.Note})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Turn(*result))
}

// SendChannel
// @Summary Send a real channel reply (Facebook or Telegram) as the assigned sale
// @Description Assigned sale under sales control only; the conversation's Page/bot connection must be active. Records the message as sending with a single-attempt delivery intent; the connector reports sent, failed (window_24h, auth_expired, rate_limit, invalid_recipient, bot_blocked, chat_not_found, transport) or unknown. Never retried automatically. Retrying the same client_message_id returns the original message.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param request body dto.SendMessageRequest true "Message"
// @Success 200 {object} response.Response[dto.MessagingMessage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/messages [post]
func (ctrl *MessagingCommandController) SendChannel(c *gin.Context) {
	var input dto.SendMessageRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.SendChannel(c.Request.Context(), tenant, actor, c.Param("id"), types.ChannelSendInput{ClientMessageID: input.ClientMessageID, Text: input.Text, ExpectedVersion: input.ExpectedVersion, ContextRevision: input.ContextRevision})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Message(*result))
}
