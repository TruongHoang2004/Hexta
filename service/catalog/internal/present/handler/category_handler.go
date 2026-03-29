package handler

import (
	"context"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/catalog/internal/core/service"
	pb "gitlab.com/ecommercehub1/catalog/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CategoryGrpcHandler struct {
	*baseHandler
	pb.UnimplementedCategoryServiceServer
	service *service.CategoryService
}

func NewCategoryGrpcHandler(service *service.CategoryService) *CategoryGrpcHandler {
	return &CategoryGrpcHandler{
		service: service,
	}
}

func (h *CategoryGrpcHandler) Register(s *grpc.Server) {
	pb.RegisterCategoryServiceServer(s, h)
}

func (h *CategoryGrpcHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	category := &model.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	createdCategory, err := h.service.CreateCategory(ctx, category)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.CreateCategoryResponse{
		Category: mapCategoryToProto(createdCategory),
	}, nil
}

func (h *CategoryGrpcHandler) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	category, err := h.service.GetCategoryByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.GetCategoryResponse{
		Category: mapCategoryToProto(category),
	}, nil
}

func (h *CategoryGrpcHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	categories, err := h.service.ListAllCategories(ctx)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbCategories []*pb.Category
	for _, category := range categories {
		pbCategories = append(pbCategories, mapCategoryToProto(category))
	}

	return &pb.ListCategoriesResponse{
		Categories: pbCategories,
	}, nil
}

func (h *CategoryGrpcHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	category, err := h.service.GetCategoryByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	if _, err := h.service.DeleteCategory(ctx, category); err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.DeleteCategoryResponse{
		Success: true,
	}, nil
}

func mapCategoryToProto(category *model.Category) *pb.Category {
	return &pb.Category{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   timestamppb.New(category.CreatedAt),
		UpdatedAt:   timestamppb.New(category.UpdatedAt),
	}
}
