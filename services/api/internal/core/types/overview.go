package types

import "time"

type OverviewQuery struct {
	From, To   *time.Time
	EmployeeID *string
	Assignment string
	Page, Size int
}
type OverviewFilters struct {
	From       *time.Time `json:"from"`
	To         *time.Time `json:"to"`
	EmployeeID *string    `json:"employee_id"`
	Assignment string     `json:"assignment"`
}
type OverviewOrderCounts struct {
	Total      int64 `json:"total"`
	Draft      int64 `json:"draft"`
	Confirmed  int64 `json:"confirmed"`
	Processing int64 `json:"processing"`
	Fulfilled  int64 `json:"fulfilled"`
	Cancelled  int64 `json:"cancelled"`
	Returned   int64 `json:"returned"`
	Other      int64 `json:"other"`
}
type OverviewWorkCounts struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Waiting    int64 `json:"waiting"`
	Completed  int64 `json:"completed"`
	Unassigned int64 `json:"unassigned"`
}
type OverviewWorkRow struct {
	ID                    string    `json:"id"`
	OrderID               string    `json:"order_id"`
	OrderNumber           string    `json:"order_number"`
	WorkflowVersionID     string    `json:"workflow_version_id"`
	WorkflowName          string    `json:"workflow_name"`
	WorkflowVersionNumber int64     `json:"workflow_version_number"`
	CurrentStep           int       `json:"current_step"`
	CurrentStepName       string    `json:"current_step_name"`
	ResponsibleID         *string   `json:"responsible_id"`
	RecipientID           *string   `json:"recipient_id"`
	State                 string    `json:"state"`
	Version               int64     `json:"version"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
type OverviewWorkPage struct {
	Items []OverviewWorkRow `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}
type OverviewResult struct {
	Filters   OverviewFilters     `json:"filters"`
	Orders    OverviewOrderCounts `json:"orders"`
	Work      OverviewWorkCounts  `json:"work"`
	WorkItems OverviewWorkPage    `json:"work_items"`
}
