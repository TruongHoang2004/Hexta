package workflowmodel

import "time"

// Workflow versions and their ordered steps are immutable after publication.
type WorkflowDefinition struct {
	ID         string            `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID   string            `gorm:"type:varchar(36);not null;index" json:"tenant_id"`
	Name       string            `gorm:"type:varchar(120);not null" json:"name"`
	Version    int64             `gorm:"not null" json:"version"`
	CreateHash string            `gorm:"type:varchar(64);not null" json:"-"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Versions   []WorkflowVersion `gorm:"foreignKey:DefinitionID" json:"versions"`
}

type WorkflowVersion struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID     string         `gorm:"type:varchar(36);not null;uniqueIndex:idx_workflow_version_number" json:"tenant_id"`
	DefinitionID string         `gorm:"type:varchar(36);not null;uniqueIndex:idx_workflow_version_number" json:"definition_id"`
	Number       int64          `gorm:"not null;uniqueIndex:idx_workflow_version_number" json:"number"`
	Name         string         `gorm:"type:varchar(120);not null" json:"name"`
	CreatedAt    time.Time      `json:"created_at"`
	Steps        []WorkflowStep `gorm:"foreignKey:WorkflowVersionID" json:"steps"`
}

type WorkflowStep struct {
	ID                string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID          string `gorm:"type:varchar(36);not null" json:"tenant_id"`
	WorkflowVersionID string `gorm:"type:varchar(36);not null;uniqueIndex:idx_workflow_step_ordinal" json:"workflow_version_id"`
	Ordinal           int    `gorm:"not null;uniqueIndex:idx_workflow_step_ordinal" json:"ordinal"`
	Name              string `gorm:"type:varchar(120);not null" json:"name"`
	HandoffMode       string `gorm:"type:varchar(16);not null" json:"handoff_mode"`
}

type WorkItem struct {
	ID                string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID          string          `gorm:"type:varchar(36);not null;index;uniqueIndex:idx_work_create_key" json:"tenant_id"`
	OrderID           string          `gorm:"type:varchar(36);not null" json:"order_id"`
	WorkflowVersionID string          `gorm:"type:varchar(36);not null" json:"workflow_version_id"`
	CurrentStep       int             `gorm:"not null" json:"current_step"`
	ResponsibleID     *string         `gorm:"type:varchar(255)" json:"responsible_id"`
	RecipientID       *string         `gorm:"type:varchar(255)" json:"recipient_id"`
	State             string          `gorm:"type:varchar(16);not null" json:"state"`
	Version           int64           `gorm:"not null" json:"version"`
	CreatedBy         string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_work_create_key" json:"created_by"`
	RequestKey        string          `gorm:"type:varchar(36);not null;uniqueIndex:idx_work_create_key" json:"-"`
	CreateHash        string          `gorm:"type:varchar(64);not null" json:"-"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Workflow          WorkflowVersion `gorm:"foreignKey:WorkflowVersionID" json:"workflow"`
	Events            []WorkEvent     `gorm:"foreignKey:WorkItemID" json:"events,omitempty"`
}

type WorkEvent struct {
	ID                string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID          string    `gorm:"type:varchar(36);not null" json:"tenant_id"`
	WorkItemID        string    `gorm:"type:varchar(36);not null;uniqueIndex:idx_work_event_version" json:"work_item_id"`
	Version           int64     `gorm:"not null;uniqueIndex:idx_work_event_version" json:"version"`
	ActorID           string    `gorm:"type:varchar(255);not null" json:"actor_id"`
	Action            string    `gorm:"type:varchar(32);not null" json:"action"`
	FromResponsibleID *string   `gorm:"type:varchar(255)" json:"from_responsible_id"`
	ResponsibleID     *string   `gorm:"type:varchar(255)" json:"responsible_id"`
	RecipientID       *string   `gorm:"type:varchar(255)" json:"recipient_id"`
	FromStep          int       `gorm:"not null" json:"from_step"`
	CurrentStep       int       `gorm:"not null" json:"current_step"`
	State             string    `gorm:"type:varchar(16);not null" json:"state"`
	CreatedAt         time.Time `json:"created_at"`
}
