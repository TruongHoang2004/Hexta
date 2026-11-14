package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/constant"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var traceIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func initErrorsWithTraceKey(t *testing.T) {
	t.Helper()
	errors.Init(constant.ServiceName, constant.TraceIdName)
	t.Cleanup(func() { errors.Init("", nil) })
}

func serveError(t *testing.T, router *gin.Engine, path string) errors.ErrorResponse {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(w, req)

	var body errors.ErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, w.Code, body.HTTPStatus)
	return body
}

func TestTracerPropagatesTraceIDToErrorResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initErrorsWithTraceKey(t)
	// Recovery logs through the global logger, which needs config to initialize.
	previousConfig := config.AppConfig
	config.AppConfig = &config.Config{}
	t.Cleanup(func() { config.AppConfig = previousConfig })
	log.GetLogger()

	router := gin.New()
	// Mirror production order: Recovery runs before Tracer.
	router.Use(middleware.Recovery(), middleware.Tracer())
	router.GET("/request-context", func(c *gin.Context) {
		err := errors.ErrNotFound(c.Request.Context(), "order", "missing")
		c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
	})
	router.GET("/gin-context", func(c *gin.Context) {
		err := errors.ErrBadRequest(c)
		c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
	})
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	for _, path := range []string{"/request-context", "/gin-context", "/panic"} {
		t.Run(path, func(t *testing.T) {
			body := serveError(t, router, path)
			assert.Regexp(t, traceIDPattern, body.TraceID)
			assert.Equal(t, constant.ServiceName, body.Source)
		})
	}

	t.Run("trace IDs are unique per request", func(t *testing.T) {
		first := serveError(t, router, "/request-context")
		second := serveError(t, router, "/request-context")
		assert.NotEqual(t, first.TraceID, second.TraceID)
	})
}

func TestErrorResponseOmitsTraceIDWithoutTracer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initErrorsWithTraceKey(t)

	router := gin.New()
	router.GET("/request-context", func(c *gin.Context) {
		err := errors.ErrNotFound(c.Request.Context(), "order", "missing")
		c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
	})

	body := serveError(t, router, "/request-context")
	assert.Empty(t, body.TraceID)
}
