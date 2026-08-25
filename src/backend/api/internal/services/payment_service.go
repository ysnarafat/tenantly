package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type PaymentService struct {
	paymentRepo                      interfaces.PaymentRepositoryInterface
	paymentTransactionRepo           interfaces.PaymentTransactionRepositoryInterface
	paymentTransactionAttachmentRepo interfaces.PaymentTransactionAttachmentRepositoryInterface
	unitRepo                         interfaces.UnitRepositoryInterface
	buildingRepo                     interfaces.BuildingRepositoryInterface
	propertyRepo                     interfaces.PropertyRepositoryInterface
	auditService                     interfaces.AuditServiceInterface
	userRepo                         interfaces.UserRepositoryInterface
}

func NewPaymentService(
	paymentRepo interfaces.PaymentRepositoryInterface,
	paymentTransactionRepo interfaces.PaymentTransactionRepositoryInterface,
	paymentTransactionAttachmentRepo interfaces.PaymentTransactionAttachmentRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo interfaces.PropertyRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
	userRepo interfaces.UserRepositoryInterface,
) *PaymentService {
	return &PaymentService{
		paymentRepo:                      paymentRepo,
		paymentTransactionRepo:           paymentTransactionRepo,
		paymentTransactionAttachmentRepo: paymentTransactionAttachmentRepo,
		unitRepo:                         unitRepo,
		buildingRepo:                     buildingRepo,
		propertyRepo:                     propertyRepo,
		auditService:                     auditService,
		userRepo:                         userRepo,
	}
}

// CreatePayment creates a new payment record with building information
func (s *PaymentService) CreatePayment(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
	// Validate unit exists and get building context
	unit, err := s.unitRepo.GetByID(req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Validate building and property IDs match unit's hierarchy
	if unit.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("building ID mismatch with unit's building")
	}
	if unit.PropertyID != req.PropertyID {
		return nil, fmt.Errorf("property ID mismatch with unit's property")
	}

	// Validate the unit actually belongs to the caller's organization — prevents
	// creating a payment record against another organization's unit (IDOR).
	if unit.OrganizationID != req.OrganizationID {
		return nil, fmt.Errorf("unit not found")
	}

	// Get building information for context
	building, err := s.buildingRepo.GetByID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get property information for context
	property, err := s.propertyRepo.GetByID(req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Reject a second payment for the same unit/month/year up front so the
	// user sees a clear error instead of a raw DB unique-constraint failure.
	exists, err := s.paymentRepo.CheckPaymentExists(req.UnitID, req.Month, req.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("a payment already exists for this unit for the selected month/year")
	}

	// Receipt numbers are always server-generated — a client-supplied value is
	// discarded so numbering stays sequential and collision-free per org/period.
	receiptNumber, err := s.paymentRepo.NextReceiptNumber(req.OrganizationID, fmt.Sprintf("%04d%02d", req.Year, req.Month))
	if err != nil {
		return nil, fmt.Errorf("failed to generate receipt number: %w", err)
	}
	req.ReceiptNumber = &receiptNumber

	// Create payment with building context
	payment, err := s.paymentRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Log audit with complete building context
	_ = s.auditService.LogUserAction(userID, "CREATE", "payments", &payment.ID, nil, map[string]interface{}{
		"payment_id":    payment.ID,
		"unit_id":       payment.UnitID,
		"building_id":   payment.BuildingID,
		"property_id":   payment.PropertyID,
		"building_name": building.BuildingName,
		"building_code": building.BuildingCode,
		"property_name": property.PropertyName,
		"unit_number":   unit.UnitNumber,
		"amount_due":    payment.AmountDue,
		"month":         payment.Month,
		"year":          payment.Year,
	})

	return payment, nil
}

// GetPayment retrieves a payment with building and property context
func (s *PaymentService) GetPayment(id, orgID int) (*models.PaymentWithDetails, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment with details: %w", err)
	}
	if payment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}
	return payment, nil
}

