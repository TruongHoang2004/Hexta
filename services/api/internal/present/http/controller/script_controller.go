package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/script/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ScriptController struct {
	*baseController
	scripts *service.ScriptService
}

func NewScriptController(validate *validator.Validate, scripts *service.ScriptService) *ScriptController {
	return &ScriptController{NewBaseController(validate), scripts}
}

// bind validates path IDs and, when input is non-nil, a bounded JSON body.
func (ctrl *ScriptController) bind(c *gin.Context, input any) bool {
	for _, id := range []string{c.Param("tenantID"), c.Param("id"), c.Param("userID")} {
		if id != "" {
			if _, err := uuid.Parse(id); err != nil {
				ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid resource ID"))
				return false
			}
		}
	}
	if input == nil {
		return true
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

func (ctrl *ScriptController) reply(c *gin.Context, result *types.ScriptDetail, err *apperrors.Error) {
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Script(*result))
}

// ListScripts
// @Summary List bot reply scripts
// @Description Active tenant membership required; members without script permissions can still read. Newest change first.
// @Tags Scripts
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param status query string false "inactive, active or archived (default: all)"
// @Param page query int false "Page"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.ScriptsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts [get]
func (ctrl *ScriptController) ListScripts(c *gin.Context) {
	if !ctrl.bind(c, nil) {
		return
	}
	var query dto.ScriptListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid query"))
		return
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), query); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	rows, total, err := ctrl.scripts.List(c.Request.Context(), tenant, actor, types.ScriptListFilter{Status: query.Status}, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Scripts(rows, total, page, size))
}

// GetScript
// @Summary Get a script with its version history
// @Description Active tenant membership required. Versions are newest first (at most 100); published versions never change.
// @Tags Scripts
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Script ID"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,404 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts/{id} [get]
func (ctrl *ScriptController) GetScript(c *gin.Context) {
	if !ctrl.bind(c, nil) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.Get(c.Request.Context(), tenant, actor, c.Param("id"))
	ctrl.reply(c, result, err)
}

// CreateScript
// @Summary Create an inactive script with draft v1
// @Description Requires script.create (owner implicitly). The UUID Idempotency-Key becomes the script ID; replaying the same payload returns the original script and a changed payload conflicts.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param Idempotency-Key header string true "UUID creation key"
// @Param request body dto.CreateScriptRequest true "Name and draft content"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts [post]
func (ctrl *ScriptController) CreateScript(c *gin.Context) {
	var input dto.CreateScriptRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.Create(c.Request.Context(), tenant, actor, c.GetHeader("Idempotency-Key"), types.ScriptCreateInput{Name: input.Name, Content: types.ScriptContent{Triggers: input.Triggers, Answer: input.Answer}})
	ctrl.reply(c, result, err)
}

// SaveScriptDraft
// @Summary Save the script draft
// @Description Requires script.edit. Rewrites the open draft, or opens draft vN+1 when every version is published (editing an active script never changes its active content). Stale expected_version or an archived script conflicts.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Script ID"
// @Param request body dto.SaveScriptDraftRequest true "Name, draft content and expected script version"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts/{id}/draft [put]
func (ctrl *ScriptController) SaveScriptDraft(c *gin.Context) {
	var input dto.SaveScriptDraftRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.SaveDraft(c.Request.Context(), tenant, actor, c.Param("id"), types.ScriptDraftInput{Name: input.Name, Content: types.ScriptContent{Triggers: input.Triggers, Answer: input.Answer}, ExpectedVersion: input.ExpectedVersion})
	ctrl.reply(c, result, err)
}

// ActivateScript
// @Summary Activate a script version
// @Description Requires script.activate. Points the script at the selected version (the draft or an earlier published version); a draft is frozen on activation.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Script ID"
// @Param request body dto.ActivateScriptRequest true "Version and expected script version"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts/{id}/activate [post]
func (ctrl *ScriptController) ActivateScript(c *gin.Context) {
	var input dto.ActivateScriptRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.Activate(c.Request.Context(), tenant, actor, c.Param("id"), input.VersionID, input.ExpectedVersion)
	ctrl.reply(c, result, err)
}

