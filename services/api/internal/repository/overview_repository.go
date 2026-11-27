package repository

import (
	"context"
	"database/sql"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/overview/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/overviewmodel"
	"gorm.io/gorm"
)

// OverviewStore reads the owner overview. Every read inside WithinSnapshot sees one consistent snapshot.
type OverviewStore interface {
	WithinSnapshot(context.Context, func(OverviewStore) *apperrors.Error) *apperrors.Error
	Tenant(ctx context.Context, tenant string) (*model.Tenant, *apperrors.Error)
	IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error)
	HasActivity(ctx context.Context, tenant string) (bool, *apperrors.Error)
	OrderStatusTotals(ctx context.Context, tenant string, filter types.OverviewFilter) ([]types.OverviewOrderStatusTotal, *apperrors.Error)
	CaseTotals(ctx context.Context, tenant string, filter types.OverviewFilter) ([]types.OverviewCaseTotal, *apperrors.Error)
	PendingCases(ctx context.Context, tenant string, filter types.OverviewFilter, limit int) ([]types.OverviewPendingCase, int64, *apperrors.Error)
	WorkCounts(ctx context.Context, tenant string, filter types.OverviewFilter, stepLimit int) (types.OverviewWork, *apperrors.Error)
	OpenWork(ctx context.Context, tenant string, filter types.OverviewFilter, idleBefore time.Time, limit int) ([]types.OverviewWorkRow, int64, *apperrors.Error)
}

type OverviewRepository struct{ db *gorm.DB }

func NewOverviewRepository(db *gorm.DB) *OverviewRepository { return &OverviewRepository{db: db} }

func overviewDBError(ctx context.Context) *apperrors.Error {
	// Never echo SQL/connection details into the API response.
	return apperrors.ErrSystemError(ctx, "unable to read overview")
}

// WithinSnapshot runs fn in one read-only REPEATABLE READ transaction so counts and rows agree (REP-009).
func (r *OverviewRepository) WithinSnapshot(ctx context.Context, fn func(OverviewStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(NewOverviewRepository(tx))
		if failure != nil {
			return failure
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if failure != nil {
		return failure
	}
	if err != nil {
		return overviewDBError(ctx)
	}
	return nil
}

// Tenant is a plain read; the locking sales helpers cannot run inside a read-only transaction.
func (r *OverviewRepository) Tenant(ctx context.Context, tenant string) (*model.Tenant, *apperrors.Error) {
	var row model.Tenant
	result := r.db.WithContext(ctx).Where("id = ?", tenant).Limit(1).Find(&row)
	if result.Error != nil {
		return nil, overviewDBError(ctx)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrNotFound(ctx, "tenant", "not found")
	}
	return &row, nil
}

func (r *OverviewRepository) IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", tenant, user).Count(&count).Error; err != nil {
		return false, overviewDBError(ctx)
	}
	return count > 0, nil
}

func (r *OverviewRepository) HasActivity(ctx context.Context, tenant string) (bool, *apperrors.Error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(`SELECT EXISTS (SELECT 1 FROM orders WHERE tenant_id = ? AND deleted_at IS NULL)
 OR EXISTS (SELECT 1 FROM work_items WHERE tenant_id = ?)`, tenant, tenant).Scan(&exists).Error
	if err != nil {
		return false, overviewDBError(ctx)
	}
	return exists, nil
}

// orderCohort is the live orders of the tenant created in the interval and sold by the employee.
func (r *OverviewRepository) orderCohort(ctx context.Context, tenant string, filter types.OverviewFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Table("orders").Where("orders.tenant_id = ? AND orders.deleted_at IS NULL", tenant)
	if filter.From != nil && filter.To != nil {
		q = q.Where("orders.created_at >= ? AND orders.created_at < ?", *filter.From, *filter.To)
	}
	if filter.EmployeeID != nil {
		q = q.Where("orders.employee_id = ?", *filter.EmployeeID)
	}
	return q
}

// workCohort is the tenant's work created in the interval whose current responsible is the employee.
func (r *OverviewRepository) workCohort(ctx context.Context, tenant string, filter types.OverviewFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Table("work_items").Where("work_items.tenant_id = ?", tenant)
	if filter.From != nil && filter.To != nil {
		q = q.Where("work_items.created_at >= ? AND work_items.created_at < ?", *filter.From, *filter.To)
	}
	if filter.EmployeeID != nil {
		q = q.Where("work_items.responsible_id = ?", *filter.EmployeeID)
	}
	return q
}

// pinnedStep joins each work item to its pinned version and current step, never the latest definition.
func pinnedStep(q *gorm.DB) *gorm.DB {
	return q.Joins("JOIN workflow_versions v ON v.tenant_id = work_items.tenant_id AND v.id = work_items.workflow_version_id").
		Joins("JOIN workflow_steps s ON s.tenant_id = work_items.tenant_id AND s.workflow_version_id = work_items.workflow_version_id AND s.ordinal = work_items.current_step")
}

// OrderStatusTotals groups by stored status so a new status needs no SQL change.
func (r *OverviewRepository) OrderStatusTotals(ctx context.Context, tenant string, filter types.OverviewFilter) ([]types.OverviewOrderStatusTotal, *apperrors.Error) {
	rows := []overviewmodel.OrderStatusRow{}
	err := r.orderCohort(ctx, tenant, filter).
		Select("orders.status AS status, count(*) AS count, COALESCE(sum(orders.total_amount), 0) AS value").
		Group("orders.status").Order("orders.status").Scan(&rows).Error
	if err != nil {
		return nil, overviewDBError(ctx)
	}
	return persistencemap.OrderStatusTotals(rows), nil
}

