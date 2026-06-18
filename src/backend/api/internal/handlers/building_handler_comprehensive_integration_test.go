package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/ysnarafat/tenantly/internal/config"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/middleware"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/services"
)

// BuildingIntegrationTestSuite provides comprehensive integration testing for building management API
type BuildingIntegrationTestSuite struct {
	suite.Suite
	db              *sql.DB
	router          *gin.Engine
	config          *config.Config
	buildingHandler *BuildingHandler
	propertyRepo    *repositories.PropertyRepository
	buildingRepo    *repositories.BuildingRepository
	userRepo        *repositories.UserRepository
	testProperty    *models.Property
	testUser        *models.User
	authToken       string
}

func (suite *BuildingIntegrationTestSuite) SetupSuite() {
	// Set test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	suite.config = &config.Config{
		DatabaseURL:   getTestDatabaseURL(),
		JWTSecret:     "test-jwt-secret-key",
		JWTExpiration: time.Hour * 24,
		Environment:   "test",
	}

	// Initialize test database
	var err error
	suite.db, err = database.Connect(suite.config.DatabaseURL)
	if err != nil {
		suite.T().Skipf("Skipping integration test: PostgreSQL not available: %v", err)
		return
	}

	// Run migrations
	err = database.RunMigrations(suite.config.DatabaseURL)
	require.NoError(suite.T(), err, "Failed to run migrations")

	// Initialize repositories
	suite.propertyRepo = repositories.NewPropertyRepository(suite.db)
	suite.buildingRepo = repositories.NewBuildingRepository(suite.db)
	suite.userRepo = repositories.NewUserRepository(suite.db)

	// Initialize services
	auditService := database.NewAuditService(suite.db)
	metadataValidator := services.NewBuildingMetadataValidator()
	userService := services.NewUserService(suite.userRepo, auditService, suite.config.JWTSecret, suite.config.JWTExpiration)
	propertyService := services.NewPropertyService(suite.propertyRepo, auditService)
	buildingService := services.NewBuildingService(suite.buildingRepo, suite.propertyRepo, auditService, metadataValidator)

	// Initialize handlers
	userHandler := NewUserHandler(userService)
	propertyHandler := NewPropertyHandler(propertyService)
	suite.buildingHandler = NewBuildingHandler(buildingService)

	// Setup router with middleware
	suite.router = gin.New()
	suite.setupTestRoutes(userHandler, propertyHandler, auditService)

	// Create test data
	suite.createTestData()
}

func (suite *BuildingIntegrationTestSuite) TearDownSuite() {
	// Clean up test data
	suite.cleanupTestData()

	// Close database connection
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *BuildingIntegrationTestSuite) SetupTest() {
	// Clean up any test-specific data before each test
	suite.cleanupBuildingTestData()
}

func (suite *BuildingIntegrationTestSuite) setupTestRoutes(userHandler *UserHandler, propertyHandler *PropertyHandler, auditService *database.AuditService) {
	// Add middleware
	suite.router.Use(middleware.SecurityHeadersMiddleware())
	suite.router.Use(middleware.CORS(suite.config.Environment))
	suite.router.Use(gin.Logger())
	suite.router.Use(gin.Recovery())

	// API v1 routes
	v1 := suite.router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired(suite.config.JWTSecret, auditService))
		{
			// Property routes
			properties := protected.Group("/properties")
			{
				properties.GET("", propertyHandler.GetProperties)
				properties.POST("", propertyHandler.CreateProperty)
				properties.GET("/:id", propertyHandler.GetProperty)
				properties.GET("/:id/buildings", middleware.PropertyValidationMiddleware(suite.propertyRepo), suite.buildingHandler.GetPropertyBuildings)
				properties.POST("/:id/buildings/bulk", middleware.PropertyValidationMiddleware(suite.propertyRepo), suite.buildingHandler.BulkCreateBuildings)
			}

			// Building routes
			buildings := protected.Group("/buildings")
			{
				buildings.GET("", suite.buildingHandler.GetBuildings)
				buildings.POST("", suite.buildingHandler.CreateBuilding)
				buildings.GET("/search", suite.buildingHandler.AdvancedSearchBuildings)
				buildings.GET("/export", suite.buildingHandler.ExportBuildingData)
				buildings.GET("/types/:type/metadata", suite.buildingHandler.GetBuildingMetadataSchema)
				buildings.GET("/:id", suite.buildingHandler.GetBuilding)
				buildings.PUT("/:id", suite.buildingHandler.UpdateBuilding)
				buildings.DELETE("/:id", suite.buildingHandler.DeleteBuilding)
				buildings.GET("/:id/analytics", suite.buildingHandler.GetBuildingAnalytics)
				buildings.GET("/:id/units", suite.buildingHandler.GetBuildingUnits)
				buildings.PUT("/:id/status", suite.buildingHandler.UpdateBuildingStatus)
			}
		}
	}
}

