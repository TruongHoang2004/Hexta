package service

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// backdate moves a turn's recorded activity into the past without touching its version, like real elapsed time.
func backdate(t *testing.T, db *gorm.DB, tenant, turn string, age time.Duration) time.Time {
	t.Helper()
	at := time.Now().UTC().Add(-age).Truncate(time.Microsecond)
	if err := db.Table("messaging_turns").Where("tenant_id = ? AND id = ?", tenant, turn).Update("last_activity_at", at).Error; err != nil {
		t.Fatal(err)
	}
	return at
}

func loadTurn(t *testing.T, db *gorm.DB, tenant, id string) types.MessagingTurn {
	t.Helper()
	turn, err := repository.NewMessagingRepository(db).Turn(context.Background(), tenant, id)
	if err != nil {
		t.Fatal(err)
	}
	return *turn
}

func expireOnce(t *testing.T, db *gorm.DB) int {
	t.Helper()
	_, closed, err := NewTurnExpiryService(repository.NewMessagingRepository(db)).ExpireDue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return closed
}

func auditCount(t *testing.T, db *gorm.DB, tenant, entity, action string) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&model.AuditLog{}).Where("tenant_id = ? AND entity_id = ? AND action = ?", tenant, entity, action).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

func TestMessagingResolveNeedsHumanAuthorization(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	if !turn.NeedsHuman {
		t.Fatal("a new turn without a bot must need a person")
	}
	// Unassigned work cannot be resolved by anyone, including the owner.
	for _, actor := range []string{f.sale, f.owner} {
		_, err := s.Command(ctx, f.tenant, actor, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: turn.Version})
		requireSalesStatus(t, err, 403)
	}
	claimed, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || !claimed.NeedsHuman {
		t.Fatal("claim must keep needs-human", err)
	}
	for _, actor := range []string{f.peer, f.owner, "stranger"} {
		_, err := s.Command(ctx, f.tenant, actor, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version})
		requireSalesStatus(t, err, 403)
	}
	_, err = s.Command(ctx, f.other, f.sale, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version})
	requireSalesStatus(t, err, 403)
	_, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version - 1})
	requireSalesStatus(t, err, 409)

	old := backdate(t, f.db, f.tenant, turn.ID, 30*time.Hour)
	resolved, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.NeedsHuman || resolved.NeedsHumanReason != "" || resolved.HandlingState() != "STAFF_ACTIVE" || resolved.Control != "sales" || *resolved.AssigneeID != f.sale {
		t.Fatal("resolve must clear the flag and keep staff control", resolved)
	}
	if resolved.ControlGeneration != claimed.ControlGeneration || resolved.Version != claimed.Version+1 {
		t.Fatal("resolve is not a control change", resolved)
	}
	if !resolved.LastActivityAt.After(old.Add(29 * time.Hour)) {
		t.Fatal("resolve must restart the inactivity window")
	}
	if auditCount(t, f.db, f.tenant, turn.ID, "messaging.resolve") != 1 {
		t.Fatal("resolve was not audited")
	}
	_, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: resolved.Version})
	requireSalesStatus(t, err, 409)
	// A resolved turn whose activity is recent does not expire.
	if expireOnce(t, f.db) != 0 {
		t.Fatal("resolved turn expired before a new 24-hour window passed")
	}
}

