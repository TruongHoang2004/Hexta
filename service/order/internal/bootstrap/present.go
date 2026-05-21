package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/order/config"
	"gitlab.com/ecommercehub1/order/internal/common/log"
	"gitlab.com/ecommercehub1/order/internal/present/http/controller"
	"go.uber.org/fx"
)

func PresentModule() fx.Option {
	return fx.Options(
		fx.Provide(controller.NewOrderController),
		fx.Invoke(startHTTPServer),
	)
}

func startHTTPServer(lc fx.Lifecycle, orderCtrl *controller.OrderController) {
	router := gin.Default()
	
	v1 := router.Group("/api/v1")
	orderCtrl.SetupRoutes(v1)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.AppConfig.Server.Port),
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info(ctx, fmt.Sprintf("Starting HTTP server on %s", config.AppConfig.Server.Port))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal("failed to listen and serve", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info(ctx, "Stopping HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
