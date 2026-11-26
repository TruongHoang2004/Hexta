package service

import (
	"context"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/google/uuid"
)

// Owner decisions D1–D3 (2026-11-26): reply mode off by default, auto_send never sends AI-written answers, and a
// bot proposal waiting for review keeps its turn in the waiting queue and out of expiry.

// botFixtureFrom builds the bot services on an existing fixture without saving a reply mode.
func botFixtureFrom(t *testing.T, f salesFixture) botFixture {
	t.Helper()
	repo := repository.NewMessagingRepository(f.db)
	b := botFixture{salesFixture: f, messaging: NewMessagingService(repo), settings: NewReplySettingsService(repo),
		scripts: NewScriptService(repository.NewScriptRepository(f.db), repository.NewTenantDBRepository(f.db)), ai: &fakeAI{}}
	b.router = NewReplyRouterService(repo, b.scripts, b.ai)
	return b
}

// receiveChannelText consumes one live inbound text from a connected page, as the connector would publish it.
func receiveChannelText(t *testing.T, f salesFixture, conn types.ChannelConnection, psid, id, text string) {
	t.Helper()
	e := messaging.MessageReceived{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: f.tenant, OccurredAt: time.Now()}, Platform: "facebook", ConnectionID: conn.ID, ConnectionVersion: conn.Version, AccountID: conn.AccountID, ConversationKey: psid, MessageID: id, SourceIdentity: "message:" + id, Sender: psid, Recipient: conn.AccountID, Direction: "inbound", Text: text, SourceAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	if err := NewIngressService(repository.NewIngressRepository(f.db)).Consume(context.Background(), e); err != nil {
		t.Fatal(err)
	}
}

func (b botFixture) setMode(t *testing.T, mode string) {
	t.Helper()
	ctx := context.Background()
	current, err := b.settings.Get(ctx, b.tenant, b.owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.settings.Update(ctx, b.tenant, b.owner, types.ReplySettingsUpdate{Mode: mode, ExpectedVersion: current.Version}); err != nil {
		t.Fatal(err)
	}
}

func (b botFixture) turnOf(t *testing.T, thread string) types.MessagingTurn {
	t.Helper()
	return *b.detail(t, thread).Conversation.Turn
}

// waiting rebuilds the read model (standing in for the projector) and returns the Desk waiting count and the
// threads listed under the waiting filter.
func (b botFixture) waiting(t *testing.T) (int64, map[string]types.MessagingTurn) {
	t.Helper()
	ctx := context.Background()
	page, err := b.messaging.Conversations(ctx, b.tenant, b.sale, types.MessagingQuery{Limit: 50, Filter: "waiting"})
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]types.MessagingTurn{}
	for _, item := range page.Items {
		listed[item.Thread.ID] = *item.Turn
	}
	summary, err := b.messaging.Summary(ctx, b.tenant, b.sale)
	if err != nil {
		t.Fatal(err)
	}
	return summary.WaitingStaff, listed
}

func TestBotOffIsTheDefaultAndNeverRoutes(t *testing.T) {
	b := newBotFixture(t, "")
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")
	message := b.customer(t, "c1", "giá áo")
	turn := b.turnOf(t, message.ThreadID)
	if turn.HandlingState() != "WAITING_STAFF" || !turn.NeedsHuman || turn.NeedsHumanReason != "" || turn.ReviewPending {
		t.Fatalf("bot off must start the turn waiting for staff exactly as before the bot: %+v", turn)
	}
	if len(b.proposals(t, turn.ID)) != 0 || b.route(t) != 0 || len(b.ai.started) != 0 {
		t.Fatal("bot off must create no routing job and call no AI")
	}
	if count, listed := b.waiting(t); count != 1 || len(listed) != 1 {
		t.Fatal("the new turn waits for staff", count, listed)
	}
	// A regenerate request is refused while the bot is off, even for the assigned sale.
	claimed, err := b.messaging.Command(context.Background(), b.tenant, b.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.messaging.RegenerateProposal(context.Background(), b.tenant, b.sale, claimed.ID)
	requireSalesStatus(t, err, 409)
}

func TestReturnToBotIsRefusedWhileBotIsOff(t *testing.T) {
	b := newBotFixture(t, "")
	ctx := context.Background()
	message := b.customer(t, "c1", "xin chào")
	turn, err := b.messaging.Command(ctx, b.tenant, b.sale, b.turnOf(t, message.ThreadID).ID, "claim", types.MessagingCommand{ExpectedVersion: b.turnOf(t, message.ThreadID).Version})
	if err != nil {
		t.Fatal(err)
	}
	_, err = b.messaging.Command(ctx, b.tenant, b.sale, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: turn.Version})
	requireSalesStatus(t, err, 409)
	if after := b.turnOf(t, message.ThreadID); after.Version != turn.Version || after.HandlingState() != "STAFF_ACTIVE" {
		t.Fatal("a refused return must not change the turn", after)
	}
	b.setMode(t, types.ReplyModeSuggest)
	returned, err := b.messaging.Command(ctx, b.tenant, b.sale, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || returned.HandlingState() != "BOT_ACTIVE" {
		t.Fatal("with the bot on, the assigned sale may return the turn", returned, err)
	}
	claimed, err := b.messaging.Command(ctx, b.tenant, b.owner, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: returned.Version})
	if err != nil {
		t.Fatal(err)
	}
	b.setMode(t, types.ReplyModeOff)
	_, err = b.messaging.Command(ctx, b.tenant, b.owner, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: claimed.Version})
	requireSalesStatus(t, err, 409)
}

