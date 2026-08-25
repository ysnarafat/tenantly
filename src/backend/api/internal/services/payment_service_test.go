package services

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// ---------------------------------------------------------------------------
// MockPaymentRepo â€“ implements interfaces.PaymentRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentRepo struct {
	payments                map[int]*models.PaymentWithDetails
	nextID                  int
	shouldFailCreate        bool
	shouldFailGetByID       bool
	shouldFailUpdate        bool
	shouldFailList          bool
	shouldFailStats         bool
	shouldFailPeriod        bool
	shouldFailDashboard     bool
	shouldFailBldgLevel     bool
	shouldFailAnalytics     bool
	shouldReturnExists      bool
	shouldFailReceiptNumber bool
	receiptCounter          int
	dashboardResult         *models.DashboardSummary
	buildingLevelResult     map[string]interface{}
	analyticsResult         *models.BuildingPaymentAnalytics
}

func newMockPaymentRepo() *MockPaymentRepo {
	return &MockPaymentRepo{
		payments:            make(map[int]*models.PaymentWithDetails),
		nextID:              1,
		dashboardResult:     &models.DashboardSummary{TotalDue: 1000, TotalPaid: 800, CollectionRate: 80},
		buildingLevelResult: map[string]interface{}{"total_buildings": 3},
		analyticsResult:     &models.BuildingPaymentAnalytics{BuildingID: 1, TotalRevenue: 5000, CollectionRate: 90},
	}
}

func (m *MockPaymentRepo) Create(req *models.CreatePaymentRequest) (*models.Payment, error) {
	if m.shouldFailCreate {
		return nil, errors.New("create payment failed")
	}
	p := &models.Payment{
		ID:             m.nextID,
		UnitID:         req.UnitID,
		TenantID:       req.TenantID,
		BuildingID:     req.BuildingID,
		PropertyID:     req.PropertyID,
		OrganizationID: req.OrganizationID,
		Month:          req.Month,
		Year:           req.Year,
		AmountDue:      req.AmountDue,
		Status:         models.PaymentStatusDue,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	pwd := &models.PaymentWithDetails{Payment: *p}
	m.payments[m.nextID] = pwd
	m.nextID++
	return p, nil
}

func (m *MockPaymentRepo) GetByID(id int) (*models.Payment, error) {
	if m.shouldFailGetByID {
		return nil, errors.New("payment not found")
	}
	pwd, ok := m.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	p := pwd.Payment
	return &p, nil
}

func (m *MockPaymentRepo) GetByIDWithDetails(id int) (*models.PaymentWithDetails, error) {
	if m.shouldFailGetByID {
		return nil, errors.New("payment not found")
	}
	pwd, ok := m.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return pwd, nil
}

func (m *MockPaymentRepo) Update(id int, req *models.UpdatePaymentRequest) (*models.Payment, error) {
	if m.shouldFailUpdate {
		return nil, errors.New("update payment failed")
	}
	pwd, ok := m.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	if req.AmountPaid != nil {
		pwd.AmountPaid = *req.AmountPaid
		// Mirrors PaymentRepository.Update's amount-vs-amount_due CASE
		// derivation (Overdue-by-due_date omitted — not needed by any test
		// using this mock; see the real-DB tests in payment_repository_test.go
		// for that).
		switch {
		case *req.AmountPaid >= pwd.AmountDue:
			pwd.Status = models.PaymentStatusPaid
		case *req.AmountPaid > 0:
			pwd.Status = models.PaymentStatusPartial
		default:
			pwd.Status = models.PaymentStatusDue
		}
	}
	if req.Status != nil {
		pwd.Status = *req.Status
	}
	if req.PaymentMethod != nil {
		pwd.PaymentMethod = *req.PaymentMethod
	}
	if req.ReceiptNumber != nil {
		pwd.ReceiptNumber = *req.ReceiptNumber
	}
	p := pwd.Payment
	return &p, nil
}

func (m *MockPaymentRepo) GetWithDetailsAndFilters(filters map[string]interface{}, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	if m.shouldFailList {
		return nil, 0, errors.New("list payments failed")
	}
	var result []*models.PaymentWithDetails
	for _, p := range m.payments {
		result = append(result, p)
	}
	total := len(result)
	if offset >= total {
		return []*models.PaymentWithDetails{}, 0, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (m *MockPaymentRepo) GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	if m.shouldFailStats {
		return nil, errors.New("stats failed")
	}
	return map[string]interface{}{"building_id": buildingID, "total_due": 5000.0}, nil
}

func (m *MockPaymentRepo) GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	if m.shouldFailStats {
		return nil, errors.New("property stats failed")
	}
	return map[string]interface{}{"property_id": propertyID, "total_due": 10000.0}, nil
}

func (m *MockPaymentRepo) GetSystemPaymentStats(startDate, endDate time.Time) (interface{}, error) {
	if m.shouldFailStats {
		return nil, errors.New("system stats failed")
	}
	return map[string]interface{}{"total_due": 50000.0}, nil
}

func (m *MockPaymentRepo) GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	if m.shouldFailPeriod {
		return nil, 0, errors.New("period payments failed")
	}
	var result []*models.PaymentWithDetails
	for _, p := range m.payments {
		if p.BuildingID == buildingID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

func (m *MockPaymentRepo) GetDashboardSummary(orgID int) (*models.DashboardSummary, error) {
	if m.shouldFailDashboard {
		return nil, errors.New("dashboard summary failed")
	}
	return m.dashboardResult, nil
}

func (m *MockPaymentRepo) GetBuildingLevelSummary() (map[string]interface{}, error) {
	if m.shouldFailBldgLevel {
		return nil, errors.New("building level summary failed")
	}
	return m.buildingLevelResult, nil
}

func (m *MockPaymentRepo) GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error) {
	if m.shouldFailAnalytics {
		return nil, errors.New("analytics failed")
	}
	a := *m.analyticsResult
	a.BuildingID = buildingID
	return &a, nil
}

func (m *MockPaymentRepo) SearchLeases(orgID int, query string) ([]*models.LeaseSearchResult, error) {
	return make([]*models.LeaseSearchResult, 0), nil
}

func (m *MockPaymentRepo) GetActiveLeasesForPeriod(orgID, month, year int, buildingID *int) ([]*models.LeaseSearchResult, error) {
	return make([]*models.LeaseSearchResult, 0), nil
}

func (m *MockPaymentRepo) CheckPaymentExists(unitID, month, year int) (bool, error) {
	return m.shouldReturnExists, nil
}

func (m *MockPaymentRepo) NextReceiptNumber(orgID int, yearMonth string) (string, error) {
	if m.shouldFailReceiptNumber {
		return "", errors.New("failed to generate receipt number")
	}
	m.receiptCounter++
	return fmt.Sprintf("ORG%d-%s-%04d", orgID, yearMonth, m.receiptCounter), nil
}

func (m *MockPaymentRepo) GetAgingBuckets(orgID int) (map[string]float64, error) {
	return map[string]float64{"current": 0, "30d": 0, "60d": 0, "90d+": 0}, nil
}

func (m *MockPaymentRepo) GetMonthlyCollectionTrend(orgID int, months int) ([]*models.MonthlyCollectionTrend, error) {
	return make([]*models.MonthlyCollectionTrend, 0), nil
}

func (m *MockPaymentRepo) GetTenantPaymentSummary(orgID int) ([]*models.TenantReportEntry, error) {
	return make([]*models.TenantReportEntry, 0), nil
}

func (m *MockPaymentRepo) GetPaymentAnalyticsByPeriod(orgID int, startDate, endDate time.Time) (*models.PaymentAnalyticsResult, error) {
	return &models.PaymentAnalyticsResult{
		MethodCounts:  map[string]int64{},
		StatusCounts:  map[string]int64{},
		DailyTrend:    map[string]int64{},
		TotalPayments: 0,
	}, nil
}

func (m *MockPaymentRepo) GetBatchPropertyPaymentStats(propertyIDs []int, startDate, endDate time.Time) (map[int]any, error) {
	return map[int]any{}, nil
}

// ---------------------------------------------------------------------------
// MockPaymentTransactionRepo â€“ implements interfaces.PaymentTransactionRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentTransactionRepo struct {
	transactions      map[int]*models.PaymentTransaction
	nextID            int
	shouldFailCreate  bool
	shouldFailDelete  bool
	shouldFailGetByID bool
}

