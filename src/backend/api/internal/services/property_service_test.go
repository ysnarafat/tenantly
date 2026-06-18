package services

import (
	"database/sql"
	"testing"

	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"

	_ "github.com/lib/pq"
)

func setupPropertyTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/tenantly_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	// Create test tables
	createTestTables(t, db)

	cleanup := func() {
		dropTestTables(t, db)
		db.Close()
	}

	return db, cleanup
}

func createTestTables(t *testing.T, db *sql.DB) {
	// Create properties table for testing
	_, err := db.Exec(`
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
			building_code VARCHAR(20) NOT NULL,
			building_type VARCHAR(50) NOT NULL,
			total_floors INTEGER,
			has_elevator BOOLEAN DEFAULT false,
			construction_year INTEGER,
			metadata JSONB,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			UNIQUE(property_id, building_code)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create buildings table: %v", err)
	}

	// Create units table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS units (
			id SERIAL PRIMARY KEY,
			building_id INTEGER NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
			property_id INTEGER NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
			unit_number VARCHAR(50) NOT NULL,
			unit_name VARCHAR(100),
			floor INTEGER,
			section VARCHAR(50),
			unit_type VARCHAR(50) NOT NULL,
			monthly_rent DECIMAL(10,2) NOT NULL,
			metadata JSONB,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			updated_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC'),
			UNIQUE(building_id, unit_number)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create units table: %v", err)
	}

	// Create audit_log table for testing
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			action VARCHAR(50) NOT NULL,
			table_name VARCHAR(50) NOT NULL,
			record_id INTEGER,
			old_values JSONB,
			new_values JSONB,
			created_at TIMESTAMP DEFAULT (NOW() AT TIME ZONE 'UTC')
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create audit_log table: %v", err)
	}
}

func dropTestTables(t *testing.T, db *sql.DB) {
	tables := []string{"units", "buildings", "properties", "audit_log"}
	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			t.Logf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
}

func TestPropertyService_CreateProperty(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	tests := []struct {
		name        string
		request     *models.CreatePropertyRequest
		userID      int
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid property creation",
			request: &models.CreatePropertyRequest{
				PropertyName: "Test Property",
				PropertyCode: "TEST001",
				Address:      "123 Test Street",
				City:         "Test City",
				PostalCode:   "12345",
				PropertyType: models.PropertyTypeCommercial,
				Metadata:     models.PropertyMetadata{"test": "value"},
			},
			userID:      1,
			expectError: false,
		},
		{
			name: "Invalid property type",
			request: &models.CreatePropertyRequest{
				PropertyName: "Test Property 2",
				PropertyCode: "TEST002",
				Address:      "456 Test Street",
				PropertyType: "InvalidType",
			},
			userID:      1,
			expectError: true,
			errorMsg:    "invalid property type",
		},
		{
			name: "Duplicate property name",
			request: &models.CreatePropertyRequest{
				PropertyName: "Test Property", // Same as first test
				PropertyCode: "TEST003",
				Address:      "789 Test Street",
				PropertyType: models.PropertyTypeResidential,
			},
			userID:      1,
			expectError: true,
			errorMsg:    "property name already exists",
		},
		{
			name: "Duplicate property code",
			request: &models.CreatePropertyRequest{
				PropertyName: "Test Property 3",
				PropertyCode: "TEST001", // Same as first test
				Address:      "789 Test Street",
				PropertyType: models.PropertyTypeResidential,
			},
			userID:      1,
			expectError: true,
			errorMsg:    "property code already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.CreateProperty(tt.request, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if property == nil {
				t.Errorf("Expected property but got nil")
				return
			}

			if property.PropertyName != tt.request.PropertyName {
				t.Errorf("Expected property name '%s', got '%s'", tt.request.PropertyName, property.PropertyName)
			}

			if property.PropertyCode != tt.request.PropertyCode {
				t.Errorf("Expected property code '%s', got '%s'", tt.request.PropertyCode, property.PropertyCode)
			}

			if property.PropertyType != tt.request.PropertyType {
				t.Errorf("Expected property type '%s', got '%s'", tt.request.PropertyType, property.PropertyType)
			}
		})
	}
}

func TestPropertyService_GetProperty(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	// Create a test property
	createReq := &models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	tests := []struct {
		name        string
		propertyID  int
		expectError bool
	}{
		{
			name:        "Valid property ID",
			propertyID:  createdProperty.ID,
			expectError: false,
		},
		{
			name:        "Invalid property ID",
			propertyID:  99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.GetProperty(tt.propertyID, 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if property == nil {
				t.Errorf("Expected property but got nil")
				return
			}

			if property.ID != tt.propertyID {
				t.Errorf("Expected property ID %d, got %d", tt.propertyID, property.ID)
			}
		})
	}
}

func TestPropertyService_ListProperties(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	// Create test properties
	properties := []*models.CreatePropertyRequest{
		{
			PropertyName: "Commercial Property",
			PropertyCode: "COM001",
			Address:      "123 Commercial St",
			PropertyType: models.PropertyTypeCommercial,
		},
		{
			PropertyName: "Residential Property",
			PropertyCode: "RES001",
			Address:      "456 Residential Ave",
			PropertyType: models.PropertyTypeResidential,
		},
		{
			PropertyName: "Mixed Property",
			PropertyCode: "MIX001",
			Address:      "789 Mixed Blvd",
			PropertyType: models.PropertyTypeMixed,
		},
	}

	for _, prop := range properties {
		_, err := service.CreateProperty(prop, 1)
		if err != nil {
			t.Fatalf("Failed to create test property: %v", err)
		}
	}

	tests := []struct {
		name          string
		filters       map[string]interface{}
		page          int
		pageSize      int
		expectedCount int
		expectError   bool
	}{
		{
			name:          "List all properties",
			filters:       map[string]interface{}{},
			page:          1,
			pageSize:      10,
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "Filter by property type",
			filters:       map[string]interface{}{"property_type": "Commercial"},
			page:          1,
			pageSize:      10,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Search by name",
			filters:       map[string]interface{}{"search": "Commercial"},
			page:          1,
			pageSize:      10,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "Invalid property type filter",
			filters:       map[string]interface{}{"property_type": "InvalidType"},
			page:          1,
			pageSize:      10,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			properties, total, err := service.ListProperties(tt.filters, tt.page, tt.pageSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(properties) != tt.expectedCount {
				t.Errorf("Expected %d properties, got %d", tt.expectedCount, len(properties))
			}

			if total < tt.expectedCount {
				t.Errorf("Expected total >= %d, got %d", tt.expectedCount, total)
			}
		})
	}
}

func TestPropertyService_UpdateProperty(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	// Create a test property
	createReq := &models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	// Create another property for uniqueness testing
	createReq2 := &models.CreatePropertyRequest{
		PropertyName: "Another Property",
		PropertyCode: "TEST002",
		Address:      "456 Another Street",
		PropertyType: models.PropertyTypeResidential,
	}
	_, err = service.CreateProperty(createReq2, 1)
	if err != nil {
		t.Fatalf("Failed to create second test property: %v", err)
	}

	newName := "Updated Property Name"
	newAddress := "Updated Address"
	invalidType := models.PropertyType("InvalidType")
	duplicateName := "Another Property"

	tests := []struct {
		name        string
		propertyID  int
		request     *models.UpdatePropertyRequest
		userID      int
		expectError bool
		errorMsg    string
	}{
		{
			name:       "Valid update",
			propertyID: createdProperty.ID,
			request: &models.UpdatePropertyRequest{
				PropertyName: &newName,
				Address:      &newAddress,
			},
			userID:      1,
			expectError: false,
		},
		{
			name:       "Invalid property type",
			propertyID: createdProperty.ID,
			request: &models.UpdatePropertyRequest{
				PropertyType: &invalidType,
			},
			userID:      1,
			expectError: true,
			errorMsg:    "invalid property type: must be Residential, Commercial, or Mixed",
		},
		{
			name:       "Duplicate property name",
			propertyID: createdProperty.ID,
			request: &models.UpdatePropertyRequest{
				PropertyName: &duplicateName,
			},
			userID:      1,
			expectError: true,
			errorMsg:    "property name already exists",
		},
		{
			name:       "Non-existent property",
			propertyID: 99999,
			request: &models.UpdatePropertyRequest{
				PropertyName: &newName,
			},
			userID:      1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.UpdateProperty(tt.propertyID, tt.request, tt.userID, 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if property == nil {
				t.Errorf("Expected property but got nil")
				return
			}

			if tt.request.PropertyName != nil && property.PropertyName != *tt.request.PropertyName {
				t.Errorf("Expected property name '%s', got '%s'", *tt.request.PropertyName, property.PropertyName)
			}

			if tt.request.Address != nil && property.Address != *tt.request.Address {
				t.Errorf("Expected address '%s', got '%s'", *tt.request.Address, property.Address)
			}
		})
	}
}

func TestPropertyService_DeleteProperty(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	// Create a test property
	createReq := &models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	tests := []struct {
		name        string
		propertyID  int
		userID      int
		expectError bool
	}{
		{
			name:        "Valid deletion",
			propertyID:  createdProperty.ID,
			userID:      1,
			expectError: false,
		},
		{
			name:        "Non-existent property",
			propertyID:  99999,
			userID:      1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteProperty(tt.propertyID, tt.userID, 0)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify property is marked as inactive
			property, err := service.GetProperty(tt.propertyID, 0)
			if err == nil && property.Active {
				t.Errorf("Expected property to be inactive after deletion")
			}
		})
	}
}

func TestPropertyService_ValidatePropertyType(t *testing.T) {
	db, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	tests := []struct {
		name         string
		propertyType string
		expectError  bool
	}{
		{
			name:         "Valid Residential type",
			propertyType: "Residential",
			expectError:  false,
		},
		{
			name:         "Valid Commercial type",
			propertyType: "Commercial",
			expectError:  false,
		},
		{
			name:         "Valid Mixed type",
			propertyType: "Mixed",
			expectError:  false,
		},
		{
			name:         "Invalid type",
			propertyType: "InvalidType",
			expectError:  true,
		},
		{
			name:         "Empty type",
			propertyType: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePropertyType(tt.propertyType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

