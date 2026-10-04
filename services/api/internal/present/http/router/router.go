package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/TruongHoang2004/Hexta/services/api/docs"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"go.uber.org/fx"
)

func RegisterRoutes(
	params Params,
	healthController *controller.HealthController,
	authController *controller.AuthController,
	tenantController *controller.TenantController,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Root level public routes
	params.Public.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	params.Public.GET("/health", healthController.HealthCheck)

	// Auth public routes
	authPublic := params.Public.Group("/auth")
	{
		authPublic.POST("/register", authController.Register)
		authPublic.POST("/login", authController.Login)
		authPublic.POST("/refresh", authController.RefreshToken)
		authPublic.GET("/google/login", authController.GoogleLogin)
		authPublic.GET("/google/callback", authController.GoogleCallback)
	}

	// Auth private routes
	authPrivate := params.Private.Group("/auth")
	{
		authPrivate.POST("/logout", authController.Logout)
	}

	// Tenant routes (protected by AuthMiddleware)
	tenantGroup := params.Public.Group("/tenants", authMiddleware.RequireAuth())
	{
		tenantGroup.POST("", tenantController.CreateTenant)
		tenantGroup.GET("", tenantController.GetUserTenants)
		tenantGroup.GET("/:id", tenantController.GetTenant)
		tenantGroup.PUT("/:id", tenantController.UpdateTenant)
		tenantGroup.GET("/:id/users", tenantController.ListMembers)
		tenantGroup.POST("/:id/invites", tenantController.InviteMember)
	}

	// Swagger UI
	params.Public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func CreatePublicRouterGroup(r *gin.Engine) *gin.RouterGroup {
	return r.Group("/api/v1")
}

func CreatePrivateRouterGroup(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) *gin.RouterGroup {
	private := r.Group("/api/v1")
	private.Use(authMiddleware.Authenticate())
	return private
}

type Params struct {
	fx.In

	Public  *gin.RouterGroup `name:"public"`
	Private *gin.RouterGroup `name:"private"`
}
