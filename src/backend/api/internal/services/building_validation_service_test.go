package services

import (
	"errors"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// Mock implementations for testing

// MockBuildingRepository implements BuildingRepositoryInterface for testing
type MockBuildingRepository struct {
	buildings map[int]*models.Building
	codeMap   map[string]map[int]*models.Building // propertyID -> buildingCode -> building
}

func NewMockBuildingRepository() *MockBuildingRepository {
	return &MockBuildingRepository{
		buildings: make(map[int]*models.Building),
		codeMap:   make(map[string]map[int]*models.Building),
	}
}

func (m *MockBuildingRepository) Create(building *models.Building) error {
	building.ID = len(m.buildings) + 1
	building.CreatedAt = time.Now()
	building.UpdatedAt = time.Now()
	m.buildings[building.ID] = building

	// Update code map
	if m.codeMap[building.BuildingCode] == nil {
		m.codeMap[building.BuildingCode] = make(map[int]*models.Building)
	}
	m.codeMap[building.BuildingCode][building.PropertyID] = building

	return nil
}

func (m *MockBuildingRepository) GetByID(id int) (*models.Building, error) {
	if building, exists := m.buildings[id]; exists {
		return building, nil
	}
	return nil, errors.New("building not found")
}

func (m *MockBuildingRepository) GetByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	if propertyBuildings, exists := m.codeMap[code]; exists {
		if building, exists := propertyBuildings[propertyID]; exists {
			return building, nil
		}
	}
	return nil, errors.New("building not found")
}

func (m *MockBuildingRepository) GetByPropertyID(propertyID int) ([]*models.Building, error) {
	var buildings []*models.Building
	for _, building := range m.buildings {
		if building.PropertyID == propertyID {
			buildings = append(buildings, building)
		}
	}
	return buildings, nil
}

func (m *MockBuildingRepository) Update(id int, updates map[string]interface{}) error {
	if _, exists := m.buildings[id]; !exists {
		return errors.New("building not found")
	}
	// Simple mock implementation
	return nil
}

func (m *MockBuildingRepository) SoftDelete(id int) error {
	if _, exists := m.buildings[id]; !exists {
		return errors.New("building not found")
	}
	// Simple mock implementation
	return nil
}

func (m *MockBuildingRepository) GetWithStats(id int) (*models.BuildingWithStats, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBuildingRepository) BulkCreate(buildings []*models.Building) error {
	return errors.New("not implemented")
}

func (m *MockBuildingRepository) Search(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBuildingRepository) GetAnalytics(id int) (*models.BuildingAnalytics, error) {
	return nil, errors.New("not implemented")
}

// CountByProperty counts buildings for a property with optional filters
func (m *MockBuildingRepository) CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error) {
	count := 0
	for _, building := range m.buildings {
		if building.PropertyID == propertyID {
			count++
		}
	}
	return count, nil
}

// GetByPropertyWithSorting retrieves buildings for a property with sorting and filtering
func (m *MockBuildingRepository) GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error) {
	var buildings []*models.Building
	for _, building := range m.buildings {
		if building.PropertyID == propertyID {
			buildings = append(buildings, building)
		}
	}
	return buildings, nil
}

// AdvancedSearch performs advanced building search with metadata queries and enhanced filtering
func (m *MockBuildingRepository) AdvancedSearch(req *models.BuildingSearchRequest) ([]*models.Building, int, error) {
	return nil, 0, errors.New("not implemented")
}

// GetBuildingUnits retrieves units for a specific building with pagination
func (m *MockBuildingRepository) GetBuildingUnits(buildingID int, offset, limit int) ([]*models.BuildingUnitSummary, int, error) {
	return nil, 0, errors.New("not implemented")
}

// MockPropertyRepository implements PropertyRepository interface for testing
type MockPropertyRepository struct {
	properties map[int]*models.Property
}

func NewMockPropertyRepository() *MockPropertyRepository {
	return &MockPropertyRepository{
		properties: make(map[int]*models.Property),
	}
}

func (m *MockPropertyRepository) GetByID(id int) (*models.Property, error) {
	if property, exists := m.properties[id]; exists {
		return property, nil
	}
	return nil, errors.New("property not found")
}

