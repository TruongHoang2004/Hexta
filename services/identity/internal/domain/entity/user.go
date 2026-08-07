package entity

import (
	"time"

	"gitlab.com/ecommercehub1/identity-service/internal/domain/valueobject"
)

type UserStatus string

const (
	UserStatusPending  UserStatus = "PENDING"
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusDeleted  UserStatus = "DELETED"
)

type User struct {
	ID        valueobject.UserID
	TenantID  valueobject.TenantID
	Email     valueobject.Email
	FullName  string
	AvatarURL string
	Status    UserStatus
	RoleID    valueobject.RoleID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(tenantID valueobject.TenantID, email valueobject.Email, fullName string, roleID valueobject.RoleID) *User {
	now := time.Now().UTC()
	return &User{
		ID:        valueobject.NewUserID(),
		TenantID:  tenantID,
		Email:     email,
		FullName:  fullName,
		Status:    UserStatusPending,
		RoleID:    roleID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now().UTC()
}
