// Package ordercasemodel holds the GORM rows of the order exception case aggregate. Atlas owns the physical schema.
package ordercasemodel

import "time"

type Case struct {
	ID           string
	TenantID     string
	OrderID      string
	CaseType     string
	Status       string
	HandlerID    string
	OpenedBy     string
	ProgressNote string
	FinalizedBy  *string
	FinalizedAt  *time.Time
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Case) TableName() string { return "order_cases" }

type Line struct {
	ID               string
	TenantID         string
	CaseID           string
	OrderID          string
	OrderItemID      string
	OriginalQuantity int64
	Quantity         int64
	CreatedAt        time.Time
}

func (Line) TableName() string { return "order_case_lines" }

// Event rows are append-only; Lines is a jsonb snapshot of the case lines after the event.
type Event struct {
	ID        string
	TenantID  string
	CaseID    string
	Version   int64
	ActorID   string
	Action    string
	Status    string
	Note      string
	Lines     string `gorm:"type:jsonb"`
	CreatedAt time.Time
}

func (Event) TableName() string { return "order_case_events" }
