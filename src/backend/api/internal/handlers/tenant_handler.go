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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.tenantService.GetAllTenants(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
