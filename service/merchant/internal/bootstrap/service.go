package bootstrap

import (
	"gitlab.com/ecommercehub1/merchant/internal/core/service"
	"go.uber.org/fx"
)

func ServiceModule() fx.Option {
	return fx.Options(
		fx.Provide(service.NewMerchantService),
	)
}
