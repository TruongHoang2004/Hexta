package service

import (
	"context"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/catalog/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/catalog/internal/present/request"
	"gitlab.com/ecommercehub1/catalog/internal/repository"
	"gitlab.com/ecommercehub1/catalog/internal/utils"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type ProductService struct {
	*baseService
	productRepo  *repository.ProductRepository
	productCache *cache.ProductCache
}

func NewProductService(productRepo *repository.ProductRepository, productCache *cache.ProductCache) *ProductService {
	return &ProductService{
		baseService:  NewBaseService(),
		productRepo:  productRepo,
		productCache: productCache,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	product.ID = utils.NewULID()
	conflictField, exists, err := s.productRepo.IsFieldExist(ctx, []string{"name"}, []string{product.Name})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrConflict(ctx, conflictField, "already exists")
	}
	return s.productRepo.Create(ctx, product)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	updatedProduct, err := s.productRepo.Update(ctx, product)
	if err == nil {
		_ = s.productCache.DeleteProduct(ctx, product.ID)
	}
	return updatedProduct, err
}

func (s *ProductService) DeleteProduct(ctx context.Context, product *model.Product) (*model.Product, *errors.Error) {
	deletedProduct, err := s.productRepo.Delete(ctx, product)
	if err == nil {
		_ = s.productCache.DeleteProduct(ctx, product.ID)
	}
	return deletedProduct, err
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*model.Product, *errors.Error) {
	if product, err := s.productCache.GetProduct(ctx, id); err == nil {
		return product, nil
	}

	product, err := s.productRepo.GetByID(ctx, id)
	if err == nil {
		_ = s.productCache.SetProduct(ctx, product)
	}
	return product, err
}

func (s *ProductService) ListProducts(ctx context.Context, req *request.PaginationRequest) ([]*model.Product, int64, *errors.Error) {
	return s.productRepo.List(ctx, req)
}

func (s *ProductService) GetProductsByIDs(ctx context.Context, ids []string) ([]*model.Product, *errors.Error) {
	return s.productRepo.GetByIDs(ctx, ids)
}
