package bootstrap

import (
	"gitlab.com/ecommercehub1/user/internal/core/service"
	"go.uber.org/fx"
)

func ServiceModule() fx.Option {
	return fx.Options(
		fx.Provide(service.NewUserService),
		fx.Provide(service.NewAddressService),
	)
}
