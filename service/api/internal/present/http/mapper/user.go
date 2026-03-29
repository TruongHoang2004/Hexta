package mapper

import (
	"strings"

	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/user/proto"
)

func getFilterField(sort string) *proto.UserFilterField {
	if sort == "" {
		return nil
	}
	sort = strings.ToUpper(strings.TrimSpace(sort))
	if val, ok := proto.UserFilterField_value[sort]; ok {
		typedVal := proto.UserFilterField(val)
		return &typedVal
	}
	if val, ok := proto.UserFilterField_value["USER_FILTER_FIELD_"+sort]; ok {
		typedVal := proto.UserFilterField(val)
		return &typedVal
	}
	return nil
}

func getOrderBy(order string) *proto.OrderBy {
	if order == "" {
		return nil
	}
	order = strings.ToUpper(strings.TrimSpace(order))
	switch order {
	case "ASC":
		v := proto.OrderBy_ORDER_BY_ASC
		return &v
	case "DESC":
		v := proto.OrderBy_ORDER_BY_DESC
		return &v
	default:
		return nil
	}
}

func ListUserRequestToPb(req dto.ListUsersRequest) *proto.ListUsersRequest {

	filter := make(map[int32]string)
	if req.Filter.UserName != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_USER_NAME)] = req.Filter.UserName
	}
	if req.Filter.FullName != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_FULL_NAME)] = req.Filter.FullName
	}
	if req.Filter.Email != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_EMAIL)] = req.Filter.Email
	}
	if req.Filter.Gender != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_GENDER)] = req.Filter.Gender
	}
	if req.Filter.Phone != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_PHONE)] = req.Filter.Phone
	}
	if req.Filter.DateOfBirth != "" {
		filter[int32(proto.UserFilterField_USER_FILTER_FIELD_DATE_OF_BIRTH)] = req.Filter.DateOfBirth
	}

	return &proto.ListUsersRequest{
		Page:    &req.Page,
		Limit:   &req.Size,
		Filters: filter,
		Sort:    getFilterField(req.SortBy),
		Order:   getOrderBy(req.Order),
		Search:  &req.Search,
	}
}

func GetUserFilterFieldEnum(field string) proto.UserFilterField {
	switch field {
	case "user_name":
		return proto.UserFilterField_USER_FILTER_FIELD_USER_NAME
	case "full_name":
		return proto.UserFilterField_USER_FILTER_FIELD_FULL_NAME
	case "email":
		return proto.UserFilterField_USER_FILTER_FIELD_EMAIL
	case "gender":
		return proto.UserFilterField_USER_FILTER_FIELD_GENDER
	case "phone":
		return proto.UserFilterField_USER_FILTER_FIELD_PHONE
	case "date_of_birth":
		return proto.UserFilterField_USER_FILTER_FIELD_DATE_OF_BIRTH
	default:
		return proto.UserFilterField_USER_FILTER_FIELD_UNSPECIFIED
	}
}
