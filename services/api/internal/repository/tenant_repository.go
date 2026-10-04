package repository

import (
	"context"
	stdErrors "errors"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
)

type ITenantRepository interface {
	CreateTenant(ctx context.Context, tenant *model.Tenant, ownerID string) (*model.Tenant, *errors.Error)
	GetTenantByID(ctx context.Context, id string) (*model.Tenant, *errors.Error)
	GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, *errors.Error)
	UpdateTenant(ctx context.Context, tenant *model.Tenant) (*model.Tenant, *errors.Error)
	GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error)
	GetTenantMember(ctx context.Context, tenantID, userID string) (*model.TenantMember, *errors.Error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]*model.TenantMember, *errors.Error)
	AddTenantMember(ctx context.Context, member *model.TenantMember) (*model.TenantMember, *errors.Error)
}

type TenantRepository struct {
	*baseRepository
}

func NewTenantDBRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{baseRepository: NewBaseRepository(db)}
}

func (r *TenantRepository) CreateTenant(ctx context.Context, tenant *model.Tenant, ownerID string) (*model.Tenant, *errors.Error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}

		ownerMember := &model.TenantMember{
			TenantID: tenant.ID,
			UserID:   ownerID,
			Role:     model.RoleOwner,
		}

		if err := tx.Create(ownerMember).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, r.returnError(ctx, err)
	}

	return tenant, nil
}

func (r *TenantRepository) GetTenantByID(ctx context.Context, id string) (*model.Tenant, *errors.Error) {
	var tenant model.Tenant
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, r.returnError(ctx, err)
	}
	return &tenant, nil
}

func (r *TenantRepository) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, *errors.Error) {
	var tenant model.Tenant
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error; err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, r.returnError(ctx, err)
	}
	return &tenant, nil
}

func (r *TenantRepository) UpdateTenant(ctx context.Context, tenant *model.Tenant) (*model.Tenant, *errors.Error) {
	if err := r.db.WithContext(ctx).Save(tenant).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return tenant, nil
}

func (r *TenantRepository) GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error) {
	var tenants []*model.Tenant
	err := r.db.WithContext(ctx).
		Joins("JOIN tenant_members ON tenant_members.tenant_id = tenants.id").
		Where("tenant_members.user_id = ?", userID).
		Order("tenants.created_at DESC").
		Find(&tenants).Error

	if err != nil {
		return nil, r.returnError(ctx, err)
	}
	return tenants, nil
}

func (r *TenantRepository) GetTenantMember(ctx context.Context, tenantID, userID string) (*model.TenantMember, *errors.Error) {
	var member model.TenantMember
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&member).Error; err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, r.returnError(ctx, err)
	}
	return &member, nil
}

func (r *TenantRepository) ListTenantMembers(ctx context.Context, tenantID string) ([]*model.TenantMember, *errors.Error) {
	var members []*model.TenantMember
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at ASC").Find(&members).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return members, nil
}

func (r *TenantRepository) AddTenantMember(ctx context.Context, member *model.TenantMember) (*model.TenantMember, *errors.Error) {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return member, nil
}