// UpdatePayment updates a payment record with building context logging
func (s *PaymentService) UpdatePayment(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error) {
	// Get existing payment for audit and building context
	existingPayment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing payment: %w", err)
	}
	if existingPayment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	// Status is never accepted from the client — PaymentRepository.Update
	// always (re)derives it from amount_paid vs amount_due.
	req.Status = nil

	// AmountPaid and ReceiptNumber are never accepted from the client either
	// — recording money received must go through RecordPaymentTransaction
	// so partial/installment payments accumulate correctly instead of
	// overwriting each other, and every payment received gets its own
	// server-generated receipt.
	req.AmountPaid = nil
	req.ReceiptNumber = nil

	// Update payment
	updatedPayment, err := s.paymentRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Log audit with building context
	_ = s.auditService.LogUserAction(userID, "UPDATE", "payments", &id, map[string]interface{}{
		"payment_id":      existingPayment.ID,
		"building_name":   existingPayment.BuildingName,
		"building_code":   existingPayment.BuildingCode,
		"property_name":   existingPayment.PropertyName,
		"unit_number":     existingPayment.UnitNumber,
		"old_status":      existingPayment.Status,
		"old_amount_paid": existingPayment.AmountPaid,
	}, map[string]interface{}{
		"payment_id":      updatedPayment.ID,
		"building_id":     updatedPayment.BuildingID,
		"property_id":     updatedPayment.PropertyID,
		"new_status":      updatedPayment.Status,
		"new_amount_paid": updatedPayment.AmountPaid,
	})

	return updatedPayment, nil
}

// RecordPaymentTransaction records a new amount received against a payment.
// This is the only way amount_paid ever changes — it adds to the existing
// total rather than replacing it, so a second (or third) installment against
// the same month's due amount accumulates correctly instead of overwriting
// the first, and each amount received gets its own receipt number.
func (s *PaymentService) RecordPaymentTransaction(paymentID int, req *models.CreatePaymentTransactionRequest, userID, orgID int) (*models.Payment, error) {
	existingPayment, err := s.paymentRepo.GetByIDWithDetails(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing payment: %w", err)
	}
	if existingPayment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	paymentDateStr := time.Now().UTC().Format("2006-01-02")
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		paymentDateStr = *req.PaymentDate
	}
	paymentDate, err := time.Parse("2006-01-02", paymentDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid payment_date format (expected YYYY-MM-DD): %w", err)
	}

	// Receipt numbers are always server-generated — every amount received
	// gets its own, matching how a landlord would actually hand out receipts
	// for separate installments.
	yearMonth := fmt.Sprintf("%04d%02d", existingPayment.Year, existingPayment.Month)
	receiptNumber, err := s.paymentRepo.NextReceiptNumber(existingPayment.OrganizationID, yearMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to generate receipt number: %w", err)
	}

	method := ""
	if req.PaymentMethod != nil {
		method = *req.PaymentMethod
	}
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}

	if _, err := s.paymentTransactionRepo.Create(paymentID, req.Amount, method, receiptNumber, notes, paymentDate); err != nil {
		return nil, fmt.Errorf("failed to record payment transaction: %w", err)
	}

	updatedPayment, err := s.refreshPaymentFromTransactions(paymentID)
	if err != nil {
		return nil, err
	}

	_ = s.auditService.LogUserAction(userID, "RECORD_PAYMENT", "payments", &paymentID, map[string]interface{}{
		"payment_id":      existingPayment.ID,
		"old_amount_paid": existingPayment.AmountPaid,
	}, map[string]interface{}{
		"payment_id":          paymentID,
		"transaction_amount":  req.Amount,
		"new_amount_paid":     updatedPayment.AmountPaid,
		"new_status":          updatedPayment.Status,
		"transaction_receipt": receiptNumber,
	})

	return updatedPayment, nil
}

// GetPaymentTransactions returns the ledger of amounts received against a
// payment, earliest first.
func (s *PaymentService) GetPaymentTransactions(paymentID, orgID int) ([]*models.PaymentTransaction, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	txns, err := s.paymentTransactionRepo.GetByPaymentID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transactions: %w", err)
	}
	return txns, nil
}

