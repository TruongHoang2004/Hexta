package bootstrap

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestOutboxComposition(t *testing.T) {
	if err := fx.ValidateApp(BuildOutbox(), fx.Supply(&gorm.DB{}, &config.Config{}, zap.NewNop().Sugar()), fx.NopLogger); err != nil {
		t.Fatalf("outbox graph is not wired: %v", err)
	}
}

func TestOutboxRelayDisabledWithoutBrokers(t *testing.T) {
	// No brokers registers no hooks, so start/stop must not touch the database or Kafka.
	app := fxtest.New(t, BuildOutbox(), fx.Supply(&gorm.DB{}, &config.Config{}, zap.NewNop().Sugar()), fx.NopLogger)
	app.RequireStart().RequireStop()
}

type relayStore struct {
	pending  []*model.DomainEvent
	finished chan string
}

func (s *relayStore) ClaimEvent(context.Context) (*model.DomainEvent, *apperrors.Error) {
	if len(s.pending) == 0 {
		return nil, nil
	}
	event := s.pending[0]
	s.pending = s.pending[1:]
	return event, nil
}

func (s *relayStore) FinishEvent(_ context.Context, event *model.DomainEvent, published bool) *apperrors.Error {
	if published {
		s.finished <- event.ID
	}
	return nil
}

type relayPublisher struct{}

func (relayPublisher) Publish(context.Context, string, []byte) error { return nil }

func TestRunOutboxRelayDrainsAndStops(t *testing.T) {
	store := &relayStore{pending: []*model.DomainEvent{{ID: "a", Payload: "{}"}, {ID: "b", Payload: "{}"}}, finished: make(chan string, 2)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		runOutboxRelay(ctx, service.NewOutboxRelay(store, relayPublisher{}), zap.NewNop().Sugar())
	}()
	// Due events are drained back to back, without waiting for the poll interval between them.
	for _, want := range []string{"a", "b"} {
		select {
		case got := <-store.finished:
			if got != want {
				t.Fatalf("expected %s, got %s", want, got)
			}
		case <-time.After(outboxPollInterval / 2):
			t.Fatalf("event %s was not relayed promptly", want)
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("relay did not stop after cancellation")
	}
}
