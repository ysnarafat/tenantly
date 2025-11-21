package repositories

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"

	_ "github.com/lib/pq"
)

func setupBuildingRepositoryTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/tenantly_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	// Create test tables
	createBuildingRepositoryTestTables(t, db)

	cleanup := func() {
		dropBuildingRepositoryTestTables(t, db)
		db.Close()
	}

	return db, cleanup
}

func createBuildingRepositoryTestTables(t *testing.T, db *sql.DB) {
	// Create building_type_enum
	_, err := db.Exec(`CREATE TYPE building_type_enum AS ENUM ('Residential', 'Commercial', 'Mixed')`)
	if err != nil {
		// Type might already exist, ignore error
	}

	// Create properties table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS properties (
			id SERIAL PRIMARY KEY,
			property_name VARCHAR(200) NOT NULL,
			property_code VARCHAR(50) UNIQUE NOT NULL,
			address TEXT NOT NULL,
			city VARCHAR(100),
			postal_code VARCHAR(20),
			property_type VARCHAR(50) NOT NULL CHECK (property_type IN ('Residential', 'Commercial', 'Mixed')),
			total_buildings INTEGER DEFAULT 1,
			metadata JSONB,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create properties table: %v", err)
	}

	// Create buildings table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS buildings (
			id SERIAL PRIMARY KEY,
			property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
			building_name VARCHAR(100) NOT NULL,
			building_code VARCHAR(50) NOT NULL,
			building_type building_type_enum NOT NULL,
			total_floors INTEGER DEFAULT 1 CHECK (total_floors > 0),
			has_elevator BOOLEAN DEFAULT FALSE,
			construction_year INTEGER CHECK (construction_year >= 1800 AND construction_year <= EXTRACT(YEAR FROM CURRENT_DATE) + 5),
			metadata JSONB DEFAULT '{}',
			active_status BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			CONSTRAINT unique_building_code_per_property UNIQUE (property_id, building_code)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create buildings table: %v", err)
	}

	// Create units table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS units (
			id SERIAL PRIMARY KEY,
			property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
			building_id INTEGER REFERENCES buildings(id) ON DELETE CASCADE,
			unit_number VARCHAR(50) NOT NULL,
			floor INTEGER,
			section VARCHAR(50),
			unit_type VARCHAR(50) NOT NULL,
			area DECIMAL(10,2),
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create units table: %v", err)
	}

	// Create leases table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS leases (
			id SERIAL PRIMARY KEY,
			unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
			tenant_name VARCHAR(200) NOT NULL,
			monthly_rent DECIMAL(10,2) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create leases table: %v", err)
	}

	// Create payments table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			unit_id INTEGER NOT NULL REFERENCES units(id) ON DELETE CASCADE,
			amount_paid DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) DEFAULT 'Paid',
			payment_date TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create payments table: %v", err)
	}
}

func dropBuildingRepositoryTestTables(t *testing.T, db *sql.DB) {
	tables := []string{"payments", "leases", "units", "buildings", "properties"}
	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			t.Logf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	// Drop enum type
	_, err := db.Exec("DROP TYPE IF EXISTS building_type_enum CASCADE")
	if err != nil {
		t.Logf("Warning: Failed to drop enum type: %v", err)
	}
}

func setupBuildingRepository(t *testing.T) (*BuildingRepository, *sql.DB, func()) {
	db, cleanup := setupBuildingRepositoryTestDB(t)
	repo := NewBuildingRepository(db)
	return repo, db, cleanup
}

// Simple metadata validator for testing purposes
type testMetadataValidator struct{}

func (v *testMetadataValidator) ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error {
	if metadata == nil {
		return nil // Empty metadata is allowed
	}

	switch buildingType {
	case models.BuildingTypeResidential:
		return v.validateResidentialMetadata(metadata)
	case models.BuildingTypeCommercial:
		return v.validateCommercialMetadata(metadata)
	case models.BuildingTypeMixed:
		return v.validateMixedMetadata(metadata)
	default:
		return fmt.Errorf("invalid building type: %s", buildingType)
	}
}

func (v *testMetadataValidator) validateResidentialMetadata(metadata models.BuildingMetadata) error {
	// Basic validation for residential metadata
	if amenities, exists := metadata["amenities"]; exists {
		if amenitiesList, ok := amenities.([]string); ok {
			validAmenities := map[string]bool{
				"gym": true, "swimming_pool": true, "playground": true, "community_hall": true,
				"rooftop_garden": true, "library": true, "prayer_room": true,
			}
			for _, amenity := range amenitiesList {
				if !validAmenities[amenity] {
					return fmt.Errorf("invalid amenity: %s", amenity)
				}
			}
		}
	}

	if securityType, exists := metadata["security_type"]; exists {
		if securityTypeStr, ok := securityType.(string); ok {
			validTypes := map[string]bool{"24_hour_guard": true, "cctv_only": true, "card_access": true, "basic": true}
			if !validTypes[securityTypeStr] {
				return fmt.Errorf("invalid security_type: %s", securityTypeStr)
			}
		}
	}

	// Validate maintenance_staff_count
	if staffCount, exists := metadata["maintenance_staff_count"]; exists {
		var count int
		switch v := staffCount.(type) {
		case float64:
			count = int(v)
		case int:
			count = v
		default:
			return fmt.Errorf("maintenance_staff_count must be an integer")
		}

		if count < 0 {
			return fmt.Errorf("maintenance_staff_count cannot be negative")
		}
	}

	return nil
}

