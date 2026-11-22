package service

import (
	"context"
	"encoding/json"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
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

// Decode validates a topic value. ok=false means the envelope is malformed and must be quarantined; a nil change
// with ok=true is an event this consumer deliberately ignores (unknown type or newer schema version).
func (p *ConversationProjector) Decode(value []byte) (tenant string, change *types.ConversationChanged, ok bool) {
	var env types.DomainEventEnvelope
	if json.Unmarshal(value, &env) != nil {
		return "", nil, false
	}
	if env.EventType != model.DomainEventConversationChanged || env.SchemaVersion != 1 {
		return "", nil, env.EventType != "" && env.SchemaVersion >= 1
	}
	var data types.ConversationChanged
	if json.Unmarshal(env.Data, &data) != nil {
		return "", nil, false
	}
	for _, id := range []string{env.TenantID, data.ThreadID, data.TurnID} {
		if _, err := uuid.Parse(id); err != nil {
			return "", nil, false
		}
	}
	if data.TurnVersion < 1 || data.ThreadID != env.AggregateID {
		return "", nil, false
	}
	return env.TenantID, &data, true
}

// Project applies one decoded change; failures are retried by the consumer and never skip the offset.
func (p *ConversationProjector) Project(ctx context.Context, tenant string, change types.ConversationChanged) *apperrors.Error {
	return p.store.Apply(ctx, tenant, change)
}
