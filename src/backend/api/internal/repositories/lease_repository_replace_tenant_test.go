package repositories

import (
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/testutil"
)

// turnoverFixture is one unit with an active lease held by outgoingTenantID, plus
// a second tenant available to take it over.
type turnoverFixture struct {
	repo             *LeaseRepository
	orgID            int
	unitID           int
	outgoingTenantID int
	incomingTenantID int
	outgoingLease    *models.Lease
}

func setupTurnoverFixture(t *testing.T) (*turnoverFixture, func()) {
	t.Helper()

	repo, cleanup := setupTestLeaseRepository(t)

	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	outgoingTenantID := testutil.CreateTestTenant(t, repo.db, orgID)
	incomingTenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	startDate := time.Now().AddDate(0, -6, 0).Format("2006-01-02")
	lease, err := repo.Create(&models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        outgoingTenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       startDate,
		DurationMonths:  12,
		MonthlyRent:     20000,
		SecurityDeposit: 40000,
		OrganizationID:  orgID,
	})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create the outgoing lease: %v", err)
	}

	return &turnoverFixture{
		repo:             repo,
		orgID:            orgID,
		unitID:           unitID,
		outgoingTenantID: outgoingTenantID,
		incomingTenantID: incomingTenantID,
		outgoingLease:    lease,
	}, cleanup
}

func (f *turnoverFixture) successorRequest(startDate string, tenantID int) *models.CreateLeaseRequest {
	return &models.CreateLeaseRequest{
		UnitID:          f.unitID,
		TenantID:        tenantID,
		LeaseType:       f.outgoingLease.LeaseType,
		StartDate:       startDate,
		DurationMonths:  f.outgoingLease.DurationMonths,
		MonthlyRent:     f.outgoingLease.MonthlyRent,
		SecurityDeposit: f.outgoingLease.SecurityDeposit,
		OrganizationID:  f.orgID,
	}
}

func (f *turnoverFixture) activeLeaseCountOnUnit(t *testing.T) int {
	t.Helper()
	var count int
	err := f.repo.db.QueryRow(
		`SELECT COUNT(*) FROM leases WHERE unit_id = $1 AND active = true`, f.unitID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count active leases: %v", err)
	}
	return count
}

func TestLeaseRepository_ReplaceTenant(t *testing.T) {
	f, cleanup := setupTurnoverFixture(t)
	defer cleanup()

	handoverStr := time.Now().Format("2006-01-02")
	handoverDate, err := time.Parse("2006-01-02", handoverStr)
	if err != nil {
		t.Fatalf("Failed to parse the handover date: %v", err)
	}

	successor, err := f.repo.ReplaceTenant(
		f.outgoingLease.ID,
		handoverDate,
		f.successorRequest(handoverStr, f.incomingTenantID),
	)
	if err != nil {
		t.Fatalf("ReplaceTenant failed: %v", err)
	}

	// The successor must be a distinct, active lease for the incoming tenant.
	if successor.ID == f.outgoingLease.ID {
		t.Error("Expected a new lease row, got the outgoing lease back")
	}
	if !successor.Active {
		t.Error("Expected the successor lease to be active")
	}
	if successor.TenantID != f.incomingTenantID {
		t.Errorf("Expected successor tenant %d, got %d", f.incomingTenantID, successor.TenantID)
	}
	if successor.UnitID != f.unitID {
		t.Errorf("Expected successor unit %d, got %d", f.unitID, successor.UnitID)
	}
	if !successor.StartDate.Equal(handoverDate) {
		t.Errorf("Expected successor to start on %v, got %v", handoverDate, successor.StartDate)
	}
	if successor.MonthlyRent != f.outgoingLease.MonthlyRent {
		t.Errorf("Expected carried-over rent %v, got %v", f.outgoingLease.MonthlyRent, successor.MonthlyRent)
	}

	// The outgoing lease must be closed, ending on the handover date.
	closed, err := f.repo.GetByID(f.outgoingLease.ID)
	if err != nil {
		t.Fatalf("Failed to re-read the outgoing lease: %v", err)
	}
	if closed.Active {
		t.Error("Expected the outgoing lease to be closed")
	}
	if !closed.EndDate.Equal(handoverDate) {
		t.Errorf("Expected the outgoing lease to end on %v, got %v", handoverDate, closed.EndDate)
	}
	if closed.TenantID != f.outgoingTenantID {
		t.Errorf("Outgoing lease tenant changed to %d — history must stay attributed to %d",
			closed.TenantID, f.outgoingTenantID)
	}

	// Exactly one active lease on the unit: no double-lease, no silent vacancy.
	if got := f.activeLeaseCountOnUnit(t); got != 1 {
		t.Errorf("Expected 1 active lease on the unit after turnover, got %d", got)
	}
}

// TestLeaseRepository_ReplaceTenant_RejectsClosedLease covers the concurrency
// guard: the UPDATE matches only active rows, so a turnover on an already-closed
// lease affects nothing and rolls back.
func TestLeaseRepository_ReplaceTenant_RejectsClosedLease(t *testing.T) {
	f, cleanup := setupTurnoverFixture(t)
	defer cleanup()

	if err := f.repo.SoftDelete(f.outgoingLease.ID, time.Now(), models.LeaseEndReasonTerminated); err != nil {
		t.Fatalf("Failed to close the outgoing lease: %v", err)
	}

	handoverStr := time.Now().Format("2006-01-02")
	handoverDate, _ := time.Parse("2006-01-02", handoverStr)

	_, err := f.repo.ReplaceTenant(
		f.outgoingLease.ID,
		handoverDate,
		f.successorRequest(handoverStr, f.incomingTenantID),
	)
	if err == nil {
		t.Fatal("Expected ReplaceTenant to reject an already-closed lease")
	}

	if got := f.activeLeaseCountOnUnit(t); got != 0 {
		t.Errorf("Expected no successor lease to be created, found %d active leases", got)
	}
}

// TestLeaseRepository_ReplaceTenant_RollsBackOnFailedInsert is the atomicity
// check: if the successor insert fails, the outgoing lease must stay active
// rather than leaving the unit silently vacant.
func TestLeaseRepository_ReplaceTenant_RollsBackOnFailedInsert(t *testing.T) {
	f, cleanup := setupTurnoverFixture(t)
	defer cleanup()

	handoverStr := time.Now().Format("2006-01-02")
	handoverDate, _ := time.Parse("2006-01-02", handoverStr)

	// A tenant_id that violates the foreign key, so the INSERT fails after the
	// outgoing lease has already been closed inside the transaction.
	successor := f.successorRequest(handoverStr, -1)

	if _, err := f.repo.ReplaceTenant(f.outgoingLease.ID, handoverDate, successor); err == nil {
		t.Fatal("Expected ReplaceTenant to fail on an invalid successor tenant")
	}

	outgoing, err := f.repo.GetByID(f.outgoingLease.ID)
	if err != nil {
		t.Fatalf("Failed to re-read the outgoing lease: %v", err)
	}
	if !outgoing.Active {
		t.Error("Outgoing lease was left closed after a failed turnover — the transaction did not roll back")
	}
	if got := f.activeLeaseCountOnUnit(t); got != 1 {
		t.Errorf("Expected the original lease to remain the only active one, got %d", got)
	}
}
