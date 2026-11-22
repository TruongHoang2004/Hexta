package service

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// ConversationProjector keeps the conversation read model current from messaging.conversation.changed events.
type ConversationProjector struct {
	store repository.ConversationViewStore
}

func NewConversationProjector(store repository.ConversationViewStore) *ConversationProjector {
	return &ConversationProjector{store: store}
}

// Decode validates a topic value. ok=false means the value is malformed and must be quarantined.
func (p *ConversationProjector) Decode(value []byte) (tenant string, change *messaging.ConversationChanged, ok bool) {
	var e messaging.ConversationChanged
	if json.Unmarshal(value, &e) != nil || e.Validate() != nil {
		return "", nil, false
	}
	for _, id := range []string{e.TenantID, e.ThreadID, e.TurnID} {
		if _, err := uuid.Parse(id); err != nil {
			return "", nil, false
		}
	}
	if e.TurnVersion < 1 {
		return "", nil, false
	}
	return e.TenantID, &e, true
}

// Project applies one decoded change; failures are retried by the consumer and never skip the offset.
func (p *ConversationProjector) Project(ctx context.Context, tenant string, change messaging.ConversationChanged) *apperrors.Error {
	return p.store.Apply(ctx, tenant, change)
}
