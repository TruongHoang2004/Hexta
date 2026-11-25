package controller

import (
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/httpmap"
	replymap "github.com/TruongHoang2004/Hexta/services/api/internal/mapper/reply/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ctrl *MessagingCommandController) proposalID(c *gin.Context) bool {
	if _, err := uuid.Parse(c.Param("proposalID")); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid proposal ID"))
		return false
	}
	return true
}

// SendProposal
// @Summary Send a bot reply proposal as-is or edited
// @Description Omit text to send the proposal unchanged. Only the assigned sale may send on a staff-handled conversation; on a bot-handled conversation the sender claims it atomically with the send. A stale proposal (conversation control changed after it was created) or one that is no longer pending is 409. Real channels get a single-attempt delivery intent; simulator conversations a simulated message. Retrying the same client_message_id returns the original message.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param proposalID path string true "Proposal ID"
// @Param request body dto.SendProposalRequest true "Send"
// @Success 200 {object} response.Response[dto.MessagingMessage]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/proposals/{proposalID}/send [post]
func (ctrl *MessagingCommandController) SendProposal(c *gin.Context) {
	var input dto.SendProposalRequest
	if !ctrl.proposalID(c) || !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.SendProposal(c.Request.Context(), tenant, actor, c.Param("id"), c.Param("proposalID"), types.ProposalSendInput{ClientMessageID: input.ClientMessageID, Text: input.Text})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Message(*result))
}

// DiscardProposal
// @Summary Discard a bot reply proposal
// @Description Nothing is sent. The assigned sale, the tenant owner, or any member while the conversation is unassigned may discard; stale proposals can be discarded. 409 when the proposal is no longer pending.
// @Tags Messaging
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Param proposalID path string true "Proposal ID"
// @Success 200 {object} response.Response[dto.ReplyProposal]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/proposals/{proposalID}/discard [post]
func (ctrl *MessagingCommandController) DiscardProposal(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) || !ctrl.proposalID(c) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.DiscardProposal(c.Request.Context(), tenant, actor, c.Param("id"), c.Param("proposalID"))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, replymap.Proposal(types.ProposalView{Proposal: *result}))
}

// RegenerateProposal
// @Summary Ask the bot for a new proposal at the current conversation control
// @Description Answers the turn's newest customer message and replaces its live proposal. Any member on a bot-handled conversation, the assigned sale on a staff-handled one. A requested proposal is never sent automatically and never changes control. Refused with 409 while the reply mode is off. Returns the routing job; poll the conversation for the result.
// @Tags Messaging
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Turn ID"
// @Success 200 {object} response.Response[dto.ReplyProposal]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/turns/{id}/regenerate-proposal [post]
func (ctrl *MessagingCommandController) RegenerateProposal(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.messaging.RegenerateProposal(c.Request.Context(), tenant, actor, c.Param("id"))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, replymap.Proposal(types.ProposalView{Proposal: *result}))
}
