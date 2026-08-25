package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/api/internal/present/http/response"
	"gorm.io/gorm"
)

type HealthController struct {
	*baseController
	db    *gorm.DB
	redis *cache.RedisClient
}

func NewHealthController(
	validate *validator.Validate,
	db *gorm.DB,
	redis *cache.RedisClient,
) *HealthController {
	return &HealthController{
		baseController: NewBaseController(validate),
		db:             db,
		redis:          redis,
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
	status := "up"
	details := make(map[string]string)

	// Check Database
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "down"
		details["database"] = "unreachable"
	} else if err := sqlDB.Ping(); err != nil {
		status = "down"
		details["database"] = "ping failed"
	} else {
		details["database"] = "ok"
	}

	// Check Redis
	if err := h.redis.Client.Ping(context.Background()).Err(); err != nil {
		status = "down"
		details["redis"] = "ping failed"
	} else {
		details["redis"] = "ok"
	}

	httpStatus := http.StatusOK
	if status == "down" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response.NewSuccessResponse(gin.H{
		"status":  status,
		"details": details,
	}, nil))
}
