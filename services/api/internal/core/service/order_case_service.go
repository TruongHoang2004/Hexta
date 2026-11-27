package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

const (
	// orderCaseMaxVersion bounds the history of one case (one event per version) so reads stay bounded.
	orderCaseMaxVersion = 500
	orderCaseMaxNote    = 5000
	orderCaseMaxLines   = 100
)

// OrderCaseService runs cancellation/return cases on confirmed orders (MVP 6, FR-6.1–6.3).
//
// Rules: a case opens only on a confirmed order, at most one case per order (a finalized case is terminal);
// the assigned handler or the owner records progress and return lines; only the owner sends back or finalizes;
// a finalized case is read-only. The original order is never changed and nothing is refunded or restocked.
type OrderCaseService struct {
	store repository.OrderCaseStore
	now   func() time.Time
}

func NewOrderCaseService(store repository.OrderCaseStore) *OrderCaseService {
	return &OrderCaseService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

// authorized rechecks current membership and workspace status inside the case transaction.
func (s *OrderCaseService) authorized(ctx context.Context, tenant, actor string, fn func(repository.OrderCaseStore, bool) *apperrors.Error) *apperrors.Error {
	if tenant == "" {
		return apperrors.ErrBadRequest(ctx).SetDetail("tenant ID is required")
	}
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx).SetDetail("authentication required")
	}
	return s.store.Within(ctx, func(r repository.OrderCaseStore) *apperrors.Error {
		access, err := r.Access(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !access.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("workspace membership required")
		}
		if !access.Active {
			return apperrors.ErrForbidden(ctx).SetDetail("workspace is not active")
		}
		return fn(r, access.Owner)
	})
}

func cleanNote(ctx context.Context, note string) (string, *apperrors.Error) {
	note = strings.TrimSpace(note)
	if len([]rune(note)) > orderCaseMaxNote {
		return "", apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("note: at most %d characters", orderCaseMaxNote))
	}
	return note, nil
}

// Case returns the order's case with lines and history, or nil when the order has none.
func (s *OrderCaseService) Case(ctx context.Context, tenant, actor, orderID string) (*types.OrderCase, *apperrors.Error) {
	var result *types.OrderCase
	err := s.authorized(ctx, tenant, actor, func(r repository.OrderCaseStore, _ bool) *apperrors.Error {
		if _, err := r.Order(ctx, tenant, orderID, false); err != nil {
			return err
		}
		var err *apperrors.Error
		result, err = r.CaseByOrder(ctx, tenant, orderID, false)
		return err
	})
	return result, err
}

func (s *OrderCaseService) Open(ctx context.Context, tenant, actor, orderID string, input types.OpenOrderCaseInput) (*types.OrderCase, *apperrors.Error) {
	if input.Type != types.OrderCaseTypeCancellation && input.Type != types.OrderCaseTypeReturn {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("type: must be cancellation or return")
	}
	note, failure := cleanNote(ctx, input.Note)
	if failure != nil {
		return nil, failure
	}
	handler := actor
	if id := cleanID(input.HandlerID); id != nil {
		handler = *id
	}
	var result *types.OrderCase
	err := s.authorized(ctx, tenant, actor, func(r repository.OrderCaseStore, owner bool) *apperrors.Error {
		if handler != actor && !owner {
			return apperrors.ErrForbidden(ctx).SetDetail("only the owner can assign another handler")
		}
		order, err := r.Order(ctx, tenant, orderID, true)
		if err != nil {
			return err
		}
		if order.Status != string(model.OrderStatusConfirmed) {
			return apperrors.ErrConflict(ctx, "order", "must be confirmed before a cancellation or return case is opened")
		}
		existing, err := r.CaseByOrder(ctx, tenant, orderID, true)
		if err != nil {
			return err
		}
		if existing != nil {
			return apperrors.ErrConflict(ctx, "order", "already has an exception case")
		}
		if handler != actor {
			member, err := r.IsMember(ctx, tenant, handler)
			if err != nil {
				return err
			}
			if !member {
				return apperrors.ErrBadRequest(ctx).SetDetail("handler_id: must identify a current workspace member")
			}
		}
		now := s.now()
		value := types.OrderCase{ID: uuid.NewString(), TenantID: tenant, OrderID: orderID, Type: input.Type, Status: types.OrderCaseStatusOpen, HandlerID: handler, OpenedBy: actor, ProgressNote: note, Version: 1, CreatedAt: now, UpdatedAt: now}
		if err := r.InsertCase(ctx, value); err != nil {
			return err
		}
		event := types.OrderCaseEvent{ID: uuid.NewString(), ActorID: actor, Action: types.OrderCaseActionOpened, Status: value.Status, Note: note, Version: 1, CreatedAt: now}
		if err := r.AppendEvent(ctx, tenant, value.ID, event); err != nil {
			return err
		}
		if err := r.Audit(ctx, tenant, actor, "order_case.opened", "order_case", value.ID, value); err != nil {
			return err
		}
		value.Lines = []types.OrderCaseLine{}
		value.Events = []types.OrderCaseEvent{event}
		result = &value
		return nil
	})
	return result, err
}

