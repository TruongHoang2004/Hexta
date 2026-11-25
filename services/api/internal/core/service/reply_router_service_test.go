package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/replymodel"
	"github.com/google/uuid"
)

// fakeAI stands in for the Python AIService; tests never call a live provider.
type fakeAI struct {
	mu        sync.Mutex
	err       error
	startErrs []error // consumed by successive StartRun calls before err applies; such a call never reaches the AI
	accepted  bool    // a StartRun committed the run; until then GetRun reports the run as not found
	pending   int     // GetRun/StartRun report running this many more times
	run       types.AIRun
	started   []types.AIRunRequest
	polled    int
	cancelled []string
}

func (f *fakeAI) result(req types.AIRunRequest) types.AIRun {
	if f.pending > 0 {
		f.pending--
		return types.AIRun{RunID: req.RunID, Status: types.AIRunRunning}
	}
	run := f.run
	run.RunID = req.RunID
	if run.Status == "" {
		run.Status = types.AIRunSucceeded
	}
	run.ContextRevision, run.ControlGeneration = req.ContextRevision, req.ControlGeneration
	return run
}

func (f *fakeAI) StartRun(_ context.Context, req types.AIRunRequest) (types.AIRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, req)
	if len(f.startErrs) > 0 {
		err := f.startErrs[0]
		f.startErrs = f.startErrs[1:]
		return types.AIRun{}, err
	}
	if f.err != nil {
		return types.AIRun{}, f.err
	}
	f.accepted = true
	return f.result(req), nil
}

func (f *fakeAI) GetRun(_ context.Context, _, runID string) (types.AIRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.polled++
	if f.err != nil {
		return types.AIRun{}, f.err
	}
	if !f.accepted {
		return types.AIRun{}, types.ErrAIRunNotFound
	}
	return f.result(f.started[len(f.started)-1]), nil
}

func (f *fakeAI) CancelRun(_ context.Context, _, runID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cancelled = append(f.cancelled, runID)
	return nil
}

type botFixture struct {
	salesFixture
	messaging *messagingFacade
	router    *ReplyRouterService
	settings  *ReplySettingsService
	scripts   *ScriptService
	ai        *fakeAI
}

// newBotFixture configures the tenant's reply mode (so new turns start under bot handling) and the router.
func newBotFixture(t *testing.T, mode string) botFixture {
	t.Helper()
	f := salesTestFixture(t)
	repo := repository.NewMessagingRepository(f.db)
	b := botFixture{salesFixture: f, messaging: NewMessagingService(repo), settings: NewReplySettingsService(repo),
		scripts: NewScriptService(repository.NewScriptRepository(f.db), repository.NewTenantDBRepository(f.db)), ai: &fakeAI{}}
	b.router = NewReplyRouterService(repo, b.scripts, b.ai)
	if mode != "" {
		if _, err := b.settings.Update(context.Background(), f.tenant, f.owner, types.ReplySettingsUpdate{Mode: mode}); err != nil {
			t.Fatal(err)
		}
	}
	return b
}

