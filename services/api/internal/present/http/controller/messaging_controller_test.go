package controller

import (
	"bytes"
	"encoding/json"
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

func TestMessagingHTTPFlowAndValidation(t *testing.T) {
	previous := config.AppConfig
	if previous == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	tenant := uuid.NewString()
	for _, row := range []any{&model.Tenant{ID: tenant, Name: "Inbox", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive}, &model.TenantMember{TenantID: tenant, UserID: "owner"}, &model.TenantMember{TenantID: tenant, UserID: "sale"}} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewMessagingController(validator.New(), service.NewMessagingService(repository.NewMessagingRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	group := router.Group("/tenants/:tenantID")
	group.GET("/conversations", ctrl.List)
	group.GET("/conversations/:id/messages", ctrl.Detail)
	group.POST("/simulator/messages", ctrl.Inbound)
	group.POST("/turns/:id/claim", ctrl.Claim)
	group.POST("/turns/:id/assign", ctrl.Assign)
	group.POST("/turns/:id/handoff", ctrl.Handoff)
	group.POST("/turns/:id/return-to-bot", ctrl.ReturnToBot)
	group.POST("/turns/:id/send-simulation", ctrl.Send)
	request := func(method, path, actor, body string, status int) json.RawMessage {
		t.Helper()
		req := httptest.NewRequest(method, "/tenants/"+tenant+path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s: got %d, want %d: %s", path, res.Code, status, res.Body.String())
		}
		if strings.Contains(res.Body.String(), "request_hash") || strings.Contains(res.Body.String(), "request_key") && status == 200 {
			t.Fatal("internal request metadata exposed")
		}
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if status == 200 {
			if err := json.Unmarshal(res.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if len(envelope.Data) == 0 {
				t.Fatal("missing success envelope")
			}
		}
		return envelope.Data
	}
	request("GET", "/conversations", "", "", 401)
	request("GET", "/conversations?limit=101", "sale", "", 400)
	request("GET", "/conversations?cursor=bad", "sale", "", 400)
	request("GET", "/conversations?filter=other", "sale", "", 400)
	request("GET", "/conversations/not-a-uuid/messages", "sale", "", 400)
	request("POST", "/simulator/messages", "sale", "{", 400)
	request("POST", "/simulator/messages", "sale", `{"body":"Hello"}`, 400)
	key := uuid.NewString()
	input := `{"external_key":"customer-1","display_name":"Customer","body":"Hello","request_key":"` + key + `"}`
	var message dto.MessagingMessage
	if err := json.Unmarshal(request("POST", "/simulator/messages", "sale", input, 200), &message); err != nil {
		t.Fatal(err)
	}
	request("POST", "/simulator/messages", "sale", input, 200)
	var detail dto.MessagingDetail
	if err := json.Unmarshal(request("GET", "/conversations/"+message.ThreadID+"/messages", "sale", "", 200), &detail); err != nil {
		t.Fatal(err)
	}
	if detail.Conversation.Source != "simulator" || len(detail.Messages) != 1 || detail.Conversation.Turn == nil {
		t.Fatal("bad detail contract")
	}
	var turn dto.MessagingTurn
	if err := json.Unmarshal(request("POST", "/turns/"+detail.Conversation.Turn.ID+"/claim", "sale", `{"expected_version":1}`, 200), &turn); err != nil {
		t.Fatal(err)
	}
	request("POST", "/turns/"+turn.ID+"/assign", "sale", `{"expected_version":2}`, 403)
	request("POST", "/turns/"+turn.ID+"/send-simulation", "owner", `{"expected_version":2,"context_revision":1,"body":"No","request_key":"`+uuid.NewString()+`"}`, 403)
	request("POST", "/turns/"+turn.ID+"/send-simulation", "sale", `{"expected_version":2,"context_revision":1,"body":"Reply","request_key":"`+uuid.NewString()+`"}`, 200)

	request("POST", "/turns/"+turn.ID+"/handoff", "sale", `{"expected_version":3}`, 400)
	request("POST", "/turns/"+turn.ID+"/handoff", "sale", `{"expected_version":3,"reason":"arbitrary"}`, 400)
	request("POST", "/turns/"+turn.ID+"/return-to-bot", "", `{"expected_version":3}`, 401)
	request("POST", "/turns/"+turn.ID+"/return-to-bot", "sale", `{"expected_version":2}`, 409)
	var returned dto.MessagingTurn
	if err := json.Unmarshal(request("POST", "/turns/"+turn.ID+"/return-to-bot", "sale", `{"expected_version":3}`, 200), &returned); err != nil {
		t.Fatal(err)
	}
	if returned.HandlingState != "BOT_ACTIVE" || returned.AssigneeID != nil {
		t.Fatal("bad bot control contract")
	}
	request("POST", "/turns/"+turn.ID+"/handoff", "sale", `{"expected_version":4,"reason":"customer_requested_staff"}`, 200)
	request("POST", "/turns/"+turn.ID+"/handoff", "sale", `{"expected_version":4,"reason":"customer_requested_staff"}`, 409)
	request("POST", "/turns/"+turn.ID+"/send-simulation", "sale", `{"expected_version":5,"context_revision":2,"body":"No","request_key":"`+uuid.NewString()+`"}`, 403)
	request("POST", "/simulator/messages", "sale", `{"external_key":"a","display_name":"A","body":"`+strings.Repeat("a", 33000)+`","request_key":"`+uuid.NewString()+`"}`, 400)
}
