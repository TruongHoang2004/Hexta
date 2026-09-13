package model

import "time"

const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"

	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"

	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type Tenant struct {
	ID        string         `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Slug      string         `gorm:"column:slug;type:varchar(255);not null;uniqueIndex" json:"slug"`
	Plan      string         `gorm:"column:plan;type:varchar(50);not null;default:'free'" json:"plan"`
	Status    string         `gorm:"column:status;type:varchar(50);not null;default:'active'" json:"status"`
	OwnerID   string         `gorm:"column:owner_id;type:varchar(255);not null;index" json:"owner_id"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Members   []TenantMember `gorm:"foreignKey:TenantID;references:ID" json:"members,omitempty"`
}

func (Tenant) TableName() string {
	return "tenants"
}

type TenantMember struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;type:varchar(36);not null;index;uniqueIndex:idx_tenant_user" json:"tenant_id"`
	UserID    string    `gorm:"column:user_id;type:varchar(255);not null;index;uniqueIndex:idx_tenant_user" json:"user_id"`
	Role      string    `gorm:"column:role;type:varchar(50);not null;default:'member'" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (TenantMember) TableName() string {
	return "tenant_members"
}
