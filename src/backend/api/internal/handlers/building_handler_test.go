package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/models"
)

// MockBuildingService is a mock implementation of BuildingService
type MockBuildingService struct {
	mock.Mock
}

func (m *MockBuildingService) CreateBuilding(req *models.CreateBuildingRequest) (*models.Building, error) {
	args := m.Called(req)
	return args.Get(0).(*models.Building), args.Error(1)
}

func (m *MockBuildingService) GetBuilding(id, orgID int) (*models.Building, error) {
	args := m.Called(id, orgID)
	return args.Get(0).(*models.Building), args.Error(1)
}

func (m *MockBuildingService) GetBuildingWithStats(id, orgID int) (*models.BuildingWithStats, error) {
	args := m.Called(id, orgID)
	return args.Get(0).(*models.BuildingWithStats), args.Error(1)
}

func (m *MockBuildingService) UpdateBuilding(id int, req *models.UpdateBuildingRequest, orgID int) (*models.Building, error) {
	args := m.Called(id, req, orgID)
	return args.Get(0).(*models.Building), args.Error(1)
}

func (m *MockBuildingService) DeleteBuilding(id, orgID int) error {
	args := m.Called(id, orgID)
	return args.Error(0)
}

func (m *MockBuildingService) GetBuildingsByProperty(propertyID int) ([]*models.Building, error) {
	args := m.Called(propertyID)
	return args.Get(0).([]*models.Building), args.Error(1)
}

func (m *MockBuildingService) GetBuildingsByPropertyWithStats(propertyID int) ([]*models.BuildingWithStats, error) {
	args := m.Called(propertyID)
	return args.Get(0).([]*models.BuildingWithStats), args.Error(1)
}

func (m *MockBuildingService) BulkCreateBuildings(req *models.BulkCreateBuildingsRequest) ([]*models.Building, error) {
	args := m.Called(req)
	return args.Get(0).([]*models.Building), args.Error(1)
}

func (m *MockBuildingService) SearchBuildings(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	args := m.Called(filters)
	return args.Get(0).([]*models.Building), args.Error(1)
}

func (m *MockBuildingService) GetBuildingAnalytics(id, orgID int) (*models.BuildingAnalytics, error) {
	args := m.Called(id, orgID)
	return args.Get(0).(*models.BuildingAnalytics), args.Error(1)
}

func (m *MockBuildingService) GetBuildingByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	args := m.Called(propertyID, code)
	return args.Get(0).(*models.Building), args.Error(1)
}

func (m *MockBuildingService) GetPropertyBuildingAnalytics(propertyID int) ([]*models.BuildingAnalytics, error) {
	args := m.Called(propertyID)
	return args.Get(0).([]*models.BuildingAnalytics), args.Error(1)
}

func (m *MockBuildingService) CalculateBuildingOccupancyRate(id int) (float64, error) {
	args := m.Called(id)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockBuildingService) CalculateBuildingRevenue(id int) (float64, error) {
	args := m.Called(id)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockBuildingService) GetBuildingUnitCounts(id int) (total int, occupied int, vacant int, err error) {
	args := m.Called(id)
	return args.Get(0).(int), args.Get(1).(int), args.Get(2).(int), args.Error(3)
}

func (m *MockBuildingService) ValidateBuildingCreation(req *models.CreateBuildingRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockBuildingService) ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func (m *MockBuildingService) ValidateBuildingDeletion(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBuildingService) ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error {
	args := m.Called(propertyID, code, excludeID)
	return args.Error(0)
}

// New methods added for enhanced property-building relationship operations
func (m *MockBuildingService) GetPropertyBuildingsWithPagination(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string, includeStats bool) (*models.BuildingListResponse, error) {
	args := m.Called(propertyID, filters, sortBy, sortOrder, includeStats)
	return args.Get(0).(*models.BuildingListResponse), args.Error(1)
}

func (m *MockBuildingService) GetPropertyStatistics(propertyID int) (*models.PropertyStatistics, error) {
	args := m.Called(propertyID)
	return args.Get(0).(*models.PropertyStatistics), args.Error(1)
}

func (m *MockBuildingService) GetPropertyWithBuildingsHierarchy(propertyID int, includeStats bool) (*models.PropertyWithBuildings, error) {
	args := m.Called(propertyID, includeStats)
	return args.Get(0).(*models.PropertyWithBuildings), args.Error(1)
}

// Advanced building API features methods
func (m *MockBuildingService) AdvancedSearchBuildings(req *models.BuildingSearchRequest) (*models.BuildingListResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.BuildingListResponse), args.Error(1)
}

func (m *MockBuildingService) GetBuildingUnits(buildingID int, page, pageSize int) (*models.BuildingUnitsResponse, error) {
	args := m.Called(buildingID, page, pageSize)
	return args.Get(0).(*models.BuildingUnitsResponse), args.Error(1)
}

func (m *MockBuildingService) GetBuildingMetadataSchema(buildingType models.BuildingType) (*models.MetadataSchemaResponse, error) {
	args := m.Called(buildingType)
	return args.Get(0).(*models.MetadataSchemaResponse), args.Error(1)
}

func (m *MockBuildingService) ExportBuildingData(req *models.BuildingExportRequest) ([]byte, string, error) {
	args := m.Called(req)
	return args.Get(0).([]byte), args.Get(1).(string), args.Error(2)
}

func (m *MockBuildingService) UpdateBuildingStatus(buildingID int, req *models.BuildingStatusRequest) (*models.Building, error) {
	args := m.Called(buildingID, req)
	return args.Get(0).(*models.Building), args.Error(1)
}

func TestBuildingHandler_CreateBuilding_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	// Mock request
	req := &models.CreateBuildingRequest{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
	}

	// Mock response
	expectedBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
		ActiveStatus: true,
	}

	mockService.On("CreateBuilding", req).Return(expectedBuilding, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/buildings", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handler.CreateBuilding(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Building created successfully", response["message"])
	assert.NotNil(t, response["building"])

	mockService.AssertExpectations(t)
}

func TestBuildingHandler_GetBuilding_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	// Mock response
	expectedBuilding := &models.Building{
		ID:           1,
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		HasElevator:  true,
		ActiveStatus: true,
	}

	mockService.On("GetBuilding", 1, 0).Return(expectedBuilding, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/buildings/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Call handler
	handler.GetBuilding(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["building"])

	mockService.AssertExpectations(t)
}

func TestBuildingHandler_GetBuilding_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	mockService.On("GetBuilding", 999, 0).Return((*models.Building)(nil), assert.AnError)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/buildings/999", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	// Call handler
	handler.GetBuilding(c)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestBuildingHandler_DeleteBuilding_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockBuildingService)
	handler := NewBuildingHandler(mockService)

	mockService.On("DeleteBuilding", 1, 0).Return(nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/v1/buildings/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Call handler
	handler.DeleteBuilding(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Building deleted successfully", response["message"])

	mockService.AssertExpectations(t)
}