// DeletePaymentTransaction removes a mistakenly-recorded transaction and
// recomputes the payment's cached amount_paid/status/etc. from what remains.
func (s *PaymentService) DeletePaymentTransaction(paymentID, transactionID, userID, orgID int) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	txn, err := s.paymentTransactionRepo.GetByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}
	if txn.PaymentID != paymentID {
		return nil, fmt.Errorf("payment transaction not found")
	}

	if err := s.paymentTransactionRepo.Delete(transactionID); err != nil {
		return nil, fmt.Errorf("failed to delete payment transaction: %w", err)
	}

	updatedPayment, err := s.refreshPaymentFromTransactions(paymentID)
	if err != nil {
		return nil, err
	}

	_ = s.auditService.LogUserAction(userID, "DELETE_PAYMENT_TRANSACTION", "payments", &paymentID, txn, map[string]interface{}{
		"payment_id":      paymentID,
		"new_amount_paid": updatedPayment.AmountPaid,
		"new_status":      updatedPayment.Status,
	})

	return updatedPayment, nil
}

// getOwnedTransaction loads a transaction and verifies it belongs both to
// the given payment and, transitively, to the caller's organization —
// mirroring the same parent-chain ownership check used throughout
// DeletePaymentTransaction/GetPaymentTransactions.
func (s *PaymentService) getOwnedTransaction(paymentID, transactionID, orgID int) (*models.PaymentTransaction, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	txn, err := s.paymentTransactionRepo.GetByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}
	if txn.PaymentID != paymentID {
		return nil, fmt.Errorf("payment transaction not found")
	}
	return txn, nil
}

// UploadPaymentTransactionAttachment stores a file (receipt photo, mobile
// banking screenshot, etc.) as evidence of one installment. The content type
// is derived from the actual file bytes (never trusted from the client) and
// restricted to images and PDF; size is capped at MaxAttachmentFileSize.
func (s *PaymentService) UploadPaymentTransactionAttachment(paymentID, transactionID int, fileName string, data []byte, userID, orgID int) (*models.PaymentTransactionAttachment, error) {
	if _, err := s.getOwnedTransaction(paymentID, transactionID, orgID); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}
	if len(data) > models.MaxAttachmentFileSize {
		return nil, fmt.Errorf("file exceeds the maximum allowed size of %dMB", models.MaxAttachmentFileSize/(1024*1024))
	}

	detectedType := http.DetectContentType(data)
	if !models.AllowedAttachmentContentTypes[detectedType] {
		return nil, fmt.Errorf("unsupported file type %q — only images and PDF files are allowed", detectedType)
	}

	if fileName == "" {
		fileName = "attachment"
	}

	attachment, err := s.paymentTransactionAttachmentRepo.Create(transactionID, fileName, detectedType, len(data), data, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to store attachment: %w", err)
	}

	_ = s.auditService.LogUserAction(userID, "UPLOAD_PAYMENT_ATTACHMENT", "payment_transaction_attachments", &attachment.ID, nil, map[string]interface{}{
		"payment_id":     paymentID,
		"transaction_id": transactionID,
		"file_name":      fileName,
		"file_size":      len(data),
	})

	return attachment, nil
}

// GetPaymentTransactionAttachments lists the attachments recorded against one transaction.
func (s *PaymentService) GetPaymentTransactionAttachments(paymentID, transactionID, orgID int) ([]*models.PaymentTransactionAttachment, error) {
	if _, err := s.getOwnedTransaction(paymentID, transactionID, orgID); err != nil {
		return nil, err
	}
	atts, err := s.paymentTransactionAttachmentRepo.GetByTransactionID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction attachments: %w", err)
	}
	return atts, nil
}

