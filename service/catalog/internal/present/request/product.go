package request

import pb "gitlab.com/ecommercehub1/catalog/proto"

type ProductFilterField string

const (
	ProductFilterFieldName       ProductFilterField = "name"
	ProductFilterFieldCategoryID ProductFilterField = "category_id"
)

func NewProductPaginationRequest(req *pb.ListProductsRequest) *PaginationRequest {
	filters := make(map[string]string)
	for key, val := range req.GetFilters() {
		fieldStr := mapProductFilterFieldToModel(pb.ProductFilterField(key))
		if fieldStr != "" {
			filters[fieldStr] = val
		}
	}

	sortField := mapProductFilterFieldToModel(req.GetSort())
	order := ""
	if req.GetOrder() == pb.OrderBy_ORDER_BY_ASC {
		order = "asc"
	}
	if req.GetOrder() == pb.OrderBy_ORDER_BY_DESC {
		order = "desc"
	}

	return &PaginationRequest{
		Page:   req.GetPage(),
		Limit:  req.GetLimit(),
		Filter: filters,
		Sort:   sortField,
		Order:  order,
		Search: req.GetSearch(),
	}
}

func mapProductFilterFieldToModel(field pb.ProductFilterField) string {
	switch field {
	case pb.ProductFilterField_PRODUCT_FILTER_FIELD_NAME:
		return "name"
	case pb.ProductFilterField_PRODUCT_FILTER_FIELD_CATEGORY_ID:
		return "category_id"
	default:
		return ""
	}
}
