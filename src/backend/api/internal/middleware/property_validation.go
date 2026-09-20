package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

// PropertyValidationMiddleware creates middleware to validate property existence
func PropertyValidationMiddleware(propertyRepo *repositories.PropertyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get property ID from URL parameter
		propertyIDStr := c.Param("propertyId")
		if propertyIDStr == "" {
			propertyIDStr = c.Param("id")
		}
		if propertyIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Property ID is required"})
			c.Abort()
			return
		}

		propertyID, err := strconv.Atoi(propertyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID format"})
			c.Abort()
			return
		}

		// Validate property exists
		property, err := propertyRepo.GetByID(propertyID)
		if err != nil {
			if err.Error() == "property not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
			} else {
				respondError(c, http.StatusInternalServerError, "PROPERTY_VALIDATION_FAILED", "Failed to validate property", err)
			}
			c.Abort()
			return
		}

		// Check if property is active
		if !property.Active {
			c.JSON(http.StatusForbidden, gin.H{"error": "Property is not active"})
			c.Abort()
			return
		}

		// Store property in context for use in handlers
		c.Set("property", property)
		c.Set("property_id", propertyID)

		c.Next()
	}
}
