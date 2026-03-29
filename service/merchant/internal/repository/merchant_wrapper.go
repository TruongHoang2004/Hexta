package repository

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/merchant/internal/core/model"
	"gitlab.com/ecommercehub1/merchant/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/merchant/internal/present/request"
)

type MerchantCacheWrapper struct {
	dbRepo *MerchantRepository
	cache  *cache.MerchantCache
}

func NewMerchantCacheWrapper(dbRepo *MerchantRepository, ch *cache.MerchantCache) IMerchantRepository {
	return &MerchantCacheWrapper{
		dbRepo: dbRepo,
		cache:  ch,
	}
}

func (w *MerchantCacheWrapper) IsExited(ctx context.Context, name, email, phone string) (bool, *errors.Error) {
	return w.dbRepo.IsExited(ctx, name, email, phone)
}

func (w *MerchantCacheWrapper) Create(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	created, err := w.dbRepo.Create(ctx, merchant)
	if err == nil {
		_ = w.cache.SetMerchant(ctx, created)
	}
	return created, err
}

func (w *MerchantCacheWrapper) Update(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	updated, err := w.dbRepo.Update(ctx, merchant)
	if err == nil {
		_ = w.cache.DeleteMerchant(ctx, merchant.ID)
	}
	return updated, err
}

func (w *MerchantCacheWrapper) Delete(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	deleted, err := w.dbRepo.Delete(ctx, merchant)
	if err == nil {
		_ = w.cache.DeleteMerchant(ctx, merchant.ID)
	}
	return deleted, err
}

func (w *MerchantCacheWrapper) GetByID(ctx context.Context, id string) (*model.Merchant, *errors.Error) {
	if cached, err := w.cache.GetMerchant(ctx, id); err == nil && cached != nil {
		return cached, nil
	}
	merchant, err := w.dbRepo.GetByID(ctx, id)
	if err == nil {
		_ = w.cache.SetMerchant(ctx, merchant)
	}
	return merchant, err
}

func (w *MerchantCacheWrapper) GetByOwnerID(ctx context.Context, ownerID string) (*model.Merchant, *errors.Error) {
	return w.dbRepo.GetByOwnerID(ctx, ownerID)
}

func (w *MerchantCacheWrapper) List(ctx context.Context, req *request.PaginationRequest) ([]*model.Merchant, int64, *errors.Error) {
	return w.dbRepo.List(ctx, req)
}
