// Package knowledgemodel holds the GORM rows of knowledge documents, knowledge sync jobs and the read projections
// the knowledge sync job takes of products and scripts.
package knowledgemodel

import "time"

type Document struct {
	ID        string
	TenantID  string
	Title     string
	Body      string
	Version   int64
	CreatedBy string
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (Document) TableName() string { return "knowledge_documents" }

// DocumentSync is a document with its knowledge sync job state (NULL when no job exists).
type DocumentSync struct {
	Document
	SyncState *string
	SyncError *string
}

type SyncJob struct {
	TenantID    string
	SourceType  string
	SourceID    string
	State       string
	Attempts    int
	Fence       int64
	AvailableAt time.Time
	LastError   *string
	Revision    *int64
	UpdatedAt   time.Time
}

func (SyncJob) TableName() string { return "knowledge_sync_jobs" }

// Product is the product projection read by the sync job.
type Product struct {
	ID          string
	Name        string
	Category    *string
	Description *string
	Status      string
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}

// Variant is a live variant with its inventory row (NULL columns when none).
type Variant struct {
	SKU               string
	VariantName       string
	Price             string
	PriceProvided     bool
	UpdatedAt         *time.Time
	AvailableQty      *int64
	InventoryUpdateAt *time.Time `gorm:"column:inventory_updated_at"`
}

// Script is the script projection with its active version content (NULL when inactive).
type Script struct {
	ID              string
	Name            string
	Status          string
	ActiveVersionID *string
	Version         int64
	Triggers        *string
	Answer          *string
}
