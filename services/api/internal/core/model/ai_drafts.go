package model

import (
	"time"
)

type AIDraftStatus string

const (
	DraftStatusProposed  AIDraftStatus = "proposed"
	DraftStatusConfirmed AIDraftStatus = "confirmed"
	DraftStatusDiscarded AIDraftStatus = "discarded"
	DraftStatusExpired   AIDraftStatus = "expired"
)

type AIDraft struct {
	ID           string        `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID     string        `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	UserID       string        `gorm:"column:user_id;type:varchar(255);not null;index" json:"user_id"`
	Intent       string        `gorm:"column:intent;type:varchar(64);not null;index" json:"intent"`
	Prompt       string        `gorm:"column:prompt;type:text;not null" json:"prompt"`
	DraftPayload string        `gorm:"column:draft_payload;type:jsonb;not null" json:"draft_payload"`
	Status       AIDraftStatus `gorm:"column:status;type:varchar(32);not null;default:'proposed';index" json:"status"`
	ExpiresAt    time.Time     `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt    time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AIDraft) TableName() string {
	return "ai_drafts"
}
