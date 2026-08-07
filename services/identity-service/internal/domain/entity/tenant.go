package entity

import (
	"time"

	"gitlab.com/ecommercehub1/identity-service/internal/domain/valueobject"
)

type TenantStatus string
type TenantPlan string

const (
	TenantStatusPending   TenantStatus = "PENDING"
	TenantStatusActive    TenantStatus = "ACTIVE"
	TenantStatusSuspended TenantStatus = "SUSPENDED"
	TenantStatusDeleted   TenantStatus = "DELETED"

	TenantPlanFree       TenantPlan = "FREE"
	TenantPlanPro        TenantPlan = "PRO"
	TenantPlanEnterprise TenantPlan = "ENTERPRISE"
)

type Tenant struct {
	ID           valueobject.TenantID
	Name         string
	Slug         string
	Status       TenantStatus
	Plan         TenantPlan
	OwnerID      valueobject.UserID
	ContactEmail valueobject.Email
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewTenant(name string, slug string, contactEmail valueobject.Email, ownerID valueobject.UserID) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		ID:           valueobject.NewTenantID(),
		Name:         name,
		Slug:         slug,
		Status:       TenantStatusPending,
		Plan:         TenantPlanFree,
		OwnerID:      ownerID,
		ContactEmail: contactEmail,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (t *Tenant) Activate() {
	t.Status = TenantStatusActive
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tenant) Suspend() {
	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now().UTC()
}
