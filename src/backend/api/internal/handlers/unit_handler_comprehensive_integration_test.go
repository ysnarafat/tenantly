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

// UnitIntegrationTestSuite exercises the full Unit API against a real PostgreSQL
// database. Each test issues an HTTP request through the live Gin router, checks
// the status code, and asserts exact response-body values.
type UnitIntegrationTestSuite struct {
	suite.Suite
	db           *sqlx.DB
	router       *gin.Engine
	cfg          *config.Config
	unitHandler  *UnitHandler
	orgRepo      *repositories.OrganizationRepository
	userRepo     *repositories.UserRepository
	propertyRepo *repositories.PropertyRepository
	buildingRepo *repositories.BuildingRepository
	unitRepo     *repositories.UnitRepository
	testOrg      *models.Organization
	testUser     *models.User
	testProperty *models.Property
	testBuilding *models.Building
	authToken    string
	unitSeq      int // monotonically-increasing suffix for unique unit numbers
}

func (s *UnitIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	nidKey := sha256.Sum256([]byte("unit-integration-test-key"))
	nidProtector, err := appcrypto.NewNIDProtector(nidKey[:], []byte("unit-integration-test-pepper"))
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

	auditSvc := database.NewAuditService(s.db)
	userSvc := services.NewUserService(s.userRepo, auditSvc, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	unitSvc := services.NewUnitService(s.unitRepo, s.buildingRepo, s.propertyRepo, auditSvc)

	userHandler := NewUserHandler(userSvc, "", false)
	s.unitHandler = NewUnitHandler(unitSvc)

	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.setupRoutes(userHandler, auditSvc)
	s.seedFixtures()
}

func (s *UnitIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		_ = s.db.Close()
	}
}

func (s *UnitIntegrationTestSuite) setupRoutes(userHandler *UserHandler, auditSvc *database.AuditService) {
	v1 := s.router.Group("/api/v1")
	v1.POST("/auth/login", userHandler.Login)

	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired(s.cfg.JWTSecret, auditSvc))
	protected.Use(middleware.RequireOrgContext())

	units := protected.Group("/units")
	{
		units.POST("", s.unitHandler.CreateUnit)
		units.GET("/:id", s.unitHandler.GetUnit)
		units.PUT("/:id", s.unitHandler.UpdateUnit)
		units.DELETE("/:id", s.unitHandler.DeleteUnit)
		units.GET("/:id/hierarchy", s.unitHandler.GetUnitHierarchyContext)
	}

	buildings := protected.Group("/buildings")
	{
		buildings.GET("/:id/units/list", s.unitHandler.GetUnitsByBuilding)
		buildings.POST("/:id/units/bulk", s.unitHandler.BulkCreateUnits)
	}
}

func (s *UnitIntegrationTestSuite) seedFixtures() {
	s.testOrg = &models.Organization{
		Name:             "Unit Test Org",
		Slug:             "unit-test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         10,
		Active:           true,
	}
	require.NoError(s.T(), s.orgRepo.Create(s.testOrg))

	hash, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	require.NoError(s.T(), err)

	s.testUser = &models.User{
		Username:       "unittestuser",
		Email:          "unittestuser@example.com",
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
		PropertyName:   "Unit Test Property",
		PropertyCode:   "UNTP001",
		PropertyType:   models.PropertyTypeCommercial,
		Address:        "2 Unit St",
		City:           "Dhaka",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	s.testBuilding = &models.Building{
		PropertyID:     s.testProperty.ID,
		OrganizationID: s.testOrg.ID,
		BuildingName:   "Unit Test Building",
		BuildingCode:   "UTB001",
		// Mixed allows every unit type — this suite exercises all of them
		// (Shop, Office, Apartment, Parking, Storage, Other) against this
		// one shared building.
		BuildingType: models.BuildingTypeMixed,
		TotalFloors:  6,
		ActiveStatus: true,
	}
	require.NoError(s.T(), s.buildingRepo.Create(s.testBuilding))
}

func (s *UnitIntegrationTestSuite) login() string {
	body, _ := json.Marshal(map[string]string{"username": s.testUser.Username, "password": testUserPassword})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	require.Equal(s.T(), http.StatusOK, w.Code, "login failed: %s", w.Body.String())
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	token, ok := resp["token"].(string)
	require.True(s.T(), ok, "no token in login response")
	return "Bearer " + token
}

func (s *UnitIntegrationTestSuite) req(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// nextUnitNum returns a unique unit number per test run, avoiding DB collisions.
func (s *UnitIntegrationTestSuite) nextUnitNum() string {
	s.unitSeq++
	return fmt.Sprintf("U%04d", s.unitSeq)
}

// ── Create Unit ───────────────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestCreateUnit_Success() {
	payload := map[string]interface{}{
		"building_id": s.testBuilding.ID,
		"property_id": s.testProperty.ID,
		"unit_number": s.nextUnitNum(),
		"unit_type":   "Apartment",
		"floor":       2,
		"unit_name":   "Corner Suite",
	}

	w := s.req(http.MethodPost, "/api/v1/units", payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var unit models.Unit
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &unit))
	assert.Greater(s.T(), unit.ID, 0)
	assert.Equal(s.T(), s.testBuilding.ID, unit.BuildingID)
	assert.Equal(s.T(), models.UnitTypeApartment, unit.UnitType)
	assert.Equal(s.T(), 2, unit.Floor)
	assert.Equal(s.T(), "Corner Suite", unit.UnitName)
	assert.True(s.T(), unit.Active)
}