func (suite *BuildingIntegrationTestSuite) createTestData() {
	// Create test user
	suite.testUser = &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "Admin",
		Active:       true,
	}

	err := suite.userRepo.Create(suite.testUser)
	require.NoError(suite.T(), err, "Failed to create test user")

	// Generate auth token
	suite.authToken = suite.generateAuthToken()

	// Create test property
	propertyReq := &models.CreatePropertyRequest{
		PropertyName: "Test Property",
		PropertyCode: "TEST001",
		PropertyType: models.PropertyTypeCommercial,
		Address:      "123 Test Street",
		City:         "Test City",
	}
	suite.testProperty, err = suite.propertyRepo.Create(propertyReq)
	require.NoError(suite.T(), err, "Failed to create test property")
}

func (suite *BuildingIntegrationTestSuite) cleanupTestData() {
	// Clean up in reverse order of creation
	if suite.testProperty != nil {
		suite.propertyRepo.Delete(suite.testProperty.ID)
	}
	if suite.testUser != nil {
		suite.userRepo.Delete(suite.testUser.ID)
	}
}

func (suite *BuildingIntegrationTestSuite) cleanupBuildingTestData() {
	// Clean up buildings created during tests
	if suite.testProperty != nil {
		buildings, _ := suite.buildingRepo.GetByPropertyID(suite.testProperty.ID)
		for _, building := range buildings {
			suite.buildingRepo.SoftDelete(building.ID)
		}
	}
}

func (suite *BuildingIntegrationTestSuite) generateAuthToken() string {
	loginReq := map[string]string{
		"username": suite.testUser.Username,
		"password": "password", // This should match the actual password used in tests
	}

	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if token, ok := response["token"].(string); ok {
			return "Bearer " + token
		}
	}

	// Fallback: generate token directly for testing
	return "Bearer test-token"
}

func (suite *BuildingIntegrationTestSuite) makeAuthenticatedRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// Test Building CRUD Operations
func (suite *BuildingIntegrationTestSuite) TestCreateBuilding_Success() {
	buildingReq := &models.CreateBuildingRequest{
		PropertyID:       suite.testProperty.ID,
		BuildingName:     "Test Building",
		BuildingCode:     "TB001",
		BuildingType:     models.BuildingTypeResidential,
		TotalFloors:      5,
		HasElevator:      true,
		ConstructionYear: intPtrComprehensive(2020),
		Metadata: models.BuildingMetadata{
			"amenities":               []string{"gym", "swimming_pool"},
			"security_type":           "24_hour_guard",
			"maintenance_staff_count": 3,
		},
	}

	w := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", buildingReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Building created successfully", response["message"])
	assert.NotNil(suite.T(), response["building"])

	// Verify building was created in database
	building := response["building"].(map[string]interface{})
	buildingID := int(building["id"].(float64))

	dbBuilding, err := suite.buildingRepo.GetByID(buildingID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Building", dbBuilding.BuildingName)
	assert.Equal(suite.T(), "TB001", dbBuilding.BuildingCode)
}

