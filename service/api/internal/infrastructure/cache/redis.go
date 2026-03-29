package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/ecommercehub1/api/common/log"
	"gitlab.com/ecommercehub1/api/config"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	addr := config.AppConfig.Redis.Addr
	log.Info(context.Background(), "🔌 Connecting to Redis at %s", addr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		PoolSize: 10,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	log.Info(context.Background(), "✅ Redis connected")
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	return r.Client.Get(ctx, key).Bytes()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
