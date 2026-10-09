package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusFulfilled  OrderStatus = "fulfilled"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusReturned   OrderStatus = "returned"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid        PaymentStatus = "unpaid"
	PaymentStatusPartiallyPaid PaymentStatus = "partially_paid"
	PaymentStatusPaid          PaymentStatus = "paid"
	PaymentStatusRefunded      PaymentStatus = "refunded"
)

type OrderSource string

const (
	OrderSourceWebForm OrderSource = "web_form"
	OrderSourceAIDraft OrderSource = "ai_draft"
	OrderSourceAPI     OrderSource = "api"
	OrderSourceImport  OrderSource = "import"
)

type Order struct {
	ID            string          `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID      string          `gorm:"column:tenant_id;type:varchar(36);not null;index;uniqueIndex:idx_orders_tenant_number" json:"tenant_id"`
	OrderNumber   string          `gorm:"column:order_number;type:varchar(64);not null;uniqueIndex:idx_orders_tenant_number" json:"order_number"`
	CustomerID    *string         `gorm:"column:customer_id;type:varchar(36);index" json:"customer_id,omitempty"`
	EmployeeID    *string         `gorm:"column:employee_id;type:varchar(36);index" json:"employee_id,omitempty"`
	Status        OrderStatus     `gorm:"column:status;type:varchar(32);not null;default:'draft';index" json:"status"`
	PaymentStatus PaymentStatus   `gorm:"column:payment_status;type:varchar(32);not null;default:'unpaid';index" json:"payment_status"`
	Subtotal      decimal.Decimal `gorm:"column:subtotal;type:numeric(15,2);not null;default:0" json:"subtotal"`
	Discount      decimal.Decimal `gorm:"column:discount;type:numeric(15,2);not null;default:0" json:"discount"`
	TotalAmount   decimal.Decimal `gorm:"column:total_amount;type:numeric(15,2);not null;default:0" json:"total_amount"`
	Source        OrderSource     `gorm:"column:source;type:varchar(32);not null;default:'web_form';index" json:"source"`
	DraftID       *string         `gorm:"column:draft_id;type:varchar(36);index" json:"draft_id,omitempty"`
	Notes         string          `gorm:"column:notes;type:text" json:"notes,omitempty"`
	CreatedAt     time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	Items         []OrderItem     `gorm:"foreignKey:OrderID;references:ID" json:"items,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID        string          `gorm:"primaryKey;type:varchar(36);column:id" json:"id"`
	TenantID  string          `gorm:"column:tenant_id;type:varchar(36);not null;index" json:"tenant_id"`
	OrderID   string          `gorm:"column:order_id;type:varchar(36);not null;index" json:"order_id"`
	VariantID string          `gorm:"column:variant_id;type:varchar(36);not null;index" json:"variant_id"`
	Quantity  int             `gorm:"column:quantity;not null" json:"quantity"`
	UnitPrice decimal.Decimal `gorm:"column:price;type:numeric(15,2);not null;default:0" json:"unit_price"`
	LineTotal decimal.Decimal `gorm:"column:line_total;type:numeric(15,2);not null;default:0" json:"line_total"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
