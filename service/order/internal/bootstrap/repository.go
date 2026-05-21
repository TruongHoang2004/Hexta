package bootstrap
import (
	"gitlab.com/ecommercehub1/order/internal/repository"
	"go.uber.org/fx"
)

func RepositoryModule() fx.Option {
	return fx.Options(
		fx.Provide(func(repo *repository.OrderRepository) repository.IOrderRepository {
			return repo
		}),
		fx.Provide(repository.NewOrderDBRepository),
	)
}
