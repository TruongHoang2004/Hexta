package bootstrap

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/internal/constant"
	"gitlab.com/ecommercehub1/api/internal/present/http/middleware"
	"gitlab.com/ecommercehub1/api/internal/present/http/router"
	"gitlab.com/ecommercehub1/shared/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)



var RouterModule = fx.Options(
	fx.Provide(func(logger *zap.SugaredLogger) *gin.Engine {
		r := gin.Default()

		r.Use(cors.Default())
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
