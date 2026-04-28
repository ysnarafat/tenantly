package models

import "time"

type SubscriptionTier string

const (
	TierBasic        SubscriptionTier = "basic"
	TierProfessional SubscriptionTier = "professional"
	TierEnterprise   SubscriptionTier = "enterprise"
)

// Organization represents a top-level container for multi-tenancy
type Organization struct {
	ID               int              `json:"id" db:"id"`
	Name             string           `json:"name" db:"name"`
	Slug             string           `json:"slug" db:"slug"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier" db:"subscription_tier"`
	MaxUsers         int              `json:"max_users" db:"max_users"`
	Active           bool             `json:"active" db:"active"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateOrganizationRequest is the request payload for creating an organization
type CreateOrganizationRequest struct {
	Name             string           `json:"name" binding:"required,min=2,max=255"`
	Slug             string           `json:"slug" binding:"required,min=2,max=100"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier" binding:"omitempty,oneof=basic professional enterprise"`
	MaxUsers         int              `json:"max_users" binding:"omitempty,min=1"`
}

// UpdateOrganizationRequest is the request payload for updating an organization
type UpdateOrganizationRequest struct {
	Name             string           `json:"name" binding:"omitempty,min=2,max=255"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier" binding:"omitempty,oneof=basic professional enterprise"`
	MaxUsers         int              `json:"max_users" binding:"omitempty,min=1"`
	Active           *bool            `json:"active"`
}

// UserInvitation represents an invitation for a user to join an organization
type UserInvitation struct {
	ID               int        `json:"id" db:"id"`
	OrganizationID   int        `json:"organization_id" db:"organization_id"`
	Email            string     `json:"email" db:"email"`
	Role             string     `json:"role" db:"role"`
	InvitationToken  string     `json:"invitation_token" db:"invitation_token"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt       *time.Time `json:"accepted_at" db:"accepted_at"`
	AcceptedByUserID *int       `json:"accepted_by_user_id" db:"accepted_by_user_id"`
	InvitedByUserID  int        `json:"invited_by_user_id" db:"invited_by_user_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// InviteUserRequest is the request payload for inviting a user to an organization
type InviteUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=SUPER_ADMIN ORG_ADMIN Admin PropertyManager Accountant"`
}

// UserOrganizationRole represents a user's membership and role in an organization
type UserOrganizationRole struct {
	ID             int          `json:"id" db:"id"`
	UserID         int          `json:"user_id" db:"user_id"`
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	Organization   Organization `json:"organization"`
	Role           string       `json:"role" db:"role"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

// SetOrganizationRequest is the request payload for switching organizations
type SetOrganizationRequest struct {
	OrganizationID int `json:"organization_id" binding:"required"`
}

// SetOrganizationResponse is returned when a user switches organizations
type SetOrganizationResponse struct {
	Token        string               `json:"token"`
	RefreshToken string               `json:"refresh_token"`
	Organization UserOrganizationRole `json:"organization"`
	ExpiresAt    time.Time            `json:"expires_at"`
}
