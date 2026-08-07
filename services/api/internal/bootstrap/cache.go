package bootstrap

import (
	"gitlab.com/ecommercehub1/api/internal/infrastructure/cache"
	"go.uber.org/fx"
)

func BuildCache() fx.Option {
	return fx.Provide(
		cache.NewRedisClient,
		cache.NewSessionCache,
	)
}
