package types

type AccessTokenPayload struct {
	UserID    string `json:"user_id"`
	SessionID int64  `json:"session_id"`
}

type RefreshTokenPayload struct {
	UserID string `json:"user_id"`
}

type AuthInfo struct {
	UserID    string `json:"user_id"`
	SessionID int64  `json:"session_id"`
}
