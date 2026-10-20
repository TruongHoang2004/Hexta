package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/telemetry"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/constant"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/router"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"strings"
)

var RouterModule = fx.Options(
	fx.Provide(func(logger *zap.SugaredLogger, cfg *config.Config) *gin.Engine {
		r := gin.Default()

		r.Use(cors.New(browserCORSConfig(cfg)))
		r.Use(middleware.Recovery())
		r.Use(otelgin.Middleware(constant.ServiceName))
		r.Use(telemetry.GinMiddleware(logger))

		r.GET("/metrics", gin.WrapH(promhttp.Handler()))

		return r

	}),

	fx.Provide(fx.Annotate(router.CreatePublicRouterGroup, fx.ResultTags(`name:"public"`))),
	fx.Provide(fx.Annotate(router.CreatePrivateRouterGroup, fx.ResultTags(`name:"private"`))),

	fx.Invoke(router.RegisterRoutes),
)

type Params struct {
	fx.In

	Public  *gin.RouterGroup `name:"public"`
	Private *gin.RouterGroup `name:"private"`
}

// Browser requests use explicit bearer authorization and a workspace header.
func browserCORSConfig(cfg *config.Config) cors.Config {
	result := cors.DefaultConfig()
	origin := strings.TrimRight(cfg.Server.FrontendURL, "/")
	if origin != "" {
		result.AllowOrigins = []string{origin}
	}
	if !cfg.Server.Production {
		result.AllowOrigins = append(result.AllowOrigins, "http://localhost:3000", "http://127.0.0.1:3000")
	}
	result.AllowHeaders = append(result.AllowHeaders, "Authorization", "X-Tenant-ID", "Idempotency-Key")
	return result
}
