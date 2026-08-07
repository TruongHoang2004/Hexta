package entity

import (
	"gitlab.com/ecommercehub1/identity-service/internal/domain/valueobject"
)

type Role struct {
	ID          valueobject.RoleID
	TenantID    valueobject.TenantID
	Name        string
	Description string
	IsSystem    bool
	Permissions []string
}

func NewRole(tenantID valueobject.TenantID, name, description string, permissions []string) *Role {
	return &Role{
		ID:          valueobject.NewRoleID(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		IsSystem:    false,
		Permissions: permissions,
	}
}

func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
