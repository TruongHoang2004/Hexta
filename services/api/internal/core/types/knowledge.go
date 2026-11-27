package types

import (
	"errors"
	"time"
)

// Knowledge source types pushed to the AI service for grounded answers (ai_runtime knowledge store).
const (
	KnowledgeSourceProduct  = "product"
	KnowledgeSourceDocument = "document"
	KnowledgeSourceScript   = "script"

	// Knowledge sync job states. pending and leased jobs are due work; done and failed wait for the next change.
	KnowledgeSyncPending = "pending"
	KnowledgeSyncLeased  = "leased"
	KnowledgeSyncDone    = "done"
	KnowledgeSyncFailed  = "failed"

	// Knowledge document limits (Go and the AI knowledge contract: title ≤ 300, body ≤ 20000).
	KnowledgeDocumentMaxTitle = 200
	KnowledgeDocumentMaxBody  = 20000
)

// KnowledgeDocument is an owner-managed business document (FAQ, policy) the bot may answer from. Version is the
// optimistic version and the AI knowledge revision; a deleted document keeps its row (DeletedAt) so the deletion
// carries a newer revision.
type KnowledgeDocument struct {
	ID, TenantID, Title, Body string
	Version                   int64
	CreatedBy, UpdatedBy      string
	CreatedAt, UpdatedAt      time.Time
	DeletedAt                 *time.Time
	// SyncState and SyncError report the AI knowledge sync job (empty when no job exists).
	SyncState, SyncError string
}

// KnowledgeDocumentInput is a create or update; ExpectedVersion is ignored on create.
type KnowledgeDocumentInput struct {
	Title, Body     string
	ExpectedVersion int64
}

// KnowledgeDocumentPage is one page of a tenant's documents plus whether the caller may change them.
type KnowledgeDocumentPage struct {
	Items     []KnowledgeDocument
	Total     int64
	CanManage bool
}

// KnowledgeJob is one leased knowledge sync job. Fence identifies the lease; a change enqueued meanwhile advances it.
type KnowledgeJob struct {
	TenantID, SourceType, SourceID string
	Fence                          int64
	Attempts                       int
}

// KnowledgeUpsert is the authorized current state of one source as sent to the AI service. Revision is monotonic
// per source and identical on every retry of the same state.
type KnowledgeUpsert struct {
	SourceType, SourceID string
	Revision             int64
	Title, Body          string
	Attributes           map[string]string
}

// KnowledgeSource is a source's current state read by the sync job: Upsert is nil when the source must be removed
// from the AI knowledge store at Revision.
type KnowledgeSource struct {
	Upsert   *KnowledgeUpsert
	Revision int64
}

// KnowledgeProduct is a product's current catalog state for knowledge sync. Removed is true when the product is
// archived or soft-deleted. Revision is the newest change time (product, variants, inventory) in microseconds.
type KnowledgeProduct struct {
	ID, Name, Category, Description string
	Removed                         bool
	Revision                        int64
	Variants                        []KnowledgeVariant
}

// KnowledgeVariant is one sellable variant. Price is the decimal text and is meaningful only when PriceProvided;
// Stock is nil when no inventory row exists.
type KnowledgeVariant struct {
	SKU, Name, Price string
	PriceProvided    bool
	Stock            *int64
}

// KnowledgeScript is a script's current state for knowledge sync; only an active script with its active version
// content is ingested. Revision is the script version.
type KnowledgeScript struct {
	ID, Name, Status, ActiveVersionID string
	Triggers                          []string
	Answer                            string
	Revision                          int64
}

// KnowledgeRejectedError is a permanent AI ingest rejection (INVALID_ARGUMENT or FAILED_PRECONDITION); Code is a
// safe category recorded on the job.
type KnowledgeRejectedError struct{ Code string }

func (e *KnowledgeRejectedError) Error() string { return "AI knowledge ingest rejected: " + e.Code }

// ErrKnowledgeRetryable means the AI service could not take the change now (UNAVAILABLE, RESOURCE_EXHAUSTED,
// timeouts, transport); the job retries later with the same revision.
var ErrKnowledgeRetryable = errors.New("AI knowledge ingest temporarily unavailable")

// KnowledgeSyncStuckAttempts is the attempt count from which a still-retrying job is reported as stuck.
const KnowledgeSyncStuckAttempts = 5

// KnowledgeSyncStatus counts knowledge sync jobs for operators: Failed jobs were rejected by the AI service (they
// stay failed until the source changes or a re-sync), Stuck jobs keep retrying (at least
// KnowledgeSyncStuckAttempts attempts), Pending jobs are due or waiting for a retry.
type KnowledgeSyncStatus struct {
	Pending, Stuck, Failed, Done int64
}
