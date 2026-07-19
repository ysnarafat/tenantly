package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// UnitHandler handles HTTP requests for unit operations
type UnitHandler struct {
	unitService interfaces.UnitServiceInterface
}

// NewUnitHandler creates a new UnitHandler
func NewUnitHandler(unitService interfaces.UnitServiceInterface) *UnitHandler {
	return &UnitHandler{unitService: unitService}
}

// CreateUnit handles unit creation
func (h *UnitHandler) CreateUnit(c *gin.Context) {
	var req models.CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "CREATE_UNIT_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")
	unit, err := h.unitService.CreateUnit(&req, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "CREATE_UNIT_FAILED", "Failed to create unit", err)
		return
	}

	c.JSON(http.StatusCreated, unit)
}

// GetUnit handles retrieving a unit by ID
func (h *UnitHandler) GetUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	orgID := c.GetInt("org_id")
	unit, err := h.unitService.GetUnit(id, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_UNIT_FAILED", "Unit not found", err)
		return
	}

	c.JSON(http.StatusOK, unit)
}

// UpdateUnit handles updating a unit
func (h *UnitHandler) UpdateUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var req models.UpdateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "UPDATE_UNIT_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")
	unit, err := h.unitService.UpdateUnit(id, &req, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPDATE_UNIT_FAILED", "Failed to update unit", err)
		return
	}

	c.JSON(http.StatusOK, unit)
}

// DeleteUnit handles deleting a unit
func (h *UnitHandler) DeleteUnit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")
	if err := h.unitService.DeleteUnit(id, userID, orgID); err != nil {
		respondError(c, http.StatusBadRequest, "DELETE_UNIT_FAILED", "Failed to delete unit", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unit deleted successfully"})
}

// GetUnitsByBuilding handles retrieving units for a building
func (h *UnitHandler) GetUnitsByBuilding(c *gin.Context) {
	buildingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orgID := c.GetInt("org_id")

	units, total, err := h.unitService.GetUnitsByBuilding(buildingID, page, pageSize, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "GET_UNITS_BY_BUILDING_FAILED", "Failed to retrieve units", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": units,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUnitsByProperty handles retrieving units for a property
func (h *UnitHandler) GetUnitsByProperty(c *gin.Context) {
	propertyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid property ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orgID := c.GetInt("org_id")

	units, total, err := h.unitService.GetUnitsByProperty(propertyID, page, pageSize, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "GET_UNITS_BY_PROPERTY_FAILED", "Failed to retrieve units", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": units,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetUnitHierarchyContext handles retrieving hierarchy context for a unit
func (h *UnitHandler) GetUnitHierarchyContext(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	orgID := c.GetInt("org_id")
	context, err := h.unitService.GetUnitHierarchyContext(id, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_UNIT_HIERARCHY_FAILED", "Unit not found", err)
		return
	}

	c.JSON(http.StatusOK, context)
}