func (m *MockPropertyRepository) AddProperty(property *models.Property) {
	m.properties[property.ID] = property
}

// Implement other methods required by PropertyRepository interface
func (m *MockPropertyRepository) Create(property *models.CreatePropertyRequest) (*models.Property, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPropertyRepository) GetByIDWithStats(id int) (*models.PropertyWithStats, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPropertyRepository) List(filters map[string]interface{}, limit, offset int) ([]*models.Property, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockPropertyRepository) Update(id int, updates *models.UpdatePropertyRequest) (*models.Property, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPropertyRepository) Delete(id int) error {
	return errors.New("not implemented")
}

func (m *MockPropertyRepository) CheckPropertyNameExists(name string, excludeID int) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockPropertyRepository) CheckPropertyCodeExists(code string, excludeID int) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockPropertyRepository) HasActiveBuildings(id int) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockPropertyRepository) HasActiveUnits(id int) (bool, error) {
	return false, errors.New("not implemented")
}

// MockMetadataValidator implements MetadataValidatorInterface for testing
type MockMetadataValidator struct {
	shouldFail bool
}

func NewMockMetadataValidator() *MockMetadataValidator {
	return &MockMetadataValidator{shouldFail: false}
}

func (m *MockMetadataValidator) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func (m *MockMetadataValidator) ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error {
	if m.shouldFail {
		return errors.New("metadata validation failed")
	}
	return nil
}

func (m *MockMetadataValidator) ValidateResidentialMetadata(metadata models.BuildingMetadata) error {
	return nil
}

func (m *MockMetadataValidator) ValidateCommercialMetadata(metadata models.BuildingMetadata) error {
	return nil
}

func (m *MockMetadataValidator) ValidateMixedMetadata(metadata models.BuildingMetadata) error {
	return nil
}

func (m *MockMetadataValidator) GetMetadataSchema(buildingType models.BuildingType) map[string]interface{} {
	return nil
}

// Test helper functions
func setupValidationService() (*BuildingValidationService, *MockBuildingRepository, *MockPropertyRepository, *MockMetadataValidator) {
	buildingRepo := NewMockBuildingRepository()
	propertyRepo := NewMockPropertyRepository()
	metadataValidator := NewMockMetadataValidator()

	// Add a test property
	propertyRepo.AddProperty(&models.Property{
		ID:           1,
		PropertyName: "Test Property",
		PropertyCode: "PROP001",
		Active:       true,
	})

	service := NewBuildingValidationService(buildingRepo, propertyRepo, metadataValidator)
	return service, buildingRepo, propertyRepo, metadataValidator
}

