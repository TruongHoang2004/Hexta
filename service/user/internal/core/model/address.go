package model

import (
	"time"
)

type Address struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	Receiver  string    `gorm:"not null" json:"receiver"`
	Phone     string    `gorm:"not null" json:"phone"`
	City      string    `gorm:"not null" json:"city"`
	District  string    `gorm:"not null" json:"district"`
	Ward      string    `gorm:"not null" json:"ward"`
	Street    string    `gorm:"not null" json:"street"`
	Details   string    `gorm:"not null" json:"details"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
