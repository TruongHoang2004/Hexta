package model

import "time"

type TenantRegistration struct {
	ID           string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ApplicantID  string     `gorm:"type:varchar(255);not null;index" json:"applicant_id"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Slug         string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_tenant_registration_active_slug,where:status <> 'rejected'" json:"slug"`
	Status       string     `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	TenantID     *string    `gorm:"type:varchar(36)" json:"tenant_id"`
	ReviewedBy   string     `gorm:"type:varchar(255);not null;default:''" json:"reviewed_by"`
	ReviewReason string     `gorm:"type:text;not null;default:''" json:"review_reason"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TenantRegistration) TableName() string { return "tenant_registration_requests" }

type TenantInvitation struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID   string    `gorm:"type:varchar(36);not null;index;uniqueIndex:idx_tenant_pending_invitation_email,where:status = 'pending'" json:"tenant_id"`
	Email      string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_pending_invitation_email,where:status = 'pending'" json:"email"`
	Role       string    `gorm:"type:varchar(50);not null;default:'member'" json:"role"`
	InvitedBy  string    `gorm:"type:varchar(255);not null" json:"invited_by"`
	TokenHash  string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	Status     string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	AcceptedBy string    `gorm:"type:varchar(255);not null;default:''" json:"accepted_by"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TenantInvitation) TableName() string { return "tenant_invitations" }

type PlatformOperator struct {
	UserID    string    `gorm:"primaryKey;type:varchar(255)" json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PlatformOperator) TableName() string { return "platform_operators" }
