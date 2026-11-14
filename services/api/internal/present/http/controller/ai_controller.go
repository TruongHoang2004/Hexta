package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AIController struct {
	*baseController
	aiService service.IAIAgentService
}

func NewAIController(validate *validator.Validate, aiService service.IAIAgentService) *AIController {
	return &AIController{
		baseController: NewBaseController(validate),
		aiService:      aiService,
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
// @Success 200 {string} string "Named message, products, draft_order, quick_replies, done and error events; data is dto.SSEEvent"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/ai/converse [post]
func (ctrl *AIController) Converse(c *gin.Context) {
	tenantID := ctrl.getTenantID(c)
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("tenant_id is required via X-Tenant-ID header or query parameter"))
		return
	}

	var req dto.AIConverseRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256<<10)
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.ErrorData(c, errors.ErrBadRequest(c.Request.Context()).SetDetail("invalid conversation request"))
		return
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	userID := ctrl.getAuthUserID(c)
	emitEvent := func(event *types.AIEvent) error {
		if event == nil {
			return nil
		}
		// Prepare SSE response headers
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		data, err := json.Marshal(dto.AIEventResponse(event))
		if err != nil {
			return err
		}
		_, writeErr := c.Writer.Write([]byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event.Type, string(data))))
		if writeErr != nil {
			return writeErr
		}
		c.Writer.Flush()
		return nil
	}

	if err := ctrl.aiService.Converse(c.Request.Context(), tenantID, userID, req.DomainInput(), emitEvent); err != nil {
		if !c.Writer.Written() {
			ctrl.ErrorData(c, err)
			return
		}
		_ = emitEvent(&types.AIEvent{
			Type:  types.AIEventError,
			Error: err.GetDetail(),
		})
	}
}
