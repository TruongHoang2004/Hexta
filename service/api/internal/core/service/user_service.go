package service

import (
	"context"
	"time"

	"gitlab.com/ecommercehub1/api/common"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/mapper"
	"gitlab.com/ecommercehub1/api/pkg/client"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/proto"
)

type UserService struct {
	*baseService
	userClient *client.UserClient
}

func NewUserService(userClient *client.UserClient) *UserService {
	return &UserService{
		userClient: userClient,
	}
}

func (s *UserService) GetProfile(ctx context.Context) (*dto.UserResponse, *errors.Error) {
	authInfo, ok := common.GetAuthInfo(ctx)
	if !ok {
		return nil, errors.ErrUnauthorized(ctx).SetDetail("Can't get authInfo")
	}

	res, rpcErr := s.userClient.GetUser(ctx, &proto.GetUserRequest{
		Id: authInfo.UserID,
	})
	if rpcErr != nil {
		return nil, s.grpToIError(ctx, rpcErr)
	}

	return &dto.UserResponse{
		ID:          res.User.Id,
		UserName:    res.User.UserName,
		FullName:    res.User.FullName,
		Email:       *res.User.Email,
		Gender:      proto.Gender(res.User.Gender).String(),
		Phone:       *res.User.Phone,
		DateOfBirth: res.User.DateOfBirth,
		CreatedAt:   res.User.CreatedAt.AsTime().Format(time.RFC3339),
		UpdatedAt:   res.User.UpdatedAt.AsTime().Format(time.RFC3339),
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, *errors.Error) {
	res, rpcErr := s.userClient.GetUser(ctx, &proto.GetUserRequest{
		Id: id,
	})
	if rpcErr != nil {
		return nil, s.grpToIError(ctx, rpcErr)
	}

	return &dto.UserResponse{
		ID:          res.User.Id,
		UserName:    res.User.UserName,
		FullName:    res.User.FullName,
		Email:       *res.User.Email,
		Gender:      proto.Gender(res.User.Gender).String(),
		Phone:       *res.User.Phone,
		DateOfBirth: res.User.DateOfBirth,
		CreatedAt:   res.User.CreatedAt.AsTime().Format(time.RFC3339),
		UpdatedAt:   res.User.UpdatedAt.AsTime().Format(time.RFC3339),
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, *errors.Error) {

	pbReq := mapper.ListUserRequestToPb(*req)

	res, rpcErr := s.userClient.ListUsers(ctx, pbReq)
	if rpcErr != nil {
		return nil, s.grpToIError(ctx, rpcErr)
	}

	users := make([]dto.UserResponse, 0, len(res.Users))
	for _, u := range res.Users {
		email := ""
		if u.Email != nil {
			email = *u.Email
		}
		phone := ""
		if u.Phone != nil {
			phone = *u.Phone
		}

		users = append(users, dto.UserResponse{
			ID:          u.Id,
			UserName:    u.UserName,
			FullName:    u.FullName,
			Email:       email,
			Gender:      proto.Gender(u.Gender).String(),
			Phone:       phone,
			DateOfBirth: u.DateOfBirth,
			CreatedAt:   u.CreatedAt.AsTime().Format(time.RFC3339),
			UpdatedAt:   u.UpdatedAt.AsTime().Format(time.RFC3339),
		})
	}

	return &dto.ListUsersResponse{
		PaginationResponse: dto.NewPaginationResponse(users, res.Total, req.PaginationRequest),
	}, nil
}