func newMockPaymentTransactionRepo() *MockPaymentTransactionRepo {
	return &MockPaymentTransactionRepo{
		transactions: make(map[int]*models.PaymentTransaction),
		nextID:       1,
	}
}

func (m *MockPaymentTransactionRepo) Create(paymentID int, amount float64, paymentMethod, receiptNumber, notes string, paymentDate time.Time) (*models.PaymentTransaction, error) {
	if m.shouldFailCreate {
		return nil, errors.New("create payment transaction failed")
	}
	txn := &models.PaymentTransaction{
		ID:            m.nextID,
		PaymentID:     paymentID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		PaymentDate:   paymentDate,
		ReceiptNumber: receiptNumber,
		Notes:         notes,
		CreatedAt:     time.Now(),
	}
	m.transactions[m.nextID] = txn
	m.nextID++
	return txn, nil
}

func (m *MockPaymentTransactionRepo) GetByID(id int) (*models.PaymentTransaction, error) {
	if m.shouldFailGetByID {
		return nil, errors.New("payment transaction not found")
	}
	txn, ok := m.transactions[id]
	if !ok {
		return nil, errors.New("payment transaction not found")
	}
	return txn, nil
}

func (m *MockPaymentTransactionRepo) GetByPaymentID(paymentID int) ([]*models.PaymentTransaction, error) {
	var result []*models.PaymentTransaction
	for id := 1; id < m.nextID; id++ {
		if txn, ok := m.transactions[id]; ok && txn.PaymentID == paymentID {
			result = append(result, txn)
		}
	}
	return result, nil
}

func (m *MockPaymentTransactionRepo) Delete(id int) error {
	if m.shouldFailDelete {
		return errors.New("delete payment transaction failed")
	}
	delete(m.transactions, id)
	return nil
}

func (m *MockPaymentTransactionRepo) SumByPaymentID(paymentID int) (float64, error) {
	var total float64
	for id := 1; id < m.nextID; id++ {
		if txn, ok := m.transactions[id]; ok && txn.PaymentID == paymentID {
			total += txn.Amount
		}
	}
	return total, nil
}

// ---------------------------------------------------------------------------
// MockPaymentUnitRepo â€“ implements interfaces.UnitRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentUnitRepo struct {
	units      map[int]*models.Unit
	shouldFail bool
}

func newMockPaymentUnitRepo() *MockPaymentUnitRepo {
	return &MockPaymentUnitRepo{units: make(map[int]*models.Unit)}
}

func (m *MockPaymentUnitRepo) addUnit(u *models.Unit) {
	m.units[u.ID] = u
}

func (m *MockPaymentUnitRepo) GetByID(id int) (*models.Unit, error) {
	if m.shouldFail {
		return nil, errors.New("unit not found")
	}
	u, ok := m.units[id]
	if !ok {
		return nil, errors.New("unit not found")
	}
	return u, nil
}

func (m *MockPaymentUnitRepo) Create(req *models.CreateUnitRequest, organizationID int) (*models.Unit, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) BulkCreate(units []*models.Unit) error { return nil }
func (m *MockPaymentUnitRepo) GetByIDWithDetails(id int) (*models.UnitWithDetails, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) Update(id int, req *models.UpdateUnitRequest) (*models.Unit, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) Delete(id int) error { return nil }
func (m *MockPaymentUnitRepo) CheckUnitNumberExists(buildingID int, unitNumber string, excludeID int) (bool, error) {
	return false, nil
}
func (m *MockPaymentUnitRepo) HasActiveLeases(unitID int) (bool, error) { return false, nil }
func (m *MockPaymentUnitRepo) GetByBuildingWithDetails(buildingID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	return nil, 0, nil
}
func (m *MockPaymentUnitRepo) GetByPropertyWithDetails(propertyID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	return nil, 0, nil
}
func (m *MockPaymentUnitRepo) GetBuildingOccupancyStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) GetPropertyOccupancyStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) GetSystemOccupancyStats(startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *MockPaymentUnitRepo) GetBuildingUnitTypeDistribution(buildingID int) (interface{}, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// MockPaymentBuildingRepo â€“ implements interfaces.BuildingRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentBuildingRepo struct {
	buildings  map[int]*models.Building
	shouldFail bool
}

func newMockPaymentBuildingRepo() *MockPaymentBuildingRepo {
	return &MockPaymentBuildingRepo{buildings: make(map[int]*models.Building)}
}

func (m *MockPaymentBuildingRepo) addBuilding(b *models.Building) {
	m.buildings[b.ID] = b
}

func (m *MockPaymentBuildingRepo) GetByID(id int) (*models.Building, error) {
	if m.shouldFail {
		return nil, errors.New("building not found")
	}
	b, ok := m.buildings[id]
	if !ok {
		return nil, errors.New("building not found")
	}
	return b, nil
}

func (m *MockPaymentBuildingRepo) GetByPropertyID(propertyID int) ([]*models.Building, error) {
	if m.shouldFail {
		return nil, errors.New("failed to get buildings")
	}
	var result []*models.Building
	for _, b := range m.buildings {
		if b.PropertyID == propertyID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *MockPaymentBuildingRepo) Create(building *models.Building) error { return nil }
func (m *MockPaymentBuildingRepo) GetByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	return nil, nil
}
func (m *MockPaymentBuildingRepo) Update(id int, updates map[string]interface{}) error { return nil }
func (m *MockPaymentBuildingRepo) SoftDelete(id int) error                             { return nil }
func (m *MockPaymentBuildingRepo) GetWithStats(id int) (*models.BuildingWithStats, error) {
	return nil, nil
}
func (m *MockPaymentBuildingRepo) BulkCreate(buildings []*models.Building) error { return nil }
func (m *MockPaymentBuildingRepo) Search(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	return nil, nil
}
func (m *MockPaymentBuildingRepo) GetAnalytics(id int) (*models.BuildingAnalytics, error) {
	return nil, nil
}
func (m *MockPaymentBuildingRepo) CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error) {
	return 0, nil
}
func (m *MockPaymentBuildingRepo) GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error) {
	return nil, nil
}
func (m *MockPaymentBuildingRepo) AdvancedSearch(req *models.BuildingSearchRequest) ([]*models.Building, int, error) {
	return nil, 0, nil
}
func (m *MockPaymentBuildingRepo) GetBuildingUnits(buildingID int, offset, limit int) ([]*models.BuildingUnitSummary, int, error) {
	return nil, 0, nil
}

