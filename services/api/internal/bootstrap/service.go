package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
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
