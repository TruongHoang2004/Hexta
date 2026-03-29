package dto

import "time"

type LocalLoginRequest struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type RegisterRequest struct {
	UserName    string    `json:"user_name" validate:"required"`
	FullName    string    `json:"full_name" validate:"required"`
	Email       string    `json:"email" validate:"required"`
	Password    string    `json:"password" validate:"required"`
	Gender      Gender    `json:"gender" validate:"required"`
	Phone       string    `json:"phone" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
}

type RegisterResponse struct {
	ID          string    `json:"id"`
	UserName    string    `json:"user_name"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Gender      Gender    `json:"gender"`
	Phone       string    `json:"phone"`
	DateOfBirth time.Time `json:"date_of_birth"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LoginRequest struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