func (v *testMetadataValidator) validateCommercialMetadata(metadata models.BuildingMetadata) error {
	// Basic validation for commercial metadata
	if businessHours, exists := metadata["business_hours"]; exists {
		if hoursMap, ok := businessHours.(map[string]interface{}); ok {
			for key, value := range hoursMap {
				if valueStr, ok := value.(string); ok {
					// Simple time format validation
					if len(valueStr) != 11 || valueStr[2] != ':' || valueStr[5] != '-' || valueStr[8] != ':' {
						return fmt.Errorf("invalid business_hours.%s format", key)
					}
				}
			}
		}
	}

	if securitySystem, exists := metadata["security_system"]; exists {
		if systemMap, ok := securitySystem.(map[string]interface{}); ok {
			if systemType, exists := systemMap["type"]; exists {
				if typeStr, ok := systemType.(string); ok {
					validTypes := map[string]bool{"basic_cctv": true, "advanced_cctv": true, "full_security": true}
					if !validTypes[typeStr] {
						return fmt.Errorf("invalid security_system.type: %s", typeStr)
					}
				}
			}
		}
	}

	return nil
}

func (v *testMetadataValidator) validateMixedMetadata(metadata models.BuildingMetadata) error {
	// Basic validation for mixed metadata
	if residentialSection, exists := metadata["residential_section"]; exists {
		if sectionMap, ok := residentialSection.(map[string]interface{}); ok {
			if floors, exists := sectionMap["floors"]; exists {
				if floorsStr, ok := floors.(string); ok {
					// Simple floor range validation (e.g., "1-5")
					if err := v.validateFloorRange(floorsStr); err != nil {
						return fmt.Errorf("invalid residential_section.floors: %w", err)
					}
				}
			}
		}
	}

	if commercialSection, exists := metadata["commercial_section"]; exists {
		if sectionMap, ok := commercialSection.(map[string]interface{}); ok {
			if floors, exists := sectionMap["floors"]; exists {
				if floorsStr, ok := floors.(string); ok {
					// Simple floor range validation (e.g., "1-5")
					if err := v.validateFloorRange(floorsStr); err != nil {
						return fmt.Errorf("invalid commercial_section.floors: %w", err)
					}
				}
			}
		}
	}

	return nil
}

