package service

import (
	"context"
	"net/mail"
	"strings"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SalesService struct{ repo repository.ISalesRepository }

func NewSalesService(repo repository.ISalesRepository) *SalesService {
	return &SalesService{repo: repo}
}

// Every read and command authorizes the current member in the same transaction.
func (s *SalesService) authorized(ctx context.Context, tenant, actor string, fn func(repository.ISalesRepository, bool) *apperrors.Error) *apperrors.Error {
	if tenant == "" {
		return apperrors.ErrBadRequest(ctx).SetDetail("X-Tenant-ID is required")
	}
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx).SetDetail("authentication required")
	}
	return s.repo.Transaction(ctx, func(r repository.ISalesRepository) *apperrors.Error {
		member, err := r.Member(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if member == nil {
			return apperrors.ErrForbidden(ctx).SetDetail("workspace membership required")
		}
		workspace, err := r.Tenant(ctx, tenant)
		if err != nil {
			return err
		}
		if workspace.Status != model.TenantStatusActive {
			return apperrors.ErrForbidden(ctx).SetDetail("workspace is not active")
		}
		return fn(r, workspace.OwnerID == actor)
	})
}
func pageBounds(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if page > 100000 {
		page = 100000
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return size, (page - 1) * size
}
func validMoney(value *decimal.Decimal) bool {
	return value == nil || (!value.IsNegative() && value.LessThan(decimal.New(1, 13)) && value.Equal(value.Round(2)))
}
func cleanID(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	v := strings.TrimSpace(*value)
	return &v
}
func (s *SalesService) Customers(ctx context.Context, tenant, actor string, page, size int) ([]model.Customer, int64, *apperrors.Error) {
	var rows []model.Customer
	var total int64
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.Customers(ctx, tenant, limit, offset)
		return err
	})
	return rows, total, err
}
func (s *SalesService) SaveCustomer(ctx context.Context, tenant, actor, id string, input types.CustomerInput) (*model.Customer, *apperrors.Error) {
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.Address = strings.TrimSpace(input.Address)
	if input.FullName == "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("full_name is required")
	}
	if input.Email != "" {
		address, err := mail.ParseAddress(input.Email)
		if err != nil || address.Address != input.Email {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("email must be a valid email address")
		}
	}
	var row *model.Customer
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		create := id == ""
		if create {
			row = &model.Customer{ID: uuid.NewString(), TenantID: tenant}
		} else {
			var err *apperrors.Error
			row, err = r.Customer(ctx, tenant, id)
			if err != nil {
				return err
			}
		}
		row.FullName = input.FullName
		row.Phone = input.Phone
		row.Email = input.Email
		row.Address = input.Address
		if err := r.SaveCustomer(ctx, row, create); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "customer.saved", "customer", row.ID, row)
	})
	return row, err
}
func (s *SalesService) ArchiveCustomer(ctx context.Context, tenant, actor, id string) *apperrors.Error {
	return s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		row, err := r.Customer(ctx, tenant, id)
		if err != nil {
			return err
		}
		if err := r.ArchiveCustomer(ctx, tenant, id); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "customer.archived", "customer", id, row)
	})
}
func (s *SalesService) Products(ctx context.Context, tenant, actor string, page, size int) ([]model.Product, int64, *apperrors.Error) {
	var rows []model.Product
	var total int64
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.Products(ctx, tenant, limit, offset)
		return err
	})
	return rows, total, err
}
func (s *SalesService) SaveProduct(ctx context.Context, tenant, actor, id string, input types.ProductInput) (*model.Product, *apperrors.Error) {
	input.Name = strings.TrimSpace(input.Name)
	input.SKU = strings.TrimSpace(input.SKU)
	input.VariantName = strings.TrimSpace(input.VariantName)
	if input.Name == "" || input.SKU == "" || input.VariantName == "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("name, sku and variant_name are required")
	}
	if !validMoney(input.Price) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("price must be nonnegative, below 10000000000000 and have at most two decimal places")
	}
	var row *model.Product
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		create := id == ""
		if create {
			row = &model.Product{ID: uuid.NewString(), TenantID: tenant, Status: model.ProductStatusActive}
			row.Variants = []model.ProductVariant{{ID: uuid.NewString(), TenantID: tenant, ProductID: row.ID}}
		} else {
			var err *apperrors.Error
			row, err = r.Product(ctx, tenant, id)
			if err != nil {
				return err
			}
			if row.Status != model.ProductStatusActive {
				return apperrors.ErrConflict(ctx, "product", "is archived")
			}
			if len(row.Variants) == 0 {
				return apperrors.ErrConflict(ctx, "product", "has no editable variant")
			}
		}
		row.Name = input.Name
		row.Description = strings.TrimSpace(input.Description)
		row.Category = strings.TrimSpace(input.Category)
		variant := &row.Variants[0]
		variant.SKU = input.SKU
		variant.VariantName = input.VariantName
		variant.PriceProvided = input.Price != nil
		variant.Price = decimal.Zero
		if input.Price != nil {
			variant.Price = *input.Price
		}
		if err := r.SaveProduct(ctx, row, create); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "product.saved", "product", row.ID, row)
	})
	return row, err
}
func (s *SalesService) ArchiveProduct(ctx context.Context, tenant, actor, id string) *apperrors.Error {
	return s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		row, err := r.Product(ctx, tenant, id)
		if err != nil {
			return err
		}
		row.Status = model.ProductStatusArchived
		if err := r.SaveProduct(ctx, row, false); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "product.archived", "product", id, row)
	})
}
func (s *SalesService) Orders(ctx context.Context, tenant, actor string, page, size int) ([]model.Order, int64, *apperrors.Error) {
	var rows []model.Order
	var total int64
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.Orders(ctx, tenant, limit, offset)
		return err
	})
	return rows, total, err
}
func (s *SalesService) Order(ctx context.Context, tenant, actor, id string) (*model.Order, *apperrors.Error) {
	var row *model.Order
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		var err *apperrors.Error
		row, err = r.Order(ctx, tenant, id)
		return err
	})
	return row, err
}
func (s *SalesService) SaveDraft(ctx context.Context, tenant, actor, id string, input types.DraftInput) (*model.Order, *apperrors.Error) {
	input.CustomerID = cleanID(input.CustomerID)
	input.EmployeeID = cleanID(input.EmployeeID)
	var row *model.Order
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, owner bool) *apperrors.Error {
		create := id == ""
		expected := input.Version
		if create {
			row = &model.Order{ID: uuid.NewString(), TenantID: tenant, OrderNumber: "ORD-" + strings.ToUpper(uuid.NewString()), Status: model.OrderStatusDraft, PaymentStatus: model.PaymentStatusUnpaid, Source: model.OrderSourceWebForm, Version: 1}
		} else {
			var err *apperrors.Error
			row, err = r.Order(ctx, tenant, id)
			if err != nil {
				return err
			}
			if row.Status != model.OrderStatusDraft || row.Version != expected {
				return apperrors.ErrConflict(ctx, "order", "changed or is no longer a draft; reload before editing")
			}
			if !owner && row.EmployeeID != nil && *row.EmployeeID != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale or owner can edit this draft")
			}
			// Legacy discounts are retained in totals; the demo does not edit discounts.
			row.Version++
		}
		if !owner {
			prior := ""
			if row.EmployeeID != nil {
				prior = *row.EmployeeID
			}
			next := ""
			if input.EmployeeID != nil {
				next = *input.EmployeeID
			}
			if next != prior && next != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the owner can assign another sale or clear an assignment")
			}
		}
		if input.CustomerID != nil {
			if _, err := r.Customer(ctx, tenant, *input.CustomerID); err != nil {
				return err
			}
		}
		if input.EmployeeID != nil {
			member, err := r.Member(ctx, tenant, *input.EmployeeID)
			if err != nil {
				return err
			}
			if member == nil {
				return apperrors.ErrBadRequest(ctx).SetDetail("employee_id must identify a current workspace member")
			}
		}
		row.CustomerID = input.CustomerID
		row.EmployeeID = input.EmployeeID
		row.RequiresDelivery = input.RequiresDelivery
		row.DeliveryAddress = strings.TrimSpace(input.DeliveryAddress)
		row.Notes = strings.TrimSpace(input.Notes)
		if len(input.Items) > 100 {
			return apperrors.ErrBadRequest(ctx).SetDetail("items: at most 100 lines are supported")
		}
		variantIDs := make([]string, len(input.Items))
		for i, item := range input.Items {
			variantIDs[i] = item.VariantID
		}
		if _, err := r.Variants(ctx, tenant, variantIDs); err != nil {
			return err
		}
		row.Items = make([]model.OrderItem, 0, len(input.Items))
		row.Subtotal = decimal.Zero
		for _, item := range input.Items {
			if item.VariantID == "" || item.Quantity <= 0 || item.Quantity > 1000000 || !validMoney(item.UnitPrice) {
				return apperrors.ErrBadRequest(ctx).SetDetail("items require a product, positive quantity (at most 1000000) and a valid optional unit_price")
			}
			price := decimal.Zero
			if item.UnitPrice != nil {
				price = *item.UnitPrice
			}
			line := price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			row.Subtotal = row.Subtotal.Add(line)
			row.Items = append(row.Items, model.OrderItem{ID: uuid.NewString(), TenantID: tenant, OrderID: row.ID, VariantID: item.VariantID, Quantity: item.Quantity, UnitPrice: price, PriceProvided: item.UnitPrice != nil, LineTotal: line})
		}
		if row.Subtotal.GreaterThanOrEqual(decimal.New(1, 13)) {
			return apperrors.ErrBadRequest(ctx).SetDetail("order total exceeds the supported amount")
		}
		row.TotalAmount = decimal.Max(decimal.Zero, row.Subtotal.Sub(row.Discount))
		if err := r.SaveOrder(ctx, row, expected, create); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "order.draft_saved", "order", row.ID, row)
	})
	return row, err
}
func (s *SalesService) Confirm(ctx context.Context, tenant, actor, id string, version int64) (*model.Order, *apperrors.Error) {
	var row *model.Order
	err := s.authorized(ctx, tenant, actor, func(r repository.ISalesRepository, _ bool) *apperrors.Error {
		var err *apperrors.Error
		row, err = r.Order(ctx, tenant, id)
		if err != nil {
			return err
		}
		if row.Status != model.OrderStatusDraft || row.Version != version {
			return apperrors.ErrConflict(ctx, "order", "changed or already confirmed; reload before confirming")
		}
		if row.EmployeeID == nil || *row.EmployeeID != actor {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale can confirm this order")
		}
		if row.CustomerID == nil {
			return apperrors.ErrBadRequest(ctx).SetDetail("customer_id: select a customer before confirmation")
		}
		customer, err := r.Customer(ctx, tenant, *row.CustomerID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(customer.FullName) == "" {
			return apperrors.ErrBadRequest(ctx).SetDetail("customer.full_name: required before confirmation")
		}
		if strings.TrimSpace(customer.Phone) == "" && strings.TrimSpace(customer.Email) == "" {
			return apperrors.ErrBadRequest(ctx).SetDetail("customer.contact: phone or email is required before confirmation")
		}
		if len(row.Items) == 0 {
			return apperrors.ErrBadRequest(ctx).SetDetail("items: add at least one product before confirmation")
		}
		for _, item := range row.Items {
			if !item.PriceProvided || item.Quantity <= 0 || !validMoney(&item.UnitPrice) {
				return apperrors.ErrBadRequest(ctx).SetDetail("items.unit_price: review each line and explicitly enter a nonnegative price")
			}
		}
		variantIDs := make([]string, len(row.Items))
		for i, item := range row.Items {
			variantIDs[i] = item.VariantID
		}
		if _, err := r.Variants(ctx, tenant, variantIDs); err != nil {
			return err
		}
		if row.RequiresDelivery && strings.TrimSpace(row.DeliveryAddress) == "" {
			return apperrors.ErrBadRequest(ctx).SetDetail("delivery_address: required when delivery is selected")
		}
		row.Status = model.OrderStatusConfirmed
		row.Version++
		if err := r.SaveOrder(ctx, row, version, false); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "order.confirmed", "order", row.ID, row)
	})
	return row, err
}