func TestTurnExpiryClosesInactiveTurnAndNextMessageOpensNewTurn(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	claimed, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	backdate(t, f.db, f.tenant, turn.ID, 72*time.Hour)
	if expireOnce(t, f.db) != 0 || loadTurn(t, f.db, f.tenant, turn.ID).Lifecycle != types.TurnOpen {
		t.Fatal("needs-human must block expiry")
	}
	resolved, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version})
	if err != nil {
		t.Fatal(err)
	}
	backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow-time.Minute)
	if expireOnce(t, f.db) != 0 {
		t.Fatal("turn closed before 24 hours of inactivity")
	}
	backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow+time.Minute)
	before := len(conversationEvents(t, f.db, f.tenant))
	if expireOnce(t, f.db) != 1 {
		t.Fatal("inactive turn did not close")
	}
	closed := loadTurn(t, f.db, f.tenant, turn.ID)
	if closed.Lifecycle != types.TurnClosed || closed.ClosedAt == nil || closed.Version != resolved.Version+1 || closed.ControlGeneration != resolved.ControlGeneration+1 {
		t.Fatal("close did not fence the turn", closed)
	}
	if auditCount(t, f.db, f.tenant, turn.ID, "messaging.turn_expired") != 1 {
		t.Fatal("expiry was not audited")
	}
	events := conversationEvents(t, f.db, f.tenant)
	var change messaging.ConversationChanged
	if len(events) != before+1 || json.Unmarshal([]byte(events[len(events)-1].Payload), &change) != nil || change.Lifecycle != types.TurnClosed || change.TurnID != turn.ID {
		t.Fatal("closing must emit the closed snapshot for the read model")
	}
	// Idempotent: a second run, or another instance, finds nothing to do.
	if expireOnce(t, f.db) != 0 || len(conversationEvents(t, f.db, f.tenant)) != before+1 {
		t.Fatal("expiry is not idempotent")
	}
	// The closed turn is read-only.
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, types.MessagingCommand{ExpectedVersion: closed.Version, ContextRevision: closed.ContextRevision, Body: "late", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 409)
	_, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "handoff", types.MessagingCommand{ExpectedVersion: closed.Version, Reason: "staff_handoff"})
	requireSalesStatus(t, err, 409)
	detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || detail.Conversation.Turn == nil || detail.Conversation.Turn.Lifecycle != types.TurnClosed || len(detail.Messages) != 1 {
		t.Fatal("closed conversation must keep its history and show the closed turn", err)
	}
	all, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20})
	if err != nil || len(all.Items) != 1 || all.Items[0].Turn.Lifecycle != types.TurnClosed {
		t.Fatal("closed conversation must stay listed under all", err)
	}
	for _, filter := range []string{"mine", "waiting", "unassigned", "needs_human", "bot"} {
		page, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20, Filter: filter})
		if err != nil || len(page.Items) != 0 {
			t.Fatal("closed conversation leaked into an open-work filter", filter, err)
		}
	}

	// A genuinely new customer message opens a new unassigned turn; history stays.
	if _, err := s.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Back again", RequestKey: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	detail, err = s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	next := detail.Conversation.Turn
	if next.ID == turn.ID || next.Lifecycle != types.TurnOpen || next.AssigneeID != nil || next.HandlingState() != "WAITING_STAFF" || next.Version != 1 {
		t.Fatal("next message must open a new unassigned turn", next)
	}
	if len(detail.Messages) != 2 || detail.Messages[0].TurnID != next.ID || detail.Messages[1].TurnID != turn.ID {
		t.Fatal("history across turns", detail.Messages)
	}
	if loadTurn(t, f.db, f.tenant, turn.ID).Lifecycle != types.TurnClosed {
		t.Fatal("old turn reopened")
	}
	// The read model replaces the closed turn with the newer one even though its version restarts at 1.
	repo := repository.NewMessagingRepository(f.db)
	projector := NewConversationProjector(repo.ConversationViews())
	if err := repo.ConversationViews().Rebuild(ctx, f.tenant); err != nil {
		t.Fatal(err)
	}
	for _, e := range conversationEvents(t, f.db, f.tenant) {
		tenant, change, ok := projector.Decode(envelope(t, e))
		if !ok || change == nil {
			t.Fatal("undecodable event")
		}
		if err := projector.Project(ctx, tenant, *change); err != nil {
			t.Fatal(err)
		}
	}
	var view messagingmodel.ConversationView
	if err := f.db.Where("tenant_id = ? AND thread_id = ?", f.tenant, message.ThreadID).First(&view).Error; err != nil || view.TurnID != next.ID || view.Lifecycle != types.TurnOpen {
		t.Fatal("replayed older-turn events overwrote the newer turn", view, err)
	}
	summary, err := s.Summary(ctx, f.tenant, f.peer)
	if err != nil || summary.WaitingStaff != 1 {
		t.Fatal("waiting count must include the new turn", summary, err)
	}
	_, err = s.Summary(ctx, f.tenant, "stranger")
	requireSalesStatus(t, err, 403)
	_, err = s.Summary(ctx, f.other, f.sale)
	requireSalesStatus(t, err, 403)
}

