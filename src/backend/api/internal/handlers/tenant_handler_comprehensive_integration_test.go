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

// TenantIntegrationTestSuite exercises the full Tenant API via a real PostgreSQL
// database: every test issues an actual HTTP request through the Gin router,
// checks the HTTP status code, and asserts exact response-body values.
type TenantIntegrationTestSuite struct {
	suite.Suite
	db            *sqlx.DB
	router        *gin.Engine
	cfg           *config.Config
	tenantHandler *TenantHandler
	leaseHandler  *LeaseHandler
	orgRepo       *repositories.OrganizationRepository
	userRepo      *repositories.UserRepository
	propertyRepo  *repositories.PropertyRepository
	buildingRepo  *repositories.BuildingRepository
	unitRepo      *repositories.UnitRepository
	tenantRepo    *repositories.TenantRepository
	leaseRepo     *repositories.LeaseRepository
	testOrg       *models.Organization
	testUser      *models.User
	testProperty  *models.Property
	testBuilding  *models.Building
	testUnit      *models.Unit
	authToken     string
}

func (s *TenantIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	nidKey := sha256.Sum256([]byte("tenant-integration-test-key"))
	nidProtector, err := appcrypto.NewNIDProtector(nidKey[:], []byte("tenant-integration-test-pepper"))
	require.NoError(s.T(), err)

	dbURL, err := testutil.EnsureTestDatabase(handlersTestDB)
	if err != nil {
		s.T().Skipf("Skipping: PostgreSQL not available: %v", err)
		return
	}

	s.cfg = &config.Config{
		DatabaseURL:   dbURL,
		JWTSecret:     "test-jwt-secret-key",
		JWTExpiration: time.Hour * 24,
		Environment:   "test",
		NIDProtector:  nidProtector,
	}

	s.db, err = database.Connect(s.cfg.DatabaseURL)
	if err != nil {
		s.T().Skipf("Skipping: PostgreSQL not available: %v", err)
		return
	}

	require.NoError(s.T(), testutil.ResetSchema(s.cfg.DatabaseURL))

	s.orgRepo = repositories.NewOrganizationRepository(s.db)
	s.userRepo = repositories.NewUserRepository(s.db)
	s.propertyRepo = repositories.NewPropertyRepository(s.db)
	s.buildingRepo = repositories.NewBuildingRepository(s.db)
	s.unitRepo = repositories.NewUnitRepository(s.db)
	s.tenantRepo = repositories.NewTenantRepository(s.db, s.cfg.NIDProtector)
	s.leaseRepo = repositories.NewLeaseRepository(s.db)
	leaseChargeRepo := repositories.NewLeaseChargeRepository(s.db)

	auditSvc := database.NewAuditService(s.db)
	userSvc := services.NewUserService(s.userRepo, auditSvc, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	tenantSvc := services.NewTenantService(s.tenantRepo, s.leaseRepo, auditSvc)
	leaseSvc := services.NewLeaseService(s.leaseRepo, s.tenantRepo, s.unitRepo, leaseChargeRepo, auditSvc)

	userHandler := NewUserHandler(userSvc, "", false)
	s.tenantHandler = NewTenantHandler(tenantSvc)
	s.leaseHandler = NewLeaseHandler(leaseSvc)

	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.setupRoutes(userHandler, auditSvc)
	s.seedFixtures()
}

func (s *TenantIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		_ = s.db.Close()
	}
}

func (s *TenantIntegrationTestSuite) SetupTest() {
	// Remove tenants and leases created during the previous test
	_, _ = s.db.Exec("DELETE FROM leases WHERE organization_id = $1", s.testOrg.ID)
	_, _ = s.db.Exec("DELETE FROM tenants WHERE organization_id = $1", s.testOrg.ID)
}

