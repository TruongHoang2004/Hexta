package controller

import (
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/overview/httpmap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type OverviewController struct {
	*baseController
	overview *service.OverviewService
}

func NewOverviewController(validate *validator.Validate, overview *service.OverviewService) *OverviewController {
	return &OverviewController{NewBaseController(validate), overview}
}

// Overview
// @Summary Read the tenant owner overview
// @Description Read-only and owner-only. Counts orders and work created in the half-open interval [from, to) (all time when both are omitted). employee_id narrows orders to their assigned sale and work to its current responsible member. Order value sums confirmed, processing and fulfilled order totals; it is not revenue. The stuck list holds open work that is unassigned or unchanged for longer than threshold_hours, longest waiting first, capped at 50 rows with the full total.
// @Tags Overview
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant ID"
// @Param from query string false "Inclusive lower creation bound, RFC3339 with offset; requires to"
// @Param to query string false "Exclusive upper creation bound, RFC3339 with offset; requires from"
// @Param employee_id query string false "Current tenant member ID"
// @Success 200 {object} response.Response[dto.OverviewResponse]
// @Failure 400,401,403,404,500 {object} apperrors.ErrorResponse
// @Router /api/v1/tenants/{tenantID}/overview [get]
func (ctrl *OverviewController) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	if _, err := uuid.Parse(c.Param("tenantID")); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(ctx).SetDetail("invalid resource ID"))
		return
	}
	for _, key := range []string{"from", "to", "employee_id"} {
		if len(c.QueryArray(key)) > 1 {
			ctrl.ErrorData(c, apperrors.ErrBadRequest(ctx).SetDetail("repeated query parameter "+key))
			return
		}
	}
	var query dto.OverviewQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(ctx).SetDetail("invalid overview query"))
		return
	}
	if err := ctrl.ValidateRequest(ctx, &query); err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	filter := types.OverviewFilter{}
	var bad bool
	if filter.From, bad = overviewTime(query.From); bad {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(ctx).SetDetail("from must be an RFC3339 timestamp with offset"))
		return
	}
	if filter.To, bad = overviewTime(query.To); bad {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(ctx).SetDetail("to must be an RFC3339 timestamp with offset"))
		return
	}
	if query.EmployeeID != "" {
		filter.EmployeeID = &query.EmployeeID
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.overview.Overview(ctx, tenant, actor, filter)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, httpmap.Overview(*result))
}

// overviewTime parses an optional RFC3339 instant; an empty value means "not given".
func overviewTime(value string) (*time.Time, bool) {
	if value == "" {
		return nil, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, true
	}
	return &parsed, false
}
