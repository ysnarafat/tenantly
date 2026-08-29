package services

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/testutil"
)

// TestGenerateMonthlyPayments_IncludesLeaseCharges is a real-DB integration
// test (unlike the mock-based tests in payment_service_test.go) — it exists
// specifically to verify that generated payments bill a lease's active
// recurring charges on top of its base rent, and that each one gets a
// receipt number, since neither is exercisable through the mocked
// GetActiveLeasesForPeriod used elsewhere in this package.
func TestGenerateMonthlyPayments_IncludesLeaseCharges(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	leaseRepo := repositories.NewLeaseRepository(db)
	chargeRepo := repositories.NewLeaseChargeRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	unitRepo := repositories.NewUnitRepository(db)
	buildingRepo := repositories.NewBuildingRepository(db)
	propertyRepo := repositories.NewPropertyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditService := database.NewAuditService(db)

	lease, err := leaseRepo.Create(&models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       "2026-01-01",
		DurationMonths:  12,
		MonthlyRent:     5000,
		SecurityDeposit: 10000,
		OrganizationID:  orgID,
	})
	if err != nil {
		t.Fatalf("failed to create lease: %v", err)
	}

	if _, err := chargeRepo.Create(lease.ID, &models.CreateLeaseChargeRequest{
		ChargeType: models.ChargeTypeUtility,
		Label:      "Electricity",
		Amount:     500,
	}); err != nil {
		t.Fatalf("failed to add lease charge: %v", err)
	}
	if _, err := chargeRepo.Create(lease.ID, &models.CreateLeaseChargeRequest{
		ChargeType: models.ChargeTypeServiceCharge,
		Label:      "Building maintenance",
		Amount:     250,
	}); err != nil {
		t.Fatalf("failed to add second lease charge: %v", err)
	}

	paymentService := NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, receiptAccessTokenRepo, notificationRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo, "http://localhost:8080")

	result, err := paymentService.GenerateMonthlyPayments(&models.GenerateMonthlyPaymentsRequest{
		Month: 6,
		Year:  2026,
	}, orgID, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Generated != 1 {
		t.Fatalf("expected 1 payment generated, got %d (failed=%d, errors=%v)", result.Generated, result.Failed, result.Errors)
	}

	payments, total, err := paymentRepo.GetWithDetailsAndFilters(map[string]interface{}{
		"organization_id": orgID,
		"month":           6,
		"year":            2026,
	}, 10, 0)
	if err != nil {
		t.Fatalf("failed to fetch generated payment: %v", err)
	}
	if total != 1 || len(payments) != 1 {
		t.Fatalf("expected exactly 1 matching payment, got %d", total)
	}

	payment := payments[0]
	// 5000 base rent + 500 utility + 250 service charge.
	const wantAmountDue = 5750.0
	if payment.AmountDue != wantAmountDue {
		t.Errorf("AmountDue: got %.2f, want %.2f (rent + active charges)", payment.AmountDue, wantAmountDue)
	}
	if payment.ReceiptNumber == "" {
		t.Error("expected a receipt number to be assigned to the generated payment, got empty string")
	}
	// Nothing has been paid yet, so status must be Due or (once the due date
	// has passed relative to whenever this test runs) Overdue — never Paid
	// or Partial.
	if payment.Status != models.PaymentStatusDue && payment.Status != models.PaymentStatusOverdue {
		t.Errorf("Status: got %q, want Due or Overdue (nothing paid yet)", payment.Status)
	}
}