func (suite *BuildingIntegrationTestSuite) TestCreateBuilding_InvalidPropertyID() {
	buildingReq := &models.CreateBuildingRequest{
		PropertyID:   99999, // Non-existent property
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
	}

	w := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", buildingReq)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Property not found", response["error"])
}

func (suite *BuildingIntegrationTestSuite) TestCreateBuilding_DuplicateBuildingCode() {
	// Create first building
	buildingReq1 := &models.CreateBuildingRequest{
		PropertyID:   suite.testProperty.ID,
		BuildingName: "First Building",
		BuildingCode: "DUPLICATE001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  3,
	}

	w1 := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", buildingReq1)
	assert.Equal(suite.T(), http.StatusCreated, w1.Code)

	// Try to create second building with same code
	buildingReq2 := &models.CreateBuildingRequest{
		PropertyID:   suite.testProperty.ID,
		BuildingName: "Second Building",
		BuildingCode: "DUPLICATE001", // Same code
		BuildingType: models.BuildingTypeCommercial,
		TotalFloors:  5,
	}

	w2 := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", buildingReq2)
	assert.Equal(suite.T(), http.StatusConflict, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"].(string), "already exists")
}

func (suite *BuildingIntegrationTestSuite) TestGetBuilding_Success() {
	// Create a building first
	building := suite.createTestBuilding("Get Test Building", "GTB001")

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/buildings/%d", building.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	buildingData := response["building"].(map[string]interface{})
	assert.Equal(suite.T(), "Get Test Building", buildingData["building_name"])
	assert.Equal(suite.T(), "GTB001", buildingData["building_code"])
}

func (suite *BuildingIntegrationTestSuite) TestGetBuilding_NotFound() {
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/buildings/99999", nil)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Building not found", response["error"])
}

func (suite *BuildingIntegrationTestSuite) TestGetBuildingWithStats() {
	// Create a building first
	building := suite.createTestBuilding("Stats Test Building", "STB001")

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/buildings/%d?include_stats=true", building.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	buildingData := response["building"].(map[string]interface{})
	assert.NotNil(suite.T(), buildingData["unit_count"])
	assert.NotNil(suite.T(), buildingData["occupancy_rate"])
}

func (suite *BuildingIntegrationTestSuite) TestUpdateBuilding_Success() {
	// Create a building first
	building := suite.createTestBuilding("Update Test Building", "UTB001")

	updateReq := &models.UpdateBuildingRequest{
		BuildingName: stringPtr("Updated Building Name"),
		TotalFloors:  intPtrComprehensive(10),
		HasElevator:  boolPtr(true),
		Metadata: &models.BuildingMetadata{
			"amenities":     []string{"gym", "swimming_pool", "playground"},
			"security_type": "cctv_only",
		},
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/buildings/%d", building.ID), updateReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Building updated successfully", response["message"])

	// Verify update in database
	updatedBuilding, err := suite.buildingRepo.GetByID(building.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Building Name", updatedBuilding.BuildingName)
	assert.Equal(suite.T(), 10, updatedBuilding.TotalFloors)
	assert.True(suite.T(), updatedBuilding.HasElevator)
}

func (suite *BuildingIntegrationTestSuite) TestDeleteBuilding_Success() {
	// Create a building first
	building := suite.createTestBuilding("Delete Test Building", "DTB001")

	w := suite.makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/v1/buildings/%d", building.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Building deleted successfully", response["message"])

	// Verify building is soft deleted
	_, err = suite.buildingRepo.GetByID(building.ID)
	assert.Error(suite.T(), err) // Should not be found after soft delete
}

// Test Property-Building Relationship Endpoints
func (suite *BuildingIntegrationTestSuite) TestGetPropertyBuildings_Success() {
	// Create multiple buildings for the property
	suite.createTestBuilding("Property Building 1", "PB001")
	suite.createTestBuilding("Property Building 2", "PB002")
	suite.createTestBuilding("Property Building 3", "PB003")

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/properties/%d/buildings", suite.testProperty.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify hierarchical response structure
	assert.NotNil(suite.T(), response["property"])
	assert.NotNil(suite.T(), response["buildings"])
	assert.NotNil(suite.T(), response["pagination"])

	property := response["property"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testProperty.PropertyName, property["property_name"])

	buildings := response["buildings"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(buildings), 3)
}

