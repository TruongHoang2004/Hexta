package persistencemap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
)

func Thread(row messagingmodel.Thread) types.MessagingThread {
	return types.MessagingThread{ID: row.ID, TenantID: row.TenantID, ExternalKey: row.ExternalKey, DisplayName: row.DisplayName, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
func ThreadRow(value types.MessagingThread) messagingmodel.Thread {
	return messagingmodel.Thread{ID: value.ID, TenantID: value.TenantID, ExternalKey: value.ExternalKey, DisplayName: value.DisplayName, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func Turn(row messagingmodel.Turn) types.MessagingTurn {
	return types.MessagingTurn{ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, Lifecycle: row.Lifecycle, Control: row.Control, AssigneeID: row.AssigneeID, NeedsHuman: row.NeedsHuman, ContextRevision: row.ContextRevision, ControlGeneration: row.ControlGeneration, Version: row.Version, LastActivityAt: row.LastActivityAt}
}
func TurnRow(value types.MessagingTurn) messagingmodel.Turn {
	return messagingmodel.Turn{ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, Lifecycle: value.Lifecycle, Control: value.Control, AssigneeID: value.AssigneeID, NeedsHuman: value.NeedsHuman, ContextRevision: value.ContextRevision, ControlGeneration: value.ControlGeneration, Version: value.Version, LastActivityAt: value.LastActivityAt}
}
func Message(row messagingmodel.Message) types.MessagingMessage {
	return types.MessagingMessage{ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, TurnID: row.TurnID, Direction: row.Direction, ActorID: row.ActorID, Body: row.Body, RequestKey: row.RequestKey, RequestHash: row.RequestHash, Status: row.Status, CreatedAt: row.CreatedAt}
}
func MessageRow(value types.MessagingMessage) messagingmodel.Message {
	return messagingmodel.Message{ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, TurnID: value.TurnID, Direction: value.Direction, ActorID: value.ActorID, Body: value.Body, RequestKey: value.RequestKey, RequestHash: value.RequestHash, Status: value.Status, CreatedAt: value.CreatedAt}
}
