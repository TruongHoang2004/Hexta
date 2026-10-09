package repository

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
)

var (
	ErrInsufficientStock         = stdErrors.New("insufficient available stock")
	ErrInvalidReservedQuantity   = stdErrors.New("cannot release or deduct more than currently reserved stock")
	ErrAdjustmentNegativeBalance = stdErrors.New("stock adjustment cannot result in negative available quantity")
)

type IInventoryRepository interface {
	GetByVariantID(ctx context.Context, tenantID, variantID string) (*model.InventoryItem, *errors.Error)
	GetByID(ctx context.Context, tenantID, id string) (*model.InventoryItem, *errors.Error)
	CreateInventoryItem(ctx context.Context, item *model.InventoryItem) (*model.InventoryItem, *errors.Error)
	ReserveStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error)
	ReleaseStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error)
	DeductStock(ctx context.Context, tenantID, variantID string, quantity int, refType string, refID *string) (*model.InventoryItem, *errors.Error)
	AdjustStock(ctx context.Context, tenantID, variantID string, newOnHand int, notes string) (*model.InventoryItem, *errors.Error)
	RecordMovement(ctx context.Context, movement *model.StockMovement) (*model.StockMovement, *errors.Error)
	ListMovements(ctx context.Context, tenantID, inventoryID string, limit, offset int) ([]*model.StockMovement, int64, *errors.Error)
}

type InventoryRepository struct {
	*baseRepository
}

func NewInventoryDBRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{baseRepository: NewBaseRepository(db)}
}

func (r *InventoryRepository) GetByVariantID(ctx context.Context, tenantID, variantID string) (*model.InventoryItem, *errors.Error) {
	var item model.InventoryItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).
		First(&item).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}
	return &item, nil
}

func (r *InventoryRepository) GetByID(ctx context.Context, tenantID, id string) (*model.InventoryItem, *errors.Error) {
	var item model.InventoryItem
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&item).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}
	return &item, nil
}

func (r *InventoryRepository) CreateInventoryItem(ctx context.Context, item *model.InventoryItem) (*model.InventoryItem, *errors.Error) {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if item.AvailableQty == 0 && item.OnHandQty > 0 && item.ReservedQty == 0 {
		item.AvailableQty = item.OnHandQty
	}

	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		return nil, r.returnError(ctx, err)
	}
	return item, nil
}

func (r *InventoryRepository) ReserveStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	var item model.InventoryItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.InventoryItem{}).
			Where("tenant_id = ? AND variant_id = ? AND available_qty >= ?", tenantID, variantID, quantity).
			Updates(map[string]interface{}{
				"reserved_qty":  gorm.Expr("reserved_qty + ?", quantity),
				"available_qty": gorm.Expr("available_qty - ?", quantity),
				"updated_at":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var exists model.InventoryItem
			if checkErr := tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&exists).Error; checkErr != nil {
				return checkErr
			}
			return ErrInsufficientStock
		}
		return tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&item).Error
	})

	if err != nil {
		if stdErrors.Is(err, ErrInsufficientStock) {
			return nil, errors.ErrConflict(ctx, "inventory stock", "insufficient available quantity to reserve")
		}
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}

	return &item, nil
}

func (r *InventoryRepository) ReleaseStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	var item model.InventoryItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.InventoryItem{}).
			Where("tenant_id = ? AND variant_id = ? AND reserved_qty >= ?", tenantID, variantID, quantity).
			Updates(map[string]interface{}{
				"reserved_qty":  gorm.Expr("reserved_qty - ?", quantity),
				"available_qty": gorm.Expr("available_qty + ?", quantity),
				"updated_at":    time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var exists model.InventoryItem
			if checkErr := tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&exists).Error; checkErr != nil {
				return checkErr
			}
			return ErrInvalidReservedQuantity
		}
		return tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&item).Error
	})

	if err != nil {
		if stdErrors.Is(err, ErrInvalidReservedQuantity) {
			return nil, errors.ErrBadRequest(ctx).SetDetail("release quantity exceeds currently reserved quantity")
		}
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}

	return &item, nil
}

