package controller

import (
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TenantController struct {
	*baseController
	tenantService service.ITenantService
}

func NewTenantController(validate *validator.Validate, tenantService *service.TenantService) *TenantController {
	return &TenantController{
		baseController: NewBaseController(validate),
		tenantService:  tenantService,
	}
}

func (ctrl *TenantController) getAuthUserID(c *gin.Context) (string, *errors.Error) {
	if info, ok := common.GetAuthInfo(c.Request.Context()); ok && info.UserID != "" {
		return info.UserID, nil
	}
	if uid := c.GetString("userID"); uid != "" {
		return uid, nil
	}
	return "", errors.ErrUnauthorized(c).SetDetail("authentication required")
}

// GetUserTenants
// @Summary List user workspaces
// @Description Get all workspaces the authenticated user belongs to
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[[]dto.TenantResponse]
// @Router /api/v1/tenants [get]
func (ctrl *TenantController) GetUserTenants(c *gin.Context) {
	userID, err := ctrl.getAuthUserID(c)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenants, err := ctrl.tenantService.GetUserTenants(c.Request.Context(), userID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := make([]dto.TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		res = append(res, dto.TenantResponse{
			ID:        t.ID,
			Name:      t.Name,
			Slug:      t.Slug,
			Plan:      t.Plan,
			Status:    t.Status,
			OwnerID:   t.OwnerID,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		})
	}

	ctrl.Success(c, res)
}

// GetTenant
// @Summary Get workspace details
// @Description Get details for a specific workspace (caller must be a member)
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.TenantResponse]
// @Router /api/v1/tenants/{tenantID} [get]
func (ctrl *TenantController) GetTenant(c *gin.Context) {
	userID, err := ctrl.getAuthUserID(c)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenantID := c.Param("tenantID")
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c).SetDetail("tenant id is required"))
		return
	}

	tenant, err := ctrl.tenantService.GetTenant(c.Request.Context(), userID, tenantID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Plan:      tenant.Plan,
		Status:    tenant.Status,
		OwnerID:   tenant.OwnerID,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	ctrl.Success(c, res)
}

// UpdateTenant
// @Summary Update workspace settings
// @Description Update workspace details (requires owner or admin role)
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.UpdateTenantRequest true "Workspace update data"
// @Success 200 {object} response.Response[dto.TenantResponse]
// @Router /api/v1/tenants/{tenantID} [put]
func (ctrl *TenantController) UpdateTenant(c *gin.Context) {
	userID, err := ctrl.getAuthUserID(c)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenantID := c.Param("tenantID")
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c).SetDetail("tenant id is required"))
		return
	}

	var req dto.UpdateTenantRequest
	if err := ctrl.BindAndValidateRequest(c, &req); err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenant, err := ctrl.tenantService.UpdateTenant(c.Request.Context(), userID, tenantID, req.Name, req.Plan, req.Status)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := dto.TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Plan:      tenant.Plan,
		Status:    tenant.Status,
		OwnerID:   tenant.OwnerID,
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	ctrl.Success(c, res)
}

// ListMembers
// @Summary List workspace members
// @Description Get all members belonging to a workspace
// @Tags Tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[[]dto.TenantMemberResponse]
// @Router /api/v1/tenants/{tenantID}/users [get]
func (ctrl *TenantController) ListMembers(c *gin.Context) {
	userID, err := ctrl.getAuthUserID(c)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	tenantID := c.Param("tenantID")
	if tenantID == "" {
		ctrl.ErrorData(c, errors.ErrBadRequest(c).SetDetail("tenant id is required"))
		return
	}

	members, err := ctrl.tenantService.ListMembers(c.Request.Context(), userID, tenantID)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}

	res := make([]dto.TenantMemberResponse, 0, len(members))
	for _, m := range members {
		res = append(res, dto.TenantMemberResponse{
			ID:        m.ID,
			TenantID:  m.TenantID,
			UserID:    m.UserID,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		})
	}

	ctrl.Success(c, res)
}
