package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type OrganizationService struct {
	orgRepo              interfaces.OrganizationRepositoryInterface
	userInvitationRepo   interfaces.UserInvitationRepositoryInterface
	auditService         interfaces.AuditServiceInterface
	invitationExpiryDays int
}

func NewOrganizationService(
	orgRepo interfaces.OrganizationRepositoryInterface,
	userInvitationRepo interfaces.UserInvitationRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:              orgRepo,
		userInvitationRepo:   userInvitationRepo,
		auditService:         auditService,
		invitationExpiryDays: 7,
	}
}

// CreateOrganization creates a new organization (SUPER_ADMIN only)
func (s *OrganizationService) CreateOrganization(req *models.CreateOrganizationRequest, userID int) (*models.Organization, error) {
	// Validate required fields
	if err := s.validateOrganizationRequest(req); err != nil {
		return nil, err
	}

	// Check slug uniqueness
	existing, err := s.orgRepo.GetBySlug(req.Slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("organization slug already exists")
	}

	// Create organization
	org := &models.Organization{
		Name:             req.Name,
		Slug:             req.Slug,
		SubscriptionTier: req.SubscriptionTier,
		MaxUsers:         req.MaxUsers,
		Active:           true,
	}

	if org.MaxUsers == 0 {
		org.MaxUsers = 50 // Default max users
	}

	err = s.orgRepo.Create(org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "CREATE", "organizations", &org.ID, nil, org)

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(id int) (*models.Organization, error) {
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return org, nil
}

// GetOrganizationBySlug retrieves an organization by slug
func (s *OrganizationService) GetOrganizationBySlug(slug string) (*models.Organization, error) {
	org, err := s.orgRepo.GetBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return org, nil
}

// ListOrganizations lists all organizations
func (s *OrganizationService) ListOrganizations(activeOnly bool) ([]*models.Organization, error) {
	orgs, err := s.orgRepo.GetAll(activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	return orgs, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(id int, req *models.UpdateOrganizationRequest, userID int) error {
	// Get current organization for audit
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.SubscriptionTier != "" {
		updates["subscription_tier"] = req.SubscriptionTier
	}
	if req.MaxUsers > 0 {
		updates["max_users"] = req.MaxUsers
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if len(updates) == 0 {
		return nil // No updates to apply
	}

	err = s.orgRepo.Update(id, updates)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "UPDATE", "organizations", &id, org, updates)

	return nil
}

// DeleteOrganization soft deletes an organization
func (s *OrganizationService) DeleteOrganization(id int, userID int) error {
	// Get organization for audit
	org, err := s.orgRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	err = s.orgRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "DELETE", "organizations", &id, org, nil)

	return nil
}

// InviteUserToOrganization sends an invitation to a user to join an organization
func (s *OrganizationService) InviteUserToOrganization(orgID int, req *models.InviteUserRequest, invitedByUserID int) (*models.UserInvitation, error) {
	// Validate organization exists
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	if !org.Active {
		return nil, fmt.Errorf("cannot invite users to inactive organization")
	}

	// Check if user already has pending invitation
	existing, err := s.userInvitationRepo.GetByEmailAndOrg(req.Email, orgID)
	if err == nil && existing != nil && existing.AcceptedAt == nil {
		return nil, fmt.Errorf("invitation already pending for this email")
	}

	// Generate invitation token
	token, err := s.generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Create invitation
	invitation := &models.UserInvitation{
		OrganizationID:  orgID,
		Email:           req.Email,
		Role:            req.Role,
		InvitationToken: token,
		ExpiresAt:       time.Now().AddDate(0, 0, s.invitationExpiryDays),
		InvitedByUserID: invitedByUserID,
	}

	err = s.userInvitationRepo.Create(invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(invitedByUserID, "INVITE_USER", "user_invitations", &invitation.ID, nil, invitation)

	return invitation, nil
}

// GetPendingInvitations gets all pending invitations for an organization
func (s *OrganizationService) GetPendingInvitations(orgID int) ([]*models.UserInvitation, error) {
	invitations, err := s.userInvitationRepo.GetPendingByOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}
	return invitations, nil
}

// RevokeInvitation revokes a pending invitation
func (s *OrganizationService) RevokeInvitation(invitationID int, revokedByUserID int) error {
	// Get invitation for audit
	invitation, err := s.userInvitationRepo.GetByID(invitationID)
	if err != nil {
		return fmt.Errorf("invitation not found: %w", err)
	}

	if invitation.AcceptedAt != nil {
		return fmt.Errorf("cannot revoke accepted invitation")
	}

	err = s.userInvitationRepo.Delete(invitationID)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(revokedByUserID, "REVOKE_INVITATION", "user_invitations", &invitationID, invitation, nil)

	return nil
}

// AcceptInvitation accepts an invitation using its token (public endpoint)
func (s *OrganizationService) AcceptInvitation(token string, userID int) (*models.UserInvitation, error) {
	// Get invitation by token
	invitation, err := s.userInvitationRepo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invitation token")
	}

	// Check if invitation is already accepted
	if invitation.AcceptedAt != nil {
		return nil, fmt.Errorf("invitation has already been accepted")
	}

	// Check if invitation is expired
	if time.Now().After(invitation.ExpiresAt) {
		return nil, fmt.Errorf("invitation has expired")
	}

	// Accept the invitation
	err = s.userInvitationRepo.AcceptInvitation(invitation.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	// Refresh the invitation to get updated data
	invitation, err = s.userInvitationRepo.GetByID(invitation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve accepted invitation: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "ACCEPT_INVITATION", "user_invitations", &invitation.ID, nil, invitation)

	return invitation, nil
}

// ValidateInvitationToken validates an invitation token without accepting it
func (s *OrganizationService) ValidateInvitationToken(token string) (*models.UserInvitation, error) {
	// Get invitation by token
	invitation, err := s.userInvitationRepo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invitation token")
	}

	// Check if invitation is already accepted
	if invitation.AcceptedAt != nil {
		return nil, fmt.Errorf("invitation has already been accepted")
	}

	// Check if invitation is expired
	if time.Now().After(invitation.ExpiresAt) {
		return nil, fmt.Errorf("invitation has expired")
	}

	return invitation, nil
}

// Helper methods

func (s *OrganizationService) validateOrganizationRequest(req *models.CreateOrganizationRequest) error {
	if req.Name == "" {
		return fmt.Errorf("organization name is required")
	}
	if req.Slug == "" {
		return fmt.Errorf("organization slug is required")
	}
	if len(req.Slug) < 2 || len(req.Slug) > 100 {
		return fmt.Errorf("slug must be between 2 and 100 characters")
	}
	return nil
}

func (s *OrganizationService) generateInvitationToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
