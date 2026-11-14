package service

import (
	"context"
	"encoding/json"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

const outboxPublishTimeout = 10 * time.Second

// IEventPublisher delivers one message; it is satisfied by the shared Kafka transport.
type IEventPublisher interface {
	Publish(ctx context.Context, key string, value []byte) error
}

// OutboxRelay publishes committed domain events at least once. Consumers deduplicate by event_id.
type OutboxRelay struct {
	store     repository.IOutboxRepository
	publisher IEventPublisher
}

func NewOutboxRelay(store repository.IOutboxRepository, publisher IEventPublisher) *OutboxRelay {
	return &OutboxRelay{store: store, publisher: publisher}
}

// RelayOne publishes the next due event. It reports whether an event was claimed, so callers can drain without waiting.
// A crash after the broker ACK but before FinishEvent republishes the same event_id.
func (r *OutboxRelay) RelayOne(ctx context.Context) (bool, *apperrors.Error) {
	event, err := r.store.ClaimEvent(ctx)
	if err != nil || event == nil {
		return false, err
	}
	value, encodeErr := json.Marshal(types.DomainEventEnvelope{SchemaVersion: 1, EventID: event.ID, EventType: event.EventType, TenantID: event.TenantID, AggregateID: event.AggregateID, OccurredAt: event.CreatedAt.UTC(), Data: json.RawMessage(event.Payload)})
	published := false
	if encodeErr == nil {
		call, cancel := context.WithTimeout(ctx, outboxPublishTimeout)
		published = r.publisher.Publish(call, event.AggregateID, value) == nil
		cancel()
	}
	if err := r.store.FinishEvent(ctx, event, published); err != nil {
		return true, err
	}
	if !published {
		return true, apperrors.ErrSystemError(ctx, "outbox publication failed; retry scheduled")
	}
	return true, nil
}