func (suite *BuildingIntegrationTestSuite) TestGetPropertyBuildings_WithFiltering() {
	// Create buildings of different types
	suite.createTestBuildingWithType("Residential Building", "RB001", models.BuildingTypeResidential)
	suite.createTestBuildingWithType("Commercial Building", "CB001", models.BuildingTypeCommercial)
	suite.createTestBuildingWithType("Mixed Building", "MB001", models.BuildingTypeMixed)

	// Filter by building type
	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/properties/%d/buildings?building_type=Residential", suite.testProperty.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	buildings := response["buildings"].([]interface{})
	for _, b := range buildings {
		building := b.(map[string]interface{})
		assert.Equal(suite.T(), "Residential", building["building_type"])
	}
}

func (suite *BuildingIntegrationTestSuite) TestBulkCreateBuildings_Success() {
	bulkReq := &models.BulkCreateBuildingsRequest{
		PropertyID: suite.testProperty.ID,
		Buildings: []models.CreateBuildingRequest{
			{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Bulk Building 1",
				BuildingCode: "BB001",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
			},
			{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Bulk Building 2",
				BuildingCode: "BB002",
				BuildingType: models.BuildingTypeCommercial,
				TotalFloors:  5,
			},
			{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Bulk Building 3",
				BuildingCode: "BB003",
				BuildingType: models.BuildingTypeMixed,
				TotalFloors:  8,
			},
		},
	}

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/properties/%d/buildings/bulk", suite.testProperty.ID), bulkReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Buildings created successfully", response["message"])

	buildings := response["buildings"].([]interface{})
	assert.Equal(suite.T(), 3, len(buildings))

	summary := response["summary"].(map[string]interface{})
	assert.Equal(suite.T(), float64(3), summary["buildings_created"])
}

func (suite *BuildingIntegrationTestSuite) TestBulkCreateBuildings_DuplicateCodes() {
	bulkReq := &models.BulkCreateBuildingsRequest{
		PropertyID: suite.testProperty.ID,
		Buildings: []models.CreateBuildingRequest{
			{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Duplicate Building 1",
				BuildingCode: "DUPLICATE",
				BuildingType: models.BuildingTypeResidential,
				TotalFloors:  3,
			},
			{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Duplicate Building 2",
				BuildingCode: "DUPLICATE", // Same code
				BuildingType: models.BuildingTypeCommercial,
				TotalFloors:  5,
			},
		},
	}

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/properties/%d/buildings/bulk", suite.testProperty.ID), bulkReq)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Duplicate building codes in request", response["error"])
}

// Test Building Search and Filtering
func (suite *BuildingIntegrationTestSuite) TestSearchBuildings_BasicFiltering() {
	// Create buildings with different characteristics
	suite.createTestBuildingWithDetails("Search Building 1", "SB001", models.BuildingTypeResidential, 3, false)
	suite.createTestBuildingWithDetails("Search Building 2", "SB002", models.BuildingTypeCommercial, 10, true)
	suite.createTestBuildingWithDetails("Search Building 3", "SB003", models.BuildingTypeMixed, 7, true)

	// Test filtering by building type
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/buildings?building_type=Commercial", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	buildings := response["buildings"].([]interface{})
	for _, b := range buildings {
		building := b.(map[string]interface{})
		assert.Equal(suite.T(), "Commercial", building["building_type"])
	}

	// Test filtering by elevator
	w = suite.makeAuthenticatedRequest("GET", "/api/v1/buildings?has_elevator=true", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	buildings = response["buildings"].([]interface{})
	for _, b := range buildings {
		building := b.(map[string]interface{})
		assert.True(suite.T(), building["has_elevator"].(bool))
	}
}

func (suite *BuildingIntegrationTestSuite) TestAdvancedSearchBuildings() {
	// Create buildings with metadata
	suite.createTestBuildingWithMetadata("Advanced Search 1", "AS001", models.BuildingTypeResidential, map[string]interface{}{
		"amenities":     []string{"gym", "swimming_pool"},
		"security_type": "24_hour_guard",
	})
	suite.createTestBuildingWithMetadata("Advanced Search 2", "AS002", models.BuildingTypeCommercial, map[string]interface{}{
		"parking_spaces": map[string]interface{}{
			"total":    100,
			"customer": 80,
		},
		"business_hours": map[string]interface{}{
			"weekdays": "09:00-22:00",
		},
	})

	// Test advanced search with metadata query
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/buildings/search?building_type=Residential&page=1&page_size=10", nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["buildings"])
	assert.NotNil(suite.T(), response["pagination"])

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), pagination["current_page"])
	assert.Equal(suite.T(), float64(10), pagination["page_size"])
}

