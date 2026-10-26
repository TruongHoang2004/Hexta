package repository

import (
	"context"
	"database/sql"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"gorm.io/gorm"
)

type OverviewRepository struct{ db *gorm.DB }

func NewOverviewRepository(db *gorm.DB) *OverviewRepository { return &OverviewRepository{db: db} }
func overviewDBError(ctx context.Context) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "OverviewReadFailed").SetDetail("unable to read overview")
}
func (r *OverviewRepository) WithinSnapshot(ctx context.Context, fn func(*OverviewRepository) *apperrors.Error) *apperrors.Error {
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

// Plain reads deliberately avoid the locking command-authorization helpers.
func (r *OverviewRepository) Tenant(ctx context.Context, tenant string) (*model.Tenant, *apperrors.Error) {
	var row model.Tenant
	result := r.db.WithContext(ctx).Where("id = ?", tenant).Limit(1).Find(&row)
	if result.Error != nil {
		return nil, overviewDBError(ctx)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrNotFound(ctx, "tenant", tenant)
	}
	return &row, nil
}
func (r *OverviewRepository) Member(ctx context.Context, tenant, actor string) (bool, *apperrors.Error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", tenant, actor).Count(&count).Error; err != nil {
		return false, overviewDBError(ctx)
	}
	return count > 0, nil
}
func overviewPopulation(db *gorm.DB, table, tenant, assignmentColumn string, query types.OverviewQuery) *gorm.DB {
	q := db.Table(table).Where(table+".tenant_id = ?", tenant)
	if query.From != nil {
		q = q.Where(table+".created_at >= ? AND "+table+".created_at < ?", *query.From, *query.To)
	}
	if query.EmployeeID != nil {
		q = q.Where(table+"."+assignmentColumn+" = ?", *query.EmployeeID)
	}
	if query.Assignment == "assigned" {
		q = q.Where(table + "." + assignmentColumn + " IS NOT NULL")
	}
	if query.Assignment == "unassigned" {
		q = q.Where(table + "." + assignmentColumn + " IS NULL")
	}
	return q
}
func (r *OverviewRepository) Read(ctx context.Context, tenant string, query types.OverviewQuery) (*types.OverviewResult, *apperrors.Error) {
	result := &types.OverviewResult{Filters: types.OverviewFilters{From: query.From, To: query.To, EmployeeID: query.EmployeeID, Assignment: query.Assignment}, WorkItems: types.OverviewWorkPage{Items: []types.OverviewWorkRow{}, Page: query.Page, Size: query.Size}}
	orders := overviewPopulation(r.db.WithContext(ctx), "orders", tenant, "employee_id", query).Where("orders.deleted_at IS NULL")
	err := orders.Select(`count(*) AS total,
 count(*) FILTER (WHERE status = 'draft') AS draft,
 count(*) FILTER (WHERE status = 'confirmed') AS confirmed,
 count(*) FILTER (WHERE status = 'processing') AS processing,
 count(*) FILTER (WHERE status = 'fulfilled') AS fulfilled,
 count(*) FILTER (WHERE status = 'cancelled') AS cancelled,
 count(*) FILTER (WHERE status = 'returned') AS returned,
 count(*) FILTER (WHERE status NOT IN ('draft','confirmed','processing','fulfilled','cancelled','returned')) AS other`).Scan(&result.Orders).Error
	if err != nil {
		return nil, overviewDBError(ctx)
	}
	work := func() *gorm.DB {
		return overviewPopulation(r.db.WithContext(ctx), "work_items", tenant, "responsible_id", query)
	}
	err = work().Select(`count(*) AS total, count(*) FILTER (WHERE state = 'active') AS active,
 count(*) FILTER (WHERE state = 'waiting') AS waiting, count(*) FILTER (WHERE state = 'completed') AS completed,
 count(*) FILTER (WHERE responsible_id IS NULL) AS unassigned`).Scan(&result.Work).Error
	if err != nil {
		return nil, overviewDBError(ctx)
	}
	err = work().Select(`work_items.id, work_items.order_id, COALESCE(o.order_number, '') AS order_number,
 work_items.workflow_version_id, v.name AS workflow_name, v.number AS workflow_version_number,
 work_items.current_step, s.name AS current_step_name, work_items.responsible_id, work_items.recipient_id,
 work_items.state, work_items.version, work_items.created_at, work_items.updated_at`).
		Joins("LEFT JOIN orders o ON o.tenant_id = work_items.tenant_id AND o.id = work_items.order_id").
		Joins("JOIN workflow_versions v ON v.tenant_id = work_items.tenant_id AND v.id = work_items.workflow_version_id").
		Joins("JOIN workflow_steps s ON s.tenant_id = work_items.tenant_id AND s.workflow_version_id = work_items.workflow_version_id AND s.ordinal = work_items.current_step").
		Order("work_items.created_at DESC, work_items.id ASC").Limit(query.Size).Offset((query.Page - 1) * query.Size).Scan(&result.WorkItems.Items).Error
	if err != nil {
		return nil, overviewDBError(ctx)
	}
	for i := range result.WorkItems.Items {
		row := &result.WorkItems.Items[i]
		row.CreatedAt = row.CreatedAt.UTC()
		row.UpdatedAt = row.UpdatedAt.UTC()
	}
	result.WorkItems.Total = result.Work.Total
	return result, nil
}
