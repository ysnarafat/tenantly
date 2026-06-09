package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentHandler handles HTTP requests for payment operations
type PaymentHandler struct {
	paymentService interfaces.PaymentServiceInterface
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentService interfaces.PaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePayment handles POST /payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OrganizationID = c.GetInt("org_id")
	userID := c.GetInt("userID")

	payment, err := h.paymentService.CreatePayment(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log successful payment creation
	h.paymentService.LogPaymentAccess(userID, "CREATE", payment.ID, true)

	c.JSON(http.StatusCreated, payment)
}

// GetPayment handles GET /payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	userID := c.GetInt("userID")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	payment, err := h.paymentService.GetPayment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Verify user has access to this payment
	if !h.paymentService.CanUserAccessPayment(userID, userRole, payment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to access this payment"})
		h.paymentService.LogPaymentAccess(userID, "GET", id, false)
		return
	}

	// Log successful access
	h.paymentService.LogPaymentAccess(userID, "GET", id, true)

	c.JSON(http.StatusOK, payment)
}

// UpdatePayment handles PUT /payments/:id
func (h *PaymentHandler) UpdatePayment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	var req models.UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("userID")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	// Get existing payment for access verification
	existingPayment, err := h.paymentService.GetPayment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	// Verify user has access to update this payment
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to update this payment"})
		h.paymentService.LogPaymentAccess(userID, "UPDATE", id, false)
		return
	}

	// Additional: Accountants should not be able to update (read-only)
	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to payments"})
		h.paymentService.LogPaymentAccess(userID, "UPDATE", id, false)
		return
	}

	payment, err := h.paymentService.UpdatePayment(id, &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log successful update
	h.paymentService.LogPaymentAccess(userID, "UPDATE", id, true)

	c.JSON(http.StatusOK, payment)
}

// GetPayments handles GET /payments with query filters
// Query params: building_id, property_id, status, month, year, page, page_size
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	userID := c.GetInt("userID")
	orgID := c.GetInt("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := map[string]interface{}{
		"organization_id": orgID,
	}

	if buildingID := c.Query("building_id"); buildingID != "" {
		if id, err := strconv.Atoi(buildingID); err == nil {
			filters["building_id"] = id
		}
	}
	if propertyID := c.Query("property_id"); propertyID != "" {
		if id, err := strconv.Atoi(propertyID); err == nil {
			filters["property_id"] = id
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if month := c.Query("month"); month != "" {
		if m, err := strconv.Atoi(month); err == nil {
			filters["month"] = m
		}
	}
	if year := c.Query("year"); year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			filters["year"] = y
		}
	}

	var payments []*models.PaymentWithDetails
	var total int
	var err error

	if buildingID, ok := filters["building_id"].(int); ok {
		payments, total, err = h.paymentService.GetPaymentsByBuilding(buildingID, page, pageSize, filters)
	} else if propertyID, ok := filters["property_id"].(int); ok {
		payments, total, err = h.paymentService.GetPaymentsByProperty(propertyID, page, pageSize, filters)
	} else {
		payments, total, err = h.paymentService.GetPayments(page, pageSize, filters)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := 1
	if pageSize > 0 && total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	// Log payment list access with filters for audit trail
	h.paymentService.LogPaymentAccess(userID, "LIST", 0, true)

	c.JSON(http.StatusOK, models.PaymentListResponse{
		Payments:   payments,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetDashboardSummary handles GET /dashboard/summary
func (h *PaymentHandler) GetDashboardSummary(c *gin.Context) {
	summary, err := h.paymentService.GetDashboardSummaryWithBuildingContext()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetBuildingPaymentReport handles GET /payments/building/:building_id/report
// Query params: start_date, end_date (YYYY-MM-DD)
func (h *PaymentHandler) GetBuildingPaymentReport(c *gin.Context) {
	buildingID, err := strconv.Atoi(c.Param("building_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid building_id"})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.paymentService.GenerateBuildingPaymentReport(buildingID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPropertyPaymentReport handles GET /payments/property/:property_id/report
// Query params: start_date, end_date (YYYY-MM-DD)
func (h *PaymentHandler) GetPropertyPaymentReport(c *gin.Context) {
	propertyID, err := strconv.Atoi(c.Param("property_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid property_id"})
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.paymentService.GeneratePropertyPaymentReport(propertyID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// BulkCreatePayments handles POST /payments/bulk
func (h *PaymentHandler) BulkCreatePayments(c *gin.Context) {
	var requests []*models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetInt("org_id")
	userID := c.GetInt("userID")
	for _, req := range requests {
		req.OrganizationID = orgID
	}

	payments, errs := h.paymentService.ProcessBulkPayments(requests, userID)

	errMessages := make([]string, 0, len(errs))
	for _, e := range errs {
		errMessages = append(errMessages, e.Error())
	}

	// Log bulk operation
	h.paymentService.LogPaymentAccess(userID, "BULK_CREATE", 0, len(errs) == 0)

	c.JSON(http.StatusMultiStatus, gin.H{
		"created": payments,
		"errors":  errMessages,
		"total":   len(requests),
		"success": len(payments),
		"failed":  len(errs),
	})
}

// SearchLeases handles GET /payments/search for autocomplete and lease discovery
func (h *PaymentHandler) SearchLeases(c *gin.Context) {
	orgID := c.GetInt("org_id")
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query parameter 'q' is required"})
		return
	}

	result, err := h.paymentService.SearchLeases(orgID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GenerateMonthlyPayments handles POST /payments/generate-monthly
func (h *PaymentHandler) GenerateMonthlyPayments(c *gin.Context) {
	var req models.GenerateMonthlyPaymentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetInt("org_id")
	userID := c.GetInt("userID")

	result, err := h.paymentService.GenerateMonthlyPayments(&req, orgID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	now := time.Now()
	startStr := c.DefaultQuery("start_date", now.AddDate(0, -1, 0).Format("2006-01-02"))
	endStr := c.DefaultQuery("end_date", now.Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date; use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date; use YYYY-MM-DD")
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	return startDate, endDate, nil
}
