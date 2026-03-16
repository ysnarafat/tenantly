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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
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
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create organization",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list organizations",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update organization",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete organization",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// InviteUser invites a user to an organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID, err := strconv.Atoi(c.Param("org_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req models.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
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
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to invite user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// GetPendingInvitations gets all pending invitations for an organization
func (h *OrganizationHandler) GetPendingInvitations(c *gin.Context) {
	orgID, err := strconv.Atoi(c.Param("org_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	invitations, err := h.organizationService.GetPendingInvitations(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get pending invitations",
			"details": err.Error(),
		})
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
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to revoke invitation",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "invitation has already been accepted" || err.Error() == "invitation has expired" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to accept invitation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation accepted successfully",
		"invitation": invitation,
	})
}
