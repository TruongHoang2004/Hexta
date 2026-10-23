package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

type fixtureHealthService struct{ status string }

func (s fixtureHealthService) CheckHealth(context.Context) service.HealthCheckResult {
	return service.HealthCheckResult{Status: s.status, Details: map[string]string{"database": "fixture-state"}}
}
func TestHealthResponseMatchesDocumentedEnvelope(t *testing.T) {
	for _, tc := range []struct {
		status string
		code   int
	}{{"up", http.StatusOK}, {"down", http.StatusServiceUnavailable}} {
		r := gin.New()
		ctrl := NewHealthController(validator.New(), fixtureHealthService{tc.status})
		r.GET("/api/v1/health", ctrl.HealthCheck)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
		require.Equal(t, tc.code, rec.Code)
		var body response.Response[dto.HealthResponse]
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, tc.status, body.Data.Status)
		require.Equal(t, "fixture-state", body.Data.Details["database"])
	}
}