func TestSwitchingBotOffDropsQueuedJobsAndHandsOffTheNextMessage(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}

	// proposed: routed, its proposal waits for review. quiet: routed, proposal discarded, no live proposal.
	proposed := b.customer(t, "proposed", "alo")
	quiet := b.customer(t, "quiet", "alo")
	b.route(t)
	quietTurn := b.turnOf(t, quiet.ThreadID)
	if _, err := b.messaging.DiscardProposal(ctx, b.tenant, b.peer, quietTurn.ID, b.detail(t, quiet.ThreadID).Proposals[0].Proposal.ID); err != nil {
		t.Fatal(err)
	}
	// queued: its routing job has not been applied yet when the owner turns the bot off.
	queued := b.customer(t, "queued", "alo")
	started := len(b.ai.started)
	b.setMode(t, types.ReplyModeOff)

	// Existing turns are not moved by the switch itself.
	for _, thread := range []string{proposed.ThreadID, quiet.ThreadID, queued.ThreadID} {
		if turn := b.turnOf(t, thread); turn.HandlingState() != "BOT_ACTIVE" {
			t.Fatal("switching the bot off must not move existing turns", turn)
		}
	}

	// The queued job is dropped without an AI call; its unanswered message is handed to staff.
	if b.route(t) != 1 || len(b.ai.started) != started {
		t.Fatal("a queued job must be dropped without calling the AI")
	}
	queuedTurn := b.turnOf(t, queued.ThreadID)
	if queuedTurn.HandlingState() != "WAITING_STAFF" || queuedTurn.NeedsHumanReason != types.HandoffReasonBotOff {
		t.Fatalf("a dropped job must hand its turn to staff with reason bot_off: %+v", queuedTurn)
	}
	if rows := b.proposals(t, queuedTurn.ID); len(rows) != 1 || rows[0].Status != types.ProposalHandedOff || deref(rows[0].Reason) != types.HandoffReasonBotOff || deref(rows[0].Text) != "" {
		t.Fatal("the dropped job must carry no candidate", rows)
	}
	if auditCount(t, b.db, b.tenant, queuedTurn.ID, "messaging.handoff") != 1 {
		t.Fatal("the bot_off handoff must be audited")
	}

	// The turn whose proposal waits for review keeps it and stays in the waiting queue; staff may still send it.
	proposedTurn := b.turnOf(t, proposed.ThreadID)
	if !proposedTurn.ReviewPending || len(b.detail(t, proposed.ThreadID).Proposals) != 1 {
		t.Fatal("a pending proposal survives the switch", proposedTurn)
	}
	_, err := b.messaging.RegenerateProposal(ctx, b.tenant, b.peer, proposedTurn.ID)
	requireSalesStatus(t, err, 409)

	// The next customer message on a bot-handled turn hands it off; nothing is routed.
	for _, thread := range []string{"quiet", "proposed"} {
		message := b.customer(t, thread, "còn đó không")
		turn := b.turnOf(t, message.ThreadID)
		if turn.HandlingState() != "WAITING_STAFF" || turn.NeedsHumanReason != types.HandoffReasonBotOff || turn.ReviewPending {
			t.Fatalf("%s: the next message must hand the turn to staff with reason bot_off: %+v", thread, turn)
		}
		for _, row := range b.proposals(t, turn.ID) {
			if row.Status == types.ProposalRouting {
				t.Fatalf("%s: no routing job may be created while the bot is off", thread)
			}
		}
		if auditCount(t, b.db, b.tenant, turn.ID, "messaging.handoff") != 1 {
			t.Fatalf("%s: the bot_off handoff must be audited", thread)
		}
	}
	if detail := b.detail(t, proposed.ThreadID); len(detail.Proposals) != 1 || !detail.Proposals[0].Stale {
		t.Fatal("the old proposal is fenced by the handoff and shows as stale", detail.Proposals)
	}
	if b.route(t) != 0 || len(b.ai.started) != started {
		t.Fatal("bot off must not call the AI")
	}
	if count, listed := b.waiting(t); count != 3 || len(listed) != 3 {
		t.Fatal("all three conversations now wait for staff", count, listed)
	}
}

