package repository

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOutboxLeaseFencingBackoffAndExhaustion(t *testing.T) {
	db := testutil.Database(t)
	repo := NewOutboxRepository(db)
	ctx := context.Background()

	idle, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.Nil(t, idle)

	event := model.DomainEvent{ID: uuid.NewString(), TenantID: "tenant", EventType: model.DomainEventOrderConfirmed, AggregateID: "order", Payload: `{"order_id":"order"}`}
	require.NoError(t, db.Create(&event).Error)

	first, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.Equal(t, event.ID, first.ID)
	require.Equal(t, model.DomainEventStateLeased, first.State)
	require.EqualValues(t, 1, first.Fence)
	require.Equal(t, 1, first.Attempts)
	// An active lease is not handed to another relay.
	other, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.Nil(t, other)

	// A failed publication returns the event to pending with backoff, so it is not immediately due.
	require.Nil(t, repo.FinishEvent(ctx, first, false))
	again, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.Nil(t, again)

	// An expired lease is recovered with a new fence; the stale lease can no longer finish the event.
	require.NoError(t, db.Exec(`UPDATE domain_events SET available_at=clock_timestamp()-interval '1 second' WHERE id=?`, event.ID).Error)
	second, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.EqualValues(t, 2, second.Fence)
	require.NotNil(t, repo.FinishEvent(ctx, first, true))
	require.Nil(t, repo.FinishEvent(ctx, second, true))
	var published model.DomainEvent
	require.NoError(t, db.First(&published, "id = ?", event.ID).Error)
	require.Equal(t, model.DomainEventStatePublished, published.State)
	require.NotNil(t, published.PublishedAt)

	// Events that exhaust their attempts are parked as failed instead of retried forever.
	exhausted := model.DomainEvent{ID: uuid.NewString(), TenantID: "tenant", EventType: model.DomainEventOrderConfirmed, AggregateID: "order-2", Payload: `{}`, Attempts: outboxMaxAttempts}
	require.NoError(t, db.Create(&exhausted).Error)
	none, failure := repo.ClaimEvent(ctx)
	require.Nil(t, failure)
	require.Nil(t, none)
	require.NoError(t, db.First(&exhausted, "id = ?", exhausted.ID).Error)
	require.Equal(t, model.DomainEventStateFailed, exhausted.State)
}
