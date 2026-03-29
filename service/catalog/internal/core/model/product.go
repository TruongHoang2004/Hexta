package model

import (
	"time"
)

type Product struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;index" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null;type:decimal(10,2)" json:"price"`
	CategoryID  string    `gorm:"not null;index" json:"category_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
