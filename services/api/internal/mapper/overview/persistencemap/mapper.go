// Package persistencemap converts owner overview scan rows to core values.
package persistencemap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/overviewmodel"
)

func OrderStatusTotals(rows []overviewmodel.OrderStatusRow) []types.OverviewOrderStatusTotal {
	result := make([]types.OverviewOrderStatusTotal, len(rows))
	for i, row := range rows {
		result[i] = types.OverviewOrderStatusTotal{Status: row.Status, Count: row.Count, Value: row.Value}
	}
	return result
}

// WorkTotals copies cohort totals; steps are attached separately by the caller.
func WorkTotals(row overviewmodel.WorkTotalsRow, steps []overviewmodel.StepCountRow) types.OverviewWork {
	result := types.OverviewWork{Total: row.Total, Active: row.Active, Waiting: row.Waiting, Completed: row.Completed, Unassigned: row.Unassigned, BySteps: make([]types.OverviewStepCount, len(steps))}
	for i, s := range steps {
		result.BySteps[i] = types.OverviewStepCount{WorkflowVersionID: s.WorkflowVersionID, WorkflowName: s.WorkflowName, WorkflowVersionNumber: s.WorkflowVersionNumber, StepOrdinal: s.StepOrdinal, StepName: s.StepName, Active: s.Active, Waiting: s.Waiting, Completed: s.Completed, Unassigned: s.Unassigned}
	}
	return result
}

// WorkRows normalizes timestamps to UTC so responses do not depend on the session time zone.
func WorkRows(rows []overviewmodel.WorkRow) []types.OverviewWorkRow {
	result := make([]types.OverviewWorkRow, len(rows))
	for i, row := range rows {
		result[i] = types.OverviewWorkRow{WorkItemID: row.WorkItemID, OrderID: row.OrderID, OrderNumber: row.OrderNumber, WorkflowName: row.WorkflowName, StepOrdinal: row.StepOrdinal, StepName: row.StepName, State: row.State, Responsible: row.ResponsibleID, Recipient: row.RecipientID, WaitingSince: row.UpdatedAt.UTC()}
	}
	return result
}

func CaseTotals(rows []overviewmodel.CaseTotalRow) []types.OverviewCaseTotal {
	result := make([]types.OverviewCaseTotal, len(rows))
	for i, row := range rows {
		result[i] = types.OverviewCaseTotal{OrderStatus: row.OrderStatus, CaseType: row.CaseType, CaseStatus: row.CaseStatus, Count: row.Count, Value: row.Value}
	}
	return result
}

// PendingCases uses the case's last change as the waiting start, in UTC.
func PendingCases(rows []overviewmodel.PendingCaseRow) []types.OverviewPendingCase {
	result := make([]types.OverviewPendingCase, len(rows))
	for i, row := range rows {
		result[i] = types.OverviewPendingCase{CaseID: row.CaseID, OrderID: row.OrderID, OrderNumber: row.OrderNumber, CaseType: row.CaseType, HandlerID: row.HandlerID, WaitingSince: row.UpdatedAt.UTC()}
	}
	return result
}