// Test Building Analytics and Aggregation
func (suite *BuildingIntegrationTestSuite) TestGetBuildingAnalytics() {
	// Create a building
	building := suite.createTestBuilding("Analytics Building", "AB001")

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/buildings/%d/analytics", building.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	analytics := response["analytics"].(map[string]interface{})
	assert.NotNil(suite.T(), analytics["building_id"])
	assert.NotNil(suite.T(), analytics["unit_count"])
	assert.NotNil(suite.T(), analytics["occupancy_rate"])
	assert.NotNil(suite.T(), analytics["monthly_revenue"])
}

func (suite *BuildingIntegrationTestSuite) TestGetBuildingUnits() {
	// Create a building
	building := suite.createTestBuilding("Units Building", "UB001")

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/buildings/%d/units", building.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(building.ID), response["building_id"])
	assert.Equal(suite.T(), building.BuildingName, response["building_name"])
	assert.NotNil(suite.T(), response["units"])
	assert.NotNil(suite.T(), response["summary"])
}

// Test Metadata Schema Endpoints
func (suite *BuildingIntegrationTestSuite) TestGetBuildingMetadataSchema() {
	buildingTypes := []string{"Residential", "Commercial", "Mixed"}

	for _, buildingType := range buildingTypes {
		w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/buildings/types/%s/metadata", buildingType), nil)

		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)

		assert.Equal(suite.T(), buildingType, response["building_type"])
		assert.NotNil(suite.T(), response["schema"])
		assert.NotNil(suite.T(), response["examples"])
		assert.NotNil(suite.T(), response["description"])
	}
}

