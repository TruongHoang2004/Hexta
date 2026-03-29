package handler

import (
	"context"

	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/core/service"
	"gitlab.com/ecommercehub1/user/internal/present/request"
	pb "gitlab.com/ecommercehub1/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcHandler struct {
	*baseHandler
	pb.UnimplementedUserServiceServer
	service *service.UserService
}

func NewGrpcHandler(service *service.UserService) *GrpcHandler {
	return &GrpcHandler{
		service: service,
	}
}

func (h *GrpcHandler) Register(s *grpc.Server) {
	pb.RegisterUserServiceServer(s, h)
}

func (h *GrpcHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Convert proto gender enum to model gender enum
	var gender model.Gender
	switch req.Gender {
	case pb.Gender_GENDER_MALE:
		gender = model.GenderMale
	case pb.Gender_GENDER_FEMALE:
		gender = model.GenderFemale
	default:
		gender = model.GenderOther
	}

	var email, phone string
	if req.Email != nil {
		email = *req.Email
	}
	if req.Phone != nil {
		phone = *req.Phone
	}

	user := &model.User{
		FullName:    req.FullName,
		Gender:      gender,
		DateOfBirth: req.DateOfBirth,
		UserName:    req.UserName,
		Email:       email,
		Phone:       phone,
	}

	createdUser, err := h.service.CreateUser(ctx, user)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.CreateUserResponse{
		User: mapUserToProto(createdUser),
	}, nil
}

func (h *GrpcHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.service.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.GetUserResponse{
		User: mapUserToProto(user),
	}, nil
}

func (h *GrpcHandler) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	users, err := h.service.GetUsersByIDs(ctx, req.Ids)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, mapUserToProto(user))
	}

	return &pb.GetUsersResponse{
		Users: pbUsers,
	}, nil
}

func (h *GrpcHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	user, err := h.service.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	if _, err := h.service.DeleteUser(ctx, user); err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

func (h *GrpcHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	paginationReq := request.NewUserPaginationRequest(req)
	users, total, err := h.service.ListUsers(ctx, paginationReq)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, mapUserToProto(user))
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: int32(total),
	}, nil
}

func mapUserToProto(user *model.User) *pb.User {
	gender := pb.Gender_GENDER_UNSPECIFIED
	switch user.Gender {
	case model.GenderMale:
		gender = pb.Gender_GENDER_MALE
	case model.GenderFemale:
		gender = pb.Gender_GENDER_FEMALE
	case model.GenderOther:
		gender = pb.Gender_GENDER_OTHER
	}

	return &pb.User{
		Id:          user.ID,
		FullName:    user.FullName,
		Gender:      gender,
		DateOfBirth: user.DateOfBirth,
		UserName:    user.UserName,
		Email:       &user.Email,
		Phone:       &user.Phone,
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
	}
}
