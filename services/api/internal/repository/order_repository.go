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

type IOrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order, items []model.OrderItem) (*model.Order, *errors.Error)
	GetByID(ctx context.Context, tenantID, orderID string) (*model.Order, *errors.Error)
	GetByOrderNumber(ctx context.Context, tenantID, orderNumber string) (*model.Order, *errors.Error)
	ListOrders(ctx context.Context, tenantID string, status *model.OrderStatus, limit, offset int) ([]*model.Order, int64, *errors.Error)
	UpdateOrderStatus(ctx context.Context, tenantID, orderID string, status model.OrderStatus) (*model.Order, *errors.Error)
	UpdatePaymentStatus(ctx context.Context, tenantID, orderID string, paymentStatus model.PaymentStatus) (*model.Order, *errors.Error)
}

type OrderRepository struct {
	*baseRepository
}

func NewOrderDBRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{baseRepository: NewBaseRepository(db)}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *model.Order, items []model.OrderItem) (*model.Order, *errors.Error) {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range items {
			if items[i].ID == "" {
				items[i].ID = uuid.New().String()
			}
			items[i].OrderID = order.ID
			items[i].TenantID = order.TenantID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, r.returnError(ctx, err)
	}

	order.Items = items
	return order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, tenantID, orderID string) (*model.Order, *errors.Error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("tenant_id = ? AND id = ?", tenantID, orderID).
		First(&order).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "order", "not found")
		}
		return nil, r.returnError(ctx, err)
	}
	return &order, nil
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, tenantID, orderNumber string) (*model.Order, *errors.Error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("tenant_id = ? AND order_number = ?", tenantID, orderNumber).
		First(&order).Error
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotFound(ctx, "order", "not found")
		}
		return nil, r.returnError(ctx, err)
	}
	return &order, nil
}

func (r *OrderRepository) ListOrders(ctx context.Context, tenantID string, status *model.OrderStatus, limit, offset int) ([]*model.Order, int64, *errors.Error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var orders []*model.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Order{}).Where("tenant_id = ?", tenantID)
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	err := query.Preload("Items").Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error
	if err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	return orders, total, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, tenantID, orderID string, status model.OrderStatus) (*model.Order, *errors.Error) {
	res := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND id = ?", tenantID, orderID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return nil, r.returnError(ctx, res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, errors.ErrNotFound(ctx, "order", "not found")
	}

	return r.GetByID(ctx, tenantID, orderID)
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, tenantID, orderID string, paymentStatus model.PaymentStatus) (*model.Order, *errors.Error) {
	res := r.db.WithContext(ctx).Model(&model.Order{}).
		Where("tenant_id = ? AND id = ?", tenantID, orderID).
		Updates(map[string]interface{}{
			"payment_status": paymentStatus,
			"updated_at":     time.Now(),
		})
	if res.Error != nil {
		return nil, r.returnError(ctx, res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, errors.ErrNotFound(ctx, "order", "not found")
	}

	return r.GetByID(ctx, tenantID, orderID)
}
