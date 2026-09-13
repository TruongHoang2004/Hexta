package model

import "time"

type Provider string

const (
	ProviderLocal    Provider = "local"
	ProviderEmail    Provider = "email"
	ProviderGoogle   Provider = "google"
	ProviderFacebook Provider = "facebook"
)

type AuthIdentities struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID     string    `gorm:"column:user_id;not null;index" json:"user_id"`
	Provider   Provider  `gorm:"column:provider;type:varchar(50);not null;uniqueIndex:idx_identities_provider_identifier" json:"provider"`
	Identifier string    `gorm:"column:identifier;type:varchar(255);not null;uniqueIndex:idx_identities_provider_identifier" json:"identifier"`
	Password   *string   `gorm:"column:password;type:varchar(255)" json:"password,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AuthIdentities) TableName() string {
	return "identities"
}
