package repository

import (
	"context"
	stderrors "errors"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/ordercase/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/ordercasemodel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderCaseAccess is the caller's current standing in a workspace, read inside the case transaction.
type OrderCaseAccess struct {
	Member, Owner, Active bool
}

// OrderCaseStore persists order exception cases. Events are insert-only: the store exposes no event update/delete.
type OrderCaseStore interface {
	Within(context.Context, func(OrderCaseStore) *apperrors.Error) *apperrors.Error
	Access(ctx context.Context, tenant, user string) (OrderCaseAccess, *apperrors.Error)
	IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error)
	// Order loads the original order with its lines; lock serializes case commands on the same order.
	Order(ctx context.Context, tenant, id string, lock bool) (*types.OrderCaseOrder, *apperrors.Error)
	// CaseByOrder returns nil when the order has no case; lock takes the case row for update.
	CaseByOrder(ctx context.Context, tenant, orderID string, lock bool) (*types.OrderCase, *apperrors.Error)
	InsertCase(ctx context.Context, value types.OrderCase) *apperrors.Error
	// SaveCase applies the case header on the expected version and, when replaceLines is set, its current lines.
	SaveCase(ctx context.Context, value types.OrderCase, expected int64, replaceLines bool) *apperrors.Error
	AppendEvent(ctx context.Context, tenant, caseID string, event types.OrderCaseEvent) *apperrors.Error
	Audit(ctx context.Context, tenant, actor, action, kind, id string, state any) *apperrors.Error
}

type OrderCaseRepository struct{ *SalesRepository }

func NewOrderCaseRepository(db *gorm.DB) *OrderCaseRepository {
	return &OrderCaseRepository{NewSalesRepository(db)}
}

func (r *OrderCaseRepository) Within(ctx context.Context, fn func(OrderCaseStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if failure = fn(NewOrderCaseRepository(tx)); failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return orderCaseDBError(ctx, err)
	}
	return nil
}

// orderCaseDBError maps the case constraints to stable public failures without SQL detail.
func orderCaseDBError(ctx context.Context, err error) *apperrors.Error {
	var pg *pgconn.PgError
	if stderrors.As(err, &pg) {
		switch pg.ConstraintName {
		case "order_cases_tenant_id_order_id_key":
			return apperrors.ErrConflict(ctx, "order", "already has an exception case")
		case "order_case_events_case_id_version_key":
			return apperrors.ErrConflict(ctx, "order case", "changed; reload before saving")
		case "order_case_lines_quantity_check":
			return apperrors.ErrBadRequest(ctx).SetDetail("lines.quantity: return quantity must be positive and cannot exceed the original quantity")
		case "order_case_lines_item_fkey", "order_case_lines_case_id_order_item_id_key":
			return apperrors.ErrBadRequest(ctx).SetDetail("lines.order_item_id: each line must match a different line of the original order")
		}
	}
	return salesDBError(ctx, err)
}

func (r *OrderCaseRepository) Access(ctx context.Context, tenant, user string) (OrderCaseAccess, *apperrors.Error) {
	member, err := r.Member(ctx, tenant, user)
	if err != nil || member == nil {
		return OrderCaseAccess{}, err
	}
	workspace, err := r.Tenant(ctx, tenant)
	if err != nil {
		return OrderCaseAccess{}, err
	}
	return OrderCaseAccess{Member: true, Owner: workspace.OwnerID == user, Active: workspace.Status == model.TenantStatusActive}, nil
}

func (r *OrderCaseRepository) IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error) {
	member, err := r.Member(ctx, tenant, user)
	return member != nil, err
}

func (r *OrderCaseRepository) Order(ctx context.Context, tenant, id string, lock bool) (*types.OrderCaseOrder, *apperrors.Error) {
	var row model.Order
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("created_at, id") }).
		Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound(ctx, "order", "not found")
	}
	if err != nil {
		return nil, orderCaseDBError(ctx, err)
	}
	value := persistencemap.Order(row)
	return &value, nil
}

func (r *OrderCaseRepository) CaseByOrder(ctx context.Context, tenant, orderID string, lock bool) (*types.OrderCase, *apperrors.Error) {
	var rows []ordercasemodel.Case
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Where("tenant_id = ? AND order_id = ?", tenant, orderID).Limit(1).Find(&rows).Error; err != nil {
		return nil, orderCaseDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	var lines []ordercasemodel.Line
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND case_id = ?", tenant, rows[0].ID).Order("created_at, order_item_id").Find(&lines).Error; err != nil {
		return nil, orderCaseDBError(ctx, err)
	}
	// History is bounded by the service's per-case version limit, so loading it whole stays bounded.
	var events []ordercasemodel.Event
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND case_id = ?", tenant, rows[0].ID).Order("version").Find(&events).Error; err != nil {
		return nil, orderCaseDBError(ctx, err)
	}
	value, err := persistencemap.Case(rows[0], lines, events)
	if err != nil {
		return nil, apperrors.ErrSystemError(ctx, "stored order case is not readable")
	}
	return &value, nil
}

func (r *OrderCaseRepository) InsertCase(ctx context.Context, value types.OrderCase) *apperrors.Error {
	row := persistencemap.CaseRow(value)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return orderCaseDBError(ctx, err)
	}
	return r.insertLines(ctx, value)
}

func (r *OrderCaseRepository) SaveCase(ctx context.Context, value types.OrderCase, expected int64, replaceLines bool) *apperrors.Error {
	row := persistencemap.CaseRow(value)
	result := r.db.WithContext(ctx).Model(&ordercasemodel.Case{}).
		Where("tenant_id = ? AND id = ? AND version = ? AND status <> ?", value.TenantID, value.ID, expected, types.OrderCaseStatusFinalized).
		Select("status", "handler_id", "progress_note", "finalized_by", "finalized_at", "version", "updated_at").Updates(&row)
	if result.Error != nil {
		return orderCaseDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "order case", "changed or is finalized; reload before saving")
	}
	if !replaceLines {
		return nil
	}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND case_id = ?", value.TenantID, value.ID).Delete(&ordercasemodel.Line{}).Error; err != nil {
		return orderCaseDBError(ctx, err)
	}
	return r.insertLines(ctx, value)
}

func (r *OrderCaseRepository) insertLines(ctx context.Context, value types.OrderCase) *apperrors.Error {
	if len(value.Lines) == 0 {
		return nil
	}
	ids := make([]string, len(value.Lines))
	for i := range ids {
		ids[i] = uuid.NewString()
	}
	rows := persistencemap.LineRows(value, ids)
	if err := r.db.WithContext(ctx).Create(&rows).Error; err != nil {
		return orderCaseDBError(ctx, err)
	}
	return nil
}

func (r *OrderCaseRepository) AppendEvent(ctx context.Context, tenant, caseID string, event types.OrderCaseEvent) *apperrors.Error {
	row := persistencemap.EventRow(tenant, caseID, event)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return orderCaseDBError(ctx, err)
	}
	return nil
}
