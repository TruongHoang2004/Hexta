package repository

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestOrderItemsMustShareTheirOrderTenant(t *testing.T) {
	db := testutil.Database(t)
	require.NoError(t, db.Exec(`INSERT INTO orders(id,tenant_id,order_number) VALUES ('order-1','alpha','A-1')`).Error)

	// An item that names another tenant's order is rejected instead of becoming a cross-tenant orphan.
	err := db.Exec(`INSERT INTO order_items(id,tenant_id,order_id,variant_id,quantity) VALUES ('item-beta','beta','order-1','variant',1)`).Error
	require.Error(t, err)
	require.NoError(t, db.Exec(`INSERT INTO order_items(id,tenant_id,order_id,variant_id,quantity) VALUES ('item-alpha','alpha','order-1','variant',1)`).Error)

	// Deleting the order still cascades to its items through the composite key.
	require.NoError(t, db.Exec(`DELETE FROM orders WHERE id='order-1'`).Error)
	var remaining int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM order_items WHERE order_id='order-1'`).Scan(&remaining).Error)
	require.Zero(t, remaining)
}
