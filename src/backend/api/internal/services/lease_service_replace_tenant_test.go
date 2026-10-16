package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// ---------------------------------------------------------------------------
// MockTurnoverLeaseRepo - implements interfaces.LeaseRepositoryInterface
// ---------------------------------------------------------------------------

type MockTurnoverLeaseRepo struct {
	leases  map[int]*models.Lease
	details map[int]*models.LeaseWithDetails

	tenantsWithActiveLease map[int]bool

	replaceCalls     int
	lastHandover     time.Time
	lastSuccessor    *models.CreateLeaseRequest
	replaceFailsWith error
	nextLeaseID      int
}

func newMockTurnoverLeaseRepo() *MockTurnoverLeaseRepo {
	return &MockTurnoverLeaseRepo{
		leases:                 make(map[int]*models.Lease),
		details:                make(map[int]*models.LeaseWithDetails),
		tenantsWithActiveLease: make(map[int]bool),
		nextLeaseID:            900,
	}
}

func (m *MockTurnoverLeaseRepo) addLease(l *models.Lease) {
	m.leases[l.ID] = l
	m.details[l.ID] = &models.LeaseWithDetails{Lease: *l}
}

func (m *MockTurnoverLeaseRepo) GetByID(id int) (*models.Lease, error) {
	l, ok := m.leases[id]
	if !ok {
		return nil, errors.New("lease not found")
	}
	copied := *l
	return &copied, nil
}

func (m *MockTurnoverLeaseRepo) GetByIDWithDetails(id int) (*models.LeaseWithDetails, error) {
	d, ok := m.details[id]
	if !ok {
		return nil, errors.New("lease not found")
	}
	copied := *d
	return &copied, nil
}

func (m *MockTurnoverLeaseRepo) ReplaceTenant(oldLeaseID int, handoverDate time.Time, successor *models.CreateLeaseRequest) (*models.Lease, error) {
	m.replaceCalls++
	m.lastHandover = handoverDate
	m.lastSuccessor = successor

	if m.replaceFailsWith != nil {
		return nil, m.replaceFailsWith
	}

	old, ok := m.leases[oldLeaseID]
	if !ok || !old.Active {
		return nil, errors.New("lease is no longer active")
	}
	old.Active = false
	old.EndDate = handoverDate

	startDate, err := time.Parse("2006-01-02", successor.StartDate)
	if err != nil {
		return nil, err
	}

	m.nextLeaseID++
	created := &models.Lease{
		ID:              m.nextLeaseID,
		UnitID:          successor.UnitID,
		TenantID:        successor.TenantID,
		LeaseType:       successor.LeaseType,
		StartDate:       startDate,
		EndDate:         startDate.AddDate(0, successor.DurationMonths, 0),
		DurationMonths:  successor.DurationMonths,
		MonthlyRent:     successor.MonthlyRent,
		SecurityDeposit: successor.SecurityDeposit,
		Active:          true,
		OrganizationID:  successor.OrganizationID,
	}
	m.addLease(created)
	return created, nil
}

func (m *MockTurnoverLeaseRepo) HasActiveLeaseForTenant(tenantID int) (bool, error) {
	return m.tenantsWithActiveLease[tenantID], nil
}

