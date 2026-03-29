package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
)

const (
	productKeyPrefix = "product:%s"
	productTTL       = 30 * time.Minute
)

type ProductCache struct {
	redis *RedisClient
}

func NewProductCache(redis *RedisClient) *ProductCache {
	return &ProductCache{redis: redis}
}

func (c *ProductCache) SetProduct(ctx context.Context, product *model.Product) error {
	key := fmt.Sprintf(productKeyPrefix, product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, productTTL)
}

func (c *ProductCache) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	key := fmt.Sprintf(productKeyPrefix, id)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var product model.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (c *ProductCache) DeleteProduct(ctx context.Context, id string) error {
	key := fmt.Sprintf(productKeyPrefix, id)
	return c.redis.Delete(ctx, key)
}
