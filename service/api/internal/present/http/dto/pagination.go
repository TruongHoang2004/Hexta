package dto

import "strings"

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationRequest struct {
	Page   int32  `json:"page" form:"page" query:"page"`
	Size   int32  `json:"size" form:"size" query:"size"`
	SortBy string `json:"sort_by" form:"sort_by" query:"sort_by"`
	Order  string `json:"order" form:"order" query:"order"` // "ASC" or "DESC"
}

func (p *PaginationRequest) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Size <= 0 {
		p.Size = DefaultPageSize
	}
	if p.Size > MaxPageSize {
		p.Size = MaxPageSize
	}
	p.Order = strings.ToUpper(strings.TrimSpace(p.Order))
	if p.Order != "ASC" && p.Order != "DESC" {
		p.Order = "ASC"
	}
}

func (p PaginationRequest) Offset() int32 {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Size
}

func (p PaginationRequest) Limit() int32 {
	if p.Size <= 0 {
		return DefaultPageSize
	}
	return p.Size
}

type PaginationResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int32 `json:"total"`
	Page       int32 `json:"page"`
	Size       int32 `json:"size"`
	TotalPages int32 `json:"total_pages"`
}

func NewPaginationResponse[T any](items []T, total int32, req PaginationRequest) PaginationResponse[T] {
	req.Normalize()
	var totalPages int32
	if total > 0 {
		totalPages = (total + req.Size - 1) / req.Size
	}
	return PaginationResponse[T]{
		Data:       items,
		Total:      total,
		Page:       req.Page,
		Size:       req.Size,
		TotalPages: totalPages,
	}
}