func (b botFixture) activeScript(t *testing.T, trigger, answer string) types.ScriptVersion {
	t.Helper()
	ctx := context.Background()
	script, err := b.scripts.Create(ctx, b.tenant, b.owner, uuid.NewString(), types.ScriptCreateInput{Name: trigger, Content: types.ScriptContent{Triggers: []string{trigger}, Answer: answer}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.scripts.Activate(ctx, b.tenant, b.owner, script.ID, script.Versions[0].ID, script.Version); err != nil {
		t.Fatal(err)
	}
	return script.Versions[0]
}

func (b botFixture) customer(t *testing.T, key, body string) *types.MessagingMessage {
	t.Helper()
	message, err := b.messaging.Inbound(context.Background(), b.tenant, b.sale, types.MessagingInbound{ExternalKey: key, DisplayName: "Customer", Body: body, RequestKey: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	return message
}

func (b botFixture) route(t *testing.T) int {
	t.Helper()
	_, decided, err := b.router.RouteDue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return decided
}

func (b botFixture) detail(t *testing.T, thread string) types.MessagingDetail {
	t.Helper()
	detail, err := b.messaging.Detail(context.Background(), b.tenant, b.sale, thread, types.MessagingQuery{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	return detail
}

func (b botFixture) proposals(t *testing.T, turn string) []replymodel.Proposal {
	t.Helper()
	rows := []replymodel.Proposal{}
	if err := b.db.Where("tenant_id = ? AND turn_id = ?", b.tenant, turn).Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

func outbound(detail types.MessagingDetail) []types.MessagingMessage {
	result := []types.MessagingMessage{}
	for _, m := range detail.Messages {
		if m.Direction == "outbound" {
			result = append(result, m)
		}
	}
	return result
}

func TestReplySettingsOwnerOnlyWithOptimisticVersion(t *testing.T) {
	b := newBotFixture(t, "")
	ctx := context.Background()
	current, err := b.settings.Get(ctx, b.tenant, b.sale)
	if err != nil || current.Configured || current.Mode != types.ReplyModeSuggest || current.Version != 0 {
		t.Fatal("unconfigured tenants read as suggest", current, err)
	}
	// Before the owner configures the bot, a new conversation keeps waiting for staff and nothing is routed.
	message := b.customer(t, "before-bot", "giá áo")
	if turn := b.detail(t, message.ThreadID).Conversation.Turn; turn.HandlingState() != "WAITING_STAFF" || len(b.proposals(t, turn.ID)) != 0 {
		t.Fatal("unconfigured bot must not handle new turns", turn)
	}
	for _, actor := range []string{b.sale, b.peer, "stranger"} {
		_, err := b.settings.Update(ctx, b.tenant, actor, types.ReplySettingsUpdate{Mode: types.ReplyModeAutoSend})
		requireSalesStatus(t, err, 403)
	}
	_, err = b.settings.Update(ctx, b.other, b.sale, types.ReplySettingsUpdate{Mode: types.ReplyModeSuggest})
	requireSalesStatus(t, err, 403)
	_, err = b.settings.Update(ctx, b.tenant, b.owner, types.ReplySettingsUpdate{Mode: "always"})
	requireSalesStatus(t, err, 400)
	saved, err := b.settings.Update(ctx, b.tenant, b.owner, types.ReplySettingsUpdate{Mode: types.ReplyModeAutoSend})
	if err != nil || !saved.Configured || saved.Mode != types.ReplyModeAutoSend || saved.Version != 1 {
		t.Fatal("owner must configure the reply mode", saved, err)
	}
	_, err = b.settings.Update(ctx, b.tenant, b.owner, types.ReplySettingsUpdate{Mode: types.ReplyModeSuggest})
	requireSalesStatus(t, err, 409)
	back, err := b.settings.Update(ctx, b.tenant, b.owner, types.ReplySettingsUpdate{Mode: types.ReplyModeSuggest, ExpectedVersion: 1})
	if err != nil || back.Mode != types.ReplyModeSuggest || back.Version != 2 {
		t.Fatal(back, err)
	}
	if auditCount(t, b.db, b.tenant, b.tenant, "messaging.reply_mode_changed") != 2 {
		t.Fatal("reply mode changes must be audited")
	}
	// Once configured, a new conversation starts under bot handling and is routed.
	message = b.customer(t, "after-bot", "xin chào")
	if turn := b.detail(t, message.ThreadID).Conversation.Turn; turn.HandlingState() != "BOT_ACTIVE" || len(b.proposals(t, turn.ID)) != 1 {
		t.Fatal("configured bot must handle new turns", turn)
	}
}

func TestRouterScriptHitInSuggestModeNeverSendsWithoutSale(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	version := b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")
	message := b.customer(t, "c1", "Shop ơi cho hỏi GIÁ ÁO này?")
	if b.route(t) != 1 || len(b.ai.started) != 0 {
		t.Fatal("a script hit must not call the AI model")
	}
	for range 3 {
		b.route(t)
	}
	detail := b.detail(t, message.ThreadID)
	if len(outbound(detail)) != 0 {
		t.Fatal("suggest mode must never send without a sale action")
	}
	if len(detail.Proposals) != 1 {
		t.Fatal("expected one live proposal", detail.Proposals)
	}
	p := detail.Proposals[0]
	if p.Proposal.Status != types.ProposalPending || p.Proposal.Route != types.RouteScript || p.Proposal.Source != types.SourceScript || p.Proposal.ScriptVersionID != version.ID || p.Proposal.Text != "Dạ áo giá 200k ạ" || p.Stale || p.ScriptName != "giá áo" || p.ScriptVersion != 1 {
		t.Fatalf("unexpected proposal %+v", p)
	}
	if p.Proposal.ControlGeneration != detail.Conversation.Turn.ControlGeneration {
		t.Fatal("proposal must record its control generation")
	}
	// A newer customer message supersedes the older proposal; the turn keeps one live proposal.
	b.customer(t, "c1", "còn màu đen không")
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn muốn hỏi mẫu nào ạ?"}
	b.route(t)
	rows := b.proposals(t, detail.Conversation.Turn.ID)
	if len(rows) != 2 || rows[0].Status != types.ProposalSuperseded || rows[1].Status != types.ProposalPending {
		t.Fatal("newer message must supersede the older proposal", rows)
	}
}

func TestRouterBranchSelectionWithFakeAI(t *testing.T) {
	cases := []struct {
		name                        string
		run                         types.AIRun
		err                         error
		status, route, source, text string
		handoff                     string
	}{
		{"ai answer", types.AIRun{Route: "generate", Text: "Áo này giá 200k"}, nil, types.ProposalPending, types.RouteAnswer, types.SourceAI, "Áo này giá 200k", ""},
		{"clarification", types.AIRun{Route: "clarify", Text: "Bạn hỏi mẫu nào ạ?"}, nil, types.ProposalPending, types.RouteClarify, types.SourceAI, "Bạn hỏi mẫu nào ạ?", ""},
		{"handoff", types.AIRun{Route: "human", Reason: "human_review_required"}, nil, types.ProposalHandedOff, types.RouteHandoff, "", "", "insufficient_information"},
		{"ai failure", types.AIRun{Status: types.AIRunFailed}, nil, types.ProposalHandedOff, types.RouteHandoff, "", "", "ai_error"},
		{"ai not configured", types.AIRun{}, types.ErrAIUnavailable, types.ProposalHandedOff, types.RouteHandoff, "", "", "ai_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newBotFixture(t, types.ReplyModeSuggest)
			b.activeScript(t, "phí ship", "Phí ship 30k")
			b.ai.run, b.ai.err = tc.run, tc.err
			message := b.customer(t, "c1", "áo này giá bao nhiêu")
			b.route(t)
			if len(b.ai.started) != 1 {
				t.Fatal("no script matched, so the AI must be asked once")
			}
			req := b.ai.started[0]
			if req.TenantID != b.tenant || req.Message != "áo này giá bao nhiêu" || len(req.Scripts) != 1 || req.Scripts[0].Trigger != "phí ship" {
				t.Fatalf("unexpected AI request %+v", req)
			}
			detail := b.detail(t, message.ThreadID)
			turn := detail.Conversation.Turn
			rows := b.proposals(t, turn.ID)
			p := rows[len(rows)-1]
			if p.Status != tc.status || deref(p.Route) != tc.route || deref(p.Source) != tc.source || deref(p.Text) != tc.text {
				t.Fatalf("unexpected proposal %+v", p)
			}
			if tc.source == types.SourceAI && deref(p.AIRunID) != p.ID {
				t.Fatal("an AI proposal must record its run ID as source")
			}
			if tc.handoff != "" {
				if turn.HandlingState() != "WAITING_STAFF" || turn.NeedsHumanReason != tc.handoff || turn.ControlGeneration != 2 {
					t.Fatalf("handoff must use the staff handoff transition: %+v", turn)
				}
				if auditCount(t, b.db, b.tenant, turn.ID, "messaging.handoff") != 1 {
					t.Fatal("router handoff must be audited")
				}
			} else if turn.HandlingState() != "BOT_ACTIVE" {
				t.Fatal("a proposal must not change control", turn)
			}
			if len(outbound(detail)) != 0 {
				t.Fatal("suggest mode sent a message")
			}
		})
	}
}

func TestRouterPollsAcceptedAIRunWithoutBlocking(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	b.ai.pending = 1
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	if b.route(t) != 0 {
		t.Fatal("a running AI run must not be decided")
	}
	turn := b.detail(t, message.ThreadID).Conversation.Turn
	rows := b.proposals(t, turn.ID)
	if rows[0].Status != types.ProposalRouting || rows[0].Attempts != 1 || !rows[0].AvailableAt.After(time.Now()) {
		t.Fatal("the job must wait for the next poll", rows[0])
	}
	b.router.now = func() time.Time { return time.Now().UTC().Add(time.Minute) }
	if b.route(t) != 1 || b.ai.polled != 1 || len(b.ai.started) != 1 {
		t.Fatal("the second run must poll the same run, not start another")
	}
	if rows = b.proposals(t, turn.ID); rows[0].Status != types.ProposalPending {
		t.Fatal(rows[0])
	}
}

func TestRouterRestartsAIRunThatNeverReachedTheService(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	// The first StartRun fails in transport before the AI service stores the run.
	b.ai.startErrs = []error{errors.New("AI run request failed")}
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	if b.route(t) != 0 {
		t.Fatal("a transient StartRun failure must wait for a retry, not decide")
	}
	turn := b.detail(t, message.ThreadID).Conversation.Turn
	if rows := b.proposals(t, turn.ID); rows[0].Status != types.ProposalRouting || rows[0].Attempts != 1 {
		t.Fatal("the job must stay routing for the next attempt", rows[0])
	}
	// The next attempt finds no run (NOT_FOUND) and starts it again with the same run ID.
	b.router.now = func() time.Time { return time.Now().UTC().Add(time.Minute) }
	if b.route(t) != 1 || b.ai.polled != 1 || len(b.ai.started) != 2 {
		t.Fatalf("a missing run must be started again: polled=%d started=%d", b.ai.polled, len(b.ai.started))
	}
	rows := b.proposals(t, turn.ID)
	if b.ai.started[0].RunID != rows[0].ID || b.ai.started[1].RunID != rows[0].ID {
		t.Fatal("a restarted run must keep the job's run ID")
	}
	if rows[0].Status != types.ProposalPending || deref(rows[0].Route) != types.RouteClarify {
		t.Fatal("the restarted run's candidate must be proposed", rows[0])
	}
	if after := b.detail(t, message.ThreadID).Conversation.Turn; after.HandlingState() != "BOT_ACTIVE" {
		t.Fatal("a recovered run must not hand the turn off", after)
	}
}

func TestRouterHandsOffWhenRunIDHoldsDifferentInput(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeAutoSend)
	b.ai.err = types.ErrAIRunConflict
	message := b.customer(t, "c1", "alo")
	if b.route(t) != 1 || b.ai.polled != 0 {
		t.Fatal("a run conflict must be decided without reading the conflicting run")
	}
	detail := b.detail(t, message.ThreadID)
	rows := b.proposals(t, detail.Conversation.Turn.ID)
	if rows[0].Status != types.ProposalHandedOff || deref(rows[0].Reason) != "ai_run_conflict" || deref(rows[0].Text) != "" {
		t.Fatal("a run conflict must hand off with a sanitized reason", rows[0])
	}
	if turn := detail.Conversation.Turn; turn.HandlingState() != "WAITING_STAFF" || turn.NeedsHumanReason != "ai_error" || len(outbound(detail)) != 0 {
		t.Fatal("a run conflict must move the turn to staff without sending", turn)
	}
}

func TestAutoSendSendsOnlyWhileEligible(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeAutoSend)
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")
	ctx := context.Background()

	// Eligible: BOT_ACTIVE at the same generation and revision -> the bot sends the script answer.
	message := b.customer(t, "c1", "giá áo")
	b.route(t)
	detail := b.detail(t, message.ThreadID)
	sent := outbound(detail)
	if len(sent) != 1 || sent[0].ActorID != types.BotActorID || sent[0].Body != "Dạ áo giá 200k ạ" || sent[0].Status != "simulated" {
		t.Fatalf("auto_send must send an eligible reply: %+v", sent)
	}
	rows := b.proposals(t, detail.Conversation.Turn.ID)
	if rows[0].Status != types.ProposalAutoSent || deref(rows[0].SentMessageID) != sent[0].ID {
		t.Fatal(rows[0])
	}

	// Not eligible: a sale takes over before the router runs -> the job is fenced and nothing is sent.
	other := b.customer(t, "c2", "giá áo")
	turn := b.detail(t, other.ThreadID).Conversation.Turn
	if _, err := b.messaging.Command(ctx, b.tenant, b.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err != nil {
		t.Fatal(err)
	}
	b.route(t)
	if len(outbound(b.detail(t, other.ThreadID))) != 0 || b.proposals(t, turn.ID)[0].Status != types.ProposalSuperseded {
		t.Fatal("a takeover must stop the pending automatic send")
	}

	// Not eligible: a person-requested proposal is never sent automatically, even in auto_send mode.
	third := b.customer(t, "c3", "xin chào")
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	b.route(t) // auto-sends the clarification
	turn = b.detail(t, third.ThreadID).Conversation.Turn
	if _, err := b.messaging.RegenerateProposal(ctx, b.tenant, b.peer, turn.ID); err != nil {
		t.Fatal(err)
	}
	b.route(t)
	rows = b.proposals(t, turn.ID)
	if last := rows[len(rows)-1]; last.Status != types.ProposalPending || deref(last.RequestedBy) != b.peer || len(outbound(b.detail(t, third.ThreadID))) != 1 {
		t.Fatal("a regenerated proposal must wait for a sale", rows)
	}
}

func TestAutoSendChannelIntentIsCancelledByTakeoverUntilDispatched(t *testing.T) {
	f, channels, _ := channelServiceFixture(t)
	ctx := context.Background()
	repo := repository.NewMessagingRepository(f.db)
	settings := NewReplySettingsService(repo)
	if _, err := settings.Update(ctx, f.tenant, f.owner, types.ReplySettingsUpdate{Mode: types.ReplyModeAutoSend}); err != nil {
		t.Fatal(err)
	}
	scripts := NewScriptService(repository.NewScriptRepository(f.db), repository.NewTenantDBRepository(f.db))
	b := botFixture{salesFixture: f, messaging: NewMessagingService(repo), scripts: scripts, ai: &fakeAI{}}
	b.router = NewReplyRouterService(repo, scripts, b.ai)
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")
	connected, _ := connectPages(t, channels, f.tenant, f.owner, "page-a")
	conn := connected[0]
	in := NewIngressService(repository.NewIngressRepository(f.db))
	receive := func(key, id string) types.MessagingTurn {
		e := messaging.MessageReceived{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: f.tenant, OccurredAt: time.Now()}, Platform: "facebook", ConnectionID: conn.ID, ConnectionVersion: conn.Version, AccountID: "page-a", ConversationKey: key, MessageID: id, SourceIdentity: "message:" + id, Sender: key, Recipient: "page-a", Direction: "inbound", Text: "giá áo", SourceAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
		if err := in.Consume(ctx, e); err != nil {
			t.Fatal(err)
		}
		b.route(t)
		var row messagingmodel.Thread
		if err := f.db.Where("tenant_id = ? AND recipient_id = ?", f.tenant, key).First(&row).Error; err != nil {
			t.Fatal(err)
		}
		return *b.detail(t, row.ID).Conversation.Turn
	}
	intentOf := func(turn types.MessagingTurn) (messagingmodel.OutboundIntent, messagingmodel.Message) {
		var msg messagingmodel.Message
		if err := f.db.Where("tenant_id = ? AND turn_id = ? AND actor_id = ?", f.tenant, turn.ID, types.BotActorID).First(&msg).Error; err != nil {
			t.Fatal("auto_send must create a bot message", err)
		}
		var intent messagingmodel.OutboundIntent
		if err := f.db.Where("tenant_id = ? AND message_id = ?", f.tenant, msg.ID).First(&intent).Error; err != nil {
			t.Fatal(err)
		}
		return intent, msg
	}

	// Undispatched: the outbox event is still pending, so the takeover cancels the send.
	turn := receive("psid-1", "m1")
	intent, _ := intentOf(turn)
	if intent.Status != types.IntentPending {
		t.Fatal("auto_send must enqueue an outbound intent", intent)
	}
	if _, err := b.messaging.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err != nil {
		t.Fatal(err)
	}
	intent, msg := intentOf(turn)
	var event model.DomainEvent
	f.db.Where("aggregate_id = ?", intent.ID).First(&event)
	if intent.Status != types.IntentCancelled || msg.DeliveryStatus != types.DeliveryFailed || deref(msg.DeliveryErrorReason) != types.DeliveryCancelledReason || event.State != "failed" {
		t.Fatalf("takeover must cancel the undispatched bot send: %+v %+v %s", intent, msg, event.State)
	}

	// Dispatched: once the relay published the event, the send is reconciled from the connector, not cancelled.
	turn = receive("psid-2", "m2")
	intent, _ = intentOf(turn)
	if err := f.db.Model(&model.DomainEvent{}).Where("aggregate_id = ?", intent.ID).Update("state", "published").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := b.messaging.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version}); err != nil {
		t.Fatal(err)
	}
	if intent, msg = intentOf(turn); intent.Status != types.IntentPending || msg.DeliveryStatus != types.DeliverySending {
		t.Fatal("a dispatched send must stay pending for reconciliation", intent, msg)
	}
}

func TestProposalSendAuthorityAndStaleGeneration(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")

	// BOT_ACTIVE: the sender claims the unassigned turn atomically with the send.
	message := b.customer(t, "c1", "giá áo")
	b.route(t)
	detail := b.detail(t, message.ThreadID)
	turn, p := detail.Conversation.Turn, detail.Proposals[0].Proposal
	_, err := b.messaging.SendProposal(ctx, b.other, b.sale, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
	requireSalesStatus(t, err, 403)
	key := uuid.NewString()
	sent, err := b.messaging.SendProposal(ctx, b.tenant, b.sale, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: key})
	if err != nil || sent.ActorID != b.sale || sent.Body != "Dạ áo giá 200k ạ" {
		t.Fatal(sent, err)
	}
	replay, err := b.messaging.SendProposal(ctx, b.tenant, b.sale, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: key})
	if err != nil || replay.ID != sent.ID {
		t.Fatal("retrying the same client_message_id must replay", err)
	}
	after := b.detail(t, message.ThreadID)
	if claimed := after.Conversation.Turn; claimed.HandlingState() != "STAFF_ACTIVE" || *claimed.AssigneeID != b.sale || claimed.ControlGeneration != turn.ControlGeneration+1 {
		t.Fatal("sending from bot handling must claim the turn", claimed)
	}
	if len(after.Proposals) != 0 || len(outbound(after)) != 1 {
		t.Fatal("the sent proposal is no longer live", after.Proposals)
	}
	_, err = b.messaging.SendProposal(ctx, b.tenant, b.sale, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
	requireSalesStatus(t, err, 409)

	// A control change makes a pending proposal stale: it cannot be sent by anyone.
	second := b.customer(t, "c2", "giá áo")
	b.route(t)
	detail = b.detail(t, second.ThreadID)
	turn, p = detail.Conversation.Turn, detail.Proposals[0].Proposal
	assignee := b.peer
	assigned, err := b.messaging.Command(ctx, b.tenant, b.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: turn.Version, AssigneeID: &assignee})
	if err != nil {
		t.Fatal(err)
	}
	detail = b.detail(t, second.ThreadID)
	if len(detail.Proposals) != 1 || !detail.Proposals[0].Stale {
		t.Fatal("the Desk must show the proposal as stale", detail.Proposals)
	}
	for _, actor := range []string{b.peer, b.sale} {
		_, err = b.messaging.SendProposal(ctx, b.tenant, actor, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
		requireSalesStatus(t, err, 409)
	}
	// Only the assigned sale may request a fresh proposal on a staffed turn; it is never sent automatically.
	_, err = b.messaging.RegenerateProposal(ctx, b.tenant, b.sale, turn.ID)
	requireSalesStatus(t, err, 403)
	if _, err = b.messaging.RegenerateProposal(ctx, b.tenant, b.peer, turn.ID); err != nil {
		t.Fatal(err)
	}
	b.route(t)
	detail = b.detail(t, second.ThreadID)
	if len(detail.Proposals) != 1 || detail.Proposals[0].Stale || detail.Proposals[0].Proposal.Status != types.ProposalPending || detail.Proposals[0].Proposal.ControlGeneration != assigned.ControlGeneration {
		t.Fatal("regenerate must propose at the current generation", detail.Proposals)
	}
	fresh := detail.Proposals[0].Proposal
	// The non-assigned sale is refused; discard by a non-assignee is refused too.
	_, err = b.messaging.SendProposal(ctx, b.tenant, b.sale, turn.ID, fresh.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
	requireSalesStatus(t, err, 403)
	_, err = b.messaging.DiscardProposal(ctx, b.tenant, b.sale, turn.ID, fresh.ID)
	requireSalesStatus(t, err, 403)
	edited, err := b.messaging.SendProposal(ctx, b.tenant, b.peer, turn.ID, fresh.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString(), Text: "Dạ áo giá 180k ạ"})
	if err != nil || edited.Body != "Dạ áo giá 180k ạ" || edited.ActorID != b.peer {
		t.Fatal("the assigned sale may edit then send", edited, err)
	}
	if auditCount(t, b.db, b.tenant, turn.ID, "messaging.proposal_sent") != 1 {
		t.Fatal("proposal send must be audited")
	}
}

func TestProposalDiscardSendsNothing(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	b.route(t)
	detail := b.detail(t, message.ThreadID)
	turn, p := detail.Conversation.Turn, detail.Proposals[0].Proposal
	discarded, err := b.messaging.DiscardProposal(ctx, b.tenant, b.peer, turn.ID, p.ID)
	if err != nil || discarded.Status != types.ProposalDiscarded || discarded.DecidedBy != b.peer {
		t.Fatal(discarded, err)
	}
	_, err = b.messaging.SendProposal(ctx, b.tenant, b.sale, turn.ID, p.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()})
	requireSalesStatus(t, err, 409)
	after := b.detail(t, message.ThreadID)
	if len(outbound(after)) != 0 || len(after.Proposals) != 0 || after.Conversation.Turn.HandlingState() != "BOT_ACTIVE" {
		t.Fatal("discard must send nothing and keep control", after)
	}
}
