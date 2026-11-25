package repository

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/google/uuid"
)

type conversationKey struct{ tenant, thread string }

// markConversation remembers a thread changed inside this transaction; only transaction-scoped repositories track.
func (r *MessagingRepository) markConversation(tenant, thread string) {
	if r.dirty != nil {
		r.dirty[conversationKey{tenant, thread}] = struct{}{}
	}
}

// flushConversationEvents writes one messaging.conversation.changed outbox row per changed thread, with the final
// thread and latest-turn state (a closed turn included), in the same transaction as the change. Threads without
// any turn emit nothing.
func (r *MessagingRepository) flushConversationEvents(ctx context.Context) *apperrors.Error {
	for key := range r.dirty {
		thread, err := r.Thread(ctx, key.tenant, key.thread, false)
		if err != nil {
			return err
		}
		turn, err := r.LatestTurn(ctx, key.tenant, key.thread)
		if err != nil {
			return err
		}
		if turn == nil {
			continue
		}
		payload, marshalErr := json.Marshal(ConversationChangedOf(*thread, *turn))
		if marshalErr != nil {
			return apperrors.ErrSystemError(ctx, "conversation event encoding failed")
		}
		event := model.DomainEvent{ID: uuid.NewString(), TenantID: key.tenant, EventType: model.DomainEventConversationChanged, AggregateID: key.thread, Payload: string(payload), State: model.DomainEventStatePending}
		if err := r.db.WithContext(ctx).Create(&event).Error; err != nil {
			return salesDBError(ctx, err)
		}
	}
	r.dirty = map[conversationKey]struct{}{}
	return nil
}

// ConversationChangedOf snapshots a thread and its latest turn as the event payload.
func ConversationChangedOf(thread types.MessagingThread, turn types.MessagingTurn) messaging.ConversationChanged {
	return messaging.ConversationChanged{
		ThreadID: thread.ID, ExternalKey: thread.ExternalKey, DisplayName: thread.DisplayName,
		ConnectionID: thread.ConnectionID, RecipientID: thread.RecipientID,
		ThreadCreatedAt: thread.CreatedAt, ThreadUpdatedAt: thread.UpdatedAt,
		TurnID: turn.ID, Lifecycle: turn.Lifecycle, Control: turn.Control, AssigneeID: turn.AssigneeID,
		NeedsHuman: turn.NeedsHuman, NeedsHumanReason: turn.NeedsHumanReason, ContextRevision: turn.ContextRevision, ControlGeneration: turn.ControlGeneration,
		TurnVersion: turn.Version, LastActivityAt: turn.LastActivityAt,
	}
}
