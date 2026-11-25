// Package overviewmodel holds the scan rows of the read-only owner overview projection.
package overviewmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatusRow is one GROUP BY status row of the order cohort.
type OrderStatusRow struct {
	Status string
	Count  int64
	Value  decimal.Decimal
}

// WorkTotalsRow is the whole work cohort counted by state.
type WorkTotalsRow struct {
	Total      int64
	Active     int64
	Waiting    int64
	Completed  int64
	Unassigned int64
}

// StepCountRow is the work cohort grouped by pinned workflow version and current step.
type StepCountRow struct {
	WorkflowVersionID     string
	WorkflowName          string
	WorkflowVersionNumber int64
	StepOrdinal           int
	StepName              string
	Active                int64
	Waiting               int64
	Completed             int64
	Unassigned            int64
}

// WorkRow is one open work item joined to its source order and pinned step.
type WorkRow struct {
	WorkItemID    string
	OrderID       string
	OrderNumber   string
	WorkflowName  string
	StepOrdinal   int
	StepName      string
	State         string
	ResponsibleID *string
	RecipientID   *string
	UpdatedAt     time.Time
}
