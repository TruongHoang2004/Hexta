package router

import (
	_ "github.com/TruongHoang2004/Hexta/services/api/docs"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
)

func RegisterRoutes(
	params Params,
	healthController *controller.HealthController,
	authController *controller.AuthController,
	tenantController *controller.TenantController,
	salesController *controller.SalesController,
	workflows *controller.WorkflowController,
	overview *controller.OverviewController,
	onboarding *controller.TenantOnboardingController,
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

	account := params.Public.Group("", authMiddleware.RequireAuth())
	account.GET("/tenant-requests", onboarding.MyApplications)
	account.POST("/tenant-requests", onboarding.ApplyForTenant)
	account.GET("/platform/access", onboarding.PlatformAccess)
	account.GET("/platform/tenant-requests", onboarding.PlatformApplications)
	account.POST("/platform/tenant-requests/:id/review", onboarding.ReviewApplication)
	account.POST("/invitations/preview", onboarding.PreviewInvitation)
	account.POST("/invitations/accept", onboarding.AcceptInvitation)
	account.POST("/invitations/decline", onboarding.DeclineInvitation)
	tenantGroup := account.Group("/tenants")
	tenantGroup.GET("", tenantController.GetUserTenants)
	tenantGroup.GET("/:tenantID", tenantController.GetTenant)
	tenantGroup.PUT("/:tenantID", tenantController.UpdateTenant)
	tenantGroup.GET("/:tenantID/users", tenantController.ListMembers)
	tenantGroup.POST("/:tenantID/invitations", onboarding.CreateEmailInvitation)
	tenantGroup.GET("/:tenantID/invitations", onboarding.ListInvitations)
	tenantGroup.POST("/:tenantID/invitations/:id/revoke", onboarding.RevokeInvitation)
	resources := tenantGroup.Group("/:tenantID")
	resources.Use(func(c *gin.Context) {
		if header := c.GetHeader("X-Tenant-ID"); header != "" && header != c.Param("tenantID") {
			c.AbortWithStatusJSON(400, gin.H{"detail": "tenant route and header conflict"})
			return
		}
	})
	resources.POST("/orders", salesController.CreateOrder)
	resources.GET("/orders", salesController.ListOrders)
	resources.GET("/orders/:id", salesController.GetOrder)
	resources.PUT("/orders/:id", salesController.UpdateOrder)
	resources.POST("/orders/:id/confirm", salesController.ConfirmOrder)
	resources.GET("/customers", salesController.ListCustomers)
	resources.POST("/customers", salesController.CreateCustomer)
	resources.PUT("/customers/:id", salesController.UpdateCustomer)
	resources.DELETE("/customers/:id", salesController.ArchiveCustomer)
	resources.GET("/products", salesController.ListProducts)
	resources.POST("/products", salesController.CreateProduct)
	resources.PUT("/products/:id", salesController.UpdateProduct)
	resources.DELETE("/products/:id", salesController.ArchiveProduct)

	resources.GET("/workflow-definitions", workflows.ListDefinitions)
	resources.POST("/workflow-definitions", workflows.CreateDefinition)
	resources.POST("/workflow-definitions/:id/versions", workflows.PublishVersion)
	resources.GET("/overview", overview.GetOverview)
	resources.GET("/work-items", workflows.ListWork)
	resources.POST("/work-items", workflows.CreateWork)
	resources.GET("/work-items/:id", workflows.GetWork)
	resources.POST("/work-items/:id/handoff", workflows.Handoff)
	resources.POST("/work-items/:id/accept", workflows.Accept)
	resources.POST("/work-items/:id/assign", workflows.Assign)
	resources.POST("/work-items/:id/complete", workflows.Complete)

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
