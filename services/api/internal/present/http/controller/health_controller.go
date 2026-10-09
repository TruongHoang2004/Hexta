package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
)

type HealthController struct {
	*baseController
	healthService service.IHealthService
}

func NewHealthController(
	validate *validator.Validate,
	healthService service.IHealthService,
) *HealthController {
	return &HealthController{
		baseController: NewBaseController(validate),
		healthService:  healthService,
	}
}

// HealthCheck godoc
// @Summary Health Check
// @Description Check the health of the application and its dependencies
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthController) HealthCheck(c *gin.Context) {
	result := h.healthService.CheckHealth(c.Request.Context())

	httpStatus := http.StatusOK
	if result.Status == "down" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response.NewSuccessResponse(gin.H{
		"status":  result.Status,
		"details": result.Details,
	}, nil))
}
