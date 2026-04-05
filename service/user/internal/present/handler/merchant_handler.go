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

type MerchantGrpcHandler struct {
	*baseHandler
	pb.UnimplementedMerchantServiceServer
	service *service.MerchantService
}

func NewMerchantGrpcHandler(service *service.MerchantService) *MerchantGrpcHandler {
	return &MerchantGrpcHandler{
		baseHandler: NewBaseHandler(),
		service:     service,
	}
}

func (h *MerchantGrpcHandler) Register(s *grpc.Server) {
	pb.RegisterMerchantServiceServer(s, h)
}

func (h *MerchantGrpcHandler) CreateMerchant(ctx context.Context, req *pb.CreateMerchantRequest) (*pb.CreateMerchantResponse, error) {
	var description, logoURL string
	if req.Description != nil {
		description = *req.Description
	}
	if req.LogoUrl != nil {
		logoURL = *req.LogoUrl
	}

	merchant := &model.Merchant{
		OwnerID:     req.OwnerId,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Description: description,
		LogoURL:     logoURL,
		Status:      model.MerchantStatusPending,
	}

	created, err := h.service.CreateMerchant(ctx, merchant)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.CreateMerchantResponse{
		Merchant: mapMerchantToProto(created),
	}, nil
}

func (h *MerchantGrpcHandler) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*pb.GetMerchantResponse, error) {
	merchant, err := h.service.GetMerchantByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.GetMerchantResponse{
		Merchant: mapMerchantToProto(merchant),
	}, nil
}

func (h *MerchantGrpcHandler) GetMerchantByOwner(ctx context.Context, req *pb.GetMerchantByOwnerRequest) (*pb.GetMerchantByOwnerResponse, error) {
	merchant, err := h.service.GetMerchantByOwnerID(ctx, req.OwnerId)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.GetMerchantByOwnerResponse{
		Merchant: mapMerchantToProto(merchant),
	}, nil
}

func (h *MerchantGrpcHandler) UpdateMerchant(ctx context.Context, req *pb.UpdateMerchantRequest) (*pb.UpdateMerchantResponse, error) {
	existing, err := h.service.GetMerchantByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.LogoUrl != nil {
		existing.LogoURL = *req.LogoUrl
	}

	updated, err := h.service.UpdateMerchant(ctx, existing)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.UpdateMerchantResponse{
		Merchant: mapMerchantToProto(updated),
	}, nil
}

func (h *MerchantGrpcHandler) ListMerchants(ctx context.Context, req *pb.ListMerchantsRequest) (*pb.ListMerchantsResponse, error) {
	page := int32(1)
	limit := int32(20)
	if req.Page != nil {
		page = *req.Page
	}
	if req.Limit != nil {
		limit = *req.Limit
	}

	paginationReq := &request.PaginationRequest{
		Page:   page,
		Limit:  limit,
		Filter: make(map[string]string),
	}
	if req.Search != nil {
		paginationReq.Search = *req.Search
	}
	if req.Status != nil && *req.Status != pb.MerchantStatus_MERCHANT_STATUS_UNSPECIFIED {
		paginationReq.Filter["status"] = mapProtoStatusToModel(*req.Status)
	}

	merchants, total, err := h.service.ListMerchants(ctx, paginationReq)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	var pbMerchants []*pb.Merchant
	for _, m := range merchants {
		pbMerchants = append(pbMerchants, mapMerchantToProto(m))
	}

	return &pb.ListMerchantsResponse{
		Merchants: pbMerchants,
		Total:     int32(total),
	}, nil
}

func (h *MerchantGrpcHandler) DeleteMerchant(ctx context.Context, req *pb.DeleteMerchantRequest) (*pb.DeleteMerchantResponse, error) {
	merchant, err := h.service.GetMerchantByID(ctx, req.Id)
	if err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	if _, err := h.service.DeleteMerchant(ctx, merchant); err != nil {
		return nil, h.IErrorToGRPCError(err)
	}

	return &pb.DeleteMerchantResponse{Success: true}, nil
}

func mapMerchantToProto(m *model.Merchant) *pb.Merchant {
	return &pb.Merchant{
		Id:          m.ID,
		OwnerId:     m.OwnerID,
		Name:        m.Name,
		Email:       m.Email,
		Phone:       m.Phone,
		Description: m.Description,
		LogoUrl:     m.LogoURL,
		Status:      mapModelStatusToProto(m.Status),
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func mapModelStatusToProto(s model.MerchantStatus) pb.MerchantStatus {
	switch s {
	case model.MerchantStatusPending:
		return pb.MerchantStatus_MERCHANT_STATUS_PENDING
	case model.MerchantStatusActive:
		return pb.MerchantStatus_MERCHANT_STATUS_ACTIVE
	case model.MerchantStatusSuspended:
		return pb.MerchantStatus_MERCHANT_STATUS_SUSPENDED
	case model.MerchantStatusClosed:
		return pb.MerchantStatus_MERCHANT_STATUS_CLOSED
	default:
		return pb.MerchantStatus_MERCHANT_STATUS_UNSPECIFIED
	}
}

func mapProtoStatusToModel(s pb.MerchantStatus) string {
	switch s {
	case pb.MerchantStatus_MERCHANT_STATUS_PENDING:
		return string(model.MerchantStatusPending)
	case pb.MerchantStatus_MERCHANT_STATUS_ACTIVE:
		return string(model.MerchantStatusActive)
	case pb.MerchantStatus_MERCHANT_STATUS_SUSPENDED:
		return string(model.MerchantStatusSuspended)
	case pb.MerchantStatus_MERCHANT_STATUS_CLOSED:
		return string(model.MerchantStatusClosed)
	default:
		return ""
	}
}
