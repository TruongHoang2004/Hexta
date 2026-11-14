package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Pool limits per API process. Keep the total across processes below the server's max_connections.
const (
	maxOpenConns    = 50
	maxIdleConns    = 25
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
)

func NewDatabase() (*gorm.DB, error) {
	log.GetLogger()
	log.Info(context.Background(), "Connecting to database")
	db, err := gorm.Open(postgres.Open(config.AppConfig.Database.Url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to access database pool: %w", err)
	}
	configurePool(sqlDB)

	log.Info(context.Background(), "Database connected")
	return db, nil
}

func configurePool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
}
