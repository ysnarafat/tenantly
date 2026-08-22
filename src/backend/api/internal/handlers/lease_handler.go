package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// LeaseHandler handles HTTP requests for lease operations
type LeaseHandler struct {
	leaseService interfaces.LeaseServiceInterface
}

// NewLeaseHandler creates a new LeaseHandler
func NewLeaseHandler(leaseService interfaces.LeaseServiceInterface) *LeaseHandler {
	return &LeaseHandler{leaseService: leaseService}
}

// CreateLease handles lease creation
func (h *LeaseHandler) CreateLease(c *gin.Context) {
	var req models.CreateLeaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "CREATE_LEASE_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")
	req.OrganizationID = orgID

	lease, err := h.leaseService.CreateLease(&req, userID)
	if err != nil {
		if err.Error() == "unit already has an active lease" {
			respondError(c, http.StatusConflict, "CREATE_LEASE_CONFLICT", err.Error(), err)
			return
		}
		respondError(c, http.StatusBadRequest, "CREATE_LEASE_FAILED", "Failed to create lease", err)
		return
	}

	c.JSON(http.StatusCreated, lease)
}

// GetLeaseByID handles fetching a single lease by ID
func (h *LeaseHandler) GetLeaseByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	orgID := c.GetInt("org_id")
	lease, err := h.leaseService.GetLeaseByID(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lease not found"})
		return
	}

	c.JSON(http.StatusOK, lease)
}

// GetAllLeases handles fetching all leases
func (h *LeaseHandler) GetAllLeases(c *gin.Context) {
	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.leaseService.GetAllLeases(page, pageSize, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_ALL_LEASES_FAILED", "Failed to retrieve leases", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateLease handles lease updates
func (h *LeaseHandler) UpdateLease(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	var req models.UpdateLeaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "UPDATE_LEASE_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	lease, err := h.leaseService.UpdateLease(id, &req, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPDATE_LEASE_FAILED", "Failed to update lease", err)
		return
	}

	c.JSON(http.StatusOK, lease)
}

// DeleteLease handles lease deletion
func (h *LeaseHandler) DeleteLease(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	if err := h.leaseService.DeleteLease(id, userID, orgID); err != nil {
		respondError(c, http.StatusBadRequest, "DELETE_LEASE_FAILED", "Failed to delete lease", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lease deleted successfully"})
}

// TerminateLease handles lease termination (soft delete with termination date)
func (h *LeaseHandler) TerminateLease(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	// Check for termination_date in query params or body
	terminationDateStr := c.Query("termination_date")
	if terminationDateStr == "" {
		var req struct {
			TerminationDate string `json:"termination_date" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err == nil {
			terminationDateStr = req.TerminationDate
		} else {
			// Default to current date if not provided
			terminationDateStr = "" // Let service handle it
		}
	}

	if err := h.leaseService.TerminateLease(id, userID, orgID, terminationDateStr); err != nil {
		respondError(c, http.StatusBadRequest, "TERMINATE_LEASE_FAILED", "Failed to terminate lease", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lease terminated successfully"})
}

// RenewLease handles starting a new lease term for a unit/tenant, closing out
// the lease being renewed rather than mutating it in place.
func (h *LeaseHandler) RenewLease(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	var req models.RenewLeaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "RENEW_LEASE_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	orgID := c.GetInt("org_id")

	lease, err := h.leaseService.RenewLease(id, &req, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "RENEW_LEASE_FAILED", "Failed to renew lease", err)
		return
	}

	c.JSON(http.StatusCreated, lease)
}

// GetLeasesByUnit handles fetching leases by unit ID
func (h *LeaseHandler) GetLeasesByUnit(c *gin.Context) {
	unitID, err := strconv.Atoi(c.Param("unit_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.leaseService.GetLeasesByUnit(unitID, page, pageSize, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_LEASES_BY_UNIT_FAILED", "Failed to retrieve leases", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLeasesByTenant handles fetching leases by tenant ID
func (h *LeaseHandler) GetLeasesByTenant(c *gin.Context) {
	tenantID, err := strconv.Atoi(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.leaseService.GetLeasesByTenant(tenantID, page, pageSize, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_LEASES_BY_TENANT_FAILED", "Failed to retrieve leases", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLeasesDue handles fetching all leases with unpaid rent for current month
func (h *LeaseHandler) GetLeasesDue(c *gin.Context) {
	orgID := c.GetInt("org_id")

	leasesDue, err := h.leaseService.GetLeasesDue(orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_LEASES_DUE_FAILED", "Failed to retrieve due leases", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"leases": leasesDue})
}

// GetDueSummary handles fetching summary statistics for unpaid rent
func (h *LeaseHandler) GetDueSummary(c *gin.Context) {
	orgID := c.GetInt("org_id")

	summary, err := h.leaseService.GetDueSummary(orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_DUE_SUMMARY_FAILED", "Failed to retrieve due summary", err)
		return
	}

	c.JSON(http.StatusOK, summary)
}
