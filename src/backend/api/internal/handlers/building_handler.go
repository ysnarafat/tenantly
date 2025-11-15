package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type BuildingHandler struct {
	buildingService interfaces.BuildingServiceInterface
}

func NewBuildingHandler(buildingService interfaces.BuildingServiceInterface) *BuildingHandler {
	return &BuildingHandler{
		buildingService: buildingService,
	}
}

// CreateBuilding creates a new building
func (h *BuildingHandler) CreateBuilding(c *gin.Context) {
	var req models.CreateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	building, err := h.buildingService.CreateBuilding(&req)
	if err != nil {
		// Handle specific error types
		switch {
		case err.Error() == "property validation failed: property not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
			return
		case err.Error() == fmt.Sprintf("building code '%s' already exists in this property", req.BuildingCode):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		case err.Error() == "metadata validation failed":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata for building type"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create building",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Building created successfully",
		"building": building,
	})
}

// GetBuildings retrieves buildings with filtering by property, type, and status
func (h *BuildingHandler) GetBuildings(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Build search filters
	filters := &models.BuildingSearchFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Parse property_id filter
	if propertyIDStr := c.Query("property_id"); propertyIDStr != "" {
		if propertyID, err := strconv.Atoi(propertyIDStr); err == nil {
			filters.PropertyID = &propertyID
		}
	}

	// Parse building_type filter
	if buildingTypeStr := c.Query("building_type"); buildingTypeStr != "" {
		buildingType := models.BuildingType(buildingTypeStr)
		// Validate building type
		if buildingType == models.BuildingTypeResidential ||
			buildingType == models.BuildingTypeCommercial ||
			buildingType == models.BuildingTypeMixed {
			filters.BuildingType = &buildingType
		}
	}

	// Parse active_status filter
	if activeStr := c.Query("active_status"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.ActiveStatus = &active
		}
	}

	// Parse has_elevator filter
	if elevatorStr := c.Query("has_elevator"); elevatorStr != "" {
		if hasElevator, err := strconv.ParseBool(elevatorStr); err == nil {
			filters.HasElevator = &hasElevator
		}
	}

	// Parse floor range filters
	if minFloorsStr := c.Query("min_floors"); minFloorsStr != "" {
		if minFloors, err := strconv.Atoi(minFloorsStr); err == nil && minFloors > 0 {
			filters.MinFloors = &minFloors
		}
	}

	if maxFloorsStr := c.Query("max_floors"); maxFloorsStr != "" {
		if maxFloors, err := strconv.Atoi(maxFloorsStr); err == nil && maxFloors > 0 {
			filters.MaxFloors = &maxFloors
		}
	}

	buildings, err := h.buildingService.SearchBuildings(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve buildings",
			"details": err.Error(),
		})
		return
	}

	// Calculate pagination info (simplified - in production you'd want total count)
	hasNext := len(buildings) == pageSize
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"buildings": buildings,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}

// GetBuilding retrieves a single building by ID with building details, unit counts, and property information
func (h *BuildingHandler) GetBuilding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	// Check if stats are requested
	includeStats := c.Query("include_stats") == "true"

	if includeStats {
		building, err := h.buildingService.GetBuildingWithStats(id)
		if err != nil {
			if err.Error() == "failed to get building with stats: building not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve building with stats",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"building": building})
	} else {
		building, err := h.buildingService.GetBuilding(id)
		if err != nil {
			if err.Error() == "failed to get building: building not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to retrieve building",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"building": building})
	}
}

// UpdateBuilding updates a building for building updates and metadata modifications
func (h *BuildingHandler) UpdateBuilding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	var req models.UpdateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	building, err := h.buildingService.UpdateBuilding(id, &req)
	if err != nil {
		// Handle specific error types
		switch {
		case err.Error() == "failed to get existing building: building not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		case err.Error() == "metadata validation failed":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metadata for building type"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update building",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Building updated successfully",
		"building": building,
	})
}

