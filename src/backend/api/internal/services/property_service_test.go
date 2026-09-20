package services

import (
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/testutil"

	_ "github.com/lib/pq"
)

// servicesTestDB is this package's own database: go test ./internal/... runs
// packages in parallel and teardown drops every table, so sharing one database
// with the repository tests would tear down their schema mid-run.
const servicesTestDB = "tenantly_test_services"

// setupPropertyTestDB migrates a real schema rather than hand-rolling one. The
// bespoke CREATE TABLEs this replaced predated multi-tenancy and carried no
// organization_id, and it connected as a "postgres" role that exists neither
// here nor in CI — so these tests silently skipped instead of running.
func setupPropertyTestDB(t *testing.T) (*sqlx.DB, int, func()) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	return db, testutil.CreateTestOrganization(t, db), cleanup
}

func TestPropertyService_CreateProperty(t *testing.T) {
	db, orgID, cleanup := setupPropertyTestDB(t)
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
			tt.request.OrganizationID = orgID
			property, err := service.CreateProperty(tt.request, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
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
	db, orgID, cleanup := setupPropertyTestDB(t)
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
	createReq.OrganizationID = orgID
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	tests := []struct {
		name        string
		propertyID  int
		expectError bool
		errorMsg    string
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
			// Regression check: the repo's "property not found" must reach the
			// handler unwrapped, or its 404 detection silently falls through to 500.
			errorMsg: "property not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.GetProperty(tt.propertyID, orgID)

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

			if property.ID != tt.propertyID {
				t.Errorf("Expected property ID %d, got %d", tt.propertyID, property.ID)
			}
		})
	}
}

func TestPropertyService_GetPropertyWithStats(t *testing.T) {
	db, orgID, cleanup := setupPropertyTestDB(t)
	defer cleanup()

	propertyRepo := repositories.NewPropertyRepository(db)
	auditService := database.NewAuditService(db)
	service := NewPropertyService(propertyRepo, auditService)

	createReq := &models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		Address:      "123 Test Street",
		PropertyType: models.PropertyTypeCommercial,
	}
	createReq.OrganizationID = orgID
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	tests := []struct {
		name        string
		propertyID  int
		orgID       int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid property ID",
			propertyID:  createdProperty.ID,
			orgID:       orgID,
			expectError: false,
		},
		{
			name:        "Invalid property ID",
			propertyID:  99999,
			orgID:       orgID,
			expectError: true,
			// Regression check: the repo's "property not found" must reach the
			// handler unwrapped, or its 404 detection silently falls through to 500.
			errorMsg: "property not found",
		},
		{
			name:        "Property belongs to a different organization",
			propertyID:  createdProperty.ID,
			orgID:       999,
			expectError: true,
			errorMsg:    "property not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.GetPropertyWithStats(tt.propertyID, tt.orgID)

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

			if property.ID != tt.propertyID {
				t.Errorf("Expected property ID %d, got %d", tt.propertyID, property.ID)
			}
		})
	}
}

func TestPropertyService_ListProperties(t *testing.T) {
	db, orgID, cleanup := setupPropertyTestDB(t)
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
		prop.OrganizationID = orgID
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
			// Scope to this test's organization, exactly as PropertyHandler
			// does. Without it the list also returns the "Default Property"
			// the initial migration seeds.
			tt.filters["organization_id"] = orgID

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
	db, orgID, cleanup := setupPropertyTestDB(t)
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
	createReq.OrganizationID = orgID
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
	createReq2.OrganizationID = orgID
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
			// Regression check: the repo's "property not found" must reach the
			// handler unwrapped, or its 404 detection silently falls through to 500.
			errorMsg: "property not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			property, err := service.UpdateProperty(tt.propertyID, tt.request, tt.userID, orgID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
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
	db, orgID, cleanup := setupPropertyTestDB(t)
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
	createReq.OrganizationID = orgID
	createdProperty, err := service.CreateProperty(createReq, 1)
	if err != nil {
		t.Fatalf("Failed to create test property: %v", err)
	}

	tests := []struct {
		name        string
		propertyID  int
		userID      int
		expectError bool
		errorMsg    string
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
			// Regression check: the repo's "property not found" must reach the
			// handler unwrapped, or its 404 detection silently falls through to 500.
			errorMsg: "property not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteProperty(tt.propertyID, tt.userID, orgID)

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

			// Verify property is marked as inactive
			property, err := service.GetProperty(tt.propertyID, orgID)
			if err == nil && property.Active {
				t.Errorf("Expected property to be inactive after deletion")
			}
		})
	}
}

func TestPropertyService_ValidatePropertyType(t *testing.T) {
	db, _, cleanup := setupPropertyTestDB(t)
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