// GetPaymentTransactionAttachmentFile returns an attachment's raw bytes for download.
func (s *PaymentService) GetPaymentTransactionAttachmentFile(paymentID, transactionID, attachmentID, orgID int) ([]byte, string, string, error) {
	if _, err := s.getOwnedTransaction(paymentID, transactionID, orgID); err != nil {
		return nil, "", "", err
	}

	attachment, err := s.paymentTransactionAttachmentRepo.GetByID(attachmentID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment.PaymentTransactionID != transactionID {
		return nil, "", "", fmt.Errorf("attachment not found")
	}

	data, fileName, contentType, err := s.paymentTransactionAttachmentRepo.GetFileData(attachmentID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get attachment file: %w", err)
	}
	return data, fileName, contentType, nil
}

// DeletePaymentTransactionAttachment removes a mistakenly-uploaded attachment.
func (s *PaymentService) DeletePaymentTransactionAttachment(paymentID, transactionID, attachmentID, userID, orgID int) error {
	if _, err := s.getOwnedTransaction(paymentID, transactionID, orgID); err != nil {
		return err
	}

	attachment, err := s.paymentTransactionAttachmentRepo.GetByID(attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}
	if attachment.PaymentTransactionID != transactionID {
		return fmt.Errorf("attachment not found")
	}

	if err := s.paymentTransactionAttachmentRepo.Delete(attachmentID); err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	_ = s.auditService.LogUserAction(userID, "DELETE_PAYMENT_ATTACHMENT", "payment_transaction_attachments", &attachmentID, attachment, nil)

	return nil
}

// refreshPaymentFromTransactions recomputes a payment's cached amount_paid/
// status/payment_method/payment_date/receipt_number from its transaction
// ledger. After RecordPaymentTransaction/DeletePaymentTransaction, this is
// the only place those fields are ever written, so payments.* always mirrors
// "the sum of what's actually been paid, plus the most recent payment
// event's details" — or a cleared/Due state once no transactions remain.
func (s *PaymentService) refreshPaymentFromTransactions(paymentID int) (*models.Payment, error) {
	total, err := s.paymentTransactionRepo.SumByPaymentID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to total payment transactions: %w", err)
	}

	txns, err := s.paymentTransactionRepo.GetByPaymentID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load payment transactions: %w", err)
	}

	method, receipt, dateStr := "", "", ""
	if len(txns) > 0 {
		latest := txns[len(txns)-1] // GetByPaymentID orders by payment_date, id ascending
		method = latest.PaymentMethod
		receipt = latest.ReceiptNumber
		dateStr = latest.PaymentDate.Format("2006-01-02")
	}

	updated, err := s.paymentRepo.Update(paymentID, &models.UpdatePaymentRequest{
		AmountPaid:    &total,
		PaymentMethod: &method,
		ReceiptNumber: &receipt,
		PaymentDate:   &dateStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	return updated, nil
}

// GetPayments retrieves payments scoped only by the provided filters (no entity validation)
func (s *PaymentService) GetPayments(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments: %w", err)
	}
	return payments, total, nil
}

// GetPaymentsByBuilding retrieves payments for a specific building
func (s *PaymentService) GetPaymentsByBuilding(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	// Validate building exists
	_, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, 0, fmt.Errorf("building not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Add building filter
	filters["building_id"] = buildingID

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments by building: %w", err)
	}

	return payments, total, nil
}

// GetPaymentsByProperty retrieves payments for a specific property with building breakdown
func (s *PaymentService) GetPaymentsByProperty(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, 0, fmt.Errorf("property not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Add property filter
	filters["property_id"] = propertyID

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments by property: %w", err)
	}

	return payments, total, nil
}

// GenerateBuildingPaymentReport generates payment report for a building
func (s *PaymentService) GenerateBuildingPaymentReport(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
	// Validate building exists and get details
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("building not found")
	}

	// Get property details for context
	property, err := s.propertyRepo.GetByID(building.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get payment statistics for the building
	stats, err := s.paymentRepo.GetBuildingPaymentStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment stats: %w", err)
	}

	// Get detailed payment records
	payments, _, err := s.paymentRepo.GetBuildingPaymentsInPeriod(buildingID, startDate, endDate, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payments: %w", err)
	}

	report := &models.BuildingPaymentReport{
		BuildingID:   buildingID,
		BuildingName: building.BuildingName,
		BuildingCode: building.BuildingCode,
		PropertyID:   property.ID,
		PropertyName: property.PropertyName,
		ReportPeriod: fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Statistics:   stats,
		Payments:     payments,
		GeneratedAt:  time.Now(),
	}

	return report, nil
}

