package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "gitlab.com/ecommercehub1/api/docs"
	"gitlab.com/ecommercehub1/api/internal/present/http/controller"
	"go.uber.org/fx"
)

func RegisterRoutes(
	params Params,
	healthController *controller.HealthController,
	authController *controller.AuthController,
) {
	// Root level public routes
	params.Public.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	params.Public.GET("/health", healthController.HealthCheck)

	// Auth routes
	authGroup := params.Public.Group("/auth")
	{
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/refresh", authController.RefreshToken)
		authGroup.POST("/logout", authController.Logout)
	}

	// Swagger UI
	params.Public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

}

func CreatePublicRouterGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1")
}

func CreatePrivateRouterGroup(r *gin.Engine) *gin.RouterGroup {
	private := r.Group("/api/v1")
	return private
}

type Params struct {
	fx.In

	Public  *gin.RouterGroup `name:"public"`
	Private *gin.RouterGroup `name:"private"`
}
