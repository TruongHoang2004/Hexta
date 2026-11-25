package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/reply/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReplySettingsController struct {
	*baseController
	settings *service.ReplySettingsService
}

func NewReplySettingsController(validate *validator.Validate, settings *service.ReplySettingsService) *ReplySettingsController {
	return &ReplySettingsController{NewBaseController(validate), settings}
}

// GetReplyMode
// @Summary Read the tenant's bot reply mode
// @Description Active tenant membership required. configured=false means the owner never chose a mode: new conversations wait for staff and reply_mode reads as suggest.
// @Tags Messaging
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.ReplySettingsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/settings/reply-mode [get]
func (ctrl *ReplySettingsController) GetReplyMode(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.settings.Get(c.Request.Context(), tenant, actor)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Settings(result))
}

// UpdateReplyMode
// @Summary Change the tenant's bot reply mode
// @Description Tenant owner only. suggest: the bot proposes and a sale sends; auto_send: the bot sends eligible replies itself while a conversation is under bot handling. The first save also enables bot handling for new conversations. expected_version is the version last read (0 before the first save); a stale version is 409. Existing proposals are never sent by a mode change.
// @Tags Messaging
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.UpdateReplySettingsRequest true "Reply mode"
// @Success 200 {object} response.Response[dto.ReplySettingsResponse]
// @Failure 400,401,403,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/settings/reply-mode [put]
func (ctrl *ReplySettingsController) UpdateReplyMode(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	var input dto.UpdateReplySettingsRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4*1024)
	if err := c.ShouldBindJSON(&input); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid JSON request"))
		return
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), input); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.settings.Update(c.Request.Context(), tenant, actor, types.ReplySettingsUpdate{Mode: input.ReplyMode, ExpectedVersion: input.ExpectedVersion})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Settings(result))
}