// DeleteBuilding soft deletes a building with constraint validation
func (h *BuildingHandler) DeleteBuilding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	err = h.buildingService.DeleteBuilding(id)
	if err != nil {
		// Handle specific error types
		switch {
		case err.Error() == "failed to get existing building: building not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		case err.Error() == "failed to delete building: cannot delete building with active units":
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete building with active units"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete building",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Building deleted successfully"})
}

// GetPropertyBuildings retrieves all buildings for a specific property with enhanced filtering, pagination, and sorting
func (h *BuildingHandler) GetPropertyBuildings(c *gin.Context) {
	// Property validation is handled by middleware, get property from context
	property, exists := c.Get("property")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Property context not found"})
		return
	}
	propertyObj := property.(*models.Property)

	// Parse request parameters
	var req models.BuildingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build search filters with pagination and sorting
	filters := &models.BuildingSearchFilters{
		PropertyID: &propertyObj.ID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
	}

	// Apply additional filters
	if req.BuildingType != "" {
		buildingType := models.BuildingType(req.BuildingType)
		filters.BuildingType = &buildingType
	}
	if req.ActiveStatus != nil {
		filters.ActiveStatus = req.ActiveStatus
	}

	// Get buildings with enhanced service method
	response, err := h.buildingService.GetPropertyBuildingsWithPagination(propertyObj.ID, filters, req.SortBy, req.SortOrder, req.IncludeStats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve property buildings",
			"details": err.Error(),
		})
		return
	}

	// Create hierarchical response
	hierarchicalResponse := gin.H{
		"property": gin.H{
			"id":            propertyObj.ID,
			"property_name": propertyObj.PropertyName,
			"property_code": propertyObj.PropertyCode,
			"property_type": propertyObj.PropertyType,
			"address":       propertyObj.Address,
			"city":          propertyObj.City,
		},
		"buildings":  response.Buildings,
		"pagination": response.Pagination,
		"statistics": response.Statistics,
	}

	c.JSON(http.StatusOK, hierarchicalResponse)
}

// BulkCreateBuildings creates multiple buildings for a property with enhanced response structure
func (h *BuildingHandler) BulkCreateBuildings(c *gin.Context) {
	// Property validation is handled by middleware, get property from context
	property, exists := c.Get("property")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Property context not found"})
		return
	}
	propertyObj := property.(*models.Property)

	var req models.BulkCreateBuildingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Set property ID from context
	req.PropertyID = propertyObj.ID

	buildings, err := h.buildingService.BulkCreateBuildings(&req)
	if err != nil {
		// Handle specific error types
		switch {
		case err.Error() == "duplicate building code in request":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate building codes in request"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to bulk create buildings",
				"details": err.Error(),
			})
			return
		}
	}

	// Get updated property statistics after bulk creation
	propertyStats, err := h.buildingService.GetPropertyStatistics(propertyObj.ID)
	if err != nil {
		// Log error but don't fail the response
		propertyStats = nil
	}

	// Create hierarchical response
	response := gin.H{
		"message": "Buildings created successfully",
		"property": gin.H{
			"id":            propertyObj.ID,
			"property_name": propertyObj.PropertyName,
			"property_code": propertyObj.PropertyCode,
			"property_type": propertyObj.PropertyType,
		},
		"buildings": buildings,
		"summary": gin.H{
			"buildings_created": len(buildings),
			"total_buildings":   len(buildings), // This would be updated if we had existing count
		},
	}

	if propertyStats != nil {
		response["statistics"] = propertyStats
	}

	c.JSON(http.StatusCreated, response)
}

