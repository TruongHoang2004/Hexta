package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"time"
)

func (s *Store) ClaimSync(ctx context.Context) (service.SyncJob, error) {
	_, cleanupErr := s.Pool.Exec(ctx, "UPDATE connector.sync_jobs SET state='quarantine',reason='retry_exhausted' WHERE state IN ('pending','leased') AND attempts>=8 AND available_at<=now()")
	if cleanupErr != nil {
		return service.SyncJob{}, cleanupErr
	}
	var j service.SyncJob
	err := s.Pool.QueryRow(ctx, `UPDATE connector.sync_jobs SET state='leased',fence=fence+1,attempts=attempts+1,available_at=now()+interval '30 seconds' WHERE id=(SELECT id FROM connector.sync_jobs WHERE state IN ('pending','leased') AND available_at<=now() AND attempts<8 ORDER BY available_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,app_id,platform,account_id,cursor,mode,page_limit,fence`).Scan(&j.ID, &j.AppID, &j.Platform, &j.AccountID, &j.Cursor, &j.Mode, &j.Limit, &j.Fence)
	return j, err
}
func (s *Store) CommitPage(ctx context.Context, j service.SyncJob, p plugin.Page) error {
	// Data first, then fenced checkpoint. A crash between commits refetches the same page;
	// provider message identity deduplicates it, including overlap with webhooks.
	b, _ := json.Marshal(p.Items)
	if err := s.Accept(ctx, j.AppID, j.Platform, ingress.Hash(j.ID, j.Cursor, string(b)), p.Items, time.Now().UTC()); err != nil {
		return err
	}
	state := "pending"
	if p.NextCursor == "" && j.Mode == "history" {
		state = "done"
	}
	cursor := p.NextCursor
	if cursor == "" && j.Mode == "incremental" {
		cursor = j.Cursor
	}
	delay := p.RetryAfter
	if delay < time.Second {
		delay = time.Second
	}
	tag, err := s.Pool.Exec(ctx, `UPDATE connector.sync_jobs SET cursor=$3,state=$4,attempts=0,available_at=now()+make_interval(secs=>$5) WHERE id=$1 AND fence=$2 AND state='leased' AND available_at>now()`, j.ID, j.Fence, cursor, state, delay.Seconds())
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("sync lease lost")
	}
	return err
}
func (s *Store) FailSync(ctx context.Context, j service.SyncJob, reason string, delay time.Duration) error {
	state := "pending"
	if reason == "unsupported" || reason == "cursor_expired" || reason == "invalid_page" || reason == "credential_revoked" {
		state = "quarantine"
	}
	_, err := s.Pool.Exec(ctx, `UPDATE connector.sync_jobs SET state=CASE WHEN attempts>=8 THEN 'quarantine' ELSE $3 END,reason=$4,available_at=now()+make_interval(secs=>$5) WHERE id=$1 AND fence=$2 AND state='leased' AND available_at>now()`, j.ID, j.Fence, state, reason, delay.Seconds())
	return err
}
