package repository

import (
	"context"
	"fmt"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/catalog/internal/present/request"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gorm.io/gorm"
)

type ProductRepository struct {
	*baseRepository
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	if err := r.db.Create(product).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return product, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	if err := r.db.Save(product).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	if err := r.db.Delete(product).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return product, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*model.Product, *errors.Error) {
	var product model.Product
	if err := r.db.First(&product, "id = ?", id).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, req *request.PaginationRequest) ([]*model.Product, int64, *errors.Error) {
	var products []*model.Product
	var total int64
	offset := (req.Page - 1) * req.Limit
	query := r.db.Model(&model.Product{})

	for field, value := range req.Filter {
		query = query.Where(field+" = ?", value)
	}

	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	if req.Sort != "" && req.Order != "" {
		query = query.Order(req.Sort + " " + req.Order)
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Offset(int(offset)).Limit(int(req.Limit)).Find(&products).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	return products, total, nil
}

func (r *ProductRepository) GetByIDs(ctx context.Context, ids []string) ([]*model.Product, *errors.Error) {
	var products []*model.Product
	if err := r.db.Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return products, nil
}

func (r *ProductRepository) IsFieldExist(ctx context.Context, fields []string, values []string) (string, bool, *errors.Error) {
	var result map[string]interface{}

	query := r.db.Model(&model.Product{})
	for i := range fields {
		if values[i] != "" {
			query = query.Or(fmt.Sprintf("%s = ?", fields[i]), values[i])
		}
	}

	if err := query.First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, r.returnError(ctx, err)
	}

	for i, field := range fields {
		if values[i] != "" {
			if val, ok := result[field]; ok {
				if fmt.Sprintf("%v", val) == values[i] {
					return field, true, nil
				}
			}
		}
	}

	return "", true, nil
}
