package controller

import (
	"net/http"
	"strconv"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SalesController struct {
	*baseController
	sales *service.SalesService
}

func NewSalesController(validate *validator.Validate, sales *service.SalesService) *SalesController {
	return &SalesController{NewBaseController(validate), sales}
}
func salesScope(c *gin.Context) (string, string) {
	info, ok := common.GetAuthInfo(c.Request.Context())
	if !ok {
		return c.Param("tenantID"), ""
	}
	return c.Param("tenantID"), info.UserID
}
func salesPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
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
	return page, size
}
func (ctrl *SalesController) bind(c *gin.Context, req any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	if err := c.ShouldBindJSON(req); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid JSON request"))
		return false
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), req); err != nil {
		ctrl.ErrorData(c, err)
		return false
	}
	return true
}

// ListCustomers
// @Summary List workspace customers
// @Description List workspace customers
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page (default 1)"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.SalesCustomersResponse]
// @Router /api/v1/tenants/{tenantID}/customers [get]
func (ctrl *SalesController) ListCustomers(c *gin.Context) {
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.sales.Customers(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.SalesCustomersResponse{Items: rows, Total: total, Page: page, Size: size})
}

// ListProducts
// @Summary List workspace products
// @Description List workspace products
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page (default 1)"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.SalesProductsResponse]
// @Router /api/v1/tenants/{tenantID}/products [get]
func (ctrl *SalesController) ListProducts(c *gin.Context) {
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.sales.Products(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.SalesProductsResponse{Items: rows, Total: total, Page: page, Size: size})
}

// ListOrders
// @Summary List workspace orders
// @Description List workspace orders
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page (default 1)"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.SalesOrdersResponse]
// @Router /api/v1/tenants/{tenantID}/orders [get]
func (ctrl *SalesController) ListOrders(c *gin.Context) {
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.sales.Orders(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.SalesOrdersResponse{Items: rows, Total: total, Page: page, Size: size})
}

// CreateCustomer
// @Summary Create customer
// @Description Create customer
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.SalesCustomerRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesCustomerResponse]
// @Router /api/v1/tenants/{tenantID}/customers [post]
func (ctrl *SalesController) CreateCustomer(c *gin.Context) {
	var req dto.SalesCustomerRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.SaveCustomer(c.Request.Context(), tenant, actor, "", types.CustomerInput{FullName: req.FullName, Phone: req.Phone, Email: req.Email, Address: req.Address})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// UpdateCustomer
// @Summary Update customer
// @Description Update customer
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Param request body dto.SalesCustomerRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesCustomerResponse]
// @Router /api/v1/tenants/{tenantID}/customers/{id} [put]
func (ctrl *SalesController) UpdateCustomer(c *gin.Context) {
	var req dto.SalesCustomerRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.SaveCustomer(c.Request.Context(), tenant, actor, c.Param("id"), types.CustomerInput{FullName: req.FullName, Phone: req.Phone, Email: req.Email, Address: req.Address})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// ArchiveCustomer
// @Summary Archive customer
// @Description Archive without deleting historical order references
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Success 200 {object} response.Response[dto.SalesArchiveResponse]
// @Router /api/v1/tenants/{tenantID}/customers/{id} [delete]
func (ctrl *SalesController) ArchiveCustomer(c *gin.Context) {
	tenant, actor := salesScope(c)
	if err := ctrl.sales.ArchiveCustomer(c.Request.Context(), tenant, actor, c.Param("id")); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.SalesArchiveResponse{ID: c.Param("id"), Archived: true})
}

