package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	"github.com/ysnarafat/tenantly/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

// LeaseIntegrationTestSuite provides comprehensive integration testing for lease management API
type LeaseIntegrationTestSuite struct {
	suite.Suite
	db              *sqlx.DB
	router          *gin.Engine
	config          *config.Config
	leaseHandler    *LeaseHandler
	tenantHandler   *TenantHandler
	propertyRepo    *repositories.PropertyRepository
	buildingRepo    *repositories.BuildingRepository
	unitRepo        *repositories.UnitRepository
	tenantRepo      *repositories.TenantRepository
	leaseRepo       *repositories.LeaseRepository
	leaseChargeRepo *repositories.LeaseChargeRepository
	userRepo        *repositories.UserRepository
	orgRepo         *repositories.OrganizationRepository
	testProperty    *models.Property
	testBuilding    *models.Building
	testUnit        *models.Unit
	testTenant      *models.Tenant
	testUser        *models.User
	testOrg         *models.Organization
	authToken       string
}

func (suite *LeaseIntegrationTestSuite) SetupSuite() {
	// Set test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	nidKey := sha256.Sum256([]byte("lease-integration-test-key"))
	nidProtector, err := appcrypto.NewNIDProtector(nidKey[:], []byte("lease-integration-test-pepper"))
	require.NoError(suite.T(), err)

	databaseURL, err := testutil.EnsureTestDatabase(handlersTestDB)
	if err != nil {
		suite.T().Skipf("Skipping integration test: PostgreSQL not available: %v", err)
		return
	}

	suite.config = &config.Config{
		DatabaseURL:   databaseURL,
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

	// Reset rather than just migrate: this suite seeds fixed slugs, usernames,
	// and property codes, which collide with whatever the previous run left
	// behind. testutil resolves the migrations directory relative to the
	// package under test; database.RunMigrations resolves it against the
	// process working directory and so only works from the module root.
	err = testutil.ResetSchema(suite.config.DatabaseURL)
	require.NoError(suite.T(), err, "Failed to run migrations")

	// Initialize repositories
	suite.orgRepo = repositories.NewOrganizationRepository(suite.db)
	suite.userRepo = repositories.NewUserRepository(suite.db)
	suite.propertyRepo = repositories.NewPropertyRepository(suite.db)
	suite.buildingRepo = repositories.NewBuildingRepository(suite.db)
	suite.unitRepo = repositories.NewUnitRepository(suite.db)
	suite.tenantRepo = repositories.NewTenantRepository(suite.db, suite.config.NIDProtector)
	suite.leaseRepo = repositories.NewLeaseRepository(suite.db)
	suite.leaseChargeRepo = repositories.NewLeaseChargeRepository(suite.db)

	// Initialize services
	auditService := database.NewAuditService(suite.db)
	userService := services.NewUserService(suite.userRepo, auditService, suite.config.JWTSecret, suite.config.JWTExpiration)
	tenantService := services.NewTenantService(suite.tenantRepo, suite.leaseRepo, auditService)
	leaseService := services.NewLeaseService(suite.leaseRepo, suite.tenantRepo, suite.unitRepo, suite.leaseChargeRepo, auditService)

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
		_ = suite.db.Close()
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

		// Protected routes. RequireOrgContext mirrors the real server: every
		// org-scoped handler reads org_id, which only this middleware sets.
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired(suite.config.JWTSecret, auditService))
		protected.Use(middleware.RequireOrgContext())
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
				leases.POST("/:id/renew", suite.leaseHandler.RenewLease)
				leases.POST("/:id/charges", suite.leaseHandler.AddLeaseCharge)
				leases.PUT("/:id/charges/:chargeId", suite.leaseHandler.UpdateLeaseCharge)
				leases.DELETE("/:id/charges/:chargeId", suite.leaseHandler.DeleteLeaseCharge)
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
	// The hash must be a real bcrypt digest of the password generateAuthToken
	// logs in with — Login compares them with bcrypt, so a placeholder string
	// leaves every authenticated request a 401.
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	require.NoError(suite.T(), err)

	suite.testUser = &models.User{
		Username:       "testuser",
		Email:          "test@example.com",
		PasswordHash:   string(passwordHash),
		Role:           "Admin",
		Active:         true,
		Status:         "active",
		OrganizationID: &suite.testOrg.ID,
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
		PropertyName:   "Test Property",
		PropertyCode:   "TEST001",
		PropertyType:   models.PropertyTypeCommercial,
		Address:        "123 Test Street",
		City:           "Test City",
		OrganizationID: suite.testOrg.ID,
	}
	suite.testProperty, err = suite.propertyRepo.Create(propertyReq)
	require.NoError(suite.T(), err)

	// Create test building
	building := &models.Building{
		PropertyID:     suite.testProperty.ID,
		OrganizationID: suite.testOrg.ID,
		BuildingName:   "Test Building",
		BuildingCode:   "TB001",
		BuildingType:   models.BuildingTypeResidential,
		TotalFloors:    5,
		ActiveStatus:   true,
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
		_ = suite.unitRepo.Delete(suite.testUnit.ID)
	}
	if suite.testBuilding != nil {
		_ = suite.buildingRepo.SoftDelete(suite.testBuilding.ID)
	}
	if suite.testProperty != nil {
		_ = suite.propertyRepo.Delete(suite.testProperty.ID)
	}
	if suite.testTenant != nil {
		_ = suite.tenantRepo.Update(suite.testTenant.ID, map[string]interface{}{"active": false})
	}
	if suite.testUser != nil {
		_ = suite.userRepo.Delete(suite.testUser.ID)
	}
	if suite.testOrg != nil {
		_ = suite.orgRepo.Delete(suite.testOrg.ID)
	}
}

