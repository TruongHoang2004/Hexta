package service

import (
	"context"
	"reflect"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

func TestRouterKnowledgeAnswerStaysPendingWithCitations(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeAutoSend)
	ctx := context.Background()
	product, err := b.service.SaveProduct(ctx, b.tenant, b.owner, "", types.ProductInput{Name: "Áo thun basic", SKU: "AO-1", VariantName: "M"})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := NewKnowledgeService(repository.NewKnowledgeRepository(b.db)).Create(ctx, b.tenant, b.owner, types.KnowledgeDocumentInput{Title: "Đổi trả", Body: "7 ngày"})
	if err != nil {
		t.Fatal(err)
	}
	cited := []string{"product:" + product.ID, "document:" + doc.ID, "fact:f1", "product:" + product.ID, "order:x", "script:bad id", "document:" + uuid.NewString()}
	b.ai.run = types.AIRun{Route: "generate", CandidateKind: "generated", RequiresReview: true, Text: "Dạ áo thun basic còn hàng ạ", CitedSourceIDs: cited}
	message := b.customer(t, "c1", "áo thun còn không")
	b.route(t)
	if len(b.ai.started) != 1 || !b.ai.started[0].AllowKnowledgeAnswer {
		t.Fatalf("the router must opt into knowledge answers: %+v", b.ai.started)
	}
	detail := b.detail(t, message.ThreadID)
	if len(outbound(detail)) != 0 || !detail.Conversation.Turn.ReviewPending {
		t.Fatal("a knowledge answer must never be sent automatically, even in auto_send")
	}
	rows := b.proposals(t, detail.Conversation.Turn.ID)
	if rows[0].Status != types.ProposalPending || deref(rows[0].Route) != types.RouteAnswer || deref(rows[0].Source) != types.SourceAI {
		t.Fatalf("unexpected proposal %+v", rows[0])
	}
	if len(detail.Proposals) != 1 {
		t.Fatalf("expected the live proposal on the Desk: %+v", detail.Proposals)
	}
	want := []types.ProposalCitation{
		{SourceID: "product:" + product.ID, Kind: "product", Label: "Áo thun basic"},
		{SourceID: "document:" + doc.ID, Kind: "document", Label: "Đổi trả"},
		{SourceID: "fact:f1", Kind: "fact"},
		{SourceID: cited[6], Kind: "document"},
	}
	if got := detail.Proposals[0].Citations; !reflect.DeepEqual(got, want) {
		t.Fatalf("citations must be validated, deduplicated and labelled:\n got %+v\nwant %+v", got, want)
	}
}

func TestAutoSendHoldsModelWrittenClarifications(t *testing.T) {
	cases := []struct {
		name string
		run  types.AIRun
	}{
		{"requires review", types.AIRun{Route: "clarify", Text: "Anh/chị hỏi mẫu nào ạ?", RequiresReview: true}},
		{"generated kind", types.AIRun{Route: "clarify", Text: "Anh/chị hỏi mẫu nào ạ?", CandidateKind: "generated"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newBotFixture(t, types.ReplyModeAutoSend)
			b.ai.run = tc.run
			message := b.customer(t, "c1", "alo")
			b.route(t)
			detail := b.detail(t, message.ThreadID)
			rows := b.proposals(t, detail.Conversation.Turn.ID)
			if len(outbound(detail)) != 0 || rows[0].Status != types.ProposalPending || deref(rows[0].Route) != types.RouteClarify {
				t.Fatalf("a model-written clarification must wait for review: %+v", rows[0])
			}
		})
	}
}

func TestRouterUsesOnlyKnownAIHandoffReasons(t *testing.T) {
	for reason, want := range map[string]string{
		"customer_requested_staff": "customer_requested_staff",
		"business_review":          "business_review",
		"bot_off":                  "insufficient_information",
		"":                         "insufficient_information",
	} {
		b := newBotFixture(t, types.ReplyModeSuggest)
		b.ai.run = types.AIRun{Route: "human", HandoffReason: reason, NeedsHuman: true, Reason: "jev_customer_requested_human"}
		message := b.customer(t, "c1", "cho gặp nhân viên")
		b.route(t)
		if turn := b.turnOf(t, message.ThreadID); turn.HandlingState() != "WAITING_STAFF" || turn.NeedsHumanReason != want {
			t.Fatalf("AI reason %q: want %q, got %+v", reason, want, turn)
		}
	}
}

// offAtRead reports the bot off to the router's first read (outside the decision transaction) while the stored mode
// is on: the owner turned the bot back on between route() and apply().
type offAtRead struct{ repository.ReplyRouterStore }

func (o offAtRead) ReplySettings(ctx context.Context, tenant string, lock bool) (types.ReplySettings, *apperrors.Error) {
	settings, err := o.ReplyRouterStore.ReplySettings(ctx, tenant, lock)
	settings.Mode = types.ReplyModeOff
	return settings, err
}

func TestRouterRequeuesJobWhenBotTurnsBackOnDuringRouting(t *testing.T) {
	b := newBotFixture(t, types.ReplyModeSuggest)
	b.ai.run = types.AIRun{Route: "clarify", Text: "Bạn cần gì ạ?"}
	message := b.customer(t, "c1", "alo")
	turn := b.turnOf(t, message.ThreadID)
	before := b.proposals(t, turn.ID)[0]

	racing := NewReplyRouterService(offAtRead{repository.NewMessagingRepository(b.db)}, b.scripts, b.ai)
	_, decided, err := racing.RouteDue(context.Background())
	if err != nil || decided != 0 {
		t.Fatalf("a requeued job is not a decision: decided=%d err=%v", decided, err)
	}
	after := b.proposals(t, turn.ID)
	if len(after) != 1 || after[0].Status != types.ProposalRouting || after[0].Version != before.Version+1 || after[0].AvailableAt.Before(before.AvailableAt) {
		t.Fatalf("the job must stay routing and be due again: %+v", after)
	}
	if current := b.turnOf(t, message.ThreadID); current.HandlingState() != "BOT_ACTIVE" || len(b.ai.started) != 0 {
		t.Fatalf("no handoff and no AI call may happen for the bot_off read: %+v", current)
	}
	// The next normal pass decides the job.
	if b.route(t) != 1 {
		t.Fatal("the requeued job must be routed on the next pass")
	}
	if rows := b.proposals(t, turn.ID); rows[0].Status != types.ProposalPending || deref(rows[0].Text) != "Bạn cần gì ạ?" {
		t.Fatalf("unexpected proposal after requeue: %+v", rows[0])
	}
}
