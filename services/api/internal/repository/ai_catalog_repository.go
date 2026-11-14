package repository

import (
	"context"
	stderrors "errors"
	"strings"
	"unicode"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		terms := productSearchTerms(query)
		if terms == "" {
			return rows, nil
		}
		// The GIN-indexed search_vector covers name, category and description; prefixes keep partial words matching.
		q = q.Where("search_vector @@ to_tsquery('simple', ?)", terms).
			Order(clause.Expr{SQL: "ts_rank(search_vector, to_tsquery('simple', ?)) DESC, id", Vars: []interface{}{terms}})
	}
	if err := q.Limit(10).Find(&rows).Error; err != nil {
		return nil, aiCatalogError(ctx)
	}
	return rows, nil
}

// productSearchTerms turns free text into an AND of prefix terms. Tokens keep only letters, marks and digits,
// so caller input cannot inject tsquery operators.
func productSearchTerms(query string) string {
	words := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsMark(r) && !unicode.IsDigit(r)
	})
	for i, word := range words {
		words[i] = word + ":*"
	}
	return strings.Join(words, " & ")
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
