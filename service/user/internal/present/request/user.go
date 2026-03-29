package request

import pb "gitlab.com/ecommercehub1/user/proto"

type UserFilterField string

const (
	UserFilterFieldUserName    UserFilterField = "user_name"
	UserFilterFieldFullName    UserFilterField = "full_name"
	UserFilterFieldEmail       UserFilterField = "email"
	UserFilterFieldGender      UserFilterField = "gender"
	UserFilterFieldPhone       UserFilterField = "phone"
	UserFilterFieldDateOfBirth UserFilterField = "date_of_birth"
)

func NewUserPaginationRequest(req *pb.ListUsersRequest) *PaginationRequest {
	filters := make(map[string]string)
	for key, val := range req.GetFilters() {
		fieldStr := mapUserFilterFieldToModel(pb.UserFilterField(key))
		if fieldStr != "" {
			filters[fieldStr] = val
		}
	}

	sortField := mapUserFilterFieldToModel(req.GetSort())
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

func mapUserFilterFieldToModel(field pb.UserFilterField) string {
	switch field {
	case pb.UserFilterField_USER_FILTER_FIELD_USER_NAME:
		return "user_name"
	case pb.UserFilterField_USER_FILTER_FIELD_FULL_NAME:
		return "full_name"
	case pb.UserFilterField_USER_FILTER_FIELD_EMAIL:
		return "email"
	case pb.UserFilterField_USER_FILTER_FIELD_GENDER:
		return "gender"
	case pb.UserFilterField_USER_FILTER_FIELD_PHONE:
		return "phone"
	case pb.UserFilterField_USER_FILTER_FIELD_DATE_OF_BIRTH:
		return "date_of_birth"
	default:
		return ""
	}
}
