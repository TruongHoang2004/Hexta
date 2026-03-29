package bootstrap

import (
	"gitlab.com/ecommercehub1/api/internal/present/http/middleware"
	"go.uber.org/fx"
)

func BuildMiddleware() fx.Option {
	return fx.Options(
		fx.Provide(middleware.NewAuthMiddleware),
	)
}
