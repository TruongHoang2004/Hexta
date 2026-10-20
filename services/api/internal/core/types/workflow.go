package types

type WorkflowStepInput struct {
	Name        string `json:"name" validate:"required,max=120"`
	HandoffMode string `json:"handoff_mode" validate:"required,oneof=direct acceptance"`
}
type WorkflowInput struct {
	Name            string              `json:"name" validate:"required,max=120"`
	ExpectedVersion int64               `json:"expected_version" validate:"min=0"`
	Steps           []WorkflowStepInput `json:"steps" validate:"required,min=1,max=30,dive"`
}
type WorkInput struct {
	OrderID           string `json:"order_id" validate:"required,uuid"`
	WorkflowVersionID string `json:"workflow_version_id" validate:"required,uuid"`
	ResponsibleID     string `json:"responsible_id" validate:"required,max=255"`
}
type WorkCommand struct {
	ExpectedVersion int64   `json:"expected_version" validate:"required,min=1"`
	RecipientID     string  `json:"recipient_id" validate:"max=255"`
	ResponsibleID   *string `json:"responsible_id" validate:"omitempty,min=1,max=255"`
}
