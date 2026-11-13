package service

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"testing"
	"time"
)

type syncMemory struct {
	job    SyncJob
	page   plugin.Page
	reason string
}

func (s *syncMemory) ClaimSync(context.Context) (SyncJob, error) { return s.job, nil }
func (s *syncMemory) CommitPage(_ context.Context, _ SyncJob, p plugin.Page) error {
	s.page = p
	return nil
}
func (s *syncMemory) FailSync(_ context.Context, _ SyncJob, reason string, _ time.Duration) error {
	s.reason = reason
	return nil
}

type fixtureFetcher struct {
	plugin.BasePlugin
	failure error
}

func (fixtureFetcher) Platform() string                  { return "fixture" }
func (fixtureFetcher) Capabilities() plugin.Capabilities { return plugin.Capabilities{Fetch: true} }
func (f fixtureFetcher) Fetch(_ context.Context, r plugin.FetchRequest) (plugin.Page, error) {
	if r.Cursor != "cursor" || r.Limit != 10 {
		return plugin.Page{}, errors.New("request mismatch")
	}
	return plugin.Page{NextCursor: "next", RetryAfter: time.Minute, Items: []plugin.Item{{AccountID: r.AccountID}}}, f.failure
}
func TestSyncContract(t *testing.T) {
	for _, mode := range []string{"history", "incremental"} {
		store := &syncMemory{job: SyncJob{Platform: "fixture", AccountID: "page", Cursor: "cursor", Mode: mode, Limit: 10}}
		r, _ := plugin.NewRegistry(fixtureFetcher{})
		s := SyncService{Registry: r, Store: store}
		if s.Step(context.Background()) != nil {
			t.Fatal("fetch failed")
		}
		purpose := "history"
		if mode == "incremental" {
			purpose = "reconciliation"
		}
		if store.page.Items[0].Event.Purpose != purpose || store.page.Items[0].Event.Acquisition != "poll" || store.page.NextCursor != "next" || store.page.RetryAfter != time.Minute {
			t.Fatal("page contract")
		}
	}
	store := &syncMemory{job: SyncJob{Platform: "fixture", Cursor: "cursor", Limit: 10}}
	r, _ := plugin.NewRegistry(fixtureFetcher{failure: ErrCursorExpired})
	s := SyncService{Registry: r, Store: store}
	if !errors.Is(s.Step(context.Background()), ErrCursorExpired) || store.reason != "cursor_expired" {
		t.Fatal("expired cursor must quarantine")
	}
}
