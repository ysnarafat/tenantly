package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type BillingRunHandler struct {
	service interfaces.BillingRunServiceInterface
}

func NewBillingRunHandler(service interfaces.BillingRunServiceInterface) *BillingRunHandler {
	return &BillingRunHandler{service: service}
}

func (h *BillingRunHandler) Create(c *gin.Context) {
	var req models.CreateBillingRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "BILLING_RUN_INVALID_BODY", "Invalid request body", err)
		return
	}
	run, err := h.service.Create(&req, c.GetInt("org_id"), c.GetInt("user_id"))
	if err != nil {
		if err.Error() == "building not found" {
			respondError(c, http.StatusNotFound, "BILLING_BUILDING_NOT_FOUND", "Building not found", err)
			return
		}
		respondError(c, http.StatusConflict, "BILLING_RUN_CREATE_FAILED", "Could not create billing run", err)
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *BillingRunHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid billing run ID"})
		return
	}
	run, err := h.service.Get(id, c.GetInt("org_id"))
	if err != nil {
		respondError(c, http.StatusNotFound, "BILLING_RUN_NOT_FOUND", "Billing run not found", err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *BillingRunHandler) GetByPeriod(c *gin.Context) {
	buildingID, err1 := strconv.Atoi(c.Query("building_id"))
	month, err2 := strconv.Atoi(c.Query("month"))
	year, err3 := strconv.Atoi(c.Query("year"))
	if err1 != nil || err2 != nil || err3 != nil || buildingID < 1 || month < 1 || month > 12 || year < 2020 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "building_id, month, and year are required"})
		return
	}
	run, err := h.service.GetByPeriod(buildingID, c.GetInt("org_id"), month, year)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "BILLING_RUN_LOOKUP_FAILED", "Could not load billing run", err)
		return
	}
	if run == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "billing run not found"})
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *BillingRunHandler) UpdateElectricity(c *gin.Context) {
	runID, err1 := strconv.Atoi(c.Param("id"))
	itemID, err2 := strconv.Atoi(c.Param("itemId"))
	if err1 != nil || err2 != nil || runID < 1 || itemID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid billing run or item ID"})
		return
	}
	var req models.UpdateBillingRunItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "BILLING_ITEM_INVALID_BODY", "Invalid electricity amount", err)
		return
	}
	run, err := h.service.UpdateElectricity(runID, itemID, c.GetInt("org_id"), c.GetInt("user_id"), req.MeterNumber, req.ElectricityAmount)
	if err != nil {
		respondError(c, http.StatusConflict, "BILLING_ITEM_UPDATE_FAILED", "Could not update electricity amount", err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *BillingRunHandler) Finalize(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid billing run ID"})
		return
	}
	run, err := h.service.Finalize(id, c.GetInt("org_id"), c.GetInt("user_id"))
	if err != nil {
		respondError(c, http.StatusConflict, "BILLING_RUN_FINALIZE_FAILED", "Could not finalize billing run", err)
		return
	}
	c.JSON(http.StatusOK, run)
}
