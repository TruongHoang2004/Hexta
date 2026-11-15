package controller

import (
	"net/http"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/channel/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ChannelController struct {
	*baseController
	channels *service.ChannelService
}

func NewChannelController(validate *validator.Validate, s *service.ChannelService) *ChannelController {
	return &ChannelController{NewBaseController(validate), s}
}

func (ctrl *ChannelController) bind(c *gin.Context, input any) bool {
	for _, id := range []string{c.Param("tenantID"), c.Param("id")} {
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
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

// StartFacebookOAuth
// @Summary Start Facebook Page authorization
// @Description Tenant owner only. Returns the Meta OAuth dialog URL with a single-use state valid for 10 minutes. redirect_uri must be the web app /channels/facebook/callback on an allowed origin.
// @Tags Channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.StartFacebookOAuthRequest true "Redirect"
// @Success 200 {object} response.Response[dto.StartFacebookOAuthResponse]
// @Failure 400,401,403,503 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/channels/facebook/oauth/start [post]
func (ctrl *ChannelController) StartFacebookOAuth(c *gin.Context) {
	var input dto.StartFacebookOAuthRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.channels.StartFacebookOAuth(c.Request.Context(), tenant, actor, input.RedirectURI)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.OAuthStart(*result))
}

// FacebookOAuthCallback
// @Summary Complete Facebook Page authorization
// @Description Called by the web callback page with Meta's code and state. Consumes the state (single use), requires the same authenticated owner that started the flow, and returns selectable Pages without tokens.
// @Tags Channels
// @Produce json
// @Security BearerAuth
// @Param code query string false "OAuth code"
// @Param state query string true "OAuth state"
// @Param error query string false "Provider error"
// @Success 200 {object} response.Response[dto.FacebookOAuthPagesResponse]
// @Failure 400,401,403,422,502,503 {object} apperrors.ErrorResponse
// @Router /api/v1/channels/facebook/oauth/callback [get]
func (ctrl *ChannelController) FacebookOAuthCallback(c *gin.Context) {
	var input dto.FacebookOAuthCallbackQuery
	if err := c.ShouldBindQuery(&input); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid query"))
		return
	}
	if err := ctrl.ValidateRequest(c.Request.Context(), input); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	_, actor := salesScope(c)
	result, err := ctrl.channels.CompleteFacebookOAuth(c.Request.Context(), actor, input.Code, input.State, input.Error)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	ctrl.Success(c, httpmap.OAuthPages(*result))
}

// ConnectFacebookPages
// @Summary Connect selected Facebook Pages
// @Description Tenant owner only. Pages must come from the owner's own unexpired OAuth session. A Page connected to another workspace returns 409 (1 Page = 1 tenant); a Page already connected here is re-authorized.
// @Tags Channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param request body dto.ConnectFacebookPagesRequest true "Pages"
// @Success 200 {object} response.Response[dto.ConnectFacebookPagesResponse]
// @Failure 400,401,403,409,422,502,503 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/channels/facebook/connections [post]
func (ctrl *ChannelController) ConnectFacebookPages(c *gin.Context) {
	var input dto.ConnectFacebookPagesRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.channels.ConnectFacebookPages(c.Request.Context(), tenant, actor, input.OAuthSessionID, input.PageIDs)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.ConnectFacebookPagesResponse{Connections: httpmap.Connections(result)})
}

// ListFacebookConnections
// @Summary List Facebook Page connections
// @Description Active tenant membership required. Returns non-disconnected connections with status (active, paused, auth_expired) for settings and Inbox banners.
// @Tags Channels
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Success 200 {object} response.Response[dto.ListFacebookConnectionsResponse]
// @Failure 400,401,403 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/channels/facebook/connections [get]
func (ctrl *ChannelController) ListFacebookConnections(c *gin.Context) {
	if !ctrl.bind(c, nil) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.channels.ListFacebookConnections(c.Request.Context(), tenant, actor)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.ListFacebookConnectionsResponse{Connections: httpmap.Connections(result)})
}

// UpdateFacebookConnection
// @Summary Pause, resume or disconnect a Facebook Page
// @Description Tenant owner only. Pause/disconnect cancel pending outbound intents; disconnect wipes the stored Page credential and unsubscribes webhooks. Repeating an action in effect re-syncs the connector.
// @Tags Channels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param id path string true "Connection ID"
// @Param request body dto.UpdateFacebookConnectionRequest true "Action"
// @Success 200 {object} response.Response[dto.FacebookConnectionDTO]
// @Failure 400,401,403,404,409,502,503 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/channels/facebook/connections/{id} [patch]
func (ctrl *ChannelController) UpdateFacebookConnection(c *gin.Context) {
	var input dto.UpdateFacebookConnectionRequest
	if !ctrl.bind(c, &input) {
		return
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.channels.UpdateFacebookConnection(c.Request.Context(), tenant, actor, c.Param("id"), input.Action)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Connection(*result))
}
