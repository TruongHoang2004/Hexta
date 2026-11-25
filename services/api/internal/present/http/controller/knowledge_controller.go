package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/knowledge/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type KnowledgeController struct {
	*baseController
	knowledge *service.KnowledgeService
}

func NewKnowledgeController(validate *validator.Validate, knowledge *service.KnowledgeService) *KnowledgeController {
	return &KnowledgeController{NewBaseController(validate), knowledge}
}

func (ctrl *KnowledgeController) body(c *gin.Context, input any) bool {
	if !ctrl.messagingResourceIDs(c) {
		return false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 128*1024)
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

// ListKnowledgeDocuments
// @Summary List the tenant's knowledge documents
// @Description Active tenant membership required. Business documents (FAQ, policies) the bot may answer from, newest change first. sync_state reports whether the AI knowledge index has the current version (pending, done, failed). can_manage is true for the tenant owner. Bot answers that rely on these documents always wait for staff review.
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param page query int false "Page"
// @Param size query int false "Page size (1-100)"
// @Success 200 {object} response.Response[dto.KnowledgeDocumentsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/knowledge-documents [get]
func (ctrl *KnowledgeController) ListKnowledgeDocuments(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	tenant, actor := salesScope(c)
	page, size := salesPage(c)
	result, err := ctrl.knowledge.List(c.Request.Context(), tenant, actor, page, size)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Documents(result, page, size))
}

// CreateKnowledgeDocument
// @Summary Add a knowledge document
// @Description Tenant owner only. title 1-200 and body 1-20000 characters after trimming. The document is indexed for the bot asynchronously (sync_state pending).
// @Tags Knowledge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.CreateKnowledgeDocumentRequest true "Document"
// @Success 200 {object} response.Response[dto.KnowledgeDocument]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/knowledge-documents [post]
func (ctrl *KnowledgeController) CreateKnowledgeDocument(c *gin.Context) {
	var input dto.CreateKnowledgeDocumentRequest
	if !ctrl.body(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.knowledge.Create(c.Request.Context(), tenant, actor, types.KnowledgeDocumentInput{Title: input.Title, Body: input.Body})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Document(*result))
}

// UpdateKnowledgeDocument
// @Summary Edit a knowledge document
// @Description Tenant owner only. expected_version is the version last read; a stale version is 409. Saving unchanged content keeps the version.
// @Tags Knowledge
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Document ID"
// @Param request body dto.UpdateKnowledgeDocumentRequest true "Document"
// @Success 200 {object} response.Response[dto.KnowledgeDocument]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/knowledge-documents/{id} [put]
func (ctrl *KnowledgeController) UpdateKnowledgeDocument(c *gin.Context) {
	var input dto.UpdateKnowledgeDocumentRequest
	if !ctrl.body(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.knowledge.Update(c.Request.Context(), tenant, actor, c.Param("id"), types.KnowledgeDocumentInput{Title: input.Title, Body: input.Body, ExpectedVersion: input.ExpectedVersion})
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Document(*result))
}

// DeleteKnowledgeDocument
// @Summary Delete a knowledge document
// @Description Tenant owner only. expected_version is the version last read; a stale version is 409. The bot stops using the document once the deletion is synced.
// @Tags Knowledge
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Document ID"
// @Param expected_version query int true "Version last read"
// @Success 200 {object} response.Response[dto.KnowledgeDocumentDeletedResponse]
// @Failure 400,401,403,404,409 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/knowledge-documents/{id} [delete]
func (ctrl *KnowledgeController) DeleteKnowledgeDocument(c *gin.Context) {
	if !ctrl.messagingResourceIDs(c) {
		return
	}
	var query dto.KnowledgeDocumentDeleteQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid query"))
		return
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), query); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	tenant, actor := salesScope(c)
	id := c.Param("id")
	if err := ctrl.knowledge.Delete(c.Request.Context(), tenant, actor, id, query.ExpectedVersion); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.KnowledgeDocumentDeletedResponse{ID: id, Deleted: true})
}
