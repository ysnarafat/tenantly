package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// TenantHandler handles HTTP requests for tenant operations
type TenantHandler struct {
	tenantService interfaces.TenantServiceInterface
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(tenantService interfaces.TenantServiceInterface) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// CreateTenant handles tenant creation
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("userID")
	orgID := c.GetInt("org_id")
	req.OrganizationID = orgID
	tenant, err := h.tenantService.CreateTenant(&req, userID)
	if err != nil {
		// Check for uniqueness errors
		if err.Error() == "email already exists" || err.Error() == "NID number already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetAllTenants handles fetching all tenants
func (h *TenantHandler) GetAllTenants(c *gin.Context) {
	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.tenantService.GetAllTenants(page, pageSize, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTenantByID handles fetching a single tenant by ID
func (h *TenantHandler) GetTenantByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	orgID := c.GetInt("org_id")
	tenant, err := h.tenantService.GetTenantByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant handles tenant updates
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var req models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("userID")
	orgID := c.GetInt("org_id")

	tenant, err := h.tenantService.UpdateTenant(id, &req, userID, orgID)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "NID number already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant handles tenant deletion (soft delete with guard for active leases)
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userID := c.GetInt("userID")
	orgID := c.GetInt("org_id")

	if err := h.tenantService.DeleteTenant(id, userID, orgID); err != nil {
		if err.Error() == "cannot delete tenant with active leases" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}
