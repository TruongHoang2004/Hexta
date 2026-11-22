package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
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
	keys   []string
	values [][]byte
	err    error
	ctxOK  bool
}

func (p *fakePublisher) Publish(ctx context.Context, key string, value []byte) error {
	_, p.ctxOK = ctx.Deadline()
	p.keys = append(p.keys, key)
	p.values = append(p.values, value)
	return p.err
}

func outboxEvent() *model.DomainEvent {
	return &model.DomainEvent{ID: "8b0f1c1e-0000-4000-8000-000000000001", TenantID: "tenant", EventType: model.DomainEventOrderConfirmed, AggregateID: "order-1", Payload: `{"order_id":"order-1","version":2}`, Fence: 1, CreatedAt: time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)}
}

func TestOutboxRelayPublishesVersionedEnvelopeKeyedByAggregate(t *testing.T) {
	store := &fakeOutboxStore{events: []*model.DomainEvent{outboxEvent()}}
	publisher := &fakePublisher{}
	claimed, err := NewOutboxRelay(store, publisher).RelayOne(context.Background())
	require.Nil(t, err)
	require.True(t, claimed)
	require.True(t, publisher.ctxOK, "publication must be bounded by a timeout")
	require.Equal(t, []string{"order-1"}, publisher.keys)

	var envelope types.DomainEventEnvelope
	require.NoError(t, json.Unmarshal(publisher.values[0], &envelope))
	require.Equal(t, 1, envelope.SchemaVersion)
	require.Equal(t, "8b0f1c1e-0000-4000-8000-000000000001", envelope.EventID)
	require.Equal(t, model.DomainEventOrderConfirmed, envelope.EventType)
	require.Equal(t, "tenant", envelope.TenantID)
	require.JSONEq(t, `{"order_id":"order-1","version":2}`, string(envelope.Data))
	require.Equal(t, map[string]bool{envelope.EventID: true}, store.finished)
}

func TestOutboxRelayReleasesFailedPublicationForRetry(t *testing.T) {
	store := &fakeOutboxStore{events: []*model.DomainEvent{outboxEvent()}}
	claimed, err := NewOutboxRelay(store, &fakePublisher{err: stderrors.New("broker unavailable")}).RelayOne(context.Background())
	require.True(t, claimed)
	require.NotNil(t, err)
	require.NotContains(t, err.GetDetail(), "broker unavailable")
	require.Equal(t, map[string]bool{outboxEvent().ID: false}, store.finished)
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

func TestOutboxRelayRoutesIntentsAsRawContractToIntentTopic(t *testing.T) {
	payload := `{"schema_version":1,"intent_id":"i","connection_id":"conn","recipient_id":"psid","text":"hi"}`
	intent := &model.DomainEvent{ID: "8b0f1c1e-0000-4000-8000-000000000002", TenantID: "tenant", EventType: model.DomainEventOutboundIntent, AggregateID: "i", Payload: payload, Fence: 1}
	orders, intents := &fakePublisher{}, &fakePublisher{}
	store := &fakeOutboxStore{events: []*model.DomainEvent{intent}}
	claimed, err := NewOutboxRelay(store, orders).WithIntentPublisher(intents).RelayOne(context.Background())
	require.True(t, claimed)
	require.Nil(t, err)
	require.Empty(t, orders.keys, "intents must never reach the order topic")
	require.Equal(t, []string{"conn:psid"}, intents.keys, "intents are ordered per Page conversation")
	require.JSONEq(t, payload, string(intents.values[0]), "the connector reads the shared contract, not an envelope")

	// Without an intent publisher the row is released for retry instead of being dropped.
	store = &fakeOutboxStore{events: []*model.DomainEvent{intent}}
	claimed, err = NewOutboxRelay(store, orders).RelayOne(context.Background())
	require.True(t, claimed)
	require.NotNil(t, err)
	require.Equal(t, map[string]bool{intent.ID: false}, store.finished)
}

func TestOutboxRelayRoutesConversationChangesToTheReadModelTopic(t *testing.T) {
	change := &model.DomainEvent{ID: "8b0f1c1e-0000-4000-8000-000000000003", TenantID: "tenant", EventType: model.DomainEventConversationChanged, AggregateID: "thread-1", Payload: `{"thread_id":"thread-1"}`, Fence: 1}
	orders, conversations := &fakePublisher{}, &fakePublisher{}
	store := &fakeOutboxStore{events: []*model.DomainEvent{change}}
	claimed, err := NewOutboxRelay(store, orders).WithConversationPublisher(conversations).RelayOne(context.Background())
	require.True(t, claimed)
	require.Nil(t, err)
	require.Empty(t, orders.keys, "conversation snapshots must never reach the order topic")
	require.Equal(t, []string{"thread-1"}, conversations.keys, "snapshots are keyed by thread so one thread stays ordered")

	// Without a conversation publisher the row is released for retry instead of leaking to the order topic.
	store = &fakeOutboxStore{events: []*model.DomainEvent{change}}
	claimed, err = NewOutboxRelay(store, orders).RelayOne(context.Background())
	require.True(t, claimed)
	require.NotNil(t, err)
	require.Empty(t, orders.keys)
}
