package models

import (
	"time"
)

type User struct {
	UserID    int        `json:"user_id" db:"user_id"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	Role      string     `json:"-" db:"role"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty" db:"last_login"`
}

type Profile struct {
	ProfileID   int       `json:"profile_id" db:"profile_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	FirstName   *string   `json:"first_name" db:"first_name"`
	LastName    *string   `json:"last_name" db:"last_name"`
	PhoneNumber *string   `json:"phone_number" db:"phone_number"`
	Avatar      *string   `json:"avatar" db:"avatar"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}