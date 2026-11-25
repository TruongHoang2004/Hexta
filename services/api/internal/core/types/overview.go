package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// OverviewFilter narrows the owner overview. From/To are a half-open creation-time interval
// [From, To); both nil means all time. EmployeeID nil means every employee.
type OverviewFilter struct {
	From       *time.Time
	To         *time.Time
	EmployeeID *string
}

// OverviewOrderStatusTotal is one grouped order status as stored, before bucketing.
type OverviewOrderStatusTotal struct {
	Status string
	Count  int64
	Value  decimal.Decimal
}

// OverviewOrderBucket is one reported order status. Status "other" collects unknown stored values.
type OverviewOrderBucket struct {
	Status string
	Count  int64
	Value  decimal.Decimal
}

// OverviewOrders summarizes the order cohort. OrderValue is the summed total of orders whose status
// counts toward order value; it is never revenue (no payment, refund or cost is considered).
type OverviewOrders struct {
	Total      int64
	OrderValue decimal.Decimal
	ByStatus   []OverviewOrderBucket
}

// OverviewStepCount counts work currently at one step of one pinned workflow version.
type OverviewStepCount struct {
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

// OverviewWork summarizes the work cohort. Unassigned is a subset of Active, not a fourth state.
type OverviewWork struct {
	Total      int64
	Active     int64
	Waiting    int64
	Completed  int64
	Unassigned int64
	BySteps    []OverviewStepCount
}

// OverviewWorkRow is open work as stored; the service decides why it is listed.
type OverviewWorkRow struct {
	WorkItemID   string
	OrderID      string
	OrderNumber  string
	WorkflowName string
	StepOrdinal  int
	StepName     string
	State        string
	Responsible  *string
	Recipient    *string
	WaitingSince time.Time
}

// OverviewStuckItem is one row of the waiting/stuck list with its listing reason.
type OverviewStuckItem struct {
	OverviewWorkRow
	Reason string
}

// OverviewStuck lists open work that is unassigned or unchanged for longer than the threshold.
type OverviewStuck struct {
	ThresholdHours int
	Total          int64
	Items          []OverviewStuckItem
}

// OverviewResult is the read-only owner overview for one tenant snapshot.
type OverviewResult struct {
	Filter OverviewFilter
	// HasActivity is true when the tenant has any order or work item at all, ignoring filters.
	HasActivity bool
	Orders      OverviewOrders
	Work        OverviewWork
	Stuck       OverviewStuck
}
