package database

import (
	"context"
	"fmt"

	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase() (*gorm.DB, error) {
	log.GetLogger()
	log.Info(context.Background(), "Connecting to database")
	db, err := gorm.Open(postgres.Open(config.AppConfig.Database.Url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	log.Info(context.Background(), "Database connected")
	return db, nil
}
