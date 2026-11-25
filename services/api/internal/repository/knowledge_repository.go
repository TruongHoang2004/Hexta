package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/knowledge/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/knowledgemodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// knowledgeLeaseSeconds bounds one sync attempt (two short AI calls); an expired lease makes the job due again.
	knowledgeLeaseSeconds = 60
	// knowledgeMaxBackoffSeconds caps the retry delay of a job the AI service could not take.
	knowledgeMaxBackoffSeconds = 600
)

// KnowledgeDocumentStore persists owner-managed knowledge documents. Authorization stays in the service.
type KnowledgeDocumentStore interface {
	WithinKnowledge(context.Context, func(KnowledgeDocumentStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	Documents(ctx context.Context, tenant string, limit, offset int) ([]types.KnowledgeDocument, int64, *apperrors.Error)
	// Document returns a live (not deleted) document, locked when lock is true.
	Document(ctx context.Context, tenant, id string, lock bool) (*types.KnowledgeDocument, *apperrors.Error)
	InsertDocument(context.Context, types.KnowledgeDocument) *apperrors.Error
	// SaveDocument updates the document only if it is still at value.Version-1.
	SaveDocument(context.Context, types.KnowledgeDocument) *apperrors.Error
	EnqueueKnowledgeSync(ctx context.Context, tenant, sourceType, sourceID string, at time.Time) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
}

// KnowledgeSyncStore is the durable knowledge sync job queue and the current-state source reads it needs. A lease
// is a committed fence: no transaction or row lock is held while the job calls the AI service.
type KnowledgeSyncStore interface {
	// ClaimKnowledgeJob leases the oldest due job (including an expired lease); nil when none is due.
	ClaimKnowledgeJob(context.Context) (*types.KnowledgeJob, *apperrors.Error)
	// FinishKnowledgeJob, RetryKnowledgeJob and FailKnowledgeJob record an outcome only while the lease is held and
	// no newer change was enqueued; they report false otherwise (the job stays due for the newer change).
	FinishKnowledgeJob(ctx context.Context, job types.KnowledgeJob, revision int64) (bool, *apperrors.Error)
	RetryKnowledgeJob(ctx context.Context, job types.KnowledgeJob) (bool, *apperrors.Error)
	FailKnowledgeJob(ctx context.Context, job types.KnowledgeJob, code string) (bool, *apperrors.Error)
	// KnowledgeProduct, KnowledgeScript and KnowledgeDocument return nil when the row does not exist.
	KnowledgeProduct(ctx context.Context, tenant, id string) (*types.KnowledgeProduct, *apperrors.Error)
	KnowledgeScript(ctx context.Context, tenant, id string) (*types.KnowledgeScript, *apperrors.Error)
	KnowledgeDocument(ctx context.Context, tenant, id string) (*types.KnowledgeDocument, *apperrors.Error)
	// KnowledgeSyncStatus counts jobs by outcome across tenants, for operator logs (no identifiers).
	KnowledgeSyncStatus(ctx context.Context) (types.KnowledgeSyncStatus, *apperrors.Error)
}

// knowledgeSyncCoalesce keeps one job per source: a new change makes the existing job pending and due now, and
// advances its fence so a lease held on an older state can no longer finish it. The worker reads the current state
// (and its revision) after commit, so coalesced changes are synced once at the newest revision.
const knowledgeSyncCoalesce = ` ON CONFLICT (tenant_id, source_type, source_id) DO UPDATE SET state='pending', attempts=0,
	fence=knowledge_sync_jobs.fence+1, available_at=EXCLUDED.available_at, last_error=NULL, updated_at=EXCLUDED.updated_at`

// EnqueueKnowledgeSync records in the caller's transaction that a knowledge source changed (transactional outbox).
func (r *SalesRepository) EnqueueKnowledgeSync(ctx context.Context, tenant, sourceType, sourceID string, at time.Time) *apperrors.Error {
	err := r.db.WithContext(ctx).Exec(`INSERT INTO knowledge_sync_jobs (tenant_id, source_type, source_id, state, attempts, fence, available_at, updated_at)
		VALUES (?, ?, ?, 'pending', 0, 1, ?, ?)`+knowledgeSyncCoalesce, tenant, sourceType, sourceID, at, at).Error
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// enqueueVariantProductSync records, in the caller's transaction, that the stock of a variant changed: its product's
// knowledge (the "stock" attribute) must be synced. Same tenant only; no job when the variant does not exist.
func enqueueVariantProductSync(tx *gorm.DB, tenant, variantID string, at time.Time) error {
	return tx.Exec(`INSERT INTO knowledge_sync_jobs (tenant_id, source_type, source_id, state, attempts, fence, available_at, updated_at)
		SELECT v.tenant_id, ?, v.product_id, 'pending', 0, 1, ?, ? FROM product_variants v WHERE v.tenant_id = ? AND v.id = ?`+
		knowledgeSyncCoalesce, types.KnowledgeSourceProduct, at, at, tenant, variantID).Error
}

// KnowledgeRepository stores knowledge documents and runs the knowledge sync queue.
type KnowledgeRepository struct{ *SalesRepository }

func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{NewSalesRepository(db)}
}

func (r *KnowledgeRepository) WithinKnowledge(ctx context.Context, fn func(KnowledgeDocumentStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if failure = fn(NewKnowledgeRepository(tx)); failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *KnowledgeRepository) Scope(ctx context.Context, tenant, actor string) (types.MessagingScope, *apperrors.Error) {
	workspace, err := r.Tenant(ctx, tenant)
	if err != nil {
		return types.MessagingScope{}, err
	}
	member, err := r.Member(ctx, tenant, actor)
	if err != nil {
		return types.MessagingScope{}, err
	}
	return types.MessagingScope{Active: workspace.Status == model.TenantStatusActive, Member: member != nil, Owner: workspace.OwnerID == actor}, nil
}

func (r *KnowledgeRepository) Documents(ctx context.Context, tenant string, limit, offset int) ([]types.KnowledgeDocument, int64, *apperrors.Error) {
	var total int64
	db := r.db.WithContext(ctx)
	if err := db.Model(&knowledgemodel.Document{}).Where("tenant_id = ? AND deleted_at IS NULL", tenant).Count(&total).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	rows := []knowledgemodel.DocumentSync{}
	if err := db.Table("knowledge_documents AS d").Select("d.*, j.state AS sync_state, j.last_error AS sync_error").
		Joins("LEFT JOIN knowledge_sync_jobs j ON j.tenant_id = d.tenant_id AND j.source_type = ? AND j.source_id = d.id", types.KnowledgeSourceDocument).
		Where("d.tenant_id = ? AND d.deleted_at IS NULL", tenant).Order("d.updated_at DESC, d.id").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, salesDBError(ctx, err)
	}
	result := make([]types.KnowledgeDocument, len(rows))
	for i, row := range rows {
		result[i] = persistencemap.DocumentWithSync(row)
	}
	return result, total, nil
}

func (r *KnowledgeRepository) Document(ctx context.Context, tenant, id string, lock bool) (*types.KnowledgeDocument, *apperrors.Error) {
	var row knowledgemodel.Document
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&row).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound(ctx, "knowledge document", "not found")
		}
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Document(row)
	return &value, nil
}

