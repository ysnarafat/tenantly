package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/services"
)

type PropertyHandler struct {
	propertyService *services.PropertyService
}

func NewPropertyHandler(propertyService *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{
		propertyService: propertyService,
	}
}

// CreateProperty creates a new property
func (h *PropertyHandler) CreateProperty(c *gin.Context) {
	var req models.CreatePropertyRequest
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

	orgID := c.GetInt("org_id")
	req.OrganizationID = orgID

	property, err := h.propertyService.CreateProperty(&req, userID.(int))
	if err != nil {
		if err.Error() == "property name already exists" || err.Error() == "property code already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, property)
}

// GetProperties retrieves properties with filtering and pagination
func (h *PropertyHandler) GetProperties(c *gin.Context) {
	orgID := c.GetInt("org_id")

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	propertyType := c.Query("property_type")
	activeStr := c.Query("active")
	search := c.Query("search")

	// Build filters
	filters := make(map[string]interface{})
	filters["organization_id"] = orgID

	if propertyType != "" {
		filters["property_type"] = propertyType
	}

	if activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			filters["active"] = active
		}
	}

	if search != "" {
		filters["search"] = search
	}

	properties, total, err := h.propertyService.ListProperties(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve properties",
			"details": err.Error(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"properties": properties,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}

// GetProperty retrieves a single property by ID with enhanced building information
func (h *PropertyHandler) GetProperty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	// Check if stats are requested
	includeStats := c.Query("include_stats") == "true"
	includeBuildings := c.Query("include_buildings") == "true"

	if includeStats {
		property, err := h.propertyService.GetPropertyWithStats(id)
		if err != nil {
			if err.Error() == "property not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve property",
				"details": err.Error(),
			})
			return
		}

		response := gin.H{"property": property}

		// Add building information if requested
		if includeBuildings {
			buildingStats, err := h.propertyService.GetPropertyBuildingSummary(id)
			if err == nil {
				response["building_summary"] = buildingStats
			}
		}

		c.JSON(http.StatusOK, response)
	} else {
		property, err := h.propertyService.GetProperty(id)
		if err != nil {
			if err.Error() == "property not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve property",
				"details": err.Error(),
			})
			return
		}

		response := gin.H{"property": property}

		// Add basic building count if requested
		if includeBuildings {
			buildingCount, err := h.propertyService.GetPropertyBuildingCount(id)
			if err == nil {
				response["building_count"] = buildingCount
			}
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateProperty updates a property
func (h *PropertyHandler) UpdateProperty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	var req models.UpdatePropertyRequest
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

	property, err := h.propertyService.UpdateProperty(id, &req, userID.(int))
	if err != nil {
		if err.Error() == "property not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
			return
		}
		if err.Error() == "property name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, property)
}

// DeleteProperty soft deletes a property
func (h *PropertyHandler) DeleteProperty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err = h.propertyService.DeleteProperty(id, userID.(int))
	if err != nil {
		if err.Error() == "property not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
			return
		}
		if err.Error() == "cannot delete property: has active buildings" ||
			err.Error() == "cannot delete property: has active units" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete property",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property deleted successfully"})
}

// GetPropertyAggregations returns property-level aggregations
func (h *PropertyHandler) GetPropertyAggregations(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	aggregations, err := h.propertyService.GetPropertyAggregations(id)
	if err != nil {
		if err.Error() == "property not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve property aggregations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"aggregations": aggregations})
}

// SearchProperties performs advanced search on properties
func (h *PropertyHandler) SearchProperties(c *gin.Context) {
	searchTerm := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	propertyType := c.Query("property_type")
	activeStr := c.Query("active")

	// Build filters
	filters := make(map[string]interface{})

	if propertyType != "" {
		filters["property_type"] = propertyType
	}

	if activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			filters["active"] = active
		}
	}

	properties, total, err := h.propertyService.SearchProperties(searchTerm, filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search properties",
			"details": err.Error(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"properties":  properties,
		"search_term": searchTerm,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  totalPages,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}
