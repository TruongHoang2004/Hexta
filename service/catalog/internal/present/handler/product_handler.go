package handler

import (
	"context"

	"gitlab.com/ecommercehub1/catalog/internal/core/model"
	"gitlab.com/ecommercehub1/catalog/internal/core/service"
	"gitlab.com/ecommercehub1/catalog/internal/present/request"
	pb "gitlab.com/ecommercehub1/catalog/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductGrpcHandler struct {
	*baseHandler
	pb.UnimplementedProductServiceServer
	service *service.ProductService
}

func NewProductGrpcHandler(service *service.ProductService) *ProductGrpcHandler {
	return &ProductGrpcHandler{
		service: service,
	}
}

func (h *ProductGrpcHandler) Register(s *grpc.Server) {
	pb.RegisterProductServiceServer(s, h)
}

func (h *ProductGrpcHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       float64(req.Price),
		CategoryID:  req.CategoryId,
	}

	createdProduct, err := h.service.CreateProduct(ctx, product)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.CreateProductResponse{
		Product: mapProductToProto(createdProduct),
	}, nil
}

func (h *ProductGrpcHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := h.service.GetProductByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.GetProductResponse{
		Product: mapProductToProto(product),
	}, nil
}

func (h *ProductGrpcHandler) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	products, err := h.service.GetProductsByIDs(ctx, req.Ids)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbProducts []*pb.Product
	for _, product := range products {
		pbProducts = append(pbProducts, mapProductToProto(product))
	}

	return &pb.GetProductsResponse{
		Products: pbProducts,
	}, nil
}

func (h *ProductGrpcHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	product, err := h.service.GetProductByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	if _, err := h.service.DeleteProduct(ctx, product); err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.DeleteProductResponse{
		Success: true,
	}, nil
}

func (h *ProductGrpcHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	paginationReq := request.NewProductPaginationRequest(req)
	products, total, err := h.service.ListProducts(ctx, paginationReq)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbProducts []*pb.Product
	for _, product := range products {
		pbProducts = append(pbProducts, mapProductToProto(product))
	}

	return &pb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total),
	}, nil
}

func mapProductToProto(product *model.Product) *pb.Product {
	return &pb.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CategoryId:  product.CategoryID,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}
}