// Test Building Status Management
func (suite *BuildingIntegrationTestSuite) TestUpdateBuildingStatus() {
	// Create a building
	building := suite.createTestBuilding("Status Building", "STB001")

	// Deactivate building
	statusReq := &models.BuildingStatusRequest{
		ActiveStatus: false,
		Reason:       "Maintenance required",
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/buildings/%d/status", building.ID), statusReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["message"].(string), "deactivated")

	// Verify status change in database
	updatedBuilding, err := suite.buildingRepo.GetByID(building.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updatedBuilding.ActiveStatus)

	// Reactivate building
	statusReq.ActiveStatus = true
	statusReq.Reason = "Maintenance completed"

	w = suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/buildings/%d/status", building.ID), statusReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["message"].(string), "activated")
}

// Test Error Handling and Validation
func (suite *BuildingIntegrationTestSuite) TestCreateBuilding_ValidationErrors() {
	testCases := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Missing required fields",
			request: map[string]interface{}{
				"building_name": "Test Building",
				// Missing property_id, building_code, building_type, total_floors
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
		{
			name: "Invalid building type",
			request: map[string]interface{}{
				"property_id":   suite.testProperty.ID,
				"building_name": "Test Building",
				"building_code": "TB001",
				"building_type": "InvalidType",
				"total_floors":  5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
		{
			name: "Invalid floor count",
			request: map[string]interface{}{
				"property_id":   suite.testProperty.ID,
				"building_name": "Test Building",
				"building_code": "TB001",
				"building_type": "Residential",
				"total_floors":  0, // Invalid
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
		{
			name: "Invalid construction year",
			request: map[string]interface{}{
				"property_id":       suite.testProperty.ID,
				"building_name":     "Test Building",
				"building_code":     "TB001",
				"building_type":     "Residential",
				"total_floors":      5,
				"construction_year": 1700, // Too old
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request data",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			w := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", tc.request)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"].(string), tc.expectedError)
		})
	}
}

func (suite *BuildingIntegrationTestSuite) TestMetadataValidation() {
	testCases := []struct {
		name           string
		buildingType   models.BuildingType
		metadata       models.BuildingMetadata
		expectedStatus int
	}{
		{
			name:         "Valid residential metadata",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities":               []string{"gym", "swimming_pool"},
				"security_type":           "24_hour_guard",
				"maintenance_staff_count": 3,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:         "Valid commercial metadata",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"parking_spaces": map[string]interface{}{
					"total":    100,
					"customer": 80,
					"staff":    20,
				},
				"loading_docks": 3,
				"business_hours": map[string]interface{}{
					"weekdays": "09:00-22:00",
					"weekends": "10:00-23:00",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:         "Invalid residential metadata",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"parking_spaces": 100, // Commercial field in residential building
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			buildingReq := &models.CreateBuildingRequest{
				PropertyID:   suite.testProperty.ID,
				BuildingName: "Metadata Test Building",
				BuildingCode: "MTB" + strconv.Itoa(int(time.Now().UnixNano()%1000)),
				BuildingType: tc.buildingType,
				TotalFloors:  5,
				Metadata:     tc.metadata,
			}

			w := suite.makeAuthenticatedRequest("POST", "/api/v1/buildings", buildingReq)
			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

// Test Authentication and Authorization
func (suite *BuildingIntegrationTestSuite) TestUnauthorizedAccess() {
	// Test without authentication token
	req := httptest.NewRequest("GET", "/api/v1/buildings", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *BuildingIntegrationTestSuite) TestInvalidAuthToken() {
	// Test with invalid authentication token
	req := httptest.NewRequest("GET", "/api/v1/buildings", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// Helper methods
func (suite *BuildingIntegrationTestSuite) createTestBuilding(name, code string) *models.Building {
	return suite.createTestBuildingWithType(name, code, models.BuildingTypeResidential)
}

func (suite *BuildingIntegrationTestSuite) createTestBuildingWithType(name, code string, buildingType models.BuildingType) *models.Building {
	return suite.createTestBuildingWithDetails(name, code, buildingType, 5, false)
}

func (suite *BuildingIntegrationTestSuite) createTestBuildingWithDetails(name, code string, buildingType models.BuildingType, floors int, hasElevator bool) *models.Building {
	building := &models.Building{
		PropertyID:   suite.testProperty.ID,
		BuildingName: name,
		BuildingCode: code,
		BuildingType: buildingType,
		TotalFloors:  floors,
		HasElevator:  hasElevator,
		ActiveStatus: true,
		Metadata:     models.BuildingMetadata{},
	}

	err := suite.buildingRepo.Create(building)
	require.NoError(suite.T(), err)
	return building
}

func (suite *BuildingIntegrationTestSuite) createTestBuildingWithMetadata(name, code string, buildingType models.BuildingType, metadata map[string]interface{}) *models.Building {
	building := &models.Building{
		PropertyID:   suite.testProperty.ID,
		BuildingName: name,
		BuildingCode: code,
		BuildingType: buildingType,
		TotalFloors:  5,
		HasElevator:  false,
		ActiveStatus: true,
		Metadata:     models.BuildingMetadata(metadata),
	}

	err := suite.buildingRepo.Create(building)
	require.NoError(suite.T(), err)
	return building
}

// Utility functions
func getTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	// Use the main database for testing if test database is not available
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:password@localhost:5432/tenantly?sslmode=disable"
}

func intPtrComprehensive(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// Run the test suite
func TestBuildingIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(BuildingIntegrationTestSuite))
}