func TestBotOffHandsOffChannelMessageOnBotTurn(t *testing.T) {
	f, channels, _ := channelServiceFixture(t)
	b := botFixtureFrom(t, f)
	b.setMode(t, types.ReplyModeSuggest)
	connected, _ := connectPages(t, channels, f.tenant, f.owner, "page-a")
	receive := func(id string) types.MessagingTurn {
		t.Helper()
		receiveChannelText(t, f, connected[0], "psid-1", id, "alo")
		var row messagingmodel.Thread
		if err := f.db.Where("tenant_id = ? AND recipient_id = ?", f.tenant, "psid-1").First(&row).Error; err != nil {
			t.Fatal(err)
		}
		return b.turnOf(t, row.ID)
	}
	turn := receive("m1")
	if turn.HandlingState() != "BOT_ACTIVE" || len(b.proposals(t, turn.ID)) != 1 {
		t.Fatal("with suggest, a channel message starts a bot turn and is routed", turn)
	}
	b.setMode(t, types.ReplyModeOff)
	// The queued job of m1 is still unapplied; m2 hands the turn off at ingress and supersedes nothing new.
	turn = receive("m2")
	if turn.HandlingState() != "WAITING_STAFF" || turn.NeedsHumanReason != types.HandoffReasonBotOff || len(b.proposals(t, turn.ID)) != 1 {
		t.Fatalf("ingress must hand a bot turn to staff while the bot is off: %+v", turn)
	}
	// The old job is now fenced by the handoff: applying it changes nothing.
	b.route(t)
	if after := b.turnOf(t, turn.ThreadID); after.Version != turn.Version || b.proposals(t, turn.ID)[0].Status != types.ProposalSuperseded {
		t.Fatal("a fenced job must be superseded without touching the turn", after)
	}
}

func TestAutoSendNeverSendsAIWrittenAnswers(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeAutoSend)
	b.activeScript(t, "giá áo", "Dạ áo giá 200k ạ")
	shipping := b.activeScript(t, "phí ship", "Phí ship 30k ạ")
	cases := []struct {
		name, body string
		run        types.AIRun
		sent       string // expected bot message body; empty means nothing is sent
		status     string
	}{
		{"script hit", "giá áo", types.AIRun{}, "Dạ áo giá 200k ạ", types.ProposalAutoSent},
		{"ai selected script", "ship bao nhiêu", types.AIRun{Route: "script", ScriptID: shipping.ID}, "Phí ship 30k ạ", types.ProposalAutoSent},
		{"ai clarification", "alo", types.AIRun{Route: "clarify", Text: "Bạn hỏi mẫu nào ạ?"}, "Bạn hỏi mẫu nào ạ?", types.ProposalAutoSent},
		{"ai written answer", "áo len còn không", types.AIRun{Route: "generate", Text: "Dạ áo len còn ạ"}, "", types.ProposalPending},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b.ai.run = tc.run
			message := b.customer(t, uuid.NewString(), tc.body)
			b.route(t)
			detail := b.detail(t, message.ThreadID)
			sent := outbound(detail)
			rows := b.proposals(t, detail.Conversation.Turn.ID)
			if rows[0].Status != tc.status {
				t.Fatalf("case %d: unexpected proposal %+v", i, rows[0])
			}
			turn := detail.Conversation.Turn
			if tc.sent == "" {
				if len(sent) != 0 || !turn.ReviewPending || turn.HandlingState() != "BOT_ACTIVE" {
					t.Fatalf("an AI-written answer must wait for staff review: sent=%v turn=%+v", sent, turn)
				}
				return
			}
			if len(sent) != 1 || sent[0].ActorID != types.BotActorID || sent[0].Body != tc.sent || turn.ReviewPending {
				t.Fatalf("expected an automatic send of %q: %+v", tc.sent, sent)
			}
		})
	}
}

