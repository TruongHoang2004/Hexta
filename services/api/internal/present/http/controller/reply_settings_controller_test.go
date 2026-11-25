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
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TestReplyModeHTTPOwnerOnly covers S-19 over HTTP: members read, only the owner changes, versions guard races and
// malformed input is a 400 before the service runs.
func TestReplyModeHTTPOwnerOnly(t *testing.T) {
	previous := config.AppConfig
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	tenant, owner, member := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, row := range []any{&model.Tenant{ID: tenant, Name: "HTTP", Slug: tenant, OwnerID: owner, Status: model.TenantStatusActive}, &model.TenantMember{TenantID: tenant, UserID: owner, Role: model.RoleOwner}, &model.TenantMember{TenantID: tenant, UserID: member}} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewReplySettingsController(validator.New(), service.NewReplySettingsService(repository.NewMessagingRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	g := router.Group("/tenants/:tenantID")
	g.GET("/settings/reply-mode", ctrl.GetReplyMode)
	g.PUT("/settings/reply-mode", ctrl.UpdateReplyMode)
	call := func(method, actor, body string, status int) dto.ReplySettingsResponse {
		t.Helper()
		req := httptest.NewRequest(method, "/tenants/"+tenant+"/settings/reply-mode", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s got %d, want %d: %s", method, res.Code, status, res.Body.String())
		}
		var envelope struct {
			Data dto.ReplySettingsResponse `json:"data"`
		}
		if status == http.StatusOK {
			if err := json.Unmarshal(res.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
		}
		return envelope.Data
	}
	if got := call(http.MethodGet, member, "", http.StatusOK); got.Configured || got.ReplyMode != types.ReplyModeSuggest {
		t.Fatal(got)
	}
	call(http.MethodPut, owner, `{"reply_mode":"always","expected_version":0}`, http.StatusBadRequest)
	call(http.MethodPut, owner, `{"reply_mode":"auto_send","expected_version":-1}`, http.StatusBadRequest)
	call(http.MethodPut, member, `{"reply_mode":"auto_send","expected_version":0}`, http.StatusForbidden)
	if got := call(http.MethodPut, owner, `{"reply_mode":"auto_send","expected_version":0}`, http.StatusOK); !got.Configured || got.ReplyMode != types.ReplyModeAutoSend || got.Version != 1 {
		t.Fatal(got)
	}
	call(http.MethodPut, owner, `{"reply_mode":"suggest","expected_version":0}`, http.StatusConflict)
}
