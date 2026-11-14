package repository

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"slices"
	"strings"
	"time"
)

type Store struct{ Pool *pgxpool.Pool }

// maxPendingEvents bounds the unpublished backlog. Admission is approximate: concurrent batches can each pass the
// check before the others commit, so the backlog may exceed it by the other in-flight batches.
var maxPendingEvents = 100000

type acceptedEvent struct {
	id, key, state, reason string
	payload                []byte
}

// Accept locks connection configuration through commit. No network I/O occurs in this transaction.
// It takes only row-level locks, so callbacks with different receipts are accepted concurrently.
func (s *Store) Accept(ctx context.Context, app, platform, receipt string, items []plugin.Item, now time.Time) error {
	room := maxPendingEvents - len(items)
	if room < 0 {
		return errors.New("storage capacity")
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Claiming the receipt first deduplicates retries: a concurrent identical batch waits on this row, then sees
	// the conflict after commit or claims the receipt itself after rollback.
	tag, err := tx.Exec(ctx, "INSERT INTO connector.receipts(id,app_id,received_at) VALUES($1,$2,$3) ON CONFLICT(id,app_id) DO NOTHING", receipt, app, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}
	// Counting stops once the backlog is known to leave no room for this batch.
	var backlog int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM (SELECT 1 FROM connector.events WHERE state IN ('pending','leased') LIMIT $1) b", room+1).Scan(&backlog); err != nil {
		return err
	}
	if backlog > room {
		return errors.New("storage capacity")
	}
	events := make([]acceptedEvent, 0, len(items))
	for i, item := range items {
		var c service.Connection
		err = tx.QueryRow(ctx, "SELECT id,version,tenant_id,platform,account_id,app_id,enabled FROM connector.connections WHERE app_id=$1 AND platform=$2 AND account_id=$3 FOR SHARE", app, platform, item.AccountID).Scan(&c.ID, &c.Version, &c.TenantID, &c.Platform, &c.AccountID, &c.AppID, &c.Enabled)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		reason := item.Reason
		if errors.Is(err, pgx.ErrNoRows) {
			reason = "unknown_connection"
		} else if !c.Enabled {
			reason = "disabled_connection"
		}
		e := service.Envelope(item.Event, c)
		v := acceptedEvent{id: e.EventID, key: ingress.Hash(app, receipt, string(rune(i))), state: "pending", reason: reason}
		if reason != "" {
			v.state = "quarantine"
		} else {
			if err = e.Validate(); err != nil {
				return err
			}
			v.key = e.Identity()
			v.payload, err = json.Marshal(e)
			if err != nil {
				return err
			}
		}
		events = append(events, v)
	}
	// Overlapping batches (webhook/poll or redelivery) insert shared event keys in one order, so they cannot deadlock.
	slices.SortStableFunc(events, func(a, b acceptedEvent) int { return strings.Compare(a.key, b.key) })
	for _, v := range events {
		if _, err = tx.Exec(ctx, "INSERT INTO connector.events(id,receipt_id,app_id,event_key,payload,state,reason,available_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(event_key) DO NOTHING", v.id, receipt, app, v.key, v.payload, v.state, v.reason, now); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type Leased struct {
	ID      string
	Fence   int64
	Payload []byte
}

func (s *Store) Claim(ctx context.Context) (Leased, error) {
	_, err := s.Pool.Exec(ctx, "UPDATE connector.events SET state='quarantine',reason='publication_exhausted' WHERE state IN ('pending','leased') AND attempts>=12 AND available_at<=now()")
	if err != nil {
		return Leased{}, err
	}
	var v Leased
	err = s.Pool.QueryRow(ctx, `UPDATE connector.events SET state='leased', fence=fence+1, attempts=attempts+1, available_at=now()+interval '30 seconds' WHERE id=(SELECT id FROM connector.events WHERE ((state='pending' AND available_at<=now()) OR (state='leased' AND available_at<=now())) AND attempts<12 ORDER BY available_at,id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,fence,payload`).Scan(&v.ID, &v.Fence, &v.Payload)
	return v, err
}
func (s *Store) Finish(ctx context.Context, v Leased, success bool) error {
	state := "pending"
	if success {
		state = "published"
	}
	tag, err := s.Pool.Exec(ctx, `UPDATE connector.events SET state=CASE WHEN attempts>=12 AND $3='pending' THEN 'quarantine' ELSE $3 END, reason=CASE WHEN attempts>=12 AND $3='pending' THEN 'publication_exhausted' ELSE reason END, available_at=now()+make_interval(secs=>LEAST(300, power(2,attempts)::int)) WHERE id=$1 AND fence=$2 AND state='leased' AND available_at>now()`, v.ID, v.Fence, state)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("lease lost")
	}
	return err
}

// PutConnection is trusted machine provisioning, with immutable tenant/account/app identity and monotonic versions.
func (s *Store) PutConnection(ctx context.Context, c service.Connection) error {
	tag, err := s.Pool.Exec(ctx, `INSERT INTO connector.connections(id,version,tenant_id,platform,account_id,app_id,enabled,credential_ref) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO UPDATE SET version=excluded.version,enabled=excluded.enabled,credential_ref=excluded.credential_ref WHERE connections.tenant_id=excluded.tenant_id AND connections.platform=excluded.platform AND connections.account_id=excluded.account_id AND connections.app_id=excluded.app_id AND connections.version+1=excluded.version`, c.ID, c.Version, c.TenantID, c.Platform, c.AccountID, c.AppID, c.Enabled, c.CredentialRef)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("configuration conflict")
	}
	return err
}

var _ service.InboundStore = (*Store)(nil)
var _ service.SyncStore = (*Store)(nil)
