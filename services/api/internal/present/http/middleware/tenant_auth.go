package middleware

import (
	"context"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/gin-gonic/gin"
)

type tenantAuthorizer interface {
	AuthorizeMember(ctx context.Context, userID, tenantID string) (*model.Tenant, *errors.Error)
}

type TenantMiddleware struct {
	tenants tenantAuthorizer
}

func NewTenantMiddleware(tenantService *service.TenantService) *TenantMiddleware {
	return &TenantMiddleware{tenants: tenantService}
}

// RequireTenantMember rejects callers who are not current members of an active :tenantID workspace.
// The decision is the tenant service's; services still re-check membership inside their transactions.
func (m *TenantMiddleware) RequireTenantMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		if userID == "" {
			err := errors.ErrUnauthorized(c.Request.Context()).SetDetail("Authorization token is required")
			c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			c.Abort()
			return
		}
		tenant, err := m.tenants.AuthorizeMember(c.Request.Context(), userID, c.Param("tenantID"))
		if err != nil {
			c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(common.SetTenant(c.Request.Context(), tenant))
		c.Next()
	}
}
