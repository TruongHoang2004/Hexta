package model

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        string         `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID  string         `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	FullName  string         `gorm:"column:full_name;type:varchar(255);not null;index" json:"full_name"`
	Phone     string         `gorm:"column:phone;type:varchar(32);index;uniqueIndex:idx_customers_tenant_phone" json:"phone"`
	Email     string         `gorm:"column:email;type:varchar(255);index" json:"email"`
	Address   string         `gorm:"column:address;type:text" json:"address"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

func (Customer) TableName() string {
	return "customers"
}
