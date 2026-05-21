package repository

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/order/internal/core/model"
	"gorm.io/gorm"
)

type IOrderRepository interface {
	Create(ctx context.Context, order *model.Order) (*model.Order, *errors.Error)
}

type OrderRepository struct {
	*baseRepository
}

func NewOrderDBRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) (*model.Order, *errors.Error) {
	if err := r.db.Create(order).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return order, nil
}
