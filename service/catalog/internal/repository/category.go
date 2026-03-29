package repository

import (
	"context"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	*baseRepository
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	if err := r.db.Create(category).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return category, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	if err := r.db.Save(category).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return category, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	if err := r.db.Delete(category).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return category, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*model.Category, *errors.Error) {
	var category model.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &category, nil
}

func (r *CategoryRepository) ListAll(ctx context.Context) ([]*model.Category, *errors.Error) {
	var categories []*model.Category
	if err := r.db.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return categories, nil
}
