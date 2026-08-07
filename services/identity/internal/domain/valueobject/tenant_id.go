package valueobject

import (
	"errors"

	"github.com/google/uuid"
)

type TenantID struct {
	value uuid.UUID
}

func NewTenantID() TenantID {
	return TenantID{value: uuid.New()}
}

func ParseTenantID(id string) (TenantID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return TenantID{}, errors.New("invalid tenant id format")
	}
	return TenantID{value: parsed}, nil
}

func (t TenantID) String() string {
	return t.value.String()
}

func (t TenantID) Value() uuid.UUID {
	return t.value
}
