package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/httpmap"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	streamHeartbeat = 20 * time.Second
	// streamMaxAge forces a reconnect so a revoked session or removed member is rechecked at connect time.
	streamMaxAge = 10 * time.Minute
)

type ConversationStreamController struct {
	*baseController
	hub *service.ConversationStreamHub
}

func NewConversationStreamController(validate *validator.Validate, hub *service.ConversationStreamHub) *ConversationStreamController {
	return &ConversationStreamController{NewBaseController(validate), hub}
}

// streamFrame maps a hint to its SSE event: `message` carries a stored message in the API message shape,
// `delivery` a delivery-state update for one message, `changed` a list refresh trigger, `resync` a full refetch.
func streamFrame(hint types.RealtimeHint) (string, []byte) {
	var data []byte
	switch {
	case hint.ThreadID == "":
		data, _ = json.Marshal(gin.H{})
		return "resync", data
	case hint.Reason == types.RealtimeMessageCreated && hint.Message != nil:
		data, _ = json.Marshal(gin.H{"thread_id": hint.ThreadID, "message": httpmap.Message(*hint.Message)})
		return "message", data
	case hint.Reason == types.RealtimeDeliveryChanged && hint.Message != nil:
		data, _ = json.Marshal(gin.H{"thread_id": hint.ThreadID, "message_id": hint.Message.ID, "delivery_status": hint.Message.DeliveryStatus, "delivery_error_reason": hint.Message.DeliveryErrorReason})
		return "delivery", data
	}
	data, _ = json.Marshal(gin.H{"thread_id": hint.ThreadID, "reason": hint.Reason})
	return "changed", data
}

// Stream
// @Summary Stream conversation change hints (SSE)
// @Description Server-sent events: `ready` on connect; `message` {thread_id, message} for each new stored message (message has the same shape as in the conversation messages response); `delivery` {thread_id, message_id, delivery_status, delivery_error_reason}; `changed` {thread_id, reason} when the conversation list should refresh; `resync` when events may have been lost and clients must refetch. PostgreSQL stays authoritative; clients dedupe by message id. The stream closes after about 10 minutes and clients reconnect. Requires current active tenant membership.
// @Tags Messaging
// @Produce text/event-stream
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {string} string "event stream"
// @Failure 400,401,403,404,429 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/conversations/stream [get]
func (ctrl *ConversationStreamController) Stream(c *gin.Context) {
	tenant, actor := salesScope(c)
	if actor == "" {
		ctrl.ErrorData(c, apperrors.ErrUnauthorized(c.Request.Context()))
		return
	}
	sub, err := ctrl.hub.Subscribe(tenant, actor)
	if errors.Is(err, service.ErrTooManyStreams) {
		c.AbortWithStatusJSON(429, gin.H{"detail": "too many open streams"})
		return
	}
	defer sub.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(200)
	write := func(format string, args ...any) bool {
		if _, err := fmt.Fprintf(c.Writer, format, args...); err != nil {
			return false
		}
		c.Writer.Flush()
		return true
	}
	if !write("event: ready\ndata: {}\n\n") {
		return
	}
	heartbeat := time.NewTicker(streamHeartbeat)
	defer heartbeat.Stop()
	expiry := time.NewTimer(streamMaxAge)
	defer expiry.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-expiry.C:
			write("event: reconnect\ndata: {}\n\n")
			return
		case <-heartbeat.C:
			if !write(": keep-alive\n\n") {
				return
			}
		case hint := <-sub.C:
			event, data := streamFrame(hint)
			if !write("event: %s\ndata: %s\n\n", event, data) {
				return
			}
		}
	}
}
