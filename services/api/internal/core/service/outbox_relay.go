package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/catalog"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

const outboxPublishTimeout = 10 * time.Second

// IEventPublisher delivers one message to a topic; it is satisfied by the shared multi-topic Kafka transport.
type IEventPublisher interface {
	PublishTo(ctx context.Context, topic, key string, value []byte) error
}

// OutboxRelay publishes committed domain events at least once to the single topic of their shared event. Enveloped
// events (order, conversation) are wrapped in events.Envelope keyed by aggregate; others are published as their raw
// payload keyed by the event. Consumers deduplicate by event_id.
type OutboxRelay struct {
	store     repository.IOutboxRepository
	publisher IEventPublisher
}

func NewOutboxRelay(store repository.IOutboxRepository, publisher IEventPublisher) *OutboxRelay {
	return &OutboxRelay{store: store, publisher: publisher}
}

// RelayOne publishes the next due event. It reports whether an event was claimed, so callers can drain without waiting.
// A crash after the broker ACK but before FinishEvent republishes the same event_id. A row of an unknown event type or
// with an invalid payload is never published and stays scheduled for retry.
func (r *OutboxRelay) RelayOne(ctx context.Context) (bool, *apperrors.Error) {
	event, err := r.store.ClaimEvent(ctx)
	if err != nil || event == nil {
		return false, err
	}
	published := false
	if entry, ok := catalog.ByEventType(event.EventType); ok && r.publisher != nil {
		if key, value, encodeErr := outboxMessage(entry, event); encodeErr == nil {
			call, cancel := context.WithTimeout(ctx, outboxPublishTimeout)
			published = r.publisher.PublishTo(call, entry.Topic, key, value) == nil
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

// outboxMessage encodes one outbox row as its topic's Kafka key and value.
func outboxMessage(entry catalog.Entry, event *model.DomainEvent) (string, []byte, error) {
	decoded, err := entry.Decode([]byte(event.Payload))
	if err == nil {
		err = decoded.Validate()
	}
	if err != nil {
		return "", nil, err
	}
	if !entry.Enveloped {
		return decoded.Key(), []byte(event.Payload), nil
	}
	value, err := json.Marshal(events.Envelope{SchemaVersion: events.EnvelopeSchemaVersion, EventID: event.ID, EventType: event.EventType, TenantID: event.TenantID, AggregateID: event.AggregateID, OccurredAt: event.CreatedAt.UTC(), Data: json.RawMessage(event.Payload)})
	return event.AggregateID, value, err
}
