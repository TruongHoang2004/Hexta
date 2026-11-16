package dto

import "time"

type StartFacebookOAuthRequest struct {
	RedirectURI string `json:"redirect_uri" validate:"required,url,max=2048"`
}
type StartFacebookOAuthResponse struct {
	AuthURL   string    `json:"auth_url"`
	ExpiresAt time.Time `json:"expires_at"`
}
type FacebookOAuthCallbackQuery struct {
	Code  string `form:"code" validate:"omitempty,max=2048"`
	State string `form:"state" validate:"required,max=128"`
	Error string `form:"error" validate:"omitempty,max=256"`
}
type SelectableFacebookPageDTO struct {
	PageID       string `json:"page_id"`
	PageName     string `json:"page_name"`
	Availability string `json:"availability"`
}
type FacebookOAuthPagesResponse struct {
	TenantID       string                      `json:"tenant_id"`
	OAuthSessionID string                      `json:"oauth_session_id"`
	ExpiresAt      time.Time                   `json:"expires_at"`
	Pages          []SelectableFacebookPageDTO `json:"pages"`
}
type ConnectFacebookPagesRequest struct {
	OAuthSessionID string   `json:"oauth_session_id" validate:"required,uuid"`
	PageIDs        []string `json:"page_ids" validate:"required,min=1,max=20,unique,dive,required,max=64"`
}
type FacebookConnectionDTO struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Platform         string    `json:"platform"`
	AccountID        string    `json:"account_id"`
	AccountName      string    `json:"account_name"`
	Version          int64     `json:"version"`
	Status           string    `json:"status"`
	LastErrorCode    string    `json:"last_error_code,omitempty"`
	LastErrorMessage string    `json:"last_error_message,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type ConnectFacebookPagesResponse struct {
	Connections []FacebookConnectionDTO `json:"connections"`
}
type ListFacebookConnectionsResponse struct {
	Connections []FacebookConnectionDTO `json:"connections"`
}
type UpdateFacebookConnectionRequest struct {
	Action string `json:"action" validate:"required,oneof=pause resume disconnect"`
}

// ConnectTelegramBotRequest carries a raw Bot Token. It is forwarded to the connector and never logged, stored or
// returned by the API.
type ConnectTelegramBotRequest struct {
	BotToken string `json:"bot_token" validate:"required,max=128"`
}

// UpdateTelegramConnectionRequest: bot_token is required for rotate_token and rejected otherwise.
type UpdateTelegramConnectionRequest struct {
	Action   string `json:"action" validate:"required,oneof=pause resume disconnect rotate_token"`
	BotToken string `json:"bot_token,omitempty" validate:"required_if=Action rotate_token,max=128"`
}
type TelegramConnectionDTO struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Platform         string    `json:"platform"`
	AccountID        string    `json:"account_id"`
	AccountName      string    `json:"account_name"`
	AccountUsername  string    `json:"account_username,omitempty"`
	Version          int64     `json:"version"`
	Status           string    `json:"status"`
	LastErrorCode    string    `json:"last_error_code,omitempty"`
	LastErrorMessage string    `json:"last_error_message,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type ListTelegramConnectionsResponse struct {
	Connections []TelegramConnectionDTO `json:"connections"`
}
