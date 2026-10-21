package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	SessionID   int64  `json:"session_id"`
	UserID      string `json:"user_id"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterResponse struct {
	AccessToken string `json:"access_token"`
	SessionID   int64  `json:"session_id"`
	UserID      string `json:"user_id"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type LogoutRequest struct {
	SessionID int64 `json:"session_id,omitempty"`
}
