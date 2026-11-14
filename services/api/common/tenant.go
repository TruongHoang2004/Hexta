package common

import (
	"context"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
)

type tenantContextKey struct{}

func SetTenant(ctx context.Context, tenant *model.Tenant) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

func GetTenant(ctx context.Context) (*model.Tenant, bool) {
	tenant, ok := ctx.Value(tenantContextKey{}).(*model.Tenant)
	return tenant, ok
}
