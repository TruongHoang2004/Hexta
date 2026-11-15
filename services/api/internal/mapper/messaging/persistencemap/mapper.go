package persistencemap

import (
	"encoding/json"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
)

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type placeholderMeta struct {
	MediaKinds []string `json:"media_kinds"`
}

func Thread(row messagingmodel.Thread) types.MessagingThread {
	return types.MessagingThread{ID: row.ID, TenantID: row.TenantID, ExternalKey: row.ExternalKey, DisplayName: row.DisplayName, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, ConnectionID: text(row.ConnectionID), RecipientID: text(row.RecipientID)}
}
func ThreadRow(value types.MessagingThread) messagingmodel.Thread {
	return messagingmodel.Thread{ID: value.ID, TenantID: value.TenantID, ExternalKey: value.ExternalKey, DisplayName: value.DisplayName, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ConnectionID: optional(value.ConnectionID), RecipientID: optional(value.RecipientID)}
}
func Turn(row messagingmodel.Turn) types.MessagingTurn {
	return types.MessagingTurn{ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, Lifecycle: row.Lifecycle, Control: row.Control, AssigneeID: row.AssigneeID, NeedsHuman: row.NeedsHuman, ContextRevision: row.ContextRevision, ControlGeneration: row.ControlGeneration, Version: row.Version, LastActivityAt: row.LastActivityAt}
}
func TurnRow(value types.MessagingTurn) messagingmodel.Turn {
	return messagingmodel.Turn{ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, Lifecycle: value.Lifecycle, Control: value.Control, AssigneeID: value.AssigneeID, NeedsHuman: value.NeedsHuman, ContextRevision: value.ContextRevision, ControlGeneration: value.ControlGeneration, Version: value.Version, LastActivityAt: value.LastActivityAt}
}

// Message tolerates malformed placeholder metadata by dropping kinds; the content type still renders a placeholder.
func Message(row messagingmodel.Message) types.MessagingMessage {
	value := types.MessagingMessage{ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, TurnID: row.TurnID, Direction: row.Direction, ActorID: row.ActorID, Body: row.Body, RequestKey: row.RequestKey, RequestHash: row.RequestHash, Status: row.Status, CreatedAt: row.CreatedAt, ContentType: row.ContentType, DeliveryStatus: row.DeliveryStatus, DeliveryErrorReason: text(row.DeliveryErrorReason)}
	if row.PlaceholderMeta != nil {
		var meta placeholderMeta
		if json.Unmarshal([]byte(*row.PlaceholderMeta), &meta) == nil {
			value.MediaKinds = meta.MediaKinds
		}
	}
	return value
}
func MessageRow(value types.MessagingMessage) messagingmodel.Message {
	row := messagingmodel.Message{ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, TurnID: value.TurnID, Direction: value.Direction, ActorID: value.ActorID, Body: value.Body, RequestKey: value.RequestKey, RequestHash: value.RequestHash, Status: value.Status, CreatedAt: value.CreatedAt, ContentType: value.ContentType, DeliveryStatus: value.DeliveryStatus, DeliveryErrorReason: optional(value.DeliveryErrorReason)}
	if row.ContentType == "" {
		row.ContentType = types.ContentText
	}
	if row.DeliveryStatus == "" {
		row.DeliveryStatus = types.DeliveryReceived
	}
	if len(value.MediaKinds) > 0 {
		data, _ := json.Marshal(placeholderMeta{MediaKinds: value.MediaKinds})
		meta := string(data)
		row.PlaceholderMeta = &meta
	}
	return row
}
