package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// maxAttachmentUploadBody bounds the raw request body accepted for an
// attachment upload — the attachment cap itself plus headroom for multipart
// framing overhead — so an oversized upload is rejected while streaming
// rather than after being fully buffered into memory.
const maxAttachmentUploadBody = models.MaxAttachmentFileSize + 1<<20

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
		respondError(c, http.StatusBadRequest, "CREATE_PAYMENT_INVALID_BODY", "Invalid request body", err)
		return
	}

	req.OrganizationID = c.GetInt("org_id")
	userID := c.GetInt("user_id")

	payment, err := h.paymentService.CreatePayment(&req, userID)
	if err != nil {
		if err.Error() == "a payment already exists for this unit for the selected month/year" {
			respondError(c, http.StatusConflict, "PAYMENT_ALREADY_EXISTS", "A payment already exists for this unit for the selected month/year", err)
			return
		}
		respondError(c, http.StatusBadRequest, "CREATE_PAYMENT_FAILED", "Failed to create payment", err)
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

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	payment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_PAYMENT_FAILED", "Payment not found", err)
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

// DownloadReceipt handles GET /payments/:id/receipt — streams a PDF receipt.
func (h *PaymentHandler) DownloadReceipt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	payment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_PAYMENT_FAILED", "Payment not found", err)
		return
	}

	if !h.paymentService.CanUserAccessPayment(userID, userRole, payment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to access this payment"})
		h.paymentService.LogPaymentAccess(userID, "DOWNLOAD_RECEIPT", id, false)
		return
	}

	pdfBytes, err := h.paymentService.GenerateReceiptPDF(id, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GENERATE_RECEIPT_FAILED", "Failed to generate receipt", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "DOWNLOAD_RECEIPT", id, true)

	filename := fmt.Sprintf("receipt-%s.pdf", payment.ReceiptNumber)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// DownloadReceiptByToken handles GET /receipts/:token — a public,
// unauthenticated download used by the "payment recorded" SMS link, since
// tenants (its recipients) have no login. The unguessable, expiring token is
// the only authorization.
func (h *PaymentHandler) DownloadReceiptByToken(c *gin.Context) {
	token := c.Param("token")

	pdfBytes, filename, err := h.paymentService.DownloadReceiptByToken(token)
	if err != nil {
		respondError(c, http.StatusNotFound, "RECEIPT_LINK_INVALID", "This receipt link is invalid or has expired", err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
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
		respondError(c, http.StatusBadRequest, "UPDATE_PAYMENT_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	// Get existing payment for access verification
	existingPayment, err := h.paymentService.GetPayment(id, orgID)
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

	payment, err := h.paymentService.UpdatePayment(id, &req, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPDATE_PAYMENT_FAILED", "Failed to update payment", err)
		return
	}

	// Log successful update
	h.paymentService.LogPaymentAccess(userID, "UPDATE", id, true)

	c.JSON(http.StatusOK, payment)
}

// RecordPaymentTransaction handles POST /payments/:id/transactions — the
// only way amount_paid ever changes. Adds to the existing total rather than
// replacing it, so a second installment against the same month's due amount
// doesn't overwrite the first.
func (h *PaymentHandler) RecordPaymentTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	var req models.CreatePaymentTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "RECORD_PAYMENT_TRANSACTION_INVALID_BODY", "Invalid request body", err)
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to update this payment"})
		h.paymentService.LogPaymentAccess(userID, "RECORD_PAYMENT", id, false)
		return
	}
	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to payments"})
		h.paymentService.LogPaymentAccess(userID, "RECORD_PAYMENT", id, false)
		return
	}

	payment, err := h.paymentService.RecordPaymentTransaction(id, &req, userID, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "exceeds the remaining due balance") {
			respondError(c, http.StatusConflict, "RECORD_PAYMENT_TRANSACTION_EXCEEDS_DUE", err.Error(), err)
			return
		}
		respondError(c, http.StatusBadRequest, "RECORD_PAYMENT_TRANSACTION_FAILED", "Failed to record payment", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "RECORD_PAYMENT", id, true)
	c.JSON(http.StatusCreated, payment)
}

