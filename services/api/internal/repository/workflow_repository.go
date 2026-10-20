package repository

import (
	"context"
	stderrors "errors"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WorkflowRepository struct{ *SalesRepository }

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{NewSalesRepository(db)}
}

func (r *WorkflowRepository) Within(ctx context.Context, fn func(*WorkflowRepository) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(NewWorkflowRepository(tx))
		if failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *WorkflowRepository) Definitions(ctx context.Context, tenant string, limit, offset int) ([]WorkflowDefinition, int64, *apperrors.Error) {
	rows := []WorkflowDefinition{}
	var total int64
	q := r.db.WithContext(ctx).Model(&WorkflowDefinition{}).Where("tenant_id = ?", tenant)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	err := q.Preload("Versions", "tenant_id = ? AND EXISTS (SELECT 1 FROM workflow_definitions d WHERE d.id = workflow_versions.definition_id AND d.version = workflow_versions.number)", tenant).
		Preload("Versions.Steps", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("ordinal") }).Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	return rows, total, nil
}
func (r *WorkflowRepository) Definition(ctx context.Context, tenant, id string) (*WorkflowDefinition, *apperrors.Error) {
	var row WorkflowDefinition
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ?", tenant, id).Preload("Versions", "tenant_id = ? AND EXISTS (SELECT 1 FROM workflow_definitions d WHERE d.id = workflow_versions.definition_id AND d.version = workflow_versions.number)", tenant).Preload("Versions.Steps", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("ordinal") }).First(&row).Error
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *WorkflowRepository) InsertDefinition(ctx context.Context, row *WorkflowDefinition) (bool, *apperrors.Error) {
	result := r.db.WithContext(ctx).Omit("Versions").Clauses(clause.OnConflict{DoNothing: true}).Create(row)
	if result.Error != nil {
		return false, salesDBError(ctx, result.Error)
	}
	return result.RowsAffected == 1, nil
}
func (r *WorkflowRepository) Publish(ctx context.Context, definition *WorkflowDefinition, version *WorkflowVersion, expected int64) *apperrors.Error {
	result := r.db.WithContext(ctx).Model(&WorkflowDefinition{}).Where("tenant_id = ? AND id = ? AND version = ?", definition.TenantID, definition.ID, expected).Select("name", "version", "updated_at").Updates(definition)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "workflow", "changed; reload before publishing")
	}
	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *WorkflowRepository) WorkflowVersion(ctx context.Context, tenant, id string) (*WorkflowVersion, *apperrors.Error) {
	var row WorkflowVersion
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id).Preload("Steps", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("ordinal") }).First(&row).Error
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *WorkflowRepository) WorkItems(ctx context.Context, tenant string, limit, offset int) ([]WorkItem, int64, *apperrors.Error) {
	rows := []WorkItem{}
	var total int64
	q := r.db.WithContext(ctx).Model(&WorkItem{}).Where("tenant_id = ?", tenant)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	err := q.Preload("Workflow", "tenant_id = ?", tenant).Preload("Workflow.Steps", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("ordinal") }).Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	return rows, total, nil
}
func (r *WorkflowRepository) WorkItem(ctx context.Context, tenant, id string) (*WorkItem, *apperrors.Error) {
	var row WorkItem
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ?", tenant, id).
		Preload("Workflow", "tenant_id = ?", tenant).Preload("Workflow.Steps", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("ordinal") }).
		Preload("Events", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("version DESC").Limit(100) }).First(&row).Error
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *WorkflowRepository) WorkByKey(ctx context.Context, tenant, actor, key string) (*WorkItem, *apperrors.Error) {
	var row WorkItem
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND created_by = ? AND request_key = ?", tenant, actor, key).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	return r.WorkItem(ctx, tenant, row.ID)
}
func (r *WorkflowRepository) SaveWork(ctx context.Context, row *WorkItem, expected int64, create bool) *apperrors.Error {
	q := r.db.WithContext(ctx).Omit("Workflow", "Events")
	if create {
		if err := q.Create(row).Error; err != nil {
			return salesDBError(ctx, err)
		}
		return nil
	}
	result := q.Model(&WorkItem{}).Where("tenant_id = ? AND id = ? AND version = ?", row.TenantID, row.ID, expected).Select("current_step", "responsible_id", "recipient_id", "state", "version", "updated_at").Updates(row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "work item", "changed; reload before retrying")
	}
	return nil
}
func (r *WorkflowRepository) WorkEvent(ctx context.Context, event *WorkEvent) *apperrors.Error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
