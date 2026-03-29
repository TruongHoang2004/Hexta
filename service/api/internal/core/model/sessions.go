package model

import "time"

type Sessions struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     string    `json:"user_id" gorm:"not null"`
	Token      string    `json:"token" gorm:"not null"`
	Provider   Provider  `json:"provider" gorm:"not null"`
	DeviceInfo string    `json:"device_info" gorm:"type:varchar(500)"`
	IpAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent  string    `json:"user_agent" gorm:"type:varchar(1000)"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	RevokedAt  time.Time `json:"revoked_at"`
}

func (Sessions) TableName() string {
	return "sessions"
}