func (r *InventoryRepository) DeductStock(ctx context.Context, tenantID, variantID string, quantity int, refType string, refID *string) (*model.InventoryItem, *errors.Error) {
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	var item model.InventoryItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.InventoryItem{}).
			Where("tenant_id = ? AND variant_id = ? AND reserved_qty >= ? AND on_hand_qty >= ?", tenantID, variantID, quantity, quantity).
			Updates(map[string]interface{}{
				"on_hand_qty":  gorm.Expr("on_hand_qty - ?", quantity),
				"reserved_qty": gorm.Expr("reserved_qty - ?", quantity),
				"updated_at":   time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			var exists model.InventoryItem
			if checkErr := tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&exists).Error; checkErr != nil {
				return checkErr
			}
			return ErrInvalidReservedQuantity
		}

		if err := tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&item).Error; err != nil {
			return err
		}

		// Log immutable stock movement record
		mv := &model.StockMovement{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			InventoryID:   item.ID,
			MovementType:  model.MovementOutbound,
			Quantity:      quantity,
			BalanceAfter:  item.OnHandQty,
			ReferenceType: refType,
			ReferenceID:   refID,
			CreatedAt:     time.Now(),
		}
		return tx.Create(mv).Error
	})

	if err != nil {
		if stdErrors.Is(err, ErrInvalidReservedQuantity) {
			return nil, errors.ErrBadRequest(ctx).SetDetail("deduction quantity exceeds reserved or on-hand stock")
		}
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}

	return &item, nil
}

func (r *InventoryRepository) AdjustStock(ctx context.Context, tenantID, variantID string, newOnHand int, notes string) (*model.InventoryItem, *errors.Error) {
	if newOnHand < 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("new on-hand quantity cannot be negative")
	}

	var item model.InventoryItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.InventoryItem
		if err := tx.Where("tenant_id = ? AND variant_id = ?", tenantID, variantID).First(&current).Error; err != nil {
			return err
		}

		newAvailable := newOnHand - current.ReservedQty
		if newAvailable < 0 {
			return ErrAdjustmentNegativeBalance
		}

		diff := newOnHand - current.OnHandQty
		if err := tx.Model(&current).Updates(map[string]interface{}{
			"on_hand_qty":   newOnHand,
			"available_qty": newAvailable,
			"updated_at":    time.Now(),
		}).Error; err != nil {
			return err
		}

		item = current
		item.OnHandQty = newOnHand
		item.AvailableQty = newAvailable

		mv := &model.StockMovement{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			InventoryID:  item.ID,
			MovementType: model.MovementAdjustment,
			Quantity:     diff,
			BalanceAfter: newOnHand,
			Notes:        notes,
			CreatedAt:    time.Now(),
		}
		return tx.Create(mv).Error
	})

	if err != nil {
		if stdErrors.Is(err, ErrAdjustmentNegativeBalance) {
			return nil, errors.ErrConflict(ctx, "inventory stock", "new on-hand stock cannot be less than current reserved stock")
		}
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
		}
		return nil, r.returnError(ctx, err)
	}

	return &item, nil
}

func (r *InventoryRepository) RecordMovement(ctx context.Context, movement *model.StockMovement) (*model.StockMovement, *errors.Error) {
	if movement.ID == "" {
		movement.ID = uuid.New().String()
	}
	err := r.db.WithContext(ctx).Create(movement).Error
	if err != nil {
		return nil, r.returnError(ctx, err)
	}
	return movement, nil
}

func (r *InventoryRepository) ListMovements(ctx context.Context, tenantID, inventoryID string, limit, offset int) ([]*model.StockMovement, int64, *errors.Error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var movements []*model.StockMovement
	var total int64

	query := r.db.WithContext(ctx).Model(&model.StockMovement{}).
		Where("tenant_id = ? AND inventory_id = ?", tenantID, inventoryID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&movements).Error
	if err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	return movements, total, nil
}
