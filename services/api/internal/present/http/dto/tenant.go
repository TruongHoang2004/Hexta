package dto

import "time"

type CreateTenantRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
	Plan string `json:"plan"`
}

type UpdateTenantRequest struct {
	Name   *string `json:"name,omitempty"`
	Plan   *string `json:"plan,omitempty"`
	Status *string `json:"status,omitempty"`
}

type InviteMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role"`
}

type TenantResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantMemberResponse struct {
	ID        int64     `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
