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
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func TestWorkflowHTTPValidationAndTransitions(t *testing.T) {
	previous := config.AppConfig
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	tenant, order := uuid.NewString(), uuid.NewString()
	for _, row := range []any{&model.Tenant{ID: tenant, Name: "HTTP", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive}, &model.TenantMember{TenantID: tenant, UserID: "owner"}, &model.TenantMember{TenantID: tenant, UserID: "employee"}, &model.Order{ID: order, TenantID: tenant, OrderNumber: order, Status: model.OrderStatusConfirmed}} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewWorkflowController(validator.New(), service.NewWorkflowService(repository.NewWorkflowRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	g := router.Group("/tenants/:tenantID")
	g.POST("/workflow-definitions", ctrl.CreateDefinition)
	g.POST("/work-items", ctrl.CreateWork)
	g.POST("/work-items/:id/handoff", ctrl.Handoff)
	g.POST("/work-items/:id/accept", ctrl.Accept)
	g.POST("/work-items/:id/complete", ctrl.Complete)
	request := func(path, actor, key string, input any, status int) json.RawMessage {
		t.Helper()
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/tenants/"+tenant+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		req.Header.Set("Idempotency-Key", key)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s got %d, want %d: %s", path, res.Code, status, res.Body.String())
		}
		if status == 200 && (strings.Contains(res.Body.String(), "create_hash") || strings.Contains(res.Body.String(), "request_key")) {
			t.Fatal("internal creation metadata exposed")
		}
		var result struct {
			Data json.RawMessage `json:"data"`
		}
		if status == 200 {
			if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
		}
		return result.Data
	}
	definitionInput := dtoWorkflowInput()
	request("/workflow-definitions", "employee", uuid.NewString(), definitionInput, 403)
	request("/workflow-definitions", "owner", "", definitionInput, 400)
	request("/workflow-definitions", "owner", uuid.NewString(), map[string]any{"name": "Bad", "steps": []map[string]string{{"name": "Step", "handoff_mode": "unsupported"}}}, 400)
	var definition repository.WorkflowDefinition
	if err := json.Unmarshal(request("/workflow-definitions", "owner", uuid.NewString(), definitionInput, 200), &definition); err != nil {
		t.Fatal(err)
	}
	request("/work-items", "owner", uuid.NewString(), map[string]string{"order_id": "invalid", "workflow_version_id": definition.Versions[0].ID, "responsible_id": "owner"}, 400)
	var work repository.WorkItem
	if err := json.Unmarshal(request("/work-items", "owner", uuid.NewString(), types.WorkInput{OrderID: order, WorkflowVersionID: definition.Versions[0].ID, ResponsibleID: "owner"}, 200), &work); err != nil {
		t.Fatal(err)
	}
	request("/work-items/"+work.ID+"/handoff", "owner", "", map[string]any{"recipient_id": "employee"}, 400)
	request("/work-items/"+work.ID+"/handoff", "owner", "", types.WorkCommand{ExpectedVersion: 1, RecipientID: "employee"}, 200)
	request("/work-items/"+work.ID+"/accept", "owner", "", types.WorkCommand{ExpectedVersion: 2}, 403)
	request("/work-items/"+work.ID+"/accept", "employee", "", types.WorkCommand{ExpectedVersion: 2}, 200)
	request("/work-items/"+work.ID+"/complete", "employee", "", types.WorkCommand{ExpectedVersion: 2}, 409)
	request("/work-items/"+work.ID+"/complete", "employee", "", types.WorkCommand{ExpectedVersion: 3}, 200)
}
func dtoWorkflowInput() types.WorkflowInput {
	return types.WorkflowInput{Name: "HTTP workflow", Steps: []types.WorkflowStepInput{{Name: "Prepare", HandoffMode: "acceptance"}, {Name: "Review", HandoffMode: "direct"}}}
}
