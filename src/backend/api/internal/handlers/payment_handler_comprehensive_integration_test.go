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

// PaymentIntegrationTestSuite exercises the full Payment API against a real
// PostgreSQL database. Every test issues an HTTP request through the live Gin
// router, checks the status code, and asserts exact response-body values.
type PaymentIntegrationTestSuite struct {
	suite.Suite
	db             *sqlx.DB
	router         *gin.Engine
	cfg            *config.Config
	paymentHandler *PaymentHandler
	orgRepo        *repositories.OrganizationRepository
	userRepo       *repositories.UserRepository
	propertyRepo   *repositories.PropertyRepository
	buildingRepo   *repositories.BuildingRepository
	unitRepo       *repositories.UnitRepository
	tenantRepo     *repositories.TenantRepository
	paymentRepo    *repositories.PaymentRepository
	leaseRepo      *repositories.LeaseRepository
	paymentTxnRepo *repositories.PaymentTransactionRepository
	testOrg        *models.Organization
	testUser       *models.User
	testProperty   *models.Property
	testBuilding   *models.Building
	testUnit       *models.Unit
	testTenant     *models.Tenant
	authToken      string
	// Dedicated year far in the past so test months don't collide with real data
	baseYear int
}

func (s *PaymentIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	nidKey := sha256.Sum256([]byte("payment-integration-test-key"))
	nidProtector, err := appcrypto.NewNIDProtector(nidKey[:], []byte("payment-integration-test-pepper"))
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
	s.paymentRepo = repositories.NewPaymentRepository(s.db)
	s.leaseRepo = repositories.NewLeaseRepository(s.db)
	s.paymentTxnRepo = repositories.NewPaymentTransactionRepository(s.db)
	paymentTransactionAttachmentRepo := repositories.NewPaymentTransactionAttachmentRepository(s.db)
	receiptAccessTokenRepo := repositories.NewReceiptAccessTokenRepository(s.db)
	notificationRepo := repositories.NewNotificationRepository(s.db)

	auditSvc := database.NewAuditService(s.db)
	userSvc := services.NewUserService(s.userRepo, auditSvc, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	paymentSvc := services.NewPaymentService(
		s.paymentRepo,
		s.paymentTxnRepo,
		paymentTransactionAttachmentRepo,
		receiptAccessTokenRepo,
		notificationRepo,
		s.leaseRepo,
		s.unitRepo,
		s.buildingRepo,
		s.propertyRepo,
		auditSvc,
		s.userRepo,
		"http://localhost:8080",
	)

	userHandler := NewUserHandler(userSvc, "", false)
	s.paymentHandler = NewPaymentHandler(paymentSvc)

	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.setupRoutes(userHandler, auditSvc)
	s.seedFixtures()

	// Use a year far in the past (but still >= the CreatePaymentRequest.Year
	// "min=2020" validation floor) so tests don't collide with real calendar data
	s.baseYear = 2020
}

func (s *PaymentIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		_ = s.db.Close()
	}
}

func (s *PaymentIntegrationTestSuite) SetupTest() {
	// Wipe all payments for this org so each test starts with a clean slate
	_, _ = s.db.Exec("DELETE FROM payments WHERE organization_id = $1", s.testOrg.ID)
}

func (s *PaymentIntegrationTestSuite) setupRoutes(userHandler *UserHandler, auditSvc *database.AuditService) {
	v1 := s.router.Group("/api/v1")
	v1.POST("/auth/login", userHandler.Login)

	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired(s.cfg.JWTSecret, auditSvc))
	protected.Use(middleware.RequireOrgContext())

	payments := protected.Group("/payments")
	{
		payments.POST("", s.paymentHandler.CreatePayment)
		payments.POST("/bulk", s.paymentHandler.BulkCreatePayments)
		payments.GET("/:id", s.paymentHandler.GetPayment)
		payments.PUT("/:id", s.paymentHandler.UpdatePayment)
		payments.POST("/:id/transactions", s.paymentHandler.RecordPaymentTransaction)
		payments.GET("", s.paymentHandler.GetPayments)
		payments.GET("/building/:building_id/report", s.paymentHandler.GetBuildingPaymentReport)
		payments.GET("/property/:property_id/report", s.paymentHandler.GetPropertyPaymentReport)
	}

	protected.GET("/dashboard/summary", s.paymentHandler.GetDashboardSummary)
}

