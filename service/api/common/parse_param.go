package common

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/utils/casting"
)

func ParseUintParam(ctx *gin.Context, param string) (uint, error) {
	valueStr := ctx.Param(param)
	valueUint, err := casting.StringToUint(valueStr)
	if err != nil {
		return 0, err
	}
	return valueUint, nil
}

func ParseIntParam(ctx *gin.Context, param string) (int, error) {
	valueStr := ctx.Param(param)
	valueInt, err := casting.StringToInt(valueStr)
	if err != nil {
		return 0, err
	}
	return valueInt, nil
}

func ParsePaginationParams(ctx *gin.Context) (*dto.PaginationRequest, error) {
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
			return nil, err
		}
	}

	if sizeStr == "" {
		size = 10
	} else {
		size, err = casting.StringToInt(sizeStr)
		if err != nil {
			return nil, err
		}
	}

	return &dto.PaginationRequest{
		Page: int32(page),
		Size: int32(size),
	}, nil
}