func createTestProperty(t *testing.T, db *sql.DB) int {
	var propertyID int
	err := db.QueryRow(`
		INSERT INTO properties (property_name, property_code, address, property_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		"Test Property", "TEST001", "123 Test Street", "Commercial").Scan(&propertyID)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}
	return propertyID
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
			name: "Valid residential building creation",
			building: &models.Building{
				PropertyID:       propertyID,
				BuildingName:     "Residential Tower A",
				BuildingCode:     "RTA001",
				BuildingType:     models.BuildingTypeResidential,
				TotalFloors:      10,
				HasElevator:      true,
				ConstructionYear: func() *int { year := 2020; return &year }(),
				Metadata:         models.BuildingMetadata{"amenities": []string{"gym", "pool"}},
				ActiveStatus:     true,
			},
			expectError: false,
		},
		{
			name: "Valid commercial building creation",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Commercial Block B",
				BuildingCode: "CBB001",
				BuildingType: models.BuildingTypeCommercial,
				TotalFloors:  5,
				HasElevator:  false,
				Metadata:     models.BuildingMetadata{"parking_spaces": 50},
				ActiveStatus: true,
			},
			expectError: false,
		},
		{
			name: "Duplicate building code should fail",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Duplicate Building",
				BuildingCode: "RTA001", // Same as first test
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
				ActiveStatus: true,
			},
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

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
		Metadata:     models.BuildingMetadata{"test": "value"},
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	tests := []struct {
		name        string
		buildingID  int
		expectError bool
	}{
		{
			name:        "Valid building ID",
			buildingID:  building.ID,
			expectError: false,
		},
		{
			name:        "Invalid building ID",
			buildingID:  99999,
			expectError: true,
		},
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
					t.Errorf("Expected building but got nil")
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

	// Create test buildings
	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "Building A",
			BuildingCode: "BA001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  3,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Building B",
			BuildingCode: "BB001",
			BuildingType: models.BuildingTypeCommercial,
			TotalFloors:  2,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Inactive Building",
			BuildingCode: "IB001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  1,
			ActiveStatus: false,
		},
	}

	for _, building := range buildings {
		err := repo.Create(building)
		if err != nil {
			t.Fatalf("Failed to create test building: %v", err)
		}
	}

	result, err := repo.GetByPropertyID(propertyID)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	// Should only return active buildings
	if len(result) != 2 {
		t.Errorf("Expected 2 active buildings, got %d", len(result))
	}

	// Test with non-existent property
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

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	tests := []struct {
		name         string
		propertyID   int
		buildingCode string
		expectError  bool
	}{
		{
			name:         "Valid property and code",
			propertyID:   propertyID,
			buildingCode: "TB001",
			expectError:  false,
		},
		{
			name:         "Invalid property ID",
			propertyID:   99999,
			buildingCode: "TB001",
			expectError:  true,
		},
		{
			name:         "Invalid building code",
			propertyID:   propertyID,
			buildingCode: "INVALID",
			expectError:  true,
		},
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
					t.Errorf("Expected building but got nil")
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

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Original Building",
		BuildingCode: "OB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		HasElevator:  false,
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	tests := []struct {
		name        string
		buildingID  int
		updates     map[string]interface{}
		expectError bool
	}{
		{
			name:       "Valid update",
			buildingID: building.ID,
			updates: map[string]interface{}{
				"building_name": "Updated Building",
				"total_floors":  5,
				"has_elevator":  true,
			},
			expectError: false,
		},
		{
			name:       "Update metadata",
			buildingID: building.ID,
			updates: map[string]interface{}{
				"metadata": models.BuildingMetadata{"updated": "metadata"},
			},
			expectError: false,
		},
		{
			name:       "Invalid building ID",
			buildingID: 99999,
			updates: map[string]interface{}{
				"building_name": "Should Fail",
			},
			expectError: true,
		},
		{
			name:        "Empty updates",
			buildingID:  building.ID,
			updates:     map[string]interface{}{},
			expectError: false,
		},
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

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// Create another building with active units
	buildingWithUnits := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Building With Units",
		BuildingCode: "BWU001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	err = repo.Create(buildingWithUnits)
	if err != nil {
		t.Fatalf("Failed to create building with units: %v", err)
	}

	// Add an active unit to the second building
	_, err = db.Exec(`
		INSERT INTO units (property_id, building_id, unit_number, unit_type, active)
		VALUES ($1, $2, $3, $4, $5)`,
		propertyID, buildingWithUnits.ID, "U001", "Apartment", true)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	tests := []struct {
		name        string
		buildingID  int
		expectError bool
	}{
		{
			name:        "Valid soft delete",
			buildingID:  building.ID,
			expectError: false,
		},
		{
			name:        "Cannot delete building with active units",
			buildingID:  buildingWithUnits.ID,
			expectError: true,
		},
		{
			name:        "Invalid building ID",
			buildingID:  99999,
			expectError: true,
		},
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

	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "Bulk Building 1",
			BuildingCode: "BB001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  3,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Bulk Building 2",
			BuildingCode: "BB002",
			BuildingType: models.BuildingTypeCommercial,
			TotalFloors:  2,
			ActiveStatus: true,
		},
	}

	err := repo.BulkCreate(buildings)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Verify buildings were created
	for _, building := range buildings {
		if building.ID == 0 {
			t.Errorf("Expected building ID to be set")
		}
	}

	// Test empty slice
	err = repo.BulkCreate([]*models.Building{})
	if err != nil {
		t.Errorf("Expected no error for empty slice but got: %v", err)
	}
}

func TestBuildingRepository_Search(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	propertyID := createTestProperty(t, db)

	// Create test buildings
	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "Residential Tower",
			BuildingCode: "RT001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  10,
			HasElevator:  true,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Commercial Block",
			BuildingCode: "CB001",
			BuildingType: models.BuildingTypeCommercial,
			TotalFloors:  3,
			HasElevator:  false,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Inactive Building",
			BuildingCode: "IB001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  2,
			ActiveStatus: false,
		},
	}

	for _, building := range buildings {
		err := repo.Create(building)
		if err != nil {
			t.Fatalf("Failed to create test building: %v", err)
		}
	}

	tests := []struct {
		name          string
		filters       *models.BuildingSearchFilters
		expectedCount int
	}{
		{
			name:          "No filters",
			filters:       &models.BuildingSearchFilters{Limit: 10},
			expectedCount: 3,
		},
		{
			name: "Filter by property ID",
			filters: &models.BuildingSearchFilters{
				PropertyID: &propertyID,
				Limit:      10,
			},
			expectedCount: 3,
		},
		{
			name: "Filter by building type",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeResidential; return &bt }(),
				Limit:        10,
			},
			expectedCount: 2,
		},
		{
			name: "Filter by active status",
			filters: &models.BuildingSearchFilters{
				ActiveStatus: func() *bool { b := true; return &b }(),
				Limit:        10,
			},
			expectedCount: 2,
		},
		{
			name: "Filter by elevator",
			filters: &models.BuildingSearchFilters{
				HasElevator: func() *bool { b := true; return &b }(),
				Limit:       10,
			},
			expectedCount: 1,
		},
		{
			name: "Filter by floor range",
			filters: &models.BuildingSearchFilters{
				MinFloors: func() *int { f := 3; return &f }(),
				MaxFloors: func() *int { f := 10; return &f }(),
				Limit:     10,
			},
			expectedCount: 2,
		},
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

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Stats Building",
		BuildingCode: "SB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// Create test units
	for i := 1; i <= 3; i++ {
		var unitID int
		err = db.QueryRow(`
			INSERT INTO units (property_id, building_id, unit_number, unit_type, active)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			propertyID, building.ID, fmt.Sprintf("U%03d", i), "Apartment", true).Scan(&unitID)
		if err != nil {
			t.Fatalf("Failed to create test unit: %v", err)
		}

		// Create lease for first two units
		if i <= 2 {
			_, err = db.Exec(`
				INSERT INTO leases (unit_id, tenant_name, monthly_rent, active)
				VALUES ($1, $2, $3, $4)`,
				unitID, fmt.Sprintf("Tenant %d", i), 1000.00, true)
			if err != nil {
				t.Fatalf("Failed to create test lease: %v", err)
			}

			// Create payment
			_, err = db.Exec(`
				INSERT INTO payments (unit_id, amount_paid, status)
				VALUES ($1, $2, $3)`,
				unitID, 1000.00, "Paid")
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

	expectedOccupancyRate := float64(2) / float64(3) * 100
	if result.OccupancyRate != expectedOccupancyRate {
		t.Errorf("Expected occupancy rate %f, got %f", expectedOccupancyRate, result.OccupancyRate)
	}

	// Test with non-existent building
	_, err = repo.GetWithStats(99999)
	if err == nil {
		t.Errorf("Expected error for non-existent building")
	}
}

func TestBuildingRepository_GetAnalytics(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	propertyID := createTestProperty(t, db)

	// Create a test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Analytics Building",
		BuildingCode: "AB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		ActiveStatus: true,
	}
	err := repo.Create(building)
	if err != nil {
		t.Fatalf("Failed to create test building: %v", err)
	}

	// Create test units with area
	for i := 1; i <= 4; i++ {
		var unitID int
		err = db.QueryRow(`
			INSERT INTO units (property_id, building_id, unit_number, unit_type, area, active)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`,
			propertyID, building.ID, fmt.Sprintf("U%03d", i), "Apartment", 100.0, true).Scan(&unitID)
		if err != nil {
			t.Fatalf("Failed to create test unit: %v", err)
		}

		// Create lease for first three units
		if i <= 3 {
			_, err = db.Exec(`
				INSERT INTO leases (unit_id, tenant_name, monthly_rent, active)
				VALUES ($1, $2, $3, $4)`,
				unitID, fmt.Sprintf("Tenant %d", i), float64(1000+i*100), true)
			if err != nil {
				t.Fatalf("Failed to create test lease: %v", err)
			}

			// Create monthly payment
			_, err = db.Exec(`
				INSERT INTO payments (unit_id, amount_paid, status, payment_date)
				VALUES ($1, $2, $3, $4)`,
				unitID, float64(1000+i*100), "Paid", time.Now())
			if err != nil {
				t.Fatalf("Failed to create test payment: %v", err)
			}
		}
	}

	result, err := repo.GetAnalytics(building.ID)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if result == nil {
		t.Fatalf("Expected building analytics but got nil")
	}

	if result.BuildingID != building.ID {
		t.Errorf("Expected building ID %d, got %d", building.ID, result.BuildingID)
	}

	if result.UnitCount != 4 {
		t.Errorf("Expected 4 units, got %d", result.UnitCount)
	}

	if result.OccupiedUnits != 3 {
		t.Errorf("Expected 3 occupied units, got %d", result.OccupiedUnits)
	}

	if result.VacantUnits != 1 {
		t.Errorf("Expected 1 vacant unit, got %d", result.VacantUnits)
	}

	if result.TotalArea != 400.0 {
		t.Errorf("Expected total area 400.0, got %f", result.TotalArea)
	}

	expectedOccupancyRate := float64(3) / float64(4) * 100
	if result.OccupancyRate != expectedOccupancyRate {
		t.Errorf("Expected occupancy rate %f, got %f", expectedOccupancyRate, result.OccupancyRate)
	}

	// Test with non-existent building
	_, err = repo.GetAnalytics(99999)
	if err == nil {
		t.Errorf("Expected error for non-existent building")
	}
}

