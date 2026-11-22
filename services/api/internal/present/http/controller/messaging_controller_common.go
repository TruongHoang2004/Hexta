package controller

import (
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (ctrl *baseController) messagingResourceIDs(c *gin.Context) bool {
	for _, id := range []string{c.Param("tenantID"), c.Param("id")} {
		if id != "" {
			if _, err := uuid.Parse(id); err != nil {
				ctrl.ErrorData(c, apperrors.ErrBadRequest(c.Request.Context()).SetDetail("invalid resource ID"))
				return false
			}
		}
	}
	return true
}
