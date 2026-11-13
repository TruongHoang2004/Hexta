package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

func messagingFixture(t *testing.T) (salesFixture, *MessagingService, *types.MessagingMessage, *types.MessagingTurn) {
	t.Helper()
	f := salesTestFixture(t)
	s := NewMessagingService(repository.NewMessagingRepository(f.db))
	message, err := s.Inbound(context.Background(), f.tenant, f.sale, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Hello", RequestKey: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	detail, err := s.Detail(context.Background(), f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	return f, s, message, detail.Conversation.Turn
}
func TestMessagingClaimReplyAuthorityAndStaleContext(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	cmd := types.MessagingCommand{ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision, Body: "Hello back", RequestKey: uuid.NewString()}
	_, err := s.SendSimulation(ctx, f.tenant, f.owner, turn.ID, cmd)
	requireSalesStatus(t, err, 403)
	turn, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	if turn.Control != "sales" || *turn.AssigneeID != f.sale || turn.ControlGeneration != 2 {
		t.Fatal("claim did not take human control")
	}
	cmd.ExpectedVersion = turn.Version
	_, err = s.SendSimulation(ctx, f.tenant, f.owner, turn.ID, cmd)
	requireSalesStatus(t, err, 403)
	_, err = s.Inbound(ctx, f.tenant, f.peer, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Another question", RequestKey: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, cmd)
	requireSalesStatus(t, err, 409)
	detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	cmd.ExpectedVersion = detail.Conversation.Turn.Version
	cmd.ContextRevision = detail.Conversation.Turn.ContextRevision
	reply, err := s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, cmd)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the hash written before handoff Reason existed in the command.
	type legacyCommand struct {
		ExpectedVersion, ContextRevision int64
		AssigneeID                       *string
		Body, RequestKey                 string
	}
	legacyHash := messagingHash(struct {
		Actor, Turn string
		Input       legacyCommand
	}{f.sale, turn.ID, legacyCommand{cmd.ExpectedVersion, cmd.ContextRevision, cmd.AssigneeID, cmd.Body, cmd.RequestKey}})
	if reply.RequestHash != legacyHash {
		t.Fatal("reply fingerprint changed across the Desk upgrade")
	}
	replay, err := s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, cmd)
	if err != nil || replay.ID != reply.ID {
		t.Fatal("exact retry did not return persisted reply", err)
	}
	cmd.Body = "Different reply"
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, cmd)
	requireSalesStatus(t, err, 409)
	detail, err = s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || len(detail.Messages) != 3 {
		t.Fatal("reply was lost or duplicated", err)
	}
	if detail.Messages[0].Status != "simulated" || detail.Messages[0].Direction != "outbound" {
		t.Fatal("simulation incorrectly labelled")
	}
	_, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: detail.Conversation.Turn.Version, AssigneeID: &f.peer})
	requireSalesStatus(t, err, 403)
	turn, err = s.Command(ctx, f.tenant, f.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: detail.Conversation.Turn.Version, AssigneeID: &f.peer})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, cmd)
	requireSalesStatus(t, err, 403)
	if turn.ControlGeneration != 3 {
		t.Fatal("reassignment did not fence previous control")
	}
}
func TestMessagingTenantRevocationAndAssignment(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	_, err := s.Detail(ctx, f.other, f.owner, message.ThreadID, types.MessagingQuery{Limit: 20})
	requireSalesStatus(t, err, 404)
	_, err = s.Conversations(ctx, f.tenant, "stranger", types.MessagingQuery{Limit: 20})
	requireSalesStatus(t, err, 403)
	_, err = s.Command(ctx, f.other, f.owner, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: 1})
	requireSalesStatus(t, err, 404)
	stranger := "stranger"
	_, err = s.Command(ctx, f.tenant, f.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: 1, AssigneeID: &stranger})
	requireSalesStatus(t, err, 400)
	turn, err = s.Command(ctx, f.tenant, f.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: 1, AssigneeID: &f.sale})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Delete(&model.TenantMember{}).Error; err != nil {
		t.Fatal(err)
	}
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, types.MessagingCommand{ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision, Body: "No", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 403)
	turn, err = s.Command(ctx, f.tenant, f.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	if turn.AssigneeID != nil || turn.Control != "sales" {
		t.Fatal("unassignment silently resumed system control")
	}
}
func TestMessagingConcurrentClaims(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	start := make(chan struct{})
	results := make(chan *apperrors.Error, 2)
	for _, actor := range []string{f.sale, f.peer} {
		go func(actor string) {
			<-start
			_, err := s.Command(ctx, f.tenant, actor, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: 1})
			results <- err
		}(actor)
	}
	close(start)
	winners, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			winners++
		} else if err.GetHttpStatus() == 409 {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("claim winners/conflicts: %d/%d", winners, conflicts)
	}
}
func TestMessagingConcurrentInboundReplay(t *testing.T) {
	f := salesTestFixture(t)
	s := NewMessagingService(repository.NewMessagingRepository(f.db))
	ctx := context.Background()
	input := types.MessagingInbound{ExternalKey: "same-customer", DisplayName: "Customer", Body: "Once", RequestKey: uuid.NewString()}
	start := make(chan struct{})
	results := make(chan string, 2)
	for range 2 {
		go func() {
			<-start
			row, err := s.Inbound(ctx, f.tenant, f.sale, input)
			if err != nil {
				results <- "error"
				return
			}
			results <- row.ID
		}()
	}
	close(start)
	first, second := <-results, <-results
	if first == "error" || first != second {
		t.Fatal("concurrent input replay did not converge", first, second)
	}
	page, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20})
	if err != nil || len(page.Items) != 1 || page.Items[0].Turn.ContextRevision != 1 {
		t.Fatal("duplicate input created extra turn or activity", err)
	}
	input.Body = "Changed"
	_, err = s.Inbound(ctx, f.tenant, f.sale, input)
	requireSalesStatus(t, err, 409)
	input.ExternalKey = "another-thread"
	_, err = s.Inbound(ctx, f.tenant, f.sale, input)
	requireSalesStatus(t, err, 409)
	var count int64
	if err := f.db.Table("messaging_threads").Where("tenant_id = ?", f.tenant).Count(&count).Error; err != nil || count != 1 {
		t.Fatal("failed key mismatch left a thread behind", err)
	}
}
func TestMessagingPaginationAndFilters(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		_, err := s.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: fmt.Sprintf("customer-%d", i+2), DisplayName: "Customer", Body: "Hello", RequestKey: uuid.NewString()})
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	cursor := ""
	for {
		page, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 2, Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range page.Items {
			if seen[row.Thread.ID] {
				t.Fatal("duplicate cursor result")
			}
			seen[row.Thread.ID] = true
		}
		cursor = page.NextCursor
		if cursor == "" {
			break
		}
	}
	if len(seen) != 5 {
		t.Fatal("pagination omitted conversations")
	}
	_, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	mine, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20, Filter: "mine"})
	if err != nil || len(mine.Items) != 1 || mine.Items[0].Thread.ID != message.ThreadID {
		t.Fatal("mine filter failed", err)
	}
	unassigned, err := s.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20, Filter: "unassigned"})
	if err != nil || len(unassigned.Items) != 4 {
		t.Fatal("unassigned filter failed", err)
	}
	for range 3 {
		_, err := s.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "More", RequestKey: uuid.NewString()})
		if err != nil {
			t.Fatal(err)
		}
	}
	detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 2})
	if err != nil || len(detail.Messages) != 2 || detail.NextCursor == "" {
		t.Fatal("history page failed", err)
	}
	older, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 2, Cursor: detail.NextCursor})
	if err != nil || len(older.Messages) != 2 || older.NextCursor != "" {
		t.Fatal("older history page failed", err)
	}
	if detail.Messages[1].ID <= older.Messages[0].ID {
		t.Fatal("history order unstable")
	}
}

