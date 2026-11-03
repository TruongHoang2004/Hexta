package service

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"time"
)

var ErrCredentialRevoked = errors.New("credential revoked")

var ErrCursorExpired = errors.New("cursor expired: restart requires explicit reconciliation job")

type SyncJob struct {
	ID, AppID, Platform, AccountID, Cursor, Mode string
	Limit                                        int
	Fence                                        int64
}
type SyncStore interface {
	ClaimSync(context.Context) (SyncJob, error)
	CommitPage(context.Context, SyncJob, plugin.Page) error
	FailSync(context.Context, SyncJob, string, time.Duration) error
}
type SyncService struct {
	Registry *plugin.Registry
	Store    SyncStore
}

func (s *SyncService) Step(ctx context.Context) error {
	j, err := s.Store.ClaimSync(ctx)
	if err != nil {
		return err
	}
	p, err := s.Registry.Fetcher(j.Platform)
	if err != nil {
		_ = s.Store.FailSync(ctx, j, "unsupported", 0)
		return err
	}
	call, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	page, err := p.Fetch(call, plugin.FetchRequest{AccountID: j.AccountID, Cursor: j.Cursor, Mode: j.Mode, Limit: j.Limit})
	if err != nil {
		reason := "fetch_failed"
		if errors.Is(err, ErrCursorExpired) {
			reason = "cursor_expired"
		}
		if errors.Is(err, ErrCredentialRevoked) {
			reason = "credential_revoked"
		}
		retry := page.RetryAfter
		if retry <= 0 {
			retry = time.Minute
		}
		if retry > time.Hour {
			retry = time.Hour
		}
		_ = s.Store.FailSync(ctx, j, reason, retry)
		return err
	}
	if len(page.Items) > j.Limit || page.RetryAfter < 0 {
		return s.Store.FailSync(ctx, j, "invalid_page", 0)
	}
	for i := range page.Items {
		page.Items[i].Event.Acquisition = "poll"
		page.Items[i].Event.Purpose = j.Mode
		if j.Mode == "incremental" {
			page.Items[i].Event.Purpose = "reconciliation"
		}
	}
	return s.Store.CommitPage(ctx, j, page)
}
