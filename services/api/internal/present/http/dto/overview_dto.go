package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// OverviewQuery is the raw owner overview query; the controller parses timestamps.
type OverviewQuery struct {
	From       string `form:"from" validate:"omitempty,max=64"`
	To         string `form:"to" validate:"omitempty,max=64"`
	EmployeeID string `form:"employee_id" validate:"omitempty,max=255"`
}

type OverviewFiltersResponse struct {
	From       *time.Time `json:"from"`
	To         *time.Time `json:"to"`
	EmployeeID *string    `json:"employee_id"`
}

type OverviewOrderBucketResponse struct {
	// Status is a stored order status or "other" for unrecognized values.
	Status     string          `json:"status"`
	Count      int64           `json:"count"`
	OrderValue decimal.Decimal `json:"order_value" swaggertype:"string"`
}

type OverviewOrdersResponse struct {
	Total int64 `json:"total"`
	// OrderValue sums totals of confirmed, processing and fulfilled orders. It is not revenue.
	OrderValue decimal.Decimal               `json:"order_value" swaggertype:"string"`
	ByStatus   []OverviewOrderBucketResponse `json:"by_status"`
}

type OverviewStepResponse struct {
	WorkflowVersionID     string `json:"workflow_version_id"`
	WorkflowName          string `json:"workflow_name"`
	WorkflowVersionNumber int64  `json:"workflow_version_number"`
	// StepOrdinal is the zero-based position of the step in the pinned workflow version.
	StepOrdinal int    `json:"step_ordinal"`
	StepName    string `json:"step_name"`
	Active      int64  `json:"active"`
	Waiting     int64  `json:"waiting"`
	Completed   int64  `json:"completed"`
	Unassigned  int64  `json:"unassigned"`
}

type OverviewWorkResponse struct {
	Total     int64 `json:"total"`
	Active    int64 `json:"active"`
	Waiting   int64 `json:"waiting"`
	Completed int64 `json:"completed"`
	// Unassigned is a subset of active.
	Unassigned int64                  `json:"unassigned"`
	BySteps    []OverviewStepResponse `json:"by_steps"`
}

type OverviewStuckItemResponse struct {
	WorkItemID    string  `json:"work_item_id"`
	OrderID       string  `json:"order_id"`
	OrderNumber   string  `json:"order_number"`
	WorkflowName  string  `json:"workflow_name"`
	StepOrdinal   int     `json:"step_ordinal"`
	StepName      string  `json:"step_name"`
	State         string  `json:"state" enums:"active,waiting"`
	ResponsibleID *string `json:"responsible_id"`
	RecipientID   *string `json:"recipient_id"`
	// Reason is unassigned or waiting_too_long.
	Reason       string    `json:"reason" enums:"unassigned,waiting_too_long"`
	WaitingSince time.Time `json:"waiting_since"`
}

type OverviewStuckResponse struct {
	ThresholdHours int                         `json:"threshold_hours"`
	Total          int64                       `json:"total"`
	Items          []OverviewStuckItemResponse `json:"items"`
}

type OverviewResponse struct {
	Filters OverviewFiltersResponse `json:"filters"`
	// HasActivity is false only for a tenant with no orders and no work at all.
	HasActivity bool                   `json:"has_activity"`
	Orders      OverviewOrdersResponse `json:"orders"`
	Work        OverviewWorkResponse   `json:"work"`
	Stuck       OverviewStuckResponse  `json:"stuck"`
}
