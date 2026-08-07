package valueobject

import (
	"errors"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

func NewUserID() UserID {
	return UserID{value: uuid.New()}
}

func ParseUserID(id string) (UserID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return UserID{}, errors.New("invalid user id format")
	}
	return UserID{value: parsed}, nil
}

func (u UserID) String() string {
	return u.value.String()
}

func (u UserID) Value() uuid.UUID {
	return u.value
}