// Test ValidateBuildingType
func TestBuildingValidationService_ValidateBuildingType(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name         string
		buildingType models.BuildingType
		expectError  bool
	}{
		{
			name:         "Valid Residential Type",
			buildingType: models.BuildingTypeResidential,
			expectError:  false,
		},
		{
			name:         "Valid Commercial Type",
			buildingType: models.BuildingTypeCommercial,
			expectError:  false,
		},
		{
			name:         "Valid Mixed Type",
			buildingType: models.BuildingTypeMixed,
			expectError:  false,
		},
		{
			name:         "Invalid Type",
			buildingType: models.BuildingType("Invalid"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBuildingType(tt.buildingType)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateBuildingCodeUniqueness
func TestBuildingValidationService_ValidateBuildingCodeUniqueness(t *testing.T) {
	service, buildingRepo, _, _ := setupValidationService()

	// Add an existing building
	existingBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingCode: "B001",
		BuildingName: "Existing Building",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	buildingRepo.Create(existingBuilding)

	tests := []struct {
		name        string
		propertyID  int
		code        string
		excludeID   *int
		expectError bool
	}{
		{
			name:        "Unique Code",
			propertyID:  1,
			code:        "B002",
			excludeID:   nil,
			expectError: false,
		},
		{
			name:        "Duplicate Code",
			propertyID:  1,
			code:        "B001",
			excludeID:   nil,
			expectError: true,
		},
		{
			name:        "Duplicate Code But Excluded",
			propertyID:  1,
			code:        "B001",
			excludeID:   &existingBuilding.ID,
			expectError: false,
		},
		{
			name:        "Empty Code",
			propertyID:  1,
			code:        "",
			excludeID:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBuildingCodeUniqueness(tt.propertyID, tt.code, tt.excludeID)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidatePropertyAssociation
func TestBuildingValidationService_ValidatePropertyAssociation(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name        string
		propertyID  int
		expectError bool
	}{
		{
			name:        "Valid Property ID",
			propertyID:  1,
			expectError: false,
		},
		{
			name:        "Invalid Property ID - Zero",
			propertyID:  0,
			expectError: true,
		},
		{
			name:        "Invalid Property ID - Negative",
			propertyID:  -1,
			expectError: true,
		},
		{
			name:        "Non-existent Property ID",
			propertyID:  999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePropertyAssociation(tt.propertyID)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateConstructionYear
func TestBuildingValidationService_ValidateConstructionYear(t *testing.T) {
	service, _, _, _ := setupValidationService()

	currentYear := time.Now().Year()
	validYear := 2020
	oldYear := 1799
	futureYear := currentYear + 10

	tests := []struct {
		name        string
		year        *int
		expectError bool
	}{
		{
			name:        "Nil Year (Optional)",
			year:        nil,
			expectError: false,
		},
		{
			name:        "Valid Year",
			year:        &validYear,
			expectError: false,
		},
		{
			name:        "Too Old Year",
			year:        &oldYear,
			expectError: true,
		},
		{
			name:        "Too Future Year",
			year:        &futureYear,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateConstructionYear(tt.year)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateFloorCountAndElevatorRequirement
func TestBuildingValidationService_ValidateFloorCountAndElevatorRequirement(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name        string
		totalFloors int
		hasElevator bool
		expectError bool
	}{
		{
			name:        "Valid Floor Count - 1 Floor",
			totalFloors: 1,
			hasElevator: false,
			expectError: false,
		},
		{
			name:        "Valid Floor Count - 5 Floors with Elevator",
			totalFloors: 5,
			hasElevator: true,
			expectError: false,
		},
		{
			name:        "Valid Floor Count - 5 Floors without Elevator (Warning)",
			totalFloors: 5,
			hasElevator: false,
			expectError: false, // This is a warning, not an error
		},
		{
			name:        "Invalid Floor Count - Zero",
			totalFloors: 0,
			hasElevator: false,
			expectError: true,
		},
		{
			name:        "Invalid Floor Count - Negative",
			totalFloors: -1,
			hasElevator: false,
			expectError: true,
		},
		{
			name:        "Invalid Floor Count - Too High",
			totalFloors: 101,
			hasElevator: true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateFloorCountAndElevatorRequirement(tt.totalFloors, tt.hasElevator)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateMetadataSchema
func TestBuildingValidationService_ValidateMetadataSchema(t *testing.T) {
	service, _, _, metadataValidator := setupValidationService()

	metadata := models.BuildingMetadata{
		"amenities": []string{"gym", "pool"},
	}

	tests := []struct {
		name           string
		buildingType   models.BuildingType
		metadata       models.BuildingMetadata
		validatorFails bool
		expectError    bool
	}{
		{
			name:           "Valid Metadata",
			buildingType:   models.BuildingTypeResidential,
			metadata:       metadata,
			validatorFails: false,
			expectError:    false,
		},
		{
			name:           "Nil Metadata",
			buildingType:   models.BuildingTypeResidential,
			metadata:       nil,
			validatorFails: false,
			expectError:    false,
		},
		{
			name:           "Invalid Building Type",
			buildingType:   models.BuildingType("Invalid"),
			metadata:       metadata,
			validatorFails: false,
			expectError:    true,
		},
		{
			name:           "Metadata Validation Fails",
			buildingType:   models.BuildingTypeResidential,
			metadata:       metadata,
			validatorFails: true,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataValidator.SetShouldFail(tt.validatorFails)
			err := service.ValidateMetadataSchema(tt.buildingType, tt.metadata)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateBuildingCreation
func TestBuildingValidationService_ValidateBuildingCreation(t *testing.T) {
	service, _, _, _ := setupValidationService()

	validRequest := &models.CreateBuildingRequest{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "B001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		HasElevator:  false,
	}

	tests := []struct {
		name        string
		request     *models.CreateBuildingRequest
		expectError bool
	}{
		{
			name:        "Valid Request",
			request:     validRequest,
			expectError: false,
		},
		{
			name: "Missing Building Name",
			request: &models.CreateBuildingRequest{
				PropertyID:   1,
				BuildingName: "",
				BuildingCode: "B002",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
			},
			expectError: true,
		},
		{
			name: "Missing Building Code",
			request: &models.CreateBuildingRequest{
				PropertyID:   1,
				BuildingName: "Test Building",
				BuildingCode: "",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
			},
			expectError: true,
		},
		{
			name: "Invalid Property ID",
			request: &models.CreateBuildingRequest{
				PropertyID:   999,
				BuildingName: "Test Building",
				BuildingCode: "B003",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
			},
			expectError: true,
		},
		{
			name: "Invalid Building Type",
			request: &models.CreateBuildingRequest{
				PropertyID:   1,
				BuildingName: "Test Building",
				BuildingCode: "B004",
				BuildingType: models.BuildingType("Invalid"),
				TotalFloors:  3,
			},
			expectError: true,
		},
		{
			name: "Invalid Floor Count",
			request: &models.CreateBuildingRequest{
				PropertyID:   1,
				BuildingName: "Test Building",
				BuildingCode: "B005",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBuildingCreation(tt.request)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateBuildingUpdate
func TestBuildingValidationService_ValidateBuildingUpdate(t *testing.T) {
	service, buildingRepo, _, _ := setupValidationService()

	// Add an existing building
	existingBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingCode: "B001",
		BuildingName: "Existing Building",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	buildingRepo.Create(existingBuilding)

	validFloors := 5
	invalidFloors := 0
	validType := models.BuildingTypeCommercial
	invalidType := models.BuildingType("Invalid")

	tests := []struct {
		name        string
		buildingID  int
		request     *models.UpdateBuildingRequest
		expectError bool
	}{
		{
			name:       "Valid Update",
			buildingID: 1,
			request: &models.UpdateBuildingRequest{
				TotalFloors: &validFloors,
			},
			expectError: false,
		},
		{
			name:       "Non-existent Building",
			buildingID: 999,
			request: &models.UpdateBuildingRequest{
				TotalFloors: &validFloors,
			},
			expectError: true,
		},
		{
			name:       "Invalid Floor Count",
			buildingID: 1,
			request: &models.UpdateBuildingRequest{
				TotalFloors: &invalidFloors,
			},
			expectError: true,
		},
		{
			name:       "Invalid Building Type",
			buildingID: 1,
			request: &models.UpdateBuildingRequest{
				BuildingType: &invalidType,
			},
			expectError: true,
		},
		{
			name:       "Valid Building Type Update",
			buildingID: 1,
			request: &models.UpdateBuildingRequest{
				BuildingType: &validType,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBuildingUpdate(tt.buildingID, tt.request)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateDeletionConstraints
func TestBuildingValidationService_ValidateDeletionConstraints(t *testing.T) {
	service, buildingRepo, _, _ := setupValidationService()

	// Add an existing building
	existingBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingCode: "B001",
		BuildingName: "Existing Building",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
		ActiveStatus: true,
	}
	buildingRepo.Create(existingBuilding)

	tests := []struct {
		name        string
		buildingID  int
		expectError bool
	}{
		{
			name:        "Valid Building ID",
			buildingID:  1,
			expectError: false,
		},
		{
			name:        "Invalid Building ID - Zero",
			buildingID:  0,
			expectError: true,
		},
		{
			name:        "Invalid Building ID - Negative",
			buildingID:  -1,
			expectError: true,
		},
		{
			name:        "Non-existent Building ID",
			buildingID:  999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateDeletionConstraints(tt.buildingID)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateFloorRange
func TestBuildingValidationService_ValidateFloorRange(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name        string
		floorRange  string
		expectError bool
	}{
		{
			name:        "Valid Floor Range",
			floorRange:  "1-5",
			expectError: false,
		},
		{
			name:        "Valid Floor Range - Double Digits",
			floorRange:  "10-15",
			expectError: false,
		},
		{
			name:        "Empty Floor Range",
			floorRange:  "",
			expectError: true,
		},
		{
			name:        "Invalid Format - No Dash",
			floorRange:  "15",
			expectError: true,
		},
		{
			name:        "Invalid Format - Too Short",
			floorRange:  "1-",
			expectError: true, // This should fail basic validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateFloorRange(tt.floorRange)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateBusinessHours
func TestBuildingValidationService_ValidateBusinessHours(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name          string
		businessHours string
		expectError   bool
	}{
		{
			name:          "Valid Business Hours",
			businessHours: "09:00-18:00",
			expectError:   false,
		},
		{
			name:          "Valid Business Hours - 24 Hour Format",
			businessHours: "08:30-22:30",
			expectError:   false,
		},
		{
			name:          "Empty Business Hours",
			businessHours: "",
			expectError:   true,
		},
		{
			name:          "Invalid Format - No Dash",
			businessHours: "09:00",
			expectError:   true,
		},
		{
			name:          "Invalid Format - No Colon",
			businessHours: "0900-1800",
			expectError:   true,
		},
		{
			name:          "Invalid Format - Too Short",
			businessHours: "9-18",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBusinessHours(tt.businessHours)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateElevatorRequirement
func TestBuildingValidationService_ValidateElevatorRequirement(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name         string
		totalFloors  int
		hasElevator  bool
		buildingType models.BuildingType
		expectError  bool
	}{
		{
			name:         "Residential - 2 Floors No Elevator",
			totalFloors:  2,
			hasElevator:  false,
			buildingType: models.BuildingTypeResidential,
			expectError:  false,
		},
		{
			name:         "Residential - 5 Floors No Elevator (Warning)",
			totalFloors:  5,
			hasElevator:  false,
			buildingType: models.BuildingTypeResidential,
			expectError:  false, // This is a warning, not an error
		},
		{
			name:         "Commercial - 4 Floors No Elevator (Warning)",
			totalFloors:  4,
			hasElevator:  false,
			buildingType: models.BuildingTypeCommercial,
			expectError:  false, // This is a warning, not an error
		},
		{
			name:         "Mixed - 4 Floors No Elevator (Warning)",
			totalFloors:  4,
			hasElevator:  false,
			buildingType: models.BuildingTypeMixed,
			expectError:  false, // This is a warning, not an error
		},
		{
			name:         "Any Type - With Elevator",
			totalFloors:  10,
			hasElevator:  true,
			buildingType: models.BuildingTypeCommercial,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateElevatorRequirement(tt.totalFloors, tt.hasElevator, tt.buildingType)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Test ValidateMetadataConsistency
func TestBuildingValidationService_ValidateMetadataConsistency(t *testing.T) {
	service, _, _, _ := setupValidationService()

	tests := []struct {
		name         string
		buildingType models.BuildingType
		metadata     models.BuildingMetadata
		expectError  bool
	}{
		{
			name:         "Nil Metadata",
			buildingType: models.BuildingTypeResidential,
			metadata:     nil,
			expectError:  false,
		},
		{
			name:         "Residential - Valid Parking Consistency",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":   float64(100),
					"covered": float64(60),
					"visitor": float64(20),
				},
			},
			expectError: false,
		},
		{
			name:         "Residential - Invalid Parking Consistency",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":   float64(100),
					"covered": float64(80),
					"visitor": float64(40), // 80 + 40 > 100
				},
			},
			expectError: true,
		},
		{
			name:         "Commercial - Valid Parking Consistency",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":    float64(200),
					"customer": float64(150),
					"staff":    float64(50),
				},
			},
			expectError: false,
		},
		{
			name:         "Commercial - Invalid Parking Consistency",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":    float64(200),
					"customer": float64(150),
					"staff":    float64(100), // 150 + 100 > 200
				},
			},
			expectError: true,
		},
		{
			name:         "Mixed - Valid Parking Consistency",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"shared_facilities": map[string]interface{}{
					"parking_total": float64(300),
				},
				"commercial_section": map[string]interface{}{
					"parking_allocation": float64(200),
				},
			},
			expectError: false,
		},
		{
			name:         "Mixed - Invalid Parking Consistency",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"shared_facilities": map[string]interface{}{
					"parking_total": float64(300),
				},
				"commercial_section": map[string]interface{}{
					"parking_allocation": float64(400), // 400 > 300
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateMetadataConsistency(tt.buildingType, tt.metadata)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
