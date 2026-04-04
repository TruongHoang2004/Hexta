package database

import (
	"fmt"
	"log"

	"gitlab.com/ecommercehub1/order/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase() (*gorm.DB, error) {
	log.Printf("🔌 Connecting to database %s", config.AppConfig.Database.Url)
	db, err := gorm.Open(postgres.Open(config.AppConfig.Database.Url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	log.Println("✅ Database connected")
	return db, nil
}
