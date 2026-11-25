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
	return types.MessagingTurn{ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, Lifecycle: row.Lifecycle, Control: row.Control, AssigneeID: row.AssigneeID, NeedsHuman: row.NeedsHuman, NeedsHumanReason: text(row.NeedsHumanReason), ContextRevision: row.ContextRevision, ControlGeneration: row.ControlGeneration, Version: row.Version, LastActivityAt: row.LastActivityAt, ClosedAt: row.ClosedAt,
		CloseCause: text(row.CloseCause), ClosedBy: text(row.ClosedBy), Outcome: text(row.Outcome), Note: text(row.Note), LastHandlerID: row.LastHandlerID}
}
func TurnRow(value types.MessagingTurn) messagingmodel.Turn {
	return messagingmodel.Turn{ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, Lifecycle: value.Lifecycle, Control: value.Control, AssigneeID: value.AssigneeID, NeedsHuman: value.NeedsHuman, NeedsHumanReason: optional(value.NeedsHumanReason), ContextRevision: value.ContextRevision, ControlGeneration: value.ControlGeneration, Version: value.Version, LastActivityAt: value.LastActivityAt, ClosedAt: value.ClosedAt,
		CloseCause: optional(value.CloseCause), ClosedBy: optional(value.ClosedBy), Outcome: optional(value.Outcome), Note: optional(value.Note), LastHandlerID: value.LastHandlerID}
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

// ConversationView maps a read model row back to the thread and latest turn the API returns.
func ConversationView(row messagingmodel.ConversationView) (types.MessagingThread, types.MessagingTurn) {
	thread := types.MessagingThread{ID: row.ThreadID, TenantID: row.TenantID, ExternalKey: row.ExternalKey, DisplayName: row.DisplayName, CreatedAt: row.ThreadCreatedAt, UpdatedAt: row.ThreadUpdatedAt, ConnectionID: text(row.ConnectionID), RecipientID: text(row.RecipientID)}
	turn := types.MessagingTurn{ID: row.TurnID, TenantID: row.TenantID, ThreadID: row.ThreadID, Lifecycle: row.Lifecycle, Control: row.Control, AssigneeID: row.AssigneeID, NeedsHuman: row.NeedsHuman, NeedsHumanReason: text(row.NeedsHumanReason), ContextRevision: row.ContextRevision, ControlGeneration: row.ControlGeneration, Version: row.TurnVersion, LastActivityAt: row.LastActivityAt,
		ClosedAt: row.ClosedAt, CloseCause: text(row.CloseCause), Outcome: text(row.Outcome)}
	return thread, turn
}

// Session joins a turn with its message statistics and linked orders.
func Session(turn messagingmodel.Turn, stats messagingmodel.SessionStats, orders []messagingmodel.SessionOrder) types.MessagingSession {
	value := types.MessagingSession{Turn: Turn(turn), OpenedAt: stats.OpenedAt, MessageCount: stats.MessageCount, FirstMessageID: stats.FirstMessageID, Orders: []types.SessionOrder{}}
	for _, order := range orders {
		value.Orders = append(value.Orders, types.SessionOrder{ID: order.ID, OrderNumber: order.OrderNumber, Status: order.Status})
	}
	return value
}