func (r *KnowledgeRepository) InsertDocument(ctx context.Context, value types.KnowledgeDocument) *apperrors.Error {
	row := persistencemap.DocumentRow(value)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *KnowledgeRepository) SaveDocument(ctx context.Context, value types.KnowledgeDocument) *apperrors.Error {
	row := persistencemap.DocumentRow(value)
	result := r.db.WithContext(ctx).Model(&knowledgemodel.Document{}).
		Where("tenant_id = ? AND id = ? AND version = ? AND deleted_at IS NULL", value.TenantID, value.ID, value.Version-1).
		Select("title", "body", "version", "updated_by", "updated_at", "deleted_at").Updates(&row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "knowledge document", "changed; refresh before retrying")
	}
	return nil
}

func knowledgeJobError(ctx context.Context) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "knowledge sync persistence operation failed")
}

func (r *KnowledgeRepository) ClaimKnowledgeJob(ctx context.Context) (*types.KnowledgeJob, *apperrors.Error) {
	var rows []knowledgemodel.SyncJob
	err := r.db.WithContext(ctx).Raw(`UPDATE knowledge_sync_jobs SET state='leased', fence=fence+1, attempts=attempts+1,
		available_at=clock_timestamp()+make_interval(secs=>?), updated_at=clock_timestamp()
		WHERE (tenant_id, source_type, source_id) = (SELECT tenant_id, source_type, source_id FROM knowledge_sync_jobs
			WHERE state IN ('pending','leased') AND available_at<=clock_timestamp()
			ORDER BY available_at, tenant_id, source_type, source_id FOR UPDATE SKIP LOCKED LIMIT 1)
		RETURNING *`, knowledgeLeaseSeconds).Scan(&rows).Error
	if err != nil {
		return nil, knowledgeJobError(ctx)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	job := persistencemap.Job(rows[0])
	return &job, nil
}

func (r *KnowledgeRepository) finishJob(ctx context.Context, job types.KnowledgeJob, set string, args ...any) (bool, *apperrors.Error) {
	args = append(args, job.TenantID, job.SourceType, job.SourceID, job.Fence)
	result := r.db.WithContext(ctx).Exec(`UPDATE knowledge_sync_jobs SET `+set+`, updated_at=clock_timestamp()
		WHERE tenant_id=? AND source_type=? AND source_id=? AND fence=? AND state='leased'`, args...)
	if result.Error != nil {
		return false, knowledgeJobError(ctx)
	}
	return result.RowsAffected == 1, nil
}

func (r *KnowledgeRepository) FinishKnowledgeJob(ctx context.Context, job types.KnowledgeJob, revision int64) (bool, *apperrors.Error) {
	return r.finishJob(ctx, job, "state='done', attempts=0, last_error=NULL, revision=?", revision)
}

// RetryKnowledgeJob backs off exponentially from 5 seconds, capped at knowledgeMaxBackoffSeconds.
func (r *KnowledgeRepository) RetryKnowledgeJob(ctx context.Context, job types.KnowledgeJob) (bool, *apperrors.Error) {
	return r.finishJob(ctx, job, "state='pending', available_at=clock_timestamp()+make_interval(secs=>LEAST(?, 5*power(2, LEAST(attempts, 10))::int))", knowledgeMaxBackoffSeconds)
}

func (r *KnowledgeRepository) FailKnowledgeJob(ctx context.Context, job types.KnowledgeJob, code string) (bool, *apperrors.Error) {
	return r.finishJob(ctx, job, "state='failed', last_error=?", code)
}

func (r *KnowledgeRepository) KnowledgeProduct(ctx context.Context, tenant, id string) (*types.KnowledgeProduct, *apperrors.Error) {
	var rows []knowledgemodel.Product
	db := r.db.WithContext(ctx)
	if err := db.Raw(`SELECT id, name, category, description, status, updated_at, deleted_at FROM products WHERE tenant_id = ? AND id = ?`, tenant, id).Scan(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	// Bounded: a product has few variants; the AI contract keeps at most 30 attributes.
	variants := []knowledgemodel.Variant{}
	if err := db.Raw(`SELECT v.sku, v.variant_name, v.price::text AS price, v.price_provided, v.updated_at,
			i.available_qty, i.updated_at AS inventory_updated_at
		FROM product_variants v LEFT JOIN inventory_items i ON i.tenant_id = v.tenant_id AND i.variant_id = v.id
		WHERE v.tenant_id = ? AND v.product_id = ? AND v.deleted_at IS NULL ORDER BY v.created_at, v.id LIMIT 10`, tenant, id).Scan(&variants).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Product(rows[0], variants, string(model.ProductStatusActive))
	return &value, nil
}

func (r *KnowledgeRepository) KnowledgeScript(ctx context.Context, tenant, id string) (*types.KnowledgeScript, *apperrors.Error) {
	var rows []knowledgemodel.Script
	if err := r.db.WithContext(ctx).Raw(`SELECT s.id, s.name, s.status, s.active_version_id, s.version, v.triggers::text AS triggers, v.answer
		FROM scripts s LEFT JOIN script_versions v ON v.tenant_id = s.tenant_id AND v.script_id = s.id AND v.id = s.active_version_id
		WHERE s.tenant_id = ? AND s.id = ?`, tenant, id).Scan(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	value, err := persistencemap.Script(rows[0])
	if err != nil {
		return nil, storedContentError(ctx)
	}
	return &value, nil
}

func (r *KnowledgeRepository) KnowledgeDocument(ctx context.Context, tenant, id string) (*types.KnowledgeDocument, *apperrors.Error) {
	var rows []knowledgemodel.Document
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id).Limit(1).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	value := persistencemap.Document(rows[0])
	return &value, nil
}

func (r *KnowledgeRepository) KnowledgeSyncStatus(ctx context.Context) (types.KnowledgeSyncStatus, *apperrors.Error) {
	var status types.KnowledgeSyncStatus
	err := r.db.WithContext(ctx).Raw(`SELECT
			count(*) FILTER (WHERE state IN ('pending','leased') AND attempts < ?) AS pending,
			count(*) FILTER (WHERE state IN ('pending','leased') AND attempts >= ?) AS stuck,
			count(*) FILTER (WHERE state = 'failed') AS failed,
			count(*) FILTER (WHERE state = 'done') AS done
		FROM knowledge_sync_jobs`, types.KnowledgeSyncStuckAttempts, types.KnowledgeSyncStuckAttempts).Scan(&status).Error
	if err != nil {
		return status, knowledgeJobError(ctx)
	}
	return status, nil
}

// EnqueueKnowledgeResync enqueues every live knowledge source (active products, active scripts, knowledge documents)
// of one tenant, or of every tenant when tenant is empty, in one transaction. Used after switching the embedding
// model: the AI service re-embeds a source whose stored chunks carry another model. Existing jobs coalesce (failed
// ones become pending again). It reports how many sources were enqueued.
func (r *KnowledgeRepository) EnqueueKnowledgeResync(ctx context.Context, tenant string, at time.Time) (int64, *apperrors.Error) {
	sources := []struct{ kind, query string }{
		{types.KnowledgeSourceProduct, `SELECT tenant_id, id FROM products WHERE status = 'active' AND deleted_at IS NULL`},
		{types.KnowledgeSourceScript, `SELECT tenant_id, id FROM scripts WHERE status = 'active'`},
		{types.KnowledgeSourceDocument, `SELECT tenant_id, id FROM knowledge_documents WHERE deleted_at IS NULL`},
	}
	var total int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, source := range sources {
			result := tx.Exec(`INSERT INTO knowledge_sync_jobs (tenant_id, source_type, source_id, state, attempts, fence, available_at, updated_at)
				SELECT s.tenant_id, ?, s.id, 'pending', 0, 1, ?, ? FROM (`+source.query+`) s WHERE ? = '' OR s.tenant_id = ?`+
				knowledgeSyncCoalesce, source.kind, at, at, tenant, tenant)
			if result.Error != nil {
				return result.Error
			}
			total += result.RowsAffected
		}
		return nil
	})
	if err != nil {
		return 0, knowledgeJobError(ctx)
	}
	return total, nil
}

var (
	_ KnowledgeDocumentStore = (*KnowledgeRepository)(nil)
	_ KnowledgeSyncStore     = (*KnowledgeRepository)(nil)
)
