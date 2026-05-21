package bootstrap

import (
	"gitlab.com/ecommercehub1/order/internal/core/service"
	"go.uber.org/fx"
)

func ServiceModule() fx.Option {
	return fx.Options(
		fx.Provide(func(srv *service.OrderService) service.IOrderService {
			return srv
		}),
		fx.Provide(service.NewOrderService),
	)
}
