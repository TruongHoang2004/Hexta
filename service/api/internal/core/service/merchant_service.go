package service

import (
	"context"

	"gitlab.com/ecommercehub1/api/common"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/present/http/mapper"
	"gitlab.com/ecommercehub1/api/pkg/client"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	pb "gitlab.com/ecommercehub1/user/proto"
)

type MerchantService struct {
	*baseService
	merchantClient *client.MerchantClient
}

func NewMerchantService(
	merchantClient *client.MerchantClient,
) *MerchantService {
	return &MerchantService{
		merchantClient: merchantClient,
	}
}

func (s *MerchantService) CreateMerchant(ctx context.Context, req *dto.CreateMerchantRequest) (*dto.MerchantResponse, *errors.Error) {
	authInfo, ok := common.GetAuthInfo(ctx)
	if !ok {
		return nil, errors.ErrUnauthorized(ctx).SetDetail("Can't get authInfo")
	}
	ownerId := authInfo.UserID
	res, gpcErr := s.merchantClient.CreateMerchant(ctx, &pb.CreateMerchantRequest{
		Name:        req.Name,
		LogoUrl:     &req.Logo,
		Description: &req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		OwnerId:     ownerId,
	})
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	return mapper.PbToMerchantResponse(res.Merchant), nil
}

func (s *MerchantService) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*dto.MerchantResponse, *errors.Error) {
	res, gpcErr := s.merchantClient.GetMerchant(ctx, req)
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	return mapper.PbToMerchantResponse(res.Merchant), nil
}

func (s *MerchantService) GetMerchantByOwner(ctx context.Context, req *pb.GetMerchantByOwnerRequest) (*dto.MerchantResponse, *errors.Error) {
	res, gpcErr := s.merchantClient.GetMerchantByOwner(ctx, req)
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	return mapper.PbToMerchantResponse(res.Merchant), nil
}

func (s *MerchantService) UpdateMerchant(ctx context.Context, req *dto.UpdateMerchantRequest) (*dto.MerchantResponse, *errors.Error) {
	res, gpcErr := s.merchantClient.UpdateMerchant(ctx, &pb.UpdateMerchantRequest{
		Name:        &req.Name,
		LogoUrl:     &req.Logo,
		Description: &req.Description,
		Phone:       &req.Phone,
		Email:       &req.Email,
		Id:          req.Id,
	})
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	return mapper.PbToMerchantResponse(res.Merchant), nil
}

func (s *MerchantService) ListMerchants(ctx context.Context, req *dto.ListMerchantsRequest) (*dto.ListMerchantsResponse, *errors.Error) {
	res, gpcErr := s.merchantClient.ListMerchants(ctx, &pb.ListMerchantsRequest{
		Page:   &req.Page,
		Limit:  &req.Size,
		Search: &req.Search,
	})
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	merchants := make([]dto.MerchantResponse, len(res.Merchants))
	for i, merchant := range res.Merchants {
		merchants[i] = *mapper.PbToMerchantResponse(merchant)
	}

	return &dto.ListMerchantsResponse{
		PaginationResponse: dto.NewPaginationResponse(merchants, res.Total, req.PaginationRequest),
	}, nil
}

func (s *MerchantService) DeleteMerchant(ctx context.Context, req *pb.DeleteMerchantRequest) (*pb.DeleteMerchantResponse, *errors.Error) {
	res, gpcErr := s.merchantClient.DeleteMerchant(ctx, req)
	if gpcErr != nil {
		return nil, s.grpToIError(ctx, gpcErr)
	}

	return res, nil
}