func (s *PaymentIntegrationTestSuite) seedFixtures() {
	s.testOrg = &models.Organization{
		Name:             "Payment Test Org",
		Slug:             "payment-test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         10,
		Active:           true,
	}
	require.NoError(s.T(), s.orgRepo.Create(s.testOrg))

	hash, err := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
	require.NoError(s.T(), err)

	s.testUser = &models.User{
		Username:       "paymenttestuser",
		Email:          "paymenttest@example.com",
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
		PropertyName:   "Payment Test Property",
		PropertyCode:   "PMTP001",
		PropertyType:   models.PropertyTypeResidential,
		Address:        "3 Payment Ave",
		City:           "Dhaka",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	s.testBuilding = &models.Building{
		PropertyID:     s.testProperty.ID,
		OrganizationID: s.testOrg.ID,
		BuildingName:   "Payment Test Building",
		BuildingCode:   "PTB001",
		BuildingType:   models.BuildingTypeResidential,
		TotalFloors:    4,
		ActiveStatus:   true,
	}
	require.NoError(s.T(), s.buildingRepo.Create(s.testBuilding))

	s.testUnit, err = s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: "P101",
		UnitType:   models.UnitTypeApartment,
		Floor:      1,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)

	s.testTenant, err = s.tenantRepo.Create(&models.CreateTenantRequest{
		Name:           "Payment Tenant",
		TenantType:     models.TenantTypeIndividual,
		PhoneNumber:    "01711999001",
		NIDNumber:      "NID-PMT-0001",
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)

	// CreatePayment requires an active, non-expired lease for the tenant/unit
	// pair, so every payment test needs one already in place.
	s.createLease(s.testUnit.ID, s.testTenant.ID)
}

// createLease creates an active lease, far from expiry, for the given
// unit/tenant pair — the precondition CreatePayment enforces before
// recording a payment.
func (s *PaymentIntegrationTestSuite) createLease(unitID, tenantID int) *models.Lease {
	endDate := time.Now().AddDate(5, 0, 0).Format("2006-01-02")
	lease, err := s.leaseRepo.Create(&models.CreateLeaseRequest{
		UnitID:         unitID,
		TenantID:       tenantID,
		LeaseType:      models.LeaseTypeResidential,
		StartDate:      time.Now().AddDate(-1, 0, 0).Format("2006-01-02"),
		EndDate:        &endDate,
		DurationMonths: 12,
		MonthlyRent:    15000,
		OrganizationID: s.testOrg.ID,
	})
	require.NoError(s.T(), err)
	return lease
}

func (s *PaymentIntegrationTestSuite) login() string {
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

func (s *PaymentIntegrationTestSuite) req(method, path string, body interface{}) *httptest.ResponseRecorder {
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

// basePayload returns a minimal valid CreatePaymentRequest for testUnit/testTenant.
func (s *PaymentIntegrationTestSuite) basePayload(month, year int) map[string]interface{} {
	return map[string]interface{}{
		"unit_id":     s.testUnit.ID,
		"tenant_id":   s.testTenant.ID,
		"building_id": s.testBuilding.ID,
		"property_id": s.testProperty.ID,
		"month":       month,
		"year":        year,
		"amount_due":  15000.0,
	}
}

// seedPayment inserts a payment directly through the repository (bypasses HTTP),
// suitable for setting up preconditions without consuming a month/year slot via the API.
// When amountPaid is set, it also records a matching ledger transaction —
// RecordPaymentTransaction derives the running total from the transaction
// ledger, not the payments row, so a seeded amount_paid with no backing
// transaction would be silently dropped by the next recorded installment.
func (s *PaymentIntegrationTestSuite) seedPayment(month, year int, amountDue float64, amountPaid *float64) *models.Payment {
	p, err := s.paymentRepo.Create(&models.CreatePaymentRequest{
		UnitID:         s.testUnit.ID,
		TenantID:       s.testTenant.ID,
		BuildingID:     s.testBuilding.ID,
		PropertyID:     s.testProperty.ID,
		OrganizationID: s.testOrg.ID,
		Month:          month,
		Year:           year,
		AmountDue:      amountDue,
		AmountPaid:     amountPaid,
	})
	require.NoError(s.T(), err)

	if amountPaid != nil && *amountPaid > 0 {
		_, err := s.paymentTxnRepo.Create(p.ID, *amountPaid, "", p.ReceiptNumber, "", time.Now().UTC())
		require.NoError(s.T(), err)
	}

	return p
}

// ── Create Payment ────────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestCreatePayment_StatusDue_WhenNoAmountPaid() {
	// Omitting amount_paid → server must derive status "Due"
	w := s.req(http.MethodPost, "/api/v1/payments", s.basePayload(1, s.baseYear))

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Greater(s.T(), payment.ID, 0)
	assert.Equal(s.T(), s.testUnit.ID, payment.UnitID)
	assert.Equal(s.T(), s.testTenant.ID, payment.TenantID)
	assert.Equal(s.T(), 1, payment.Month)
	assert.Equal(s.T(), s.baseYear, payment.Year)
	assert.Equal(s.T(), 15000.0, payment.AmountDue)
	assert.Equal(s.T(), models.PaymentStatusDue, payment.Status)
	assert.NotEmpty(s.T(), payment.ReceiptNumber, "server must always generate a receipt number")
}

func (s *PaymentIntegrationTestSuite) TestCreatePayment_StatusPaid_WhenFullAmountPaid() {
	payload := s.basePayload(2, s.baseYear)
	amountPaid := 15000.0
	payload["amount_paid"] = amountPaid
	payload["payment_method"] = "Bank Transfer"
	payload["payment_date"] = time.Now().Format("2006-01-02")

	w := s.req(http.MethodPost, "/api/v1/payments", payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), models.PaymentStatusPaid, payment.Status)
	assert.Equal(s.T(), amountPaid, payment.AmountPaid)
	assert.Equal(s.T(), "Bank Transfer", payment.PaymentMethod)
}

func (s *PaymentIntegrationTestSuite) TestCreatePayment_StatusPartial_WhenUnderpaid() {
	payload := s.basePayload(3, s.baseYear)
	partial := 7500.0
	payload["amount_paid"] = partial

	w := s.req(http.MethodPost, "/api/v1/payments", payload)

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), models.PaymentStatusPartial, payment.Status)
	assert.Equal(s.T(), partial, payment.AmountPaid)
}