// TestBuildingRepository_CreateWithMetadataValidation tests building creation with various building types and metadata
func TestBuildingRepository_CreateWithMetadataValidation(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	propertyID := createTestProperty(t, db)
	validator := &testMetadataValidator{}

	tests := []struct {
		name         string
		building     *models.Building
		expectError  bool
		errorMessage string
	}{
		{
			name: "Residential building with valid amenities metadata",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Residential Tower",
				BuildingCode: "RT001",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  10,
				HasElevator:  true,
				Metadata: models.BuildingMetadata{
					"amenities":               []string{"gym", "swimming_pool", "playground"},
					"security_type":           "24_hour_guard",
					"maintenance_staff_count": 5,
					"parking_spaces": map[string]interface{}{
						"total":   100,
						"covered": 60,
						"visitor": 20,
					},
					"utilities": map[string]interface{}{
						"backup_generator": true,
						"water_supply":     "24_hour",
						"internet_ready":   true,
					},
				},
				ActiveStatus: true,
			},
			expectError: false,
		},
		{
			name: "Commercial building with valid business metadata",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Commercial Plaza",
				BuildingCode: "CP001",
				BuildingType: models.BuildingTypeCommercial,
				TotalFloors:  5,
				HasElevator:  true,
				Metadata: models.BuildingMetadata{
					"parking_spaces": map[string]interface{}{
						"total":    200,
						"customer": 150,
						"staff":    50,
					},
					"loading_docks": 3,
					"security_system": map[string]interface{}{
						"type":           "advanced_cctv",
						"access_control": true,
						"fire_safety":    "sprinkler_system",
					},
					"business_hours": map[string]interface{}{
						"weekdays": "09:00-22:00",
						"weekends": "10:00-23:00",
						"holidays": "10:00-20:00",
					},
					"facilities": map[string]interface{}{
						"elevators":  4,
						"escalators": 2,
						"food_court": true,
						"atm":        true,
					},
				},
				ActiveStatus: true,
			},
			expectError: false,
		},
		{
			name: "Mixed building with valid combined metadata",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Mixed Use Complex",
				BuildingCode: "MUC001",
				BuildingType: models.BuildingTypeMixed,
				TotalFloors:  15,
				HasElevator:  true,
				Metadata: models.BuildingMetadata{
					"residential_section": map[string]interface{}{
						"floors":        "5-15",
						"amenities":     []string{"gym", "rooftop_garden"},
						"security_type": "card_access",
					},
					"commercial_section": map[string]interface{}{
						"floors":             "1-4",
						"business_hours":     "09:00-22:00",
						"parking_allocation": 80,
					},
					"shared_facilities": map[string]interface{}{
						"elevators":        6,
						"parking_total":    200,
						"backup_generator": true,
					},
				},
				ActiveStatus: true,
			},
			expectError: false,
		},
		{
			name: "Building with empty metadata (should be allowed)",
			building: &models.Building{
				PropertyID:   propertyID,
				BuildingName: "Simple Building",
				BuildingCode: "SB001",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
				Metadata:     models.BuildingMetadata{},
				ActiveStatus: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate metadata before creation
			if tt.building.Metadata != nil && len(tt.building.Metadata) > 0 {
				err := validator.ValidateMetadata(tt.building.BuildingType, tt.building.Metadata)
				if err != nil {
					t.Errorf("Metadata validation failed: %v", err)
					return
				}
			}

			err := repo.Create(tt.building)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.building.ID == 0 {
					t.Errorf("Expected building ID to be set")
				}
			}
		})
	}
}

