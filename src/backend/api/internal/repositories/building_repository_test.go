package repositories

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/testutil"

	_ "github.com/lib/pq"
)

// setupBuildingRepositoryTestDB creates a test DB using migrations.
func setupBuildingRepositoryTestDB(t *testing.T) (*sqlx.DB, func()) {
	db, cleanup := testutil.SetupTestDB(t)
	return db, cleanup
}

func setupBuildingRepository(t *testing.T) (*BuildingRepository, *sqlx.DB, func()) {
	db, cleanup := setupBuildingRepositoryTestDB(t)
	repo := NewBuildingRepository(db)
	return repo, db, cleanup
}

func createTestProperty(t *testing.T, db *sqlx.DB) int {
	return testutil.CreateTestProperty(t, db)
}

func TestBuildingRepository_Create(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)

	tests := []struct {
		name        string
		building    *models.Building
		expectError bool
	}{
		{
			name:        "Valid residential building creation",
			building:    &models.Building{PropertyID: propertyID, BuildingName: "Residential Tower A", BuildingCode: "RTA001", BuildingType: models.BuildingTypeResidential, TotalFloors: 10, HasElevator: true, ConstructionYear: func() *int { y := 2020; return &y }(), Metadata: models.BuildingMetadata{"amenities": []string{"gym", "pool"}}, ActiveStatus: true},
			expectError: false,
		},
		{
			name:        "Valid commercial building creation",
			building:    &models.Building{PropertyID: propertyID, BuildingName: "Commercial Block B", BuildingCode: "CBB001", BuildingType: models.BuildingTypeCommercial, TotalFloors: 5, HasElevator: false, Metadata: models.BuildingMetadata{"parking_spaces": 50}, ActiveStatus: true},
			expectError: false,
		},
		{
			name:        "Duplicate building code should fail",
			building:    &models.Building{PropertyID: propertyID, BuildingName: "Duplicate Building", BuildingCode: "RTA001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, ActiveStatus: true},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.building)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.building.ID == 0 {
					t.Errorf("Expected building ID to be set")
				}
				if tt.building.CreatedAt.IsZero() {
					t.Errorf("Expected CreatedAt to be set")
				}
			}
		})
	}
}

func TestBuildingRepository_GetByID(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	building := &models.Building{PropertyID: propertyID, BuildingName: "Test Building", BuildingCode: "TB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 5, HasElevator: true, Metadata: models.BuildingMetadata{"test": "value"}, ActiveStatus: true}
	if err := repo.Create(building); err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	tests := []struct {
		name        string
		buildingID  int
		expectError bool
	}{
		{name: "Valid building ID", buildingID: building.ID, expectError: false},
		{name: "Invalid building ID", buildingID: 99999, expectError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.buildingID)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected building but got nil")
				}
				if result.ID != tt.buildingID {
					t.Errorf("Expected building ID %d, got %d", tt.buildingID, result.ID)
				}
			}
		})
	}
}

func TestBuildingRepository_GetByPropertyID(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	buildings := []*models.Building{{PropertyID: propertyID, BuildingName: "Building A", BuildingCode: "BA001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, ActiveStatus: true}, {PropertyID: propertyID, BuildingName: "Building B", BuildingCode: "BB001", BuildingType: models.BuildingTypeCommercial, TotalFloors: 2, ActiveStatus: true}, {PropertyID: propertyID, BuildingName: "Inactive Building", BuildingCode: "IB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 1, ActiveStatus: false}}
	for _, b := range buildings {
		if err := repo.Create(b); err != nil {
			t.Fatalf("Failed to create test building: %v", err)
		}
	}
	result, err := repo.GetByPropertyID(propertyID)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 active buildings, got %d", len(result))
	}
	// Non-existent property
	result, err = repo.GetByPropertyID(99999)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 buildings for non-existent property, got %d", len(result))
	}
}

