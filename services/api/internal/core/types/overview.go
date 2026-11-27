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
// counts toward order value, less orders whose cancellation case is finalized; it is never revenue (no
// payment, refund or cost is considered). ByStatus buckets keep the stored totals unchanged.
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

// OverviewCaseTotal is the order cohort's exception cases grouped as stored: by the original order
// status, case type and case status, with the summed original order totals.
type OverviewCaseTotal struct {
	OrderStatus string
	CaseType    string
	CaseStatus  string
	Count       int64
	Value       decimal.Decimal
}

// OverviewCaseCounts counts exception cases by type; Total is their sum.
type OverviewCaseCounts struct {
	Total        int64
	Cancellation int64
	Return       int64
}

// OverviewExceptions summarizes the cancellation/return cases of the order cohort. Cases never change the
// order's status, so they are reported here and not as order status buckets.
type OverviewExceptions struct {
	// InProgress counts cases not yet finalized (open or sent back to the handler).
	InProgress OverviewCaseCounts
	Finalized  OverviewCaseCounts
	// CancelledOrderValue is the total of value-counting orders whose cancellation case is finalized;
	// it is left out of OverviewOrders.OrderValue. Return cases never change order value.
	CancelledOrderValue decimal.Decimal
}

// OverviewPendingCase is an exception case waiting for the owner to finalize or send it back.
type OverviewPendingCase struct {
	CaseID       string
	OrderID      string
	OrderNumber  string
	CaseType     string
	HandlerID    string
	WaitingSince time.Time
}

// OverviewPendingCases lists open cases of the order cohort, longest waiting first, with the full total.
type OverviewPendingCases struct {
	Total int64
	Items []OverviewPendingCase
}

// OverviewResult is the read-only owner overview for one tenant snapshot.
type OverviewResult struct {
	Filter OverviewFilter
	// HasActivity is true when the tenant has any order or work item at all, ignoring filters.
	HasActivity  bool
	Orders       OverviewOrders
	Exceptions   OverviewExceptions
	Work         OverviewWork
	Stuck        OverviewStuck
	PendingCases OverviewPendingCases
}
