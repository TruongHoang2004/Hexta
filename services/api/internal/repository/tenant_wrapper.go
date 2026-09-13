package repository

import (
	"context"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type TenantCacheWrapper struct {
	dbRepo *TenantRepository
	cache  *cache.TenantCache
}

func NewTenantCacheWrapper(dbRepo *TenantRepository, ch *cache.TenantCache) ITenantRepository {
	return &TenantCacheWrapper{
		dbRepo: dbRepo,
		cache:  ch,
	}
}

func (w *TenantCacheWrapper) CreateTenant(ctx context.Context, tenant *model.Tenant, ownerID string) (*model.Tenant, *errors.Error) {
	createdTenant, err := w.dbRepo.CreateTenant(ctx, tenant, ownerID)
	if err == nil && createdTenant != nil {
		_ = w.cache.SetTenant(ctx, createdTenant)
	}
	return createdTenant, err
}

func (w *TenantCacheWrapper) GetTenantByID(ctx context.Context, id string) (*model.Tenant, *errors.Error) {
	if cached, err := w.cache.GetTenantByID(ctx, id); err == nil && cached != nil {
		return cached, nil
	}
	tenant, err := w.dbRepo.GetTenantByID(ctx, id)
	if err == nil && tenant != nil {
		_ = w.cache.SetTenant(ctx, tenant)
	}
	return tenant, err
}

func (w *TenantCacheWrapper) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, *errors.Error) {
	if cached, err := w.cache.GetTenantBySlug(ctx, slug); err == nil && cached != nil {
		return cached, nil
	}
	tenant, err := w.dbRepo.GetTenantBySlug(ctx, slug)
	if err == nil && tenant != nil {
		_ = w.cache.SetTenant(ctx, tenant)
	}
	return tenant, err
}

func (w *TenantCacheWrapper) UpdateTenant(ctx context.Context, tenant *model.Tenant) (*model.Tenant, *errors.Error) {
	updatedTenant, err := w.dbRepo.UpdateTenant(ctx, tenant)
	if err == nil && updatedTenant != nil {
		_ = w.cache.DeleteTenant(ctx, updatedTenant.ID, updatedTenant.Slug)
		_ = w.cache.SetTenant(ctx, updatedTenant)
	}
	return updatedTenant, err
}

func (w *TenantCacheWrapper) GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error) {
	return w.dbRepo.GetUserTenants(ctx, userID)
}

func (w *TenantCacheWrapper) GetTenantMember(ctx context.Context, tenantID, userID string) (*model.TenantMember, *errors.Error) {
	return w.dbRepo.GetTenantMember(ctx, tenantID, userID)
}

func (w *TenantCacheWrapper) ListTenantMembers(ctx context.Context, tenantID string) ([]*model.TenantMember, *errors.Error) {
	return w.dbRepo.ListTenantMembers(ctx, tenantID)
}

func (w *TenantCacheWrapper) AddTenantMember(ctx context.Context, member *model.TenantMember) (*model.TenantMember, *errors.Error) {
	return w.dbRepo.AddTenantMember(ctx, member)
}
