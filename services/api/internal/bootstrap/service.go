package bootstrap

import (
	"strings"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/connector"
	"go.uber.org/fx"
)

func BuildService() fx.Option {
	return fx.Provide(
		service.NewTenantOnboardingService,
		func() *config.Config { return config.AppConfig },
		service.NewBaseService,
		service.NewSalesService,
		service.NewWorkflowService,
		service.NewMessagingQueryService,
		service.NewMessagingCommandService,
		service.NewChannelService,
		newChannelSettings,
		fx.Annotate(
			connector.NewClient,
			fx.As(new(service.IChannelConnector)),
		),
		service.NewHealthService,
		service.NewAuthService,
		service.NewTenantService,
		fx.Annotate(
			service.NewInventoryService,
			fx.As(new(service.IInventoryService)),
		),
		fx.Annotate(
			service.NewOrderService,
			fx.As(new(service.IOrderService)),
		),
		fx.Annotate(
			ai.NewConsultClient,
			fx.As(new(service.IConsultClient)),
		),
		fx.Annotate(
			service.NewConsultantService,
			fx.As(new(service.IAIAgentService)),
		),
	)
}

// newChannelSettings allows OAuth redirects only to the configured frontend origin (plus localhost outside production).
func newChannelSettings(cfg *config.Config) service.ChannelSettings {
	settings := service.ChannelSettings{AppID: cfg.Facebook.AppID, GraphVersion: cfg.Facebook.GraphVersion, DialogURL: cfg.Facebook.DialogURL}
	if origin := strings.TrimRight(cfg.Server.FrontendURL, "/"); origin != "" {
		settings.AllowedOrigins = append(settings.AllowedOrigins, origin)
	}
	if !cfg.Server.Production {
		settings.AllowedOrigins = append(settings.AllowedOrigins, "http://localhost:3000", "http://127.0.0.1:3000")
	}
	return settings
}
