package bootstrap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/constant"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRouterAddsTraceIDToErrorResponses(t *testing.T) {
	// Exercise the production middleware chain so an unregistered Tracer is caught.
	gin.SetMode(gin.TestMode)
	errors.Init(constant.ServiceName, constant.TraceIdName)
	t.Cleanup(func() { errors.Init("", nil) })

	r := newRouter(zap.NewNop().Sugar(), &config.Config{})
	r.GET("/trace-error", func(c *gin.Context) {
		err := errors.ErrNotFound(c.Request.Context(), "order", "missing")
		c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/trace-error", nil)
	r.ServeHTTP(w, req)

	var body errors.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("error response is not JSON: %v", err)
	}
	if len(body.TraceID) != 32 {
		t.Fatalf("expected a 32-character trace_id in the error response, got %q", body.TraceID)
	}
}
