package models

import (
	"time"
)

type User struct {
	UserID       int        `json:"userId" db:"user_id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         string     `json:"-" db:"role"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	LastLogin    *time.Time `json:"lastLogin,omitempty" db:"last_login"`
}

type Profile struct {
	UserID      *int       `json:"userId" db:"user_id"`
	FirstName   *string    `json:"firstName" db:"first_name"`
	LastName    *string    `json:"lastName" db:"last_name"`
	Email       string     `json:"email" db:"email"`
	AvatarPath  *string    `json:"avatar_path" db:"avatar_path"`
	PhoneNumber *string    `json:"phoneNumber" db:"phone_number"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	LastLogin   *time.Time `json:"lastLogin" db:"last_login"`
}