// CreateProduct
// @Summary Create product
// @Description Create product
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.SalesProductRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesProductResponse]
// @Router /api/v1/tenants/{tenantID}/products [post]
func (ctrl *SalesController) CreateProduct(c *gin.Context) {
	var req dto.SalesProductRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.SaveProduct(c.Request.Context(), tenant, actor, "", types.ProductInput{Name: req.Name, Description: req.Description, Category: req.Category, SKU: req.SKU, VariantName: req.VariantName, Price: req.Price})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// UpdateProduct
// @Summary Update product
// @Description Update product
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Param request body dto.SalesProductRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesProductResponse]
// @Router /api/v1/tenants/{tenantID}/products/{id} [put]
func (ctrl *SalesController) UpdateProduct(c *gin.Context) {
	var req dto.SalesProductRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.SaveProduct(c.Request.Context(), tenant, actor, c.Param("id"), types.ProductInput{Name: req.Name, Description: req.Description, Category: req.Category, SKU: req.SKU, VariantName: req.VariantName, Price: req.Price})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// ArchiveProduct
// @Summary Archive product
// @Description Archive without deleting historical order references
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Success 200 {object} response.Response[dto.SalesArchiveResponse]
// @Router /api/v1/tenants/{tenantID}/products/{id} [delete]
func (ctrl *SalesController) ArchiveProduct(c *gin.Context) {
	tenant, actor := salesScope(c)
	if err := ctrl.sales.ArchiveProduct(c.Request.Context(), tenant, actor, c.Param("id")); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.SalesArchiveResponse{ID: c.Param("id"), Archived: true})
}

// GetOrder
// @Summary Get workspace order
// @Description Get workspace order
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Success 200 {object} response.Response[dto.SalesOrderResponse]
// @Router /api/v1/tenants/{tenantID}/orders/{id} [get]
func (ctrl *SalesController) GetOrder(c *gin.Context) {
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.Order(c.Request.Context(), tenant, actor, c.Param("id"))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// CreateOrder
// @Summary Save an incomplete order draft
// @Description Save draft fields and optional line prices; update requires the current version. Confirmation is a separate assigned-sale action.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.SalesDraftRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesOrderResponse]
// @Router /api/v1/tenants/{tenantID}/orders [post]
func (ctrl *SalesController) CreateOrder(c *gin.Context) {
	var req dto.SalesDraftRequest
	if !ctrl.bind(c, &req) {
		return
	}
	if req.AutoConfirm {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("auto_confirm is unsupported; save then explicitly confirm"))
		return
	}
	tenant, actor := salesScope(c)
	items := make([]types.DraftItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = types.DraftItemInput{VariantID: item.VariantID, Quantity: item.Quantity, UnitPrice: item.UnitPrice}
	}
	row, err := ctrl.sales.SaveDraft(c.Request.Context(), tenant, actor, "", types.DraftInput{Version: req.Version, CustomerID: req.CustomerID, EmployeeID: req.EmployeeID, RequiresDelivery: req.RequiresDelivery, DeliveryAddress: req.DeliveryAddress, Notes: req.Notes, Items: items})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// UpdateOrder
// @Summary Save an incomplete order draft
// @Description Save draft fields and optional line prices; update requires the current version. Confirmation is a separate assigned-sale action.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Param request body dto.SalesDraftRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesOrderResponse]
// @Router /api/v1/tenants/{tenantID}/orders/{id} [put]
func (ctrl *SalesController) UpdateOrder(c *gin.Context) {
	var req dto.SalesDraftRequest
	if !ctrl.bind(c, &req) {
		return
	}
	if req.AutoConfirm {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("auto_confirm is unsupported; save then explicitly confirm"))
		return
	}
	tenant, actor := salesScope(c)
	items := make([]types.DraftItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = types.DraftItemInput{VariantID: item.VariantID, Quantity: item.Quantity, UnitPrice: item.UnitPrice}
	}
	row, err := ctrl.sales.SaveDraft(c.Request.Context(), tenant, actor, c.Param("id"), types.DraftInput{Version: req.Version, CustomerID: req.CustomerID, EmployeeID: req.EmployeeID, RequiresDelivery: req.RequiresDelivery, DeliveryAddress: req.DeliveryAddress, Notes: req.Notes, Items: items})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// ConfirmOrder
// @Summary Confirm an order as the assigned sale
// @Description Validate current version, customer/contact, products, explicit prices, sale and optional delivery address; audit atomically without stock or payment side effects.
// @Tags Sales
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Record ID"
// @Param request body dto.SalesConfirmRequest true "Record data"
// @Success 200 {object} response.Response[dto.SalesOrderResponse]
// @Router /api/v1/tenants/{tenantID}/orders/{id}/confirm [post]
func (ctrl *SalesController) ConfirmOrder(c *gin.Context) {
	var req dto.SalesConfirmRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.sales.Confirm(c.Request.Context(), tenant, actor, c.Param("id"), req.Version)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}