// TestGenerateMonthlyPayments_SkipsDeactivatedCharges verifies a discontinued
// charge (Active=false) is not billed, mirroring
// TestPaymentRepository_GetActiveLeasesForPeriod's repository-level check but
// through the actual service entry point tenants/landlords go through.
func TestGenerateMonthlyPayments_SkipsDeactivatedCharges(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	leaseRepo := repositories.NewLeaseRepository(db)
	chargeRepo := repositories.NewLeaseChargeRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	unitRepo := repositories.NewUnitRepository(db)
	buildingRepo := repositories.NewBuildingRepository(db)
	propertyRepo := repositories.NewPropertyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditService := database.NewAuditService(db)

	lease, err := leaseRepo.Create(&models.CreateLeaseRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		LeaseType:      models.LeaseTypeResidential,
		StartDate:      "2026-01-01",
		DurationMonths: 12,
		MonthlyRent:    5000,
		OrganizationID: orgID,
	})
	if err != nil {
		t.Fatalf("failed to create lease: %v", err)
	}

	charge, err := chargeRepo.Create(lease.ID, &models.CreateLeaseChargeRequest{
		ChargeType: models.ChargeTypeParking,
		Label:      "Parking",
		Amount:     300,
	})
	if err != nil {
		t.Fatalf("failed to add lease charge: %v", err)
	}
	inactive := false
	if _, err := chargeRepo.Update(charge.ID, &models.UpdateLeaseChargeRequest{Active: &inactive}); err != nil {
		t.Fatalf("failed to deactivate lease charge: %v", err)
	}

	paymentService := NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, receiptAccessTokenRepo, notificationRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo, "http://localhost:8080")
	if _, err := paymentService.GenerateMonthlyPayments(&models.GenerateMonthlyPaymentsRequest{
		Month: 6,
		Year:  2026,
	}, orgID, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payments, _, err := paymentRepo.GetWithDetailsAndFilters(map[string]interface{}{
		"organization_id": orgID,
		"month":           6,
		"year":            2026,
	}, 10, 0)
	if err != nil {
		t.Fatalf("failed to fetch generated payment: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected exactly 1 matching payment, got %d", len(payments))
	}
	if payments[0].AmountDue != 5000.0 {
		t.Errorf("AmountDue: got %.2f, want 5000.00 (deactivated charge must not be billed)", payments[0].AmountDue)
	}
}

// TestRecordPaymentTransaction_RealDB_InstallmentsAccumulate is the
// real-DB counterpart to payment_service_test.go's mock-based
// TestRecordPaymentTransaction_AccumulatesAcrossInstallments — it exists to
// exercise the actual SQL status-derivation CASE and receipt-number
// sequence (NextReceiptNumber), neither of which the mock replicates, for
// exactly the scenario this feature was built for: a tenant paying a
// 12,000 due amount as 10,000 today and the remaining 2,000 later.
func TestRecordPaymentTransaction_RealDB_InstallmentsAccumulate(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	paymentRepo := repositories.NewPaymentRepository(db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	unitRepo := repositories.NewUnitRepository(db)
	buildingRepo := repositories.NewBuildingRepository(db)
	propertyRepo := repositories.NewPropertyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditService := database.NewAuditService(db)

	payment, err := paymentRepo.Create(&models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		BuildingID:     bldgID,
		PropertyID:     propID,
		OrganizationID: orgID,
		Month:          6,
		Year:           2026,
		AmountDue:      12000,
	})
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	paymentService := NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, receiptAccessTokenRepo, notificationRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo, "http://localhost:8080")

	afterFirst, err := paymentService.RecordPaymentTransaction(payment.ID, &models.CreatePaymentTransactionRequest{
		Amount:        10000,
		PaymentMethod: strPtr("Cash"),
		PaymentDate:   strPtr("2026-06-05"),
	}, 1, orgID)
	if err != nil {
		t.Fatalf("unexpected error recording first installment: %v", err)
	}
	if afterFirst.AmountPaid != 10000 {
		t.Errorf("after first installment: amount_paid got %.2f, want 10000.00", afterFirst.AmountPaid)
	}
	if afterFirst.Status != models.PaymentStatusPartial {
		t.Errorf("after first installment: status got %q, want %q", afterFirst.Status, models.PaymentStatusPartial)
	}

	afterSecond, err := paymentService.RecordPaymentTransaction(payment.ID, &models.CreatePaymentTransactionRequest{
		Amount:        2000,
		PaymentMethod: strPtr("bKash"),
		PaymentDate:   strPtr("2026-06-20"),
	}, 1, orgID)
	if err != nil {
		t.Fatalf("unexpected error recording second installment: %v", err)
	}
	if afterSecond.AmountPaid != 12000 {
		t.Errorf("after second installment: amount_paid got %.2f, want 12000.00", afterSecond.AmountPaid)
	}
	if afterSecond.Status != models.PaymentStatusPaid {
		t.Errorf("after second installment: status got %q, want %q", afterSecond.Status, models.PaymentStatusPaid)
	}
	// The cached snapshot on the payments row reflects the *latest* event.
	if afterSecond.PaymentMethod != "bKash" {
		t.Errorf("expected payment_method to mirror the latest installment (bKash), got %q", afterSecond.PaymentMethod)
	}

	txns, err := paymentService.GetPaymentTransactions(payment.ID, orgID)
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txns))
	}
	if txns[0].ReceiptNumber == txns[1].ReceiptNumber {
		t.Error("expected each installment to get its own receipt number")
	}
	if txns[0].Amount+txns[1].Amount != 12000 {
		t.Errorf("transaction amounts sum to %.2f, want 12000.00", txns[0].Amount+txns[1].Amount)
	}
}

