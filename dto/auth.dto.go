package dto

import "time"

type RegisterRequest struct {
	Email           string `form:"email" binding:"required,email"`
	Password        string `form:"password" binding:"required,min=8"`
	ConfirmPassword string `form:"confirmPassword" binding:"required"`
}

type LoginRequest struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

type LogoutRequest struct {
	Token string `form:"token" binding:"required"`
}

type UserResponse struct {
	UserID    int        `json:"user_d"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

type AuthResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

type PasswordResetRequest struct {
	Email string `form:"email" json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	NewPassword string `form:"new_password" json:"new_password" binding:"required,min=6"`
}