// GeneratePropertyPaymentReport generates payment report for a property with building breakdowns
func (s *PaymentService) GeneratePropertyPaymentReport(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
	// Validate property exists and get details
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}
	if property.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}

	// Get buildings in the property
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property buildings: %w", err)
	}

	// Get overall property payment statistics
	propertyStats, err := s.paymentRepo.GetPropertyPaymentStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}

	// Get building-level breakdowns
	buildingBreakdowns := make([]*models.BuildingPaymentBreakdown, 0, len(buildings))
	for _, building := range buildings {
		buildingStats, err := s.paymentRepo.GetBuildingPaymentStats(building.ID, startDate, endDate)
		if err != nil {
			continue // Skip buildings with errors but don't fail the entire report
		}

		breakdown := &models.BuildingPaymentBreakdown{
			BuildingID:   building.ID,
			BuildingName: building.BuildingName,
			BuildingCode: building.BuildingCode,
			BuildingType: string(building.BuildingType),
			Statistics:   buildingStats,
		}
		buildingBreakdowns = append(buildingBreakdowns, breakdown)
	}

	report := &models.PropertyPaymentReport{
		PropertyID:         propertyID,
		PropertyName:       property.PropertyName,
		PropertyCode:       property.PropertyCode,
		ReportPeriod:       fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		OverallStatistics:  propertyStats,
		BuildingBreakdowns: buildingBreakdowns,
		GeneratedAt:        time.Now(),
	}

	return report, nil
}

// GetDashboardSummaryWithBuildingContext returns dashboard summary scoped to the given org.
func (s *PaymentService) GetDashboardSummaryWithBuildingContext(orgID int) (*models.DashboardSummary, error) {
	summary, err := s.paymentRepo.GetDashboardSummary(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	buildingLevel, err := s.paymentRepo.GetBuildingLevelSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get building-level summary: %w", err)
	}
	if count, ok := buildingLevel["total_buildings"].(int); ok {
		summary.BuildingCount = count
	}
	return summary, nil
}

// ProcessBulkPayments processes multiple payments with building context validation
func (s *PaymentService) ProcessBulkPayments(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
	payments := make([]*models.Payment, 0, len(requests))
	errors := make([]error, 0)

	for i, req := range requests {
		payment, err := s.CreatePayment(req, userID)
		if err != nil {
			errors = append(errors, fmt.Errorf("payment %d: %w", i+1, err))
			continue
		}
		payments = append(payments, payment)
	}

	return payments, errors
}

// GetPaymentAnalyticsByBuilding returns payment analytics for a specific building
func (s *PaymentService) GetPaymentAnalyticsByBuilding(buildingID int, period string) (*models.BuildingPaymentAnalytics, error) {
	// Validate building exists
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Calculate date range based on period
	endDate := time.Now()
	var startDate time.Time
	switch period {
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "quarter":
		startDate = endDate.AddDate(0, -3, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, -1, 0) // Default to month
	}

	// Get payment analytics
	analytics, err := s.paymentRepo.GetBuildingPaymentAnalytics(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment analytics: %w", err)
	}

	// Add building context
	analytics.BuildingName = building.BuildingName
	analytics.BuildingCode = building.BuildingCode
	analytics.BuildingType = string(building.BuildingType)

	return analytics, nil
}

// CanUserAccessPayment verifies if a user can access a specific payment based on their role and organization
func (s *PaymentService) CanUserAccessPayment(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
	// Must belong to same organization
	if payment.OrganizationID != userOrgID {
		return false
	}

	switch userRole {
	case "SUPER_ADMIN":
		return true // Can access all payments in their organization
	case "ORG_ADMIN":
		return true // Can access all org payments
	case "Admin":
		return true // Can access all org payments
	case "PropertyManager":
		// Can only access payments for properties they manage
		return s.canPropertyManagerAccessPayment(userID, payment.PropertyID)
	case "Accountant":
		return true // Read-only access to all payments in org
	default:
		return false
	}
}

// canPropertyManagerAccessPayment checks if a PropertyManager manages the specified property
func (s *PaymentService) canPropertyManagerAccessPayment(userID int, propertyID int) bool {
	// Get user to verify they manage this property
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	// PropertyManagers should have property assignment info
	// This assumes user model has a field indicating which properties they manage
	// For now, we'll check if they're assigned to this property via metadata or a relationship
	if user == nil {
		return false
	}

	// TODO: Implement property manager assignment lookup
	// For now, allow access if user is PropertyManager (assumes assignment validation elsewhere)
	return true
}

// LogPaymentAccess logs access to payment data for audit trail
func (s *PaymentService) LogPaymentAccess(userID int, action string, paymentID int, allowed bool) {
	accessLog := map[string]interface{}{
		"payment_id": paymentID,
		"allowed":    allowed,
		"action":     action,
	}

	if allowed {
		_ = s.auditService.LogUserAction(userID, action, "payments", &paymentID, nil, accessLog)
	} else {
		_ = s.auditService.LogUserAction(userID, fmt.Sprintf("%s_DENIED", action), "payments", &paymentID, nil, accessLog)
	}
}

