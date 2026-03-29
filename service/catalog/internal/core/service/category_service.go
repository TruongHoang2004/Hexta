package service

import (
	"context"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/catalog/internal/repository"
	"gitlab.com/ecommercehub1/catalog/internal/utils"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type CategoryService struct {
	*baseService
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		baseService:  NewBaseService(),
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	category.ID = utils.NewULID()
	return s.categoryRepo.Create(ctx, category)
}

func (s *CategoryService) UpdateCategory(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	return s.categoryRepo.Update(ctx, category)
}

func (s *CategoryService) DeleteCategory(ctx context.Context, category *model.Category) (*model.Category, *errors.Error) {
	return s.categoryRepo.Delete(ctx, category)
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, id string) (*model.Category, *errors.Error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *CategoryService) ListAllCategories(ctx context.Context) ([]*model.Category, *errors.Error) {
	return s.categoryRepo.ListAll(ctx)
}
