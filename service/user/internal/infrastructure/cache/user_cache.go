package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/ecommercehub1/user/internal/core/model"
)

const (
	userKeyPrefix = "user:%s"
	userTTL       = 30 * time.Minute
)

type UserCache struct {
	redis *RedisClient
}

func NewUserCache(redis *RedisClient) *UserCache {
	return &UserCache{redis: redis}
}

func (c *UserCache) SetUser(ctx context.Context, user *model.User) error {
	key := fmt.Sprintf(userKeyPrefix, user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, data, userTTL)
}

func (c *UserCache) GetUser(ctx context.Context, id string) (*model.User, error) {
	key := fmt.Sprintf(userKeyPrefix, id)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCache) DeleteUser(ctx context.Context, id string) error {
	key := fmt.Sprintf(userKeyPrefix, id)
	return c.redis.Delete(ctx, key)
}
