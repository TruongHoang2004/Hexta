package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/google/uuid"
)

func closeInput(version int64) types.SessionInput {
	return types.SessionInput{ExpectedVersion: version, Outcome: "consulted", Note: "Asked about sizes"}
}

func auditState(t *testing.T, f salesFixture, entity, action string) map[string]any {
	t.Helper()
	var row model.AuditLog
	if err := f.db.Where("tenant_id = ? AND entity_id = ? AND action = ?", f.tenant, entity, action).First(&row).Error; err != nil {
		t.Fatal(action, err)
	}
	state := map[string]any{}
	if err := json.Unmarshal([]byte(row.AfterState), &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func TestSessionManualCloseAuthorizationMatrix(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	// Validation comes first: outcome is required and must be known; the note is bounded.
	for _, input := range []types.SessionInput{
		{ExpectedVersion: turn.Version},
		{ExpectedVersion: turn.Version, Outcome: "sold"},
		{ExpectedVersion: turn.Version, Outcome: "consulted", Note: strings.Repeat("x", types.MaxSessionNoteRunes+1)},
		{ExpectedVersion: 0, Outcome: "consulted"},
	} {
		_, err := s.CloseSession(ctx, f.tenant, f.owner, turn.ID, input)
		requireSalesStatus(t, err, 400)
	}
	// WAITING_STAFF (unassigned): no sale may close it; outsiders and other tenants are refused.
	for _, actor := range []string{f.sale, f.peer, "stranger"} {
		_, err := s.CloseSession(ctx, f.tenant, actor, turn.ID, closeInput(turn.Version))
		requireSalesStatus(t, err, 403)
	}
	_, err := s.CloseSession(ctx, f.other, f.owner, turn.ID, closeInput(turn.Version))
	requireSalesStatus(t, err, 404)

	claimed, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	// STAFF_ACTIVE: another sale is refused, a stale version conflicts.
	_, err = s.CloseSession(ctx, f.tenant, f.peer, turn.ID, closeInput(claimed.Version))
	requireSalesStatus(t, err, 403)
	_, err = s.CloseSession(ctx, f.tenant, f.sale, turn.ID, closeInput(turn.Version))
	requireSalesStatus(t, err, 409)
	closed, err := s.CloseSession(ctx, f.tenant, f.sale, turn.ID, types.SessionInput{ExpectedVersion: claimed.Version, Outcome: "order_placed", Note: "  Two shirts  "})
	if err != nil {
		t.Fatal(err)
	}
	if closed.Lifecycle != types.TurnClosed || closed.CloseCause != types.TurnCloseManual || closed.ClosedBy != f.sale || closed.Outcome != "order_placed" || closed.Note != "Two shirts" || closed.ClosedAt == nil {
		t.Fatal("manual close must record cause, closer, outcome and trimmed note", closed)
	}
	if closed.LastHandlerID == nil || *closed.LastHandlerID != f.sale {
		t.Fatal("the closing sale is the session's handler", closed.LastHandlerID)
	}
	// A closed session cannot be closed again, by anyone.
	for _, actor := range []string{f.sale, f.owner} {
		_, err = s.CloseSession(ctx, f.tenant, actor, turn.ID, closeInput(closed.Version))
		requireSalesStatus(t, err, 409)
	}
	if auditCount(t, f.db, f.tenant, turn.ID, "messaging.turn_closed_manually") != 1 {
		t.Fatal("manual close must be audited once")
	}

	// BOT_ACTIVE without an assignee: only the owner may close it.
	b := newBotFixture(t, types.ReplyModeSuggest)
	message := b.customer(t, "bot-customer", "hello")
	bot := b.detail(t, message.ThreadID).Conversation.Turn
	if bot.HandlingState() != "BOT_ACTIVE" {
		t.Fatal("configured reply mode must start the turn under bot handling", bot.HandlingState())
	}
	for _, actor := range []string{b.sale, b.peer} {
		_, err = b.messaging.CloseSession(ctx, b.tenant, actor, bot.ID, closeInput(bot.Version))
		requireSalesStatus(t, err, 403)
	}
	if _, err = b.messaging.CloseSession(ctx, b.tenant, b.owner, bot.ID, closeInput(bot.Version)); err != nil {
		t.Fatal("the owner may close a bot-handled session", err)
	}
}

func TestSessionManualCloseFencesBotWorkAndClearsNeedsHuman(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")

	// A pending proposal becomes stale through the generation bump and can never be sent.
	message := b.customer(t, "c1", "giá áo")
	b.route(t)
	detail := b.detail(t, message.ThreadID)
	turn, proposal := detail.Conversation.Turn, detail.Proposals[0].Proposal
	closed, err := b.messaging.CloseSession(ctx, b.tenant, b.owner, turn.ID, closeInput(turn.Version))
	if err != nil {
		t.Fatal(err)
	}
	if closed.Version != turn.Version+1 || closed.ControlGeneration != turn.ControlGeneration+1 {
		t.Fatal("close must bump version and control generation", closed)
	}
	// A proposal waiting for review blocks expiry but not a manual close, which clears the review flag.
	if !turn.ReviewPending || closed.ReviewPending {
		t.Fatal("manual close must clear review_pending like every control change", turn.ReviewPending, closed.ReviewPending)
	}
	detail = b.detail(t, message.ThreadID)
	if len(detail.Proposals) != 1 || !detail.Proposals[0].Stale {
		t.Fatal("the pending proposal must read as stale after close", detail.Proposals)
	}
	_, err = b.messaging.SendProposal(ctx, b.tenant, b.owner, turn.ID, proposal.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
	requireSalesStatus(t, err, 409)

	// A queued router job is dropped: closing before routing leaves no live proposal and sends nothing.
	queued := b.customer(t, "c2", "giá áo")
	queuedTurn := b.detail(t, queued.ThreadID).Conversation.Turn
	if _, err = b.messaging.CloseSession(ctx, b.tenant, b.owner, queuedTurn.ID, closeInput(queuedTurn.Version)); err != nil {
		t.Fatal(err)
	}
	b.route(t)
	for _, p := range b.proposals(t, queuedTurn.ID) {
		if p.Status == types.ProposalPending {
			t.Fatal("a router job of a closed session must not produce a proposal", p.Status)
		}
	}
	if len(outbound(b.detail(t, queued.ThreadID))) != 0 {
		t.Fatal("nothing may be sent for a closed session")
	}

	// Closing a flagged turn clears needs_human and records that it was flagged.
	flagged := b.customer(t, "c3", "hello")
	flaggedTurn := b.detail(t, flagged.ThreadID).Conversation.Turn
	handed, err := b.messaging.Command(ctx, b.tenant, b.sale, flaggedTurn.ID, "handoff", types.MessagingCommand{ExpectedVersion: flaggedTurn.Version, Reason: "customer_requested_staff"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := b.messaging.Command(ctx, b.tenant, b.sale, flaggedTurn.ID, "claim", types.MessagingCommand{ExpectedVersion: handed.Version})
	if err != nil || !claimed.NeedsHuman {
		t.Fatal("claim keeps needs-human", err)
	}
	done, err := b.messaging.CloseSession(ctx, b.tenant, b.sale, flaggedTurn.ID, closeInput(claimed.Version))
	if err != nil {
		t.Fatal(err)
	}
	if done.NeedsHuman || done.NeedsHumanReason != "" {
		t.Fatal("close must clear needs-human", done)
	}
	state := auditState(t, b.salesFixture, flaggedTurn.ID, "messaging.turn_closed_manually")
	if state["closed_while_needs_human"] != true || state["needs_human_reason"] != "customer_requested_staff" || state["outcome"] != "consulted" {
		t.Fatal("audit must record the close while flagged", state)
	}
	if _, leaked := state["note"]; leaked {
		t.Fatal("the audit must not copy the note text")
	}
}

func TestSessionManualCloseCancelsUnstartedSendsAndNextMessageOpensNewTurn(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	_, ingress, base, turn := channelThread(t, f)
	commands := NewMessagingCommandService(repository.NewMessagingRepository(f.db))
	unstarted := sendChannel(t, f, turn, "not dispatched")
	turn = loadTurn(t, f.db, f.tenant, turn.ID)
	dispatched := sendChannel(t, f, turn, "already dispatched")
	if err := f.db.Model(&model.DomainEvent{}).Where("aggregate_id = ?", dispatched.IntentID).Update("state", "published").Error; err != nil {
		t.Fatal(err)
	}
	turn = loadTurn(t, f.db, f.tenant, turn.ID)
	closed, err := commands.CloseSession(ctx, f.tenant, f.sale, turn.ID, closeInput(turn.Version))
	if err != nil {
		t.Fatal(err)
	}
	intentState := func(id string) (string, string) {
		var intent messagingmodel.OutboundIntent
		var message messagingmodel.Message
		f.db.Where("id = ?", id).First(&intent)
		f.db.Where("id = ?", intent.MessageID).First(&message)
		return intent.Status, message.DeliveryStatus
	}
	if status, delivery := intentState(unstarted.IntentID); status != types.IntentCancelled || delivery != types.DeliveryFailed {
		t.Fatal("an undispatched staff send must be cancelled", status, delivery)
	}
	if status, delivery := intentState(dispatched.IntentID); status != types.IntentPending || delivery != types.DeliverySending {
		t.Fatal("a dispatched send is reconciled from the connector, not cancelled", status, delivery)
	}
	if state := auditState(t, f, turn.ID, "messaging.turn_closed_manually"); state["cancelled_sends"] != float64(1) {
		t.Fatal("audit must count cancelled sends", state)
	}
	countTurns := func() int64 {
		var n int64
		f.db.Model(&messagingmodel.Turn{}).Where("tenant_id = ?", f.tenant).Count(&n)
		return n
	}
	// Replayed ingress (an already processed provider message) never reopens or replaces the closed session.
	if err := ingress.Consume(ctx, base); err != nil || countTurns() != 1 {
		t.Fatal("duplicate ingress opened a session", err)
	}
	// A genuinely new customer message opens a new unassigned turn under the current reply mode.
	next := base
	next.EventID, next.MessageID, next.SourceIdentity, next.Text, next.SourceAt = uuid.NewString(), "m-next", "message:m-next", "again", time.Now()
	if err := ingress.Consume(ctx, next); err != nil || countTurns() != 2 {
		t.Fatal("a new message after close must open a new session", err)
	}
	var latest messagingmodel.Turn
	if err := f.db.Where("tenant_id = ? AND lifecycle = ?", f.tenant, types.TurnOpen).First(&latest).Error; err != nil {
		t.Fatal(err)
	}
	if latest.ID == turn.ID || latest.AssigneeID != nil || latest.Version != 1 {
		t.Fatal("the new session must be unassigned", latest)
	}
	if old := loadTurn(t, f.db, f.tenant, turn.ID); old.Lifecycle != types.TurnClosed || old.Version != closed.Version {
		t.Fatal("the closed session changed")
	}
}

func TestSessionReplayNeverResurrectsClosedSession(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	if _, err := s.CloseSession(ctx, f.tenant, f.owner, turn.ID, closeInput(turn.Version)); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewMessagingRepository(f.db)
	projector := NewConversationProjector(repo.ConversationViews())
	project := func(rows []model.DomainEvent) {
		t.Helper()
		for _, e := range rows {
			tenant, change, ok := projector.Decode(envelope(t, e))
			if !ok || change == nil {
				t.Fatal("undecodable event")
			}
			if err := projector.Project(ctx, tenant, *change); err != nil {
				t.Fatal(err)
			}
		}
	}
	events := conversationEvents(t, f.db, f.tenant)
	var last messaging.ConversationChanged
	if err := json.Unmarshal([]byte(events[len(events)-1].Payload), &last); err != nil {
		t.Fatal(err)
	}
	if last.Lifecycle != types.TurnClosed || last.CloseCause != types.TurnCloseManual || last.Outcome != "consulted" || last.ClosedAt == nil {
		t.Fatal("the close event must carry the session fields", last)
	}
	// In order, then the older open snapshots again (replay/out-of-order): the view stays closed.
	project(events)
	project(events[:len(events)-1])
	var view messagingmodel.ConversationView
	if err := f.db.Where("tenant_id = ? AND thread_id = ?", f.tenant, message.ThreadID).First(&view).Error; err != nil {
		t.Fatal(err)
	}
	if view.Lifecycle != types.TurnClosed || deref(view.CloseCause) != types.TurnCloseManual || deref(view.Outcome) != "consulted" || view.ClosedAt == nil {
		t.Fatal("a replayed open snapshot resurrected the closed session", view)
	}
	// Rebuild reproduces the same session fields from the authoritative turns.
	if err := repo.ConversationViews().Rebuild(ctx, f.tenant); err != nil {
		t.Fatal(err)
	}
	f.db.Where("tenant_id = ? AND thread_id = ?", f.tenant, message.ThreadID).First(&view)
	if view.Lifecycle != types.TurnClosed || deref(view.CloseCause) != types.TurnCloseManual {
		t.Fatal("rebuild lost the session fields", view)
	}
}

// A snapshot written before session fields existed, replayed at the same turn version, must not erase the
// closed_at/close_cause/outcome that the migration backfilled into the view; a newer turn still replaces them.
func TestPreSessionSnapshotReplayKeepsBackfilledSessionFields(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	if _, err := s.CloseSession(ctx, f.tenant, f.owner, turn.ID, closeInput(turn.Version)); err != nil {
		t.Fatal(err)
	}
	views := repository.NewMessagingRepository(f.db).ConversationViews()
	if err := views.Rebuild(ctx, f.tenant); err != nil {
		t.Fatal(err)
	}
	events := conversationEvents(t, f.db, f.tenant)
	var legacy messaging.ConversationChanged
	if err := json.Unmarshal([]byte(events[len(events)-1].Payload), &legacy); err != nil {
		t.Fatal(err)
	}
	legacy.ClosedAt, legacy.CloseCause, legacy.Outcome = nil, "", ""
	if err := views.Apply(ctx, f.tenant, legacy); err != nil {
		t.Fatal(err)
	}
	load := func() messagingmodel.ConversationView {
		t.Helper()
		var view messagingmodel.ConversationView
		if err := f.db.Where("tenant_id = ? AND thread_id = ?", f.tenant, message.ThreadID).First(&view).Error; err != nil {
			t.Fatal(err)
		}
		return view
	}
	if view := load(); view.TurnVersion != legacy.TurnVersion || view.ClosedAt == nil || deref(view.CloseCause) != types.TurnCloseManual || deref(view.Outcome) != "consulted" {
		t.Fatal("a pre-session snapshot replay erased the backfilled session fields", view)
	}
	next := legacy
	next.TurnID, next.TurnVersion, next.Lifecycle = uuid.Must(uuid.NewV7()).String(), 1, types.TurnOpen
	if err := views.Apply(ctx, f.tenant, next); err != nil {
		t.Fatal(err)
	}
	if view := load(); view.TurnID != next.TurnID || view.ClosedAt != nil || view.CloseCause != nil || view.Outcome != nil {
		t.Fatal("a newer turn must replace the previous session fields", view)
	}
}

func TestSessionListPaginationAndTenantIsolation(t *testing.T) {
	f, s, message, first := messagingFixture(t)
	ctx := context.Background()
	inbound := func(body string) {
		t.Helper()
		if _, err := s.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: body, RequestKey: uuid.NewString()}); err != nil {
			t.Fatal(err)
		}
	}
	inbound("second message")
	// Session 1 expires (no outcome); session 2 closes manually; session 3 stays open.
	claimed, err := s.Command(ctx, f.tenant, f.sale, first.ID, "claim", types.MessagingCommand{ExpectedVersion: loadTurn(t, f.db, f.tenant, first.ID).Version})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Command(ctx, f.tenant, f.sale, first.ID, "resolve", types.MessagingCommand{ExpectedVersion: claimed.Version}); err != nil {
		t.Fatal(err)
	}
	backdate(t, f.db, f.tenant, first.ID, types.TurnInactivityWindow+time.Hour)
	if expireOnce(t, f.db) != 1 {
		t.Fatal("first session did not expire")
	}
	inbound("new session")
	second := s.mustLatest(t, f, message.ThreadID)
	if _, err := s.CloseSession(ctx, f.tenant, f.owner, second.ID, types.SessionInput{ExpectedVersion: second.Version, Outcome: "price_inquiry"}); err != nil {
		t.Fatal(err)
	}
	inbound("third session")
	third := s.mustLatest(t, f, message.ThreadID)

	page, err := s.Sessions(ctx, f.tenant, f.peer, message.ThreadID, types.MessagingQuery{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.NextCursor == "" || page.Items[0].Turn.ID != third.ID || page.Items[1].Turn.ID != second.ID {
		t.Fatal("sessions must be listed newest first with a cursor", page)
	}
	if page.Items[0].Turn.Lifecycle != types.TurnOpen || page.Items[1].Turn.CloseCause != types.TurnCloseManual || page.Items[1].Turn.Outcome != "price_inquiry" {
		t.Fatal("session state", page.Items)
	}
	rest, err := s.Sessions(ctx, f.tenant, f.peer, message.ThreadID, types.MessagingQuery{Limit: 2, Cursor: page.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
	if len(rest.Items) != 1 || rest.NextCursor != "" {
		t.Fatal("second page", rest)
	}
	expired := rest.Items[0]
	if expired.Turn.ID != first.ID || expired.Turn.CloseCause != types.TurnCloseExpired || expired.Turn.Outcome != "" || expired.MessageCount != 2 || expired.FirstMessageID != message.ID {
		t.Fatal("expired session: cause expired, no outcome, its messages", expired)
	}
	if expired.Turn.LastHandlerID == nil || *expired.Turn.LastHandlerID != f.sale || expired.OpenedAt.IsZero() || len(expired.Orders) != 0 {
		t.Fatal("expired session handler and opening time", expired)
	}
	// Tenant isolation: another tenant's member cannot see the conversation; outsiders are refused.
	_, err = s.Sessions(ctx, f.other, f.owner, message.ThreadID, types.MessagingQuery{Limit: 20})
	requireSalesStatus(t, err, 404)
	_, err = s.Sessions(ctx, f.tenant, "stranger", message.ThreadID, types.MessagingQuery{Limit: 20})
	requireSalesStatus(t, err, 403)
	for _, input := range []types.MessagingQuery{{Limit: 0}, {Limit: 101}, {Limit: 20, Cursor: "nope"}} {
		_, err = s.Sessions(ctx, f.tenant, f.sale, message.ThreadID, input)
		requireSalesStatus(t, err, 400)
	}
	_, err = s.Sessions(ctx, f.tenant, f.sale, "not-a-uuid", types.MessagingQuery{Limit: 20})
	requireSalesStatus(t, err, 400)
}

// mustLatest reads the thread's latest turn through the query side.
func (f *messagingFacade) mustLatest(t *testing.T, fx salesFixture, thread string) types.MessagingTurn {
	t.Helper()
	detail, err := f.Detail(context.Background(), fx.tenant, fx.sale, thread, types.MessagingQuery{Limit: 1})
	if err != nil || detail.Conversation.Turn == nil {
		t.Fatal(err)
	}
	return *detail.Conversation.Turn
}

func TestSessionOutcomeEditAuthorization(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	claimed, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	// Return to the bot (allowed once a bot reply mode is chosen) clears the assignee, but the sale stays the
	// session's last handler.
	if _, err := NewReplySettingsService(repository.NewMessagingRepository(f.db)).Update(ctx, f.tenant, f.owner, types.ReplySettingsUpdate{Mode: types.ReplyModeSuggest}); err != nil {
		t.Fatal(err)
	}
	returned, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: claimed.Version})
	if err != nil || returned.AssigneeID != nil {
		t.Fatal(err)
	}
	// An open session has no outcome to edit yet.
	_, err = s.UpdateSessionOutcome(ctx, f.tenant, f.owner, turn.ID, closeInput(returned.Version))
	requireSalesStatus(t, err, 409)
	backdate(t, f.db, f.tenant, turn.ID, types.TurnInactivityWindow+time.Hour)
	if expireOnce(t, f.db) != 1 {
		t.Fatal("session did not expire")
	}
	expired := loadTurn(t, f.db, f.tenant, turn.ID)
	if expired.Outcome != "" || expired.CloseCause != types.TurnCloseExpired || expired.LastHandlerID == nil || *expired.LastHandlerID != f.sale {
		t.Fatal("expired session starts without an outcome and keeps its handler", expired)
	}
	for _, actor := range []string{f.peer, "stranger"} {
		_, err = s.UpdateSessionOutcome(ctx, f.tenant, actor, turn.ID, closeInput(expired.Version))
		requireSalesStatus(t, err, 403)
	}
	_, err = s.UpdateSessionOutcome(ctx, f.other, f.owner, turn.ID, closeInput(expired.Version))
	requireSalesStatus(t, err, 404)
	_, err = s.UpdateSessionOutcome(ctx, f.tenant, f.sale, turn.ID, types.SessionInput{ExpectedVersion: expired.Version, Outcome: "unknown"})
	requireSalesStatus(t, err, 400)
	edited, err := s.UpdateSessionOutcome(ctx, f.tenant, f.sale, turn.ID, types.SessionInput{ExpectedVersion: expired.Version, Outcome: "no_response", Note: "Customer went quiet"})
	if err != nil || edited.Outcome != "no_response" || edited.Note != "Customer went quiet" || edited.Version != expired.Version+1 || edited.Lifecycle != types.TurnClosed {
		t.Fatal("the last handler may record the outcome", edited, err)
	}
	_, err = s.UpdateSessionOutcome(ctx, f.tenant, f.owner, turn.ID, closeInput(expired.Version))
	requireSalesStatus(t, err, 409)
	if _, err = s.UpdateSessionOutcome(ctx, f.tenant, f.owner, turn.ID, closeInput(edited.Version)); err != nil {
		t.Fatal("the owner may edit any session outcome", err)
	}
	if auditCount(t, f.db, f.tenant, turn.ID, "messaging.turn_outcome_updated") != 2 {
		t.Fatal("outcome edits must be audited")
	}
	if stored := loadTurn(t, f.db, f.tenant, turn.ID); stored.Lifecycle != types.TurnClosed || stored.Outcome != "consulted" {
		t.Fatal("outcome edit must not reopen the session", stored)
	}
}

func TestSessionOrderLinking(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	link := func(tenant, actor, id string) (*model.Order, int) {
		t.Helper()
		row, err := f.service.SaveDraft(ctx, tenant, actor, "", types.DraftInput{MessagingTurnID: &id})
		if err != nil {
			return nil, err.GetHttpStatus()
		}
		return row, 0
	}
	order, status := link(f.tenant, f.sale, turn.ID)
	if status != 0 || order.MessagingTurnID == nil || *order.MessagingTurnID != turn.ID || order.Status != model.OrderStatusDraft {
		t.Fatal("a draft created from Desk must link its session", order, status)
	}
	// Editing the draft never moves or clears the link.
	other := uuid.NewString()
	updated, err := f.service.SaveDraft(ctx, f.tenant, f.sale, order.ID, types.DraftInput{Version: order.Version, EmployeeID: &f.sale, MessagingTurnID: &other})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := f.service.Order(ctx, f.tenant, f.sale, updated.ID)
	if err != nil || stored.MessagingTurnID == nil || *stored.MessagingTurnID != turn.ID {
		t.Fatal("the session link is fixed at creation", stored, err)
	}
	// Another tenant's or a missing session is not found; malformed IDs are rejected. The owner is a member of both
	// tenants, so the cross-tenant attempt passes membership and must be stopped by the session's tenant scope.
	if _, status = link(f.other, f.owner, turn.ID); status != 404 {
		t.Fatal("cross-tenant session link must not find the other tenant's session", status)
	}
	if _, status = link(f.other, f.sale, turn.ID); status != 403 {
		t.Fatal("a non-member must be refused before the session is looked up", status)
	}
	if _, status = link(f.tenant, f.sale, uuid.NewString()); status != 404 {
		t.Fatal("unknown session", status)
	}
	if _, status = link(f.tenant, f.sale, "nope"); status != 400 {
		t.Fatal("malformed session ID", status)
	}
	// Orders without a session keep working unchanged.
	if plain, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", types.DraftInput{}); err != nil || plain.MessagingTurnID != nil {
		t.Fatal("plain drafts have no session", err)
	}
	page, err := s.Sessions(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || len(page.Items) != 1 || len(page.Items[0].Orders) != 1 || page.Items[0].Orders[0].ID != order.ID || page.Items[0].Orders[0].OrderNumber != order.OrderNumber {
		t.Fatal("the session list must show its linked order", page, err)
	}
	// A closed session no longer accepts new orders.
	current := loadTurn(t, f.db, f.tenant, turn.ID)
	if _, err := s.CloseSession(ctx, f.tenant, f.owner, turn.ID, closeInput(current.Version)); err != nil {
		t.Fatal(err)
	}
	if _, status = link(f.tenant, f.sale, turn.ID); status != 409 {
		t.Fatal("closed session must refuse new linked orders", status)
	}
}
