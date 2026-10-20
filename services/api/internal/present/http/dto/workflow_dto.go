package dto

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/workflowmodel"
)

type WorkflowRequest types.WorkflowInput
type WorkCreateRequest types.WorkInput
type WorkCommandRequest types.WorkCommand
type WorkflowResponse workflowmodel.WorkflowDefinition
type WorkResponse workflowmodel.WorkItem
type WorkflowsResponse struct {
	Items []workflowmodel.WorkflowDefinition `json:"items"`
	Total int64                              `json:"total"`
	Page  int                                `json:"page"`
	Size  int                                `json:"size"`
}
type WorkItemsResponse struct {
	Items []workflowmodel.WorkItem `json:"items"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}
