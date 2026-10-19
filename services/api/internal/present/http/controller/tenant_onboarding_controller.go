package controller

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TenantOnboardingController struct {
	*SalesController
	onboarding *service.TenantOnboardingService
}

func NewTenantOnboardingController(validate *validator.Validate, sales *service.SalesService, onboarding *service.TenantOnboardingService) *TenantOnboardingController {
	return &TenantOnboardingController{NewSalesController(validate, sales), onboarding}
}

// PlatformAccess
// @Summary PlatformAccess
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[dto.PlatformAccessResponse]
// @Router /api/v1/platform/access [get]
func (c *TenantOnboardingController) PlatformAccess(ctx *gin.Context) {
	_, actor := salesScope(ctx)
	allowed, err := c.onboarding.Operator(ctx.Request.Context(), actor)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, dto.PlatformAccessResponse{Operator: allowed})
}

// MyApplications
// @Summary MyApplications
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[dto.TenantApplicationsResponse]
// @Router /api/v1/tenant-requests [get]
func (c *TenantOnboardingController) MyApplications(ctx *gin.Context) {
	_, actor := salesScope(ctx)
	rows, err := c.onboarding.Requests(ctx.Request.Context(), actor, false)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, dto.TenantApplicationsResponse{Items: rows})
}

// PlatformApplications
// @Summary PlatformApplications
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[dto.TenantApplicationsResponse]
// @Router /api/v1/platform/tenant-requests [get]
func (c *TenantOnboardingController) PlatformApplications(ctx *gin.Context) {
	_, actor := salesScope(ctx)
	rows, err := c.onboarding.Requests(ctx.Request.Context(), actor, true)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, dto.TenantApplicationsResponse{Items: rows})
}

// ApplyForTenant
// @Summary ApplyForTenant
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TenantApplicationRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantApplicationResponse]
// @Router /api/v1/tenant-requests [post]
func (c *TenantOnboardingController) ApplyForTenant(ctx *gin.Context) {
	var req dto.TenantApplicationRequest
	if !c.bind(ctx, &req) {
		return
	}
	_, actor := salesScope(ctx)
	row, err := c.onboarding.Apply(ctx.Request.Context(), actor, req.Name, req.Slug)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, row)
}

// ReviewApplication
// @Summary ReviewApplication
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Request or invitation ID"
// @Param request body dto.TenantReviewRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantApplicationResponse]
// @Router /api/v1/platform/tenant-requests/{id}/review [post]
func (c *TenantOnboardingController) ReviewApplication(ctx *gin.Context) {
	var req dto.TenantReviewRequest
	if !c.bind(ctx, &req) {
		return
	}
	_, actor := salesScope(ctx)
	row, err := c.onboarding.Review(ctx.Request.Context(), actor, ctx.Param("id"), req.Decision, req.Reason)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, row)
}

// CreateEmailInvitation
// @Summary CreateEmailInvitation
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.TenantEmailInvitationRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantInvitationResponse]
// @Router /api/v1/tenants/{tenantID}/invitations [post]
func (c *TenantOnboardingController) CreateEmailInvitation(ctx *gin.Context) {
	var req dto.TenantEmailInvitationRequest
	if !c.bind(ctx, &req) {
		return
	}
	tenant, actor := salesScope(ctx)
	row, token, err := c.onboarding.Invite(ctx.Request.Context(), actor, tenant, req.Email)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	c.Success(ctx, dto.TenantInvitationResponse{Invitation: row, Token: token, Delivery: "manual_link_not_emailed"})
}

// ListInvitations
// @Summary ListInvitations
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.TenantInvitationsResponse]
// @Router /api/v1/tenants/{tenantID}/invitations [get]
func (c *TenantOnboardingController) ListInvitations(ctx *gin.Context) {
	tenant, actor := salesScope(ctx)
	rows, err := c.onboarding.Invitations(ctx.Request.Context(), actor, tenant)
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, dto.TenantInvitationsResponse{Items: rows})
}

// RevokeInvitation
// @Summary RevokeInvitation
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Request or invitation ID"
// @Success 200 {object} response.Response[dto.TenantRevokedResponse]
// @Router /api/v1/tenants/{tenantID}/invitations/{id}/revoke [post]
func (c *TenantOnboardingController) RevokeInvitation(ctx *gin.Context) {
	tenant, actor := salesScope(ctx)
	if err := c.onboarding.Revoke(ctx.Request.Context(), actor, tenant, ctx.Param("id")); err != nil {
		c.ErrorData(ctx, err)
		return
	}
	c.Success(ctx, dto.TenantRevokedResponse{Revoked: true})
}

// PreviewInvitation
// @Summary PreviewInvitation
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TenantInvitationTokenRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantInvitationResponse]
// @Router /api/v1/invitations/preview [post]
func (c *TenantOnboardingController) PreviewInvitation(ctx *gin.Context) {
	var req dto.TenantInvitationTokenRequest
	if !c.bind(ctx, &req) {
		return
	}
	_, actor := salesScope(ctx)
	row, name, err := c.onboarding.UseInvitation(ctx.Request.Context(), actor, req.Token, "preview")
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	c.Success(ctx, dto.TenantInvitationResponse{Invitation: row, TenantName: name})
}

// AcceptInvitation
// @Summary AcceptInvitation
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TenantInvitationTokenRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantInvitationResponse]
// @Router /api/v1/invitations/accept [post]
func (c *TenantOnboardingController) AcceptInvitation(ctx *gin.Context) {
	var req dto.TenantInvitationTokenRequest
	if !c.bind(ctx, &req) {
		return
	}
	_, actor := salesScope(ctx)
	row, name, err := c.onboarding.UseInvitation(ctx.Request.Context(), actor, req.Token, "accept")
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	c.Success(ctx, dto.TenantInvitationResponse{Invitation: row, TenantName: name})
}

// DeclineInvitation
// @Summary DeclineInvitation
// @Description Tenant approval or invitation operation with server-side actor checks
// @Tags Tenant onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TenantInvitationTokenRequest true "Request data"
// @Success 200 {object} response.Response[dto.TenantInvitationResponse]
// @Router /api/v1/invitations/decline [post]
func (c *TenantOnboardingController) DeclineInvitation(ctx *gin.Context) {
	var req dto.TenantInvitationTokenRequest
	if !c.bind(ctx, &req) {
		return
	}
	_, actor := salesScope(ctx)
	row, name, err := c.onboarding.UseInvitation(ctx.Request.Context(), actor, req.Token, "decline")
	if err != nil {
		c.ErrorData(ctx, err)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	c.Success(ctx, dto.TenantInvitationResponse{Invitation: row, TenantName: name})
}
