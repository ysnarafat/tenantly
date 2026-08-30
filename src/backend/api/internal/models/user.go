package models

import "time"

type User struct {
	ID             int       `json:"id" db:"id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	Role           string    `json:"role" db:"role"`
	Active         bool      `json:"active" db:"active"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	OrganizationID *int      `json:"organization_id" db:"organization_id"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
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
	Username  string `json:"username" binding:"omitempty,min=3,max=50"`
	Email     string `json:"email" binding:"omitempty,email"`
	FirstName string `json:"first_name" binding:"omitempty,min=1,max=100"`
	LastName  string `json:"last_name" binding:"omitempty,min=1,max=100"`
	Role      string `json:"role" binding:"omitempty,oneof=SUPER_ADMIN ORG_ADMIN Admin PropertyManager Accountant"`
	Status    string `json:"status" binding:"omitempty,oneof=active inactive pending_invite"`
	Active    *bool  `json:"active"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	// RefreshToken is deliberately never serialized (json:"-") — it's set as an
	// httpOnly cookie by the handler instead of being exposed to JS-readable
	// response bodies, which would otherwise let XSS steal a long-lived
	// credential rather than just a session token.
	RefreshToken          string                 `json:"-"`
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

// AdminResetPasswordRequest is used by an admin to set a new password for
// another user directly (e.g. the user forgot theirs and can't self-serve
// via email reset). Unlike ChangePasswordRequest, it has no current-password
// field — the caller's own admin role is the authorization, checked by the
// handler/service, not proof of knowing the old password.
type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
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
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	FirstName       string `json:"first_name" binding:"required,min=1,max=100"`
	LastName        string `json:"last_name" binding:"required,min=1,max=100"`
	InvitationToken string `json:"invitation_token" binding:"required"`
	Username        string `json:"username" binding:"omitempty,min=3,max=50"`
}
