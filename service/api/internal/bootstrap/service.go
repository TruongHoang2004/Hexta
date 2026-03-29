package bootstrap

import (
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"go.uber.org/fx"
)

func BuildService() fx.Option {
	return fx.Provide(
		service.NewBaseService,
		service.NewAuthService,
		service.NewUserService,
		service.NewMerchantService,
	)
}
