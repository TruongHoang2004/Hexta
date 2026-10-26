package controller

import (
	"encoding/json"
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

func TestOverviewHTTPQueryAndAuthorization(t *testing.T) {
	previous := config.AppConfig
	if previous == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	db := testutil.Database(t)
	tenant := uuid.NewString()
	for _, row := range []any{&model.Tenant{ID: tenant, Name: "Overview", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive}, &model.TenantMember{TenantID: tenant, UserID: "owner"}, &model.TenantMember{TenantID: tenant, UserID: "opaque:member"}} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctrl := NewOverviewController(validator.New(), service.NewOverviewService(repository.NewOverviewRepository(db)))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: c.GetHeader("Test-Actor")}))
	})
	router.GET("/tenants/:tenantID/overview", ctrl.GetOverview)
	request := func(path, actor string, status int) []byte {
		t.Helper()
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Test-Actor", actor)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("%s got %d want %d: %s", path, res.Code, status, res.Body.String())
		}
		return res.Body.Bytes()
	}
	base := "/tenants/" + tenant + "/overview"
	for _, query := range []string{"from=2026-10-19T00:00:00Z", "to=2026-10-20T00:00:00Z", "from=2026-10-20T00:00:00Z&to=2026-10-19T00:00:00Z", "from=2026-10-19&to=2026-10-20", "from=2026-10-19T00:00:00&to=2026-10-20T00:00:00", "from=2026-10-19T00:00:00Z&to=2026-10-19T00:00:00Z", "assignment=invalid", "assignment=unassigned&employee_id=owner", "page=0", "page=100001", "page=1.2", "page=garbage", "size=0", "size=101", "employee_id=missing", "employee_id="} {
		request(base+"?"+query, "owner", 400)
	}
	for _, key := range []string{"from", "to", "employee_id", "assignment", "page", "size"} {
		request(base+"?"+key+"=", "owner", 400)
		request(base+"?"+key+"=1&"+key+"=1", "owner", 400)
	}
	request(base, "", 401)
	request(base, "opaque:member", 403)
	request(base+"?employee_id=missing", "opaque:member", 403)
	request("/tenants/"+uuid.NewString()+"/overview", "owner", 404)
	request("/tenants/invalid/overview", "owner", 400)
	raw := request(base+"?from=2026-10-19T00:00:00.123456%2B07:00&to=2026-10-20T00:00:00%2B07:00&employee_id=opaque%3Amember&page=100000&size=100&unknown=ignored", "owner", 200)
	var result struct {
		Data dto.OverviewResponse `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Filters.From.Format("2006-01-02T15:04:05.999999999Z07:00") != "2026-10-18T17:00:00.123456Z" || result.Data.Filters.Assignment != "all" || result.Data.WorkItems.Items == nil || result.Data.WorkItems.Page != 100000 || result.Data.WorkItems.Size != 100 {
		t.Fatalf("normalized empty contract: %s", raw)
	}
}
