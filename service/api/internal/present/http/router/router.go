package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "gitlab.com/ecommercehub1/api/docs"
	"gitlab.com/ecommercehub1/api/internal/present/http/controller"
	"gitlab.com/ecommercehub1/api/internal/present/http/middleware"
	"go.uber.org/fx"
)

func RegisterRoutes(
	params Params,
	authController *controller.AuthController,
	userController *controller.UserController,
) {
	// Root level public routes
	params.Public.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Swagger UI
	params.Public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Delegate registration to controllers
	authController.RegisterRoutes(params.Public, params.Private)
	userController.RegisterRoutes(params.Private)
}

func CreatePublicRouterGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1")
}

func CreatePrivateRouterGroup(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) *gin.RouterGroup {
	private := r.Group("/api/v1")
	private.Use(authMiddleware.Authentication())
	return private
}

type Params struct {
	fx.In

	Public  *gin.RouterGroup `name:"public"`
	Private *gin.RouterGroup `name:"private"`
}