// SearchBuildings performs advanced building search
func (h *BuildingHandler) SearchBuildings(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Build search filters
	filters := &models.BuildingSearchFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Parse all possible filters
	if propertyIDStr := c.Query("property_id"); propertyIDStr != "" {
		if propertyID, err := strconv.Atoi(propertyIDStr); err == nil {
			filters.PropertyID = &propertyID
		}
	}

	if buildingTypeStr := c.Query("building_type"); buildingTypeStr != "" {
		buildingType := models.BuildingType(buildingTypeStr)
		if buildingType == models.BuildingTypeResidential ||
			buildingType == models.BuildingTypeCommercial ||
			buildingType == models.BuildingTypeMixed {
			filters.BuildingType = &buildingType
		}
	}

	if activeStr := c.Query("active_status"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters.ActiveStatus = &active
		}
	}

	if elevatorStr := c.Query("has_elevator"); elevatorStr != "" {
		if hasElevator, err := strconv.ParseBool(elevatorStr); err == nil {
			filters.HasElevator = &hasElevator
		}
	}

	if minFloorsStr := c.Query("min_floors"); minFloorsStr != "" {
		if minFloors, err := strconv.Atoi(minFloorsStr); err == nil && minFloors > 0 {
			filters.MinFloors = &minFloors
		}
	}

	if maxFloorsStr := c.Query("max_floors"); maxFloorsStr != "" {
		if maxFloors, err := strconv.Atoi(maxFloorsStr); err == nil && maxFloors > 0 {
			filters.MaxFloors = &maxFloors
		}
	}

	buildings, err := h.buildingService.SearchBuildings(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search buildings",
			"details": err.Error(),
		})
		return
	}

	// Calculate pagination info
	hasNext := len(buildings) == pageSize
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"buildings": buildings,
		"filters":   filters,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}

// GetBuildingAnalytics retrieves building performance metrics
func (h *BuildingHandler) GetBuildingAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	analytics, err := h.buildingService.GetBuildingAnalytics(id)
	if err != nil {
		if err.Error() == "building validation failed: building not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve building analytics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// AdvancedSearchBuildings performs advanced building search with metadata queries and enhanced filtering
func (h *BuildingHandler) AdvancedSearchBuildings(c *gin.Context) {
	var req models.BuildingSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid search parameters",
			"details": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	response, err := h.buildingService.AdvancedSearchBuildings(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search buildings",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"buildings":  response.Buildings,
		"pagination": response.Pagination,
		"filters":    req,
	})
}

// GetBuildingUnits retrieves units for a specific building with pagination
func (h *BuildingHandler) GetBuildingUnits(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.buildingService.GetBuildingUnits(id, page, pageSize)
	if err != nil {
		if strings.Contains(err.Error(), "building validation failed") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve building units",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBuildingMetadataSchema returns metadata schema information for building types
func (h *BuildingHandler) GetBuildingMetadataSchema(c *gin.Context) {
	buildingTypeStr := c.Param("type")
	buildingType := models.BuildingType(buildingTypeStr)

	response, err := h.buildingService.GetBuildingMetadataSchema(buildingType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid building type",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ExportBuildingData exports building data in various formats for reporting and analysis
func (h *BuildingHandler) ExportBuildingData(c *gin.Context) {
	var req models.BuildingExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid export parameters",
			"details": err.Error(),
		})
		return
	}

	data, contentType, err := h.buildingService.ExportBuildingData(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to export building data",
			"details": err.Error(),
		})
		return
	}

	// Set appropriate headers for file download
	filename := "buildings_export"
	switch req.Format {
	case "csv":
		filename += ".csv"
	case "json":
		filename += ".json"
	case "xlsx":
		filename += ".xlsx"
	default:
		filename += ".json"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

// UpdateBuildingStatus updates building activation/deactivation status
func (h *BuildingHandler) UpdateBuildingStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	var req models.BuildingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	building, err := h.buildingService.UpdateBuildingStatus(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "building not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
			return
		}
		if strings.Contains(err.Error(), "cannot deactivate building") {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Cannot deactivate building with active units",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update building status",
			"details": err.Error(),
		})
		return
	}

	action := "activated"
	if !req.ActiveStatus {
		action = "deactivated"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Building %s successfully", action),
		"building": building,
	})
}
