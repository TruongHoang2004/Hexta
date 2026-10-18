package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
)

type AIController struct {
	*baseController
	aiService service.IAIAgentService
}

func NewAIController(validate *validator.Validate, aiService service.IAIAgentService) *AIController {
	return &AIController{
		baseController: NewBaseController(validate),
		aiService:     aiService,
	}
}

func (ctrl *AIController) getTenantID(c *gin.Context) string {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	return tenantID
}

func (ctrl *AIController) getAuthUserID(c *gin.Context) string {
    if info, ok := common.GetAuthInfo(c.Request.Context()); ok && info.UserID != "" {
        return info.UserID
    }
    return c.GetString("userID")
}

// Converse
// @Summary Converse with AI Agent
// @Description Stream interactive conversational turns and order drafts via Server-Sent Events (SSE)
// @Tags AI Agent
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param request body dto.AIConverseRequest true "Conversational prompt and history"
// @Success 200 {string} string "Server-Sent Event stream emitting JSON dto.SSEEvent chunks"
// @Router /api/v1/ai/agent/converse [post]
func (ctrl *AIController) Converse(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	var req dto.AIConverseRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	// Prepare SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	userID := ctrl.getAuthUserID(c)
	emitEvent := func(event *dto.SSEEvent) error {
		if event == nil {
			return nil
		}
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		_, writeErr := c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", string(data))))
		if writeErr != nil {
			return writeErr
		}
		c.Writer.Flush()
		return nil
	}

	if err := ctrl.aiService.Converse(c.Request.Context(), tenantID, userID, &req, emitEvent); err != nil {
		_ = emitEvent(&dto.SSEEvent{
			Type:  dto.SSEEventError,
			Error: err.Error(),
		})
	}
}

// GetDraft
// @Summary Retrieve interactive order draft
// @Description Retrieve a Redis-cached order draft by draft ID for human review and confirmation
// @Tags AI Agent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Tenant-ID header string false "Tenant Workspace ID"
// @Param tenant_id query string false "Tenant Workspace ID"
// @Param id path string true "Draft ID (e.g. draft_8f3a9b21-4c5d)"
// @Success 200 {object} response.Response[dto.OrderDraftDTO]
// @Router /api/v1/ai/agent/drafts/{id} [get]
func (ctrl *AIController) GetDraft(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	draftID := c.Param("id")
	if draftID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("draft id is required"))
		return
	}

	draft, err := ctrl.aiService.GetDraft(c.Request.Context(), tenantID, draftID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	ctrl.Success(c, draft)
}
