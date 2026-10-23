package controller

import (
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	_ "github.com/TruongHoang2004/Hexta/services/api/internal/present/http/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strconv"
	"time"
)

type OverviewController struct {
	*baseController
	overview *service.OverviewService
}

func NewOverviewController(validate *validator.Validate, overview *service.OverviewService) *OverviewController {
	return &OverviewController{NewBaseController(validate), overview}
}

// GetOverview
// @Summary Read the tenant owner overview
// @Description Current-state counts in independent order/work creation cohorts, with paginated pinned workflow rows. Dates use inclusive from and exclusive to; employee means assigned sale/current work responsibility.
// @Tags Overview
// @Produce json
// @Security BearerAuth
// @Param tenantID path string true "Tenant UUID"
// @Param from query string false "Inclusive RFC3339 creation timestamp (requires to)"
// @Param to query string false "Exclusive RFC3339 creation timestamp (requires from)"
// @Param employee_id query string false "Current member's opaque user ID"
// @Param assignment query string false "Assignment filter" Enums(all,assigned,unassigned) default(all)
// @Param page query int false "Page (1-100000)" default(1)
// @Param size query int false "Page size (1-100)" default(20)
// @Success 200 {object} response.Response[dto.OverviewResponse]
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantID}/overview [get]
func (ctrl *OverviewController) GetOverview(c *gin.Context) {
	query := types.OverviewQuery{Assignment: "all", Page: 1, Size: 20}
	values := c.Request.URL.Query()
	invalid := func() {
		ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid overview query"))
	}
	for _, key := range []string{"from", "to", "employee_id", "assignment", "page", "size"} {
		entries, exists := values[key]
		if !exists {
			continue
		}
		if len(entries) != 1 || entries[0] == "" {
			invalid()
			return
		}
		value := entries[0]
		switch key {
		case "from", "to":
			stamp, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				invalid()
				return
			}
			stamp = stamp.UTC()
			if key == "from" {
				query.From = &stamp
			} else {
				query.To = &stamp
			}
		case "employee_id":
			query.EmployeeID = &value
		case "assignment":
			query.Assignment = value
		case "page", "size":
			number, err := strconv.Atoi(value)
			if err != nil {
				invalid()
				return
			}
			if key == "page" {
				query.Page = number
			} else {
				query.Size = number
			}
		}
	}
	tenant, actor := salesScope(c)
	result, err := ctrl.overview.Read(c.Request.Context(), tenant, actor, query)
	if err != nil {
		ctrl.ErrorData(c, err)
		return
	}
	ctrl.Success(c, dto.OverviewResponse(*result))
}
