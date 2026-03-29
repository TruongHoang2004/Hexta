package mapper

import (
	"time"

	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	pb "gitlab.com/ecommercehub1/merchant/proto"
)

func PbToMerchantResponse(merchant *pb.Merchant) *dto.MerchantResponse {
	return &dto.MerchantResponse{
		ID:          merchant.Id,
		Name:        merchant.Name,
		Logo:        merchant.LogoUrl,
		Description: merchant.Description,
		Phone:       merchant.Phone,
		Email:       merchant.Email,
		Status:      merchant.Status.String(),
		CreatedAt:   merchant.CreatedAt.AsTime().Format(time.RFC3339),
		UpdatedAt:   merchant.UpdatedAt.AsTime().Format(time.RFC3339),
	}
}
