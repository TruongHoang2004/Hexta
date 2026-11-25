package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/gin-gonic/gin"
)

func commandRouteEngine(handle func(string) gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerMessagingCommandRoutes(engine.Group("/tenants/:tenantID"), handle)
	return engine
}

func TestMessagingCommandRoutesAreAllWrites(t *testing.T) {
	engine := commandRouteEngine((&MessagingCommandGateway{}).Handle)
	routes := engine.Routes()
	if len(routes) != 11 {
		t.Fatalf("want 11 command routes, got %d", len(routes))
	}
	for _, route := range routes {
		if route.Method != http.MethodPost {
			t.Fatalf("command side must only expose writes, found %s %s", route.Method, route.Path)
		}
	}
}

func TestMessagingCommandGatewayWithoutTargetFailsClosed(t *testing.T) {
	engine := commandRouteEngine(NewMessagingCommandGateway(&config.Config{}).Handle)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tenants/t1/turns/x/claim", strings.NewReader("{}")))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestMessagingCommandGatewayForwardsRequest(t *testing.T) {
	var gotPath, gotAuth, gotBody string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotPath, gotAuth, gotBody = r.URL.Path, r.Header.Get("Authorization"), string(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer upstream.Close()
	cfg := &config.Config{}
	cfg.Messaging.CommandURL = upstream.URL
	engine := commandRouteEngine(NewMessagingCommandGateway(cfg).Handle)
	front := httptest.NewServer(engine) // a real server: the proxy needs a CloseNotifier-capable writer
	defer front.Close()
	req, _ := http.NewRequest(http.MethodPost, front.URL+"/tenants/t1/turns/x/claim", strings.NewReader(`{"expected_version":1}`))
	req.Header.Set("Authorization", "Bearer token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	rec := struct{ Code int }{resp.StatusCode}
	if rec.Code != http.StatusCreated || gotPath != "/tenants/t1/turns/x/claim" || gotAuth != "Bearer token" || gotBody != `{"expected_version":1}` {
		t.Fatalf("not forwarded intact: status=%d path=%q auth=%q body=%q", rec.Code, gotPath, gotAuth, gotBody)
	}
}