func (m *MockTurnoverLeaseRepo) Create(req *models.CreateLeaseRequest) (*models.Lease, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseRepo) GetAll(page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	return nil, 0, nil
}
func (m *MockTurnoverLeaseRepo) GetByUnitID(unitID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	return nil, 0, nil
}
func (m *MockTurnoverLeaseRepo) GetByTenantID(tenantID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	return nil, 0, nil
}
func (m *MockTurnoverLeaseRepo) Update(id int, req *models.UpdateLeaseRequest) (*models.Lease, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseRepo) Delete(id int) error { return nil }
func (m *MockTurnoverLeaseRepo) SoftDelete(id int, endDate time.Time, reason models.LeaseEndReason) error {
	return nil
}
func (m *MockTurnoverLeaseRepo) RenewLease(oldLeaseID int, req *models.RenewLeaseRequest) (*models.Lease, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseRepo) HasActiveLeaseOnUnit(unitID int, excludeLeaseID *int) (bool, error) {
	return false, nil
}
func (m *MockTurnoverLeaseRepo) HasPayableLeaseForUnitAndTenant(unitID, tenantID int) (bool, error) {
	return false, nil
}
func (m *MockTurnoverLeaseRepo) GetLeasesDueForMonth(orgID int) ([]models.LeaseDue, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseRepo) GetDueSummary(orgID int) (*models.DueSummary, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// MockTurnoverTenantRepo - implements interfaces.TenantRepositoryInterface
// ---------------------------------------------------------------------------

type MockTurnoverTenantRepo struct {
	tenants map[int]*models.Tenant
}

func newMockTurnoverTenantRepo() *MockTurnoverTenantRepo {
	return &MockTurnoverTenantRepo{tenants: make(map[int]*models.Tenant)}
}

func (m *MockTurnoverTenantRepo) addTenant(t *models.Tenant) { m.tenants[t.ID] = t }

func (m *MockTurnoverTenantRepo) GetByID(id int) (*models.Tenant, error) {
	t, ok := m.tenants[id]
	if !ok {
		return nil, errors.New("tenant not found")
	}
	return t, nil
}

func (m *MockTurnoverTenantRepo) GetByIDIncludingInactive(id int) (*models.Tenant, error) {
	return m.GetByID(id)
}

func (m *MockTurnoverTenantRepo) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	return nil, nil
}
func (m *MockTurnoverTenantRepo) CheckEmailExists(email string, excludeID int) (bool, error) {
	return false, nil
}
func (m *MockTurnoverTenantRepo) CheckNIDExists(nid string, excludeID int) (bool, error) {
	return false, nil
}
func (m *MockTurnoverTenantRepo) GetDecryptedNID(id int) (string, int, error) { return "", 0, nil }
func (m *MockTurnoverTenantRepo) GetByUnitID(unitID int) (*models.Tenant, error) {
	return nil, nil
}
func (m *MockTurnoverTenantRepo) GetAll(page, pageSize, orgID int) ([]*models.Tenant, int, error) {
	return nil, 0, nil
}
func (m *MockTurnoverTenantRepo) Update(id int, updates map[string]any) error { return nil }

// ---------------------------------------------------------------------------
// MockTurnoverLeaseChargeRepo - implements interfaces.LeaseChargeRepositoryInterface
// ---------------------------------------------------------------------------

type MockTurnoverLeaseChargeRepo struct{}

func newMockTurnoverLeaseChargeRepo() *MockTurnoverLeaseChargeRepo {
	return &MockTurnoverLeaseChargeRepo{}
}