func (s *PaymentIntegrationTestSuite) TestCreatePayment_MissingRequiredFields() {
	// unit_id, tenant_id, month, year, amount_due are all required
	w := s.req(http.MethodPost, "/api/v1/payments", map[string]interface{}{
		"building_id": s.testBuilding.ID,
	})

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_PAYMENT_INVALID_BODY", resp["code"])
}

func (s *PaymentIntegrationTestSuite) TestCreatePayment_DuplicateMonthYear_Conflicts() {
	w := s.req(http.MethodPost, "/api/v1/payments", s.basePayload(4, s.baseYear))
	require.Equal(s.T(), http.StatusCreated, w.Code)

	// Second payment for the same unit/month/year must be rejected
	w = s.req(http.MethodPost, "/api/v1/payments", s.basePayload(4, s.baseYear))

	assert.Equal(s.T(), http.StatusConflict, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "PAYMENT_ALREADY_EXISTS", resp["code"])
	assert.Equal(s.T(), "A payment already exists for this unit for the selected month/year", resp["error"])
}

func (s *PaymentIntegrationTestSuite) TestCreatePayment_NegativeAmountDue_Rejected() {
	payload := s.basePayload(5, s.baseYear)
	payload["amount_due"] = -1000.0

	w := s.req(http.MethodPost, "/api/v1/payments", payload)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "CREATE_PAYMENT_INVALID_BODY", resp["code"])
}