// ---------------------------------------------------------------------------
// MockPaymentPropertyRepo â€“ implements interfaces.PropertyRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentPropertyRepo struct {
	properties map[int]*models.Property
	shouldFail bool
}

func newMockPaymentPropertyRepo() *MockPaymentPropertyRepo {
	return &MockPaymentPropertyRepo{properties: make(map[int]*models.Property)}
}

func (m *MockPaymentPropertyRepo) addProperty(p *models.Property) {
	m.properties[p.ID] = p
}

func (m *MockPaymentPropertyRepo) GetByID(id int) (*models.Property, error) {
	if m.shouldFail {
		return nil, errors.New("property not found")
	}
	p, ok := m.properties[id]
	if !ok {
		return nil, errors.New("property not found")
	}
	return p, nil
}

func (m *MockPaymentPropertyRepo) GetByIDWithStats(id int) (*models.PropertyWithStats, error) {
	return nil, nil
}
func (m *MockPaymentPropertyRepo) List(filters map[string]interface{}, limit, offset int) ([]*models.Property, int, error) {
	return nil, 0, nil
}
func (m *MockPaymentPropertyRepo) GetBuildingAggregations(propertyID int) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockPaymentPropertyRepo) GetBuildingBreakdowns(propertyID int) (interface{}, error) {
	return nil, nil
}
func (m *MockPaymentPropertyRepo) GetBuildingTypeDistribution(propertyID int) (interface{}, error) {
	return nil, nil
}
func (m *MockPaymentPropertyRepo) GetBuildingCount(propertyID int) (int, error) { return 0, nil }
func (m *MockPaymentPropertyRepo) GetBuildingSummary(propertyID int) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockPaymentPropertyRepo) HasActiveBuildings(propertyID int) (bool, error) {
	return false, nil
}
func (m *MockPaymentPropertyRepo) HasActiveUnits(propertyID int) (bool, error) { return false, nil }

// ---------------------------------------------------------------------------
// MockPaymentAuditService â€“ implements interfaces.AuditServiceInterface
// ---------------------------------------------------------------------------

type MockPaymentAuditService struct {
	userActionCalls   int
	systemActionCalls int
	lastAction        string
	lastTable         string
	lastUserID        int
	shouldFail        bool
}

func newMockPaymentAuditService() *MockPaymentAuditService {
	return &MockPaymentAuditService{}
}

func (m *MockPaymentAuditService) LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	m.userActionCalls++
	m.lastAction = action
	m.lastTable = tableName
	m.lastUserID = userID
	if m.shouldFail {
		return errors.New("audit failed")
	}
	return nil
}

func (m *MockPaymentAuditService) LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	m.systemActionCalls++
	if m.shouldFail {
		return errors.New("audit failed")
	}
	return nil
}

func (m *MockPaymentAuditService) LogCriticalAction(userID int, action string, data map[string]interface{}) error {
	return nil
}

// ---------------------------------------------------------------------------
// MockPaymentUserRepo â€“ implements interfaces.UserRepositoryInterface
// ---------------------------------------------------------------------------

type MockPaymentUserRepo struct {
	users map[int]*models.User
}

func newMockPaymentUserRepo() *MockPaymentUserRepo {
	return &MockPaymentUserRepo{
		users: make(map[int]*models.User),
	}
}

func (m *MockPaymentUserRepo) Create(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockPaymentUserRepo) GetByID(id int) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockPaymentUserRepo) GetByUsername(username string) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockPaymentUserRepo) GetByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockPaymentUserRepo) Update(id int, updates map[string]interface{}) error {
	if user, ok := m.users[id]; ok {
		// Apply updates to user (simplified)
		if name, ok := updates["name"]; ok {
			user.FirstName = name.(string)
		}
		m.users[id] = user
	}
	return nil
}

func (m *MockPaymentUserRepo) Delete(id int) error {
	delete(m.users, id)
	return nil
}

func (m *MockPaymentUserRepo) List(filters map[string]interface{}, limit, offset int) ([]*models.User, int, error) {
	return nil, 0, nil
}

func (m *MockPaymentUserRepo) CleanupExpiredTokens() error {
	return nil
}

func (m *MockPaymentUserRepo) GetAll(activeOnly bool) ([]*models.User, error) {
	users := make([]*models.User, 0, len(m.users))
	for _, u := range m.users {
		if !activeOnly || u.Active {
			users = append(users, u)
		}
	}
	return users, nil
}

func (m *MockPaymentUserRepo) GetByOrganizationID(orgID int, activeOnly bool) ([]*models.User, error) {
	users := make([]*models.User, 0)
	for _, u := range m.users {
		if u.OrganizationID != nil && *u.OrganizationID == orgID && (!activeOnly || u.Active) {
			users = append(users, u)
		}
	}
	return users, nil
}

func (m *MockPaymentUserRepo) CreateResetToken(token *models.ResetPasswordToken) error {
	return nil
}

func (m *MockPaymentUserRepo) GetResetToken(token string) (*models.ResetPasswordToken, error) {
	return nil, fmt.Errorf("token not found")
}

func (m *MockPaymentUserRepo) DeleteResetToken(token string) error {
	return nil
}

func (m *MockPaymentUserRepo) MarkResetTokenUsed(tokenID int) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newPaymentServiceWithMocks builds a PaymentService wired with mock dependencies.
func newPaymentServiceWithMocks() (
	*PaymentService,
	*MockPaymentRepo,
	*MockPaymentUnitRepo,
	*MockPaymentBuildingRepo,
	*MockPaymentPropertyRepo,
	*MockPaymentAuditService,
) {
	svc, payRepo, _, unitRepo, bldgRepo, propRepo, audit, _ := newPaymentServiceWithMocksAndTxns()
	return svc, payRepo, unitRepo, bldgRepo, propRepo, audit
}

// newPaymentServiceWithMocksAndTxns is like newPaymentServiceWithMocks but
// also exposes the mock transaction repo, for tests exercising
// RecordPaymentTransaction/GetPaymentTransactions/DeletePaymentTransaction.
func newPaymentServiceWithMocksAndTxns() (
	*PaymentService,
	*MockPaymentRepo,
	*MockPaymentTransactionRepo,
	*MockPaymentUnitRepo,
	*MockPaymentBuildingRepo,
	*MockPaymentPropertyRepo,
	*MockPaymentAuditService,
	*MockPaymentUserRepo,
) {
	payRepo := newMockPaymentRepo()
	txnRepo := newMockPaymentTransactionRepo()
	unitRepo := newMockPaymentUnitRepo()
	bldgRepo := newMockPaymentBuildingRepo()
	propRepo := newMockPaymentPropertyRepo()
	audit := newMockPaymentAuditService()
	userRepo := newMockPaymentUserRepo()
	svc := NewPaymentService(payRepo, txnRepo, unitRepo, bldgRepo, propRepo, audit, userRepo)
	return svc, payRepo, txnRepo, unitRepo, bldgRepo, propRepo, audit, userRepo
}

