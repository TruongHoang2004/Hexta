package service

import (
	"context"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	// OverviewStuckAfter is how long open work may stay unchanged (no handoff, acceptance, reassignment
	// or completion) before the overview lists it as stuck. Unassigned open work is listed immediately.
	OverviewStuckAfter = 48 * time.Hour
	// overviewStuckLimit bounds the waiting/stuck rows; the response still reports the full total.
	overviewStuckLimit = 50
	// overviewStepLimit bounds (workflow version, step) groups; cohort totals are counted separately.
	overviewStepLimit = 300

	OverviewReasonUnassigned     = "unassigned"
	OverviewReasonWaitingTooLong = "waiting_too_long"
	OverviewOrderStatusOther     = "other"
)

// overviewOrderStatuses is the reported bucket order. A new order status (for example an exception
// state) is reported by appending it here; anything not listed is counted in "other", never dropped.
var overviewOrderStatuses = []string{
	string(model.OrderStatusDraft),
	string(model.OrderStatusConfirmed),
	string(model.OrderStatusProcessing),
	string(model.OrderStatusFulfilled),
	string(model.OrderStatusCancelled),
	string(model.OrderStatusReturned),
}

// overviewValueStatuses count toward order value: committed orders that were not cancelled or returned.
// Drafts are not yet orders. Order value is a sum of order totals, never revenue.
var overviewValueStatuses = map[string]bool{
	string(model.OrderStatusConfirmed):  true,
	string(model.OrderStatusProcessing): true,
	string(model.OrderStatusFulfilled):  true,
}

type OverviewService struct {
	store repository.OverviewStore
	now   func() time.Time
}

func NewOverviewService(store repository.OverviewStore) *OverviewService {
	return &OverviewService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

// Overview returns the read-only owner overview. Only the tenant owner may read it (PRD MVP 6 rule 3).
func (s *OverviewService) Overview(ctx context.Context, tenant, actor string, filter types.OverviewFilter) (*types.OverviewResult, *apperrors.Error) {
	if actor == "" {
		return nil, apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	if (filter.From == nil) != (filter.To == nil) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("from and to must be given together")
	}
	if filter.From != nil {
		from, to := filter.From.UTC(), filter.To.UTC()
		if !from.Before(to) {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("from must be before to")
		}
		filter.From, filter.To = &from, &to
	}
	if filter.EmployeeID != nil && *filter.EmployeeID == "" {
		filter.EmployeeID = nil
	}
	idleBefore := s.now().Add(-OverviewStuckAfter)
	result := &types.OverviewResult{Filter: filter}
	err := s.store.WithinSnapshot(ctx, func(r repository.OverviewStore) *apperrors.Error {
		workspace, err := r.Tenant(ctx, tenant)
		if err != nil {
			return err
		}
		if workspace.Status != model.TenantStatusActive {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant is not active")
		}
		member, err := r.IsMember(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !member {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant membership required")
		}
		// A membership role label alone does not grant the overview; the tenant owner identity does.
		if workspace.OwnerID != actor {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
		}
		if filter.EmployeeID != nil {
			employee, err := r.IsMember(ctx, tenant, *filter.EmployeeID)
			if err != nil {
				return err
			}
			if !employee {
				return apperrors.ErrBadRequest(ctx).SetDetail("employee must be a current tenant member")
			}
		}
		if result.HasActivity, err = r.HasActivity(ctx, tenant); err != nil {
			return err
		}
		totals, err := r.OrderStatusTotals(ctx, tenant, filter)
		if err != nil {
			return err
		}
		result.Orders = overviewOrders(totals)
		if result.Work, err = r.WorkCounts(ctx, tenant, filter, overviewStepLimit); err != nil {
			return err
		}
		rows, total, err := r.OpenWork(ctx, tenant, filter, idleBefore, overviewStuckLimit)
		if err != nil {
			return err
		}
		result.Stuck = overviewStuck(rows, total)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// overviewOrders fills every known bucket (zero when absent), folds unknown statuses into "other"
// and sums order value from value-counting statuses only.
func overviewOrders(totals []types.OverviewOrderStatusTotal) types.OverviewOrders {
	result := types.OverviewOrders{OrderValue: decimal.Zero, ByStatus: make([]types.OverviewOrderBucket, 0, len(overviewOrderStatuses)+1)}
	index := map[string]int{}
	for _, status := range overviewOrderStatuses {
		index[status] = len(result.ByStatus)
		result.ByStatus = append(result.ByStatus, types.OverviewOrderBucket{Status: status, Value: decimal.Zero})
	}
	other := types.OverviewOrderBucket{Status: OverviewOrderStatusOther, Value: decimal.Zero}
	for _, row := range totals {
		result.Total += row.Count
		if overviewValueStatuses[row.Status] {
			result.OrderValue = result.OrderValue.Add(row.Value)
		}
		bucket := &other
		if i, known := index[row.Status]; known {
			bucket = &result.ByStatus[i]
		}
		bucket.Count += row.Count
		bucket.Value = bucket.Value.Add(row.Value)
	}
	result.ByStatus = append(result.ByStatus, other)
	return result
}

func overviewStuck(rows []types.OverviewWorkRow, total int64) types.OverviewStuck {
	result := types.OverviewStuck{ThresholdHours: int(OverviewStuckAfter / time.Hour), Total: total, Items: make([]types.OverviewStuckItem, len(rows))}
	for i, row := range rows {
		reason := OverviewReasonWaitingTooLong
		if row.Responsible == nil {
			reason = OverviewReasonUnassigned
		}
		result.Items[i] = types.OverviewStuckItem{OverviewWorkRow: row, Reason: reason}
	}
	return result
}
