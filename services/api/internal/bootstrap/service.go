package bootstrap

import (
	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"go.uber.org/fx"
)

func BuildService() fx.Option {
	return fx.Provide(
		func() *config.Config { return config.AppConfig },
		service.NewBaseService,
		service.NewAuthService,
		service.NewTenantService,
	)
}
