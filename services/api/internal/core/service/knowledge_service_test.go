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
