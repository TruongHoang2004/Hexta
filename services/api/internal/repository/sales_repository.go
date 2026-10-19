package repository

import (
	"context"
	"encoding/json"
	stderrors "errors"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ISalesRepository keeps transaction and locking details outside business services.
type ISalesRepository interface {
	Transaction(context.Context, func(ISalesRepository) *apperrors.Error) *apperrors.Error
	Member(context.Context, string, string) (*model.TenantMember, *apperrors.Error)
	Tenant(context.Context, string) (*model.Tenant, *apperrors.Error)
	Customer(context.Context, string, string) (*model.Customer, *apperrors.Error)
	Customers(context.Context, string, int, int) ([]model.Customer, int64, *apperrors.Error)
	SaveCustomer(context.Context, *model.Customer, bool) *apperrors.Error
	ArchiveCustomer(context.Context, string, string) *apperrors.Error
	Product(context.Context, string, string) (*model.Product, *apperrors.Error)
	Products(context.Context, string, int, int) ([]model.Product, int64, *apperrors.Error)
	SaveProduct(context.Context, *model.Product, bool) *apperrors.Error
	Variants(context.Context, string, []string) (map[string]model.ProductVariant, *apperrors.Error)
	Order(context.Context, string, string) (*model.Order, *apperrors.Error)
	Orders(context.Context, string, int, int) ([]model.Order, int64, *apperrors.Error)
	SaveOrder(context.Context, *model.Order, int64, bool) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
}

type SalesRepository struct{ db *gorm.DB }

func NewSalesRepository(db *gorm.DB) *SalesRepository { return &SalesRepository{db: db} }

func salesDBError(ctx context.Context, err error) *apperrors.Error {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound(ctx, "record", "not found")
	}
	var pg *pgconn.PgError
	if stderrors.As(err, &pg) && pg.Code == "23505" {
		return apperrors.ErrConflict(ctx, "record", "already exists; refresh and check the submitted values")
	}
	// Never echo SQL/connection details into the API response.
	return apperrors.ErrSystemError(ctx, "sales persistence operation failed")
}
func (r *SalesRepository) Transaction(ctx context.Context, fn func(ISalesRepository) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(&SalesRepository{db: tx})
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
func (r *SalesRepository) Member(ctx context.Context, tenant, user string) (*model.TenantMember, *apperrors.Error) {
	var row model.TenantMember
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("tenant_id = ? AND user_id = ?", tenant, user).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *SalesRepository) Tenant(ctx context.Context, id string) (*model.Tenant, *apperrors.Error) {
	var row model.Tenant
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *SalesRepository) Customer(ctx context.Context, tenant, id string) (*model.Customer, *apperrors.Error) {
	var row model.Customer
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *SalesRepository) Customers(ctx context.Context, tenant string, limit, offset int) ([]model.Customer, int64, *apperrors.Error) {
	rows := []model.Customer{}
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Customer{}).Where("tenant_id = ?", tenant)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	if err := q.Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	return rows, total, nil
}
func (r *SalesRepository) SaveCustomer(ctx context.Context, row *model.Customer, create bool) *apperrors.Error {
	var err error
	if create {
		err = r.db.WithContext(ctx).Create(row).Error
	} else {
		err = r.db.WithContext(ctx).Model(&model.Customer{}).Where("tenant_id = ? AND id = ?", row.TenantID, row.ID).Select("full_name", "phone", "email", "address", "updated_at").Updates(row).Error
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *SalesRepository) ArchiveCustomer(ctx context.Context, tenant, id string) *apperrors.Error {
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id).Delete(&model.Customer{}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *SalesRepository) Product(ctx context.Context, tenant, id string) (*model.Product, *apperrors.Error) {
	var row model.Product
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("created_at, id") }).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *SalesRepository) Products(ctx context.Context, tenant string, limit, offset int) ([]model.Product, int64, *apperrors.Error) {
	rows := []model.Product{}
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Product{}).Where("tenant_id = ? AND status = ?", tenant, model.ProductStatusActive)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	if err := q.Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("created_at, id") }).Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	return rows, total, nil
}
func (r *SalesRepository) SaveProduct(ctx context.Context, row *model.Product, create bool) *apperrors.Error {
	if create {
		if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
			return salesDBError(ctx, err)
		}
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Where("tenant_id = ? AND id = ?", row.TenantID, row.ID).Select("name", "description", "category", "status", "updated_at").Updates(row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if len(row.Variants) > 0 {
		v := &row.Variants[0]
		if err := r.db.WithContext(ctx).Model(&model.ProductVariant{}).Where("tenant_id = ? AND product_id = ? AND id = ?", row.TenantID, row.ID, v.ID).Select("sku", "variant_name", "price", "price_provided", "updated_at").Updates(v).Error; err != nil {
			return salesDBError(ctx, err)
		}
	}

	return nil
}

// Lock active parent products in a consistent order. Catalog writes acquire the
// same product lock before changing variant prices or archiving a product.
func (r *SalesRepository) Variants(ctx context.Context, tenant string, ids []string) (map[string]model.ProductVariant, *apperrors.Error) {
	result := make(map[string]model.ProductVariant)
	if len(ids) == 0 {
		return result, nil
	}
	var rows []model.ProductVariant
	err := r.db.WithContext(ctx).Model(&model.ProductVariant{}).
		Joins("JOIN products ON products.id = product_variants.product_id AND products.tenant_id = product_variants.tenant_id AND products.deleted_at IS NULL").
		Where("product_variants.tenant_id = ? AND product_variants.id IN ? AND products.status = ?", tenant, ids, model.ProductStatusActive).
		Order("products.id, product_variants.id").Clauses(clause.Locking{Strength: "SHARE", Table: clause.Table{Name: "products"}}).Find(&rows).Error
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	for _, row := range rows {
		result[row.ID] = row
	}
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			return nil, apperrors.ErrNotFound(ctx, "product variant", "not found or archived in this workspace")
		}
	}
	return result, nil
}
func (r *SalesRepository) Order(ctx context.Context, tenant, id string) (*model.Order, *apperrors.Error) {
	var row model.Order
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Where("tenant_id = ?", tenant).Order("created_at, id") }).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *SalesRepository) Orders(ctx context.Context, tenant string, limit, offset int) ([]model.Order, int64, *apperrors.Error) {
	rows := []model.Order{}
	var total int64
	q := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ?", tenant)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	if err := q.Preload("Items", "tenant_id = ?", tenant).Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	return rows, total, nil
}
func (r *SalesRepository) SaveOrder(ctx context.Context, row *model.Order, expected int64, create bool) *apperrors.Error {
	if create {
		if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
			return salesDBError(ctx, err)
		}
		return nil
	}
	result := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ? AND id = ? AND version = ? AND status = ?", row.TenantID, row.ID, expected, model.OrderStatusDraft).Select("customer_id", "employee_id", "status", "subtotal", "total_amount", "notes", "version", "requires_delivery", "delivery_address", "updated_at").Updates(row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "order", "changed; reload before saving")
	}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND order_id = ?", row.TenantID, row.ID).Delete(&model.OrderItem{}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if len(row.Items) > 0 {
		if err := r.db.WithContext(ctx).Create(&row.Items).Error; err != nil {
			return salesDBError(ctx, err)
		}
	}
	return nil
}
func (r *SalesRepository) Audit(ctx context.Context, tenant, actor, action, kind, id string, state any) *apperrors.Error {
	data, err := json.Marshal(state)
	if err != nil {
		return apperrors.ErrSystemError(ctx, "could not encode audit state")
	}
	row := model.AuditLog{ID: uuid.NewString(), TenantID: tenant, UserID: actor, Action: action, EntityType: kind, EntityID: id, Source: "web_form", AfterState: string(data)}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
