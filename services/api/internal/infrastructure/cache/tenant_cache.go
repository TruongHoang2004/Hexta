package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
)

const (
	tenantIDKeyPrefix   = "tenant:id:%s"
	tenantSlugKeyPrefix = "tenant:slug:%s"
	tenantTTL           = 24 * time.Hour
)

type TenantCache struct {
	redis *RedisClient
}

func NewTenantCache(redis *RedisClient) *TenantCache {
	return &TenantCache{redis: redis}
}

func (c *TenantCache) SetTenant(ctx context.Context, tenant *model.Tenant) error {
	if tenant == nil || tenant.ID == "" {
		return nil
	}
	data, err := json.Marshal(tenant)
	if err != nil {
		return err
	}
	idKey := fmt.Sprintf(tenantIDKeyPrefix, tenant.ID)
	if err := c.redis.Set(ctx, idKey, data, tenantTTL); err != nil {
		return err
	}
	if tenant.Slug != "" {
		slugKey := fmt.Sprintf(tenantSlugKeyPrefix, tenant.Slug)
		return c.redis.Set(ctx, slugKey, data, tenantTTL)
	}
	return nil
}

func (c *TenantCache) GetTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	key := fmt.Sprintf(tenantIDKeyPrefix, id)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var tenant model.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (c *TenantCache) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	key := fmt.Sprintf(tenantSlugKeyPrefix, slug)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var tenant model.Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (c *TenantCache) DeleteTenant(ctx context.Context, id string, slug string) error {
	if id != "" {
		_ = c.redis.Delete(ctx, fmt.Sprintf(tenantIDKeyPrefix, id))
	}
	if slug != "" {
		_ = c.redis.Delete(ctx, fmt.Sprintf(tenantSlugKeyPrefix, slug))
	}
	return nil
}
