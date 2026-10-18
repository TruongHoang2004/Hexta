package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"go.uber.org/fx"
)

func BuildCache() fx.Option {
	return fx.Provide(
		cache.NewRedisClient,
		func(client *cache.RedisClient) service.IDraftCache { return client },
		cache.NewSessionCache,
		cache.NewTenantCache,
	)
}
