package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestScriptHTTPPermissionAcceptance covers the PRD acceptance scenario over HTTP: a member with only script.edit
// is refused activation, while the owner's grant and activation succeed with the response envelope.
func TestScriptHTTPPermissionAcceptance(t *testing.T) {
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
	ctrl := NewScriptController(validator.New(), service.NewScriptService(repository.NewScriptRepository(db), repository.NewTenantDBRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	g := router.Group("/tenants/:tenantID")
	g.POST("/scripts", ctrl.CreateScript)
	g.PUT("/scripts/:id/draft", ctrl.SaveScriptDraft)
	g.POST("/scripts/:id/activate", ctrl.ActivateScript)
	g.GET("/script-permissions/me", ctrl.MyScriptPermissions)
	g.POST("/script-permissions/:userID/grant", ctrl.GrantScriptPermission)
	call := func(method, path, actor string, input any, status int, out any) {
		t.Helper()
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(method, "/tenants/"+tenant+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		req.Header.Set("Idempotency-Key", uuid.NewString())
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s %s got %d, want %d: %s", method, path, res.Code, status, res.Body.String())
		}
		if status == http.StatusOK && strings.Contains(res.Body.String(), "create_hash") {
			t.Fatal("internal idempotency hash exposed")
		}
		if out != nil {
			envelope := struct {
				Data json.RawMessage `json:"data"`
			}{}
			if err := json.Unmarshal(res.Body.Bytes(), &envelope); err != nil || len(envelope.Data) == 0 {
				t.Fatalf("missing response envelope: %s", res.Body.String())
			}
			if err := json.Unmarshal(envelope.Data, out); err != nil {
				t.Fatal(err)
			}
		}
	}

	call(http.MethodPost, "/script-permissions/"+member+"/grant", owner, map[string]any{"permission": "script.edit", "expected_version": 0}, http.StatusOK, nil)
	call(http.MethodPost, "/script-permissions/not-a-uuid/grant", owner, map[string]any{"permission": "script.edit", "expected_version": 1}, http.StatusBadRequest, nil)
	call(http.MethodPost, "/script-permissions/"+member+"/grant", owner, map[string]any{"permission": "script.delete", "expected_version": 1}, http.StatusBadRequest, nil)
	var mine dto.MyScriptPermissionsResponse
	call(http.MethodGet, "/script-permissions/me", member, nil, http.StatusOK, &mine)
	if mine.Owner || len(mine.Permissions) != 1 || mine.Permissions[0] != types.ScriptPermissionEdit {
		t.Fatalf("member must hold only script.edit: %+v", mine)
	}

	call(http.MethodPost, "/scripts", member, map[string]any{"name": "Price", "triggers": []string{"price"}, "answer": "200k"}, http.StatusForbidden, nil)
	call(http.MethodPost, "/scripts", owner, map[string]any{"name": "Price", "triggers": []string{}, "answer": "200k"}, http.StatusBadRequest, nil)
	var script dto.ScriptDTO
	call(http.MethodPost, "/scripts", owner, map[string]any{"name": "Price", "triggers": []string{"price"}, "answer": "200k"}, http.StatusOK, &script)
	if script.Status != types.ScriptStatusInactive || len(script.Versions) != 1 || !script.Versions[0].Draft {
		t.Fatalf("new script must be inactive with draft v1: %+v", script)
	}

	var edited dto.ScriptDTO
	call(http.MethodPut, "/scripts/"+script.ID+"/draft", member, map[string]any{"name": "Price", "triggers": []string{"price", "cost"}, "answer": "190k", "expected_version": script.Version}, http.StatusOK, &edited)
	activate := map[string]any{"version_id": edited.Versions[0].ID, "expected_version": edited.Version}
	call(http.MethodPost, "/scripts/"+script.ID+"/activate", member, activate, http.StatusForbidden, nil)
	var active dto.ScriptDTO
	call(http.MethodPost, "/scripts/"+script.ID+"/activate", owner, activate, http.StatusOK, &active)
	if active.Status != types.ScriptStatusActive || active.ActiveVersionID != edited.Versions[0].ID || active.Versions[0].Draft {
		t.Fatalf("owner activation must publish the draft: %+v", active)
	}
	call(http.MethodPost, "/scripts/"+script.ID+"/activate", owner, activate, http.StatusConflict, nil)
}