func TestPendingProposalWaitsForStaffAndBlocksExpiry(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	if count, _ := b.waiting(t); count != 0 {
		t.Fatal("a turn still being routed is not waiting for staff")
	}
	b.route(t)
	turn := b.turnOf(t, message.ThreadID)
	if turn.HandlingState() != "BOT_ACTIVE" || !turn.ReviewPending {
		t.Fatalf("a pending proposal on a bot turn must mark it for review: %+v", turn)
	}
	count, listed := b.waiting(t)
	if listed[message.ThreadID].ID != turn.ID || !listed[message.ThreadID].ReviewPending || count != 1 {
		t.Fatal("the turn must be counted and listed as waiting for staff", count, listed)
	}

	// The pending proposal blocks the 24-hour expiry.
	backdate(t, b.db, b.tenant, turn.ID, 25*time.Hour)
	if expireOnce(t, b.db) != 0 || b.turnOf(t, message.ThreadID).Lifecycle != types.TurnOpen {
		t.Fatal("a turn whose proposal waits for review must not expire")
	}

	// A newer customer message supersedes the proposal: the turn leaves the queue until the new proposal is ready.
	b.customer(t, "c1", "alo alo")
	if turn = b.turnOf(t, message.ThreadID); turn.ReviewPending {
		t.Fatal("a superseded proposal no longer holds the turn")
	}
	b.route(t)
	turn = b.turnOf(t, message.ThreadID)
	p := b.detail(t, message.ThreadID).Proposals[0].Proposal
	if !turn.ReviewPending || p.Status != types.ProposalPending {
		t.Fatal("the new pending proposal marks the turn again", turn, p)
	}

	// Discarding it resumes the normal rules: not waiting, and the inactive turn expires.
	if _, err := b.messaging.DiscardProposal(ctx, b.tenant, b.peer, turn.ID, p.ID); err != nil {
		t.Fatal(err)
	}
	turn = b.turnOf(t, message.ThreadID)
	if turn.ReviewPending || turn.HandlingState() != "BOT_ACTIVE" {
		t.Fatal("discard must clear the review mark and keep control", turn)
	}
	if count, listed := b.waiting(t); count != 0 || len(listed) != 0 {
		t.Fatal("a discarded proposal no longer counts as waiting", count, listed)
	}
	// The read model follows the conversation events, including out-of-order replays.
	projectAllEvents(t, b, true)
	if view := conversationView(t, b, message.ThreadID); view.ReviewPending || view.TurnVersion != turn.Version {
		t.Fatal("replayed events must leave the projected review mark at the latest state", view)
	}
	backdate(t, b.db, b.tenant, turn.ID, 25*time.Hour)
	if expireOnce(t, b.db) != 1 || b.turnOf(t, message.ThreadID).Lifecycle != types.TurnClosed {
		t.Fatal("after the discard the inactive turn must expire")
	}
}

func TestReviewMarkIsClearedBySendAndProjectedFromEvents(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	ctx := context.Background()
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	b.route(t)
	pending := b.turnOf(t, message.ThreadID)
	projectAllEvents(t, b, false)
	if view := conversationView(t, b, message.ThreadID); !view.ReviewPending || view.TurnVersion != pending.Version {
		t.Fatal("the projected view must carry the review mark", view)
	}
	p := b.detail(t, message.ThreadID).Proposals[0].Proposal
	if _, err := b.messaging.SendProposal(ctx, b.tenant, b.sale, pending.ID, p.ID, types.ProposalSendInput{ClientMessageID: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	sent := b.turnOf(t, message.ThreadID)
	if sent.ReviewPending || sent.HandlingState() != "STAFF_ACTIVE" {
		t.Fatal("sending the proposal claims the turn and clears the review mark", sent)
	}
	// Replaying every snapshot newest-first, then oldest-first, must still end at the latest state.
	projectAllEvents(t, b, true)
	projectAllEvents(t, b, false)
	if view := conversationView(t, b, message.ThreadID); view.ReviewPending || view.TurnVersion != sent.Version {
		t.Fatal("an older snapshot with the review mark must not overwrite the newer state", view)
	}
}

// projectAllEvents applies the tenant's conversation events through the projector, oldest first or newest first.
func projectAllEvents(t *testing.T, b botFixture, newestFirst bool) {
	t.Helper()
	ctx := context.Background()
	views := b.messaging.views
	projector := NewConversationProjector(views)
	all := conversationEvents(t, b.db, b.tenant)
	for i := range all {
		e := all[i]
		if newestFirst {
			e = all[len(all)-1-i]
		}
		tenant, change, ok := projector.Decode(envelope(t, e))
		if !ok || change == nil {
			t.Fatal("undecodable event")
		}
		if err := projector.Project(ctx, tenant, *change); err != nil {
			t.Fatal(err)
		}
	}
}

func conversationView(t *testing.T, b botFixture, thread string) messagingmodel.ConversationView {
	t.Helper()
	var view messagingmodel.ConversationView
	if err := b.db.Where("tenant_id = ? AND thread_id = ?", b.tenant, thread).First(&view).Error; err != nil {
		t.Fatal(err)
	}
	return view
}
