package service

import (
	"context"
	"sync"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type orderCaseFixture struct {
	salesFixture
	cases *OrderCaseService
}

func orderCaseTestFixture(t *testing.T) orderCaseFixture {
	t.Helper()
	f := salesTestFixture(t)
	return orderCaseFixture{salesFixture: f, cases: NewOrderCaseService(repository.NewOrderCaseRepository(f.db))}
}

// confirmedOrder creates an order confirmed by the sale with one line of the given quantity.
func (f orderCaseFixture) confirmedOrder(t *testing.T, quantity int) *model.Order {
	t.Helper()
	ctx := context.Background()
	input := f.completeInput(t)
	input.Items[0].Quantity = quantity
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	confirmed, err := f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	if err != nil {
		t.Fatal(err)
	}
	return confirmed
}

func TestOrderCaseOpensOnlyOnConfirmedOrders(t *testing.T) {
	f := orderCaseTestFixture(t)
	ctx := context.Background()
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", types.DraftInput{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.cases.Open(ctx, f.tenant, f.sale, draft.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeCancellation})
	requireSalesStatus(t, err, 409)
	_, err = f.cases.Open(ctx, f.tenant, f.sale, uuid.NewString(), types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	requireSalesStatus(t, err, 404)

	order := f.confirmedOrder(t, 2)
	_, err = f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: "refund"})
	requireSalesStatus(t, err, 400)
	// Another workspace cannot see or open a case on this order, and outsiders are refused.
	_, err = f.cases.Open(ctx, f.other, f.owner, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	requireSalesStatus(t, err, 404)
	_, err = f.cases.Open(ctx, f.tenant, "stranger", order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	requireSalesStatus(t, err, 403)
	// Only the owner assigns someone else as handler.
	_, err = f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn, HandlerID: &f.peer})
	requireSalesStatus(t, err, 403)
	stranger := "stranger"
	_, err = f.cases.Open(ctx, f.tenant, f.owner, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn, HandlerID: &stranger})
	requireSalesStatus(t, err, 400)

	opened, err := f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn, Note: " customer asked "})
	if err != nil {
		t.Fatal(err)
	}
	if opened.Status != types.OrderCaseStatusOpen || opened.HandlerID != f.sale || opened.OpenedBy != f.sale || opened.Version != 1 || opened.ProgressNote != "customer asked" || len(opened.Events) != 1 || opened.Events[0].Action != types.OrderCaseActionOpened {
		t.Fatalf("unexpected opened case %+v", opened)
	}
	// At most one case per order, including a second type.
	_, err = f.cases.Open(ctx, f.tenant, f.owner, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeCancellation})
	requireSalesStatus(t, err, 409)

	got, err := f.cases.Case(ctx, f.tenant, f.peer, order.ID)
	if err != nil || got == nil || got.ID != opened.ID {
		t.Fatalf("members read the case: %+v %v", got, err)
	}
	none, err := f.cases.Case(ctx, f.tenant, f.peer, f.confirmedOrder(t, 1).ID)
	if err != nil || none != nil {
		t.Fatalf("an order without a case returns nil: %+v %v", none, err)
	}
	_, err = f.cases.Case(ctx, f.other, f.owner, order.ID)
	requireSalesStatus(t, err, 404)
}