// TestPaymentTransactionAttachment_RealDB_UploadListDownloadDelete exercises
// the actual bytea storage/round-trip through Postgres — the mock-based
// tests in payment_service_test.go verify the validation and access-control
// logic, but not that a real file survives a genuine INSERT/SELECT.
func TestPaymentTransactionAttachment_RealDB_UploadListDownloadDelete(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	paymentRepo := repositories.NewPaymentRepository(db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	unitRepo := repositories.NewUnitRepository(db)
	buildingRepo := repositories.NewBuildingRepository(db)
	propertyRepo := repositories.NewPropertyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditService := database.NewAuditService(db)

	payment, err := paymentRepo.Create(&models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		BuildingID:     bldgID,
		PropertyID:     propID,
		OrganizationID: orgID,
		Month:          6,
		Year:           2026,
		AmountDue:      12000,
	})
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	paymentService := NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, receiptAccessTokenRepo, notificationRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo, "http://localhost:8080")

	afterFirst, err := paymentService.RecordPaymentTransaction(payment.ID, &models.CreatePaymentTransactionRequest{
		Amount: 10000,
	}, 1, orgID)
	if err != nil {
		t.Fatalf("unexpected error recording installment: %v", err)
	}
	txns, err := paymentService.GetPaymentTransactions(payment.ID, orgID)
	if err != nil || len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d (err=%v)", len(txns), err)
	}
	transactionID := txns[0].ID
	_ = afterFirst

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x02, 0x03}
	uploaded, err := paymentService.UploadPaymentTransactionAttachment(payment.ID, transactionID, "bkash-screenshot.jpg", jpegBytes, 1, orgID)
	if err != nil {
		t.Fatalf("unexpected error uploading attachment: %v", err)
	}
	if uploaded.ContentType != "image/jpeg" {
		t.Errorf("content type got %q, want image/jpeg", uploaded.ContentType)
	}
	if uploaded.FileSize != len(jpegBytes) {
		t.Errorf("file size got %d, want %d", uploaded.FileSize, len(jpegBytes))
	}

	atts, err := paymentService.GetPaymentTransactionAttachments(payment.ID, transactionID, orgID)
	if err != nil {
		t.Fatalf("unexpected error listing attachments: %v", err)
	}
	if len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(atts))
	}

	data, fileName, contentType, err := paymentService.GetPaymentTransactionAttachmentFile(payment.ID, transactionID, uploaded.ID, orgID)
	if err != nil {
		t.Fatalf("unexpected error downloading attachment: %v", err)
	}
	if string(data) != string(jpegBytes) {
		t.Error("downloaded bytes do not match uploaded bytes after a real DB round-trip")
	}
	if fileName != "bkash-screenshot.jpg" {
		t.Errorf("file name got %q, want bkash-screenshot.jpg", fileName)
	}
	if contentType != "image/jpeg" {
		t.Errorf("content type got %q, want image/jpeg", contentType)
	}

	if err := paymentService.DeletePaymentTransactionAttachment(payment.ID, transactionID, uploaded.ID, 1, orgID); err != nil {
		t.Fatalf("unexpected error deleting attachment: %v", err)
	}
	attsAfterDelete, err := paymentService.GetPaymentTransactionAttachments(payment.ID, transactionID, orgID)
	if err != nil {
		t.Fatalf("unexpected error listing attachments after delete: %v", err)
	}
	if len(attsAfterDelete) != 0 {
		t.Errorf("expected 0 attachments after delete, got %d", len(attsAfterDelete))
	}
}

