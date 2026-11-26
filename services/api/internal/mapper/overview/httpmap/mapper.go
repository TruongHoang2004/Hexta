// Package httpmap converts the owner overview use-case result to its HTTP DTO.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

// Overview always emits arrays (never null) so clients can render empty states directly.
func Overview(value types.OverviewResult) dto.OverviewResponse {
	result := dto.OverviewResponse{
		Filters:     dto.OverviewFiltersResponse{From: value.Filter.From, To: value.Filter.To, EmployeeID: value.Filter.EmployeeID},
		HasActivity: value.HasActivity,
		Orders:      dto.OverviewOrdersResponse{Total: value.Orders.Total, OrderValue: value.Orders.OrderValue, ByStatus: make([]dto.OverviewOrderBucketResponse, len(value.Orders.ByStatus))},
		Work:        dto.OverviewWorkResponse{Total: value.Work.Total, Active: value.Work.Active, Waiting: value.Work.Waiting, Completed: value.Work.Completed, Unassigned: value.Work.Unassigned, BySteps: make([]dto.OverviewStepResponse, len(value.Work.BySteps))},
		Stuck:       dto.OverviewStuckResponse{ThresholdHours: value.Stuck.ThresholdHours, Total: value.Stuck.Total, Items: make([]dto.OverviewStuckItemResponse, len(value.Stuck.Items))},
		Exceptions: dto.OverviewExceptionsResponse{
			InProgress:          caseCounts(value.Exceptions.InProgress),
			Finalized:           caseCounts(value.Exceptions.Finalized),
			CancelledOrderValue: value.Exceptions.CancelledOrderValue,
		},
		PendingCases: dto.OverviewPendingCasesResponse{Total: value.PendingCases.Total, Items: make([]dto.OverviewPendingCaseResponse, len(value.PendingCases.Items))},
	}
	for i, b := range value.Orders.ByStatus {
		result.Orders.ByStatus[i] = dto.OverviewOrderBucketResponse{Status: b.Status, Count: b.Count, OrderValue: b.Value}
	}
	for i, s := range value.Work.BySteps {
		result.Work.BySteps[i] = dto.OverviewStepResponse{WorkflowVersionID: s.WorkflowVersionID, WorkflowName: s.WorkflowName, WorkflowVersionNumber: s.WorkflowVersionNumber, StepOrdinal: s.StepOrdinal, StepName: s.StepName, Active: s.Active, Waiting: s.Waiting, Completed: s.Completed, Unassigned: s.Unassigned}
	}
	for i, item := range value.Stuck.Items {
		result.Stuck.Items[i] = dto.OverviewStuckItemResponse{WorkItemID: item.WorkItemID, OrderID: item.OrderID, OrderNumber: item.OrderNumber, WorkflowName: item.WorkflowName, StepOrdinal: item.StepOrdinal, StepName: item.StepName, State: item.State, ResponsibleID: item.Responsible, RecipientID: item.Recipient, Reason: item.Reason, WaitingSince: item.WaitingSince}
	}
	for i, item := range value.PendingCases.Items {
		result.PendingCases.Items[i] = dto.OverviewPendingCaseResponse{CaseID: item.CaseID, OrderID: item.OrderID, OrderNumber: item.OrderNumber, CaseType: item.CaseType, HandlerID: item.HandlerID, WaitingSince: item.WaitingSince}
	}
	return result
}

func caseCounts(value types.OverviewCaseCounts) dto.OverviewCaseCountsResponse {
	return dto.OverviewCaseCountsResponse{Total: value.Total, Cancellation: value.Cancellation, Return: value.Return}
}