func TestBuildingRepository_GetByPropertyAndCode(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	building := &models.Building{PropertyID: propertyID, BuildingName: "Test Building", BuildingCode: "TB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 5, ActiveStatus: true}
	if err := repo.Create(building); err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	tests := []struct {
		name         string
		propertyID   int
		buildingCode string
		expectError  bool
	}{
		{name: "Valid property and code", propertyID: propertyID, buildingCode: "TB001", expectError: false},
		{name: "Invalid property ID", propertyID: 99999, buildingCode: "TB001", expectError: true},
		{name: "Invalid building code", propertyID: propertyID, buildingCode: "INVALID", expectError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByPropertyAndCode(tt.propertyID, tt.buildingCode)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected building but got nil")
				}
				if result.BuildingCode != tt.buildingCode {
					t.Errorf("Expected building code %s, got %s", tt.buildingCode, result.BuildingCode)
				}
			}
		})
	}
}

func TestBuildingRepository_Update(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	building := &models.Building{PropertyID: propertyID, BuildingName: "Original Building", BuildingCode: "OB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, HasElevator: false, ActiveStatus: true}
	if err := repo.Create(building); err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	tests := []struct {
		name        string
		buildingID  int
		updates     map[string]interface{}
		expectError bool
	}{
		{name: "Valid update", buildingID: building.ID, updates: map[string]interface{}{"building_name": "Updated Building", "total_floors": 5, "has_elevator": true}, expectError: false},
		{name: "Update metadata", buildingID: building.ID, updates: map[string]interface{}{"metadata": models.BuildingMetadata{"updated": "metadata"}}, expectError: false},
		{name: "Invalid building ID", buildingID: 99999, updates: map[string]interface{}{"building_name": "Should Fail"}, expectError: true},
		{name: "Empty updates", buildingID: building.ID, updates: map[string]interface{}{}, expectError: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.buildingID, tt.updates)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingRepository_SoftDelete(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	// Building without units
	building := &models.Building{PropertyID: propertyID, BuildingName: "Test Building", BuildingCode: "TB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, ActiveStatus: true}
	if err := repo.Create(building); err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	// Building with active unit
	buildingWithUnits := &models.Building{PropertyID: propertyID, BuildingName: "Building With Units", BuildingCode: "BWU001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, ActiveStatus: true}
	if err := repo.Create(buildingWithUnits); err != nil {
		t.Fatalf("Failed to create building with units: %v", err)
	}
	// Add an active unit to the second building
	_, err := db.Exec(`INSERT INTO units (property_id, building_id, unit_number, unit_type, active) VALUES ($1, $2, $3, $4, $5)`, propertyID, buildingWithUnits.ID, "U001", "Apartment", true)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}
	tests := []struct {
		name        string
		buildingID  int
		expectError bool
	}{
		{name: "Valid soft delete", buildingID: building.ID, expectError: false},
		{name: "Cannot delete building with active units", buildingID: buildingWithUnits.ID, expectError: true},
		{name: "Invalid building ID", buildingID: 99999, expectError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SoftDelete(tt.buildingID)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingRepository_BulkCreate(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	buildings := []*models.Building{{PropertyID: propertyID, BuildingName: "Bulk Building 1", BuildingCode: "BB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 3, ActiveStatus: true}, {PropertyID: propertyID, BuildingName: "Bulk Building 2", BuildingCode: "BB002", BuildingType: models.BuildingTypeCommercial, TotalFloors: 2, ActiveStatus: true}}
	if err := repo.BulkCreate(buildings); err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	for _, b := range buildings {
		if b.ID == 0 {
			t.Errorf("Expected building ID to be set")
		}
	}
	// Empty slice should not error
	if err := repo.BulkCreate([]*models.Building{}); err != nil {
		t.Errorf("Expected no error for empty slice but got: %v", err)
	}
}

func TestBuildingRepository_Search(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	// Create test buildings
	buildings := []*models.Building{{PropertyID: propertyID, BuildingName: "Residential Tower", BuildingCode: "RT001", BuildingType: models.BuildingTypeResidential, TotalFloors: 10, HasElevator: true, ActiveStatus: true}, {PropertyID: propertyID, BuildingName: "Commercial Block", BuildingCode: "CB001", BuildingType: models.BuildingTypeCommercial, TotalFloors: 3, HasElevator: false, ActiveStatus: true}, {PropertyID: propertyID, BuildingName: "Inactive Building", BuildingCode: "IB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 2, ActiveStatus: false}}
	for _, b := range buildings {
		if err := repo.Create(b); err != nil {
			t.Fatalf("Failed to create test building: %v", err)
		}
	}
	tests := []struct {
		name          string
		filters       *models.BuildingSearchFilters
		expectedCount int
	}{
		{name: "No filters", filters: &models.BuildingSearchFilters{Limit: 10}, expectedCount: 3},
		{name: "Filter by property ID", filters: &models.BuildingSearchFilters{PropertyID: &propertyID, Limit: 10}, expectedCount: 3},
		{name: "Filter by building type", filters: &models.BuildingSearchFilters{BuildingType: func() *models.BuildingType { bt := models.BuildingTypeResidential; return &bt }(), Limit: 10}, expectedCount: 2},
		{name: "Filter by active status", filters: &models.BuildingSearchFilters{ActiveStatus: func() *bool { b := true; return &b }(), Limit: 10}, expectedCount: 2},
		{name: "Filter by elevator", filters: &models.BuildingSearchFilters{HasElevator: func() *bool { b := true; return &b }(), Limit: 10}, expectedCount: 1},
		{name: "Filter by floor range", filters: &models.BuildingSearchFilters{MinFloors: func() *int { f := 3; return &f }(), MaxFloors: func() *int { f := 10; return &f }(), Limit: 10}, expectedCount: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.Search(tt.filters)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d buildings, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestBuildingRepository_GetWithStats(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()
	propertyID := createTestProperty(t, db)
	building := &models.Building{PropertyID: propertyID, BuildingName: "Stats Building", BuildingCode: "SB001", BuildingType: models.BuildingTypeResidential, TotalFloors: 5, ActiveStatus: true}
	if err := repo.Create(building); err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}
	// Create test units and related data
	for i := 1; i <= 3; i++ {
		var unitID int
		err := db.QueryRow(`INSERT INTO units (property_id, building_id, unit_number, unit_type, active) VALUES ($1, $2, $3, $4, $5) RETURNING id`, propertyID, building.ID, fmt.Sprintf("U%03d", i), "Apartment", true).Scan(&unitID)
		if err != nil {
			t.Fatalf("Failed to create test unit: %v", err)
		}
		if i <= 2 { // first two units have leases
			_, err = db.Exec(`INSERT INTO leases (unit_id, tenant_name, monthly_rent, active) VALUES ($1, $2, $3, $4)`, unitID, fmt.Sprintf("Tenant %d", i), 1000.00, true)
			if err != nil {
				t.Fatalf("Failed to create test lease: %v", err)
			}
			_, err = db.Exec(`INSERT INTO payments (unit_id, amount_paid, status) VALUES ($1, $2, $3)`, unitID, 1000.00, "Paid")
			if err != nil {
				t.Fatalf("Failed to create test payment: %v", err)
			}
		}
	}
	result, err := repo.GetWithStats(building.ID)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected building with stats but got nil")
	}
	if result.UnitCount != 3 {
		t.Errorf("Expected 3 units, got %d", result.UnitCount)
	}
	if result.OccupiedUnits != 2 {
		t.Errorf("Expected 2 occupied units, got %d", result.OccupiedUnits)
	}
	if result.TotalRevenue != 2000.00 {
		t.Errorf("Expected total revenue 2000.00, got %f", result.TotalRevenue)
	}
}
