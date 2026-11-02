package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func Thread(value types.MessagingThread) dto.MessagingThread {
	return dto.MessagingThread{ID: value.ID, DisplayName: value.DisplayName, ExternalKey: value.ExternalKey, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func Turn(value types.MessagingTurn) dto.MessagingTurn {
	return dto.MessagingTurn{ID: value.ID, ThreadID: value.ThreadID, Lifecycle: value.Lifecycle, Control: value.Control, AssigneeID: value.AssigneeID, NeedsHuman: value.NeedsHuman, ContextRevision: value.ContextRevision, ControlGeneration: value.ControlGeneration, Version: value.Version, LastActivityAt: value.LastActivityAt}
}
func Message(value types.MessagingMessage) dto.MessagingMessage {
	return dto.MessagingMessage{ID: value.ID, ThreadID: value.ThreadID, TurnID: value.TurnID, Direction: value.Direction, ActorID: value.ActorID, Body: value.Body, Status: value.Status, CreatedAt: value.CreatedAt}
}

func Conversation(value types.MessagingConversation) dto.MessagingConversation {
	result := dto.MessagingConversation{Thread: Thread(value.Thread), Source: "simulator"}
	if value.Turn != nil {
		turn := Turn(*value.Turn)
		result.Turn = &turn
	}
	return result
}
func Page(value types.MessagingPage) dto.MessagingPage {
	result := dto.MessagingPage{Items: make([]dto.MessagingConversation, len(value.Items)), NextCursor: value.NextCursor}
	for i, row := range value.Items {
		result.Items[i] = Conversation(row)
	}
	return result
}
func Detail(value types.MessagingDetail) dto.MessagingDetail {
	result := dto.MessagingDetail{Conversation: Conversation(value.Conversation), Messages: make([]dto.MessagingMessage, len(value.Messages)), NextCursor: value.NextCursor}
	for i, row := range value.Messages {
		result.Messages[i] = Message(row)
	}
	return result
}