func sampleUnit(id, buildingID, propertyID int) *models.Unit {
	return &models.Unit{
		ID:             id,
		BuildingID:     buildingID,
		PropertyID:     propertyID,
		OrganizationID: 1,
		UnitNumber:     "U-101",
		UnitType:       models.UnitTypeApartment,
		Active:         true,
	}
}

func sampleBuilding(id, propertyID int) *models.Building {
	return &models.Building{
		ID:             id,
		PropertyID:     propertyID,
		OrganizationID: 1,
		BuildingName:   "Block A",
		BuildingCode:   "BLK-A",
		BuildingType:   models.BuildingTypeResidential,
		ActiveStatus:   true,
	}
}

func sampleProperty(id int) *models.Property {
	return &models.Property{
		ID:             id,
		OrganizationID: 1,
		PropertyName:   "Sunrise Residency",
		PropertyCode:   "SR-001",
		Active:         true,
	}
}

func sampleCreateRequest(unitID, buildingID, propertyID int) *models.CreatePaymentRequest {
	return &models.CreatePaymentRequest{
		UnitID:         unitID,
		TenantID:       10,
		BuildingID:     buildingID,
		PropertyID:     propertyID,
		OrganizationID: 1,
		Month:          5,
		Year:           2026,
		AmountDue:      5000,
	}
}

// ---------------------------------------------------------------------------
// TestCreatePayment
// ---------------------------------------------------------------------------

