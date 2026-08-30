package repositories

import (
	"strconv"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/testutil"
)

// ---------------------------------------------------------------------------
// TestPaymentRepository_Create
// ---------------------------------------------------------------------------

func TestPaymentRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     *models.CreatePaymentRequest
		wantErr bool
	}{
		{
			name: "creates payment with all required fields",
			req: &models.CreatePaymentRequest{
				Month:     5,
				Year:      2026,
				AmountDue: 5000.0,
			},
			wantErr: false,
		},
		{
			name: "handles leap year dates correctly",
			req: &models.CreatePaymentRequest{
				Month:     2,
				Year:      2024,
				AmountDue: 3500.0,
				DueDate:   "2024-02-29",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()

			orgID := testutil.CreateTestOrganization(t, db)
			propID := testutil.CreateTestProperty(t, db)
			bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
			unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
			tenantID := testutil.CreateTestTenant(t, db, orgID)

			tc.req.OrganizationID = orgID
			tc.req.PropertyID = propID
			tc.req.BuildingID = bldgID
			tc.req.UnitID = unitID
			tc.req.TenantID = tenantID

			repo := NewPaymentRepository(db)
			payment, err := repo.Create(tc.req)

			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr: %v, got error: %v", tc.wantErr, err)
				return
			}

			if !tc.wantErr && payment == nil {
				t.Errorf("expected payment to be created, got nil")
				return
			}

			if !tc.wantErr && payment.ID == 0 {
				t.Errorf("payment ID should be auto-incremented")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_Create_StatusIsServerDerived
// ---------------------------------------------------------------------------

func TestPaymentRepository_Create_StatusIsServerDerived(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	repo := NewPaymentRepository(db)

	// A client claiming Status=Paid with AmountPaid=0 must not be trusted —
	// status is always derived from amount_paid vs amount_due.
	claimedPaid := models.PaymentStatusPaid
	payment, err := repo.Create(&models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		BuildingID:     bldgID,
		PropertyID:     propID,
		OrganizationID: orgID,
		Month:          6,
		Year:           2026,
		AmountDue:      5000.0,
		AmountPaid:     nil,
		Status:         &claimedPaid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != models.PaymentStatusDue {
		t.Errorf("expected server-derived status %q, got %q", models.PaymentStatusDue, payment.Status)
	}
}

func TestPaymentRepository_Create_PaymentDateDefaultsToToday(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	repo := NewPaymentRepository(db)
	today := time.Now().UTC().Format("2006-01-02")

	t.Run("amount_paid > 0 and no payment_date given defaults to today", func(t *testing.T) {
		amountPaid := 5000.0
		payment, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          9,
			Year:           2026,
			AmountDue:      5000.0,
			AmountPaid:     &amountPaid,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.PaymentDate == nil {
			t.Fatal("expected payment_date to default to today, got nil")
		}
		if got := payment.PaymentDate.Format("2006-01-02"); got != today {
			t.Errorf("payment_date: got %q, want %q", got, today)
		}
	})

	t.Run("amount_paid unset leaves payment_date nil", func(t *testing.T) {
		payment, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          10,
			Year:           2026,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.PaymentDate != nil {
			t.Errorf("expected nil payment_date for an unpaid Due record, got %v", payment.PaymentDate)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetByID
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetByID(t *testing.T) {
	t.Run("returns payment with correct fields", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		createdPayment, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          5,
			Year:           2026,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		retrieved, err := repo.GetByID(createdPayment.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if retrieved == nil {
			t.Errorf("expected payment to be retrieved, got nil")
			return
		}

		if retrieved.ID != createdPayment.ID {
			t.Errorf("ID mismatch: got %d, want %d", retrieved.ID, createdPayment.ID)
		}

		if retrieved.UnitID != unitID {
			t.Errorf("UnitID mismatch: got %d, want %d", retrieved.UnitID, unitID)
		}
	})

	t.Run("returns error for non-existent payment", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		repo := NewPaymentRepository(db)
		_, err := repo.GetByID(99999)

		if err == nil {
			t.Errorf("expected error for non-existent payment, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetByIDWithDetails
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetByIDWithDetails(t *testing.T) {
	t.Run("returns payment with related entity details", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		createdPayment, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          6,
			Year:           2026,
			AmountDue:      4500.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		paymentDetails, err := repo.GetByIDWithDetails(createdPayment.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if paymentDetails == nil {
			t.Errorf("expected PaymentWithDetails, got nil")
			return
		}

		if paymentDetails.PropertyName == "" {
			t.Errorf("PropertyName should be populated")
		}

		if paymentDetails.BuildingName == "" {
			t.Errorf("BuildingName should be populated")
		}

		if paymentDetails.UnitNumber == "" {
			t.Errorf("UnitNumber should be populated")
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_Update
// ---------------------------------------------------------------------------

func TestPaymentRepository_Update(t *testing.T) {
	t.Run("updates payment status and amount paid", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          7,
			Year:           2026,
			AmountDue:      6000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		newStatus := models.PaymentStatusPaid
		newAmount := 6000.0
		updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{
			Status:        &newStatus,
			AmountPaid:    &newAmount,
			PaymentMethod: ptrString("Bank Transfer"),
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if updated.Status != newStatus {
			t.Errorf("Status: got %q, want %q", updated.Status, newStatus)
		}

		if updated.AmountPaid != newAmount {
			t.Errorf("AmountPaid: got %.2f, want %.2f", updated.AmountPaid, newAmount)
		}
	})

	t.Run("returns error for non-existent payment", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		repo := NewPaymentRepository(db)
		newStatus := models.PaymentStatusPaid
		_, err := repo.Update(99999, &models.UpdatePaymentRequest{
			Status: &newStatus,
		})

		if err == nil {
			t.Errorf("expected error for non-existent payment, got nil")
		}
	})

	t.Run("client-supplied status is ignored — always derived from amount_paid", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          8,
			Year:           2026,
			AmountDue:      6000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		// Claim Due while actually paying the full amount — the claimed value
		// must be ignored and the server-derived one (Paid) must win.
		claimedDue := models.PaymentStatusDue
		fullAmount := 6000.0
		updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{
			Status:     &claimedDue,
			AmountPaid: &fullAmount,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != models.PaymentStatusPaid {
			t.Errorf("expected server-derived status %q (ignoring claimed %q), got %q",
				models.PaymentStatusPaid, claimedDue, updated.Status)
		}
	})

	t.Run("derives Overdue when amount_paid stays zero past due_date", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		pastDueDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          8,
			Year:           2026,
			AmountDue:      6000.0,
			DueDate:        pastDueDate,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}
		if created.Status != models.PaymentStatusOverdue {
			t.Fatalf("expected newly-created payment past its due_date to be Overdue, got %q", created.Status)
		}

		// Re-saving amount_paid=0 (e.g. correcting an unrelated field) must
		// keep deriving Overdue, not silently reset to Due.
		zero := 0.0
		updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{AmountPaid: &zero})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Status != models.PaymentStatusOverdue {
			t.Errorf("expected status to stay %q, got %q", models.PaymentStatusOverdue, updated.Status)
		}
	})

	t.Run("amount_paid > 0 and no payment_date given defaults to today", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          11,
			Year:           2026,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		amountPaid := 5000.0
		updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{AmountPaid: &amountPaid})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.PaymentDate == nil {
			t.Fatal("expected payment_date to default to today, got nil")
		}
		today := time.Now().UTC().Format("2006-01-02")
		if got := updated.PaymentDate.Format("2006-01-02"); got != today {
			t.Errorf("payment_date: got %q, want %q", got, today)
		}
	})

	t.Run("explicit payment_date is honored over the today default", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          12,
			Year:           2026,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		amountPaid := 5000.0
		backdated := "2026-01-15"
		updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{
			AmountPaid:  &amountPaid,
			PaymentDate: &backdated,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.PaymentDate == nil || updated.PaymentDate.Format("2006-01-02") != backdated {
			t.Errorf("payment_date: got %v, want %q (explicit date must win over today default)", updated.PaymentDate, backdated)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetWithDetailsAndFilters
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetWithDetailsAndFilters(t *testing.T) {
	t.Run("filters payments by organization_id", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID1 := testutil.CreateTestOrganization(t, db)
		_ = testutil.CreateTestOrganization(t, db)

		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID1)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID1)
		tenantID := testutil.CreateTestTenant(t, db, orgID1)

		repo := NewPaymentRepository(db)

		// Create payments for org1
		for i := range 3 {
			_, err := repo.Create(&models.CreatePaymentRequest{
				UnitID:         unitID,
				TenantID:       tenantID,
				BuildingID:     bldgID,
				PropertyID:     propID,
				OrganizationID: orgID1,
				Month:          int((time.Now().Month())) + i,
				Year:           2026,
				AmountDue:      5000.0,
			})
			if err != nil {
				t.Fatalf("failed to create payment: %v", err)
			}
		}

		filters := map[string]interface{}{"organization_id": orgID1}
		payments, total, err := repo.GetWithDetailsAndFilters(filters, 10, 0)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if total < 3 {
			t.Errorf("expected at least 3 payments, got %d", total)
		}

		// Verify all returned payments belong to org1
		for _, p := range payments {
			if p.OrganizationID != orgID1 {
				t.Errorf("expected OrganizationID %d, got %d", orgID1, p.OrganizationID)
			}
		}
	})

	t.Run("respects pagination limit and offset", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create 10 payments
		for i := range 10 {
			_, err := repo.Create(&models.CreatePaymentRequest{
				UnitID:         unitID,
				TenantID:       tenantID,
				BuildingID:     bldgID,
				PropertyID:     propID,
				OrganizationID: orgID,
				Month:          (i % 12) + 1,
				Year:           2026,
				AmountDue:      1000.0 * float64(i+1),
			})
			if err != nil {
				t.Fatalf("failed to create payment: %v", err)
			}
		}

		filters := map[string]interface{}{"organization_id": orgID}

		// Get first page
		payments1, total1, err := repo.GetWithDetailsAndFilters(filters, 3, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(payments1) != 3 {
			t.Errorf("expected 3 payments in first page, got %d", len(payments1))
		}

		// Get second page
		payments2, _, err := repo.GetWithDetailsAndFilters(filters, 3, 3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(payments2) != 3 {
			t.Errorf("expected 3 payments in second page, got %d", len(payments2))
		}

		// Verify no overlap between pages
		for _, p1 := range payments1 {
			for _, p2 := range payments2 {
				if p1.ID == p2.ID {
					t.Errorf("found duplicate payment ID across pages")
				}
			}
		}

		if total1 < 10 {
			t.Errorf("expected total >= 10, got %d", total1)
		}
	})

	t.Run("filters by building_id", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID1 := testutil.CreateTestBuilding(t, db, propID, orgID)
		bldgID2 := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID1 := testutil.CreateTestUnit(t, db, bldgID1, orgID)
		unitID2 := testutil.CreateTestUnit(t, db, bldgID2, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create payments in different buildings
		_, _ = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID1,
			TenantID:       tenantID,
			BuildingID:     bldgID1,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          1,
			Year:           2026,
			AmountDue:      5000.0,
		})

		_, _ = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID2,
			TenantID:       tenantID,
			BuildingID:     bldgID2,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          1,
			Year:           2026,
			AmountDue:      5000.0,
		})

		filters := map[string]interface{}{
			"organization_id": orgID,
			"building_id":     bldgID1,
		}
		payments, _, err := repo.GetWithDetailsAndFilters(filters, 10, 0)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		// Verify all returned payments belong to bldgID1
		for _, p := range payments {
			if p.BuildingID != bldgID1 {
				t.Errorf("expected BuildingID %d, got %d", bldgID1, p.BuildingID)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetBuildingPaymentStats
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetBuildingPaymentStats(t *testing.T) {
	t.Run("returns payment statistics for building", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create payments with different statuses
		paidAmount := 5000.0
		paidStatus := models.PaymentStatusPaid
		_, _ = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          1,
			Year:           2026,
			AmountDue:      5000.0,
		})
		paymentID := 1
		_, _ = repo.Update(paymentID, &models.UpdatePaymentRequest{
			Status:     &paidStatus,
			AmountPaid: &paidAmount,
		})

		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

		stats, err := repo.GetBuildingPaymentStats(bldgID, startDate, endDate)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if stats == nil {
			t.Errorf("expected stats, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_DashboardSummary
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetDashboardSummary(t *testing.T) {
	t.Run("returns dashboard summary", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create some payments
		_, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          1,
			Year:           2026,
			AmountDue:      10000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		summary, err := repo.GetDashboardSummary(orgID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if summary == nil {
			t.Errorf("expected summary, got nil")
			return
		}

		if summary.TotalDue < 0 {
			t.Errorf("TotalDue should be non-negative, got %.2f", summary.TotalDue)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetAgingBuckets
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetAgingBuckets(t *testing.T) {
	t.Run("returns outstanding balances in correct buckets", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		now := time.Now()
		currentYear := now.Year()
		currentMonth := int(now.Month())

		// Helper to compute a (year, month) that is N months back from current
		monthsBack := func(n int) (int, int) {
			y, m := currentYear, currentMonth-n
			for m <= 0 {
				m += 12
				y--
			}
			return y, m
		}

		// current bucket: this month, unpaid (Due)
		_, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          currentMonth,
			Year:           currentYear,
			AmountDue:      1000.0,
		})
		if err != nil {
			t.Fatalf("failed to create current-bucket payment: %v", err)
		}

		// 30d bucket: one month ago, unpaid (Due)
		y30, m30 := monthsBack(1)
		_, err = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          m30,
			Year:           y30,
			AmountDue:      2000.0,
		})
		if err != nil {
			t.Fatalf("failed to create 30d-bucket payment: %v", err)
		}

		// 60d bucket: two months ago, unpaid (Due)
		y60, m60 := monthsBack(2)
		_, err = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          m60,
			Year:           y60,
			AmountDue:      3000.0,
		})
		if err != nil {
			t.Fatalf("failed to create 60d-bucket payment: %v", err)
		}

		// 90d+ bucket: three months ago, unpaid (Due)
		y90, m90 := monthsBack(3)
		_, err = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          m90,
			Year:           y90,
			AmountDue:      4000.0,
		})
		if err != nil {
			t.Fatalf("failed to create 90d+-bucket payment: %v", err)
		}

		buckets, err := repo.GetAgingBuckets(orgID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if buckets == nil {
			t.Fatal("expected non-nil buckets map")
		}

		// Verify the four keys are always present
		for _, key := range []string{"current", "30d", "60d", "90d+"} {
			if _, ok := buckets[key]; !ok {
				t.Errorf("missing expected bucket key %q", key)
			}
		}

		// Each bucket should contain at least the amount we inserted
		if buckets["current"] < 1000 {
			t.Errorf("current bucket: got %v, want >= 1000", buckets["current"])
		}
		if buckets["30d"] < 2000 {
			t.Errorf("30d bucket: got %v, want >= 2000", buckets["30d"])
		}
		if buckets["60d"] < 3000 {
			t.Errorf("60d bucket: got %v, want >= 3000", buckets["60d"])
		}
		if buckets["90d+"] < 4000 {
			t.Errorf("90d+ bucket: got %v, want >= 4000", buckets["90d+"])
		}

		// All bucket values must be non-negative
		for key, val := range buckets {
			if val < 0 {
				t.Errorf("bucket %q has negative value %v", key, val)
			}
		}
	})

	t.Run("returns zero buckets for org with no payments", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)

		repo := NewPaymentRepository(db)
		buckets, err := repo.GetAgingBuckets(orgID)
		if err != nil {
			t.Fatalf("unexpected error for empty org: %v", err)
		}

		if buckets == nil {
			t.Fatal("expected non-nil buckets map even for empty org")
		}

		for _, key := range []string{"current", "30d", "60d", "90d+"} {
			if val, ok := buckets[key]; !ok {
				t.Errorf("missing bucket key %q", key)
			} else if val != 0 {
				t.Errorf("bucket %q: got %v, want 0 for empty org", key, val)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetMonthlyCollectionTrend
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetMonthlyCollectionTrend(t *testing.T) {
	t.Run("returns trend rows for months with payments", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		now := time.Now()
		currentYear := now.Year()
		currentMonth := int(now.Month())

		// Create a paid payment this month
		created, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          currentMonth,
			Year:           currentYear,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}

		paidAmount := 5000.0
		paidStatus := models.PaymentStatusPaid
		_, err = repo.Update(created.ID, &models.UpdatePaymentRequest{
			Status:     &paidStatus,
			AmountPaid: &paidAmount,
		})
		if err != nil {
			t.Fatalf("failed to update payment: %v", err)
		}

		trend, err := repo.GetMonthlyCollectionTrend(orgID, 6)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(trend) == 0 {
			t.Fatal("expected at least one trend entry, got none")
		}

		// Find the entry for the current month
		var found *models.MonthlyCollectionTrend
		for _, entry := range trend {
			if entry.Month.Year() == currentYear && int(entry.Month.Month()) == currentMonth {
				found = entry
				break
			}
		}
		if found == nil {
			t.Fatalf("no trend entry found for current month %d/%d", currentMonth, currentYear)
		}

		if found.AmountDue < 5000 {
			t.Errorf("AmountDue: got %v, want >= 5000", found.AmountDue)
		}
		if found.AmountCollected < 5000 {
			t.Errorf("AmountCollected: got %v, want >= 5000", found.AmountCollected)
		}
		if found.CollectionRate <= 0 {
			t.Errorf("CollectionRate: got %f, want > 0 for a fully paid month", found.CollectionRate)
		}
	})

	t.Run("returns empty slice for org with no payments", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)

		repo := NewPaymentRepository(db)
		trend, err := repo.GetMonthlyCollectionTrend(orgID, 6)
		if err != nil {
			t.Fatalf("unexpected error for empty org: %v", err)
		}

		if len(trend) != 0 {
			t.Errorf("expected empty trend for org with no payments, got %d entries", len(trend))
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetTenantPaymentSummary
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetTenantPaymentSummary(t *testing.T) {
	t.Run("returns summary entry per tenant with correct aggregates", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create two payments for the same tenant
		for i, month := range []int{1, 2} {
			p, err := repo.Create(&models.CreatePaymentRequest{
				UnitID:         unitID,
				TenantID:       tenantID,
				BuildingID:     bldgID,
				PropertyID:     propID,
				OrganizationID: orgID,
				Month:          month,
				Year:           2026,
				AmountDue:      3000.0,
			})
			if err != nil {
				t.Fatalf("failed to create payment %d: %v", i, err)
			}

			// Mark first payment as fully paid
			if month == 1 {
				paidAmount := 3000.0
				paidStatus := models.PaymentStatusPaid
				_, err = repo.Update(p.ID, &models.UpdatePaymentRequest{
					Status:     &paidStatus,
					AmountPaid: &paidAmount,
				})
				if err != nil {
					t.Fatalf("failed to update payment: %v", err)
				}
			}
		}

		entries, err := repo.GetTenantPaymentSummary(orgID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(entries) == 0 {
			t.Fatal("expected at least one tenant summary entry, got none")
		}

		// Find our tenant's entry
		var found *models.TenantReportEntry
		for _, e := range entries {
			if e.TenantID == tenantID {
				found = e
				break
			}
		}
		if found == nil {
			t.Fatalf("no entry found for tenantID %d", tenantID)
		}

		if found.TenantName == "" {
			t.Errorf("expected non-empty TenantName")
		}

		// Total due across both payments = 6000
		if found.TotalDue < 6000 {
			t.Errorf("TotalDue: got %.2f, want >= 6000", found.TotalDue)
		}

		// Total paid = 3000 (only first payment was paid)
		if found.TotalPaid < 3000 {
			t.Errorf("TotalPaid: got %.2f, want >= 3000", found.TotalPaid)
		}

		// BalanceDue should be TotalDue - TotalPaid
		expectedBalance := found.TotalDue - found.TotalPaid
		if found.BalanceDue != expectedBalance {
			t.Errorf("BalanceDue: got %.2f, want %.2f", found.BalanceDue, expectedBalance)
		}
	})

	t.Run("returns empty slice for org with no tenants", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)

		repo := NewPaymentRepository(db)
		entries, err := repo.GetTenantPaymentSummary(orgID)
		if err != nil {
			t.Fatalf("unexpected error for empty org: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("expected empty summary for org with no tenants, got %d entries", len(entries))
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetPaymentAnalyticsByPeriod
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetPaymentAnalyticsByPeriod(t *testing.T) {
	t.Run("returns method, status, and daily counts for payments in range", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		repo := NewPaymentRepository(db)

		// Create a paid payment via bank transfer
		p1, err := repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          1,
			Year:           2026,
			AmountDue:      5000.0,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}
		paidAmount := 5000.0
		paidStatus := models.PaymentStatusPaid
		_, err = repo.Update(p1.ID, &models.UpdatePaymentRequest{
			Status:        &paidStatus,
			AmountPaid:    &paidAmount,
			PaymentMethod: ptrString("Bank Transfer"),
		})
		if err != nil {
			t.Fatalf("failed to update payment: %v", err)
		}

		// Create a second unpaid (Due) payment
		_, err = repo.Create(&models.CreatePaymentRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			BuildingID:     bldgID,
			PropertyID:     propID,
			OrganizationID: orgID,
			Month:          2,
			Year:           2026,
			AmountDue:      4000.0,
		})
		if err != nil {
			t.Fatalf("failed to create second payment: %v", err)
		}

		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

		result, err := repo.GetPaymentAnalyticsByPeriod(orgID, startDate, endDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil analytics result")
		}

		if result.MethodCounts == nil {
			t.Error("MethodCounts map should not be nil")
		}
		if result.StatusCounts == nil {
			t.Error("StatusCounts map should not be nil")
		}
		if result.DailyTrend == nil {
			t.Error("DailyTrend map should not be nil")
		}

		if result.TotalPayments < 2 {
			t.Errorf("TotalPayments: got %d, want >= 2", result.TotalPayments)
		}

		// Verify the paid payment's method appears
		if result.MethodCounts["Bank Transfer"] < 1 {
			t.Errorf("expected at least 1 Bank Transfer payment, got %d", result.MethodCounts["Bank Transfer"])
		}

		// Verify Paid status is counted
		if result.StatusCounts[string(models.PaymentStatusPaid)] < 1 {
			t.Errorf("expected at least 1 Paid status entry, got %d", result.StatusCounts[string(models.PaymentStatusPaid)])
		}

		// DailyTrend should have at least one entry (today's created_at)
		if len(result.DailyTrend) == 0 {
			t.Errorf("expected at least one DailyTrend entry, got none")
		}
	})

	t.Run("returns zero totals for org with no payments in period", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)

		repo := NewPaymentRepository(db)
		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

		result, err := repo.GetPaymentAnalyticsByPeriod(orgID, startDate, endDate)
		if err != nil {
			t.Fatalf("unexpected error for empty org: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result even for empty org")
		}

		if result.TotalPayments != 0 {
			t.Errorf("TotalPayments: got %d, want 0 for empty org", result.TotalPayments)
		}
		if len(result.MethodCounts) != 0 {
			t.Errorf("expected empty MethodCounts for empty org, got %v", result.MethodCounts)
		}
		if len(result.StatusCounts) != 0 {
			t.Errorf("expected empty StatusCounts for empty org, got %v", result.StatusCounts)
		}
		if len(result.DailyTrend) != 0 {
			t.Errorf("expected empty DailyTrend for empty org, got %v", result.DailyTrend)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_GetActiveLeasesForPeriod
// ---------------------------------------------------------------------------

func TestPaymentRepository_GetActiveLeasesForPeriod(t *testing.T) {
	t.Run("ChargesTotal sums only active lease charges", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		leaseRepo := NewLeaseRepository(db)
		chargeRepo := NewLeaseChargeRepository(db)

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
			t.Fatalf("failed to add active charge: %v", err)
		}
		discontinued, err := chargeRepo.Create(lease.ID, &models.CreateLeaseChargeRequest{
			ChargeType: models.ChargeTypeParking,
			Label:      "Parking (discontinued)",
			Amount:     300,
		})
		if err != nil {
			t.Fatalf("failed to add charge to deactivate: %v", err)
		}
		inactive := false
		if _, err := chargeRepo.Update(discontinued.ID, &models.UpdateLeaseChargeRequest{Active: &inactive}); err != nil {
			t.Fatalf("failed to deactivate charge: %v", err)
		}

		repo := NewPaymentRepository(db)
		results, err := repo.GetActiveLeasesForPeriod(orgID, 6, 2026, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var found *models.LeaseSearchResult
		for _, r := range results {
			if r.LeaseID == lease.ID {
				found = r
			}
		}
		if found == nil {
			t.Fatalf("expected lease %d to be among active leases for the period", lease.ID)
		}
		// Only the active $500 charge should count — the deactivated $300
		// parking charge must be excluded from the total billed to the tenant.
		if found.ChargesTotal != 500.0 {
			t.Errorf("ChargesTotal: got %.2f, want 500.00", found.ChargesTotal)
		}
	})

	t.Run("ChargesTotal is zero for a lease with no charges", func(t *testing.T) {
		db, cleanup := testutil.SetupTestDB(t)
		defer cleanup()

		orgID := testutil.CreateTestOrganization(t, db)
		propID := testutil.CreateTestProperty(t, db)
		bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)

		leaseRepo := NewLeaseRepository(db)
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

		repo := NewPaymentRepository(db)
		results, err := repo.GetActiveLeasesForPeriod(orgID, 6, 2026, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var found *models.LeaseSearchResult
		for _, r := range results {
			if r.LeaseID == lease.ID {
				found = r
			}
		}
		if found == nil {
			t.Fatalf("expected lease %d to be among active leases for the period", lease.ID)
		}
		if found.ChargesTotal != 0.0 {
			t.Errorf("ChargesTotal: got %.2f, want 0.00", found.ChargesTotal)
		}
	})
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_SearchLeases
//
// SearchLeases backs the lease-picker used by the manual "add payment"
// dialog (and the property-menu quick-add flow) — it must not surface a
// lease that is past its end_date, even if active hasn't been flipped to
// false yet, so a user can't pick an expired lease to bill against.
// ---------------------------------------------------------------------------

func TestPaymentRepository_SearchLeases(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	leaseRepo := NewLeaseRepository(db)
	repo := NewPaymentRepository(db)

	t.Run("active lease far from expiry is returned", func(t *testing.T) {
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)
		endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		lease, err := leaseRepo.Create(&models.CreateLeaseRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			LeaseType:      models.LeaseTypeResidential,
			StartDate:      time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
			EndDate:        &endDate,
			DurationMonths: 12,
			MonthlyRent:    5000,
			OrganizationID: orgID,
		})
		if err != nil {
			t.Fatalf("failed to create lease: %v", err)
		}

		results, err := repo.SearchLeases(orgID, strconv.Itoa(lease.ID))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !leaseSearchContains(results, lease.ID) {
			t.Errorf("expected lease %d (active, far from expiry) to be returned", lease.ID)
		}
	})

	t.Run("active lease whose end_date has already passed is excluded", func(t *testing.T) {
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)
		endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		lease, err := leaseRepo.Create(&models.CreateLeaseRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			LeaseType:      models.LeaseTypeResidential,
			StartDate:      time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
			EndDate:        &endDate,
			DurationMonths: 12,
			MonthlyRent:    5000,
			OrganizationID: orgID,
		})
		if err != nil {
			t.Fatalf("failed to create lease: %v", err)
		}
		if _, err := db.Exec(`UPDATE leases SET end_date = $1 WHERE id = $2`, time.Now().AddDate(0, 0, -1), lease.ID); err != nil {
			t.Fatalf("failed to force end_date into the past: %v", err)
		}

		results, err := repo.SearchLeases(orgID, strconv.Itoa(lease.ID))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if leaseSearchContains(results, lease.ID) {
			t.Errorf("expected expired lease %d to be excluded from search results", lease.ID)
		}
	})

	t.Run("terminated (active=false) lease is excluded", func(t *testing.T) {
		unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
		tenantID := testutil.CreateTestTenant(t, db, orgID)
		endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		lease, err := leaseRepo.Create(&models.CreateLeaseRequest{
			UnitID:         unitID,
			TenantID:       tenantID,
			LeaseType:      models.LeaseTypeResidential,
			StartDate:      time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
			EndDate:        &endDate,
			DurationMonths: 12,
			MonthlyRent:    5000,
			OrganizationID: orgID,
		})
		if err != nil {
			t.Fatalf("failed to create lease: %v", err)
		}
		if err := leaseRepo.SoftDelete(lease.ID, time.Now(), models.LeaseEndReasonTerminated); err != nil {
			t.Fatalf("failed to terminate lease: %v", err)
		}

		results, err := repo.SearchLeases(orgID, strconv.Itoa(lease.ID))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if leaseSearchContains(results, lease.ID) {
			t.Errorf("expected terminated lease %d to be excluded from search results", lease.ID)
		}
	})
}

func leaseSearchContains(results []*models.LeaseSearchResult, leaseID int) bool {
	for _, r := range results {
		if r.LeaseID == leaseID {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// TestPaymentRepository_Update_ClearsFieldsToNull
// ---------------------------------------------------------------------------

func TestPaymentRepository_Update_ClearsFieldsToNull(t *testing.T) {
	// PaymentService.refreshPaymentFromTransactions passes an empty string
	// (rather than nil) for payment_method/payment_date/receipt_number when
	// a payment's last remaining transaction is deleted, to explicitly clear
	// them instead of leaving stale values from the deleted transaction.
	// payment_date is a real DATE column, so binding "" as its value would
	// error at the driver level unless Update special-cases it — this test
	// exists specifically to catch a regression there.
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propID := testutil.CreateTestProperty(t, db)
	bldgID := testutil.CreateTestBuilding(t, db, propID, orgID)
	unitID := testutil.CreateTestUnit(t, db, bldgID, orgID)
	tenantID := testutil.CreateTestTenant(t, db, orgID)

	repo := NewPaymentRepository(db)
	amountPaid := 5000.0
	created, err := repo.Create(&models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		BuildingID:     bldgID,
		PropertyID:     propID,
		OrganizationID: orgID,
		Month:          9,
		Year:           2026,
		AmountDue:      5000.0,
		AmountPaid:     &amountPaid,
		PaymentMethod:  ptrString("Cash"),
		ReceiptNumber:  ptrString("RCP-CLEAR-TEST-1"),
	})
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}
	if created.PaymentDate == nil || created.PaymentMethod == "" || created.ReceiptNumber == "" {
		t.Fatalf("expected payment to start with method/date/receipt populated, got %+v", created)
	}

	empty := ""
	zero := 0.0
	updated, err := repo.Update(created.ID, &models.UpdatePaymentRequest{
		AmountPaid:    &zero,
		PaymentMethod: &empty,
		PaymentDate:   &empty,
		ReceiptNumber: &empty,
	})
	if err != nil {
		t.Fatalf("unexpected error clearing fields: %v", err)
	}
	if updated.PaymentDate != nil {
		t.Errorf("expected payment_date to be cleared to nil, got %v", updated.PaymentDate)
	}
	if updated.PaymentMethod != "" {
		t.Errorf("expected payment_method to be cleared, got %q", updated.PaymentMethod)
	}
	if updated.ReceiptNumber != "" {
		t.Errorf("expected receipt_number to be cleared, got %q", updated.ReceiptNumber)
	}
	if updated.Status != models.PaymentStatusDue {
		t.Errorf("expected status to revert to Due, got %q", updated.Status)
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func ptrString(s string) *string {
	return &s
}
