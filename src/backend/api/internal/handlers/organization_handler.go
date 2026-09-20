package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/services"
)

type OrganizationHandler struct {
	organizationService *services.OrganizationService
}

func NewOrganizationHandler(organizationService *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		organizationService: organizationService,
	}
}

// CreateOrganization creates a new organization (SUPER_ADMIN only)
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req models.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	org, err := h.organizationService.CreateOrganization(&req, userID.(int))
	if err != nil {
		if err.Error() == "organization slug already exists" {
			respondError(c, http.StatusConflict, "ORGANIZATION_SLUG_EXISTS", "Organization slug already exists", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "CREATE_ORGANIZATION_FAILED", "Failed to create organization", err)
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetOrganization retrieves a single organization by ID
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.organizationService.GetOrganization(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations lists all organizations with optional filtering
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	orgs, err := h.organizationService.ListOrganizations(activeOnly)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LIST_ORGANIZATIONS_FAILED", "Failed to list organizations", err)
		return
	}

	if orgs == nil {
		orgs = make([]*models.Organization, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgs,
		"total":         len(orgs),
	})
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req models.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = h.organizationService.UpdateOrganization(id, &req, userID.(int))
	if err != nil {
		if err.Error() == "organization not found: organization not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		}
		respondError(c, http.StatusInternalServerError, "UPDATE_ORGANIZATION_FAILED", "Failed to update organization", err)
		return
	}

	// Return updated organization
	org, _ := h.organizationService.GetOrganization(id)
	c.JSON(http.StatusOK, org)
}

// DeleteOrganization soft deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = h.organizationService.DeleteOrganization(id, userID.(int))
	if err != nil {
		if err.Error() == "organization not found: organization not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		}
		respondError(c, http.StatusInternalServerError, "DELETE_ORGANIZATION_FAILED", "Failed to delete organization", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// InviteUser invites a user to an organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req models.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request data", err)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	invitation, err := h.organizationService.InviteUserToOrganization(orgID, &req, userID.(int))
	if err != nil {
		if err.Error() == "organization not found: organization not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		}
		if err.Error() == "cannot invite users to inactive organization" || err.Error() == "invitation already pending for this email" {
			respondError(c, http.StatusConflict, "INVITE_USER_CONFLICT", "Unable to invite user to this organization", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "INVITE_USER_FAILED", "Failed to invite user", err)
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// GetPendingInvitations gets all pending invitations for an organization
func (h *OrganizationHandler) GetPendingInvitations(c *gin.Context) {
	orgID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	invitations, err := h.organizationService.GetPendingInvitations(orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_PENDING_INVITATIONS_FAILED", "Failed to get pending invitations", err)
		return
	}

	if invitations == nil {
		invitations = make([]*models.UserInvitation, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       len(invitations),
	})
}

// RevokeInvitation revokes a pending invitation
func (h *OrganizationHandler) RevokeInvitation(c *gin.Context) {
	invitationID, err := strconv.Atoi(c.Param("invitation_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = h.organizationService.RevokeInvitation(invitationID, userID.(int))
	if err != nil {
		if err.Error() == "invitation not found: invitation not found" || err.Error() == "invitation not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}
		if err.Error() == "cannot revoke accepted invitation" {
			respondError(c, http.StatusConflict, "REVOKE_INVITATION_CONFLICT", "Cannot revoke this invitation", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "REVOKE_INVITATION_FAILED", "Failed to revoke invitation", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked successfully"})
}

// ValidateInvitationToken validates an invitation token (public endpoint)
func (h *OrganizationHandler) ValidateInvitationToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation token is required"})
		return
	}

	invitation, err := h.organizationService.ValidateInvitationToken(token)
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_INVITATION_TOKEN", "Invalid or expired invitation token", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitation": gin.H{
			"id":              invitation.ID,
			"organization_id": invitation.OrganizationID,
			"email":           invitation.Email,
			"role":            invitation.Role,
			"expires_at":      invitation.ExpiresAt,
			"token":           invitation.InvitationToken,
		},
	})
}

// AcceptInvitation accepts an invitation for an authenticated user
func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation token is required"})
			return
		}
		token = req.Token
	}

	// Get user ID from context (requires authentication)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	invitation, err := h.organizationService.AcceptInvitation(token, userID.(int))
	if err != nil {
		if err.Error() == "invalid or expired invitation token" {
			respondError(c, http.StatusBadRequest, "INVALID_INVITATION_TOKEN", "Invalid or expired invitation token", err)
			return
		}
		if err.Error() == "invitation has already been accepted" || err.Error() == "invitation has expired" {
			respondError(c, http.StatusConflict, "ACCEPT_INVITATION_CONFLICT", "Unable to accept this invitation", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "ACCEPT_INVITATION_FAILED", "Failed to accept invitation", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation accepted successfully",
		"invitation": invitation,
	})
}
