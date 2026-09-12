package bootstrap

import (
	"context"

	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func BuildDatabase() fx.Option {
	return fx.Module("database",
		fx.Provide(database.NewDatabase),

		fx.Invoke(func(lc fx.Lifecycle, db *gorm.DB, log *zap.SugaredLogger) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Info("Closing database connection")
					sqlDB, err := db.DB()
					if err != nil {
						return err
					}

					return sqlDB.Close()
				},
			})
		}),
	)
}