// lockedCase loads the order's case for a command, holding its row lock until the transaction ends.
func lockedCase(ctx context.Context, r repository.OrderCaseStore, tenant, orderID string) (*types.OrderCase, *apperrors.Error) {
	value, err := r.CaseByOrder(ctx, tenant, orderID, true)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, apperrors.ErrNotFound(ctx, "order case", "not found")
	}
	return value, nil
}

// checkEditable applies the shared terminal-state, optimistic-version and history-bound checks.
func checkEditable(ctx context.Context, value *types.OrderCase, version int64) *apperrors.Error {
	if value.Status == types.OrderCaseStatusFinalized {
		return apperrors.ErrConflict(ctx, "order case", "is finalized and read-only")
	}
	if value.Version != version {
		return apperrors.ErrConflict(ctx, "order case", "changed; reload before saving")
	}
	if value.Version >= orderCaseMaxVersion {
		return apperrors.ErrConflict(ctx, "order case", "reached its history limit")
	}
	return nil
}

// returnLines validates requested lines against the original order. Zero quantities are dropped.
func returnLines(ctx context.Context, caseType string, order *types.OrderCaseOrder, input []types.OrderCaseLineInput) ([]types.OrderCaseLine, *apperrors.Error) {
	if len(input) > orderCaseMaxLines {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("lines: at most %d lines are supported", orderCaseMaxLines))
	}
	original := make(map[string]int64, len(order.Items))
	for _, item := range order.Items {
		original[item.ID] = item.Quantity
	}
	seen := make(map[string]bool, len(input))
	lines := make([]types.OrderCaseLine, 0, len(input))
	for i, line := range input {
		quantity, ok := original[line.OrderItemID]
		if !ok {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("lines[%d].order_item_id: not a line of this order", i))
		}
		if seen[line.OrderItemID] {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("lines[%d].order_item_id: listed more than once", i))
		}
		seen[line.OrderItemID] = true
		if line.Quantity < 0 {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("lines[%d].quantity: cannot be negative", i))
		}
		if line.Quantity > quantity {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("lines[%d].quantity: return quantity cannot exceed the original quantity (%d)", i, quantity))
		}
		if line.Quantity == 0 {
			continue
		}
		lines = append(lines, types.OrderCaseLine{OrderItemID: line.OrderItemID, OriginalQuantity: quantity, Quantity: line.Quantity})
	}
	if caseType == types.OrderCaseTypeCancellation && len(lines) > 0 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("lines: a cancellation case has no return lines")
	}
	return lines, nil
}

// Update records progress and the current return lines. A sent-back case returns to open.
func (s *OrderCaseService) Update(ctx context.Context, tenant, actor, orderID string, input types.UpdateOrderCaseInput) (*types.OrderCase, *apperrors.Error) {
	note, failure := cleanNote(ctx, input.Note)
	if failure != nil {
		return nil, failure
	}
	var result *types.OrderCase
	err := s.authorized(ctx, tenant, actor, func(r repository.OrderCaseStore, owner bool) *apperrors.Error {
		// Confirmed order lines are immutable, so reading the original quantities needs no order lock.
		order, err := r.Order(ctx, tenant, orderID, false)
		if err != nil {
			return err
		}
		value, err := lockedCase(ctx, r, tenant, orderID)
		if err != nil {
			return err
		}
		if !owner && value.HandlerID != actor {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned handler or the owner can update this case")
		}
		if err := checkEditable(ctx, value, input.Version); err != nil {
			return err
		}
		if id := cleanID(input.HandlerID); id != nil && *id != value.HandlerID {
			if !owner {
				return apperrors.ErrForbidden(ctx).SetDetail("only the owner can change the handler")
			}
			member, err := r.IsMember(ctx, tenant, *id)
			if err != nil {
				return err
			}
			if !member {
				return apperrors.ErrBadRequest(ctx).SetDetail("handler_id: must identify a current workspace member")
			}
			value.HandlerID = *id
		}
		lines, err := returnLines(ctx, value.Type, order, input.Lines)
		if err != nil {
			return err
		}
		value.Lines = lines
		value.ProgressNote = note
		value.Status = types.OrderCaseStatusOpen
		result, err = s.commit(ctx, r, value, actor, types.OrderCaseActionUpdated, note, true)
		return err
	})
	return result, err
}

