package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
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

// LeaseIntegrationTestSuite provides comprehensive integration testing for lease management API
type LeaseIntegrationTestSuite struct {
	suite.Suite
	db            *sqlx.DB
	router        *gin.Engine
	config        *config.Config
	leaseHandler  *LeaseHandler
	tenantHandler *TenantHandler
	propertyRepo  *repositories.PropertyRepository
	buildingRepo  *repositories.BuildingRepository
	unitRepo      *repositories.UnitRepository
	tenantRepo    *repositories.TenantRepository
	leaseRepo     *repositories.LeaseRepository
	userRepo      *repositories.UserRepository
	orgRepo       *repositories.OrganizationRepository
	testProperty  *models.Property
	testBuilding  *models.Building
	testUnit      *models.Unit
	testTenant    *models.Tenant
	testUser      *models.User
	testOrg       *models.Organization
	authToken     string
}

func (suite *LeaseIntegrationTestSuite) SetupSuite() {
	// Set test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	nidKey := sha256.Sum256([]byte("lease-integration-test-key"))
	nidProtector, err := appcrypto.NewNIDProtector(nidKey[:], []byte("lease-integration-test-pepper"))
	require.NoError(suite.T(), err)

	suite.config = &config.Config{
		DatabaseURL:   getLeaseTestDatabaseURL(),
		JWTSecret:     "test-jwt-secret-key",
		JWTExpiration: time.Hour * 24,
		Environment:   "test",
		NIDProtector:  nidProtector,
	}

	// Initialize test database
	suite.db, err = database.Connect(suite.config.DatabaseURL)
	if err != nil {
		suite.T().Skipf("Skipping integration test: PostgreSQL not available: %v", err)
		return
	}

	// Run migrations
	err = database.RunMigrations(suite.config.DatabaseURL)
	require.NoError(suite.T(), err, "Failed to run migrations")

	// Initialize repositories
	suite.orgRepo = repositories.NewOrganizationRepository(suite.db)
	suite.userRepo = repositories.NewUserRepository(suite.db)
	suite.propertyRepo = repositories.NewPropertyRepository(suite.db)
	suite.buildingRepo = repositories.NewBuildingRepository(suite.db)
	suite.unitRepo = repositories.NewUnitRepository(suite.db)
	suite.tenantRepo = repositories.NewTenantRepository(suite.db, suite.config.NIDProtector)
	suite.leaseRepo = repositories.NewLeaseRepository(suite.db)

	// Initialize services
	auditService := database.NewAuditService(suite.db)
	userService := services.NewUserService(suite.userRepo, auditService, suite.config.JWTSecret, suite.config.JWTExpiration)
	tenantService := services.NewTenantService(suite.tenantRepo, suite.leaseRepo, auditService)
	leaseService := services.NewLeaseService(suite.leaseRepo, suite.tenantRepo, suite.unitRepo, auditService)

	// Initialize handlers
	userHandler := NewUserHandler(userService, "", false)
	suite.tenantHandler = NewTenantHandler(tenantService)
	suite.leaseHandler = NewLeaseHandler(leaseService)

	// Setup router with middleware
	suite.router = gin.New()
	suite.setupTestRoutes(userHandler, auditService)

	// Create test data
	suite.createTestData()
}

func (suite *LeaseIntegrationTestSuite) TearDownSuite() {
	// Clean up test data
	suite.cleanupTestData()

	// Close database connection
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *LeaseIntegrationTestSuite) SetupTest() {
	// Clean up any test-specific data before each test
	suite.cleanupLeaseTestData()
}

func (suite *LeaseIntegrationTestSuite) setupTestRoutes(userHandler *UserHandler, auditService *database.AuditService) {
	// Add middleware
	suite.router.Use(middleware.SecurityHeadersMiddleware())
	suite.router.Use(middleware.CORS(suite.config.AllowedOrigins))
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
			// Tenant routes
			tenants := protected.Group("/tenants")
			{
				tenants.GET("", suite.tenantHandler.GetAllTenants)
				tenants.POST("", suite.tenantHandler.CreateTenant)
				tenants.GET("/:id", suite.tenantHandler.GetTenantByID)
				tenants.PUT("/:id", suite.tenantHandler.UpdateTenant)
				tenants.DELETE("/:id", suite.tenantHandler.DeleteTenant)
			}

			// Lease routes
			leases := protected.Group("/leases")
			{
				leases.GET("", suite.leaseHandler.GetAllLeases)
				leases.POST("", suite.leaseHandler.CreateLease)
				leases.GET("/:id", suite.leaseHandler.GetLeaseByID)
				leases.PUT("/:id", suite.leaseHandler.UpdateLease)
				leases.DELETE("/:id", suite.leaseHandler.DeleteLease)
				leases.POST("/:id/terminate", suite.leaseHandler.TerminateLease)
				leases.GET("/units/:unit_id", suite.leaseHandler.GetLeasesByUnit)
				leases.GET("/tenants/:tenant_id", suite.leaseHandler.GetLeasesByTenant)
			}
		}
	}
}

