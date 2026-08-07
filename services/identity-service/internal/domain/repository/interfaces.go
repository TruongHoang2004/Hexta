package repository

import (
	"context"

	"gitlab.com/ecommercehub1/identity-service/internal/domain/entity"
	"gitlab.com/ecommercehub1/identity-service/internal/domain/valueobject"
)

type TenantRepository interface {
	Save(ctx context.Context, tenant *entity.Tenant) error
	FindByID(ctx context.Context, id valueobject.TenantID) (*entity.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
}

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id valueobject.UserID) (*entity.User, error)
	FindByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}

type RoleRepository interface {
	Save(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id valueobject.RoleID) (*entity.Role, error)
	FindByTenantID(ctx context.Context, tenantID valueobject.TenantID) ([]*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
}