func (suite *LeaseIntegrationTestSuite) cleanupLeaseTestData() {
	// Clean up leases created during tests
	leases, _, _ := suite.leaseRepo.GetAll(1, 100, suite.testOrg.ID)
	for _, lease := range leases {
		_ = suite.leaseRepo.Delete(lease.ID)
	}
}

func (suite *LeaseIntegrationTestSuite) generateAuthToken() string {
	loginReq := map[string]string{
		"username": suite.testUser.Username,
		"password": testUserPassword,
	}

	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Fail here rather than returning a placeholder token: a bad token turns
	// every assertion in the suite into an indistinguishable 401.
	require.Equal(suite.T(), http.StatusOK, w.Code, "login failed: %s", w.Body.String())

	var response map[string]interface{}
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &response))

	token, ok := response["token"].(string)
	require.True(suite.T(), ok, "login response carried no token: %s", w.Body.String())

	return "Bearer " + token
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

	// GetLeaseByID returns the lease itself, not a wrapper object.
	assert.Equal(suite.T(), float64(lease.ID), response["id"])
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

// TestUpdateLease_StartDateOnlyRecalculatesEndDate covers correcting a
// data-entry mistake on just the start date (e.g. it was set wrong when the
// lease was created) — end_date must still shift using the lease's existing
// duration_months, not go stale, even though duration_months itself isn't
// part of this request.
func (suite *LeaseIntegrationTestSuite) TestUpdateLease_StartDateOnlyRecalculatesEndDate() {
	lease := suite.createTestLease() // duration_months: 12

	correctedStart := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	updateReq := map[string]interface{}{
		"start_date": correctedStart,
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/leases/%d", lease.ID), updateReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	updatedLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 12, updatedLease.DurationMonths, "duration_months must be unchanged")

	wantEndDate := time.Now().AddDate(0, -1, 0).AddDate(0, 12, 0).Format("2006-01-02")
	assert.Equal(suite.T(), wantEndDate, updatedLease.EndDate.Format("2006-01-02"),
		"end_date must be recalculated from the corrected start_date + existing duration_months")
}

// TestUpdateLease_ExplicitEndDateOverridesDuration covers correcting a lease
// onto an exact end date that doesn't land on a whole-month boundary —
// duration_months (an integer column) can never express that precisely, so
// an explicit end_date must win outright rather than being derived from it.
func (suite *LeaseIntegrationTestSuite) TestUpdateLease_ExplicitEndDateOverridesDuration() {
	lease := suite.createTestLease() // start_date: today, duration_months: 12

	// 45 days out — not a whole number of months from start_date.
	correctedEnd := time.Now().AddDate(0, 0, 45).Format("2006-01-02")
	updateReq := map[string]interface{}{
		"end_date": correctedEnd,
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/leases/%d", lease.ID), updateReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	updatedLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), correctedEnd, updatedLease.EndDate.Format("2006-01-02"),
		"end_date must be set exactly as requested, not rounded onto a whole-month boundary")
	// duration_months is derived only for display — 45 days rounds to 1 month
	// (30.44-day average), not the original 12.
	assert.Equal(suite.T(), 1, updatedLease.DurationMonths)
}

// TestUpdateLease_RoundTrippingEndDateDoesNotDriftDuration covers an edit
// that resends the lease's own current start/end dates unchanged (e.g. the
// UI submits them alongside an unrelated field like monthly_rent) — the
// derived duration_months must reproduce the exact existing value, not drift
// down from averaging error (a fixed 30.44-day-per-month approximation would
// undercount an exact whole-month span).
func (suite *LeaseIntegrationTestSuite) TestUpdateLease_RoundTrippingEndDateDoesNotDriftDuration() {
	startDate := time.Now().Format("2006-01-02")
	endDate := time.Now().AddDate(0, 6, 0).Format("2006-01-02") // exactly 6 calendar months
	leaseReq := &models.CreateLeaseRequest{
		UnitID:          suite.testUnit.ID,
		TenantID:        suite.testTenant.ID,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       startDate,
		EndDate:         &endDate,
		DurationMonths:  6,
		MonthlyRent:     15000,
		SecurityDeposit: 30000,
		OrganizationID:  suite.testOrg.ID,
	}
	lease, err := suite.leaseRepo.Create(leaseReq)
	require.NoError(suite.T(), err)

	updateReq := map[string]interface{}{
		"start_date":   startDate,
		"end_date":     endDate,
		"monthly_rent": 16000.0,
	}
	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/leases/%d", lease.ID), updateReq)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	updatedLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 6, updatedLease.DurationMonths,
		"duration_months must round-trip exactly, not drift from date-averaging error")
}