func (m *MockTurnoverLeaseChargeRepo) Create(leaseID int, req *models.CreateLeaseChargeRequest) (*models.LeaseCharge, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseChargeRepo) GetByID(id int) (*models.LeaseCharge, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseChargeRepo) GetByLeaseID(leaseID int) ([]*models.LeaseCharge, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseChargeRepo) Update(id int, req *models.UpdateLeaseChargeRequest) (*models.LeaseCharge, error) {
	return nil, nil
}
func (m *MockTurnoverLeaseChargeRepo) Delete(id int) error { return nil }
func (m *MockTurnoverLeaseChargeRepo) SumActiveChargesByLeaseID(leaseID int) (float64, error) {
	return 0, nil
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const (
	turnoverOrgID     = 7
	turnoverOtherOrg  = 8
	turnoverLeaseID   = 100
	turnoverOldTenant = 11
	turnoverNewTenant = 22
	turnoverUnitID    = 55
	turnoverUserID    = 3
)

// newTurnoverFixture wires a LeaseService holding one active lease on unit 55
// for tenant 11, plus an available tenant 22 in the same organization.
func newTurnoverFixture(t *testing.T) (*LeaseService, *MockTurnoverLeaseRepo, *MockTurnoverTenantRepo, *MockPaymentAuditService) {
	t.Helper()

	leaseRepo := newMockTurnoverLeaseRepo()
	leaseRepo.addLease(&models.Lease{
		ID:              turnoverLeaseID,
		UnitID:          turnoverUnitID,
		TenantID:        turnoverOldTenant,
		LeaseType:       models.LeaseTypeResidential,
		StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		DurationMonths:  12,
		MonthlyRent:     25000,
		SecurityDeposit: 50000,
		Active:          true,
		OrganizationID:  turnoverOrgID,
	})

	tenantRepo := newMockTurnoverTenantRepo()
	tenantRepo.addTenant(&models.Tenant{ID: turnoverOldTenant, Name: "Outgoing", Active: true, OrganizationID: turnoverOrgID})
	tenantRepo.addTenant(&models.Tenant{ID: turnoverNewTenant, Name: "Incoming", Active: true, OrganizationID: turnoverOrgID})

	audit := newMockPaymentAuditService()
	svc := NewLeaseService(leaseRepo, tenantRepo, newMockPaymentUnitRepo(), newMockTurnoverLeaseChargeRepo(), audit)
	return svc, leaseRepo, tenantRepo, audit
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestReplaceTenant_CarriesOverTermsFromOutgoingLease(t *testing.T) {
	svc, leaseRepo, _, audit := newTurnoverFixture(t)

	result, err := svc.ReplaceTenant(turnoverLeaseID, &models.ReplaceTenantRequest{
		NewTenantID:  turnoverNewTenant,
		HandoverDate: "2026-07-01",
	}, turnoverUserID, turnoverOrgID)
	if err != nil {
		t.Fatalf("expected turnover to succeed, got %v", err)
	}

	successor := leaseRepo.lastSuccessor
	if successor == nil {
		t.Fatal("expected the repository to receive a successor lease")
	}
	if successor.UnitID != turnoverUnitID {
		t.Errorf("successor unit = %d, want %d", successor.UnitID, turnoverUnitID)
	}
	if successor.TenantID != turnoverNewTenant {
		t.Errorf("successor tenant = %d, want %d", successor.TenantID, turnoverNewTenant)
	}
	if successor.MonthlyRent != 25000 {
		t.Errorf("successor rent = %v, want the outgoing lease rent 25000", successor.MonthlyRent)
	}
	if successor.SecurityDeposit != 50000 {
		t.Errorf("successor deposit = %v, want the outgoing lease deposit 50000", successor.SecurityDeposit)
	}
	if successor.DurationMonths != 12 {
		t.Errorf("successor duration = %d, want the outgoing lease duration 12", successor.DurationMonths)
	}
	if successor.LeaseType != models.LeaseTypeResidential {
		t.Errorf("successor lease type = %q, want Residential from the outgoing lease", successor.LeaseType)
	}
	if successor.StartDate != "2026-07-01" {
		t.Errorf("successor start = %q, want the handover date", successor.StartDate)
	}
	if successor.OrganizationID != turnoverOrgID {
		t.Errorf("successor org = %d, want %d", successor.OrganizationID, turnoverOrgID)
	}

	// The outgoing lease must be reported as closed, not still active.
	if result.PreviousLease.Active {
		t.Error("previous lease in the response is still marked active")
	}
	wantEnd := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !result.PreviousLease.EndDate.Equal(wantEnd) {
		t.Errorf("previous lease end date = %v, want the handover date %v", result.PreviousLease.EndDate, wantEnd)
	}
	if !result.NewLease.Active {
		t.Error("successor lease in the response is not active")
	}
	if result.NewLease.TenantID != turnoverNewTenant {
		t.Errorf("successor lease tenant = %d, want %d", result.NewLease.TenantID, turnoverNewTenant)
	}

	// One audit entry for the closure, one for the successor.
	if audit.userActionCalls != 2 {
		t.Errorf("audit user actions = %d, want 2", audit.userActionCalls)
	}
}

func TestReplaceTenant_AppliesProvidedTermOverrides(t *testing.T) {
	svc, leaseRepo, _, _ := newTurnoverFixture(t)

	newType := models.LeaseTypeCommercial
	newDuration := 24
	newRent := 32000.0
	newDeposit := 64000.0

	if _, err := svc.ReplaceTenant(turnoverLeaseID, &models.ReplaceTenantRequest{
		NewTenantID:     turnoverNewTenant,
		HandoverDate:    "2026-07-01",
		LeaseType:       &newType,
		DurationMonths:  &newDuration,
		MonthlyRent:     &newRent,
		SecurityDeposit: &newDeposit,
	}, turnoverUserID, turnoverOrgID); err != nil {
		t.Fatalf("expected turnover to succeed, got %v", err)
	}

	successor := leaseRepo.lastSuccessor
	if successor.LeaseType != newType {
		t.Errorf("lease type = %q, want %q", successor.LeaseType, newType)
	}
	if successor.DurationMonths != newDuration {
		t.Errorf("duration = %d, want %d", successor.DurationMonths, newDuration)
	}
	if successor.MonthlyRent != newRent {
		t.Errorf("rent = %v, want %v", successor.MonthlyRent, newRent)
	}
	if successor.SecurityDeposit != newDeposit {
		t.Errorf("deposit = %v, want %v", successor.SecurityDeposit, newDeposit)
	}
}

func TestReplaceTenant_Rejections(t *testing.T) {
	const alreadyClosedCase = "outgoing lease already closed"

	tests := []struct {
		name    string
		setup   func(*MockTurnoverLeaseRepo, *MockTurnoverTenantRepo)
		leaseID int
		orgID   int
		req     *models.ReplaceTenantRequest
		wantErr string
	}{
		{
			name:    "lease belonging to another organization",
			leaseID: turnoverLeaseID,
			orgID:   turnoverOtherOrg,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "lease not found",
		},
		{
			name: alreadyClosedCase,
			setup: func(l *MockTurnoverLeaseRepo, _ *MockTurnoverTenantRepo) {
				l.leases[turnoverLeaseID].Active = false
			},
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "lease is not active",
		},
		{
			name:    "handover before the outgoing lease started",
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2025-12-31"},
			wantErr: "handover date cannot be before the current lease start date",
		},
		{
			name:    "unparseable handover date",
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "01-07-2026"},
			wantErr: "invalid handover date format",
		},
		{
			name:    "incoming tenant is the current tenant",
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverOldTenant, HandoverDate: "2026-07-01"},
			wantErr: "incoming tenant is already the tenant on this lease",
		},
		{
			name: "incoming tenant from another organization",
			setup: func(_ *MockTurnoverLeaseRepo, tr *MockTurnoverTenantRepo) {
				tr.tenants[turnoverNewTenant].OrganizationID = turnoverOtherOrg
			},
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "tenant not found",
		},
		{
			name: "incoming tenant is inactive",
			setup: func(_ *MockTurnoverLeaseRepo, tr *MockTurnoverTenantRepo) {
				tr.tenants[turnoverNewTenant].Active = false
			},
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "cannot assign an inactive tenant",
		},
		{
			name: "incoming tenant already holds an active lease",
			setup: func(l *MockTurnoverLeaseRepo, _ *MockTurnoverTenantRepo) {
				l.tenantsWithActiveLease[turnoverNewTenant] = true
			},
			leaseID: turnoverLeaseID,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "incoming tenant already has an active lease",
		},
		{
			name:    "unknown lease",
			leaseID: 4242,
			orgID:   turnoverOrgID,
			req:     &models.ReplaceTenantRequest{NewTenantID: turnoverNewTenant, HandoverDate: "2026-07-01"},
			wantErr: "failed to get lease",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, leaseRepo, tenantRepo, _ := newTurnoverFixture(t)
			if tc.setup != nil {
				tc.setup(leaseRepo, tenantRepo)
			}

			_, err := svc.ReplaceTenant(tc.leaseID, tc.req, turnoverUserID, tc.orgID)
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}

			// A rejected turnover must never reach the database.
			if leaseRepo.replaceCalls != 0 {
				t.Errorf("repository ReplaceTenant called %d times for a rejected turnover, want 0", leaseRepo.replaceCalls)
			}
			if outgoing := leaseRepo.leases[turnoverLeaseID]; outgoing != nil && tc.name != alreadyClosedCase && !outgoing.Active {
				t.Error("outgoing lease was closed even though the turnover was rejected")
			}
		})
	}
}

// TestReplaceTenant_PropagatesRepositoryFailure covers the lost race: a
// concurrent turnover committed first, so the guarded UPDATE matched no rows.
func TestReplaceTenant_PropagatesRepositoryFailure(t *testing.T) {
	svc, leaseRepo, _, audit := newTurnoverFixture(t)
	leaseRepo.replaceFailsWith = errors.New("lease is no longer active")

	_, err := svc.ReplaceTenant(turnoverLeaseID, &models.ReplaceTenantRequest{
		NewTenantID:  turnoverNewTenant,
		HandoverDate: "2026-07-01",
	}, turnoverUserID, turnoverOrgID)
	if err == nil {
		t.Fatal("expected the repository failure to surface")
	}
	if !strings.Contains(err.Error(), "failed to replace tenant") {
		t.Errorf("error = %q, want it to wrap the repository failure", err.Error())
	}
	if audit.userActionCalls != 0 {
		t.Errorf("audit user actions = %d, want 0 when the turnover failed", audit.userActionCalls)
	}
}
