package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/order"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/stretchr/testify/require"
)

type fakeOutboxStore struct {
	events   []*model.DomainEvent
	claimErr *apperrors.Error
	finished map[string]bool
}

func (s *fakeOutboxStore) ClaimEvent(ctx context.Context) (*model.DomainEvent, *apperrors.Error) {
	if s.claimErr != nil {
		return nil, s.claimErr
	}
	if len(s.events) == 0 {
		return nil, nil
	}
	event := s.events[0]
	s.events = s.events[1:]
	return event, nil
}

func (s *fakeOutboxStore) FinishEvent(ctx context.Context, event *model.DomainEvent, published bool) *apperrors.Error {
	if s.finished == nil {
		s.finished = map[string]bool{}
	}
	s.finished[event.ID] = published
	return nil
}

type fakePublisher struct {
	topics []string
	keys   []string
	values [][]byte
	err    error
	ctxOK  bool
}

func (p *fakePublisher) PublishTo(ctx context.Context, topic, key string, value []byte) error {
	_, p.ctxOK = ctx.Deadline()
	p.topics = append(p.topics, topic)
	p.keys = append(p.keys, key)
	p.values = append(p.values, value)
	return p.err
}

func row(t *testing.T, id, eventType, aggregate string, payload any) *model.DomainEvent {
	b, err := json.Marshal(payload)
	require.NoError(t, err)
	return &model.DomainEvent{ID: id, TenantID: "tenant", EventType: eventType, AggregateID: aggregate, Payload: string(b), Fence: 1, CreatedAt: time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)}
}

func outboxEvent(t *testing.T) *model.DomainEvent {
	return row(t, "8b0f1c1e-0000-4000-8000-000000000001", model.DomainEventOrderConfirmed, "order-1", order.OrderConfirmed{OrderID: "order-1", OrderNumber: "SO-1", Version: 2, TotalAmount: "10.00", ConfirmedBy: "owner"})
}

func TestOutboxRelayPublishesVersionedEnvelopeKeyedByAggregate(t *testing.T) {
	event := outboxEvent(t)
	store := &fakeOutboxStore{events: []*model.DomainEvent{event}}
	publisher := &fakePublisher{}
	claimed, err := NewOutboxRelay(store, publisher).RelayOne(context.Background())
	require.Nil(t, err)
	require.True(t, claimed)
	require.True(t, publisher.ctxOK, "publication must be bounded by a timeout")
	require.Equal(t, []string{order.OrderConfirmedTopic}, publisher.topics)
	require.Equal(t, []string{"order-1"}, publisher.keys)

	var envelope events.Envelope
	require.NoError(t, json.Unmarshal(publisher.values[0], &envelope))
	require.Equal(t, 1, envelope.SchemaVersion)
	require.Equal(t, event.ID, envelope.EventID)
	require.Equal(t, model.DomainEventOrderConfirmed, envelope.EventType)
	require.Equal(t, "tenant", envelope.TenantID)
	require.Equal(t, "order-1", envelope.AggregateID)
	require.JSONEq(t, event.Payload, string(envelope.Data))
	require.Equal(t, map[string]bool{event.ID: true}, store.finished)
}

func TestOutboxRelayReleasesFailedPublicationForRetry(t *testing.T) {
	event := outboxEvent(t)
	store := &fakeOutboxStore{events: []*model.DomainEvent{event}}
	claimed, err := NewOutboxRelay(store, &fakePublisher{err: stderrors.New("broker unavailable")}).RelayOne(context.Background())
	require.True(t, claimed)
	require.NotNil(t, err)
	require.NotContains(t, err.GetDetail(), "broker unavailable")
	require.Equal(t, map[string]bool{event.ID: false}, store.finished)
}

func TestOutboxRelayIdleAndClaimFailure(t *testing.T) {
	publisher := &fakePublisher{}
	claimed, err := NewOutboxRelay(&fakeOutboxStore{}, publisher).RelayOne(context.Background())
	require.False(t, claimed)
	require.Nil(t, err)

	failing := &fakeOutboxStore{claimErr: apperrors.ErrSystemError(context.Background(), "outbox persistence operation failed")}
	claimed, err = NewOutboxRelay(failing, publisher).RelayOne(context.Background())
	require.False(t, claimed)
	require.NotNil(t, err)
	require.Empty(t, publisher.keys)
}

func TestOutboxRelayRoutesEachEventToItsOwnTopic(t *testing.T) {
	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	intent := row(t, "8b0f1c1e-0000-4000-8000-000000000002", model.DomainEventOutboundIntent, "i", delivery.SendRequested{Meta: events.Meta{EventID: "8b0f1c1e-0000-4000-8000-000000000002", CorrelationID: "8b0f1c1e-0000-4000-8000-0000000000ff", TenantID: "8b0f1c1e-0000-4000-8000-0000000000aa", OccurredAt: now}, IntentID: "8b0f1c1e-0000-4000-8000-000000000010", AttemptID: "8b0f1c1e-0000-4000-8000-000000000011", ConnectionID: "8b0f1c1e-0000-4000-8000-000000000012", ConnectionVersion: 1, Platform: "facebook", AccountID: "page", RecipientID: "psid", Text: "hi", ExpiresAt: now.Add(5 * time.Minute)})
	change := row(t, "8b0f1c1e-0000-4000-8000-000000000003", model.DomainEventConversationChanged, "thread-1", messaging.ConversationChanged{ThreadID: "thread-1", TurnID: "turn-1", TurnVersion: 1})
	store := &fakeOutboxStore{events: []*model.DomainEvent{intent, change}}
	publisher := &fakePublisher{}
	relay := NewOutboxRelay(store, publisher)
	for range 2 {
		claimed, err := relay.RelayOne(context.Background())
		require.True(t, claimed)
		require.Nil(t, err)
	}
	require.Equal(t, []string{delivery.SendRequestedTopic, messaging.ConversationChangedTopic}, publisher.topics)
	require.Equal(t, []string{"8b0f1c1e-0000-4000-8000-000000000012:psid", "thread-1"}, publisher.keys, "intents are ordered per provider conversation, snapshots per thread")
	require.JSONEq(t, intent.Payload, string(publisher.values[0]), "the connector reads the raw shared contract, not an envelope")
	var envelope events.Envelope
	require.NoError(t, json.Unmarshal(publisher.values[1], &envelope))
	require.Equal(t, model.DomainEventConversationChanged, envelope.EventType)
}

func TestOutboxRelayNeverPublishesUnknownTypesOrInvalidPayloads(t *testing.T) {
	unknown := &model.DomainEvent{ID: "8b0f1c1e-0000-4000-8000-000000000004", TenantID: "tenant", EventType: "order.cancelled", Payload: `{}`, Fence: 1}
	invalid := &model.DomainEvent{ID: "8b0f1c1e-0000-4000-8000-000000000005", TenantID: "tenant", EventType: model.DomainEventOrderConfirmed, Payload: `{"order_id":"o"}`, Fence: 1}
	store := &fakeOutboxStore{events: []*model.DomainEvent{unknown, invalid}}
	publisher := &fakePublisher{}
	relay := NewOutboxRelay(store, publisher)
	for range 2 {
		claimed, err := relay.RelayOne(context.Background())
		require.True(t, claimed)
		require.NotNil(t, err)
	}
	require.Empty(t, publisher.topics)
	require.Equal(t, map[string]bool{unknown.ID: false, invalid.ID: false}, store.finished)
}