// TestUpdateLease_EndDateBeforeStartDateRejected covers the validation
// guarding against an obviously-inverted correction.
func (suite *LeaseIntegrationTestSuite) TestUpdateLease_EndDateBeforeStartDateRejected() {
	lease := suite.createTestLease()

	pastEnd := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	updateReq := map[string]interface{}{
		"end_date": pastEnd,
	}

	w := suite.makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/v1/leases/%d", lease.ID), updateReq)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
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
	// The actual move-out date and reason must be persisted, not just the
	// active flag — otherwise the lease record keeps its original,
	// now-inaccurate end_date.
	assert.Equal(suite.T(), terminationDate, terminatedLease.EndDate.Format("2006-01-02"))
	if assert.NotNil(suite.T(), terminatedLease.EndReason) {
		assert.Equal(suite.T(), models.LeaseEndReasonTerminated, *terminatedLease.EndReason)
	}
}

func (suite *LeaseIntegrationTestSuite) TestRenewLease_Success() {
	// Create a lease first
	lease := suite.createTestLease()

	renewReq := map[string]interface{}{
		"duration_months": 6,
		"monthly_rent":    17000,
	}

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/renew", lease.ID), renewReq)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var renewed models.LeaseWithDetails
	err := json.Unmarshal(w.Body.Bytes(), &renewed)
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), lease.ID, renewed.ID)
	assert.True(suite.T(), renewed.Active)
	assert.Equal(suite.T(), 17000.0, renewed.MonthlyRent)
	assert.Equal(suite.T(), 6, renewed.DurationMonths)
	if assert.NotNil(suite.T(), renewed.RenewedFromLeaseID) {
		assert.Equal(suite.T(), lease.ID, *renewed.RenewedFromLeaseID)
	}

	// The original lease must be closed out, not left dangling as still active
	// — otherwise the unit would appear to have two active leases at once.
	oldLease, err := suite.leaseRepo.GetByID(lease.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), oldLease.Active)
	if assert.NotNil(suite.T(), oldLease.EndReason) {
		assert.Equal(suite.T(), models.LeaseEndReasonRenewed, *oldLease.EndReason)
	}
	// Coverage must be back-to-back: the old lease's new end_date should
	// match the renewed lease's start_date, with no gap or overlap.
	assert.Equal(suite.T(), oldLease.EndDate.Format("2006-01-02"), renewed.StartDate.Format("2006-01-02"))
}

