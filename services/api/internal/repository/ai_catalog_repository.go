package repository

import (
	"context"
	stderrors "errors"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
)

// IAICatalogRepository exposes only tenant-qualified reads to the legacy agent.
// It intentionally cannot create orders or mutate stock/customer records.
type IAICatalogRepository interface {
	SearchProducts(context.Context, string, string) ([]model.Product, *apperrors.Error)
	StockBySKU(context.Context, string, string) (*model.ProductVariant, *model.InventoryItem, *apperrors.Error)
	LookupCustomers(context.Context, string, string, string) ([]model.Customer, *apperrors.Error)
}

type AICatalogRepository struct{ db *gorm.DB }

func NewAICatalogRepository(db *gorm.DB) *AICatalogRepository { return &AICatalogRepository{db: db} }

func aiCatalogError(ctx context.Context) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "AI catalog persistence operation failed")
}

func (r *AICatalogRepository) SearchProducts(ctx context.Context, tenantID, query string) ([]model.Product, *apperrors.Error) {
	var rows []model.Product
	q := r.db.WithContext(ctx).Preload("Variants", "tenant_id = ?", tenantID).
		Where("tenant_id = ? AND status = ?", tenantID, model.ProductStatusActive)
	if query != "" {
		q = q.Where("name LIKE ? OR category LIKE ?", "%"+query+"%", "%"+query+"%")
	}
	if err := q.Limit(10).Find(&rows).Error; err != nil {
		return nil, aiCatalogError(ctx)
	}
	return rows, nil
}

func (r *AICatalogRepository) StockBySKU(ctx context.Context, tenantID, sku string) (*model.ProductVariant, *model.InventoryItem, *apperrors.Error) {
	var variant model.ProductVariant
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND sku = ?", tenantID, sku).First(&variant).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, aiCatalogError(ctx)
	}
	var inventory model.InventoryItem
	err = r.db.WithContext(ctx).Where("tenant_id = ? AND variant_id = ?", tenantID, variant.ID).First(&inventory).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return &variant, nil, nil
	}
	if err != nil {
		return nil, nil, aiCatalogError(ctx)
	}
	return &variant, &inventory, nil
}

func (r *AICatalogRepository) LookupCustomers(ctx context.Context, tenantID, phone, name string) ([]model.Customer, *apperrors.Error) {
	var rows []model.Customer
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if phone != "" {
		q = q.Where("phone LIKE ?", "%"+phone+"%")
	}
	if name != "" {
		q = q.Where("full_name LIKE ?", "%"+name+"%")
	}
	if err := q.Limit(5).Find(&rows).Error; err != nil {
		return nil, aiCatalogError(ctx)
	}
	return rows, nil
}