func (s *UnitIntegrationTestSuite) TestCreateUnit_AllSupportedTypes() {
	types := []models.UnitType{
		models.UnitTypeShop,
		models.UnitTypeOffice,
		models.UnitTypeParking,
		models.UnitTypeStorage,
		models.UnitTypeOther,
	}
	for _, ut := range types {
		payload := map[string]interface{}{
			"building_id": s.testBuilding.ID,
			"property_id": s.testProperty.ID,
			"unit_number": s.nextUnitNum(),
			"unit_type":   string(ut),
		}
		w := s.req(http.MethodPost, "/api/v1/units", payload)
		assert.Equal(s.T(), http.StatusCreated, w.Code, "expected 201 for unit type %s, got: %s", ut, w.Body.String())
	}
}

func (s *UnitIntegrationTestSuite) TestCreateUnit_MissingRequiredFields() {
	// unit_number and unit_type are required
	payload := map[string]interface{}{
		"building_id": s.testBuilding.ID,
		"floor":       1,
	}

	w := s.req(http.MethodPost, "/api/v1/units", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_UNIT_INVALID_BODY", resp["code"])
}

func (s *UnitIntegrationTestSuite) TestCreateUnit_InvalidUnitType() {
	payload := map[string]interface{}{
		"building_id": s.testBuilding.ID,
		"property_id": s.testProperty.ID,
		"unit_number": s.nextUnitNum(),
		"unit_type":   "Penthouse", // not a valid enum value
	}

	w := s.req(http.MethodPost, "/api/v1/units", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_UNIT_INVALID_BODY", resp["code"])
}

func (s *UnitIntegrationTestSuite) TestCreateUnit_DuplicateUnitNumber() {
	unitNum := s.nextUnitNum()

	payload := map[string]interface{}{
		"building_id": s.testBuilding.ID,
		"property_id": s.testProperty.ID,
		"unit_number": unitNum,
		"unit_type":   "Apartment",
	}
	w := s.req(http.MethodPost, "/api/v1/units", payload)
	require.Equal(s.T(), http.StatusCreated, w.Code)

	// Same unit number in the same building — must conflict
	w = s.req(http.MethodPost, "/api/v1/units", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_UNIT_FAILED", resp["code"])
}

// ── Get Unit ──────────────────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestGetUnit_ReturnsEnrichedDetails() {
	created, err := s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: s.nextUnitNum(),
		UnitType:   models.UnitTypeApartment,
		Floor:      3,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/units/%d", created.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), float64(created.ID), resp["id"])
	// UnitWithDetails must include enriched building/property names
	assert.Equal(s.T(), "Unit Test Building", resp["building_name"])
	assert.Equal(s.T(), "Unit Test Property", resp["property_name"])
	assert.Equal(s.T(), "UTB001", resp["building_code"])
	assert.Equal(s.T(), float64(3), resp["floor"])
}

func (s *UnitIntegrationTestSuite) TestGetUnit_NotFound() {
	w := s.req(http.MethodGet, "/api/v1/units/999999", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "GET_UNIT_FAILED", resp["code"])
}

// ── Update Unit ───────────────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestUpdateUnit_FloorAndName() {
	created, err := s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: s.nextUnitNum(),
		UnitType:   models.UnitTypeApartment,
		Floor:      1,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)

	w := s.req(http.MethodPut, fmt.Sprintf("/api/v1/units/%d", created.ID), map[string]interface{}{
		"floor":     5,
		"unit_name": "Penthouse A",
	})

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), float64(5), resp["floor"])
	assert.Equal(s.T(), "Penthouse A", resp["unit_name"])
	// unit_type must be preserved when not included in the update
	assert.Equal(s.T(), string(models.UnitTypeApartment), resp["unit_type"])
}