type messagingAuditFailure struct{ repository.MessagingStore }

func (r messagingAuditFailure) Within(ctx context.Context, fn func(repository.MessagingStore) *apperrors.Error) *apperrors.Error {
	return r.MessagingStore.Within(ctx, func(tx repository.MessagingStore) *apperrors.Error { return fn(messagingAuditFailure{tx}) })
}
func (r messagingAuditFailure) Audit(ctx context.Context, tenant, actor, action, resource, id string, payload any) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "test audit failure")
}
func TestMessagingAuditFailureRollsBack(t *testing.T) {
	f := salesTestFixture(t)
	s := NewMessagingService(messagingAuditFailure{repository.NewMessagingRepository(f.db)})
	_, err := s.Inbound(context.Background(), f.tenant, f.sale, types.MessagingInbound{ExternalKey: "rollback", DisplayName: "Customer", Body: "Hello", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 500)
	for _, table := range []string{"messaging_threads", "messaging_turns", "messaging_messages"} {
		var count int64
		if err := f.db.Table(table).Where("tenant_id = ?", f.tenant).Count(&count).Error; err != nil || count != 0 {
			t.Fatal("partial transaction persisted", table, err)
		}
	}
}
func TestMessagingInputBounds(t *testing.T) {
	s := NewMessagingService(nil)
	ctx := context.Background()
	tenant := uuid.NewString()
	for _, q := range []types.MessagingQuery{{Limit: 0}, {Limit: 101}, {Limit: 20, Cursor: "bad"}, {Limit: 20, Filter: "arbitrary"}} {
		_, err := s.Conversations(ctx, tenant, "sale", q)
		requireSalesStatus(t, err, 400)
	}
	for _, body := range []string{" ", strings.Repeat("a", 4001)} {
		_, err := s.Inbound(ctx, tenant, "sale", types.MessagingInbound{ExternalKey: "c", DisplayName: "C", Body: body, RequestKey: uuid.NewString()})
		requireSalesStatus(t, err, 400)
	}
	_, err := s.Inbound(ctx, tenant, "sale", types.MessagingInbound{ExternalKey: "c", DisplayName: "C", Body: "Hi", RequestKey: "bad"})
	requireSalesStatus(t, err, 400)
}

func TestMessagingDeskHandoffLifecycle(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	if turn.HandlingState() != "WAITING_STAFF" {
		t.Fatal("simulator without AI must start in staff queue")
	}
	// Reloading the inbox through a new service instance sees persisted work.
	s = NewMessagingService(repository.NewMessagingRepository(f.db))
	page, err := s.Conversations(ctx, f.tenant, f.peer, types.MessagingQuery{Limit: 20, Filter: "needs_human"})
	if err != nil || len(page.Items) != 1 {
		t.Fatal("durable queue lost", err)
	}
	turn, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || turn.HandlingState() != "STAFF_ACTIVE" {
		t.Fatal("claim failed", err)
	}
	oldVersion, oldGeneration, oldContext := turn.Version, turn.ControlGeneration, turn.ContextRevision
	for _, action := range []string{"handoff", "return-to-bot"} {
		_, err = s.Command(ctx, f.tenant, f.peer, turn.ID, action, types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "staff_handoff"})
		requireSalesStatus(t, err, 403)
	}
	turn, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || turn.HandlingState() != "BOT_ACTIVE" || turn.AssigneeID != nil || turn.ControlGeneration != oldGeneration+1 || turn.ContextRevision != oldContext {
		t.Fatal("explicit return failed or replayed old context", err)
	}
	_, err = s.SendSimulation(ctx, f.tenant, f.sale, turn.ID, types.MessagingCommand{ExpectedVersion: oldVersion, ContextRevision: oldContext, Body: "Stale reply", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 403)
	_, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: oldVersion})
	requireSalesStatus(t, err, 409)
	// New input preserves the chosen controller; no implicit return to staff or bot.
	_, err = s.Inbound(ctx, f.tenant, f.peer, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Next question", RequestKey: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || detail.Conversation.Turn.HandlingState() != "BOT_ACTIVE" || len(detail.Messages) != 2 {
		t.Fatal("new input changed bot control", err)
	}
	turn = detail.Conversation.Turn
	turn, err = s.Command(ctx, f.tenant, f.peer, turn.ID, "handoff", types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "customer_requested_staff"})
	if err != nil || turn.HandlingState() != "WAITING_STAFF" {
		t.Fatal("handoff failed", err)
	}
	var audit model.AuditLog
	if err := f.db.Where("tenant_id = ? AND entity_id = ? AND action = ?", f.tenant, turn.ID, "messaging.handoff").First(&audit).Error; err != nil {
		t.Fatal(err)
	}
	var state struct {
		Before, After, Reason string
		ControlGeneration     int64 `json:"control_generation"`
	}
	if err := json.Unmarshal([]byte(audit.AfterState), &state); err != nil {
		t.Fatal(err)
	}
	if state.Before != "BOT_ACTIVE" || state.After != "WAITING_STAFF" || state.Reason != "customer_requested_staff" || state.ControlGeneration != turn.ControlGeneration || audit.UserID != f.peer || audit.CreatedAt.IsZero() {
		t.Fatal("incomplete handoff audit")
	}
	_, err = s.Command(ctx, f.tenant, f.peer, turn.ID, "handoff", types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "staff_handoff"})
	requireSalesStatus(t, err, 409)
	_, err = s.Command(ctx, f.tenant, f.peer, turn.ID, "return-to-bot", types.MessagingCommand{ExpectedVersion: turn.Version})
	requireSalesStatus(t, err, 403)
	_, err = s.Inbound(ctx, f.tenant, f.peer, types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Still waiting", RequestKey: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	detail, err = s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || detail.Conversation.Turn.HandlingState() != "WAITING_STAFF" {
		t.Fatal("waiting queue silently resumed bot", err)
	}
	turn = detail.Conversation.Turn
	turn, err = s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || turn.NeedsHuman || turn.HandlingState() != "STAFF_ACTIVE" {
		t.Fatal("claim did not clear queue flag", err)
	}
	page, err = s.Conversations(ctx, f.tenant, f.peer, types.MessagingQuery{Limit: 20, Filter: "needs_human"})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("claimed work still in queue", err)
	}
	turn, err = s.Command(ctx, f.tenant, f.owner, turn.ID, "assign", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil || turn.HandlingState() != "WAITING_STAFF" {
		t.Fatal("unassignment resumed bot", err)
	}
}

