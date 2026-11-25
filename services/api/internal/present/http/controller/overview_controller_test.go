package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	applog "github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type overviewStoreStub struct {
	tenant string
	filter types.OverviewFilter
}

func (s *overviewStoreStub) WithinSnapshot(_ context.Context, fn func(repository.OverviewStore) *apperrors.Error) *apperrors.Error {
	return fn(s)
}
func (s *overviewStoreStub) Tenant(ctx context.Context, id string) (*model.Tenant, *apperrors.Error) {
	if id != s.tenant {
		return nil, apperrors.ErrNotFound(ctx, "tenant", "not found")
	}
	return &model.Tenant{ID: id, OwnerID: "owner", Status: model.TenantStatusActive}, nil
}
func (s *overviewStoreStub) IsMember(_ context.Context, _, user string) (bool, *apperrors.Error) {
	return user == "owner" || user == "sale", nil
}
func (s *overviewStoreStub) HasActivity(context.Context, string) (bool, *apperrors.Error) {
	return true, nil
}
func (s *overviewStoreStub) OrderStatusTotals(_ context.Context, _ string, filter types.OverviewFilter) ([]types.OverviewOrderStatusTotal, *apperrors.Error) {
	s.filter = filter
	return []types.OverviewOrderStatusTotal{{Status: "confirmed", Count: 1, Value: decimal.RequireFromString("1500.50")}}, nil
}
func (s *overviewStoreStub) WorkCounts(context.Context, string, types.OverviewFilter, int) (types.OverviewWork, *apperrors.Error) {
	return types.OverviewWork{BySteps: []types.OverviewStepCount{}}, nil
}
func (s *overviewStoreStub) OpenWork(context.Context, string, types.OverviewFilter, time.Time, int) ([]types.OverviewWorkRow, int64, *apperrors.Error) {
	return nil, 0, nil
}

func TestOverviewHTTPContract(t *testing.T) {
	previous := config.AppConfig
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	gin.SetMode(gin.TestMode)
	stub := &overviewStoreStub{tenant: uuid.NewString()}
	ctrl := NewOverviewController(validator.New(), service.NewOverviewService(stub))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	router.GET("/tenants/:tenantID/overview", ctrl.Overview)
	get := func(tenant, query, actor string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/tenants/"+tenant+"/overview"+query, nil)
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	for name, tc := range map[string]struct {
		tenant, query, actor string
		status               int
	}{
		"member denied":        {stub.tenant, "", "sale", 403},
		"non-member denied":    {stub.tenant, "", "stranger", 403},
		"missing actor":        {stub.tenant, "", "", 401},
		"invalid tenant":       {"not-a-uuid", "", "owner", 400},
		"missing tenant":       {uuid.NewString(), "", "owner", 404},
		"date only":            {stub.tenant, "?from=2026-11-01&to=2026-11-02", "owner", 400},
		"offsetless timestamp": {stub.tenant, "?from=2026-11-01T00:00:00&to=2026-11-02T00:00:00", "owner", 400},
		"one bound":            {stub.tenant, "?from=2026-11-01T00:00:00Z", "owner", 400},
		"inverted":             {stub.tenant, "?from=2026-11-02T00:00:00Z&to=2026-11-01T00:00:00Z", "owner", 400},
		"repeated key":         {stub.tenant, "?employee_id=a&employee_id=b", "owner", 400},
		"unknown employee":     {stub.tenant, "?employee_id=nobody", "owner", 400},
	} {
		if res := get(tc.tenant, tc.query, tc.actor); res.Code != tc.status {
			t.Fatalf("%s: got %d, want %d: %s", name, res.Code, tc.status, res.Body.String())
		}
	}
	res := get(stub.tenant, "?from=2026-11-01T00:00:00%2B07:00&to=2026-11-08T00:00:00%2B07:00&employee_id=sale", "owner")
	if res.Code != http.StatusOK {
		t.Fatalf("owner overview: %d %s", res.Code, res.Body.String())
	}
	if !stub.filter.From.Equal(time.Date(2026, 10, 31, 17, 0, 0, 0, time.UTC)) || *stub.filter.EmployeeID != "sale" {
		t.Fatalf("filter was not passed through: %+v", stub.filter)
	}
	var body struct {
		Data struct {
			Filters struct {
				From string `json:"from"`
			} `json:"filters"`
			Orders struct {
				OrderValue string            `json:"order_value"`
				ByStatus   []json.RawMessage `json:"by_status"`
			} `json:"orders"`
			Work struct {
				BySteps []json.RawMessage `json:"by_steps"`
			} `json:"work"`
			Stuck struct {
				ThresholdHours int               `json:"threshold_hours"`
				Items          []json.RawMessage `json:"items"`
			} `json:"stuck"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	d := body.Data
	if d.Filters.From != "2026-10-31T17:00:00Z" || d.Orders.OrderValue != "1500.5" || len(d.Orders.ByStatus) != 7 || d.Stuck.ThresholdHours != 48 {
		t.Fatalf("unexpected body: %s", res.Body.String())
	}
	if d.Work.BySteps == nil || d.Stuck.Items == nil {
		t.Fatalf("lists must serialize as [] rather than null: %s", res.Body.String())
	}
}