func TestTurnExpiryWaitsForHeldThreadAndConcurrentInstances(t *testing.T) {
	f := salesTestFixture(t)
	s := NewMessagingService(repository.NewMessagingRepository(f.db))
	ctx := context.Background()
	turns := make([]types.MessagingTurn, 0, 6)
	for i := range 6 {
		m, err := s.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: uuid.NewString(), DisplayName: "Customer", Body: "Hi", RequestKey: uuid.NewString()})
		if err != nil {
			t.Fatal(err)
		}
		turn := loadTurn(t, f.db, f.tenant, m.TurnID)
		if err := f.db.Table("messaging_turns").Where("id = ?", turn.ID).Update("needs_human", false).Error; err != nil {
			t.Fatal(err)
		}
		backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow+time.Duration(i+1)*time.Minute)
		turns = append(turns, turn)
	}
	// A transaction holding the thread (a concurrent inbound message) makes expiry skip it instead of waiting.
	held := f.db.Begin()
	if err := held.Exec("SELECT id FROM messaging_threads WHERE id = ? FOR UPDATE", turns[0].ThreadID).Error; err != nil {
		t.Fatal(err)
	}
	if closed := expireOnce(t, f.db); closed != 5 {
		held.Rollback()
		t.Fatalf("expected the 5 unheld turns to close, got %d", closed)
	}
	held.Rollback()
	if loadTurn(t, f.db, f.tenant, turns[0].ID).Lifecycle != types.TurnOpen {
		t.Fatal("locked thread was closed")
	}
	// Reopen every turn as due and race four job instances: each turn closes exactly once.
	if err := f.db.Exec("UPDATE messaging_turns SET lifecycle = 'open', closed_at = NULL, close_cause = NULL WHERE tenant_id = ?", f.tenant).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Where("tenant_id = ? AND action = ?", f.tenant, "messaging.turn_expired").Delete(&model.AuditLog{}).Error; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	totals := make(chan int, 4)
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, closed, err := NewTurnExpiryService(repository.NewMessagingRepository(f.db)).ExpireDue(ctx)
			if err != nil {
				t.Error(err)
			}
			totals <- closed
		}()
	}
	wg.Wait()
	close(totals)
	sum := 0
	for n := range totals {
		sum += n
	}
	// SKIP LOCKED may leave a turn for the next run; nothing is ever closed twice.
	sum += expireOnce(t, f.db)
	if sum != 6 {
		t.Fatalf("turns closed %d times in total, want 6", sum)
	}
	for _, turn := range turns {
		if auditCount(t, f.db, f.tenant, turn.ID, "messaging.turn_expired") != 1 || loadTurn(t, f.db, f.tenant, turn.ID).Lifecycle != types.TurnClosed {
			t.Fatal("turn was not closed exactly once", turn.ID)
		}
	}
}