func (s *TenantIntegrationTestSuite) setupRoutes(userHandler *UserHandler, auditSvc *database.AuditService) {
	v1 := s.router.Group("/api/v1")
	v1.POST("/auth/login", userHandler.Login)

	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired(s.cfg.JWTSecret, auditSvc))
	protected.Use(middleware.RequireOrgContext())

	tenants := protected.Group("/tenants")
	{
		tenants.GET("", s.tenantHandler.GetAllTenants)
		tenants.POST("", s.tenantHandler.CreateTenant)
		tenants.GET("/:id", s.tenantHandler.GetTenantByID)
		tenants.PUT("/:id", s.tenantHandler.UpdateTenant)
		tenants.DELETE("/:id", s.tenantHandler.DeleteTenant)
	}

	leases := protected.Group("/leases")
	{
		leases.POST("", s.leaseHandler.CreateLease)
	}
}

func (s *TenantIntegrationTestSuite) seedFixtures() {
	s.testOrg = &models.Organization{
		Name:             "Tenant Test Org",
		Slug:             "tenant-test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         10,
		Active:           true,
	}
	require.NoError(s.T(), s.orgRepo.Create(s.testOrg))

	hash, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	require.NoError(s.T(), err)

	s.testUser = &models.User{
		Username:       "tenanttest",
		Email:          "tenanttest@example.com",
		PasswordHash:   string(hash),
		Role:           "Admin",
		Active:         true,
		Status:         "active",
		OrganizationID: &s.testOrg.ID,
	}
	require.NoError(s.T(), s.userRepo.Create(s.testUser))

	orgRoles := repositories.NewUserOrganizationRoleRepository(s.db)
	require.NoError(s.T(), orgRoles.Upsert(s.testUser.ID, s.testOrg.ID, "Admin"))

	s.authToken = s.login()

	s.testProperty, err = s.propertyRepo.Create(&models.CreatePropertyRequest{
		PropertyName:   "Tenant Test Property",
		PropertyCode:   "TNTP001",
		PropertyType:   models.PropertyTypeResidential,
		Address:        "1 Tenant St",
		City:           "Dhaka",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	s.testBuilding = &models.Building{
		PropertyID:     s.testProperty.ID,
		OrganizationID: s.testOrg.ID,
		BuildingName:   "Tenant Test Building",
		BuildingCode:   "TTB001",
		BuildingType:   models.BuildingTypeResidential,
		TotalFloors:    4,
		ActiveStatus:   true,
	}
	require.NoError(s.T(), s.buildingRepo.Create(s.testBuilding))

	s.testUnit, err = s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: "T101",
		UnitType:   models.UnitTypeApartment,
		Floor:      1,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)
}

func (s *TenantIntegrationTestSuite) login() string {
	body, _ := json.Marshal(map[string]string{"username": s.testUser.Username, "password": testUserPassword})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	require.Equal(s.T(), http.StatusOK, w.Code, "login failed: %s", w.Body.String())
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	token, ok := resp["token"].(string)
	require.True(s.T(), ok, "no token in login response: %s", w.Body.String())
	return "Bearer " + token
}

func (s *TenantIntegrationTestSuite) req(method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	r := httptest.NewRequest(method, path, buf)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.authToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	return w
}

// ── Create Tenant ─────────────────────────────────────────────────────────────

func (s *TenantIntegrationTestSuite) TestCreateTenant_Individual_Success() {
	payload := map[string]interface{}{
		"name":         "Alice Rahman",
		"tenant_type":  "Individual",
		"phone_number": "01711000001",
		"email":        "alice@example.com",
		"nid_number":   "19901234560001",
		"address":      "5 Mirpur Road, Dhaka",
	}

	w := s.req(http.MethodPost, "/api/v1/tenants", payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var resp models.TenantResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Greater(s.T(), resp.ID, 0)
	assert.Equal(s.T(), "Alice Rahman", resp.Name)
	assert.Equal(s.T(), models.TenantTypeIndividual, resp.TenantType)
	assert.Equal(s.T(), "01711000001", resp.PhoneNumber)
	assert.Equal(s.T(), "alice@example.com", resp.Email)
	assert.Equal(s.T(), "5 Mirpur Road, Dhaka", resp.Address)
	assert.True(s.T(), resp.Active)
	// NIDLastFour contains the last 4 characters of the original NID — full NID is never returned
	assert.Equal(s.T(), "0001", resp.NIDLastFour)
}

func (s *TenantIntegrationTestSuite) TestCreateTenant_Business_Success() {
	payload := map[string]interface{}{
		"name":         "Acme Corp",
		"tenant_type":  "Business",
		"phone_number": "01711000010",
		"nid_number":   "BIZ-REG-0010",
	}

	w := s.req(http.MethodPost, "/api/v1/tenants", payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var resp models.TenantResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), models.TenantTypeBusiness, resp.TenantType)
	assert.Equal(s.T(), "Acme Corp", resp.Name)
}

func (s *TenantIntegrationTestSuite) TestCreateTenant_MissingRequiredFields() {
	// name, phone_number, and nid_number are all required
	payload := map[string]interface{}{
		"tenant_type": "Individual",
		"email":       "nophone@example.com",
	}

	w := s.req(http.MethodPost, "/api/v1/tenants", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_TENANT_INVALID_BODY", resp["code"])
}

func (s *TenantIntegrationTestSuite) TestCreateTenant_InvalidTenantType() {
	payload := map[string]interface{}{
		"name":         "Bob",
		"tenant_type":  "Partnership", // not a valid enum value
		"phone_number": "01711000020",
		"nid_number":   "19900000020",
	}

	w := s.req(http.MethodPost, "/api/v1/tenants", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_TENANT_INVALID_BODY", resp["code"])
}

func (s *TenantIntegrationTestSuite) TestCreateTenant_DuplicateNID() {
	first := map[string]interface{}{
		"name": "Bob", "tenant_type": "Individual",
		"phone_number": "01711000030", "nid_number": "NID-DUP-0030",
	}
	w := s.req(http.MethodPost, "/api/v1/tenants", first)
	require.Equal(s.T(), http.StatusCreated, w.Code)

	// Same NID, different phone — must conflict
	second := map[string]interface{}{
		"name": "Bob Clone", "tenant_type": "Individual",
		"phone_number": "01711000031", "nid_number": "NID-DUP-0030",
	}
	w = s.req(http.MethodPost, "/api/v1/tenants", second)

	assert.Equal(s.T(), http.StatusConflict, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_TENANT_CONFLICT", resp["code"])
	assert.Equal(s.T(), "NID number already exists", resp["error"])
}

// ── List Tenants ──────────────────────────────────────────────────────────────

func (s *TenantIntegrationTestSuite) TestGetAllTenants_ReturnsSeededTenants() {
	for i := 0; i < 3; i++ {
		_, err := s.tenantRepo.Create(&models.CreateTenantRequest{
			Name: fmt.Sprintf("List Tenant %d", i), TenantType: models.TenantTypeIndividual,
			PhoneNumber:    fmt.Sprintf("0171100%04d", i+100),
			NIDNumber:      fmt.Sprintf("NID-LIST-%04d", i),
			OrganizationID: s.testOrg.ID,
		})
		require.NoError(s.T(), err)
	}

	w := s.req(http.MethodGet, "/api/v1/tenants?page=1&page_size=20", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp models.TenantListResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.GreaterOrEqual(s.T(), len(resp.Tenants), 3)
	require.NotNil(s.T(), resp.Pagination)
	assert.GreaterOrEqual(s.T(), resp.Pagination.TotalItems, 3)
}

func (s *TenantIntegrationTestSuite) TestGetAllTenants_DefaultPagination() {
	w := s.req(http.MethodGet, "/api/v1/tenants", nil)

	// Endpoint must respond even with no query params
	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp models.TenantListResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	require.NotNil(s.T(), resp.Pagination)
}

// ── Get Tenant ────────────────────────────────────────────────────────────────

func (s *TenantIntegrationTestSuite) TestGetTenantByID_ReturnsCorrectTenant() {
	created, err := s.tenantRepo.Create(&models.CreateTenantRequest{
		Name: "Charlie", TenantType: models.TenantTypeIndividual,
		PhoneNumber: "01711000200", NIDNumber: "NID-CHARLIE-0200",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/tenants/%d", created.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), float64(created.ID), resp["id"])
	assert.Equal(s.T(), "Charlie", resp["name"])
	assert.Equal(s.T(), "01711000200", resp["phone_number"])
	// Full NID must not appear anywhere in the response
	assert.Nil(s.T(), resp["nid_number"])
	assert.NotEmpty(s.T(), resp["nid_last_four"])
}

func (s *TenantIntegrationTestSuite) TestGetTenantByID_NotFound() {
	w := s.req(http.MethodGet, "/api/v1/tenants/999999", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "Tenant not found", resp["error"])
}

// ── Update Tenant ─────────────────────────────────────────────────────────────

func (s *TenantIntegrationTestSuite) TestUpdateTenant_NameAndPhone() {
	created, err := s.tenantRepo.Create(&models.CreateTenantRequest{
		Name: "Dave", TenantType: models.TenantTypeIndividual,
		PhoneNumber: "01711000300", NIDNumber: "NID-DAVE-0300",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	w := s.req(http.MethodPut, fmt.Sprintf("/api/v1/tenants/%d", created.ID), map[string]interface{}{
		"name":         "David",
		"phone_number": "01711000301",
	})

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "David", resp["name"])
	assert.Equal(s.T(), "01711000301", resp["phone_number"])
	// tenant_type must be preserved when not included in the update
	assert.Equal(s.T(), string(models.TenantTypeIndividual), resp["tenant_type"])
}

func (s *TenantIntegrationTestSuite) TestUpdateTenant_DeactivateTenant() {
	created, err := s.tenantRepo.Create(&models.CreateTenantRequest{
		Name: "Eve", TenantType: models.TenantTypeIndividual,
		PhoneNumber: "01711000400", NIDNumber: "NID-EVE-0400",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)
	assert.True(s.T(), created.Active)

	active := false
	w := s.req(http.MethodPut, fmt.Sprintf("/api/v1/tenants/%d", created.ID), map[string]interface{}{
		"active": active,
	})

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), false, resp["active"])
}

// ── Delete Tenant ─────────────────────────────────────────────────────────────

func (s *TenantIntegrationTestSuite) TestDeleteTenant_Success() {
	created, err := s.tenantRepo.Create(&models.CreateTenantRequest{
		Name: "Frank", TenantType: models.TenantTypeIndividual,
		PhoneNumber: "01711000500", NIDNumber: "NID-FRANK-0500",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	w := s.req(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%d", created.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "Tenant deleted successfully", resp["message"])
}

func (s *TenantIntegrationTestSuite) TestDeleteTenant_WithActiveLease_Conflicts() {
	tenant, err := s.tenantRepo.Create(&models.CreateTenantRequest{
		Name: "Grace", TenantType: models.TenantTypeIndividual,
		PhoneNumber: "01711000600", NIDNumber: "NID-GRACE-0600",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	// Create an active lease for this tenant via the API
	endDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	w := s.req(http.MethodPost, "/api/v1/leases", map[string]interface{}{
		"unit_id":          s.testUnit.ID,
		"tenant_id":        tenant.ID,
		"lease_type":       "Residential",
		"start_date":       time.Now().Format("2006-01-02"),
		"end_date":         endDate,
		"duration_months":  12,
		"monthly_rent":     12000,
		"security_deposit": 24000,
		"organization_id":  s.testOrg.ID,
	})
	require.Equal(s.T(), http.StatusCreated, w.Code)

	// Delete must fail with a specific, actionable error code
	w = s.req(http.MethodDelete, fmt.Sprintf("/api/v1/tenants/%d", tenant.ID), nil)

	assert.Equal(s.T(), http.StatusConflict, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "DELETE_TENANT_CONFLICT", resp["code"])
	assert.Equal(s.T(), "cannot delete tenant with active leases", resp["error"])
}

func TestTenantIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TenantIntegrationTestSuite))
}