// TestBuildingRepository_BuildingCodeUniqueness tests building code uniqueness validation within properties
func TestBuildingRepository_BuildingCodeUniqueness(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	// Create two different properties
	property1ID := createTestProperty(t, db)

	var property2ID int
	err := db.QueryRow(`
		INSERT INTO properties (property_name, property_code, address, property_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		"Test Property 2", "TEST002", "456 Test Avenue", "Residential").Scan(&property2ID)
	if err != nil {
		t.Fatalf("Failed to create second test property: %v", err)
	}

	tests := []struct {
		name        string
		buildings   []*models.Building
		expectError []bool
		description string
	}{
		{
			name: "Same building code in different properties should succeed",
			buildings: []*models.Building{
				{
					PropertyID:   property1ID,
					BuildingName: "Building A in Property 1",
					BuildingCode: "SAME_CODE",
					BuildingType: models.BuildingTypeResidential,
					TotalFloors:  3,
					ActiveStatus: true,
				},
				{
					PropertyID:   property2ID,
					BuildingName: "Building A in Property 2",
					BuildingCode: "SAME_CODE", // Same code but different property
					BuildingType: models.BuildingTypeCommercial,
					TotalFloors:  5,
					ActiveStatus: true,
				},
			},
			expectError: []bool{false, false},
			description: "Buildings with same code in different properties should be allowed",
		},
		{
			name: "Duplicate building code in same property should fail",
			buildings: []*models.Building{
				{
					PropertyID:   property1ID,
					BuildingName: "First Building",
					BuildingCode: "DUPLICATE",
					BuildingType: models.BuildingTypeResidential,
					TotalFloors:  3,
					ActiveStatus: true,
				},
				{
					PropertyID:   property1ID,
					BuildingName: "Second Building",
					BuildingCode: "DUPLICATE", // Same code and same property
					BuildingType: models.BuildingTypeCommercial,
					TotalFloors:  5,
					ActiveStatus: true,
				},
			},
			expectError: []bool{false, true},
			description: "Duplicate building codes in same property should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, building := range tt.buildings {
				err := repo.Create(building)

				if tt.expectError[i] {
					if err == nil {
						t.Errorf("Building %d: Expected error but got none. %s", i+1, tt.description)
					}
				} else {
					if err != nil {
						t.Errorf("Building %d: Expected no error but got: %v. %s", i+1, err, tt.description)
					}
				}
			}
		})
	}
}

// TestBuildingRepository_SoftDeleteConstraints tests soft delete functionality and constraint validation
func TestBuildingRepository_SoftDeleteConstraints(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	propertyID := createTestProperty(t, db)

	// Create buildings for testing different scenarios
	buildingWithoutUnits := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Empty Building",
		BuildingCode: "EB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	err := repo.Create(buildingWithoutUnits)
	if err != nil {
		t.Fatalf("Failed to create building without units: %v", err)
	}

	buildingWithInactiveUnits := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Building With Inactive Units",
		BuildingCode: "BWIU001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	err = repo.Create(buildingWithInactiveUnits)
	if err != nil {
		t.Fatalf("Failed to create building with inactive units: %v", err)
	}

	buildingWithActiveUnits := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Building With Active Units",
		BuildingCode: "BWAU001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	err = repo.Create(buildingWithActiveUnits)
	if err != nil {
		t.Fatalf("Failed to create building with active units: %v", err)
	}

	// Add inactive units to second building
	_, err = db.Exec(`
		INSERT INTO units (property_id, building_id, unit_number, unit_type, active)
		VALUES ($1, $2, $3, $4, $5)`,
		propertyID, buildingWithInactiveUnits.ID, "IU001", "Apartment", false)
	if err != nil {
		t.Fatalf("Failed to create inactive unit: %v", err)
	}

	// Add active units to third building
	_, err = db.Exec(`
		INSERT INTO units (property_id, building_id, unit_number, unit_type, active)
		VALUES ($1, $2, $3, $4, $5)`,
		propertyID, buildingWithActiveUnits.ID, "AU001", "Apartment", true)
	if err != nil {
		t.Fatalf("Failed to create active unit: %v", err)
	}

	tests := []struct {
		name        string
		buildingID  int
		expectError bool
		description string
	}{
		{
			name:        "Delete building without units should succeed",
			buildingID:  buildingWithoutUnits.ID,
			expectError: false,
			description: "Buildings without any units should be deletable",
		},
		{
			name:        "Delete building with only inactive units should succeed",
			buildingID:  buildingWithInactiveUnits.ID,
			expectError: false,
			description: "Buildings with only inactive units should be deletable",
		},
		{
			name:        "Delete building with active units should fail",
			buildingID:  buildingWithActiveUnits.ID,
			expectError: true,
			description: "Buildings with active units should not be deletable",
		},
		{
			name:        "Delete non-existent building should fail",
			buildingID:  99999,
			expectError: true,
			description: "Non-existent buildings should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SoftDelete(tt.buildingID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. %s", err, tt.description)
				}

				// Verify building is marked as inactive
				building, err := repo.GetByID(tt.buildingID)
				if err != nil {
					t.Errorf("Failed to retrieve building after soft delete: %v", err)
				} else if building.ActiveStatus {
					t.Errorf("Building should be marked as inactive after soft delete")
				}
			}
		})
	}
}

// TestBuildingRepository_BulkCreateErrorHandling tests bulk creation operations with error handling
func TestBuildingRepository_BulkCreateErrorHandling(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	propertyID := createTestProperty(t, db)

	tests := []struct {
		name        string
		buildings   []*models.Building
		expectError bool
		description string
	}{
		{
			name: "Bulk create valid buildings should succeed",
			buildings: []*models.Building{
				{
					PropertyID:   propertyID,
					BuildingName: "Bulk Building 1",
					BuildingCode: "BB001",
					BuildingType: models.BuildingTypeResidential,
					TotalFloors:  3,
					ActiveStatus: true,
				},
				{
					PropertyID:   propertyID,
					BuildingName: "Bulk Building 2",
					BuildingCode: "BB002",
					BuildingType: models.BuildingTypeCommercial,
					TotalFloors:  5,
					ActiveStatus: true,
				},
				{
					PropertyID:   propertyID,
					BuildingName: "Bulk Building 3",
					BuildingCode: "BB003",
					BuildingType: models.BuildingTypeMixed,
					TotalFloors:  8,
					ActiveStatus: true,
				},
			},
			expectError: false,
			description: "Valid buildings should be created successfully",
		},
		{
			name: "Bulk create with duplicate codes should fail and rollback",
			buildings: []*models.Building{
				{
					PropertyID:   propertyID,
					BuildingName: "Valid Building",
					BuildingCode: "VALID001",
					BuildingType: models.BuildingTypeResidential,
					TotalFloors:  3,
					ActiveStatus: true,
				},
				{
					PropertyID:   propertyID,
					BuildingName: "Duplicate Building 1",
					BuildingCode: "DUPLICATE",
					BuildingType: models.BuildingTypeCommercial,
					TotalFloors:  5,
					ActiveStatus: true,
				},
				{
					PropertyID:   propertyID,
					BuildingName: "Duplicate Building 2",
					BuildingCode: "DUPLICATE", // Duplicate code
					BuildingType: models.BuildingTypeMixed,
					TotalFloors:  8,
					ActiveStatus: true,
				},
			},
			expectError: true,
			description: "Duplicate codes should cause transaction rollback",
		},
		{
			name: "Bulk create with invalid property should fail",
			buildings: []*models.Building{
				{
					PropertyID:   99999, // Non-existent property
					BuildingName: "Invalid Property Building",
					BuildingCode: "IPB001",
					BuildingType: models.BuildingTypeResidential,
					TotalFloors:  3,
					ActiveStatus: true,
				},
			},
			expectError: true,
			description: "Invalid property ID should cause failure",
		},
		{
			name:        "Bulk create empty slice should succeed",
			buildings:   []*models.Building{},
			expectError: false,
			description: "Empty building slice should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Count buildings before operation
			var countBefore int
			err := db.QueryRow("SELECT COUNT(*) FROM buildings WHERE property_id = $1", propertyID).Scan(&countBefore)
			if err != nil {
				t.Fatalf("Failed to count buildings before operation: %v", err)
			}

			err = repo.BulkCreate(tt.buildings)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}

				// Verify rollback - count should be the same
				var countAfter int
				err = db.QueryRow("SELECT COUNT(*) FROM buildings WHERE property_id = $1", propertyID).Scan(&countAfter)
				if err != nil {
					t.Fatalf("Failed to count buildings after operation: %v", err)
				}

				if countAfter != countBefore {
					t.Errorf("Expected building count to remain %d after failed bulk create, got %d. Transaction should have rolled back.", countBefore, countAfter)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. %s", err, tt.description)
				}

				// Verify all buildings were created (if not empty)
				if len(tt.buildings) > 0 {
					for _, building := range tt.buildings {
						if building.ID == 0 {
							t.Errorf("Expected building ID to be set after bulk create")
						}
						if building.CreatedAt.IsZero() {
							t.Errorf("Expected CreatedAt to be set after bulk create")
						}
					}

					// Verify count increased correctly
					var countAfter int
					err = db.QueryRow("SELECT COUNT(*) FROM buildings WHERE property_id = $1", propertyID).Scan(&countAfter)
					if err != nil {
						t.Fatalf("Failed to count buildings after operation: %v", err)
					}

					expectedCount := countBefore + len(tt.buildings)
					if countAfter != expectedCount {
						t.Errorf("Expected building count to be %d after bulk create, got %d", expectedCount, countAfter)
					}
				}
			}
		})
	}
}

// TestBuildingRepository_MetadataValidationAllTypes tests metadata validation for all building types
func TestBuildingRepository_MetadataValidationAllTypes(t *testing.T) {
	validator := &testMetadataValidator{}

	tests := []struct {
		name         string
		buildingType models.BuildingType
		metadata     models.BuildingMetadata
		expectError  bool
		description  string
	}{
		// Residential metadata tests
		{
			name:         "Valid residential metadata with all fields",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities":               []string{"gym", "swimming_pool", "playground", "community_hall"},
				"security_type":           "24_hour_guard",
				"maintenance_staff_count": 5,
				"parking_spaces": map[string]interface{}{
					"total":   100,
					"covered": 60,
					"visitor": 20,
				},
				"utilities": map[string]interface{}{
					"backup_generator": true,
					"water_supply":     "24_hour",
					"internet_ready":   true,
				},
			},
			expectError: false,
			description: "All valid residential metadata fields should pass validation",
		},
		{
			name:         "Invalid residential amenities",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities": []string{"invalid_amenity", "gym"},
			},
			expectError: true,
			description: "Invalid amenities should fail validation",
		},
		{
			name:         "Invalid residential security type",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"security_type": "invalid_security",
			},
			expectError: true,
			description: "Invalid security type should fail validation",
		},
		{
			name:         "Invalid maintenance staff count (negative)",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"maintenance_staff_count": -1,
			},
			expectError: true,
			description: "Negative maintenance staff count should fail validation",
		},

		// Commercial metadata tests
		{
			name:         "Valid commercial metadata with all fields",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":    200,
					"customer": 150,
					"staff":    50,
				},
				"loading_docks": 3,
				"security_system": map[string]interface{}{
					"type":           "advanced_cctv",
					"access_control": true,
					"fire_safety":    "sprinkler_system",
				},
				"business_hours": map[string]interface{}{
					"weekdays": "09:00-22:00",
					"weekends": "10:00-23:00",
					"holidays": "10:00-20:00",
				},
				"facilities": map[string]interface{}{
					"elevators":  4,
					"escalators": 2,
					"food_court": true,
					"atm":        true,
				},
			},
			expectError: false,
			description: "All valid commercial metadata fields should pass validation",
		},
		{
			name:         "Invalid commercial business hours format",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"business_hours": map[string]interface{}{
					"weekdays": "invalid_format",
				},
			},
			expectError: true,
			description: "Invalid business hours format should fail validation",
		},
		{
			name:         "Invalid commercial security system type",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"security_system": map[string]interface{}{
					"type": "invalid_security_type",
				},
			},
			expectError: true,
			description: "Invalid security system type should fail validation",
		},

		// Mixed metadata tests
		{
			name:         "Valid mixed metadata with all sections",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"residential_section": map[string]interface{}{
					"floors":        "5-15",
					"amenities":     []string{"gym", "rooftop_garden"},
					"security_type": "card_access",
				},
				"commercial_section": map[string]interface{}{
					"floors":             "1-4",
					"business_hours":     "09:00-22:00",
					"parking_allocation": 80,
				},
				"shared_facilities": map[string]interface{}{
					"elevators":        6,
					"parking_total":    200,
					"backup_generator": true,
				},
			},
			expectError: false,
			description: "All valid mixed metadata sections should pass validation",
		},
		{
			name:         "Invalid mixed floor range format",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"residential_section": map[string]interface{}{
					"floors": "invalid_range",
				},
			},
			expectError: true,
			description: "Invalid floor range format should fail validation",
		},
		{
			name:         "Invalid mixed floor range (start >= end)",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"commercial_section": map[string]interface{}{
					"floors": "5-3", // Start floor >= end floor
				},
			},
			expectError: true,
			description: "Invalid floor range logic should fail validation",
		},

		// Edge cases
		{
			name:         "Empty metadata should be valid",
			buildingType: models.BuildingTypeResidential,
			metadata:     models.BuildingMetadata{},
			expectError:  false,
			description:  "Empty metadata should be allowed",
		},
		{
			name:         "Nil metadata should be valid",
			buildingType: models.BuildingTypeCommercial,
			metadata:     nil,
			expectError:  false,
			description:  "Nil metadata should be allowed",
		},
		{
			name:         "Invalid building type should fail",
			buildingType: "InvalidType",
			metadata:     models.BuildingMetadata{"test": "value"},
			expectError:  true,
			description:  "Invalid building type should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMetadata(tt.buildingType, tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v. %s", err, tt.description)
				}
			}
		})
	}
}

// TestBuildingRepository_SearchAdvancedFiltering tests search and filtering functionality with various criteria
func TestBuildingRepository_SearchAdvancedFiltering(t *testing.T) {
	repo, db, cleanup := setupBuildingRepository(t)
	defer cleanup()

	// Create multiple properties for testing
	property1ID := createTestProperty(t, db)

	var property2ID int
	err := db.QueryRow(`
		INSERT INTO properties (property_name, property_code, address, property_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		"Test Property 2", "TEST002", "456 Test Avenue", "Commercial").Scan(&property2ID)
	if err != nil {
		t.Fatalf("Failed to create second test property: %v", err)
	}

	// Create diverse buildings for testing filters
	testBuildings := []*models.Building{
		{
			PropertyID:       property1ID,
			BuildingName:     "Residential Tower A",
			BuildingCode:     "RTA001",
			BuildingType:     models.BuildingTypeResidential,
			TotalFloors:      15,
			HasElevator:      true,
			ConstructionYear: func() *int { year := 2020; return &year }(),
			ActiveStatus:     true,
		},
		{
			PropertyID:       property1ID,
			BuildingName:     "Commercial Block B",
			BuildingCode:     "CBB001",
			BuildingType:     models.BuildingTypeCommercial,
			TotalFloors:      8,
			HasElevator:      true,
			ConstructionYear: func() *int { year := 2018; return &year }(),
			ActiveStatus:     true,
		},
		{
			PropertyID:       property1ID,
			BuildingName:     "Mixed Use Building C",
			BuildingCode:     "MUC001",
			BuildingType:     models.BuildingTypeMixed,
			TotalFloors:      12,
			HasElevator:      true,
			ConstructionYear: func() *int { year := 2022; return &year }(),
			ActiveStatus:     false, // Inactive
		},
		{
			PropertyID:   property2ID,
			BuildingName: "Small Commercial D",
			BuildingCode: "SCD001",
			BuildingType: models.BuildingTypeCommercial,
			TotalFloors:  3,
			HasElevator:  false,
			ActiveStatus: true,
		},
		{
			PropertyID:   property2ID,
			BuildingName: "Residential Low Rise E",
			BuildingCode: "RLE001",
			BuildingType: models.BuildingTypeResidential,
			TotalFloors:  4,
			HasElevator:  false,
			ActiveStatus: true,
		},
	}

	// Create all test buildings
	for _, building := range testBuildings {
		err := repo.Create(building)
		if err != nil {
			t.Fatalf("Failed to create test building %s: %v", building.BuildingName, err)
		}
	}

	tests := []struct {
		name          string
		filters       *models.BuildingSearchFilters
		expectedCount int
		description   string
	}{
		{
			name:          "No filters - return all buildings",
			filters:       &models.BuildingSearchFilters{Limit: 10},
			expectedCount: 5,
			description:   "Should return all buildings when no filters applied",
		},
		{
			name: "Filter by property ID",
			filters: &models.BuildingSearchFilters{
				PropertyID: &property1ID,
				Limit:      10,
			},
			expectedCount: 3,
			description:   "Should return only buildings from property 1",
		},
		{
			name: "Filter by building type - Residential",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeResidential; return &bt }(),
				Limit:        10,
			},
			expectedCount: 2,
			description:   "Should return only residential buildings",
		},
		{
			name: "Filter by building type - Commercial",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeCommercial; return &bt }(),
				Limit:        10,
			},
			expectedCount: 2,
			description:   "Should return only commercial buildings",
		},
		{
			name: "Filter by building type - Mixed",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeMixed; return &bt }(),
				Limit:        10,
			},
			expectedCount: 1,
			description:   "Should return only mixed buildings",
		},
		{
			name: "Filter by active status - Active only",
			filters: &models.BuildingSearchFilters{
				ActiveStatus: func() *bool { b := true; return &b }(),
				Limit:        10,
			},
			expectedCount: 4,
			description:   "Should return only active buildings",
		},
		{
			name: "Filter by active status - Inactive only",
			filters: &models.BuildingSearchFilters{
				ActiveStatus: func() *bool { b := false; return &b }(),
				Limit:        10,
			},
			expectedCount: 1,
			description:   "Should return only inactive buildings",
		},
		{
			name: "Filter by elevator - Has elevator",
			filters: &models.BuildingSearchFilters{
				HasElevator: func() *bool { b := true; return &b }(),
				Limit:       10,
			},
			expectedCount: 3,
			description:   "Should return only buildings with elevators",
		},
		{
			name: "Filter by elevator - No elevator",
			filters: &models.BuildingSearchFilters{
				HasElevator: func() *bool { b := false; return &b }(),
				Limit:       10,
			},
			expectedCount: 2,
			description:   "Should return only buildings without elevators",
		},
		{
			name: "Filter by minimum floors",
			filters: &models.BuildingSearchFilters{
				MinFloors: func() *int { f := 10; return &f }(),
				Limit:     10,
			},
			expectedCount: 2,
			description:   "Should return buildings with 10 or more floors",
		},
		{
			name: "Filter by maximum floors",
			filters: &models.BuildingSearchFilters{
				MaxFloors: func() *int { f := 5; return &f }(),
				Limit:     10,
			},
			expectedCount: 2,
			description:   "Should return buildings with 5 or fewer floors",
		},
		{
			name: "Filter by floor range",
			filters: &models.BuildingSearchFilters{
				MinFloors: func() *int { f := 5; return &f }(),
				MaxFloors: func() *int { f := 10; return &f }(),
				Limit:     10,
			},
			expectedCount: 1,
			description:   "Should return buildings with 5-10 floors",
		},
		{
			name: "Combined filters - Property and type",
			filters: &models.BuildingSearchFilters{
				PropertyID:   &property1ID,
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeCommercial; return &bt }(),
				Limit:        10,
			},
			expectedCount: 1,
			description:   "Should return commercial buildings in property 1",
		},
		{
			name: "Combined filters - Type, elevator, and active status",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeResidential; return &bt }(),
				HasElevator:  func() *bool { b := true; return &b }(),
				ActiveStatus: func() *bool { b := true; return &b }(),
				Limit:        10,
			},
			expectedCount: 1,
			description:   "Should return active residential buildings with elevators",
		},
		{
			name: "Limit and offset test",
			filters: &models.BuildingSearchFilters{
				Limit:  2,
				Offset: 1,
			},
			expectedCount: 2,
			description:   "Should return 2 buildings starting from offset 1",
		},
		{
			name: "No matches filter",
			filters: &models.BuildingSearchFilters{
				BuildingType: func() *models.BuildingType { bt := models.BuildingTypeCommercial; return &bt }(),
				MinFloors:    func() *int { f := 20; return &f }(),
				Limit:        10,
			},
			expectedCount: 0,
			description:   "Should return no buildings when filters don't match any",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.Search(tt.filters)
			if err != nil {
				t.Errorf("Expected no error but got: %v. %s", err, tt.description)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d buildings, got %d. %s", tt.expectedCount, len(result), tt.description)
			}

			// Verify that returned buildings match the filters
			for _, building := range result {
				if tt.filters.PropertyID != nil && building.PropertyID != *tt.filters.PropertyID {
					t.Errorf("Building %d has wrong property ID: expected %d, got %d", building.ID, *tt.filters.PropertyID, building.PropertyID)
				}
				if tt.filters.BuildingType != nil && building.BuildingType != *tt.filters.BuildingType {
					t.Errorf("Building %d has wrong type: expected %s, got %s", building.ID, *tt.filters.BuildingType, building.BuildingType)
				}
				if tt.filters.ActiveStatus != nil && building.ActiveStatus != *tt.filters.ActiveStatus {
					t.Errorf("Building %d has wrong active status: expected %t, got %t", building.ID, *tt.filters.ActiveStatus, building.ActiveStatus)
				}
				if tt.filters.HasElevator != nil && building.HasElevator != *tt.filters.HasElevator {
					t.Errorf("Building %d has wrong elevator status: expected %t, got %t", building.ID, *tt.filters.HasElevator, building.HasElevator)
				}
				if tt.filters.MinFloors != nil && building.TotalFloors < *tt.filters.MinFloors {
					t.Errorf("Building %d has too few floors: expected >= %d, got %d", building.ID, *tt.filters.MinFloors, building.TotalFloors)
				}
				if tt.filters.MaxFloors != nil && building.TotalFloors > *tt.filters.MaxFloors {
					t.Errorf("Building %d has too many floors: expected <= %d, got %d", building.ID, *tt.filters.MaxFloors, building.TotalFloors)
				}
			}
		})
	}
}

func (v *testMetadataValidator) validateFloorRange(floorRange string) error {
	// Simple floor range validation (e.g., "1-5")
	if len(floorRange) < 3 {
		return fmt.Errorf("invalid floor range format, expected 'start-end'")
	}

	dashIndex := -1
	for i, char := range floorRange {
		if char == '-' {
			dashIndex = i
			break
		}
	}

	if dashIndex == -1 {
		return fmt.Errorf("invalid floor range format, expected 'start-end'")
	}

	startStr := floorRange[:dashIndex]
	endStr := floorRange[dashIndex+1:]

	if startStr == "" || endStr == "" {
		return fmt.Errorf("invalid floor range format, expected 'start-end'")
	}

	// Convert to integers for proper comparison
	var startFloor, endFloor int
	if _, err := fmt.Sscanf(startStr, "%d", &startFloor); err != nil {
		return fmt.Errorf("invalid start floor number")
	}
	if _, err := fmt.Sscanf(endStr, "%d", &endFloor); err != nil {
		return fmt.Errorf("invalid end floor number")
	}

	// For the test case "5-3", this should fail
	if startFloor >= endFloor {
		return fmt.Errorf("start floor must be less than end floor")
	}

	return nil
}
