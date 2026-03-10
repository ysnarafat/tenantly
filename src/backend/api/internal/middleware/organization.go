package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to validate organization",
					"details": err.Error(),
				})
			}
			c.Abort()
			return
		}

		// Store organization in context for use in handlers
		c.Set("organization", org)
		c.Set("organization_id", orgID)

		c.Next()
	}
}

// RequireOrganization middleware ensures user has access to the requested organization
func RequireOrganization(orgRepo *repositories.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get organization ID from context (set by validation middleware)
		orgIDVal, exists := c.Get("organization_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization context not found"})
			c.Abort()
			return
		}

		orgID := orgIDVal.(int)

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

		// Get user's organization from context (should be set by auth middleware)
		userOrgID, exists := c.Get("organization_id")
		if !exists {
			// If organization_id not in token, cannot verify access
			c.JSON(http.StatusForbidden, gin.H{"error": "User organization context not found"})
			c.Abort()
			return
		}

		// Check if user belongs to this organization
		if userOrgID.(int) != orgID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this organization"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgAdmin middleware ensures user is an ORG_ADMIN of the organization
func RequireOrgAdmin(orgRepo *repositories.OrganizationRepository) gin.HandlerFunc {
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
