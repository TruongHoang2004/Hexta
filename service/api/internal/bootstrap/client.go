package bootstrap

import (
	"gitlab.com/ecommercehub1/api/pkg/client"
	"go.uber.org/fx"
)

func BuildClient() fx.Option {
	return fx.Options(
		fx.Provide(client.NewUserClient),
	)
}
