package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"go.uber.org/fx"
)

func BuildMiddleware() fx.Option {
	return fx.Provide(
		middleware.NewAuthMiddleware,
	)
}
