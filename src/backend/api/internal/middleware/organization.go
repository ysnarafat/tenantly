package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

// OrganizationValidationMiddleware creates middleware to validate organization existence
func OrganizationValidationMiddleware(orgRepo *repositories.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get organization ID from URL parameter
		orgIDStr := c.Param("org_id")
		if orgIDStr == "" {
			orgIDStr = c.Param("id")
		}
		if orgIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required"})
			c.Abort()
			return
		}

		orgID, err := strconv.Atoi(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
			c.Abort()
			return
		}

		// Validate organization exists
		org, err := orgRepo.GetByID(orgID)
		if err != nil {
			if err.Error() == "organization not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			} else {
				respondError(c, http.StatusInternalServerError, "ORG_VALIDATION_FAILED", "Failed to validate organization", err)
			}
			c.Abort()
			return
		}

		// Store organization in context for use in handlers. Stored under a
		// distinct key from the auth middleware's own "organization_id" (the
		// caller's org, from the JWT) — RequireOrganization below compares the
		// two; overwriting "organization_id" here would make that check
		// tautological (comparing a value to itself).
		c.Set("organization", org)
		c.Set("requested_organization_id", orgID)

		c.Next()
	}
}

// RequireOrganization middleware ensures the caller has access to the
// organization requested via the URL (set by OrganizationValidationMiddleware
// under "requested_organization_id"), by comparing it against the caller's own
// organization from the JWT ("organization_id", set by AuthRequired). Must run
// after both AuthRequired and OrganizationValidationMiddleware.
func RequireOrganization(orgRepo *repositories.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get the requested organization ID (set by OrganizationValidationMiddleware)
		requestedOrgIDVal, exists := c.Get("requested_organization_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization context not found"})
			c.Abort()
			return
		}
		requestedOrgID := requestedOrgIDVal.(int)

		// Get user's role from context (use "role" key set by auth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// SUPER_ADMIN can access any organization
		if userRole.(string) == "SUPER_ADMIN" {
			c.Next()
			return
		}

		// Get the caller's own organization from the JWT (set by AuthRequired as a *int)
		callerOrgIDVal, exists := c.Get("organization_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User organization context not found"})
			c.Abort()
			return
		}
		callerOrgID, ok := callerOrgIDVal.(*int)
		if !ok || callerOrgID == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "User organization context not found"})
			c.Abort()
			return
		}

		// Check if user belongs to the requested organization
		if *callerOrgID != requestedOrgID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this organization"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgAdmin middleware ensures the caller is an ORG_ADMIN of the specific
// organization referenced by the ":id"/":org_id" URL param, or a SUPER_ADMIN (who
// may act on any organization). A JWT role claim of ORG_ADMIN alone is not
// sufficient — that only proves the caller is an ORG_ADMIN of *some* organization,
// not this one, so membership is verified against user_organization_roles.
func RequireOrgAdmin(orgRepo *repositories.OrganizationRepository, userOrgRoleRepo interfaces.UserOrganizationRoleRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (use "role" key set by auth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Check if user is SUPER_ADMIN or ORG_ADMIN
		role := userRole.(string)
		if role != "SUPER_ADMIN" && role != "ORG_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. ORG_ADMIN required"})
			c.Abort()
			return
		}

		// SUPER_ADMIN may act on any organization
		if role == "SUPER_ADMIN" {
			c.Next()
			return
		}

		// ORG_ADMIN: verify membership in the specific organization targeted by this request
		orgIDStr := c.Param("org_id")
		if orgIDStr == "" {
			orgIDStr = c.Param("id")
		}
		orgID, err := strconv.Atoi(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			c.Abort()
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			c.Abort()
			return
		}
		userID, ok := userIDVal.(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		orgRole, err := userOrgRoleRepo.GetByUserAndOrg(userID, orgID)
		if err != nil || orgRole.Role != "ORG_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. ORG_ADMIN of this organization required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin middleware ensures user is a SUPER_ADMIN
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (use "role" key set by auth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Check if user is SUPER_ADMIN
		if userRole.(string) != "SUPER_ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions. SUPER_ADMIN required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
