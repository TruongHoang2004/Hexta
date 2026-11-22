package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
)

func BuildRepository() fx.Option {
	return fx.Provide(
		fx.Annotate(repository.NewAICatalogRepository, fx.As(new(repository.IAICatalogRepository))),
		repository.NewTenantOnboardingRepository,
		repository.NewWorkflowRepository,
		repository.NewMessagingRepository,
		func(r *repository.MessagingRepository) repository.MessagingStore { return r },
		func(r *repository.MessagingRepository) repository.MessagingQueryStore { return r },
		fx.Annotate(repository.NewChannelRepository, fx.As(new(repository.ChannelStore))),
		repository.NewBaseRepository,
		fx.Annotate(repository.NewSalesRepository, fx.As(new(repository.ISalesRepository))),
		fx.Annotate(repository.NewIdentityRepository, fx.As(new(repository.IIdentityRepository))),
		// Session validity is an auth decision and must stay DB-authoritative (SEC-001, CACHE-002, CACHE-009).
		fx.Annotate(repository.NewSessionDBRepository, fx.As(new(repository.ISessionRepository))),
		repository.NewTenantDBRepository,
		repository.NewTenantCacheWrapper,
		fx.Annotate(
			repository.NewInventoryDBRepository,
			fx.As(new(repository.IInventoryRepository)),
		),
		fx.Annotate(
			repository.NewOrderDBRepository,
			fx.As(new(repository.IOrderRepository)),
		),
	)
}
