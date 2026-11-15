// Package httpmap converts channel use-case values to HTTP DTOs. Credentials never exist in these values.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func Connection(value types.ChannelConnection) dto.FacebookConnectionDTO {
	return dto.FacebookConnectionDTO{ID: value.ID, TenantID: value.TenantID, Platform: value.Platform, AccountID: value.AccountID, AccountName: value.AccountName, Version: value.Version, Status: value.Status, LastErrorCode: value.LastErrorCode, LastErrorMessage: value.LastErrorMessage, UpdatedAt: value.UpdatedAt}
}
func Connections(values []types.ChannelConnection) []dto.FacebookConnectionDTO {
	result := make([]dto.FacebookConnectionDTO, len(values))
	for i, v := range values {
		result[i] = Connection(v)
	}
	return result
}
func OAuthStart(value types.ChannelOAuthStart) dto.StartFacebookOAuthResponse {
	return dto.StartFacebookOAuthResponse{AuthURL: value.AuthURL, ExpiresAt: value.ExpiresAt}
}
func OAuthPages(value types.ChannelOAuthPages) dto.FacebookOAuthPagesResponse {
	result := dto.FacebookOAuthPagesResponse{TenantID: value.TenantID, OAuthSessionID: value.OAuthSessionID, ExpiresAt: value.ExpiresAt, Pages: make([]dto.SelectableFacebookPageDTO, len(value.Pages))}
	for i, p := range value.Pages {
		result.Pages[i] = dto.SelectableFacebookPageDTO{PageID: p.PageID, PageName: p.PageName, Availability: p.Availability}
	}
	return result
}
