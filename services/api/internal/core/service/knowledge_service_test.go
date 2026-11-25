package service

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/knowledgemodel"
	"github.com/shopspring/decimal"
)

func syncRow(t *testing.T, f salesFixture, kind, id string) knowledgemodel.SyncJob {
	t.Helper()
	var row knowledgemodel.SyncJob
	if err := f.db.Where("tenant_id = ? AND source_type = ? AND source_id = ?", f.tenant, kind, id).First(&row).Error; err != nil {
		t.Fatalf("expected a %s sync job: %v", kind, err)
	}
	return row
}

func TestKnowledgeDocumentsOwnerWritesMembersRead(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	svc := NewKnowledgeService(repository.NewKnowledgeRepository(f.db))

	_, err := svc.Create(ctx, f.tenant, f.sale, types.KnowledgeDocumentInput{Title: "Đổi trả", Body: "7 ngày"})
	requireSalesStatus(t, err, http.StatusForbidden)
	_, err = svc.Create(ctx, f.tenant, f.owner, types.KnowledgeDocumentInput{Title: "  ", Body: "x"})
	requireSalesStatus(t, err, http.StatusBadRequest)
	_, err = svc.Create(ctx, f.tenant, f.owner, types.KnowledgeDocumentInput{Title: "x", Body: strings.Repeat("a", types.KnowledgeDocumentMaxBody+1)})
	requireSalesStatus(t, err, http.StatusBadRequest)

	doc, err := svc.Create(ctx, f.tenant, f.owner, types.KnowledgeDocumentInput{Title: " Chính sách đổi trả ", Body: "Đổi trả trong 7 ngày."})
	if err != nil || doc.Title != "Chính sách đổi trả" || doc.Version != 1 || doc.SyncState != types.KnowledgeSyncPending {
		t.Fatalf("create: %+v %v", doc, err)
	}
	if job := syncRow(t, f, types.KnowledgeSourceDocument, doc.ID); job.State != types.KnowledgeSyncPending || job.Fence != 1 {
		t.Fatalf("create must enqueue the document: %+v", job)
	}

	page, err := svc.List(ctx, f.tenant, f.sale, 1, 20)
	if err != nil || page.Total != 1 || page.CanManage || page.Items[0].SyncState != types.KnowledgeSyncPending {
		t.Fatalf("a member reads documents without manage rights: %+v %v", page, err)
	}
	if page, _ := svc.List(ctx, f.tenant, f.owner, 1, 20); !page.CanManage {
		t.Fatal("the owner manages documents")
	}
	_, err = svc.List(ctx, f.other, f.sale, 1, 20)
	requireSalesStatus(t, err, http.StatusForbidden)
	_, err = svc.Update(ctx, f.other, f.owner, doc.ID, types.KnowledgeDocumentInput{Title: "x", Body: "y", ExpectedVersion: 1})
	requireSalesStatus(t, err, http.StatusNotFound)

	_, err = svc.Update(ctx, f.tenant, f.sale, doc.ID, types.KnowledgeDocumentInput{Title: "x", Body: "y", ExpectedVersion: 1})
	requireSalesStatus(t, err, http.StatusForbidden)
	updated, err := svc.Update(ctx, f.tenant, f.owner, doc.ID, types.KnowledgeDocumentInput{Title: "Đổi trả", Body: "Đổi trả trong 14 ngày.", ExpectedVersion: 1})
	if err != nil || updated.Version != 2 {
		t.Fatalf("update: %+v %v", updated, err)
	}
	if job := syncRow(t, f, types.KnowledgeSourceDocument, doc.ID); job.Fence != 2 || job.State != types.KnowledgeSyncPending {
		t.Fatalf("update must re-enqueue: %+v", job)
	}
	_, err = svc.Update(ctx, f.tenant, f.owner, doc.ID, types.KnowledgeDocumentInput{Title: "x", Body: "y", ExpectedVersion: 1})
	requireSalesStatus(t, err, http.StatusConflict)

	requireSalesStatus(t, svc.Delete(ctx, f.tenant, f.owner, doc.ID, 1), http.StatusConflict)
	requireSalesStatus(t, svc.Delete(ctx, f.tenant, f.sale, doc.ID, 2), http.StatusForbidden)
	if err := svc.Delete(ctx, f.tenant, f.owner, doc.ID, 2); err != nil {
		t.Fatal(err)
	}
	if page, _ := svc.List(ctx, f.tenant, f.owner, 1, 20); page.Total != 0 {
		t.Fatal("a deleted document is not listed")
	}
	// The tombstone keeps a newer version so the AI removal outranks every earlier upsert.
	stored, err := repository.NewKnowledgeRepository(f.db).KnowledgeDocument(ctx, f.tenant, doc.ID)
	if err != nil || stored.DeletedAt == nil || stored.Version != 3 || syncRow(t, f, types.KnowledgeSourceDocument, doc.ID).Fence != 3 {
		t.Fatalf("delete must leave a versioned tombstone and enqueue: %+v %v", stored, err)
	}
	if auditCount(t, f.db, f.tenant, doc.ID, "knowledge.document_deleted") != 1 {
		t.Fatal("deletion must be audited")
	}
}

