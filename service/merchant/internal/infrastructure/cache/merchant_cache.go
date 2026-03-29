package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/ecommercehub1/merchant/internal/core/model"
)

const (
	merchantKeyPrefix = "merchant:%s"
	merchantTTL       = 30 * time.Minute
)

type MerchantCache struct {
	redis *RedisClient
}

func NewMerchantCache(redis *RedisClient) *MerchantCache {
	return &MerchantCache{redis: redis}
}

func (c *MerchantCache) SetMerchant(ctx context.Context, merchant *model.Merchant) error {
	key := fmt.Sprintf(merchantKeyPrefix, merchant.ID)
	data, err := json.Marshal(merchant)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, data, merchantTTL)
}

func (c *MerchantCache) GetMerchant(ctx context.Context, id string) (*model.Merchant, error) {
	key := fmt.Sprintf(merchantKeyPrefix, id)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var merchant model.Merchant
	if err := json.Unmarshal(data, &merchant); err != nil {
		return nil, err
	}

	return &merchant, nil
}

func (c *MerchantCache) DeleteMerchant(ctx context.Context, id string) error {
	key := fmt.Sprintf(merchantKeyPrefix, id)
	return c.redis.Delete(ctx, key)
}
