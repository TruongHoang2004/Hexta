package valueobject

import (
	"errors"

	"github.com/google/uuid"
)

type RoleID struct {
	value uuid.UUID
}

func NewRoleID() RoleID {
	return RoleID{value: uuid.New()}
}

func ParseRoleID(id string) (RoleID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return RoleID{}, errors.New("invalid role id format")
	}
	return RoleID{value: parsed}, nil
}

func (r RoleID) String() string {
	return r.value.String()
}

func (r RoleID) Value() uuid.UUID {
	return r.value
}