func TestKnowledgeSyncQueueAgainstPostgres(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	repo := repository.NewKnowledgeRepository(f.db)
	price := decimal.RequireFromString("199000")
	product, err := f.service.SaveProduct(ctx, f.tenant, f.sale, "", types.ProductInput{Name: "Áo thun", SKU: "AO-1", VariantName: "M", Price: &price, Category: "Áo"})
	if err != nil {
		t.Fatal(err)
	}
	if job := syncRow(t, f, types.KnowledgeSourceProduct, product.ID); job.State != types.KnowledgeSyncPending {
		t.Fatalf("product save must enqueue: %+v", job)
	}
	current, err := repo.KnowledgeProduct(ctx, f.tenant, product.ID)
	if err != nil || current.Removed || current.Revision <= 0 || len(current.Variants) != 1 || current.Variants[0].Price != "199000.00" || !current.Variants[0].PriceProvided {
		t.Fatalf("unexpected product state %+v %v", current, err)
	}
	if other, err := repo.KnowledgeProduct(ctx, f.other, product.ID); err != nil || other != nil {
		t.Fatal("another tenant must not read the product", other, err)
	}

	job, err := repo.ClaimKnowledgeJob(ctx)
	if err != nil || job == nil || job.SourceID != product.ID || job.Attempts != 1 {
		t.Fatalf("claim: %+v %v", job, err)
	}
	if again, _ := repo.ClaimKnowledgeJob(ctx); again != nil {
		t.Fatal("a leased job is not claimed twice", again)
	}
	// A change while the job is leased fences the lease: the old attempt cannot finish the newer change.
	if err := repo.EnqueueKnowledgeSync(ctx, f.tenant, types.KnowledgeSourceProduct, product.ID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.FinishKnowledgeJob(ctx, *job, current.Revision); err != nil || ok {
		t.Fatal("a fenced lease must not finish the job", ok, err)
	}
	job, err = repo.ClaimKnowledgeJob(ctx)
	if err != nil || job == nil {
		t.Fatal("the newer change is due again", err)
	}
	if ok, err := repo.RetryKnowledgeJob(ctx, *job); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if row := syncRow(t, f, types.KnowledgeSourceProduct, product.ID); row.State != types.KnowledgeSyncPending || !row.AvailableAt.After(time.Now().UTC()) {
		t.Fatalf("a retry backs off: %+v", row)
	}
	if err := f.db.Exec("UPDATE knowledge_sync_jobs SET available_at = now() - interval '1 second' WHERE source_id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	job, _ = repo.ClaimKnowledgeJob(ctx)
	if ok, err := repo.FailKnowledgeJob(ctx, *job, "invalid_argument"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if row := syncRow(t, f, types.KnowledgeSourceProduct, product.ID); row.State != types.KnowledgeSyncFailed || deref(row.LastError) != "invalid_argument" {
		t.Fatalf("a rejection is recorded: %+v", row)
	}

	// Archiving re-enqueues and reads as removed with a newer revision.
	if err := f.service.ArchiveProduct(ctx, f.tenant, f.sale, product.ID); err != nil {
		t.Fatal(err)
	}
	archived, err := repo.KnowledgeProduct(ctx, f.tenant, product.ID)
	if err != nil || !archived.Removed || archived.Revision <= current.Revision {
		t.Fatalf("archived product must be removed at a newer revision: %+v %v", archived, err)
	}
	job, _ = repo.ClaimKnowledgeJob(ctx)
	if job == nil {
		t.Fatal("archive must make the job due")
	}
	if ok, err := repo.FinishKnowledgeJob(ctx, *job, archived.Revision); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if row := syncRow(t, f, types.KnowledgeSourceProduct, product.ID); row.State != types.KnowledgeSyncDone || row.Revision == nil || *row.Revision != archived.Revision || row.LastError != nil {
		t.Fatalf("finished job: %+v", row)
	}
}

// Owner decision 2026-11-26: every stock change syncs the product's knowledge immediately, through the same
// coalesced outbox row (one job per product) in the stock change's own transaction.
func TestStockChangesEnqueueOneCoalescedProductSync(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	repo := repository.NewKnowledgeRepository(f.db)
	inventory := NewInventoryService(repository.NewInventoryDBRepository(f.db))
	price := decimal.RequireFromString("299000")
	product, err := f.service.SaveProduct(ctx, f.tenant, f.sale, "", types.ProductInput{Name: "Áo polo", SKU: "POLO-1", VariantName: "M", Price: &price})
	if err != nil {
		t.Fatal(err)
	}
	var variant string
	if err := f.db.Raw("SELECT id FROM product_variants WHERE tenant_id = ? AND product_id = ?", f.tenant, product.ID).Scan(&variant).Error; err != nil || variant == "" {
		t.Fatal("variant", err)
	}
	// available_at comes from the API host clock; tolerate a container clock that is a few milliseconds behind.
	claim := func() *types.KnowledgeJob {
		t.Helper()
		for i := 0; i < 40; i++ {
			job, err := repo.ClaimKnowledgeJob(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if job != nil {
				return job
			}
			time.Sleep(25 * time.Millisecond)
		}
		return nil
	}
	// Drain the product-save job so each stock change is observed on its own.
	finish := func() {
		t.Helper()
		job := claim()
		if job == nil || job.SourceID != product.ID {
			t.Fatalf("claim: %+v", job)
		}
		current, _ := repo.KnowledgeProduct(ctx, f.tenant, product.ID)
		if ok, err := repo.FinishKnowledgeJob(ctx, *job, current.Revision); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	jobs := func() []knowledgemodel.SyncJob {
		t.Helper()
		var rows []knowledgemodel.SyncJob
		if err := f.db.Where("source_type = ? AND source_id = ?", types.KnowledgeSourceProduct, product.ID).Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		return rows
	}
	finish()
	lastRevision := int64(0)
	expectSynced := func(step string, stock int64) {
		t.Helper()
		rows := jobs()
		if len(rows) != 1 || rows[0].TenantID != f.tenant || rows[0].State != types.KnowledgeSyncPending {
			t.Fatalf("%s must leave exactly one pending job for the product: %+v", step, rows)
		}
		current, err := repo.KnowledgeProduct(ctx, f.tenant, product.ID)
		if err != nil || len(current.Variants) != 1 || current.Variants[0].Stock == nil || *current.Variants[0].Stock != stock {
			t.Fatalf("%s: unexpected knowledge state %+v %v", step, current, err)
		}
		if current.Revision <= lastRevision {
			t.Fatalf("%s: revision must increase (%d after %d)", step, current.Revision, lastRevision)
		}
		lastRevision = current.Revision
		finish()
	}

	if _, err := inventory.InitInventory(ctx, f.tenant, variant, 10, 2); err != nil {
		t.Fatal(err)
	}
	expectSynced("init", 10)
	if _, err := inventory.ReserveStock(ctx, f.tenant, variant, 3); err != nil {
		t.Fatal(err)
	}
	expectSynced("reserve", 7)
	if _, err := inventory.ReleaseStock(ctx, f.tenant, variant, 1); err != nil {
		t.Fatal(err)
	}
	expectSynced("release", 8)
	ref := "order-1"
	if _, err := inventory.DeductStock(ctx, f.tenant, variant, 2, "order", &ref); err != nil {
		t.Fatal(err)
	}
	expectSynced("deduct", 8)
	if _, err := inventory.AdjustStock(ctx, f.tenant, variant, 20, "recount"); err != nil {
		t.Fatal(err)
	}
	expectSynced("adjust", 20)

	// Several changes before the worker runs coalesce into the same single row; the fence fences an older lease.
	leased, err := repo.ClaimKnowledgeJob(ctx)
	if err != nil || leased != nil {
		t.Fatal("nothing is due after the last sync", leased, err)
	}
	for i := 0; i < 3; i++ {
		if _, err := inventory.ReserveStock(ctx, f.tenant, variant, 1); err != nil {
			t.Fatal(err)
		}
	}
	if rows := jobs(); len(rows) != 1 || rows[0].State != types.KnowledgeSyncPending {
		t.Fatalf("changes coalesce into one pending job: %+v", rows)
	}
	if leased = claim(); leased == nil {
		t.Fatal("the coalesced job is due")
	}
	if _, err := inventory.ReleaseStock(ctx, f.tenant, variant, 1); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.FinishKnowledgeJob(ctx, *leased, lastRevision); err != nil || ok {
		t.Fatal("a stock change during the lease must keep the job due", ok, err)
	}
	finish()

	// A rejected change (insufficient stock) and another tenant's call enqueue nothing.
	before := jobs()[0]
	if _, err := inventory.ReserveStock(ctx, f.tenant, variant, 1000); err == nil {
		t.Fatal("over-reservation must fail")
	}
	if _, err := inventory.ReserveStock(ctx, f.other, variant, 1); err == nil {
		t.Fatal("another tenant has no inventory for this variant")
	}
	if after := jobs(); len(after) != 1 || after[0].Fence != before.Fence || after[0].State != types.KnowledgeSyncDone {
		t.Fatalf("a failed change must not enqueue: %+v -> %+v", before, after)
	}
}

// The re-sync command (embedding model switch) enqueues live sources of one tenant or all tenants and revives
// failed jobs; archived products and other tenants are untouched; the status counts failed and stuck jobs.
func TestKnowledgeResyncAndStatusAgainstPostgres(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	repo := repository.NewKnowledgeRepository(f.db)
	docs := NewKnowledgeService(repo)
	price := decimal.RequireFromString("10000")
	live, err := f.service.SaveProduct(ctx, f.tenant, f.sale, "", types.ProductInput{Name: "Áo", SKU: "R-1", VariantName: "M", Price: &price})
	if err != nil {
		t.Fatal(err)
	}
	archived, err := f.service.SaveProduct(ctx, f.tenant, f.sale, "", types.ProductInput{Name: "Mũ", SKU: "R-2", VariantName: "M", Price: &price})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.service.ArchiveProduct(ctx, f.tenant, f.sale, archived.ID); err != nil {
		t.Fatal(err)
	}
	doc, err := docs.Create(ctx, f.tenant, f.owner, types.KnowledgeDocumentInput{Title: "Đổi trả", Body: "7 ngày"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := f.service.SaveProduct(ctx, f.other, f.owner, "", types.ProductInput{Name: "Khác", SKU: "R-3", VariantName: "M", Price: &price})
	if err != nil {
		t.Fatal(err)
	}
	// Settle every job: the live product failed permanently, the others are done, the other tenant's keeps retrying.
	if err := f.db.Exec("UPDATE knowledge_sync_jobs SET state = 'done', fence = fence + 1").Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Exec("UPDATE knowledge_sync_jobs SET state = 'failed', last_error = 'invalid_argument' WHERE source_id = ?", live.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Exec("UPDATE knowledge_sync_jobs SET state = 'pending', attempts = ? WHERE source_id = ?", types.KnowledgeSyncStuckAttempts, other.ID).Error; err != nil {
		t.Fatal(err)
	}
	if status, err := repo.KnowledgeSyncStatus(ctx); err != nil || status.Failed != 1 || status.Stuck != 1 || status.Pending != 0 || status.Done != 2 {
		t.Fatalf("status before re-sync: %+v %v", status, err)
	}

	count, err := repo.EnqueueKnowledgeResync(ctx, f.tenant, time.Now().UTC())
	if err != nil || count != 2 {
		t.Fatalf("one tenant: the live product and the document, not the archived product: %d %v", count, err)
	}
	for _, id := range []string{live.ID, doc.ID} {
		kind := types.KnowledgeSourceProduct
		if id == doc.ID {
			kind = types.KnowledgeSourceDocument
		}
		if row := syncRow(t, f, kind, id); row.State != types.KnowledgeSyncPending || row.Attempts != 0 || row.LastError != nil {
			t.Fatalf("re-synced source must be pending again: %+v", row)
		}
	}
	if row := syncRow(t, f, types.KnowledgeSourceProduct, archived.ID); row.State != types.KnowledgeSyncDone {
		t.Fatalf("archived product is not re-synced: %+v", row)
	}
	var otherRow knowledgemodel.SyncJob
	if err := f.db.Where("source_id = ?", other.ID).First(&otherRow).Error; err != nil || otherRow.Attempts != types.KnowledgeSyncStuckAttempts {
		t.Fatalf("another tenant is untouched: %+v %v", otherRow, err)
	}
	if count, err := repo.EnqueueKnowledgeResync(ctx, "", time.Now().UTC()); err != nil || count != 3 {
		t.Fatalf("all tenants: %d %v", count, err)
	}
	if status, err := repo.KnowledgeSyncStatus(ctx); err != nil || status.Pending != 3 || status.Failed != 0 || status.Stuck != 0 {
		t.Fatalf("status after re-sync: %+v %v", status, err)
	}
}
