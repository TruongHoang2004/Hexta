package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TestBrowserPreflightAllowsAuthenticatedWorkspaceRequests(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Production = true
	cfg.Server.FrontendURL = "https://workspace.example.test"
	router := gin.New()
	router.Use(cors.New(browserCORSConfig(cfg)))
	router.GET("/orders", func(c *gin.Context) { c.Status(200) })
	request := httptest.NewRequest(http.MethodOptions, "/orders", nil)
	request.Header.Set("Origin", cfg.Server.FrontendURL)
	request.Header.Set("Access-Control-Request-Method", "GET")
	request.Header.Set("Access-Control-Request-Headers", "authorization,x-tenant-id,idempotency-key")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("preflight status %d", response.Code)
	}
	headers := strings.ToLower(response.Header().Get("Access-Control-Allow-Headers"))
	if !strings.Contains(headers, "authorization") || !strings.Contains(headers, "x-tenant-id") || !strings.Contains(headers, "idempotency-key") {
		t.Fatalf("missing workspace headers: %s", headers)
	}
	request.Header.Set("Origin", "https://unrelated.example.test")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("unexpected origin accepted: %d", response.Code)
	}
}

func TestBrowserCORSAllowsCredentialsOnlyForConfiguredOrigin(t *testing.T) {
	cfg := &config.Config{}
	cfg.Server.Production = true
	cfg.Server.FrontendURL = "https://workspace.example.test"
	router := gin.New()
	router.Use(cors.New(browserCORSConfig(cfg)))
	router.POST("/refresh", func(c *gin.Context) { c.Status(http.StatusOK) })
	for _, origin := range []string{cfg.Server.FrontendURL, "https://unrelated.example.test"} {
		request := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if origin == cfg.Server.FrontendURL {
			if response.Code != http.StatusOK || response.Header().Get("Access-Control-Allow-Credentials") != "true" || response.Header().Get("Access-Control-Allow-Origin") != origin {
				t.Fatal("credentialed response must allow only the configured origin")
			}
		} else if response.Code != http.StatusForbidden || response.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("untrusted credentialed origin accepted")
		}
	}
}
