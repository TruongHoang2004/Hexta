package model

import "time"

type MerchantStatus string

const (
	MerchantStatusPending   MerchantStatus = "pending"
	MerchantStatusActive    MerchantStatus = "active"
	MerchantStatusSuspended MerchantStatus = "suspended"
	MerchantStatusClosed    MerchantStatus = "closed"
)

type Merchant struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	OwnerID     string         `gorm:"not null;index" json:"owner_id"`
	Name        string         `gorm:"not null;unique" json:"name"`
	Email       string         `gorm:"not null;unique" json:"email"`
	Phone       string         `gorm:"not null;unique" json:"phone"`
	Description string         `gorm:"type:text" json:"description"`
	LogoURL     string         `gorm:"type:varchar(500)" json:"logo_url"`
	Status      MerchantStatus `gorm:"not null;default:'pending'" json:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time     `gorm:"index" json:"deleted_at,omitempty"`
}
