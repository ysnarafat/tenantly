package repositories

import (
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/testutil"
)

func setupTestLeaseRepository(t *testing.T) (*LeaseRepository, func()) {
	db, cleanup := testutil.SetupTestDB(t)
	return NewLeaseRepository(db), cleanup
}

func TestLeaseRepository_Create(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	lease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	if lease.ID == 0 {
		t.Error("Expected lease ID to be set")
	}

	if lease.UnitID != req.UnitID {
		t.Errorf("Expected unit ID %d, got %d", req.UnitID, lease.UnitID)
	}

	if lease.TenantID != req.TenantID {
		t.Errorf("Expected tenant ID %d, got %d", req.TenantID, lease.TenantID)
	}

	if !lease.Active {
		t.Error("Expected lease to be active")
	}
}

func TestLeaseRepository_GetByID(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Get by ID
	lease, err := repo.GetByID(createdLease.ID)
	if err != nil {
		t.Fatalf("Failed to get lease by ID: %v", err)
	}

	if lease.ID != createdLease.ID {
		t.Errorf("Expected lease ID %d, got %d", createdLease.ID, lease.ID)
	}

	if lease.UnitID != unitID {
		t.Errorf("Expected unit ID %d, got %d", unitID, lease.UnitID)
	}
}

func TestLeaseRepository_GetAll(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	// Create multiple leases
	for i := 0; i < 3; i++ {
		unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
		endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

		req := &models.CreateLeaseRequest{
			UnitID:          unitID,
			TenantID:        tenantID,
			LeaseType:       models.LeaseTypeResidential,
			StartDate:       time.Now().Format("2006-01-02"),
			EndDate:         &endDate,
			DurationMonths:  12,
			MonthlyRent:     15000,
			SecurityDeposit: 30000,
			OrganizationID:  orgID,
		}
		_, err := repo.Create(req)
		if err != nil {
			t.Fatalf("Failed to create lease: %v", err)
		}
	}

	// Get all leases
	leases, total, err := repo.GetAll(1, 10, orgID)
	if err != nil {
		t.Fatalf("Failed to get all leases: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}

	if len(leases) != 3 {
		t.Errorf("Expected 3 leases, got %d", len(leases))
	}
}

func TestLeaseRepository_Update(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Update lease
	newRent := 20000.0
	newDuration := 24
	updateReq := &models.UpdateLeaseRequest{
		MonthlyRent:    &newRent,
		DurationMonths: &newDuration,
	}

	lease, err := repo.Update(createdLease.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update lease: %v", err)
	}

	if lease.MonthlyRent != newRent {
		t.Errorf("Expected monthly rent %f, got %f", newRent, lease.MonthlyRent)
	}

	if lease.DurationMonths != newDuration {
		t.Errorf("Expected duration %d, got %d", newDuration, lease.DurationMonths)
	}
}

func TestLeaseRepository_Delete(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Delete lease
	err = repo.Delete(createdLease.ID)
	if err != nil {
		t.Fatalf("Failed to delete lease: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(createdLease.ID)
	if err == nil {
		t.Error("Expected error when getting deleted lease")
	}
}

func TestLeaseRepository_HasActiveLeaseOnUnit(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	// Check before creating lease
	hasLease, err := repo.HasActiveLeaseOnUnit(unitID, nil)
	if err != nil {
		t.Fatalf("Failed to check active lease: %v", err)
	}

	if hasLease {
		t.Error("Expected no active lease for unit")
	}

	// Create active lease
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Check after creating lease
	hasLease, err = repo.HasActiveLeaseOnUnit(unitID, &createdLease.ID)
	if err != nil {
		t.Fatalf("Failed to check active lease: %v", err)
	}

	if !hasLease {
		t.Error("Expected active lease for unit")
	}
}

func TestLeaseRepository_GetByUnitID(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	// Create lease
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Get by unit ID
	leases, total, err := repo.GetByUnitID(unitID, 1, 10, orgID)
	if err != nil {
		t.Fatalf("Failed to get leases by unit ID: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total count 1, got %d", total)
	}

	if len(leases) != 1 {
		t.Errorf("Expected 1 lease, got %d", len(leases))
	}

	if leases[0].ID != createdLease.ID {
		t.Errorf("Expected lease ID %d, got %d", createdLease.ID, leases[0].ID)
	}
}

func TestLeaseRepository_GetByTenantID(t *testing.T) {
	repo, cleanup := setupTestLeaseRepository(t)
	defer cleanup()

	// Create test data
	orgID := testutil.CreateTestOrganization(t, repo.db)
	propertyID := testutil.CreateTestProperty(t, repo.db)
	buildingID := testutil.CreateTestBuilding(t, repo.db, propertyID, orgID)
	unitID := testutil.CreateTestUnit(t, repo.db, buildingID, orgID)
	tenantID := testutil.CreateTestTenant(t, repo.db, orgID)

	// Create lease
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	req := &models.CreateLeaseRequest{
		UnitID:          unitID,
		TenantID:        tenantID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  orgID,
	}

	createdLease, err := repo.Create(req)
	if err != nil {
		t.Fatalf("Failed to create lease: %v", err)
	}

	// Get by tenant ID
	leases, total, err := repo.GetByTenantID(tenantID, 1, 10, orgID)
	if err != nil {
		t.Fatalf("Failed to get leases by tenant ID: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected total count 1, got %d", total)
	}

	if len(leases) != 1 {
		t.Errorf("Expected 1 lease, got %d", len(leases))
	}

	if leases[0].ID != createdLease.ID {
		t.Errorf("Expected lease ID %d, got %d", createdLease.ID, leases[0].ID)
	}
}
