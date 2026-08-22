package services

import (
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

	paymentService := NewPaymentService(paymentRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo)

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

	paymentService := NewPaymentService(paymentRepo, unitRepo, buildingRepo, propertyRepo, auditService, userRepo)
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