func (suite *LeaseIntegrationTestSuite) TestRenewLease_InactiveLeaseFails() {
	lease := suite.createTestLease()

	// End the tenancy first, so the lease is no longer active.
	w := suite.makeAuthenticatedRequest(
		"POST", fmt.Sprintf("/api/v1/leases/%d/terminate", lease.ID),
		map[string]string{"termination_date": time.Now().Format("2006-01-02")},
	)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Renewing it must fail with a specific, actionable reason — not a
	// generic "Failed to renew lease" that hides why.
	w = suite.makeAuthenticatedRequest(
		"POST", fmt.Sprintf("/api/v1/leases/%d/renew", lease.ID),
		map[string]interface{}{"duration_months": 6},
	)
	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "RENEW_LEASE_INACTIVE", response["code"])
	assert.Equal(suite.T(), "only an active lease can be renewed", response["error"])
}

func (suite *LeaseIntegrationTestSuite) TestTerminateLease_AlreadyTerminatedFails() {
	lease := suite.createTestLease()

	terminationDate := time.Now().Format("2006-01-02")
	w := suite.makeAuthenticatedRequest(
		"POST", fmt.Sprintf("/api/v1/leases/%d/terminate", lease.ID),
		map[string]string{"termination_date": terminationDate},
	)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Terminating an already-terminated lease must fail with a specific
	// reason rather than a generic, unhelpful message.
	w = suite.makeAuthenticatedRequest(
		"POST", fmt.Sprintf("/api/v1/leases/%d/terminate", lease.ID),
		map[string]string{"termination_date": terminationDate},
	)
	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "TERMINATE_LEASE_INACTIVE", response["code"])
	assert.Equal(suite.T(), "lease is not active", response["error"])
}

func (suite *LeaseIntegrationTestSuite) TestAddLeaseCharge_Success() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Utility",
		"label":       "Electricity",
		"amount":      500,
	})
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var charge models.LeaseCharge
	err := json.Unmarshal(w.Body.Bytes(), &charge)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), lease.ID, charge.LeaseID)
	assert.Equal(suite.T(), models.ChargeTypeUtility, charge.ChargeType)
	assert.Equal(suite.T(), "Electricity", charge.Label)
	assert.Equal(suite.T(), 500.0, charge.Amount)
	assert.True(suite.T(), charge.Active)
}

func (suite *LeaseIntegrationTestSuite) TestAddLeaseCharge_InvalidChargeTypeRejected() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "NotARealType",
		"label":       "Bogus",
		"amount":      100,
	})
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *LeaseIntegrationTestSuite) TestUpdateLeaseCharge_Success() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Utility",
		"label":       "Electricity",
		"amount":      500,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var created models.LeaseCharge
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created))

	w = suite.makeAuthenticatedRequest(
		"PUT", fmt.Sprintf("/api/v1/leases/%d/charges/%d", lease.ID, created.ID),
		map[string]interface{}{"amount": 650},
	)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updated models.LeaseCharge
	assert.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(suite.T(), 650.0, updated.Amount)
	// Fields not included in the update request must be left untouched.
	assert.Equal(suite.T(), "Electricity", updated.Label)
	assert.Equal(suite.T(), models.ChargeTypeUtility, updated.ChargeType)
}

func (suite *LeaseIntegrationTestSuite) TestDeleteLeaseCharge_Success() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Maintenance",
		"label":       "Lift maintenance",
		"amount":      200,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var created models.LeaseCharge
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &created))

	w = suite.makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/v1/leases/%d/charges/%d", lease.ID, created.ID), nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	charges, err := suite.leaseChargeRepo.GetByLeaseID(lease.ID)
	assert.NoError(suite.T(), err)
	for _, c := range charges {
		assert.NotEqual(suite.T(), created.ID, c.ID, "deleted charge must not still exist")
	}
}

