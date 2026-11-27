package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SaveOAuthSession stages sealed Page tokens and purges expired sessions in the same transaction.
func (s *Store) SaveOAuthSession(ctx context.Context, id, app string, sealed service.Sealed, expires, now time.Time) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM connector.oauth_sessions WHERE expires_at<=$1`, now); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO connector.oauth_sessions(id,app_id,payload_ciphertext,key_version,expires_at,created_at) VALUES($1,$2,$3,$4,$5,$6)`, id, app, sealed.Ciphertext, sealed.KeyVersion, expires, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) OAuthSession(ctx context.Context, id, app string, now time.Time) (service.Sealed, error) {
	var sealed service.Sealed
	if _, err := uuid.Parse(id); err != nil {
		return sealed, service.ErrSessionNotFound
	}
	err := s.Pool.QueryRow(ctx, `SELECT payload_ciphertext,key_version FROM connector.oauth_sessions WHERE id=$1 AND app_id=$2 AND expires_at>$3`, id, app, now).Scan(&sealed.Ciphertext, &sealed.KeyVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return sealed, service.ErrSessionNotFound
	}
	return sealed, err
}

func enqueueOutcome(ctx context.Context, tx pgx.Tx, o service.AttemptOutcome, now time.Time) error {
	payload, err := json.Marshal(o.Result)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO connector.outbound_events(id,topic,event_key,payload,state,available_at) VALUES($1,$2,$3,$4,'pending',$5)`, uuid.NewString(), delivery.SendCompletedTopic, o.Result.IntentID, payload, now); err != nil {
		return err
	}
	if o.Health == nil {
		return nil
	}
	payload, err = json.Marshal(o.Health)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO connector.outbound_events(id,topic,event_key,payload,state,available_at) VALUES($1,$2,$3,$4,'pending',$5)`, uuid.NewString(), connection.HealthChangedTopic, o.Health.ConnectionID, payload, now)
	return err
}

// BeginAttempt claims the attempt before any provider call. A redelivered attempt still marked dispatching may
// already have reached the provider, so it becomes unknown and is never sent again.
func (s *Store) BeginAttempt(ctx context.Context, intent delivery.SendRequested, now time.Time, reject func(service.Connection, *service.Sealed) *service.AttemptOutcome) (*service.Dispatch, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO connector.outbound_attempts(attempt_id,intent_id,connection_id,status,created_at) VALUES($1,$2,$3,'dispatching',$4) ON CONFLICT(attempt_id) DO NOTHING`, intent.AttemptID, intent.IntentID, intent.ConnectionID, now)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		tag, err = tx.Exec(ctx, `UPDATE connector.outbound_attempts SET status='unknown',error_category=$2,error_code='interrupted' WHERE attempt_id=$1 AND status='dispatching'`, intent.AttemptID, delivery.ErrorTransport)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 1 {
			o := service.AttemptOutcome{Result: delivery.SendCompleted{Meta: events.NewMeta(intent.TenantID, intent.CorrelationID, now), IntentID: intent.IntentID, AttemptID: intent.AttemptID, ConnectionID: intent.ConnectionID, Status: delivery.StatusUnknown, ErrorCategory: delivery.ErrorTransport, ErrorCode: "interrupted"}}
			if err = enqueueOutcome(ctx, tx, o, now); err != nil {
				return nil, err
			}
		}
		return nil, tx.Commit(ctx)
	}
	var c service.Connection
	var sealed service.Sealed
	var key *int
	err = tx.QueryRow(ctx, `SELECT id,version,tenant_id,platform,account_id,app_id,enabled,credential_ciphertext,key_version FROM connector.connections WHERE id=$1 FOR SHARE`, intent.ConnectionID).Scan(&c.ID, &c.Version, &c.TenantID, &c.Platform, &c.AccountID, &c.AppID, &c.Enabled, &sealed.Ciphertext, &key)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	var credential *service.Sealed
	if key != nil {
		sealed.KeyVersion = *key
		credential = &sealed
	}
	if o := reject(c, credential); o != nil {
		if _, err = tx.Exec(ctx, `UPDATE connector.outbound_attempts SET status=$2,error_category=$3,error_code=$4 WHERE attempt_id=$1`, intent.AttemptID, o.Result.Status, o.Result.ErrorCategory, o.Result.ErrorCode); err != nil {
			return nil, err
		}
		if err = enqueueOutcome(ctx, tx, *o, now); err != nil {
			return nil, err
		}
		return nil, tx.Commit(ctx)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &service.Dispatch{Connection: c, Credential: *credential}, nil
}

// FinishAttempt records the provider outcome and its result/health events atomically.
func (s *Store) FinishAttempt(ctx context.Context, o service.AttemptOutcome) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE connector.outbound_attempts SET status=$2,error_category=NULLIF($3,''),error_code=NULLIF($4,''),provider_message_id=NULLIF($5,'') WHERE attempt_id=$1 AND status='dispatching'`, o.Result.AttemptID, o.Result.Status, o.Result.ErrorCategory, o.Result.ErrorCode, o.Result.ProviderMessageID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return tx.Commit(ctx)
	}
	if err = enqueueOutcome(ctx, tx, o, time.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type LeasedOutbound struct {
	ID, Topic, Key string
	Fence          int64
	Payload        []byte
}

// ClaimOutbound leases one result/health event, mirroring the ingress event lease and fencing.
func (s *Store) ClaimOutbound(ctx context.Context) (LeasedOutbound, error) {
	if _, err := s.Pool.Exec(ctx, "UPDATE connector.outbound_events SET state='quarantine' WHERE state IN ('pending','leased') AND attempts>=12 AND available_at<=now()"); err != nil {
		return LeasedOutbound{}, err
	}
	var v LeasedOutbound
	err := s.Pool.QueryRow(ctx, `UPDATE connector.outbound_events SET state='leased', fence=fence+1, attempts=attempts+1, available_at=now()+interval '30 seconds' WHERE id=(SELECT id FROM connector.outbound_events WHERE state IN ('pending','leased') AND available_at<=now() AND attempts<12 ORDER BY available_at,id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,topic,event_key,fence,payload`).Scan(&v.ID, &v.Topic, &v.Key, &v.Fence, &v.Payload)
	return v, err
}

func (s *Store) FinishOutbound(ctx context.Context, v LeasedOutbound, success bool) error {
	state := "pending"
	if success {
		state = "published"
	}
	tag, err := s.Pool.Exec(ctx, `UPDATE connector.outbound_events SET state=CASE WHEN attempts>=12 AND $3='pending' THEN 'quarantine' ELSE $3 END, available_at=now()+make_interval(secs=>LEAST(300, power(2,attempts)::int)) WHERE id=$1 AND fence=$2 AND state='leased' AND available_at>now()`, v.ID, v.Fence, state)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("lease lost")
	}
	return err
}
