package router

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
)

// RegisterOrderCaseRoutes mounts the cancellation/return case endpoints under the same authenticated,
// tenant-member-checked group as the other tenant resources. The service decides handler/owner permission.
func RegisterOrderCaseRoutes(params Params, cases *controller.OrderCaseController, auth *middleware.AuthMiddleware, tenants *middleware.TenantMiddleware) {
	account := params.Public.Group("", auth.RequireAuth())
	resources := tenantResources(account.Group("/tenants"), tenants)
	resources.GET("/orders/:id/case", cases.Get)
	resources.POST("/orders/:id/case", cases.Open)
	resources.PUT("/orders/:id/case", cases.Update)
	resources.POST("/orders/:id/case/send-back", cases.SendBack)
	resources.POST("/orders/:id/case/finalize", cases.Finalize)
}
