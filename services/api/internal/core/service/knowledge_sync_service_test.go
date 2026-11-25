package service

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

// fakeSyncStore is an in-memory knowledge sync queue with one job; outcomes are recorded, not applied.
type fakeSyncStore struct {
	job                       *types.KnowledgeJob
	product                   *types.KnowledgeProduct
	script                    *types.KnowledgeScript
	document                  *types.KnowledgeDocument
	readErr                   *apperrors.Error
	finished, retried, failed int
	revision                  int64
	failCode                  string
}

func (f *fakeSyncStore) ClaimKnowledgeJob(context.Context) (*types.KnowledgeJob, *apperrors.Error) {
	job := f.job
	f.job = nil
	return job, nil
}
func (f *fakeSyncStore) FinishKnowledgeJob(_ context.Context, _ types.KnowledgeJob, revision int64) (bool, *apperrors.Error) {
	f.finished++
	f.revision = revision
	return true, nil
}
func (f *fakeSyncStore) RetryKnowledgeJob(context.Context, types.KnowledgeJob) (bool, *apperrors.Error) {
	f.retried++
	return true, nil
}
func (f *fakeSyncStore) FailKnowledgeJob(_ context.Context, _ types.KnowledgeJob, code string) (bool, *apperrors.Error) {
	f.failed++
	f.failCode = code
	return true, nil
}
func (f *fakeSyncStore) KnowledgeProduct(context.Context, string, string) (*types.KnowledgeProduct, *apperrors.Error) {
	return f.product, f.readErr
}
func (f *fakeSyncStore) KnowledgeScript(context.Context, string, string) (*types.KnowledgeScript, *apperrors.Error) {
	return f.script, f.readErr
}
func (f *fakeSyncStore) KnowledgeDocument(context.Context, string, string) (*types.KnowledgeDocument, *apperrors.Error) {
	return f.document, f.readErr
}

// fakeKnowledgeAI records ingest calls; err applies to every call.
type fakeKnowledgeAI struct {
	upserts []types.KnowledgeUpsert
	deletes []int64
	err     error
}

func (f *fakeKnowledgeAI) UpsertKnowledge(_ context.Context, _ string, document types.KnowledgeUpsert) (string, error) {
	f.upserts = append(f.upserts, document)
	return "created", f.err
}
func (f *fakeKnowledgeAI) DeleteKnowledge(_ context.Context, _, _, _ string, revision int64) (string, error) {
	f.deletes = append(f.deletes, revision)
	return "deleted", f.err
}

func syncJob(kind string) *types.KnowledgeJob {
	return &types.KnowledgeJob{TenantID: "t", SourceType: kind, SourceID: "s1", Fence: 2, Attempts: 1}
}

func TestKnowledgeSyncBuildsProductDocumentWithAttributes(t *testing.T) {
	stock := int64(7)
	store := &fakeSyncStore{job: syncJob(types.KnowledgeSourceProduct), product: &types.KnowledgeProduct{ID: "s1", Name: "Áo thun basic", Category: "Áo",
		Description: "Cotton 100%", Revision: 1700000000000001, Variants: []types.KnowledgeVariant{
			{SKU: "AO-1", Name: "Size M", Price: "199000.00", PriceProvided: true, Stock: &stock},
			{SKU: "AO-2", Name: "Size L", Price: "0.00", PriceProvided: false},
		}}}
	ai := &fakeKnowledgeAI{}
	claimed, err := NewKnowledgeSyncService(store, ai).SyncOne(context.Background())
	if !claimed || err != nil || store.finished != 1 || store.revision != 1700000000000001 {
		t.Fatalf("claimed=%v err=%v store=%+v", claimed, err, store)
	}
	got := ai.upserts[0]
	want := map[string]string{"sku": "AO-1", "price": "199000", "stock": "7", "sku_2": "AO-2"}
	if !reflect.DeepEqual(got.Attributes, want) {
		t.Fatalf("attributes: price only when provided, stock only with inventory: %v", got.Attributes)
	}
	if got.Title != "Áo thun basic" || got.Revision != 1700000000000001 || got.Body != "Danh mục: Áo\nMô tả: Cotton 100%\nPhân loại: Size M (SKU AO-1)\nPhân loại: Size L (SKU AO-2)" {
		t.Fatalf("unexpected document %+v", got)
	}
}

