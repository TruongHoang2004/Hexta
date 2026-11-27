package repository

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

// The database bounds return lines even when a caller bypasses the service checks.
func TestOrderCaseLinesAreBoundedByTheOriginalOrderLine(t *testing.T) {
	db := testutil.Database(t)
	for _, statement := range []string{
		`INSERT INTO orders(id,tenant_id,order_number,status) VALUES ('order-1','alpha','A-1','confirmed'),('order-2','alpha','A-2','confirmed')`,
		`INSERT INTO order_items(id,tenant_id,order_id,variant_id,quantity) VALUES ('item-1','alpha','order-1','variant',2),('item-2','alpha','order-2','variant',5)`,
		`INSERT INTO order_cases(id,tenant_id,order_id,case_type,status,handler_id,opened_by,version) VALUES ('case-1','alpha','order-1','return','open','sale','sale',1)`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	insertLine := func(id, order, item string, original, quantity int) error {
		return db.Exec(`INSERT INTO order_case_lines(id,tenant_id,case_id,order_id,order_item_id,original_quantity,quantity) VALUES (?,?,?,?,?,?,?)`, id, "alpha", "case-1", order, item, original, quantity).Error
	}
	require.Error(t, insertLine("over", "order-1", "item-1", 2, 3), "returned quantity above the original is refused")
	require.Error(t, insertLine("forged", "order-1", "item-1", 9, 3), "a forged original quantity does not match the order line")
	require.Error(t, insertLine("foreign", "order-2", "item-2", 5, 1), "a line of another order cannot join this case")
	require.Error(t, insertLine("zero", "order-1", "item-1", 2, 0))
	require.NoError(t, insertLine("ok", "order-1", "item-1", 2, 2))

	// One case per order, and confirmed order lines referenced by a case cannot be changed or removed.
	require.Error(t, db.Exec(`INSERT INTO order_cases(id,tenant_id,order_id,case_type,status,handler_id,opened_by,version) VALUES ('case-2','alpha','order-1','cancellation','open','sale','sale',1)`).Error)
	require.Error(t, db.Exec(`UPDATE order_items SET quantity = 1 WHERE id = 'item-1'`).Error)
	require.Error(t, db.Exec(`DELETE FROM orders WHERE id = 'order-1'`).Error)
	// A finalized case must name its finalizer and time.
	require.Error(t, db.Exec(`UPDATE order_cases SET status = 'finalized' WHERE id = 'case-1'`).Error)
}

func TestOrderCaseRepositoryMapsConstraintFailures(t *testing.T) {
	db := testutil.Database(t)
	require.NoError(t, db.Exec(`INSERT INTO orders(id,tenant_id,order_number,status) VALUES ('order-1','alpha','A-1','confirmed')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_items(id,tenant_id,order_id,variant_id,quantity) VALUES ('item-1','alpha','order-1','variant',2)`).Error)
	repo := NewOrderCaseRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	value := types.OrderCase{ID: "case-1", TenantID: "alpha", OrderID: "order-1", Type: types.OrderCaseTypeReturn, Status: types.OrderCaseStatusOpen, HandlerID: "sale", OpenedBy: "sale", Version: 1, CreatedAt: now, UpdatedAt: now}
	require.Nil(t, repo.InsertCase(ctx, value))

	duplicate := value
	duplicate.ID = "case-2"
	err := repo.InsertCase(ctx, duplicate)
	require.NotNil(t, err)
	require.Equal(t, 409, err.GetHttpStatus())

	value.Version = 2
	value.Lines = []types.OrderCaseLine{{OrderItemID: "item-1", OriginalQuantity: 2, Quantity: 3}}
	err = repo.Within(ctx, func(store OrderCaseStore) *apperrors.Error { return store.SaveCase(ctx, value, 1, true) })
	require.NotNil(t, err)
	require.Equal(t, 400, err.GetHttpStatus())

	// The rollback kept version 1, so a stale expected version now conflicts.
	value.Lines = []types.OrderCaseLine{{OrderItemID: "item-1", OriginalQuantity: 2, Quantity: 1}}
	require.Nil(t, repo.SaveCase(ctx, value, 1, true))
	err = repo.SaveCase(ctx, value, 1, true)
	require.NotNil(t, err)
	require.Equal(t, 409, err.GetHttpStatus())

	event := types.OrderCaseEvent{ID: "event-1", ActorID: "sale", Action: types.OrderCaseActionUpdated, Status: types.OrderCaseStatusOpen, Version: 2, Lines: value.Lines, CreatedAt: now}
	require.Nil(t, repo.AppendEvent(ctx, "alpha", "case-1", event))
	event.ID = "event-2"
	err = repo.AppendEvent(ctx, "alpha", "case-1", event)
	require.NotNil(t, err)
	require.Equal(t, 409, err.GetHttpStatus())

	stored, err := repo.CaseByOrder(ctx, "alpha", "order-1", false)
	require.Nil(t, err)
	require.Equal(t, int64(2), stored.Version)
	require.Equal(t, []types.OrderCaseLine{{OrderItemID: "item-1", OriginalQuantity: 2, Quantity: 1}}, stored.Lines)
	require.Len(t, stored.Events, 1)
	require.Equal(t, value.Lines, stored.Events[0].Lines)
	// Another tenant sees nothing.
	other, err := repo.CaseByOrder(ctx, "beta", "order-1", false)
	require.Nil(t, err)
	require.Nil(t, other)
}
