package handlers

import (
	"fmt"
	"io"
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
		respondError(c, http.StatusBadRequest, "CREATE_TENANT_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")
	req.OrganizationID = orgID

	tenant, err := h.tenantService.CreateTenant(&req, userID)
	if err != nil {
		// Check for uniqueness errors
		if err.Error() == "email already exists" || err.Error() == "NID number already exists" {
			respondError(c, http.StatusConflict, "CREATE_TENANT_CONFLICT", err.Error(), err)
			return
		}
		respondError(c, http.StatusBadRequest, "CREATE_TENANT_FAILED", "Failed to create tenant", err)
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
		respondError(c, http.StatusInternalServerError, "GET_ALL_TENANTS_FAILED", "Failed to retrieve tenants", err)
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

// RevealNID returns the full, decrypted NID for a tenant. This endpoint is
// role-gated (excludes read-only Accountants) and every access is audit-logged.
func (h *TenantHandler) RevealNID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	nid, err := h.tenantService.RevealNID(id, orgID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nid_number": nid})
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
		respondError(c, http.StatusBadRequest, "UPDATE_TENANT_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	tenant, err := h.tenantService.UpdateTenant(id, &req, userID, orgID)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "NID number already exists" {
			respondError(c, http.StatusConflict, "UPDATE_TENANT_CONFLICT", err.Error(), err)
			return
		}
		respondError(c, http.StatusBadRequest, "UPDATE_TENANT_FAILED", "Failed to update tenant", err)
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

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	if err := h.tenantService.DeleteTenant(id, userID, orgID); err != nil {
		if err.Error() == "cannot delete tenant with active leases" {
			respondError(c, http.StatusConflict, "DELETE_TENANT_CONFLICT", err.Error(), err)
			return
		}
		respondError(c, http.StatusBadRequest, "DELETE_TENANT_FAILED", "Failed to delete tenant", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// UploadTenantAttachment handles POST /tenants/:id/attachments
func (h *TenantHandler) UploadTenantAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingTenant, err := h.tenantService.GetTenantByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	if existingTenant.OrganizationID != orgID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to tenants"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_INVALID_BODY", "No file provided", err)
		return
	}

	// Open and read file data
	data, err := file.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "UPLOAD_ATTACHMENT_READ_FAILED", "Failed to open uploaded file", err)
		return
	}
	defer func() { _ = data.Close() }()

	dataBytes, err := io.ReadAll(data)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_READ_FAILED", "Failed to read uploaded file", err)
		return
	}

	// Handle attachment upload
	attachment, err := h.tenantService.UploadTenantAttachment(id, file.Filename, dataBytes, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_FAILED", "Failed to upload attachment", err)
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

// GetTenantAttachments lists all attachments for a tenant
func (h *TenantHandler) GetTenantAttachments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	orgID := c.GetInt("org_id")

	attachments, err := h.tenantService.GetTenantAttachments(id, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_TENANT_ATTACHMENTS_FAILED", "Failed to get tenant attachments", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"attachments": attachments})
}

// GetTenantAttachmentFile returns an attachment's raw bytes for download
func (h *TenantHandler) GetTenantAttachmentFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	orgID := c.GetInt("org_id")

	data, fileName, contentType, err := h.tenantService.GetTenantAttachmentFile(id, attachmentID, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_TENANT_ATTACHMENT_FAILED", "Attachment not found", err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Data(http.StatusOK, contentType, data)
}

// DeleteTenantAttachment removes a mistakenly-uploaded attachment
func (h *TenantHandler) DeleteTenantAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	orgID := c.GetInt("org_id")

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")

	// Verify tenant exists and belongs to org
	tenant, err := h.tenantService.GetTenantByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	if tenant.OrganizationID != orgID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to tenants"})
		return
	}

	if err := h.tenantService.DeleteTenantAttachment(id, attachmentID, userID, orgID); err != nil {
		respondError(c, http.StatusBadRequest, "DELETE_TENANT_ATTACHMENT_FAILED", "Failed to delete attachment", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted successfully"})
}
