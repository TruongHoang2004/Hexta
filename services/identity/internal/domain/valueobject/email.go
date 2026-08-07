package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRegex.MatchString(email) {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: email}, nil
}

func (e Email) String() string {
	return e.value
}