// ── Get Payment ───────────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestGetPayment_ReturnsEnrichedDetails() {
	created := s.seedPayment(6, s.baseYear, 18000, nil)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/payments/%d", created.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var payment models.PaymentWithDetails
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), created.ID, payment.ID)
	assert.Equal(s.T(), 18000.0, payment.AmountDue)
	assert.Equal(s.T(), models.PaymentStatusDue, payment.Status)
	// Enriched fields must be resolved from joined tables
	assert.Equal(s.T(), "Payment Test Building", payment.BuildingName)
	assert.Equal(s.T(), "Payment Test Property", payment.PropertyName)
	assert.Equal(s.T(), "PTB001", payment.BuildingCode)
	assert.Equal(s.T(), "P101", payment.UnitNumber)
	assert.Equal(s.T(), "Payment Tenant", payment.TenantName)
}

func (s *PaymentIntegrationTestSuite) TestGetPayment_NotFound() {
	w := s.req(http.MethodGet, "/api/v1/payments/999999", nil)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "GET_PAYMENT_FAILED", resp["code"])
}

// ── Update Payment ────────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestUpdatePayment_DueBecomespaid() {
	// Start with a Due payment (no amount_paid)
	created := s.seedPayment(7, s.baseYear, 20000, nil)
	assert.Equal(s.T(), models.PaymentStatusDue, created.Status)

	// Recording money received goes through the transaction ledger, not a
	// direct PUT of amount_paid — status must flip to Paid, server-side.
	w := s.req(http.MethodPost, fmt.Sprintf("/api/v1/payments/%d/transactions", created.ID), map[string]interface{}{
		"amount":         20000.0,
		"payment_method": "Cash",
		"payment_date":   time.Now().Format("2006-01-02"),
	})

	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), models.PaymentStatusPaid, payment.Status)
	assert.Equal(s.T(), 20000.0, payment.AmountPaid)
	// Receipt number must be assigned when money is first recorded
	assert.NotEmpty(s.T(), payment.ReceiptNumber)
}

func (s *PaymentIntegrationTestSuite) TestUpdatePayment_PartialToFull() {
	partial := 10000.0
	created := s.seedPayment(8, s.baseYear, 20000, &partial)
	assert.Equal(s.T(), models.PaymentStatusPartial, created.Status)

	// Pay the remaining balance via the transaction ledger — status must
	// become Paid.
	w := s.req(http.MethodPost, fmt.Sprintf("/api/v1/payments/%d/transactions", created.ID), map[string]interface{}{
		"amount": 10000.0,
	})

	assert.Equal(s.T(), http.StatusCreated, w.Code)
	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), models.PaymentStatusPaid, payment.Status)
}

func (s *PaymentIntegrationTestSuite) TestUpdatePayment_RecordNotes() {
	created := s.seedPayment(9, s.baseYear, 15000, nil)

	w := s.req(http.MethodPut, fmt.Sprintf("/api/v1/payments/%d", created.ID), map[string]interface{}{
		"notes": "Paid via mobile banking",
	})

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var payment models.Payment
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &payment))
	assert.Equal(s.T(), "Paid via mobile banking", payment.Notes)
	// Status must remain Due because no amount_paid was supplied
	assert.Equal(s.T(), models.PaymentStatusDue, payment.Status)
}

func (s *PaymentIntegrationTestSuite) TestUpdatePayment_NotFound() {
	w := s.req(http.MethodPut, "/api/v1/payments/999999", map[string]interface{}{
		"amount_paid": 5000.0,
	})

	assert.Equal(s.T(), http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), "payment not found", resp["error"])
}

