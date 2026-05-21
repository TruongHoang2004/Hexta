package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string         `gorm:"type:varchar(36);not null" json:"user_id"`
	ProductID string         `gorm:"type:varchar(36);not null" json:"product_id"`
	Quantity  int            `gorm:"not null" json:"quantity"`
	Status    string         `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