func (s *UnitIntegrationTestSuite) TestUpdateUnit_Deactivate() {
	created, err := s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: s.nextUnitNum(),
		UnitType:   models.UnitTypeOffice,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)
	assert.True(s.T(), created.Active)

	active := false
	w := s.req(http.MethodPut, fmt.Sprintf("/api/v1/units/%d", created.ID), map[string]interface{}{
		"active": active,
	})

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), false, resp["active"])
}

// ── Delete Unit ───────────────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestDeleteUnit_Success() {
	created, err := s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: s.nextUnitNum(),
		UnitType:   models.UnitTypeApartment,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)

	w := s.req(http.MethodDelete, fmt.Sprintf("/api/v1/units/%d", created.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "Unit deleted successfully", resp["message"])

	// Unit must no longer be fetchable after deletion
	w = s.req(http.MethodGet, fmt.Sprintf("/api/v1/units/%d", created.ID), nil)
	assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

// ── Bulk Create ───────────────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestBulkCreateUnits_Success() {
	payload := map[string]interface{}{
		"units": []map[string]interface{}{
			{"unit_number": s.nextUnitNum(), "unit_type": "Apartment", "floor": 1},
			{"unit_number": s.nextUnitNum(), "unit_type": "Apartment", "floor": 2},
			{"unit_number": s.nextUnitNum(), "unit_type": "Office", "floor": 3},
		},
	}

	w := s.req(http.MethodPost, fmt.Sprintf("/api/v1/buildings/%d/units/bulk", s.testBuilding.ID), payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "Units created successfully", resp["message"])
	assert.Equal(s.T(), float64(s.testBuilding.ID), resp["building_id"])

	units, ok := resp["units"].([]interface{})
	require.True(s.T(), ok)
	assert.Len(s.T(), units, 3)

	summary, ok := resp["summary"].(map[string]interface{})
	require.True(s.T(), ok)
	assert.Equal(s.T(), float64(3), summary["units_created"])
}

func (s *UnitIntegrationTestSuite) TestBulkCreateUnits_DuplicateWithinRequest() {
	dup := s.nextUnitNum()
	payload := map[string]interface{}{
		"units": []map[string]interface{}{
			{"unit_number": dup, "unit_type": "Apartment"},
			{"unit_number": dup, "unit_type": "Apartment"}, // duplicate
		},
	}

	w := s.req(http.MethodPost, fmt.Sprintf("/api/v1/buildings/%d/units/bulk", s.testBuilding.ID), payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "BULK_CREATE_UNITS_DUPLICATE_UNIT_NUMBER", resp["code"])
}

func (s *UnitIntegrationTestSuite) TestBulkCreateUnits_BuildingNotFound() {
	payload := map[string]interface{}{
		"units": []map[string]interface{}{
			{"unit_number": s.nextUnitNum(), "unit_type": "Apartment"},
		},
	}

	w := s.req(http.MethodPost, "/api/v1/buildings/999999/units/bulk", payload)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "BULK_CREATE_UNITS_BUILDING_NOT_FOUND", resp["code"])
}

func (s *UnitIntegrationTestSuite) TestBulkCreateUnits_EmptyPayload() {
	payload := map[string]interface{}{
		"units": []map[string]interface{}{},
	}

	w := s.req(http.MethodPost, fmt.Sprintf("/api/v1/buildings/%d/units/bulk", s.testBuilding.ID), payload)

	// An empty units array must be rejected (binding: min=1)
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

// ── List Units by Building ────────────────────────────────────────────────────

func (s *UnitIntegrationTestSuite) TestGetUnitsByBuilding_ReturnsPaginatedList() {
	// Seed two units
	for i := 0; i < 2; i++ {
		_, err := s.unitRepo.Create(&models.CreateUnitRequest{
			BuildingID: s.testBuilding.ID,
			PropertyID: s.testProperty.ID,
			UnitNumber: s.nextUnitNum(),
			UnitType:   models.UnitTypeApartment,
		}, s.testOrg.ID)
		require.NoError(s.T(), err)
	}

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/buildings/%d/units/list?page=1&page_size=20", s.testBuilding.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))

	data, ok := resp["data"].([]interface{})
	require.True(s.T(), ok, "expected 'data' array in response")
	assert.GreaterOrEqual(s.T(), len(data), 2)

	meta, ok := resp["meta"].(map[string]interface{})
	require.True(s.T(), ok, "expected 'meta' object in response")
	assert.NotNil(s.T(), meta["total"])
}

func TestUnitIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UnitIntegrationTestSuite))
}
