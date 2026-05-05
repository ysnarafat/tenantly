package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

type LeaseHandler struct {
	leaseRepo *repositories.LeaseRepository
}

func NewLeaseHandler(leaseRepo *repositories.LeaseRepository) *LeaseHandler {
	return &LeaseHandler{leaseRepo: leaseRepo}
}

func (h *LeaseHandler) GetActiveLeases(c *gin.Context) {
	orgID := c.GetInt("org_id")
	leases, err := h.leaseRepo.GetActiveLeases(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leases"})
		return
	}
	if leases == nil {
		leases = []*models.LeaseWithDetails{}
	}
	c.JSON(http.StatusOK, gin.H{"leases": leases})
}
