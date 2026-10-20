package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/common"
	applog "github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// This is an isolated Atlas-migrated PostgreSQL HTTP test. No operator or membership in the
// running demo is changed by its fixtures.
func TestTenantOnboardingHTTPWorkflow(t *testing.T) {
	previous := config.AppConfig
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	if err := db.Create(&model.PlatformOperator{UserID: "test-operator"}).Error; err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"applicant", "recipient"} {
		if err := db.Create(&model.AuthIdentities{UserID: user, Identifier: user + "@example.test", Provider: model.ProviderLocal}).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewTenantOnboardingController(validator.New(), service.NewSalesService(repository.NewSalesRepository(db)), service.NewTenantOnboardingService(repository.NewTenantOnboardingRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	router.POST("/tenant-requests", ctrl.ApplyForTenant)
	router.POST("/platform/tenant-requests/:id/review", ctrl.ReviewApplication)
	router.POST("/tenants/:tenantID/invitations", ctrl.CreateEmailInvitation)
	router.POST("/invitations/accept", ctrl.AcceptInvitation)
	request := func(path, actor string, payload any, status int) map[string]json.RawMessage {
		t.Helper()
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s got %d, expected %d: %s", path, res.Code, status, res.Body.String())
		}
		var envelope struct {
			Data map[string]json.RawMessage `json:"data"`
		}
		if status == 200 {
			if err := json.Unmarshal(res.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
		}
		return envelope.Data
	}
	application := request("/tenant-requests", "applicant", map[string]string{"name": "HTTP Business", "slug": "http-business"}, 200)
	var id string
	json.Unmarshal(application["id"], &id)
	request("/platform/tenant-requests/"+id+"/review", "applicant", map[string]string{"decision": "approved"}, 403)
	approved := request("/platform/tenant-requests/"+id+"/review", "test-operator", map[string]string{"decision": "approved"}, 200)
	var tenant string
	json.Unmarshal(approved["tenant_id"], &tenant)
	invitation := request("/tenants/"+tenant+"/invitations", "applicant", map[string]string{"email": "recipient@example.test"}, 200)
	var token string
	json.Unmarshal(invitation["token"], &token)
	var members int64
	db.Model(&model.TenantMember{}).Where("tenant_id = ?", tenant).Count(&members)
	if members != 1 {
		t.Fatal("pending invitation activated a member")
	}
	request("/invitations/accept", "applicant", map[string]string{"token": token}, 403)
	request("/invitations/accept", "recipient", map[string]string{"token": token}, 200)
	request("/invitations/accept", "recipient", map[string]string{"token": token}, 200)
	db.Model(&model.TenantMember{}).Where("tenant_id = ?", tenant).Count(&members)
	if members != 2 {
		t.Fatal("accept must provision exactly one member")
	}
}
