package service

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type salesFixture struct {
	db                               *gorm.DB
	service                          *SalesService
	tenant, other, owner, sale, peer string
}

func salesTestFixture(t *testing.T) salesFixture {
	t.Helper()
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	var db *gorm.DB
	var err error
	if dsn := os.Getenv("SALES_TEST_DATABASE_URL"); dsn != "" {
		admin, e := gorm.Open(postgres.Open(dsn), cfg)
		if e != nil {
			t.Fatal("test database unavailable")
		}
		schema := "sales_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		if e := admin.Exec("CREATE SCHEMA " + schema).Error; e != nil {
			t.Fatal(e)
		}
		parsed, e := url.Parse(dsn)
		if e != nil {
			t.Fatal(e)
		}
		q := parsed.Query()
		q.Set("search_path", schema)
		parsed.RawQuery = q.Encode()
		db, err = gorm.Open(postgres.Open(parsed.String()), cfg)
		t.Cleanup(func() {
			if db != nil {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}
			admin.Exec("DROP SCHEMA " + schema + " CASCADE")
			sqlDB, _ := admin.DB()
			sqlDB.Close()
		})
	} else {
		db, err = gorm.Open(sqlite.Open("file:sales_"+uuid.NewString()+"?mode=memory&cache=shared"), cfg)
		if err == nil {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { sqlDB.Close() })
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Tenant{}, &model.TenantMember{}, &model.Customer{}, &model.Product{}, &model.ProductVariant{}, &model.Order{}, &model.OrderItem{}, &model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	f := salesFixture{db: db, service: NewSalesService(repository.NewSalesRepository(db)), tenant: uuid.NewString(), other: uuid.NewString(), owner: "owner", sale: "sale", peer: "peer"}
	for _, id := range []string{f.tenant, f.other} {
		if err := db.Create(&model.Tenant{ID: id, Name: id, Slug: id, OwnerID: f.owner, Status: model.TenantStatusActive}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&model.TenantMember{TenantID: id, UserID: f.owner, Role: model.RoleOwner}).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, user := range []string{f.sale, f.peer} {
		if err := db.Create(&model.TenantMember{TenantID: f.tenant, UserID: user, Role: model.RoleMember}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return f
}
func requireSalesStatus(t *testing.T, err *apperrors.Error, status int) {
	t.Helper()
	if err == nil || err.GetHttpStatus() != status {
		t.Fatalf("expected status %d, got %v", status, err)
	}
}
func (f salesFixture) completeInput(t *testing.T) types.DraftInput {
	t.Helper()
	ctx := context.Background()
	customer, err := f.service.SaveCustomer(ctx, f.tenant, f.sale, "", types.CustomerInput{FullName: "Customer", Email: "customer@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	product, err := f.service.SaveProduct(ctx, f.tenant, f.sale, "", types.ProductInput{Name: "Product", SKU: uuid.NewString(), VariantName: "Unit"})
	if err != nil {
		t.Fatal(err)
	}
	zero := decimal.Zero
	return types.DraftInput{CustomerID: &customer.ID, EmployeeID: &f.sale, Items: []types.DraftItemInput{{VariantID: product.Variants[0].ID, Quantity: 1, UnitPrice: &zero}}}
}
func TestSalesIncompleteDraftAndConfirmation(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", types.DraftInput{})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Status != model.OrderStatusDraft || draft.Version != 1 || len(draft.Items) != 0 {
		t.Fatalf("unexpected draft: %+v", draft)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, 1)
	requireSalesStatus(t, err, 403)
	input := f.completeInput(t)
	input.Version = 1
	input.Items[0].UnitPrice = nil
	draft, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	requireSalesStatus(t, err, 400)
	zero := decimal.Zero
	input.Items[0].UnitPrice = &zero
	input.Version = draft.Version
	input.RequiresDelivery = true
	draft, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	requireSalesStatus(t, err, 400)
	input.Version = draft.Version
	input.DeliveryAddress = "Delivery address"
	draft, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	confirmed, err := f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Status != model.OrderStatusConfirmed || !confirmed.TotalAmount.IsZero() {
		t.Fatal("explicit zero must be confirmable")
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	requireSalesStatus(t, err, 409)
	input.Version = confirmed.Version
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	requireSalesStatus(t, err, 409)
	var count int64
	f.db.Model(&model.AuditLog{}).Where("action = ? AND entity_id = ?", "order.confirmed", draft.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected one confirmation audit, got %d", count)
	}
}
func TestSalesTenantAndAssignmentBoundaries(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	input := f.completeInput(t)
	_, _, err := f.service.Customers(ctx, f.other, f.sale, 1, 20)
	requireSalesStatus(t, err, 403)
	_, err = f.service.SaveDraft(ctx, f.other, f.owner, "", input)
	requireSalesStatus(t, err, 404)
	foreign := input
	foreign.CustomerID = nil
	foreign.EmployeeID = nil
	_, err = f.service.SaveDraft(ctx, f.other, f.owner, "", foreign)
	requireSalesStatus(t, err, 404)
	input.EmployeeID = &f.peer
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	requireSalesStatus(t, err, 403)
	input.EmployeeID = &f.sale
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.owner, draft.ID, draft.Version)
	requireSalesStatus(t, err, 403)
	input.Version = draft.Version
	_, err = f.service.SaveDraft(ctx, f.tenant, f.peer, draft.ID, input)
	requireSalesStatus(t, err, 403)
	input.EmployeeID = &f.peer
	draft, err = f.service.SaveDraft(ctx, f.tenant, f.owner, draft.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	requireSalesStatus(t, err, 403)
	if _, err = f.service.Confirm(ctx, f.tenant, f.peer, draft.ID, draft.Version); err != nil {
		t.Fatal(err)
	}
}
func TestSalesValidationRollbackAndStaleEdit(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	input := f.completeInput(t)
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	bad := input
	bad.Version = draft.Version
	bad.Notes = "must roll back"
	bad.Items = []types.DraftItemInput{{VariantID: "missing", Quantity: 1}}
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, bad)
	requireSalesStatus(t, err, 404)
	stored, err := f.service.Order(ctx, f.tenant, f.sale, draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Version != 1 || stored.Notes != "" || len(stored.Items) != 1 {
		t.Fatal("failed command modified the draft")
	}
	input.Version = 1
	input.Notes = "latest"
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input)
	requireSalesStatus(t, err, 409)
	negative := decimal.NewFromInt(-1)
	bad.Items = []types.DraftItemInput{{VariantID: input.Items[0].VariantID, Quantity: 1, UnitPrice: &negative}}
	bad.Version = 2
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, bad)
	requireSalesStatus(t, err, 400)
}
func TestSalesCatalogMissingContactsAndPrices(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	for _, name := range []string{"One", "Two"} {
		if _, err := f.service.SaveCustomer(ctx, f.tenant, f.sale, "", types.CustomerInput{FullName: name}); err != nil {
			t.Fatal(err)
		}
	}
	input := f.completeInput(t)
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	customer, err := f.service.SaveCustomer(ctx, f.tenant, f.sale, *input.CustomerID, types.CustomerInput{FullName: "No contact"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	requireSalesStatus(t, err, 400)
	if customer.Email != "" {
		t.Fatal("clearing optional fields must persist")
	}
	rows, _, err := f.service.Products(ctx, f.tenant, f.sale, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Variants[0].PriceProvided {
		t.Fatal("missing catalog price became zero")
	}
	if err := f.service.ArchiveProduct(ctx, f.tenant, f.sale, rows[0].ID); err != nil {
		t.Fatal(err)
	}
	_, err = f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	requireSalesStatus(t, err, 404)
}
func TestSalesConcurrentConfirmationPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" {
		t.Skip("set SALES_TEST_DATABASE_URL to a disposable PostgreSQL database for row-lock concurrency verification")
	}
	f := salesTestFixture(t)
	ctx := context.Background()
	input := f.completeInput(t)
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan *apperrors.Error, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if err.GetHttpStatus() == 409 {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("expected one winner and one conflict: %d/%d", success, conflict)
	}
}

func TestSalesConcurrentEditAndConfirmationPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" {
		t.Skip("requires disposable PostgreSQL row locks")
	}
	f := salesTestFixture(t)
	ctx := context.Background()
	input := f.completeInput(t)
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", input)
	if err != nil {
		t.Fatal(err)
	}
	input.Version = draft.Version
	input.Notes = "Concurrent edit"
	start := make(chan struct{})
	results := make(chan *apperrors.Error, 2)
	go func() { <-start; _, err := f.service.SaveDraft(ctx, f.tenant, f.sale, draft.ID, input); results <- err }()
	go func() {
		<-start
		_, err := f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
		results <- err
	}()
	close(start)
	winners, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			winners++
		} else if err.GetHttpStatus() == 409 {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("expected one edit/confirm winner, got %d winners and %d conflicts", winners, conflicts)
	}
	stored, err := f.service.Order(ctx, f.tenant, f.sale, draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Version != 2 {
		t.Fatalf("state changed more than once: version %d", stored.Version)
	}
}
