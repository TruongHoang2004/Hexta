package service

import (
	"context"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/catalog"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

const outboxPublishTimeout = 10 * time.Second

// IEventPublisher delivers one message to a topic; it is satisfied by the shared multi-topic Kafka transport.
type IEventPublisher interface {
	PublishTo(ctx context.Context, topic, key string, value []byte) error
}

// OutboxRelay publishes committed domain events at least once. Each row's event type is its topic and its payload is
// the complete shared event; consumers deduplicate by event_id.
type OutboxRelay struct {
	store     repository.IOutboxRepository
	publisher IEventPublisher
}

func NewOutboxRelay(store repository.IOutboxRepository, publisher IEventPublisher) *OutboxRelay {
	return &OutboxRelay{store: store, publisher: publisher}
}

// RelayOne publishes the next due event. It reports whether an event was claimed, so callers can drain without waiting.
// A crash after the broker ACK but before FinishEvent republishes the same event_id. A row for an unknown topic or an
// invalid payload is never published and stays scheduled for retry.
func (r *OutboxRelay) RelayOne(ctx context.Context) (bool, *apperrors.Error) {
	event, err := r.store.ClaimEvent(ctx)
	if err != nil || event == nil {
		return false, err
	}
	published := false
	if entry, ok := catalog.Lookup(event.EventType); ok && r.publisher != nil {
		if decoded, decodeErr := entry.Decode([]byte(event.Payload)); decodeErr == nil && decoded.Validate() == nil {
			call, cancel := context.WithTimeout(ctx, outboxPublishTimeout)
			published = r.publisher.PublishTo(call, entry.Topic, decoded.Key(), []byte(event.Payload)) == nil
			cancel()
		}
	}
	if err := r.store.FinishEvent(ctx, event, published); err != nil {
		return true, err
	}
	if !published {
		return true, apperrors.ErrSystemError(ctx, "outbox publication failed; retry scheduled")
	}
	return true, nil
}