func TestOrderCaseReturnQuantityCannotExceedOriginal(t *testing.T) {
	f := orderCaseTestFixture(t)
	ctx := context.Background()
	order := f.confirmedOrder(t, 2)
	item := order.Items[0].ID
	opened, err := f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: opened.Version, Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 3}}})
	requireSalesStatus(t, err, 400)
	if err.GetDetail() != "lines[0].quantity: return quantity cannot exceed the original quantity (2)" {
		t.Fatalf("unexpected detail %q", err.GetDetail())
	}
	for _, bad := range [][]types.OrderCaseLineInput{
		{{OrderItemID: uuid.NewString(), Quantity: 1}},
		{{OrderItemID: item, Quantity: 1}, {OrderItemID: item, Quantity: 1}},
		{{OrderItemID: item, Quantity: -1}},
	} {
		_, err = f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: opened.Version, Lines: bad})
		requireSalesStatus(t, err, 400)
	}
	// Refused updates leave no trace.
	current, _ := f.cases.Case(ctx, f.tenant, f.sale, order.ID)
	if current.Version != 1 || len(current.Lines) != 0 || len(current.Events) != 1 {
		t.Fatalf("refused updates must not change the case: %+v", current)
	}
	updated, err := f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: 1, Note: "received", Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || len(updated.Lines) != 1 || updated.Lines[0].Quantity != 2 || updated.Lines[0].OriginalQuantity != 2 {
		t.Fatalf("unexpected lines %+v", updated.Lines)
	}
	// A zero quantity removes the line.
	cleared, err := f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: 2, Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 0}}})
	if err != nil || len(cleared.Lines) != 0 {
		t.Fatalf("zero quantity clears the line: %+v %v", cleared, err)
	}
	// A return case cannot be finalized without returned lines.
	_, err = f.cases.Finalize(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: cleared.Version})
	requireSalesStatus(t, err, 400)

	cancellation := f.confirmedOrder(t, 1)
	c, err := f.cases.Open(ctx, f.tenant, f.sale, cancellation.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeCancellation})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.cases.Update(ctx, f.tenant, f.sale, cancellation.ID, types.UpdateOrderCaseInput{Version: c.Version, Lines: []types.OrderCaseLineInput{{OrderItemID: cancellation.Items[0].ID, Quantity: 1}}})
	requireSalesStatus(t, err, 400)
}