func (suite *LeaseIntegrationTestSuite) createTestData() {
	// Create test organization
	suite.testOrg = &models.Organization{
		Name:             "Test Org",
		Slug:             "test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         10,
		Active:           true,
	}
	err := suite.orgRepo.Create(suite.testOrg)
	require.NoError(suite.T(), err)

	// Create test user
	suite.testUser = &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "Admin",
		Active:       true,
	}
	err = suite.userRepo.Create(suite.testUser)
	require.NoError(suite.T(), err)

	// Link user to organization
	orgRoleRepo := repositories.NewUserOrganizationRoleRepository(suite.db)
	err = orgRoleRepo.Upsert(suite.testUser.ID, suite.testOrg.ID, "Admin")
	require.NoError(suite.T(), err)

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
	require.NoError(suite.T(), err)

	// Create test building
	building := &models.Building{
		PropertyID:   suite.testProperty.ID,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		ActiveStatus: true,
	}
	err = suite.buildingRepo.Create(building)
	require.NoError(suite.T(), err)
	suite.testBuilding = building

	// Create test unit
	unitReq := &models.CreateUnitRequest{
		BuildingID: suite.testBuilding.ID,
		PropertyID: suite.testProperty.ID,
		UnitNumber: "101",
		UnitType:   models.UnitTypeApartment,
		Floor:      1,
	}
	suite.testUnit, err = suite.unitRepo.Create(unitReq, suite.testOrg.ID)
	require.NoError(suite.T(), err)

	// Create test tenant
	tenantReq := &models.CreateTenantRequest{
		Name:           "John Doe",
		TenantType:     models.TenantTypeIndividual,
		PhoneNumber:    "1234567890",
		Email:          "john@example.com",
		NIDNumber:      "NID123",
		Address:        "123 Main St",
		OrganizationID: suite.testOrg.ID,
	}
	suite.testTenant, err = suite.tenantRepo.Create(tenantReq)
	require.NoError(suite.T(), err)
}

func (suite *LeaseIntegrationTestSuite) cleanupTestData() {
	// Clean up in reverse order of creation
	if suite.testUnit != nil {
		suite.unitRepo.Delete(suite.testUnit.ID)
	}
	if suite.testBuilding != nil {
		suite.buildingRepo.SoftDelete(suite.testBuilding.ID)
	}
	if suite.testProperty != nil {
		suite.propertyRepo.Delete(suite.testProperty.ID)
	}
	if suite.testTenant != nil {
		suite.tenantRepo.Update(suite.testTenant.ID, map[string]interface{}{"active": false})
	}
	if suite.testUser != nil {
		suite.userRepo.Delete(suite.testUser.ID)
	}
	if suite.testOrg != nil {
		suite.orgRepo.Delete(suite.testOrg.ID)
	}
}

func (suite *LeaseIntegrationTestSuite) cleanupLeaseTestData() {
	// Clean up leases created during tests
	leases, _, _ := suite.leaseRepo.GetAll(1, 100, suite.testOrg.ID)
	for _, lease := range leases {
		suite.leaseRepo.Delete(lease.ID)
	}
}

