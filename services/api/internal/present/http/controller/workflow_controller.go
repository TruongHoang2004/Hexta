package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type WorkflowController struct {
	*baseController
	workflows *service.WorkflowService
}

func NewWorkflowController(validate *validator.Validate, workflows *service.WorkflowService) *WorkflowController {
	return &WorkflowController{NewBaseController(validate), workflows}
}
func (ctrl *WorkflowController) validScope(c *gin.Context) bool {
	for _, value := range []string{c.Param("tenantID"), c.Param("id")} {
		if value != "" {
			if _, err := uuid.Parse(value); err != nil {
				ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid resource ID"))
				return false
			}
		}
	}
	return true
}
func (ctrl *WorkflowController) bind(c *gin.Context, input any) bool {
	if !ctrl.validScope(c) {
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
	if err := c.ShouldBindJSON(input); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid JSON request"))
		return false
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), input); err != nil {
		ctrl.ErrorData(c, err)
		return false
	}
	return true
}

// ListDefinitions
// @Summary List current tenant workflow definitions
// @Description Returns the latest immutable version and ordered steps for each definition.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.WorkflowsResponse]
// @Router /api/v1/tenants/{tenantID}/workflow-definitions [get]
func (ctrl *WorkflowController) ListDefinitions(c *gin.Context) {
	if !ctrl.validScope(c) {
		return
	}
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.workflows.Definitions(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.WorkflowsResponse{Items: rows, Total: total, Page: page, Size: size})
}

// CreateDefinition
// @Summary Create and publish a workflow as tenant owner
// @Description The UUID creation key is durable; changed payload reuse conflicts.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param Idempotency-Key header string true "UUID creation key"
// @Param request body dto.WorkflowRequest true "Named ordered steps"
// @Success 200 {object} response.Response[dto.WorkflowResponse]
// @Router /api/v1/tenants/{tenantID}/workflow-definitions [post]
func (ctrl *WorkflowController) CreateDefinition(c *gin.Context) { ctrl.publish(c) }

// PublishVersion
// @Summary Publish a new immutable workflow version as tenant owner
// @Description Existing work keeps its pinned version; stale expected_version conflicts.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Definition ID"
// @Param request body dto.WorkflowRequest true "Expected version and replacement steps"
// @Success 200 {object} response.Response[dto.WorkflowResponse]
// @Router /api/v1/tenants/{tenantID}/workflow-definitions/{id}/versions [post]
func (ctrl *WorkflowController) PublishVersion(c *gin.Context) { ctrl.publish(c) }
func (ctrl *WorkflowController) publish(c *gin.Context) {
	var req dto.WorkflowRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.workflows.Publish(c.Request.Context(), tenant, actor, c.Param("id"), c.GetHeader("Idempotency-Key"), types.WorkflowInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// ListWork
// @Summary List tenant work and pinned workflow steps
// @Description Includes current responsibility, step and pending handoff recipient.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.WorkItemsResponse]
// @Router /api/v1/tenants/{tenantID}/work-items [get]
func (ctrl *WorkflowController) ListWork(c *gin.Context) {
	if !ctrl.validScope(c) {
		return
	}
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.workflows.Items(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.WorkItemsResponse{Items: rows, Total: total, Page: page, Size: size})
}

// GetWork
// @Summary Get tenant work detail and recent history
// @Description Returns pinned workflow steps and the latest 100 transition events.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Work item ID"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items/{id} [get]
func (ctrl *WorkflowController) GetWork(c *gin.Context) {
	if !ctrl.validScope(c) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.workflows.Item(c.Request.Context(), tenant, actor, c.Param("id"))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// CreateWork
// @Summary Explicitly create work from a confirmed order as tenant owner
// @Description Pins a published workflow version and nominates an initial current member.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param Idempotency-Key header string true "UUID creation key"
// @Param request body dto.WorkCreateRequest true "Confirmed order, workflow version and responsible member"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items [post]
func (ctrl *WorkflowController) CreateWork(c *gin.Context) {
	var req dto.WorkCreateRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.workflows.Create(c.Request.Context(), tenant, actor, c.GetHeader("Idempotency-Key"), types.WorkInput(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}

// Handoff
// @Summary Hand work to the next step as its current responsible member
// @Description Source step policy selects direct transfer or nominated acceptance while retaining current responsibility.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Work item ID"
// @Param request body dto.WorkCommandRequest true "Expected version and recipient_id"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items/{id}/handoff [post]
func (ctrl *WorkflowController) Handoff(c *gin.Context) { ctrl.command(c, "handoff") }

// Accept
// @Summary Accept a pending handoff as the nominated recipient
// @Description Atomically advances one step and transfers responsibility after checking current membership.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Work item ID"
// @Param request body dto.WorkCommandRequest true "Expected version"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items/{id}/accept [post]
func (ctrl *WorkflowController) Accept(c *gin.Context) { ctrl.command(c, "accept") }

// Assign
// @Summary Reassign or unassign work as tenant owner
// @Description Retains the current step and pinned workflow; cancels any pending nomination.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Work item ID"
// @Param request body dto.WorkCommandRequest true "Expected version and nullable responsible_id"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items/{id}/assign [post]
func (ctrl *WorkflowController) Assign(c *gin.Context) { ctrl.command(c, "assign") }

// Complete
// @Summary Complete the final step as its current responsible member
// @Description Completes operational work without changing the order, payment, inventory or messages.
// @Tags Workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Work item ID"
// @Param request body dto.WorkCommandRequest true "Expected version"
// @Success 200 {object} response.Response[dto.WorkResponse]
// @Router /api/v1/tenants/{tenantID}/work-items/{id}/complete [post]
func (ctrl *WorkflowController) Complete(c *gin.Context) { ctrl.command(c, "complete") }
func (ctrl *WorkflowController) command(c *gin.Context, action string) {
	var req dto.WorkCommandRequest
	if !ctrl.bind(c, &req) {
		return
	}
	tenant, actor := salesScope(c)
	row, err := ctrl.workflows.Command(c.Request.Context(), tenant, actor, c.Param("id"), action, types.WorkCommand(req))
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, row)
}