func TestOrderCaseOnlyOwnerFinalizesAndHistoryIsAppended(t *testing.T) {
	f := orderCaseTestFixture(t)
	ctx := context.Background()
	order := f.confirmedOrder(t, 2)
	item := order.Items[0].ID
	opened, err := f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	if err != nil {
		t.Fatal(err)
	}
	// A member who is not the handler cannot record progress.
	_, err = f.cases.Update(ctx, f.tenant, f.peer, order.ID, types.UpdateOrderCaseInput{Version: 1, Note: "not mine"})
	requireSalesStatus(t, err, 403)
	updated, err := f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: opened.Version, Note: "goods received", Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	// Handlers can neither finalize nor send back, even with the current version.
	_, err = f.cases.Finalize(ctx, f.tenant, f.sale, order.ID, types.OrderCaseDecisionInput{Version: updated.Version})
	requireSalesStatus(t, err, 403)
	if err.GetDetail() != "only the workspace owner can finalize a case" {
		t.Fatalf("unexpected detail %q", err.GetDetail())
	}
	_, err = f.cases.SendBack(ctx, f.tenant, f.sale, order.ID, types.OrderCaseDecisionInput{Version: updated.Version, Note: "x"})
	requireSalesStatus(t, err, 403)
	_, err = f.cases.SendBack(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: updated.Version})
	requireSalesStatus(t, err, 400)
	sent, err := f.cases.SendBack(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: updated.Version, Note: "photo missing"})
	if err != nil || sent.Status != types.OrderCaseStatusSentBack {
		t.Fatalf("owner sends back: %+v %v", sent, err)
	}
	_, err = f.cases.SendBack(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: sent.Version, Note: "again"})
	requireSalesStatus(t, err, 409)
	resubmitted, err := f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: sent.Version, Note: "photo attached", Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 2}}})
	if err != nil || resubmitted.Status != types.OrderCaseStatusOpen {
		t.Fatalf("handler update reopens a sent-back case: %+v %v", resubmitted, err)
	}
	// Only the owner reassigns the handler.
	_, err = f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: resubmitted.Version, Note: "photo attached", HandlerID: &f.peer, Lines: []types.OrderCaseLineInput{{OrderItemID: item, Quantity: 2}}})
	requireSalesStatus(t, err, 403)
	finalized, err := f.cases.Finalize(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: resubmitted.Version, Note: "accepted"})
	if err != nil {
		t.Fatal(err)
	}
	if finalized.Status != types.OrderCaseStatusFinalized || finalized.FinalizedBy == nil || *finalized.FinalizedBy != f.owner || finalized.FinalizedAt == nil {
		t.Fatalf("unexpected finalized case %+v", finalized)
	}
	// Finalized cases are read-only for everyone, including the owner.
	_, err = f.cases.Update(ctx, f.tenant, f.owner, order.ID, types.UpdateOrderCaseInput{Version: finalized.Version, Note: "late edit"})
	requireSalesStatus(t, err, 409)
	_, err = f.cases.SendBack(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: finalized.Version, Note: "reopen"})
	requireSalesStatus(t, err, 409)
	_, err = f.cases.Finalize(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: finalized.Version})
	requireSalesStatus(t, err, 409)

	stored, err := f.cases.Case(ctx, f.tenant, f.peer, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantActions := []string{types.OrderCaseActionOpened, types.OrderCaseActionUpdated, types.OrderCaseActionSentBack, types.OrderCaseActionUpdated, types.OrderCaseActionFinalized}
	if len(stored.Events) != len(wantActions) {
		t.Fatalf("expected %d history entries, got %+v", len(wantActions), stored.Events)
	}
	for i, event := range stored.Events {
		if event.Action != wantActions[i] || event.Version != int64(i+1) {
			t.Fatalf("history entry %d: %+v", i, event)
		}
	}
	if stored.Events[1].Lines[0].Quantity != 1 || stored.Events[3].Lines[0].Quantity != 2 || stored.Events[2].Note != "photo missing" || stored.Events[2].ActorID != f.owner {
		t.Fatalf("history keeps each step's lines and notes: %+v", stored.Events)
	}
	// The original order is untouched by the whole case.
	after, salesErr := f.service.Order(ctx, f.tenant, f.owner, order.ID)
	if salesErr != nil || after.Status != model.OrderStatusConfirmed || after.Version != order.Version || after.Items[0].Quantity != 2 || !after.TotalAmount.Equal(order.TotalAmount) {
		t.Fatalf("original order must be preserved: %+v %v", after, salesErr)
	}
	var audits int64
	f.db.Model(&model.AuditLog{}).Where("entity_type = ? AND entity_id = ?", "order_case", finalized.ID).Count(&audits)
	if audits != 5 {
		t.Fatalf("expected one audit per committed case change, got %d", audits)
	}
}

func TestOrderCaseStaleVersionConflicts(t *testing.T) {
	f := orderCaseTestFixture(t)
	ctx := context.Background()
	order := f.confirmedOrder(t, 3)
	opened, err := f.cases.Open(ctx, f.tenant, f.sale, order.ID, types.OpenOrderCaseInput{Type: types.OrderCaseTypeReturn})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make([]*apperrors.Error, 4)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = f.cases.Update(ctx, f.tenant, f.sale, order.ID, types.UpdateOrderCaseInput{Version: opened.Version, Note: "racing", Lines: []types.OrderCaseLineInput{{OrderItemID: order.Items[0].ID, Quantity: int64(i%3 + 1)}}})
		}(i)
	}
	wg.Wait()
	succeeded := 0
	for _, result := range results {
		if result == nil {
			succeeded++
			continue
		}
		requireSalesStatus(t, result, 409)
	}
	if succeeded != 1 {
		t.Fatalf("exactly one racing update may win, got %d", succeeded)
	}
	_, err = f.cases.Finalize(ctx, f.tenant, f.owner, order.ID, types.OrderCaseDecisionInput{Version: opened.Version})
	requireSalesStatus(t, err, 409)
	current, _ := f.cases.Case(ctx, f.tenant, f.owner, order.ID)
	if current.Version != 2 || len(current.Events) != 2 {
		t.Fatalf("conflicts must not append history: %+v", current)
	}
}