func (suite *LeaseIntegrationTestSuite) TestAddLeaseCharge_CrossOrgLeaseNotFound() {
	// A lease belonging to a *different* organization must not be reachable
	// through this org's authenticated session — otherwise one org could
	// tamper with another org's billing by adding charges to its leases.
	// Built directly via the suite's own repos (rather than the
	// testutil.CreateTest* helpers, which hardcode property/building codes
	// that collide with this suite's own shared fixtures) with unique codes.
	otherOrg := &models.Organization{Name: "Other Org", Slug: "other-org-lease-charges", SubscriptionTier: models.TierBasic}
	require.NoError(suite.T(), suite.orgRepo.Create(otherOrg))

	otherProperty, err := suite.propertyRepo.Create(&models.CreatePropertyRequest{
		PropertyName:   "Other Org Property",
		PropertyCode:   "OTHERORG001",
		PropertyType:   models.PropertyTypeCommercial,
		Address:        "1 Other Street",
		City:           "Other City",
		OrganizationID: otherOrg.ID,
	})
	require.NoError(suite.T(), err)

	otherBuilding := &models.Building{
		PropertyID:     otherProperty.ID,
		OrganizationID: otherOrg.ID,
		BuildingName:   "Other Org Building",
		BuildingCode:   "OOB001",
		BuildingType:   models.BuildingTypeResidential,
		TotalFloors:    3,
		ActiveStatus:   true,
	}
	require.NoError(suite.T(), suite.buildingRepo.Create(otherBuilding))

	otherUnit, err := suite.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: otherBuilding.ID,
		PropertyID: otherProperty.ID,
		UnitNumber: "101",
		UnitType:   models.UnitTypeApartment,
		Floor:      1,
	}, otherOrg.ID)
	require.NoError(suite.T(), err)

	otherTenant, err := suite.tenantRepo.Create(&models.CreateTenantRequest{
		Name:           "Other Org Tenant",
		TenantType:     models.TenantTypeIndividual,
		PhoneNumber:    "9999999999",
		Email:          "other-org-tenant@example.com",
		NIDNumber:      "NID-OTHERORG",
		Address:        "1 Other Street",
		OrganizationID: otherOrg.ID,
	})
	require.NoError(suite.T(), err)

	otherLease, err := suite.leaseRepo.Create(&models.CreateLeaseRequest{
		UnitID:         otherUnit.ID,
		TenantID:       otherTenant.ID,
		LeaseType:      models.LeaseTypeResidential,
		StartDate:      time.Now().Format("2006-01-02"),
		DurationMonths: 12,
		MonthlyRent:    5000,
		OrganizationID: otherOrg.ID,
	})
	require.NoError(suite.T(), err)

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", otherLease.ID), map[string]interface{}{
		"charge_type": "Utility",
		"label":       "Electricity",
		"amount":      500,
	})
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	charges, err := suite.leaseChargeRepo.GetByLeaseID(otherLease.ID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), charges, "charge must not have been added to another org's lease")
}

func (suite *LeaseIntegrationTestSuite) TestRenewLease_CarriesForwardActiveCharges() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Utility",
		"label":       "Electricity",
		"amount":      500,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)

	// A discontinued charge must NOT carry forward into the renewed lease.
	w = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Parking",
		"label":       "Parking (discontinued)",
		"amount":      300,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var discontinued models.LeaseCharge
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &discontinued))
	w = suite.makeAuthenticatedRequest(
		"PUT", fmt.Sprintf("/api/v1/leases/%d/charges/%d", lease.ID, discontinued.ID),
		map[string]interface{}{"active": false},
	)
	require.Equal(suite.T(), http.StatusOK, w.Code)

	w = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/renew", lease.ID), map[string]interface{}{
		"duration_months": 6,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var renewed models.LeaseWithDetails
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &renewed))

	newCharges, err := suite.leaseChargeRepo.GetByLeaseID(renewed.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), newCharges, 1, "only the one active charge should carry forward")
	if len(newCharges) == 1 {
		assert.Equal(suite.T(), "Electricity", newCharges[0].Label)
		assert.Equal(suite.T(), 500.0, newCharges[0].Amount)
		assert.True(suite.T(), newCharges[0].Active)
	}
}

func (suite *LeaseIntegrationTestSuite) TestRenewLease_CanOptOutOfCarryingForwardCharges() {
	lease := suite.createTestLease()

	w := suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/charges", lease.ID), map[string]interface{}{
		"charge_type": "Utility",
		"label":       "Electricity",
		"amount":      500,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)

	w = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/leases/%d/renew", lease.ID), map[string]interface{}{
		"duration_months":       6,
		"carry_forward_charges": false,
	})
	require.Equal(suite.T(), http.StatusCreated, w.Code)
	var renewed models.LeaseWithDetails
	require.NoError(suite.T(), json.Unmarshal(w.Body.Bytes(), &renewed))

	newCharges, err := suite.leaseChargeRepo.GetByLeaseID(renewed.ID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), newCharges, "charges must not carry forward when explicitly opted out")
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

// Run the test suite
func TestLeaseIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LeaseIntegrationTestSuite))
}