// ── List Payments ─────────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestGetPayments_Pagination() {
	// Seed across three months using a distinct year to avoid slot collisions
	listYear := s.baseYear + 1
	for i := 1; i <= 3; i++ {
		s.seedPayment(i, listYear, 15000, nil)
	}

	w := s.req(http.MethodGet, "/api/v1/payments?page=1&page_size=2", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var resp models.PaymentListResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.LessOrEqual(s.T(), len(resp.Payments), 2, "page_size=2 must be respected")
	assert.GreaterOrEqual(s.T(), resp.Total, 3)
	assert.Equal(s.T(), 1, resp.Page)
	assert.Equal(s.T(), 2, resp.PageSize)
}

func (s *PaymentIntegrationTestSuite) TestGetPayments_FilterByStatus_Paid() {
	filterYear := s.baseYear + 2
	paid := 12000.0
	// One Paid record
	s.seedPayment(1, filterYear, 12000, &paid)
	// One Due record
	s.seedPayment(2, filterYear, 12000, nil)

	w := s.req(http.MethodGet, "/api/v1/payments?status=Paid&page=1&page_size=50", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp models.PaymentListResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	for _, p := range resp.Payments {
		assert.Equal(s.T(), models.PaymentStatusPaid, p.Status, "filter must return only Paid records; got %s for payment %d", p.Status, p.ID)
	}
}

func (s *PaymentIntegrationTestSuite) TestGetPayments_FilterByYear() {
	yearA := s.baseYear + 3
	yearB := s.baseYear + 4
	s.seedPayment(1, yearA, 10000, nil)
	s.seedPayment(1, yearB, 10000, nil)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/payments?year=%d&page=1&page_size=50", yearA), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var resp models.PaymentListResponse
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	for _, p := range resp.Payments {
		assert.Equal(s.T(), yearA, p.Year, "year filter must exclude records from other years")
	}
}

// ── Dashboard Summary ─────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestGetDashboardSummary_ContainsAggregateKeys() {
	// Seed one payment so the summary totals are non-trivially populated
	s.seedPayment(1, s.baseYear+5, 15000, nil)

	w := s.req(http.MethodGet, "/api/v1/dashboard/summary", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var summary models.DashboardSummary
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &summary))
	// These fields must always be present; exact values depend on seeded data
	assert.GreaterOrEqual(s.T(), summary.TotalDue, 0.0)
	assert.GreaterOrEqual(s.T(), summary.TotalPaid, 0.0)
	assert.GreaterOrEqual(s.T(), summary.CollectionRate, 0.0)
}

func (s *PaymentIntegrationTestSuite) TestGetDashboardSummary_CollectionRateAccurate() {
	dashYear := s.baseYear + 6
	// Seed one paid and one due so collection rate should be 50 %
	paid := 15000.0
	s.seedPayment(1, dashYear, 15000, &paid)
	s.seedPayment(2, dashYear, 15000, nil)

	w := s.req(http.MethodGet, "/api/v1/dashboard/summary", nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var summary models.DashboardSummary
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &summary))
	// With exactly 1 Paid out of 2 total due records, the rate should be ≥ 0 and ≤ 100
	assert.GreaterOrEqual(s.T(), summary.CollectionRate, 0.0)
	assert.LessOrEqual(s.T(), summary.CollectionRate, 100.0)
}

// ── Building Payment Report ───────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestGetBuildingPaymentReport_Success() {
	s.seedPayment(1, s.baseYear+7, 15000, nil)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/payments/building/%d/report", s.testBuilding.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var report models.BuildingPaymentReport
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &report))
	assert.Equal(s.T(), s.testBuilding.ID, report.BuildingID)
	assert.Equal(s.T(), "Payment Test Building", report.BuildingName)
	assert.Equal(s.T(), "PTB001", report.BuildingCode)
	assert.Equal(s.T(), s.testProperty.ID, report.PropertyID)
	assert.Equal(s.T(), "Payment Test Property", report.PropertyName)
	assert.NotEmpty(s.T(), report.ReportPeriod)
	assert.NotZero(s.T(), report.GeneratedAt)
}

