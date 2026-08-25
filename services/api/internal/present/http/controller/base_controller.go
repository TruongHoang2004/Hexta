package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gitlab.com/ecommercehub1/api/common/log"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/response"
	"gitlab.com/ecommercehub1/api/utils/casting"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type baseController struct {
	validate *validator.Validate
}

func NewBaseController(validate *validator.Validate) *baseController {
	return &baseController{
		validate: validate,
	}
}

func (b *baseController) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response.NewSuccessResponse(data, nil))
}

func (b *baseController) ErrorData(c *gin.Context, err *errors.Error) {
	log.IErr(c.Request.Context(), err)
	c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
}

func (b *baseController) GetUintParam(c *gin.Context, key string) (uint, *errors.Error) {
	param := c.Param(key)
	if param == "" {
		log.Warn(c, "param %s is empty", key)
		return 0, errors.ErrBadRequest(c).SetDetail(fmt.Sprintf("param %s is empty", key))
	}

	id, err := casting.StringToUint(param)
	if err != nil {
		log.Warn(c, "invalid param %s, err:[%s]", key, err)
		return 0, errors.ErrBadRequest(c).SetDetail(fmt.Sprintf("invalid param %s", key))
	}
	return id, nil
}

func (b *baseController) GetStringParams(c *gin.Context, key string) (string, *errors.Error) {
	param := c.Param(key)
	if param == "" {
		log.Warn(c, "param %s is empty", key)
		return "", errors.ErrBadRequest(c).SetDetail(fmt.Sprintf("param %s is empty", key))
	}
	return param, nil
}

func (b *baseController) GetFile(c *gin.Context, key string) (*multipart.FileHeader, *errors.Error) {
	file, err := c.FormFile(key)
	if err != nil {
		log.Warn(c, "get file %s err, err:[%s]", key, err)
		return nil, errors.ErrBadRequest(c).SetDetail(fmt.Sprintf("file %s is required", key))
	}
	return file, nil
}

func (b *baseController) GetPaginationParams(ctx *gin.Context) (*dto.PaginationRequest, *errors.Error) {
	pageStr := ctx.Query("page")
	sizeStr := ctx.Query("size")

	var (
		page int
		size int
		err  error
	)

	if pageStr == "" {
		page = 1
	} else {
		page, err = casting.StringToInt(pageStr)
		if err != nil {
			return nil, errors.ErrBadRequest(ctx).SetDetail(err.Error())
		}
	}

	if sizeStr == "" {
		size = 10
	} else {
		size, err = casting.StringToInt(sizeStr)
		if err != nil {
			return nil, errors.ErrBadRequest(ctx).SetDetail(err.Error())
		}
	}

	return &dto.PaginationRequest{
		Page: int32(page),
		Size: int32(size),
	}, nil
}

func (b *baseController) BindAndValidateRequest(c *gin.Context, req interface{}) *errors.Error {
	if err := c.BindUri(req); err != nil {
		log.Warn(c, "bind request err, err:[%s]", err)
		return errors.ErrBadRequest(c).SetDetail(err.Error())
	}
	if err := c.Bind(req); err != nil {
		log.Warn(c, "bind request err, err:[%s]", err)
		return errors.ErrBadRequest(c).SetDetail(err.Error())
	}
	return b.ValidateRequest(c, req)
}

func (b *baseController) ValidateRequest(ctx context.Context, req interface{}) *errors.Error {
	err := b.validate.Struct(req)

	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			log.Error(ctx, "Cannot parse validate error: %+v", err)
			return errors.ErrSystemError(ctx, "ValidateFailed").SetDetail(err.Error())
		}
		var filedErrors []string
		for _, errValidate := range errs {
			log.Debug(ctx, "field invalid, err:[%s]", errValidate.Field())
			filedErrors = append(filedErrors, errValidate.Error())
		}
		str := strings.Join(filedErrors, ",")
		log.Warn(ctx, "invalid request, err:[%s]", err.Error())
		return errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("field invalidate [%s]", str))
	}
	return nil
}
