package service

import (
	"context"

	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"gorm.io/gorm"
)

type HealthCheckResult struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details"`
}

type IHealthService interface {
	CheckHealth(ctx context.Context) HealthCheckResult
}

type healthService struct {
	*baseService
	db    *gorm.DB
	redis *cache.RedisClient
}

func NewHealthService(
	db *gorm.DB,
	redis *cache.RedisClient,
) IHealthService {
	return &healthService{
		baseService: NewBaseService(),
		db:          db,
		redis:       redis,
	}
}

func (s *healthService) CheckHealth(ctx context.Context) HealthCheckResult {
	status := "up"
	details := make(map[string]string)

	// Check Database
	if s.db == nil {
		status = "down"
		details["database"] = "unconfigured"
	} else {
		sqlDB, err := s.db.DB()
		if err != nil {
			status = "down"
			details["database"] = "unreachable"
		} else if err := sqlDB.Ping(); err != nil {
			status = "down"
			details["database"] = "ping failed"
		} else {
			details["database"] = "ok"
		}
	}

	// Check Redis
	if s.redis != nil && s.redis.Client != nil {
		if err := s.redis.Client.Ping(ctx).Err(); err != nil {
			status = "down"
			details["redis"] = "ping failed"
		} else {
			details["redis"] = "ok"
		}
	} else {
		details["redis"] = "unconfigured"
	}

	return HealthCheckResult{
		Status:  status,
		Details: details,
	}
}
