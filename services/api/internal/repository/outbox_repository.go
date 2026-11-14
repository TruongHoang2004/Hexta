package repository

import (
	"context"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
)

const (
	outboxMaxAttempts  = 12
	outboxLeaseSeconds = 30
)

type IOutboxRepository interface {
	ClaimEvent(context.Context) (*model.DomainEvent, *apperrors.Error)
	FinishEvent(context.Context, *model.DomainEvent, bool) *apperrors.Error
}

// OutboxRepository leases domain events for publication. A lease is a committed fence, so no
// transaction or row lock is held while the relay performs network I/O.
type OutboxRepository struct{ db *gorm.DB }

func NewOutboxRepository(db *gorm.DB) *OutboxRepository { return &OutboxRepository{db: db} }

func outboxError(ctx context.Context) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "outbox persistence operation failed")
}

// ClaimEvent leases the oldest due event, including one whose earlier lease expired. It returns nil when none is due.
func (r *OutboxRepository) ClaimEvent(ctx context.Context) (*model.DomainEvent, *apperrors.Error) {
	db := r.db.WithContext(ctx)
	if err := db.Exec(`UPDATE domain_events SET state='failed' WHERE state IN ('pending','leased') AND attempts>=? AND available_at<=clock_timestamp()`, outboxMaxAttempts).Error; err != nil {
		return nil, outboxError(ctx)
	}
	var rows []model.DomainEvent
	err := db.Raw(`UPDATE domain_events SET state='leased', fence=fence+1, attempts=attempts+1,
		available_at=clock_timestamp()+make_interval(secs=>?)
		WHERE id=(SELECT id FROM domain_events WHERE state IN ('pending','leased') AND available_at<=clock_timestamp() AND attempts<?
			ORDER BY available_at, id FOR UPDATE SKIP LOCKED LIMIT 1)
		RETURNING *`, outboxLeaseSeconds, outboxMaxAttempts).Scan(&rows).Error
	if err != nil {
		return nil, outboxError(ctx)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// FinishEvent records the publication outcome only while the caller still holds the lease. A failed attempt backs
// off exponentially (capped at 300 seconds) and is marked failed after the final attempt.
func (r *OutboxRepository) FinishEvent(ctx context.Context, event *model.DomainEvent, published bool) *apperrors.Error {
	var result *gorm.DB
	if published {
		result = r.db.WithContext(ctx).Exec(`UPDATE domain_events SET state='published', published_at=clock_timestamp()
			WHERE id=? AND fence=? AND state='leased' AND available_at>clock_timestamp()`, event.ID, event.Fence)
	} else {
		result = r.db.WithContext(ctx).Exec(`UPDATE domain_events SET state=CASE WHEN attempts>=? THEN 'failed' ELSE 'pending' END,
			available_at=clock_timestamp()+make_interval(secs=>LEAST(300, power(2, attempts)::int))
			WHERE id=? AND fence=? AND state='leased' AND available_at>clock_timestamp()`, outboxMaxAttempts, event.ID, event.Fence)
	}
	if result.Error != nil {
		return outboxError(ctx)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "outbox event", "lease lost")
	}
	return nil
}
