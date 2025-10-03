package models

import "time"

type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=Admin PropertyManager Accountant"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role" binding:"omitempty,oneof=Admin PropertyManager Accountant"`
	Active   *bool  `json:"active"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	User         User      `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ConfirmPasswordResetRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
