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

func TestOrderCaseHTTPValidationAndOwnerDecisions(t *testing.T) {
	previous := config.AppConfig
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	tenant, order, item := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, row := range []any{
		&model.Tenant{ID: tenant, Name: "HTTP", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive},
		&model.TenantMember{TenantID: tenant, UserID: "owner"},
		&model.TenantMember{TenantID: tenant, UserID: "employee"},
		&model.Order{ID: order, TenantID: tenant, OrderNumber: order, Status: model.OrderStatusConfirmed},
		&model.OrderItem{ID: item, TenantID: tenant, OrderID: order, VariantID: uuid.NewString(), Quantity: 2, PriceProvided: true},
	} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewOrderCaseController(validator.New(), service.NewOrderCaseService(repository.NewOrderCaseRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	g := router.Group("/tenants/:tenantID")
	g.GET("/orders/:id/case", ctrl.Get)
	g.POST("/orders/:id/case", ctrl.Open)
	g.PUT("/orders/:id/case", ctrl.Update)
	g.POST("/orders/:id/case/send-back", ctrl.SendBack)
	g.POST("/orders/:id/case/finalize", ctrl.Finalize)
	request := func(method, path, actor string, input any, status int) dto.OrderCaseDTO {
		t.Helper()
		var body []byte
		if input != nil {
			body, _ = json.Marshal(input)
		}
		req := httptest.NewRequest(method, "/tenants/"+tenant+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s %s got %d, want %d: %s", method, path, res.Code, status, res.Body.String())
		}
		var result struct {
			Data dto.OrderCaseDTO `json:"data"`
		}
		if status == 200 && method != http.MethodGet {
			if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
		}
		return result.Data
	}
	base := "/orders/" + order + "/case"
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenant+base, nil)
	req.Header.Set("Test-Actor", "employee")
	router.ServeHTTP(res, req)
	if res.Code != 200 || res.Body.String() != `{"data":{"case":null},"meta":null}` {
		t.Fatalf("an order without a case returns a null case: %d %s", res.Code, res.Body.String())
	}
	request(http.MethodGet, "/orders/not-a-uuid/case", "employee", nil, 400)
	request(http.MethodPost, base, "employee", map[string]any{"type": "refund"}, 400)
	opened := request(http.MethodPost, base, "employee", dto.OpenOrderCaseRequest{Type: "return"}, 200)
	if opened.Lines == nil || len(opened.Events) != 1 || opened.HandlerID != "employee" {
		t.Fatalf("unexpected open response %+v", opened)
	}
	request(http.MethodPut, base, "employee", dto.UpdateOrderCaseRequest{Version: 1, Lines: []dto.OrderCaseLineRequest{{OrderItemID: item, Quantity: 3}}}, 400)
	request(http.MethodPut, base, "employee", map[string]any{"version": 0}, 400)
	updated := request(http.MethodPut, base, "employee", dto.UpdateOrderCaseRequest{Version: 1, Note: "received", Lines: []dto.OrderCaseLineRequest{{OrderItemID: item, Quantity: 2}}}, 200)
	request(http.MethodPost, base+"/finalize", "employee", dto.OrderCaseDecisionRequest{Version: updated.Version}, 403)
	request(http.MethodPost, base+"/send-back", "employee", dto.OrderCaseDecisionRequest{Version: updated.Version, Note: "x"}, 403)
	request(http.MethodPost, base+"/finalize", "owner", dto.OrderCaseDecisionRequest{Version: 1}, 409)
	finalized := request(http.MethodPost, base+"/finalize", "owner", dto.OrderCaseDecisionRequest{Version: updated.Version}, 200)
	if finalized.Status != "finalized" || finalized.FinalizedBy == nil || *finalized.FinalizedBy != "owner" {
		t.Fatalf("unexpected finalize response %+v", finalized)
	}
	request(http.MethodPut, base, "owner", dto.UpdateOrderCaseRequest{Version: finalized.Version}, 409)
}
