package model

import (
	"time"
)

type AuditLog struct {
	ID          string    `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	UserID      string    `gorm:"column:user_id;type:varchar(255);not null;index" json:"user_id"`
	Action      string    `gorm:"column:action;type:varchar(64);not null;index" json:"action"`
	EntityType  string    `gorm:"column:entity_type;type:varchar(64);not null;index" json:"entity_type"`
	EntityID    string    `gorm:"column:entity_id;type:varchar(36);not null;index" json:"entity_id"`
	Source      string    `gorm:"column:source;type:varchar(32);not null;default:'web_form';index" json:"source"`
	DraftID     *string   `gorm:"column:draft_id;type:varchar(36);index" json:"draft_id,omitempty"`
	BeforeState *string   `gorm:"column:before_state;type:jsonb" json:"before_state,omitempty"`
	AfterState  string    `gorm:"column:after_state;type:jsonb;not null" json:"after_state"`
	IPAddress   *string   `gorm:"column:ip_address;type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent   *string   `gorm:"column:user_agent;type:text" json:"user_agent,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
