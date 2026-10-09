package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusArchived ProductStatus = "archived"
)

type Product struct {
	ID          string          `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID    string          `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	Name        string          `gorm:"column:name;type:varchar(255);not null;index" json:"name"`
	Description string          `gorm:"column:description;type:text" json:"description"`
	Category    string          `gorm:"column:category;type:varchar(100);index" json:"category"`
	Status      ProductStatus   `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	Variants    []ProductVariant `gorm:"foreignKey:ProductID;references:ID" json:"variants,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

type ProductVariant struct {
	ID          string          `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID    string          `gorm:"column:tenant_id;type:varchar(36);not null;index;uniqueIndex:idx_variants_tenant_sku" json:"tenant_id"`
	ProductID   string          `gorm:"column:product_id;type:varchar(36);not null;index" json:"product_id"`
	SKU         string          `gorm:"column:sku;type:varchar(64);not null;uniqueIndex:idx_variants_tenant_sku" json:"sku"`
	VariantName string          `gorm:"column:variant_name;type:varchar(128);not null" json:"variant_name"`
	Price       decimal.Decimal `gorm:"column:price;type:numeric(15,2);not null;default:0" json:"price"`
	CostPrice   decimal.Decimal `gorm:"column:cost_price;type:numeric(15,2);not null;default:0" json:"cost_price"`
	Barcode     string          `gorm:"column:barcode;type:varchar(64);index" json:"barcode"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