// SendBack is the owner's review decision returning the case to the handler with a required reason.
func (s *OrderCaseService) SendBack(ctx context.Context, tenant, actor, orderID string, input types.OrderCaseDecisionInput) (*types.OrderCase, *apperrors.Error) {
	note, failure := cleanNote(ctx, input.Note)
	if failure != nil {
		return nil, failure
	}
	if note == "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("note: explain what the handler must change")
	}
	var result *types.OrderCase
	err := s.authorized(ctx, tenant, actor, func(r repository.OrderCaseStore, owner bool) *apperrors.Error {
		value, err := lockedCase(ctx, r, tenant, orderID)
		if err != nil {
			return err
		}
		if !owner {
			return apperrors.ErrForbidden(ctx).SetDetail("only the workspace owner can send a case back")
		}
		if err := checkEditable(ctx, value, input.Version); err != nil {
			return err
		}
		if value.Status != types.OrderCaseStatusOpen {
			return apperrors.ErrConflict(ctx, "order case", "is already sent back; wait for the handler's update")
		}
		value.Status = types.OrderCaseStatusSentBack
		result, err = s.commit(ctx, r, value, actor, types.OrderCaseActionSentBack, note, false)
		return err
	})
	return result, err
}

// Finalize is the owner's final decision; the case becomes read-only. A return case needs at least one line.
func (s *OrderCaseService) Finalize(ctx context.Context, tenant, actor, orderID string, input types.OrderCaseDecisionInput) (*types.OrderCase, *apperrors.Error) {
	note, failure := cleanNote(ctx, input.Note)
	if failure != nil {
		return nil, failure
	}
	var result *types.OrderCase
	err := s.authorized(ctx, tenant, actor, func(r repository.OrderCaseStore, owner bool) *apperrors.Error {
		value, err := lockedCase(ctx, r, tenant, orderID)
		if err != nil {
			return err
		}
		if !owner {
			return apperrors.ErrForbidden(ctx).SetDetail("only the workspace owner can finalize a case")
		}
		if err := checkEditable(ctx, value, input.Version); err != nil {
			return err
		}
		if value.Type == types.OrderCaseTypeReturn && len(value.Lines) == 0 {
			return apperrors.ErrBadRequest(ctx).SetDetail("lines: record at least one returned line before finalizing")
		}
		now := s.now()
		finalizer := actor
		value.Status = types.OrderCaseStatusFinalized
		value.FinalizedBy = &finalizer
		value.FinalizedAt = &now
		result, err = s.commit(ctx, r, value, actor, types.OrderCaseActionFinalized, note, false)
		return err
	})
	return result, err
}

// commit bumps the version, saves the case, appends its history event and audits in the same transaction.
func (s *OrderCaseService) commit(ctx context.Context, r repository.OrderCaseStore, value *types.OrderCase, actor, action, note string, replaceLines bool) (*types.OrderCase, *apperrors.Error) {
	expected := value.Version
	now := s.now()
	value.Version++
	value.UpdatedAt = now
	if err := r.SaveCase(ctx, *value, expected, replaceLines); err != nil {
		return nil, err
	}
	snapshot := append([]types.OrderCaseLine{}, value.Lines...)
	event := types.OrderCaseEvent{ID: uuid.NewString(), ActorID: actor, Action: action, Status: value.Status, Note: note, Version: value.Version, Lines: snapshot, CreatedAt: now}
	if err := r.AppendEvent(ctx, value.TenantID, value.ID, event); err != nil {
		return nil, err
	}
	// The audit row keeps the case state, not the whole history already stored in order_case_events.
	audited := *value
	audited.Events = nil
	if err := r.Audit(ctx, value.TenantID, actor, "order_case."+action, "order_case", value.ID, audited); err != nil {
		return nil, err
	}
	value.Events = append(value.Events, event)
	return value, nil
}