func TestMessagingDeskHandoffAuditRollback(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	turn, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	broken := NewMessagingService(messagingAuditFailure{repository.NewMessagingRepository(f.db)})
	for _, action := range []string{"handoff", "return-to-bot"} {
		_, err = broken.Command(ctx, f.tenant, f.sale, turn.ID, action, types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "staff_handoff"})
		requireSalesStatus(t, err, 500)
		detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
		if err != nil || detail.Conversation.Turn.Version != turn.Version || detail.Conversation.Turn.ControlGeneration != turn.ControlGeneration || detail.Conversation.Turn.HandlingState() != "STAFF_ACTIVE" {
			t.Fatal("failed audit left partial handoff", err)
		}
	}
}

func TestMessagingDeskConcurrentReturnAndReassignment(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	ctx := context.Background()
	turn, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan *apperrors.Error, 2)
	for _, action := range []string{"assign", "return-to-bot"} {
		go func(action string) {
			<-start
			_, err := s.Command(ctx, f.tenant, f.owner, turn.ID, action, types.MessagingCommand{ExpectedVersion: turn.Version, AssigneeID: &f.peer})
			results <- err
		}(action)
	}
	close(start)
	wins, conflicts := 0, 0
	for range 2 {
		err := <-results
		if err == nil {
			wins++
		} else if err.GetHttpStatus() == 409 {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("expected one winner, got %d wins/%d conflicts", wins, conflicts)
	}
	detail, err := s.Detail(ctx, f.tenant, f.sale, message.ThreadID, types.MessagingQuery{Limit: 20})
	if err != nil || detail.Conversation.Turn.Version != turn.Version+1 || detail.Conversation.Turn.ControlGeneration != turn.ControlGeneration+1 {
		t.Fatal("lost control transition", err)
	}
}