// orderCases joins the order cohort to its exception cases. A case is unique per (tenant, order), so the join
// never duplicates an order, and the tenant is matched on both sides.
func orderCases(q *gorm.DB) *gorm.DB {
	return q.Joins("JOIN order_cases c ON c.tenant_id = orders.tenant_id AND c.order_id = orders.id")
}

// CaseTotals groups the cohort's cases as stored so the service decides stages and order value.
func (r *OverviewRepository) CaseTotals(ctx context.Context, tenant string, filter types.OverviewFilter) ([]types.OverviewCaseTotal, *apperrors.Error) {
	rows := []overviewmodel.CaseTotalRow{}
	err := orderCases(r.orderCohort(ctx, tenant, filter)).
		Select("orders.status AS order_status, c.case_type AS case_type, c.status AS case_status, count(*) AS count, COALESCE(sum(orders.total_amount), 0) AS value").
		Group("orders.status, c.case_type, c.status").Order("orders.status, c.case_type, c.status").Scan(&rows).Error
	if err != nil {
		return nil, overviewDBError(ctx)
	}
	return persistencemap.CaseTotals(rows), nil
}

// PendingCases lists the cohort's open cases (waiting for the owner's decision), longest unchanged first.
func (r *OverviewRepository) PendingCases(ctx context.Context, tenant string, filter types.OverviewFilter, limit int) ([]types.OverviewPendingCase, int64, *apperrors.Error) {
	open := func() *gorm.DB {
		return orderCases(r.orderCohort(ctx, tenant, filter)).Where("c.status = ?", types.OrderCaseStatusOpen)
	}
	var total int64
	if err := open().Count(&total).Error; err != nil {
		return nil, 0, overviewDBError(ctx)
	}
	rows := []overviewmodel.PendingCaseRow{}
	err := open().Select(`c.id AS case_id, orders.id AS order_id, orders.order_number AS order_number,
 c.case_type AS case_type, c.handler_id AS handler_id, c.updated_at AS updated_at`).
		Order("c.updated_at ASC, c.id ASC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, 0, overviewDBError(ctx)
	}
	return persistencemap.PendingCases(rows), total, nil
}

func (r *OverviewRepository) WorkCounts(ctx context.Context, tenant string, filter types.OverviewFilter, stepLimit int) (types.OverviewWork, *apperrors.Error) {
	var totals overviewmodel.WorkTotalsRow
	err := r.workCohort(ctx, tenant, filter).Select(`count(*) AS total,
 count(*) FILTER (WHERE work_items.state = 'active') AS active,
 count(*) FILTER (WHERE work_items.state = 'waiting') AS waiting,
 count(*) FILTER (WHERE work_items.state = 'completed') AS completed,
 count(*) FILTER (WHERE work_items.state = 'active' AND work_items.responsible_id IS NULL) AS unassigned`).Scan(&totals).Error
	if err != nil {
		return types.OverviewWork{}, overviewDBError(ctx)
	}
	steps := []overviewmodel.StepCountRow{}
	err = pinnedStep(r.workCohort(ctx, tenant, filter)).Select(`work_items.workflow_version_id AS workflow_version_id,
 v.name AS workflow_name, v.number AS workflow_version_number,
 work_items.current_step AS step_ordinal, s.name AS step_name,
 count(*) FILTER (WHERE work_items.state = 'active') AS active,
 count(*) FILTER (WHERE work_items.state = 'waiting') AS waiting,
 count(*) FILTER (WHERE work_items.state = 'completed') AS completed,
 count(*) FILTER (WHERE work_items.state = 'active' AND work_items.responsible_id IS NULL) AS unassigned`).
		Group("work_items.workflow_version_id, v.name, v.number, work_items.current_step, s.name").
		Order("v.name, v.number, work_items.workflow_version_id, work_items.current_step").Limit(stepLimit).Scan(&steps).Error
	if err != nil {
		return types.OverviewWork{}, overviewDBError(ctx)
	}
	return persistencemap.WorkTotals(totals, steps), nil
}

// OpenWork lists active/waiting work that is unassigned or unchanged since idleBefore, longest waiting first.
func (r *OverviewRepository) OpenWork(ctx context.Context, tenant string, filter types.OverviewFilter, idleBefore time.Time, limit int) ([]types.OverviewWorkRow, int64, *apperrors.Error) {
	open := func() *gorm.DB {
		return r.workCohort(ctx, tenant, filter).
			Where("work_items.state IN ('active', 'waiting') AND (work_items.responsible_id IS NULL OR work_items.updated_at <= ?)", idleBefore)
	}
	var total int64
	if err := open().Count(&total).Error; err != nil {
		return nil, 0, overviewDBError(ctx)
	}
	rows := []overviewmodel.WorkRow{}
	err := pinnedStep(open()).
		Joins("LEFT JOIN orders o ON o.tenant_id = work_items.tenant_id AND o.id = work_items.order_id").
		Select(`work_items.id AS work_item_id, work_items.order_id AS order_id, COALESCE(o.order_number, '') AS order_number,
 v.name AS workflow_name, work_items.current_step AS step_ordinal, s.name AS step_name, work_items.state AS state,
 work_items.responsible_id AS responsible_id, work_items.recipient_id AS recipient_id, work_items.updated_at AS updated_at`).
		Order("work_items.updated_at ASC, work_items.id ASC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, 0, overviewDBError(ctx)
	}
	return persistencemap.WorkRows(rows), total, nil
}