func (suite *LeaseIntegrationTestSuite) generateAuthToken() string {
	loginReq := map[string]string{
		"username": suite.testUser.Username,
		"password": "password",
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

	return "Bearer test-token"
}

func (suite *LeaseIntegrationTestSuite) makeAuthenticatedRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// Test Lease CRUD Operations
func (suite *LeaseIntegrationTestSuite) TestCreateLease_Success() {
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	leaseReq := &models.CreateLeaseRequest{
		UnitID:          suite.testUnit.ID,
		TenantID:        suite.testTenant.ID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Now().Format("2006-01-02"),
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  suite.testOrg.ID,
	}

	w := suite.makeAuthenticatedRequest("POST", "/api/v1/leases", leaseReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
}

func (suite *LeaseIntegrationTestSuite) TestGetLeaseByID_Success() {
	// Create a lease first
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/leases/%d", lease.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	leaseData := response[""].(map[string]interface{})
	assert.Equal(suite.T(), float64(lease.ID), leaseData["id"])
}

func (suite *LeaseIntegrationTestSuite) TestGetLeaseByID_NotFound() {
	w := suite.makeAuthenticatedRequest("GET", "/api/v1/leases/99999", nil)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Lease not found", response["error"])
}

func (suite *LeaseIntegrationTestSuite) TestGetAllLeases_Success() {
	// Create multiple leases
	for i := 0; i < 3; i++ {
		suite.createTestLease()
	}

	w := suite.makeAuthenticatedRequest("GET", "/api/v1/leases?page=1&page_size=10", nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["leases"])
	assert.NotNil(suite.T(), response["pagination"])

	leases := response["leases"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(leases), 3)
}

func (suite *LeaseIntegrationTestSuite) TestUpdateLease_Success() {
	// Create a lease first
	lease := suite.createTestLease()

	newRent := 20000.0
	newDuration := 24
	updateReq := map[string]interface{}{
		"monthly_rent":    newRent,
		"duration_months": newDuration,
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/leases/%d", lease.ID), updateReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify update in database
	updatedLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newRent, updatedLease.MonthlyRent)
	assert.Equal(suite.T(), newDuration, updatedLease.DurationMonths)
}

func (suite *LeaseIntegrationTestSuite) TestDeleteLease_Success() {
	// Create a lease with future start date
	futureStartDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	endDate := time.Now().AddDate(2, 0, 0).Format("2006-01-02")

	lease := suite.createTestLeaseWithDates(futureStartDate, endDate)

	w := suite.makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/v1/leases/%d", lease.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Lease deleted successfully", response["message"])

	// Verify deletion
	_, err = suite.leaseRepo.GetByID(lease.ID)
	assert.Error(suite.T(), err)
}

func (suite *LeaseIntegrationTestSuite) TestTerminateLease_Success() {
	// Create a lease first
	lease := suite.createTestLease()

	terminationDate := time.Now().AddDate(6, 0, 0).Format("2006-01-02")
	terminateReq := map[string]string{
		"termination_date": terminationDate,
	}

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/terminate", lease.ID), terminateReq)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Lease terminated successfully", response["message"])

	// Verify termination
	terminatedLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), terminatedLease.Active)
}

func (suite *LeaseIntegrationTestSuite) TestGetLeasesByUnit_Success() {
	// Create a lease first
	_ = suite.createTestLease()

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/leases/units/%d?page=1&page_size=10", suite.testUnit.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	leases := response["leases"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(leases), 1)
}

func (suite *LeaseIntegrationTestSuite) TestGetLeasesByTenant_Success() {
	// Create a lease first
	_ = suite.createTestLease()

	w := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/leases/tenants/%d?page=1&page_size=10", suite.testTenant.ID), nil)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	leases := response["leases"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(leases), 1)
}

// Test Lease Validation and Error Handling
func (suite *LeaseIntegrationTestSuite) TestUnauthorizedAccess() {
	req := httptest.NewRequest("GET", "/api/v1/leases", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// Helper methods
func (suite *LeaseIntegrationTestSuite) createTestLease() *models.Lease {
	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	return suite.createTestLeaseWithDates(startDate, endDate)
}

func (suite *LeaseIntegrationTestSuite) createTestLeaseWithDates(startDate, endDate string) *models.Lease {
	leaseReq := &models.CreateLeaseRequest{
		UnitID:          suite.testUnit.ID,
		TenantID:        suite.testTenant.ID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       startDate,
		EndDate:         &endDate,
		DurationMonths:  12,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  suite.testOrg.ID,
	}

	lease, err := suite.leaseRepo.Create(leaseReq)
	require.NoError(suite.T(), err)
	return lease
}

// Utility functions
func getLeaseTestDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:password@localhost:5432/tenantly?sslmode=disable"
}

// Run the test suite
func TestLeaseIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LeaseIntegrationTestSuite))
}
