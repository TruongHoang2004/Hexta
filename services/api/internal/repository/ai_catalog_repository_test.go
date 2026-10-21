package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestAICatalogTenantFiltersAndLimits(t *testing.T) {
	db := testutil.Database(t)
	repo := NewAICatalogRepository(db)
	ctx := context.Background()
	for _, tenant := range []string{"alpha", "beta"} {
		for i := 0; i < 12; i++ {
			product := model.Product{ID: fmt.Sprintf("%s-product-%d", tenant, i), TenantID: tenant, Name: "Catalog Item", Category: "category", Status: model.ProductStatusActive}
			require.NoError(t, db.Create(&product).Error)
			customer := model.Customer{ID: fmt.Sprintf("%s-customer-%d", tenant, i), TenantID: tenant, FullName: "Customer", Phone: fmt.Sprintf("%03d", i)}
			require.NoError(t, db.Create(&customer).Error)
		}
	}
	require.NoError(t, db.Create(&model.Product{ID: "archived", TenantID: "alpha", Name: "Catalog Item", Status: model.ProductStatusArchived}).Error)
	require.NoError(t, db.Create(&model.Product{ID: "variants", TenantID: "alpha", Name: "Unique", Status: model.ProductStatusActive}).Error)
	// The legacy schema permits a mismatched variant tenant. Preload must still scope it.
	for _, tenant := range []string{"alpha", "beta"} {
		variant := model.ProductVariant{ID: tenant + "-variant", TenantID: tenant, ProductID: "variants", SKU: "SKU", VariantName: "Unit", Price: decimal.NewFromInt(25)}
		require.NoError(t, db.Create(&variant).Error)
	}
	require.NoError(t, db.Create(&model.InventoryItem{ID: "inventory", TenantID: "alpha", VariantID: "alpha-variant", AvailableQty: 8, OnHandQty: 10, ReservedQty: 2}).Error)

	products, failure := repo.SearchProducts(ctx, "alpha", "Catalog")
	require.Nil(t, failure)
	require.Len(t, products, 10)
	for _, product := range products {
		require.Equal(t, "alpha", product.TenantID)
		require.Equal(t, model.ProductStatusActive, product.Status)
	}
	products, failure = repo.SearchProducts(ctx, "alpha", "Unique")
	require.Nil(t, failure)
	require.Len(t, products, 1)
	require.Len(t, products[0].Variants, 1)
	require.Equal(t, "alpha", products[0].Variants[0].TenantID)
	products, failure = repo.SearchProducts(ctx, "alpha", "category")
	require.Nil(t, failure)
	require.Len(t, products, 10)
	customers, failure := repo.LookupCustomers(ctx, "alpha", "", "Customer")
	require.Nil(t, failure)
	require.Len(t, customers, 5)
	for _, customer := range customers {
		require.Equal(t, "alpha", customer.TenantID)
	}
	customers, failure = repo.LookupCustomers(ctx, "alpha", "011", "Customer")
	require.Nil(t, failure)
	require.Len(t, customers, 1)
	customers, failure = repo.LookupCustomers(ctx, "alpha", "011", "Missing")
	require.Nil(t, failure)
	require.Empty(t, customers)

	variant, inventory, failure := repo.StockBySKU(ctx, "alpha", "SKU")
	require.Nil(t, failure)
	require.Equal(t, "alpha-variant", variant.ID)
	require.Equal(t, 8, inventory.AvailableQty)
	variant, inventory, failure = repo.StockBySKU(ctx, "beta", "SKU")
	require.Nil(t, failure)
	require.Equal(t, "beta-variant", variant.ID)
	require.Nil(t, inventory)
	variant, inventory, failure = repo.StockBySKU(ctx, "unknown", "SKU")
	require.Nil(t, failure)
	require.Nil(t, variant)
	require.Nil(t, inventory)

	// Operational errors must not masquerade as an absent SKU or expose driver detail.
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, failure = repo.SearchProducts(cancelled, "alpha", "")
	require.NotNil(t, failure)
	require.Equal(t, 500, failure.GetHttpStatus())
	require.Equal(t, "AI catalog persistence operation failed", failure.GetDetail())
	_, _, failure = repo.StockBySKU(cancelled, "alpha", "SKU")
	require.NotNil(t, failure)
	require.Equal(t, 500, failure.GetHttpStatus())
	_, failure = repo.LookupCustomers(cancelled, "alpha", "", "")
	require.NotNil(t, failure)
	require.Equal(t, 500, failure.GetHttpStatus())
}
