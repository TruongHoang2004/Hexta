package service

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func conversationEvents(t *testing.T, db *gorm.DB, tenant string) []model.DomainEvent {
	t.Helper()
	var events []model.DomainEvent
	if err := db.Where("tenant_id = ? AND event_type = ?", tenant, model.DomainEventConversationChanged).Order("created_at").Find(&events).Error; err != nil {
		t.Fatal(err)
	}
	return events
}

// envelope wraps an outbox row exactly like the relay does before publishing.
func envelope(t *testing.T, e model.DomainEvent) []byte {
	t.Helper()
	value, err := json.Marshal(events.Envelope{SchemaVersion: 1, EventID: e.ID, EventType: e.EventType, TenantID: e.TenantID, AggregateID: e.AggregateID, OccurredAt: time.Now().UTC(), Data: json.RawMessage(e.Payload)})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestConversationEventsAreEmittedOncePerThreadAndTransaction(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	events := conversationEvents(t, f.db, f.tenant)
	if len(events) != 1 {
		t.Fatalf("inbound touches thread, turn and thread timestamp but must emit one event, got %d", len(events))
	}
	if _, err := s.Command(context.Background(), f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err != nil {
		t.Fatal(err)
	}
	events = conversationEvents(t, f.db, f.tenant)
	if len(events) != 2 || events[1].AggregateID != message.ThreadID || events[1].State != model.DomainEventStatePending {
		t.Fatalf("claim must add one pending event for the thread: %+v", events)
	}
	var change messaging.ConversationChanged
	if err := json.Unmarshal([]byte(events[1].Payload), &change); err != nil {
		t.Fatal(err)
	}
	if change.TurnVersion != turn.Version+1 || change.AssigneeID == nil || *change.AssigneeID != f.sale || change.NeedsHuman {
		t.Fatalf("snapshot must carry the post-claim turn state: %+v", change)
	}
}

func TestConversationEventsRollBackWithTheCommand(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	before := len(conversationEvents(t, f.db, f.tenant))
	broken := NewMessagingCommandService(messagingAuditFailure{repository.NewMessagingRepository(f.db)})
	if _, err := broken.Command(context.Background(), f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err == nil {
		t.Fatal("audit failure should fail the command")
	}
	if after := len(conversationEvents(t, f.db, f.tenant)); after != before {
		t.Fatalf("a rolled-back command leaked %d conversation events", after-before)
	}
	_ = s
}

func TestConversationProjectionMatchesAuthoritativeState(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	if _, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewMessagingRepository(f.db)
	projector := NewConversationProjector(repo.ConversationViews())
	events := conversationEvents(t, f.db, f.tenant)
	apply := func(e model.DomainEvent) {
		tenant, change, ok := projector.Decode(envelope(t, e))
		if !ok || change == nil {
			t.Fatal("valid event must decode")
		}
		if err := projector.Project(ctx, tenant, *change); err != nil {
			t.Fatal(err)
		}
	}
	// Newest first, then the older snapshot and a replay: the stale ones must not roll the row back.
	apply(events[1])
	apply(events[0])
	apply(events[1])
	queries := NewMessagingQueryService(repo)
	page, err := queries.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20, Filter: "mine"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Turn.Version != turn.Version+1 {
		t.Fatalf("projected view must reflect the claimed turn: %+v %v", page, err)
	}
	projected := page.Items[0]
	if err := repo.ConversationViews().Rebuild(ctx, f.tenant); err != nil {
		t.Fatal(err)
	}
	rebuilt, _ := queries.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20})
	if len(rebuilt.Items) != 1 || rebuilt.Items[0].Turn.Version != projected.Turn.Version || rebuilt.Items[0].Thread.UpdatedAt.UnixMicro() != projected.Thread.UpdatedAt.UnixMicro() {
		t.Fatal("projector and rebuild must agree")
	}
}

func TestConversationProjectorDecode(t *testing.T) {
	p := NewConversationProjector(nil)
	tenant, thread, turn := uuid.NewString(), uuid.NewString(), uuid.NewString()
	data, _ := json.Marshal(messaging.ConversationChanged{ThreadID: thread, TurnID: turn, TurnVersion: 1, Lifecycle: "open", Control: "sales"})
	wrap := func(kind string, version int, aggregate string) []byte {
		v, _ := json.Marshal(events.Envelope{SchemaVersion: version, EventID: uuid.NewString(), EventType: kind, TenantID: tenant, AggregateID: aggregate, Data: data})
		return v
	}
	for name, c := range map[string]struct {
		value    []byte
		ok, skip bool
	}{
		"valid":          {wrap(model.DomainEventConversationChanged, 1, thread), true, false},
		"unknown type":   {wrap("order.confirmed", 1, thread), true, true},
		"newer schema":   {wrap(model.DomainEventConversationChanged, 2, thread), true, true},
		"aggregate skew": {wrap(model.DomainEventConversationChanged, 1, uuid.NewString()), false, false},
		"not json":       {[]byte("nope"), false, false},
	} {
		_, change, ok := p.Decode(c.value)
		if ok != c.ok || (ok && (change == nil) != c.skip) {
			t.Errorf("%s: ok=%v change=%v", name, ok, change)
		}
	}
}
