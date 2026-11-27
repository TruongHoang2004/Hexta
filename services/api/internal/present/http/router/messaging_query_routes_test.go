package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// The static summary path must coexist with the conversation ID routes registered in RegisterRoutes.
func TestConversationSummaryRouteCoexistsWithDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	resources := engine.Group("/tenants/:tenantID")
	hit := ""
	resources.GET("/conversations", func(*gin.Context) { hit = "list" })
	resources.GET("/conversations/summary", func(*gin.Context) { hit = "summary" })
	resources.GET("/conversations/:id/messages", func(c *gin.Context) { hit = "detail:" + c.Param("id") })
	for path, want := range map[string]string{
		"/tenants/t/conversations":            "list",
		"/tenants/t/conversations/summary":    "summary",
		"/tenants/t/conversations/x/messages": "detail:x",
	} {
		hit = ""
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || hit != want {
			t.Fatalf("%s routed to %q (%d), want %q", path, hit, rec.Code, want)
		}
	}
}
