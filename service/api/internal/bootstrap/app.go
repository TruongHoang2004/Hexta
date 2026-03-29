package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/config"
	"go.uber.org/fx"
)

var ServerModule = fx.Module("http",
	fx.Provide(func(r *gin.Engine) *http.Server {
		return &http.Server{
			Addr:    fmt.Sprintf(":%s", config.AppConfig.Server.Port),
			Handler: r,
		}
	}),
	fx.Invoke(func(lc fx.Lifecycle, srv *http.Server) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						panic(err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				return srv.Shutdown(ctx)
			},
		})
	}),
)