// TestReceiptLinkSms_RealDB_EndToEnd exercises the actual flow behind "give
// the tenant a URL to download their receipt via SMS": recording a payment
// queues a real notification_queue row with a working link, and that link
// (a bare token, no auth) actually renders the receipt via a genuine
// receipt_access_tokens round-trip through Postgres.
func TestReceiptLinkSms_RealDB_EndToEnd(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID) // seeded with phone_number +8801234567890

	paymentRepo := repositories.NewPaymentRepository(db)
	paymentTransactionRepo := repositories.NewPaymentTransactionRepository(db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	unitRepo := repositories.NewUnitRepository(db)
	buildingRepo := repositories.NewBuildingRepository(db)
	propertyRepo := repositories.NewPropertyRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditService := database.NewAuditService(db)

	payment, err := paymentRepo.Create(&models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		BuildingID:     bldgID,
		PropertyID:     propID,
		OrganizationID: orgID,
		Month:          6,
		Year:           2026,
		AmountDue:      12000,
	})
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	paymentService := NewPaymentService(paymentRepo, paymentTransactionRepo, paymentTransactionAttachmentRepo, receiptAccessTokenRepo, notificationRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo, "https://app.example.com")

	if _, err := paymentService.RecordPaymentTransaction(payment.ID, &models.CreatePaymentTransactionRequest{
		Amount: 12000,
	}, 1, orgID); err != nil {
		t.Fatalf("unexpected error recording transaction: %v", err)
	}

	// A real row must have been queued for the .NET notification-service to
	// pick up and deliver.
	var queuedMessage, recipient, notificationType, status string
	err = db.QueryRow(`SELECT message, recipient, notification_type, status FROM notification_queue WHERE tenant_id = $1 ORDER BY id DESC LIMIT 1`, tenantID).
		Scan(&queuedMessage, &recipient, &notificationType, &status)
	if err != nil {
		t.Fatalf("expected a notification_queue row to have been inserted: %v", err)
	}
	if recipient != "+8801234567890" {
		t.Errorf("recipient got %q, want the tenant's phone number", recipient)
	}
	if notificationType != models.NotificationTypeSMS {
		t.Errorf("notification type got %q, want %q", notificationType, models.NotificationTypeSMS)
	}
	if status != models.NotificationStatusPending {
		t.Errorf("status got %q, want %q", status, models.NotificationStatusPending)
	}

	const linkPrefix = "https://app.example.com/api/v1/receipts/"
	idx := strings.Index(queuedMessage, linkPrefix)
	if idx == -1 {
		t.Fatalf("expected message to contain a receipt link, got %q", queuedMessage)
	}
	token := queuedMessage[idx+len(linkPrefix):]

	// The bare token — exactly what a tenant would have after tapping the
	// SMS link, no session/auth of any kind — must resolve to the receipt.
	pdfBytes, filename, err := paymentService.DownloadReceiptByToken(token)
	if err != nil {
		t.Fatalf("unexpected error downloading receipt by token: %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Error("expected valid PDF bytes")
	}
	if !strings.HasSuffix(filename, ".pdf") {
		t.Errorf("expected a .pdf filename, got %q", filename)
	}

	if _, _, err := paymentService.DownloadReceiptByToken("not-a-real-token"); err == nil {
		t.Error("expected an unknown token to be rejected")
	}
}

func strPtr(s string) *string {
	return &s
}