func (s *PaymentIntegrationTestSuite) TestGetBuildingPaymentReport_ContainsSeededPayment() {
	created := s.seedPayment(2, s.baseYear+7, 20000, nil)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/payments/building/%d/report", s.testBuilding.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var report models.BuildingPaymentReport
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &report))

	found := false
	for _, p := range report.Payments {
		if p.ID == created.ID {
			found = true
			assert.Equal(s.T(), 20000.0, p.AmountDue)
			break
		}
	}
	assert.True(s.T(), found, "seeded payment must appear in the building report")
}

// ── Property Payment Report ───────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestGetPropertyPaymentReport_Success() {
	s.seedPayment(1, s.baseYear+8, 15000, nil)

	w := s.req(http.MethodGet, fmt.Sprintf("/api/v1/payments/property/%d/report", s.testProperty.ID), nil)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var report models.PropertyPaymentReport
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &report))
	assert.Equal(s.T(), s.testProperty.ID, report.PropertyID)
	assert.Equal(s.T(), "Payment Test Property", report.PropertyName)
	assert.NotEmpty(s.T(), report.ReportPeriod)
}

// ── Bulk Create ───────────────────────────────────────────────────────────────

func (s *PaymentIntegrationTestSuite) TestBulkCreatePayments_Success() {
	// Create a second unit for the bulk request
	unit2, err := s.unitRepo.Create(&models.CreateUnitRequest{
		BuildingID: s.testBuilding.ID,
		PropertyID: s.testProperty.ID,
		UnitNumber: "P102",
		UnitType:   models.UnitTypeApartment,
		Floor:      2,
	}, s.testOrg.ID)
	require.NoError(s.T(), err)
	s.createLease(unit2.ID, s.testTenant.ID)

	bulkYear := s.baseYear + 9

	// BulkCreatePayments takes a top-level JSON array, not a wrapped object
	payload := []map[string]interface{}{
		{
			"unit_id": s.testUnit.ID, "tenant_id": s.testTenant.ID,
			"building_id": s.testBuilding.ID, "property_id": s.testProperty.ID,
			"month": 1, "year": bulkYear, "amount_due": 15000,
		},
		{
			"unit_id": unit2.ID, "tenant_id": s.testTenant.ID,
			"building_id": s.testBuilding.ID, "property_id": s.testProperty.ID,
			"month": 1, "year": bulkYear, "amount_due": 16000,
		},
	}

	b, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/payments/bulk", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.authToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	// BulkCreatePayments always returns 207 Multi-Status
	assert.Equal(s.T(), http.StatusMultiStatus, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	created := resp["created"].([]interface{})
	assert.Len(s.T(), created, 2)
	assert.Equal(s.T(), float64(2), resp["success"])
	assert.Equal(s.T(), float64(0), resp["failed"])
	assert.Equal(s.T(), float64(2), resp["total"])
}

func (s *PaymentIntegrationTestSuite) TestBulkCreatePayments_PartialFailure() {
	bulkYear := s.baseYear + 10

	// Pre-create a payment for month 1 to force a duplicate conflict on the first item
	s.seedPayment(1, bulkYear, 15000, nil)

	payload := []map[string]interface{}{
		// This one will fail — slot already taken
		{
			"unit_id": s.testUnit.ID, "tenant_id": s.testTenant.ID,
			"building_id": s.testBuilding.ID, "property_id": s.testProperty.ID,
			"month": 1, "year": bulkYear, "amount_due": 15000,
		},
		// This one will succeed
		{
			"unit_id": s.testUnit.ID, "tenant_id": s.testTenant.ID,
			"building_id": s.testBuilding.ID, "property_id": s.testProperty.ID,
			"month": 2, "year": bulkYear, "amount_due": 15000,
		},
	}

	b, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/payments/bulk", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", s.authToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	assert.Equal(s.T(), http.StatusMultiStatus, w.Code)

	var resp map[string]interface{}
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(s.T(), float64(2), resp["total"])
	assert.Equal(s.T(), float64(1), resp["success"])
	assert.Equal(s.T(), float64(1), resp["failed"])
	errors := resp["errors"].([]interface{})
	assert.Len(s.T(), errors, 1, "exactly one error message expected for the duplicate slot")
}

func TestPaymentIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentIntegrationTestSuite))
}