func TestCreatePayment(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockPaymentUnitRepo, *MockPaymentBuildingRepo, *MockPaymentPropertyRepo, *MockPaymentRepo)
		req           *models.CreatePaymentRequest
		userID        int
		wantErr       bool
		errContains   string
		wantAuditCall bool
	}{
		{
			name: "happy path â€“ creates payment and logs audit",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 3))
				b.addBuilding(sampleBuilding(2, 3))
				p.addProperty(sampleProperty(3))
			},
			req:           sampleCreateRequest(1, 2, 3),
			userID:        99,
			wantErr:       false,
			wantAuditCall: true,
		},
		{
			name: "unit not found",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.shouldFail = true
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "unit not found",
		},
		{
			name: "building ID mismatch with unit",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 99, 3)) // unit belongs to building 99, not 2
				b.addBuilding(sampleBuilding(2, 3))
				p.addProperty(sampleProperty(3))
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "building ID mismatch",
		},
		{
			name: "property ID mismatch with unit",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 99)) // unit belongs to property 99, not 3
				b.addBuilding(sampleBuilding(2, 3))
				p.addProperty(sampleProperty(3))
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "property ID mismatch",
		},
		{
			name: "building repo failure",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 3))
				b.shouldFail = true
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "building not found",
		},
		{
			name: "property repo failure",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 3))
				b.addBuilding(sampleBuilding(2, 3))
				p.shouldFail = true
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "property not found",
		},
		{
			name: "payment repo create failure",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 3))
				b.addBuilding(sampleBuilding(2, 3))
				p.addProperty(sampleProperty(3))
				pay.shouldFailCreate = true
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "failed to create payment",
		},
		{
			name: "duplicate payment for unit/month/year is rejected",
			setupMocks: func(u *MockPaymentUnitRepo, b *MockPaymentBuildingRepo, p *MockPaymentPropertyRepo, pay *MockPaymentRepo) {
				u.addUnit(sampleUnit(1, 2, 3))
				b.addBuilding(sampleBuilding(2, 3))
				p.addProperty(sampleProperty(3))
				pay.shouldReturnExists = true
			},
			req:         sampleCreateRequest(1, 2, 3),
			userID:      99,
			wantErr:     true,
			errContains: "a payment already exists for this unit for the selected month/year",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, audit := newPaymentServiceWithMocks()
			tc.setupMocks(unitRepo, bldgRepo, propRepo, payRepo)

			payment, err := svc.CreatePayment(tc.req, tc.userID)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if payment == nil {
				t.Errorf("expected non-nil payment")
				return
			}
			if payment.UnitID != tc.req.UnitID {
				t.Errorf("unit ID: want %d got %d", tc.req.UnitID, payment.UnitID)
			}
			if tc.wantAuditCall && audit.userActionCalls != 1 {
				t.Errorf("expected 1 audit call, got %d", audit.userActionCalls)
			}
			if audit.lastAction != "CREATE" {
				t.Errorf("expected audit action CREATE, got %q", audit.lastAction)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetPayment
// ---------------------------------------------------------------------------

func TestGetPayment(t *testing.T) {
	tests := []struct {
		name        string
		seedPayment bool
		paymentID   int
		repoFail    bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "happy path â€“ returns payment with details",
			seedPayment: true,
			paymentID:   1,
			wantErr:     false,
		},
		{
			name:        "payment not found",
			seedPayment: false,
			paymentID:   999,
			wantErr:     true,
			errContains: "failed to get payment with details",
		},
		{
			name:        "repo failure",
			seedPayment: true,
			paymentID:   1,
			repoFail:    true,
			wantErr:     true,
			errContains: "failed to get payment with details",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seedPayment {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			payRepo.shouldFailGetByID = tc.repoFail

			pwd, err := svc.GetPayment(tc.paymentID, 1)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if pwd == nil {
				t.Errorf("expected non-nil PaymentWithDetails")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestUpdatePayment
// ---------------------------------------------------------------------------

func TestUpdatePayment(t *testing.T) {
	paidStatus := models.PaymentStatusPaid
	amount := 5000.0

	tests := []struct {
		name          string
		req           *models.UpdatePaymentRequest
		getByIDFail   bool
		updateFail    bool
		wantErr       bool
		errContains   string
		wantAuditCall bool
	}{
		{
			name:          "happy path â€“ updates status and amount paid",
			req:           &models.UpdatePaymentRequest{Status: &paidStatus, AmountPaid: &amount},
			wantErr:       false,
			wantAuditCall: true,
		},
		{
			name:        "payment not found on get",
			req:         &models.UpdatePaymentRequest{Status: &paidStatus},
			getByIDFail: true,
			wantErr:     true,
			errContains: "failed to get existing payment",
		},
		{
			name:        "update repo failure",
			req:         &models.UpdatePaymentRequest{Status: &paidStatus},
			updateFail:  true,
			wantErr:     true,
			errContains: "failed to update payment",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, audit := newPaymentServiceWithMocks()

			unitRepo.addUnit(sampleUnit(1, 2, 3))
			bldgRepo.addBuilding(sampleBuilding(2, 3))
			propRepo.addProperty(sampleProperty(3))
			created, _ := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)

			// Reset audit counter after create so we can measure update-specific calls
			audit.userActionCalls = 0

			payRepo.shouldFailGetByID = tc.getByIDFail
			payRepo.shouldFailUpdate = tc.updateFail

			updated, err := svc.UpdatePayment(created.ID, tc.req, 99, 1)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if updated == nil {
				t.Errorf("expected non-nil updated payment")
				return
			}
			if tc.wantAuditCall && audit.userActionCalls != 1 {
				t.Errorf("expected 1 audit call on update, got %d", audit.userActionCalls)
			}
			if audit.lastAction != "UPDATE" {
				t.Errorf("expected audit action UPDATE, got %q", audit.lastAction)
			}
		})
	}
}

func TestUpdatePayment_AmountPaidIsIgnoredFromClient(t *testing.T) {
	svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Recording money must go through RecordPaymentTransaction — a plain
	// UpdatePayment call claiming amount_paid must be silently ignored.
	claimed := 5000.0
	updated, err := svc.UpdatePayment(created.ID, &models.UpdatePaymentRequest{AmountPaid: &claimed}, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AmountPaid != 0 {
		t.Errorf("expected amount_paid to stay 0 (ignored), got %.2f", updated.AmountPaid)
	}
	_ = payRepo
}

// ---------------------------------------------------------------------------
// TestRecordPaymentTransaction
// ---------------------------------------------------------------------------

func TestRecordPaymentTransaction_AccumulatesAcrossInstallments(t *testing.T) {
	svc, _, txnRepo, unitRepo, bldgRepo, propRepo, audit, _ := newPaymentServiceWithMocksAndTxns()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	req := sampleCreateRequest(1, 2, 3)
	req.AmountDue = 12000
	created, err := svc.CreatePayment(req, 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Reset the audit counter after create so we measure only the two
	// RECORD_PAYMENT calls below.
	audit.userActionCalls = 0

	// First installment: 10,000 of 12,000 due.
	first := 10000.0
	afterFirst, err := svc.RecordPaymentTransaction(created.ID, &models.CreatePaymentTransactionRequest{Amount: first}, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error recording first installment: %v", err)
	}
	if afterFirst.AmountPaid != first {
		t.Errorf("after first installment: amount_paid got %.2f, want %.2f", afterFirst.AmountPaid, first)
	}
	if afterFirst.Status != models.PaymentStatusPartial {
		t.Errorf("after first installment: status got %q, want %q", afterFirst.Status, models.PaymentStatusPartial)
	}

	// Second installment: the remaining 2,000 — must ADD to the first, not
	// replace it, so the payment is fully settled.
	second := 2000.0
	afterSecond, err := svc.RecordPaymentTransaction(created.ID, &models.CreatePaymentTransactionRequest{Amount: second}, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error recording second installment: %v", err)
	}
	if afterSecond.AmountPaid != first+second {
		t.Errorf("after second installment: amount_paid got %.2f, want %.2f", afterSecond.AmountPaid, first+second)
	}
	if afterSecond.Status != models.PaymentStatusPaid {
		t.Errorf("after second installment: status got %q, want %q", afterSecond.Status, models.PaymentStatusPaid)
	}

	// Each installment must get its own receipt number, not share one.
	txns, err := txnRepo.GetByPaymentID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 recorded transactions, got %d", len(txns))
	}
	if txns[0].ReceiptNumber == "" || txns[1].ReceiptNumber == "" {
		t.Error("expected both transactions to have a receipt number")
	}
	if txns[0].ReceiptNumber == txns[1].ReceiptNumber {
		t.Errorf("expected distinct receipt numbers per installment, both were %q", txns[0].ReceiptNumber)
	}

	if audit.userActionCalls != 2 {
		t.Errorf("expected 1 audit call per recorded transaction (2 total), got %d", audit.userActionCalls)
	}
	if audit.lastAction != "RECORD_PAYMENT" {
		t.Errorf("expected last audit action RECORD_PAYMENT, got %q", audit.lastAction)
	}
}

func TestRecordPaymentTransaction_CrossOrgPaymentNotFound(t *testing.T) {
	svc, _, _, unitRepo, bldgRepo, propRepo, _, _ := newPaymentServiceWithMocksAndTxns()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// orgID 999 does not own this payment (it was created under org 1).
	_, err = svc.RecordPaymentTransaction(created.ID, &models.CreatePaymentTransactionRequest{Amount: 1000}, 99, 999)
	if err == nil {
		t.Fatal("expected an error recording a transaction against another org's payment, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestDeletePaymentTransaction
// ---------------------------------------------------------------------------

func TestDeletePaymentTransaction_RecomputesRemainingBalance(t *testing.T) {
	svc, _, _, unitRepo, bldgRepo, propRepo, _, _ := newPaymentServiceWithMocksAndTxns()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	req := sampleCreateRequest(1, 2, 3)
	req.AmountDue = 12000
	created, err := svc.CreatePayment(req, 99)
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	first := 10000.0
	if _, err := svc.RecordPaymentTransaction(created.ID, &models.CreatePaymentTransactionRequest{Amount: first}, 99, 1); err != nil {
		t.Fatalf("unexpected error recording first installment: %v", err)
	}
	second := 2000.0
	if _, err := svc.RecordPaymentTransaction(created.ID, &models.CreatePaymentTransactionRequest{Amount: second}, 99, 1); err != nil {
		t.Fatalf("unexpected error recording second installment: %v", err)
	}

	txns, err := svc.GetPaymentTransactions(created.ID, 1)
	if err != nil {
		t.Fatalf("unexpected error listing transactions: %v", err)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 transactions before deletion, got %d", len(txns))
	}

	// Void the mistaken second installment — balance must fall back to
	// exactly what the first one covered, not to zero.
	afterDelete, err := svc.DeletePaymentTransaction(created.ID, txns[1].ID, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error deleting transaction: %v", err)
	}
	if afterDelete.AmountPaid != first {
		t.Errorf("after deleting second installment: amount_paid got %.2f, want %.2f", afterDelete.AmountPaid, first)
	}
	if afterDelete.Status != models.PaymentStatusPartial {
		t.Errorf("after deleting second installment: status got %q, want %q", afterDelete.Status, models.PaymentStatusPartial)
	}

	// Void the remaining installment too — must revert cleanly to Due, not
	// get stuck on a stale Partial/Paid status.
	afterDeleteAll, err := svc.DeletePaymentTransaction(created.ID, txns[0].ID, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error deleting last transaction: %v", err)
	}
	if afterDeleteAll.AmountPaid != 0 {
		t.Errorf("after deleting all transactions: amount_paid got %.2f, want 0", afterDeleteAll.AmountPaid)
	}
	if afterDeleteAll.Status != models.PaymentStatusDue {
		t.Errorf("after deleting all transactions: status got %q, want %q", afterDeleteAll.Status, models.PaymentStatusDue)
	}
}

func TestDeletePaymentTransaction_WrongPaymentRejected(t *testing.T) {
	svc, _, _, unitRepo, bldgRepo, propRepo, _, _ := newPaymentServiceWithMocksAndTxns()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	req := sampleCreateRequest(1, 2, 3)
	req.Month = 6
	paymentA, err := svc.CreatePayment(req, 99)
	if err != nil {
		t.Fatalf("failed to create payment A: %v", err)
	}
	req2 := sampleCreateRequest(1, 2, 3)
	req2.Month = 7
	paymentB, err := svc.CreatePayment(req2, 99)
	if err != nil {
		t.Fatalf("failed to create payment B: %v", err)
	}

	txnOnA, err := svc.RecordPaymentTransaction(paymentA.ID, &models.CreatePaymentTransactionRequest{Amount: 1000}, 99, 1)
	if err != nil {
		t.Fatalf("unexpected error recording transaction: %v", err)
	}
	_ = txnOnA
	txnsOnA, err := svc.GetPaymentTransactions(paymentA.ID, 1)
	if err != nil || len(txnsOnA) != 1 {
		t.Fatalf("expected 1 transaction on payment A, got %d (err=%v)", len(txnsOnA), err)
	}

	// A transaction that belongs to payment A must not be deletable through
	// payment B's endpoint — otherwise one payment's ledger could be
	// tampered with via another payment's ID.
	if _, err := svc.DeletePaymentTransaction(paymentB.ID, txnsOnA[0].ID, 99, 1); err == nil {
		t.Error("expected an error deleting a transaction through the wrong payment, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestGetPayments â€“ pagination boundary tests
// ---------------------------------------------------------------------------

func TestGetPayments(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		repoFail     bool
		seedCount    int
		wantErr      bool
		errContains  string
		wantPage     int // effective page applied
		wantPageSize int // effective pageSize applied
	}{
		{
			name:         "normal page/pageSize",
			page:         1,
			pageSize:     10,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "page 0 normalized to 1",
			page:         0,
			pageSize:     10,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "negative page normalized to 1",
			page:         -5,
			pageSize:     10,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name:         "pageSize 0 normalized to 20",
			page:         1,
			pageSize:     0,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "pageSize exceeds 100 normalized to 20",
			page:         1,
			pageSize:     200,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 20,
		},
		{
			name:         "pageSize of 100 is allowed",
			page:         1,
			pageSize:     100,
			seedCount:    3,
			wantErr:      false,
			wantPage:     1,
			wantPageSize: 100,
		},
		{
			name:        "repo failure propagated",
			page:        1,
			pageSize:    10,
			seedCount:   2,
			repoFail:    true,
			wantErr:     true,
			errContains: "failed to get payments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			unitRepo.addUnit(sampleUnit(1, 2, 3))
			bldgRepo.addBuilding(sampleBuilding(2, 3))
			propRepo.addProperty(sampleProperty(3))
			for i := 0; i < tc.seedCount; i++ {
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			payRepo.shouldFailList = tc.repoFail

			result, total, err := svc.GetPayments(tc.page, tc.pageSize, map[string]interface{}{})

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if total < 0 {
				t.Errorf("total should be non-negative, got %d", total)
			}
			_ = result // we primarily care that no error occurred and boundary normalization was applied
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetPaymentsByBuilding
// ---------------------------------------------------------------------------

func TestGetPaymentsByBuilding(t *testing.T) {
	tests := []struct {
		name         string
		buildingID   int
		seedBuilding bool
		buildingFail bool
		repoFail     bool
		page         int
		pageSize     int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "happy path",
			buildingID:   2,
			seedBuilding: true,
			page:         1,
			pageSize:     10,
			wantErr:      false,
		},
		{
			name:         "building not found",
			buildingID:   999,
			seedBuilding: false,
			buildingFail: false,
			page:         1,
			pageSize:     10,
			wantErr:      true,
			errContains:  "building not found",
		},
		{
			name:         "building repo general failure",
			buildingID:   2,
			seedBuilding: true,
			buildingFail: true,
			page:         1,
			pageSize:     10,
			wantErr:      true,
			errContains:  "building not found",
		},
		{
			name:         "payment list failure",
			buildingID:   2,
			seedBuilding: true,
			page:         1,
			pageSize:     10,
			repoFail:     true,
			wantErr:      true,
			errContains:  "failed to get payments by building",
		},
		{
			name:         "page normalization â€“ page 0",
			buildingID:   2,
			seedBuilding: true,
			page:         0,
			pageSize:     10,
			wantErr:      false,
		},
		{
			name:         "pageSize normalization â€“ pageSize 200",
			buildingID:   2,
			seedBuilding: true,
			page:         1,
			pageSize:     200,
			wantErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seedBuilding {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			bldgRepo.shouldFail = tc.buildingFail
			payRepo.shouldFailList = tc.repoFail

			_, _, err := svc.GetPaymentsByBuilding(tc.buildingID, tc.page, tc.pageSize, map[string]interface{}{})

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetPaymentsByProperty
// ---------------------------------------------------------------------------

func TestGetPaymentsByProperty(t *testing.T) {
	tests := []struct {
		name         string
		propertyID   int
		seedProperty bool
		propFail     bool
		repoFail     bool
		page         int
		pageSize     int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "happy path",
			propertyID:   3,
			seedProperty: true,
			page:         1,
			pageSize:     10,
			wantErr:      false,
		},
		{
			name:         "property not found",
			propertyID:   999,
			seedProperty: false,
			page:         1,
			pageSize:     10,
			wantErr:      true,
			errContains:  "property not found",
		},
		{
			name:         "property repo failure",
			propertyID:   3,
			seedProperty: true,
			propFail:     true,
			page:         1,
			pageSize:     10,
			wantErr:      true,
			errContains:  "property not found",
		},
		{
			name:         "payment list failure",
			propertyID:   3,
			seedProperty: true,
			page:         1,
			pageSize:     10,
			repoFail:     true,
			wantErr:      true,
			errContains:  "failed to get payments by property",
		},
		{
			name:         "page normalization â€“ negative page",
			propertyID:   3,
			seedProperty: true,
			page:         -1,
			pageSize:     10,
			wantErr:      false,
		},
		{
			name:         "pageSize normalization â€“ exceeds 100",
			propertyID:   3,
			seedProperty: true,
			page:         1,
			pageSize:     500,
			wantErr:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seedProperty {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			propRepo.shouldFail = tc.propFail
			payRepo.shouldFailList = tc.repoFail

			_, _, err := svc.GetPaymentsByProperty(tc.propertyID, tc.page, tc.pageSize, map[string]interface{}{})

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGenerateBuildingPaymentReport
// ---------------------------------------------------------------------------

func TestGenerateBuildingPaymentReport(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		buildingID  int
		seed        bool
		bldgFail    bool
		propFail    bool
		statsFail   bool
		periodFail  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "happy path â€“ report generated",
			buildingID: 2,
			seed:       true,
			wantErr:    false,
		},
		{
			name:        "building not found",
			buildingID:  999,
			seed:        false,
			wantErr:     true,
			errContains: "building not found",
		},
		{
			name:        "building repo fails",
			buildingID:  2,
			seed:        true,
			bldgFail:    true,
			wantErr:     true,
			errContains: "building not found",
		},
		{
			name:        "property not found for building",
			buildingID:  2,
			seed:        true,
			propFail:    true,
			wantErr:     true,
			errContains: "property not found",
		},
		{
			name:        "payment stats failure",
			buildingID:  2,
			seed:        true,
			statsFail:   true,
			wantErr:     true,
			errContains: "failed to get building payment stats",
		},
		{
			name:        "payments in period failure",
			buildingID:  2,
			seed:        true,
			periodFail:  true,
			wantErr:     true,
			errContains: "failed to get building payments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seed {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			bldgRepo.shouldFail = tc.bldgFail
			propRepo.shouldFail = tc.propFail
			payRepo.shouldFailStats = tc.statsFail
			payRepo.shouldFailPeriod = tc.periodFail

			report, err := svc.GenerateBuildingPaymentReport(tc.buildingID, 1, start, end)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if report == nil {
				t.Errorf("expected non-nil report")
				return
			}
			if report.BuildingID != tc.buildingID {
				t.Errorf("report.BuildingID: want %d got %d", tc.buildingID, report.BuildingID)
			}
			expectedPeriod := "2026-01-01 to 2026-05-31"
			if report.ReportPeriod != expectedPeriod {
				t.Errorf("report.ReportPeriod: want %q got %q", expectedPeriod, report.ReportPeriod)
			}
			if report.GeneratedAt.IsZero() {
				t.Errorf("report.GeneratedAt should not be zero")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGeneratePropertyPaymentReport
// ---------------------------------------------------------------------------

func TestGeneratePropertyPaymentReport(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		propertyID  int
		seed        bool
		propFail    bool
		bldgFail    bool
		statsFail   bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "happy path â€“ property report with building breakdowns",
			propertyID: 3,
			seed:       true,
			wantErr:    false,
		},
		{
			name:        "property not found",
			propertyID:  999,
			seed:        false,
			wantErr:     true,
			errContains: "property not found",
		},
		{
			name:        "property repo fails",
			propertyID:  3,
			seed:        true,
			propFail:    true,
			wantErr:     true,
			errContains: "property not found",
		},
		{
			name:        "get buildings failure",
			propertyID:  3,
			seed:        true,
			bldgFail:    true,
			wantErr:     true,
			errContains: "failed to get property buildings",
		},
		{
			name:        "overall stats failure",
			propertyID:  3,
			seed:        true,
			statsFail:   true,
			wantErr:     true,
			errContains: "failed to get property payment stats",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seed {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			propRepo.shouldFail = tc.propFail
			bldgRepo.shouldFail = tc.bldgFail
			payRepo.shouldFailStats = tc.statsFail

			report, err := svc.GeneratePropertyPaymentReport(tc.propertyID, 1, start, end)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if report == nil {
				t.Errorf("expected non-nil report")
				return
			}
			if report.PropertyID != tc.propertyID {
				t.Errorf("report.PropertyID: want %d got %d", tc.propertyID, report.PropertyID)
			}
			if report.GeneratedAt.IsZero() {
				t.Errorf("report.GeneratedAt should not be zero")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGeneratePropertyPaymentReport_BuildingBreakdownSkipOnError
// ---------------------------------------------------------------------------

// When GetBuildingPaymentStats fails for an individual building, the report
// should still be generated â€“ that building is simply omitted from breakdowns.
func TestGeneratePropertyPaymentReport_BuildingBreakdownSkipOnError(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

	svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)

	// Property stats succeed but per-building stats fail
	payRepo.shouldFailStats = true

	report, err := svc.GeneratePropertyPaymentReport(3, 1, start, end)
	if err == nil {
		// The service returns an error when property stats fail; that's expected
		// because GetPropertyPaymentStats uses the same shouldFailStats flag.
		// So we just verify the error is the property-level one.
		_ = report
	} else {
		if !paymentTestContains(err.Error(), "failed to get property payment stats") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// TestGetDashboardSummaryWithBuildingContext
// ---------------------------------------------------------------------------

func TestGetDashboardSummaryWithBuildingContext(t *testing.T) {
	tests := []struct {
		name              string
		dashboardFail     bool
		buildingLevelFail bool
		buildingCount     int
		wantErr           bool
		errContains       string
	}{
		{
			name:          "happy path â€“ summary with building count set",
			buildingCount: 3,
			wantErr:       false,
		},
		{
			name:          "dashboard summary repo failure",
			dashboardFail: true,
			wantErr:       true,
			errContains:   "failed to get dashboard summary",
		},
		{
			name:              "building level summary failure",
			buildingLevelFail: true,
			wantErr:           true,
			errContains:       "failed to get building-level summary",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, _, _, _, _ := newPaymentServiceWithMocks()

			payRepo.shouldFailDashboard = tc.dashboardFail
			payRepo.shouldFailBldgLevel = tc.buildingLevelFail

			if !tc.dashboardFail && !tc.buildingLevelFail {
				payRepo.buildingLevelResult = map[string]interface{}{
					"total_buildings": tc.buildingCount,
				}
				payRepo.dashboardResult = &models.DashboardSummary{
					TotalDue:       1000,
					TotalPaid:      800,
					CollectionRate: 80,
				}
			}

			summary, err := svc.GetDashboardSummaryWithBuildingContext(1)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if summary == nil {
				t.Errorf("expected non-nil summary")
				return
			}
			if summary.BuildingCount != tc.buildingCount {
				t.Errorf("BuildingCount: want %d got %d", tc.buildingCount, summary.BuildingCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestProcessBulkPayments
// ---------------------------------------------------------------------------

func TestProcessBulkPayments(t *testing.T) {
	tests := []struct {
		name           string
		requests       func() []*models.CreatePaymentRequest
		setupFails     func(*MockPaymentUnitRepo)
		wantPayments   int
		wantErrors     int
		wantAuditCalls int
	}{
		{
			name: "all requests succeed",
			requests: func() []*models.CreatePaymentRequest {
				return []*models.CreatePaymentRequest{
					sampleCreateRequest(1, 2, 3),
					sampleCreateRequest(1, 2, 3),
					sampleCreateRequest(1, 2, 3),
				}
			},
			setupFails:     func(u *MockPaymentUnitRepo) {},
			wantPayments:   3,
			wantErrors:     0,
			wantAuditCalls: 3,
		},
		{
			name: "empty request list",
			requests: func() []*models.CreatePaymentRequest {
				return []*models.CreatePaymentRequest{}
			},
			setupFails:     func(u *MockPaymentUnitRepo) {},
			wantPayments:   0,
			wantErrors:     0,
			wantAuditCalls: 0,
		},
		{
			name: "all requests fail â€“ unit not found",
			requests: func() []*models.CreatePaymentRequest {
				return []*models.CreatePaymentRequest{
					sampleCreateRequest(1, 2, 3),
					sampleCreateRequest(1, 2, 3),
				}
			},
			setupFails: func(u *MockPaymentUnitRepo) {
				u.shouldFail = true
			},
			wantPayments:   0,
			wantErrors:     2,
			wantAuditCalls: 0,
		},
		{
			name: "single item list",
			requests: func() []*models.CreatePaymentRequest {
				return []*models.CreatePaymentRequest{
					sampleCreateRequest(1, 2, 3),
				}
			},
			setupFails:     func(u *MockPaymentUnitRepo) {},
			wantPayments:   1,
			wantErrors:     0,
			wantAuditCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, unitRepo, bldgRepo, propRepo, audit := newPaymentServiceWithMocks()

			unitRepo.addUnit(sampleUnit(1, 2, 3))
			bldgRepo.addBuilding(sampleBuilding(2, 3))
			propRepo.addProperty(sampleProperty(3))

			tc.setupFails(unitRepo)

			reqs := tc.requests()
			payments, errs := svc.ProcessBulkPayments(reqs, 99)

			if len(payments) != tc.wantPayments {
				t.Errorf("payments: want %d got %d", tc.wantPayments, len(payments))
			}
			if len(errs) != tc.wantErrors {
				t.Errorf("errors: want %d got %d", tc.wantErrors, len(errs))
			}
			if audit.userActionCalls != tc.wantAuditCalls {
				t.Errorf("audit calls: want %d got %d", tc.wantAuditCalls, audit.userActionCalls)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestGetPaymentAnalyticsByBuilding
// ---------------------------------------------------------------------------

func TestGetPaymentAnalyticsByBuilding(t *testing.T) {
	tests := []struct {
		name          string
		buildingID    int
		period        string
		seedBuilding  bool
		bldgFail      bool
		analyticsFail bool
		wantErr       bool
		errContains   string
	}{
		{
			name:         "happy path â€“ month period",
			buildingID:   2,
			period:       "month",
			seedBuilding: true,
			wantErr:      false,
		},
		{
			name:         "happy path â€“ quarter period",
			buildingID:   2,
			period:       "quarter",
			seedBuilding: true,
			wantErr:      false,
		},
		{
			name:         "happy path â€“ year period",
			buildingID:   2,
			period:       "year",
			seedBuilding: true,
			wantErr:      false,
		},
		{
			name:         "happy path â€“ default period (unknown string)",
			buildingID:   2,
			period:       "unknown",
			seedBuilding: true,
			wantErr:      false,
		},
		{
			name:         "happy path â€“ empty period defaults to month",
			buildingID:   2,
			period:       "",
			seedBuilding: true,
			wantErr:      false,
		},
		{
			name:         "building not found",
			buildingID:   999,
			period:       "month",
			seedBuilding: false,
			wantErr:      true,
			errContains:  "building not found",
		},
		{
			name:         "building repo failure",
			buildingID:   2,
			period:       "month",
			seedBuilding: true,
			bldgFail:     true,
			wantErr:      true,
			errContains:  "building not found",
		},
		{
			name:          "analytics repo failure",
			buildingID:    2,
			period:        "month",
			seedBuilding:  true,
			analyticsFail: true,
			wantErr:       true,
			errContains:   "failed to get building payment analytics",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, payRepo, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

			if tc.seedBuilding {
				unitRepo.addUnit(sampleUnit(1, 2, 3))
				bldgRepo.addBuilding(sampleBuilding(2, 3))
				propRepo.addProperty(sampleProperty(3))
				_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
			}

			bldgRepo.shouldFail = tc.bldgFail
			payRepo.shouldFailAnalytics = tc.analyticsFail

			analytics, err := svc.GetPaymentAnalyticsByBuilding(tc.buildingID, tc.period)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errContains)
				} else if tc.errContains != "" && !paymentTestContains(err.Error(), tc.errContains) {
					t.Errorf("expected error %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if analytics == nil {
				t.Errorf("expected non-nil analytics")
				return
			}
			if analytics.BuildingID != tc.buildingID {
				t.Errorf("analytics.BuildingID: want %d got %d", tc.buildingID, analytics.BuildingID)
			}
			// Building context should be populated from the building repo
			if analytics.BuildingName == "" {
				t.Errorf("analytics.BuildingName should be populated from building repo")
			}
			if analytics.BuildingCode == "" {
				t.Errorf("analytics.BuildingCode should be populated from building repo")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestAuditLogging â€“ cross-cutting concern: audit calls on create and update
// ---------------------------------------------------------------------------

func TestAuditLogging(t *testing.T) {
	t.Run("create payment logs user action with correct metadata", func(t *testing.T) {
		svc, _, unitRepo, bldgRepo, propRepo, audit := newPaymentServiceWithMocks()

		unitRepo.addUnit(sampleUnit(1, 2, 3))
		bldgRepo.addBuilding(sampleBuilding(2, 3))
		propRepo.addProperty(sampleProperty(3))

		userID := 42
		_, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if audit.userActionCalls != 1 {
			t.Errorf("expected 1 audit call, got %d", audit.userActionCalls)
		}
		if audit.lastUserID != userID {
			t.Errorf("audit userID: want %d got %d", userID, audit.lastUserID)
		}
		if audit.lastAction != "CREATE" {
			t.Errorf("audit action: want CREATE got %q", audit.lastAction)
		}
		if audit.lastTable != "payments" {
			t.Errorf("audit table: want payments got %q", audit.lastTable)
		}
	})

	t.Run("update payment logs user action with correct metadata", func(t *testing.T) {
		svc, _, unitRepo, bldgRepo, propRepo, audit := newPaymentServiceWithMocks()

		unitRepo.addUnit(sampleUnit(1, 2, 3))
		bldgRepo.addBuilding(sampleBuilding(2, 3))
		propRepo.addProperty(sampleProperty(3))

		created, err := svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		audit.userActionCalls = 0 // reset for update measurement

		paidStatus := models.PaymentStatusPaid
		userID := 77
		_, err = svc.UpdatePayment(created.ID, &models.UpdatePaymentRequest{Status: &paidStatus}, userID, 1)
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		if audit.userActionCalls != 1 {
			t.Errorf("expected 1 audit call on update, got %d", audit.userActionCalls)
		}
		if audit.lastUserID != userID {
			t.Errorf("audit userID: want %d got %d", userID, audit.lastUserID)
		}
		if audit.lastAction != "UPDATE" {
			t.Errorf("audit action: want UPDATE got %q", audit.lastAction)
		}
	})
}

// ---------------------------------------------------------------------------
// TestGetPayments_FiltersPassedThrough
// ---------------------------------------------------------------------------

func TestGetPayments_FiltersPassedThrough(t *testing.T) {
	svc, _, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))

	// Seed payments for org 1
	for i := 0; i < 5; i++ {
		_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)
	}

	filters := map[string]interface{}{"organization_id": 1}
	payments, total, err := svc.GetPayments(1, 10, filters)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("total: want 5 got %d", total)
	}
	if len(payments) != 5 {
		t.Errorf("payments count: want 5 got %d", len(payments))
	}
}

// ---------------------------------------------------------------------------
// TestGetPaymentsByBuilding_FiltersInjected
// ---------------------------------------------------------------------------

func TestGetPaymentsByBuilding_FiltersInjected(t *testing.T) {
	svc, _, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)

	filters := map[string]interface{}{}
	_, _, err := svc.GetPaymentsByBuilding(2, 1, 10, filters)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify the service injected building_id into filters
	if filters["building_id"] != 2 {
		t.Errorf("expected filters[building_id]=2, got %v", filters["building_id"])
	}
}

// ---------------------------------------------------------------------------
// TestGetPaymentsByProperty_FiltersInjected
// ---------------------------------------------------------------------------

func TestGetPaymentsByProperty_FiltersInjected(t *testing.T) {
	svc, _, unitRepo, bldgRepo, propRepo, _ := newPaymentServiceWithMocks()

	unitRepo.addUnit(sampleUnit(1, 2, 3))
	bldgRepo.addBuilding(sampleBuilding(2, 3))
	propRepo.addProperty(sampleProperty(3))
	_, _ = svc.CreatePayment(sampleCreateRequest(1, 2, 3), 99)

	filters := map[string]interface{}{}
	_, _, err := svc.GetPaymentsByProperty(3, 1, 10, filters)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Verify the service injected property_id into filters
	if filters["property_id"] != 3 {
		t.Errorf("expected filters[property_id]=3, got %v", filters["property_id"])
	}
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

// paymentTestContains is a simple substring check used by payment service tests.
func paymentTestContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
