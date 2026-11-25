package dto

import "time"

// KnowledgeDocument is an owner-managed business document (FAQ, policy) the bot may answer from. sync_state is the
// AI knowledge sync status: pending (waiting to be indexed), done (indexed), failed (rejected; sync_error holds the
// reason code) or empty when no sync was ever recorded.
type KnowledgeDocument struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Version   int64     `json:"version"`
	SyncState string    `json:"sync_state" enums:"pending,done,failed"`
	SyncError string    `json:"sync_error,omitempty"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// KnowledgeDocumentsResponse is one page of documents, newest change first. can_manage is true for the tenant
// owner, the only member who may create, edit or delete documents.
type KnowledgeDocumentsResponse struct {
	Items     []KnowledgeDocument `json:"items"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	Size      int                 `json:"size"`
	CanManage bool                `json:"can_manage"`
}

// CreateKnowledgeDocumentRequest adds a document. Length limits count characters after trimming.
type CreateKnowledgeDocumentRequest struct {
	Title string `json:"title" validate:"required,max=200"`
	Body  string `json:"body" validate:"required,max=20000"`
}

// UpdateKnowledgeDocumentRequest rewrites a document; expected_version is the version last read.
type UpdateKnowledgeDocumentRequest struct {
	Title           string `json:"title" validate:"required,max=200"`
	Body            string `json:"body" validate:"required,max=20000"`
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
}

// KnowledgeDocumentDeleteQuery carries the version last read.
type KnowledgeDocumentDeleteQuery struct {
	ExpectedVersion int64 `form:"expected_version" validate:"min=1"`
}

// KnowledgeDocumentDeletedResponse confirms a deletion.
type KnowledgeDocumentDeletedResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}