func TestKnowledgeSyncRemovesArchivedMissingAndInactiveSources(t *testing.T) {
	cases := []struct {
		name     string
		store    *fakeSyncStore
		revision int64
	}{
		{"archived product", &fakeSyncStore{job: syncJob(types.KnowledgeSourceProduct), product: &types.KnowledgeProduct{ID: "s1", Removed: true, Revision: 42}}, 42},
		{"missing product", &fakeSyncStore{job: syncJob(types.KnowledgeSourceProduct)}, math.MaxInt64},
		{"inactive script", &fakeSyncStore{job: syncJob(types.KnowledgeSourceScript), script: &types.KnowledgeScript{ID: "s1", Status: types.ScriptStatusInactive, Revision: 5}}, 5},
		{"archived script", &fakeSyncStore{job: syncJob(types.KnowledgeSourceScript), script: &types.KnowledgeScript{ID: "s1", Status: types.ScriptStatusArchived, Revision: 9}}, 9},
		{"deleted document", &fakeSyncStore{job: syncJob(types.KnowledgeSourceDocument), document: &types.KnowledgeDocument{ID: "s1", Version: 3, DeletedAt: &time.Time{}}}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ai := &fakeKnowledgeAI{}
			if _, err := NewKnowledgeSyncService(tc.store, ai).SyncOne(context.Background()); err != nil {
				t.Fatal(err)
			}
			if len(ai.upserts) != 0 || len(ai.deletes) != 1 || ai.deletes[0] != tc.revision || tc.store.finished != 1 || tc.store.revision != tc.revision {
				t.Fatalf("expected one delete at %d: %+v %+v", tc.revision, ai, tc.store)
			}
		})
	}
}

func TestKnowledgeSyncActiveScriptCarriesVersionAndDocumentBody(t *testing.T) {
	store := &fakeSyncStore{job: syncJob(types.KnowledgeSourceScript), script: &types.KnowledgeScript{ID: "s1", Name: "Phí ship", Status: types.ScriptStatusActive,
		ActiveVersionID: "v3", Triggers: []string{"phí ship", "ship bao nhiêu"}, Answer: "Phí ship 30k ạ", Revision: 4}}
	ai := &fakeKnowledgeAI{}
	if _, err := NewKnowledgeSyncService(store, ai).SyncOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := ai.upserts[0]; got.Revision != 4 || got.Attributes["version_id"] != "v3" || got.Body != "Khách hỏi: phí ship; ship bao nhiêu\nTrả lời: Phí ship 30k ạ" {
		t.Fatalf("unexpected script document %+v", got)
	}
	store = &fakeSyncStore{job: syncJob(types.KnowledgeSourceDocument), document: &types.KnowledgeDocument{ID: "s1", Title: "Đổi trả", Body: "Đổi trả trong 7 ngày", Version: 2}}
	ai = &fakeKnowledgeAI{}
	if _, err := NewKnowledgeSyncService(store, ai).SyncOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := ai.upserts[0]; got.SourceType != types.KnowledgeSourceDocument || got.Revision != 2 || got.Title != "Đổi trả" || got.Body != "Đổi trả trong 7 ngày" {
		t.Fatalf("unexpected knowledge document %+v", got)
	}
}

func TestKnowledgeSyncRetriesTransientAndRecordsPermanentFailures(t *testing.T) {
	document := &types.KnowledgeDocument{ID: "s1", Title: "FAQ", Body: "x", Version: 1}
	for _, transient := range []error{types.ErrKnowledgeRetryable, types.ErrAIUnavailable, errors.New("transport")} {
		store := &fakeSyncStore{job: syncJob(types.KnowledgeSourceDocument), document: document}
		claimed, err := NewKnowledgeSyncService(store, &fakeKnowledgeAI{err: transient}).SyncOne(context.Background())
		if !claimed || err == nil || store.retried != 1 || store.finished != 0 || store.failed != 0 {
			t.Fatalf("%v must be retried with the same revision: %+v", transient, store)
		}
	}
	store := &fakeSyncStore{job: syncJob(types.KnowledgeSourceDocument), document: document}
	if _, err := NewKnowledgeSyncService(store, &fakeKnowledgeAI{err: &types.KnowledgeRejectedError{Code: "invalid_argument"}}).SyncOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.failed != 1 || store.failCode != "invalid_argument" || store.retried != 0 {
		t.Fatalf("a rejection is permanent with its reason: %+v", store)
	}
	// A source that cannot be read is retried without calling the AI service.
	ai := &fakeKnowledgeAI{}
	store = &fakeSyncStore{job: syncJob(types.KnowledgeSourceDocument), readErr: apperrors.ErrSystemError(context.Background(), "db down")}
	if _, err := NewKnowledgeSyncService(store, ai).SyncOne(context.Background()); err == nil || store.retried != 1 || len(ai.upserts)+len(ai.deletes) != 0 {
		t.Fatalf("read failure must retry without AI calls: %+v", store)
	}
	// No due job: nothing claimed.
	if claimed, err := NewKnowledgeSyncService(&fakeSyncStore{}, ai).SyncOne(context.Background()); claimed || err != nil {
		t.Fatal(claimed, err)
	}
}
