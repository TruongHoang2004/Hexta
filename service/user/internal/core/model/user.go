package model

import (
	"time"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	FullName    string    `gorm:"not null" json:"full_name"`
	UserName    string    `gorm:"not null;unique" json:"user_name"`
	Email       string    `gorm:"not null;unique" json:"email"`
	Phone       string    `gorm:"not null;unique" json:"phone"`
	Gender      Gender    `gorm:"not null" json:"gender"`
	DateOfBirth string    `gorm:"not null" json:"date_of_birth"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
