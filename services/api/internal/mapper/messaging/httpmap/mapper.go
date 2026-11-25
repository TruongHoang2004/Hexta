package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	replymap "github.com/TruongHoang2004/Hexta/services/api/internal/mapper/reply/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func Thread(value types.MessagingThread) dto.MessagingThread {
	return dto.MessagingThread{ID: value.ID, DisplayName: value.DisplayName, ExternalKey: value.ExternalKey, ConnectionID: value.ConnectionID, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func Turn(value types.MessagingTurn) dto.MessagingTurn {
	return dto.MessagingTurn{ID: value.ID, ThreadID: value.ThreadID, Lifecycle: value.Lifecycle, Control: value.Control, HandlingState: value.HandlingState(), AssigneeID: value.AssigneeID, NeedsHuman: value.NeedsHuman, NeedsHumanReason: value.NeedsHumanReason, ContextRevision: value.ContextRevision, ControlGeneration: value.ControlGeneration, Version: value.Version, LastActivityAt: value.LastActivityAt, ClosedAt: value.ClosedAt,
		CloseCause: value.CloseCause, ClosedBy: value.ClosedBy, Outcome: value.Outcome, Note: value.Note, LastHandlerID: value.LastHandlerID}
}
func Sessions(value types.MessagingSessionPage) dto.MessagingSessionPage {
	result := dto.MessagingSessionPage{Items: make([]dto.MessagingSession, len(value.Items)), NextCursor: value.NextCursor}
	for i, row := range value.Items {
		session := dto.MessagingSession{Turn: Turn(row.Turn), OpenedAt: row.OpenedAt, MessageCount: row.MessageCount, FirstMessageID: row.FirstMessageID, Orders: make([]dto.MessagingSessionOrder, len(row.Orders))}
		for j, order := range row.Orders {
			session.Orders[j] = dto.MessagingSessionOrder{ID: order.ID, OrderNumber: order.OrderNumber, Status: order.Status}
		}
		result.Items[i] = session
	}
	return result
}
func Summary(value types.MessagingSummary) dto.MessagingSummary {
	return dto.MessagingSummary{WaitingStaffCount: value.WaitingStaff}
}
func Message(value types.MessagingMessage) dto.MessagingMessage {
	result := dto.MessagingMessage{ID: value.ID, ThreadID: value.ThreadID, TurnID: value.TurnID, Direction: value.Direction, ActorID: value.ActorID, Body: value.Body, Status: value.Status, CreatedAt: value.CreatedAt, ContentType: value.ContentType, DeliveryStatus: value.DeliveryStatus, DeliveryErrorReason: value.DeliveryErrorReason}
	if result.ContentType == "" {
		result.ContentType = types.ContentText
	}
	if value.ContentType == types.ContentUnsupportedMedia {
		result.PlaceholderMeta = &dto.PlaceholderMetaDTO{MediaKinds: append([]string{}, value.MediaKinds...)}
	}
	return result
}

func Conversation(value types.MessagingConversation) dto.MessagingConversation {
	result := dto.MessagingConversation{Thread: Thread(value.Thread), Source: "simulator"}
	if platform := types.ChannelThreadPlatform(value.Thread.ExternalKey); platform != "" {
		result.Source = platform
	}
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
	result := dto.MessagingDetail{Conversation: Conversation(value.Conversation), Messages: make([]dto.MessagingMessage, len(value.Messages)), NextCursor: value.NextCursor, Proposals: replymap.Proposals(value.Proposals)}
	for i, row := range value.Messages {
		result.Messages[i] = Message(row)
	}
	return result
}
