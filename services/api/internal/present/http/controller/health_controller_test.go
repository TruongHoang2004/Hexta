package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
)

type mockHealthService struct {
	result service.HealthCheckResult
}

func (m *mockHealthService) CheckHealth(ctx context.Context) service.HealthCheckResult {
	return m.result
}

func TestHealthController_HealthCheck_Up(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validate := validator.New()

	mockSvc := &mockHealthService{
		result: service.HealthCheckResult{
			Status: "up",
			Details: map[string]string{
				"database": "ok",
				"redis":    "ok",
			},
		},
	}

	ctrl := controller.NewHealthController(validate, mockSvc)

	router := gin.New()
	router.GET("/health", ctrl.HealthCheck)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response[dto.HealthCheckResponse]
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, "up", res.Data.Status)
	assert.Equal(t, "ok", res.Data.Details["database"])
	assert.Equal(t, "ok", res.Data.Details["redis"])
}

func TestHealthController_HealthCheck_Down(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validate := validator.New()

	mockSvc := &mockHealthService{
		result: service.HealthCheckResult{
			Status: "down",
			Details: map[string]string{
				"database": "ping failed",
				"redis":    "ok",
			},
		},
	}

	ctrl := controller.NewHealthController(validate, mockSvc)

	router := gin.New()
	router.GET("/health", ctrl.HealthCheck)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var res response.Response[dto.HealthCheckResponse]
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, "down", res.Data.Status)
	assert.Equal(t, "ping failed", res.Data.Details["database"])
}
