package dto

import "github.com/TruongHoang2004/Hexta/services/api/internal/core/model"

type TenantApplicationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}
type TenantReviewRequest struct {
	Decision string `json:"decision" validate:"required,oneof=approved rejected"`
	Reason   string `json:"reason" validate:"max=2000"`
}
type TenantEmailInvitationRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}
type TenantInvitationTokenRequest struct {
	Token string `json:"token" validate:"required,len=64"`
}
type TenantApplicationResponse = model.TenantRegistration
type TenantApplicationsResponse struct {
	Items []model.TenantRegistration `json:"items"`
}
type TenantInvitationsResponse struct {
	Items []model.TenantInvitation `json:"items"`
}
type TenantInvitationResponse struct {
	Invitation *model.TenantInvitation `json:"invitation"`
	Token      string                  `json:"token,omitempty"`
	TenantName string                  `json:"tenant_name,omitempty"`
	Delivery   string                  `json:"delivery,omitempty"`
}
type PlatformAccessResponse struct {
	Operator bool `json:"operator"`
}
type TenantRevokedResponse struct {
	Revoked bool `json:"revoked"`
}