// GenerateMonthlyPayments creates Due payment records for every active lease in the given month/year.
// Leases that already have a payment record for that unit/month/year are skipped.
func (s *PaymentService) GenerateMonthlyPayments(req *models.GenerateMonthlyPaymentsRequest, orgID, userID int) (*models.GenerateMonthlyPaymentsResult, error) {
	now := time.Now().UTC()
	currentYear, currentMonth := now.Year(), int(now.Month())

	// Reject dates more than 1 month ahead of today.
	reqMonths := req.Year*12 + req.Month
	currentMonths := currentYear*12 + currentMonth
	if reqMonths > currentMonths+1 {
		return nil, fmt.Errorf("cannot generate payments more than 1 month in the future (requested %04d-%02d)", req.Year, req.Month)
	}

	// Reject dates older than 24 months to prevent mass back-generation.
	if reqMonths < currentMonths-24 {
		return nil, fmt.Errorf("cannot generate payments more than 24 months in the past (requested %04d-%02d)", req.Year, req.Month)
	}

	leases, err := s.paymentRepo.GetActiveLeasesForPeriod(orgID, req.Month, req.Year, req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active leases: %w", err)
	}

	dueDay := req.DueDayOfMonth
	if dueDay < 1 || dueDay > 28 {
		dueDay = 7
	}
	lastDay := time.Date(req.Year, time.Month(req.Month+1), 0, 0, 0, 0, 0, time.UTC).Day()
	if dueDay > lastDay {
		dueDay = lastDay
	}
	dueDateStr := fmt.Sprintf("%04d-%02d-%02d", req.Year, req.Month, dueDay)

	result := &models.GenerateMonthlyPaymentsResult{Errors: []string{}}

	for _, lease := range leases {
		exists, err := s.paymentRepo.CheckPaymentExists(lease.UnitID, req.Month, req.Year)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("unit %d: existence check failed: %v", lease.UnitID, err))
			continue
		}
		if exists {
			result.Skipped++
			continue
		}

		// Assign a receipt number up front, matching CreatePayment, so every
		// payment record has one from the moment it exists rather than only
		// when created through the single-entry flow.
		receiptNumber, err := s.paymentRepo.NextReceiptNumber(orgID, fmt.Sprintf("%04d%02d", req.Year, req.Month))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("unit %d: failed to generate receipt number: %v", lease.UnitID, err))
			continue
		}

		createReq := &models.CreatePaymentRequest{
			UnitID:         lease.UnitID,
			TenantID:       lease.TenantID,
			BuildingID:     lease.BuildingID,
			PropertyID:     lease.PropertyID,
			OrganizationID: orgID,
			Month:          req.Month,
			Year:           req.Year,
			// AmountDue includes the lease's recurring charges (utility, service
			// charge, etc.) on top of the base rent — see GetActiveLeasesForPeriod.
			AmountDue:     lease.MonthlyRent + lease.ChargesTotal,
			DueDate:       dueDateStr,
			ReceiptNumber: &receiptNumber,
		}

		if _, err := s.paymentRepo.Create(createReq); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("unit %d: %v", lease.UnitID, err))
			continue
		}
		result.Generated++
	}

	_ = s.auditService.LogUserAction(userID, "GENERATE_MONTHLY", "payments", nil, nil, map[string]interface{}{
		"month":       req.Month,
		"year":        req.Year,
		"org_id":      orgID,
		"generated":   result.Generated,
		"skipped":     result.Skipped,
		"failed":      result.Failed,
		"building_id": req.BuildingID,
	})

	return result, nil
}

// SearchLeases searches for active leases by tenant name, property, building, unit, or lease ID
func (s *PaymentService) SearchLeases(orgID int, query string) (*models.LeaseSearchResponse, error) {
	if query == "" {
		return &models.LeaseSearchResponse{Results: make([]*models.LeaseSearchResult, 0), Total: 0}, nil
	}

	results, err := s.paymentRepo.SearchLeases(orgID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search leases: %w", err)
	}

	return &models.LeaseSearchResponse{
		Results: results,
		Total:   len(results),
	}, nil
}