// GetPaymentTransactions handles GET /payments/:id/transactions — the ledger
// of amounts received against a payment.
func (h *PaymentHandler) GetPaymentTransactions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to view this payment"})
		return
	}

	txns, err := h.paymentService.GetPaymentTransactions(id, orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_PAYMENT_TRANSACTIONS_FAILED", "Failed to get payment transactions", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}

// DeletePaymentTransaction handles DELETE /payments/:id/transactions/:transactionId
// — removes a mistakenly-recorded transaction and recomputes the payment's
// cached amount_paid/status from what remains.
func (h *PaymentHandler) DeletePaymentTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}
	transactionID, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to update this payment"})
		h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_TRANSACTION", id, false)
		return
	}
	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to payments"})
		h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_TRANSACTION", id, false)
		return
	}

	payment, err := h.paymentService.DeletePaymentTransaction(id, transactionID, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "DELETE_PAYMENT_TRANSACTION_FAILED", "Failed to delete payment transaction", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_TRANSACTION", id, true)
	c.JSON(http.StatusOK, payment)
}

// UploadPaymentTransactionAttachment handles POST
// /payments/:id/transactions/:transactionId/attachments — attaches a
// receipt photo or similar evidence file to one installment. Limited to
// 10MB; only images and PDF are accepted, checked against the actual file
// bytes rather than the client-supplied filename or Content-Type header.
func (h *PaymentHandler) UploadPaymentTransactionAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}
	transactionID, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to update this payment"})
		h.paymentService.LogPaymentAccess(userID, "UPLOAD_PAYMENT_ATTACHMENT", id, false)
		return
	}
	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to payments"})
		h.paymentService.LogPaymentAccess(userID, "UPLOAD_PAYMENT_ATTACHMENT", id, false)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachmentUploadBody)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_INVALID_BODY", "No file provided, or the file exceeds the 10MB limit", err)
		return
	}
	if fileHeader.Size > models.MaxAttachmentFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file exceeds the maximum allowed size of 10MB"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "UPLOAD_ATTACHMENT_OPEN_FAILED", "Failed to read uploaded file", err)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_READ_FAILED", "Failed to read uploaded file, or it exceeds the 10MB limit", err)
		return
	}

	attachment, err := h.paymentService.UploadPaymentTransactionAttachment(id, transactionID, fileHeader.Filename, data, userID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "UPLOAD_ATTACHMENT_FAILED", "Failed to upload attachment", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "UPLOAD_PAYMENT_ATTACHMENT", id, true)
	c.JSON(http.StatusCreated, attachment)
}

// GetPaymentTransactionAttachments handles GET
// /payments/:id/transactions/:transactionId/attachments
func (h *PaymentHandler) GetPaymentTransactionAttachments(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}
	transactionID, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to view this payment"})
		return
	}

	attachments, err := h.paymentService.GetPaymentTransactionAttachments(id, transactionID, orgID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "GET_PAYMENT_ATTACHMENTS_FAILED", "Failed to get payment attachments", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"attachments": attachments})
}

// DownloadPaymentTransactionAttachment handles GET
// /payments/:id/transactions/:transactionId/attachments/:attachmentId —
// streams the stored file bytes.
func (h *PaymentHandler) DownloadPaymentTransactionAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}
	transactionID, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction ID"})
		return
	}
	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to view this payment"})
		h.paymentService.LogPaymentAccess(userID, "DOWNLOAD_PAYMENT_ATTACHMENT", id, false)
		return
	}

	data, fileName, contentType, err := h.paymentService.GetPaymentTransactionAttachmentFile(id, transactionID, attachmentID, orgID)
	if err != nil {
		respondError(c, http.StatusNotFound, "GET_PAYMENT_ATTACHMENT_FAILED", "Attachment not found", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "DOWNLOAD_PAYMENT_ATTACHMENT", id, true)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	c.Data(http.StatusOK, contentType, data)
}

