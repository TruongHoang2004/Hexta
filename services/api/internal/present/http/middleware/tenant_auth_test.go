package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeTenantAuthorizer struct {
	tenant *model.Tenant
	err    func(context.Context) *errors.Error
	calls  []string
}

func (f *fakeTenantAuthorizer) AuthorizeMember(ctx context.Context, userID, tenantID string) (*model.Tenant, *errors.Error) {
	f.calls = append(f.calls, userID+"@"+tenantID)
	if f.err != nil {
		return nil, f.err(ctx)
	}
	return f.tenant, nil
}

func tenantRouter(authorizer *fakeTenantAuthorizer, userID string) (*gin.Engine, *bool) {
	gin.SetMode(gin.TestMode)
	reached := false
	r := gin.New()
	resources := r.Group("/tenants/:tenantID", func(c *gin.Context) {
		if userID != "" {
			c.Set("userID", userID)
		}
	})
	resources.Use((&TenantMiddleware{tenants: authorizer}).RequireTenantMember())
	resources.GET("/orders", func(c *gin.Context) {
		reached = true
		tenant, ok := common.GetTenant(c.Request.Context())
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, tenant.ID)
	})
	return r, &reached
}

func serveTenantRequest(r *gin.Engine) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tenants/t1/orders", nil)
	r.ServeHTTP(w, req)
	return w
}

func TestRequireTenantMemberInjectsAuthorizedTenant(t *testing.T) {
	authorizer := &fakeTenantAuthorizer{tenant: &model.Tenant{ID: "t1", Status: model.TenantStatusActive}}
	r, reached := tenantRouter(authorizer, "user-1")

	w := serveTenantRequest(r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "t1", w.Body.String())
	assert.True(t, *reached)
	assert.Equal(t, []string{"user-1@t1"}, authorizer.calls)
}

func TestRequireTenantMemberRejectsBeforeHandler(t *testing.T) {
	forbidden := func(ctx context.Context) *errors.Error {
		return errors.ErrForbidden(ctx).SetDetail("workspace membership required")
	}
	t.Run("non-member or inactive workspace", func(t *testing.T) {
		r, reached := tenantRouter(&fakeTenantAuthorizer{err: forbidden}, "user-1")
		w := serveTenantRequest(r)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.False(t, *reached)
	})
	t.Run("missing authenticated user", func(t *testing.T) {
		authorizer := &fakeTenantAuthorizer{tenant: &model.Tenant{ID: "t1"}}
		r, reached := tenantRouter(authorizer, "")
		w := serveTenantRequest(r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
		assert.Empty(t, authorizer.calls)
	})
}
