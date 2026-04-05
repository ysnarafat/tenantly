package models

import "time"

type User struct {
	ID             int        `json:"id" db:"id"`
	Username       string     `json:"username" db:"username"`
	Email          string     `json:"email" db:"email"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	Role           string     `json:"role" db:"role"`
	Active         bool       `json:"active" db:"active"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       string     `json:"last_name" db:"last_name"`
	OrganizationID *int       `json:"organization_id" db:"organization_id"`
	Status         string     `json:"status" db:"status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=50"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	FirstName      string `json:"first_name" binding:"required,min=1,max=100"`
	LastName       string `json:"last_name" binding:"required,min=1,max=100"`
	Role           string `json:"role" binding:"required,oneof=SUPER_ADMIN ORG_ADMIN Admin PropertyManager Accountant"`
	OrganizationID *int   `json:"organization_id" binding:"omitempty"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	FirstName string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=SUPER_ADMIN ORG_ADMIN Admin PropertyManager Accountant"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive pending_invite"`
	Active   *bool  `json:"active"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token                 string                 `json:"token"`
	RefreshToken          string                 `json:"refresh_token"`
	User                  User                   `json:"user"`
	Organizations         []UserOrganizationRole `json:"organizations"`
	DefaultOrganizationID int                    `json:"default_organization_id"`
	ExpiresAt             time.Time              `json:"expires_at"`
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

type RegisterWithInvitationRequest struct {
	Email              string `json:"email" binding:"required,email"`
	Password           string `json:"password" binding:"required,min=8"`
	FirstName          string `json:"first_name" binding:"required,min=1,max=100"`
	LastName           string `json:"last_name" binding:"required,min=1,max=100"`
	InvitationToken    string `json:"invitation_token" binding:"required"`
	Username           string `json:"username" binding:"omitempty,min=3,max=50"`
}