// DeletePaymentTransactionAttachment handles DELETE
// /payments/:id/transactions/:transactionId/attachments/:attachmentId
func (h *PaymentHandler) DeletePaymentTransactionAttachment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}
	transactionID, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment transaction ID"})
		return
	}
	attachmentID, err := strconv.Atoi(c.Param("attachmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attachment ID"})
		return
	}

	userID := c.GetInt("user_id")
	userRole := c.GetString("role")
	orgID := c.GetInt("org_id")

	existingPayment, err := h.paymentService.GetPayment(id, orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if !h.paymentService.CanUserAccessPayment(userID, userRole, existingPayment, orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to update this payment"})
		h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_ATTACHMENT", id, false)
		return
	}
	if userRole == "Accountant" {
		c.JSON(http.StatusForbidden, gin.H{"error": "accountants have read-only access to payments"})
		h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_ATTACHMENT", id, false)
		return
	}

	if err := h.paymentService.DeletePaymentTransactionAttachment(id, transactionID, attachmentID, userID, orgID); err != nil {
		respondError(c, http.StatusBadRequest, "DELETE_PAYMENT_ATTACHMENT_FAILED", "Failed to delete attachment", err)
		return
	}

	h.paymentService.LogPaymentAccess(userID, "DELETE_PAYMENT_ATTACHMENT", id, true)
	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted successfully"})
}

// GetPayments handles GET /payments with query filters
// Query params: building_id, property_id, status, month, year, page, page_size
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	userID := c.GetInt("user_id")
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
		respondError(c, http.StatusInternalServerError, "GET_PAYMENTS_FAILED", "Failed to retrieve payments", err)
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
	orgID := c.GetInt("org_id")
	summary, err := h.paymentService.GetDashboardSummaryWithBuildingContext(orgID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_DASHBOARD_SUMMARY_FAILED", "Failed to retrieve dashboard summary", err)
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
		respondError(c, http.StatusBadRequest, "GET_BUILDING_PAYMENT_REPORT_INVALID_RANGE", "Invalid date range", err)
		return
	}

	orgID := c.GetInt("org_id")
	report, err := h.paymentService.GenerateBuildingPaymentReport(buildingID, orgID, startDate, endDate)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_BUILDING_PAYMENT_REPORT_FAILED", "Failed to generate building payment report", err)
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
		respondError(c, http.StatusBadRequest, "GET_PROPERTY_PAYMENT_REPORT_INVALID_RANGE", "Invalid date range", err)
		return
	}

	orgID := c.GetInt("org_id")
	report, err := h.paymentService.GeneratePropertyPaymentReport(propertyID, orgID, startDate, endDate)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_PROPERTY_PAYMENT_REPORT_FAILED", "Failed to generate property payment report", err)
		return
	}

	c.JSON(http.StatusOK, report)
}

// BulkCreatePayments handles POST /payments/bulk
func (h *PaymentHandler) BulkCreatePayments(c *gin.Context) {
	var requests []*models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		respondError(c, http.StatusBadRequest, "BULK_CREATE_PAYMENTS_INVALID_BODY", "Invalid request body", err)
		return
	}

	orgID := c.GetInt("org_id")
	userID := c.GetInt("user_id")
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
		respondError(c, http.StatusInternalServerError, "SEARCH_LEASES_FAILED", "Failed to search leases", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GenerateMonthlyPayments handles POST /payments/generate-monthly
func (h *PaymentHandler) GenerateMonthlyPayments(c *gin.Context) {
	var req models.GenerateMonthlyPaymentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "GENERATE_MONTHLY_PAYMENTS_INVALID_BODY", "Invalid request body", err)
		return
	}

	orgID := c.GetInt("org_id")
	userID := c.GetInt("user_id")

	result, err := h.paymentService.GenerateMonthlyPayments(&req, orgID, userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GENERATE_MONTHLY_PAYMENTS_FAILED", "Failed to generate monthly payments", err)
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
