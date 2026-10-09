package model

import (
	"time"
)

type StockMovementType string

const (
	MovementInbound    StockMovementType = "inbound"
	MovementOutbound   StockMovementType = "outbound"
	MovementAdjustment StockMovementType = "adjustment"
	MovementReserve    StockMovementType = "reserve"
	MovementRelease    StockMovementType = "release"
)

type InventoryItem struct {
	ID              string    `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;type:varchar(36);not null;index;uniqueIndex:idx_inventory_tenant_variant" json:"tenant_id"`
	VariantID       string    `gorm:"column:variant_id;type:varchar(36);not null;index;uniqueIndex:idx_inventory_tenant_variant" json:"variant_id"`
	OnHandQty       int       `gorm:"column:on_hand_qty;not null;default:0" json:"on_hand_qty"`
	ReservedQty     int       `gorm:"column:reserved_qty;not null;default:0" json:"reserved_qty"`
	AvailableQty    int       `gorm:"column:available_qty;not null;default:0" json:"available_qty"`
	SafetyThreshold int       `gorm:"column:safety_threshold;not null;default:5" json:"safety_threshold"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (InventoryItem) TableName() string {
	return "inventory_items"
}

type StockMovement struct {
	ID            string            `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID      string            `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	InventoryID   string            `gorm:"column:inventory_id;type:varchar(36);not null;index" json:"inventory_id"`
	MovementType  StockMovementType `gorm:"column:movement_type;type:varchar(32);not null;index" json:"movement_type"`
	Quantity      int               `gorm:"column:quantity;not null" json:"quantity"`
	BalanceAfter  int               `gorm:"column:balance_after;not null" json:"balance_after"`
	ReferenceType string            `gorm:"column:reference_type;type:varchar(64)" json:"reference_type,omitempty"`
	ReferenceID   *string           `gorm:"column:reference_id;type:varchar(36);index" json:"reference_id,omitempty"`
	Notes         string            `gorm:"column:notes;type:text" json:"notes,omitempty"`
	CreatedAt     time.Time         `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
}

func (StockMovement) TableName() string {
	return "stock_movements"
}