func TestTurnExpiryRejectsStaleObservedState(t *testing.T) {
	f, _, _, turn := messagingFixture(t)
	ctx := context.Background()
	if err := f.db.Table("messaging_turns").Where("id = ?", turn.ID).Update("needs_human", false).Error; err != nil {
		t.Fatal(err)
	}
	backdate(t, f.db, f.tenant, turn.ID, 30*time.Hour)
	observed := loadTurn(t, f.db, f.tenant, turn.ID)
	// Activity lands after the job read the turn: the conditional close must not apply.
	backdate(t, f.db, f.tenant, turn.ID, time.Minute)
	repo := repository.NewMessagingRepository(f.db)
	err := repo.WithinExpiry(ctx, func(r repository.TurnExpiryStore) *apperrors.Error {
		ok, err := r.CloseTurn(ctx, observed, types.TurnClosure{Cause: types.TurnCloseExpired}, time.Now().UTC())
		if err == nil && ok {
			t.Error("stale observation closed a turn with newer activity")
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if loadTurn(t, f.db, f.tenant, turn.ID).Lifecycle != types.TurnOpen {
		t.Fatal("turn closed from stale state")
	}
}

// channelThread creates a real (Facebook) conversation claimed by the sale, ready for channel sends.
func channelThread(t *testing.T, f salesFixture) (conn string, ingress *IngressService, base messaging.MessageReceived, turn types.MessagingTurn) {
	t.Helper()
	ctx := context.Background()
	conn = uuid.NewString()
	if err := f.db.Exec("INSERT INTO messaging_connections(id,tenant_id,version,platform,account_id,enabled) VALUES(?,?,1,'facebook','page',true)", conn, f.tenant).Error; err != nil {
		t.Fatal(err)
	}
	ingress = NewIngressService(repository.NewIngressRepository(f.db))
	base = messaging.MessageReceived{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: f.tenant, OccurredAt: time.Now()}, Platform: "facebook", ConnectionID: conn, ConnectionVersion: 1, AccountID: "page", ConversationKey: "psid", MessageID: "m1", SourceIdentity: "message:m1", Sender: "psid", Recipient: "page", Direction: "inbound", Text: "hello", SourceAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	if err := ingress.Consume(ctx, base); err != nil {
		t.Fatal(err)
	}
	var row messagingmodel.Turn
	if err := f.db.Where("tenant_id = ?", f.tenant).First(&row).Error; err != nil {
		t.Fatal(err)
	}
	claimed, err := NewMessagingCommandService(repository.NewMessagingRepository(f.db)).Command(ctx, f.tenant, f.sale, row.ID, "claim", types.MessagingCommand{ExpectedVersion: row.Version})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Table("messaging_turns").Where("id = ?", row.ID).Update("needs_human", false).Error; err != nil {
		t.Fatal(err)
	}
	return conn, ingress, base, loadTurn(t, f.db, f.tenant, claimed.ID)
}

func sendChannel(t *testing.T, f salesFixture, turn types.MessagingTurn, text string) delivery.SendRequested {
	t.Helper()
	if _, err := NewMessagingCommandService(repository.NewMessagingRepository(f.db)).SendChannel(context.Background(), f.tenant, f.sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: text, ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision}); err != nil {
		t.Fatal(err)
	}
	var event model.DomainEvent
	if err := f.db.Where("tenant_id = ? AND event_type = ? AND payload->>'text' = ?", f.tenant, model.DomainEventOutboundIntent, text).First(&event).Error; err != nil {
		t.Fatal(err)
	}
	var request delivery.SendRequested
	if err := json.Unmarshal([]byte(event.Payload), &request); err != nil {
		t.Fatal(err)
	}
	return request
}

func TestTurnExpiryActivityCountsOnlyConfirmedSends(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	conn, _, _, turn := channelThread(t, f)
	deliveries := NewDeliveryService(repository.NewDeliveryRepository(f.db))
	result := func(r delivery.SendRequested, status, category string, at time.Time) {
		t.Helper()
		e := delivery.SendCompleted{Meta: events.Meta{EventID: uuid.NewString(), TenantID: f.tenant, OccurredAt: at}, IntentID: r.IntentID, AttemptID: r.AttemptID, ConnectionID: conn, Status: status, ErrorCategory: category}
		if err := deliveries.ConsumeResult(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	old := backdate(t, f.db, f.tenant, turn.ID, 30*time.Hour)
	turn = loadTurn(t, f.db, f.tenant, turn.ID)

	// A pending send is not activity.
	failed := sendChannel(t, f, turn, "will fail")
	turn = loadTurn(t, f.db, f.tenant, turn.ID)
	if !turn.LastActivityAt.Equal(old) {
		t.Fatal("pending send counted as activity")
	}
	result(failed, delivery.StatusFailed, delivery.ErrorTransport, time.Now())
	unknown := sendChannel(t, f, turn, "lost outcome")
	turn = loadTurn(t, f.db, f.tenant, turn.ID)
	result(unknown, delivery.StatusUnknown, delivery.ErrorTransport, time.Now())
	if !loadTurn(t, f.db, f.tenant, turn.ID).LastActivityAt.Equal(old) {
		t.Fatal("failed or unknown send counted as activity")
	}
	// A confirmed send counts from the provider outcome time, not processing time.
	confirmed := sendChannel(t, f, turn, "delivered")
	sentAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	result(confirmed, delivery.StatusSent, "", sentAt)
	turn = loadTurn(t, f.db, f.tenant, turn.ID)
	if !turn.LastActivityAt.Equal(sentAt) {
		t.Fatal("confirmed send must set activity to its source time", turn.LastActivityAt, sentAt)
	}
	// A redelivered (duplicate) result never extends activity again.
	result(confirmed, delivery.StatusSent, "", time.Now())
	if !loadTurn(t, f.db, f.tenant, turn.ID).LastActivityAt.Equal(sentAt) {
		t.Fatal("duplicate result extended activity")
	}
	// A confirmation reporting an earlier provider time never moves activity backwards.
	earlier := sendChannel(t, f, turn, "confirmed out of order")
	result(earlier, delivery.StatusSent, "", sentAt.Add(-time.Hour))
	if !loadTurn(t, f.db, f.tenant, turn.ID).LastActivityAt.Equal(sentAt) {
		t.Fatal("out-of-order confirmation moved activity backwards")
	}
}

func TestTurnExpiryDefersInFlightSendsAndCancelsStaleIntents(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	conn, _, _, turn := channelThread(t, f)
	pending := sendChannel(t, f, turn, "still sending")
	backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow+time.Hour)
	if expireOnce(t, f.db) != 0 {
		t.Fatal("a turn with a recent in-flight send must not close")
	}
	// Once the intent is past the connector's TTL it can no longer start: the turn closes and the intent is cancelled.
	if err := f.db.Exec("UPDATE outbound_intents SET created_at = ? WHERE id = ?", time.Now().UTC().Add(-outboundSettleWindow-time.Minute), pending.IntentID).Error; err != nil {
		t.Fatal(err)
	}
	if expireOnce(t, f.db) != 1 {
		t.Fatal("turn with only stale pending sends did not close")
	}
	var intent messagingmodel.OutboundIntent
	var attempt messagingmodel.OutboundAttempt
	var message messagingmodel.Message
	f.db.Where("id = ?", pending.IntentID).First(&intent)
	f.db.Where("id = ?", pending.AttemptID).First(&attempt)
	f.db.Where("id = ?", intent.MessageID).First(&message)
	if intent.Status != types.IntentCancelled || attempt.Status != types.IntentCancelled || message.DeliveryStatus != types.DeliveryUnknown {
		t.Fatal("unstarted send was not cancelled", intent.Status, attempt.Status, message.DeliveryStatus)
	}
	// A late provider confirmation is recorded truthfully but never reopens or extends the closed turn.
	closed := loadTurn(t, f.db, f.tenant, turn.ID)
	late := delivery.SendCompleted{Meta: events.Meta{EventID: uuid.NewString(), TenantID: f.tenant, OccurredAt: time.Now()}, IntentID: pending.IntentID, AttemptID: pending.AttemptID, ConnectionID: conn, Status: delivery.StatusSent}
	if err := NewDeliveryService(repository.NewDeliveryRepository(f.db)).ConsumeResult(ctx, late); err != nil {
		t.Fatal(err)
	}
	after := loadTurn(t, f.db, f.tenant, turn.ID)
	if after.Lifecycle != types.TurnClosed || !after.LastActivityAt.Equal(closed.LastActivityAt) || after.Version != closed.Version {
		t.Fatal("late confirmation changed the closed turn")
	}
}

func TestTurnExpiryReplayedIngressDoesNotOpenTurn(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	_, ingress, base, turn := channelThread(t, f)
	last := backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow+time.Hour)
	if expireOnce(t, f.db) != 1 {
		t.Fatal("turn did not close")
	}
	countTurns := func() int64 {
		var n int64
		f.db.Model(&messagingmodel.Turn{}).Where("tenant_id = ?", f.tenant).Count(&n)
		return n
	}
	// Exact redelivery of an already processed message is a duplicate.
	if err := ingress.Consume(ctx, base); err != nil || countTurns() != 1 {
		t.Fatal("duplicate ingress opened a turn", err)
	}
	// A delayed message sent before the closed turn's last activity is history only.
	delayed := base
	delayed.EventID, delayed.MessageID, delayed.SourceIdentity, delayed.Text = uuid.NewString(), "m-delayed", "message:m-delayed", "delayed"
	delayed.SourceAt = last.Add(-time.Minute)
	if err := ingress.Consume(ctx, delayed); err != nil || countTurns() != 1 {
		t.Fatal("delayed historical message opened a turn", err)
	}
	closed := loadTurn(t, f.db, f.tenant, turn.ID)
	if closed.Lifecycle != types.TurnClosed || !closed.LastActivityAt.Equal(last) {
		t.Fatal("historical message reopened or extended the closed turn")
	}
	var kept messagingmodel.Message
	if err := f.db.Where("tenant_id = ? AND body = ?", f.tenant, "delayed").First(&kept).Error; err != nil || kept.TurnID != turn.ID {
		t.Fatal("historical message must be kept on the closed turn", err)
	}
	// A genuinely new message opens a new unassigned turn from its source time.
	fresh := base
	fresh.EventID, fresh.MessageID, fresh.SourceIdentity, fresh.Text = uuid.NewString(), "m-new", "message:m-new", "new question"
	fresh.SourceAt = time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
	if err := ingress.Consume(ctx, fresh); err != nil || countTurns() != 2 {
		t.Fatal("new message did not open a new turn", err)
	}
	next, err := repository.NewMessagingRepository(f.db).CurrentTurn(ctx, f.tenant, turn.ThreadID)
	if err != nil || next == nil || next.AssigneeID != nil || next.HandlingState() != "WAITING_STAFF" || !next.LastActivityAt.Equal(fresh.SourceAt) {
		t.Fatal("new turn must be unassigned and start at the message source time", next, err)
	}
}