func TestMessagingDeskAuthorizationAndBounds(t *testing.T) {
	f, s, _, turn := messagingFixture(t)
	ctx := context.Background()
	for _, action := range []string{"handoff", "return-to-bot"} {
		_, err := s.Command(ctx, f.other, f.owner, turn.ID, action, types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "business_review"})
		requireSalesStatus(t, err, 404)
		_, err = s.Command(ctx, f.tenant, "stranger", turn.ID, action, types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "business_review"})
		requireSalesStatus(t, err, 403)
	}
	for _, reason := range []string{"", "free text", strings.Repeat("x", 500)} {
		_, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "handoff", types.MessagingCommand{ExpectedVersion: turn.Version, Reason: reason})
		requireSalesStatus(t, err, 400)
	}
	if err := f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Delete(&model.TenantMember{}).Error; err != nil {
		t.Fatal(err)
	}
	_, err := s.Command(ctx, f.tenant, f.sale, turn.ID, "handoff", types.MessagingCommand{ExpectedVersion: turn.Version, Reason: "ai_error"})
	requireSalesStatus(t, err, 403)
}

// Older unassignment retained sales control without setting needs_human.
// Keep that durable work visible in the new Desk queue without a data migration.
func TestMessagingDeskLegacyUnassignedQueue(t *testing.T) {
	f, s, message, turn := messagingFixture(t)
	if err := f.db.Table("messaging_turns").Where("tenant_id = ? AND id = ?", f.tenant, turn.ID).Update("needs_human", false).Error; err != nil {
		t.Fatal(err)
	}
	page, err := s.Conversations(context.Background(), f.tenant, f.sale, types.MessagingQuery{Limit: 20, Filter: "needs_human"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Thread.ID != message.ThreadID || page.Items[0].Turn.HandlingState() != "WAITING_STAFF" {
		t.Fatal("legacy unassigned staff work disappeared", err)
	}
}