// DeactivateScript
// @Summary Deactivate a script
// @Description Requires script.activate. The script becomes inactive; its versions stay in history.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Script ID"
// @Param request body dto.ScriptTransitionRequest true "Expected script version"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts/{id}/deactivate [post]
func (ctrl *ScriptController) DeactivateScript(c *gin.Context) {
	var input dto.ScriptTransitionRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.Deactivate(c.Request.Context(), tenant, actor, c.Param("id"), input.ExpectedVersion)
	ctrl.reply(c, result, err)
}

// ArchiveScript
// @Summary Archive a script
// @Description Requires script.archive. Archived is terminal; the script and all versions are kept, never deleted.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Script ID"
// @Param request body dto.ScriptTransitionRequest true "Expected script version"
// @Success 200 {object} response.Response[dto.ScriptDTO]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/scripts/{id}/archive [post]
func (ctrl *ScriptController) ArchiveScript(c *gin.Context) {
	var input dto.ScriptTransitionRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.Archive(c.Request.Context(), tenant, actor, c.Param("id"), input.ExpectedVersion)
	ctrl.reply(c, result, err)
}

// MyScriptPermissions
// @Summary Get the caller's script permissions
// @Description Active tenant membership required. For showing or hiding actions only; every command is authorized again.
// @Tags Scripts
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.MyScriptPermissionsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/script-permissions/me [get]
func (ctrl *ScriptController) MyScriptPermissions(c *gin.Context) {
	if !ctrl.bind(c, nil) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.MyPermissions(c.Request.Context(), tenant, actor)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.MyPermissions(result))
}

// ListScriptPermissions
// @Summary List members' script permissions
// @Description Tenant owner only. Current members with their grants; the owner row always lists all four permissions.
// @Tags Scripts
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.ScriptPermissionsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/script-permissions [get]
func (ctrl *ScriptController) ListScriptPermissions(c *gin.Context) {
	if !ctrl.bind(c, nil) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.scripts.MemberPermissions(c.Request.Context(), tenant, actor)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Members(result))
}

// GrantScriptPermission
// @Summary Grant one script permission to a member
// @Description Tenant owner only; the target must be a current, non-owner member. expected_version is the member's permission version (0 before any grant); a stale value conflicts. Granting a held permission changes nothing.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param userID path string true "Member user ID"
// @Param request body dto.ScriptPermissionChangeRequest true "Permission and expected version"
// @Success 200 {object} response.Response[dto.MemberScriptPermissionsDTO]
// @Failure 400,401,403,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/script-permissions/{userID}/grant [post]
func (ctrl *ScriptController) GrantScriptPermission(c *gin.Context) { ctrl.permission(c, true) }

// RevokeScriptPermission
// @Summary Revoke one script permission from a member
// @Description Tenant owner only. Takes effect on the member's next action. expected_version is the member's permission version; a stale value conflicts.
// @Tags Scripts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param userID path string true "Member user ID"
// @Param request body dto.ScriptPermissionChangeRequest true "Permission and expected version"
// @Success 200 {object} response.Response[dto.MemberScriptPermissionsDTO]
// @Failure 400,401,403,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/script-permissions/{userID}/revoke [post]
func (ctrl *ScriptController) RevokeScriptPermission(c *gin.Context) { ctrl.permission(c, false) }

func (ctrl *ScriptController) permission(c *gin.Context, grant bool) {
	var input dto.ScriptPermissionChangeRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	change := ctrl.scripts.Revoke
	if grant {
		change = ctrl.scripts.Grant
	}
	result, err := change(c.Request.Context(), tenant, actor, c.Param("userID"), input.Permission, input.ExpectedVersion)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Member(*result))
}
