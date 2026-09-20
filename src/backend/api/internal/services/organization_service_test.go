package services

import (
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// MockOrganizationRepository is a mock implementation for testing
type MockOrganizationRepository struct {
	organizations map[int]*models.Organization
	nextID        int
}

func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		organizations: make(map[int]*models.Organization),
		nextID:        1,
	}
}

func (m *MockOrganizationRepository) Create(org *models.Organization) error {
	org.ID = m.nextID
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()
	m.organizations[org.ID] = org
	m.nextID++
	return nil
}

func (m *MockOrganizationRepository) GetByID(id int) (*models.Organization, error) {
	org, exists := m.organizations[id]
	if !exists {
		return nil, nil
	}
	return org, nil
}

func (m *MockOrganizationRepository) GetBySlug(slug string) (*models.Organization, error) {
	for _, org := range m.organizations {
		if org.Slug == slug {
			return org, nil
		}
	}
	return nil, nil
}

func (m *MockOrganizationRepository) GetAll(activeOnly bool) ([]*models.Organization, error) {
	var orgs []*models.Organization
	for _, org := range m.organizations {
		if !activeOnly || org.Active {
			orgs = append(orgs, org)
		}
	}
	return orgs, nil
}

func (m *MockOrganizationRepository) Update(id int, updates map[string]interface{}) error {
	org, exists := m.organizations[id]
	if !exists {
		return nil
	}

	if name, ok := updates["name"]; ok {
		org.Name = name.(string)
	}
	if tier, ok := updates["subscription_tier"]; ok {
		org.SubscriptionTier = tier.(models.SubscriptionTier)
	}
	if max, ok := updates["max_users"]; ok {
		org.MaxUsers = max.(int)
	}
	if active, ok := updates["active"]; ok {
		org.Active = active.(bool)
	}

	org.UpdatedAt = time.Now()
	return nil
}

func (m *MockOrganizationRepository) Delete(id int) error {
	if org, exists := m.organizations[id]; exists {
		org.Active = false
		org.UpdatedAt = time.Now()
	}
	return nil
}

// MockUserInvitationRepository is a mock implementation for testing
type MockUserInvitationRepository struct {
	invitations map[int]*models.UserInvitation
	nextID      int
}

func NewMockUserInvitationRepository() *MockUserInvitationRepository {
	return &MockUserInvitationRepository{
		invitations: make(map[int]*models.UserInvitation),
		nextID:      1,
	}
}

func (m *MockUserInvitationRepository) Create(invitation *models.UserInvitation) error {
	invitation.ID = m.nextID
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	m.invitations[invitation.ID] = invitation
	m.nextID++
	return nil
}

func (m *MockUserInvitationRepository) GetByID(id int) (*models.UserInvitation, error) {
	inv, exists := m.invitations[id]
	if !exists {
		return nil, nil
	}
	return inv, nil
}

func (m *MockUserInvitationRepository) GetByToken(token string) (*models.UserInvitation, error) {
	for _, inv := range m.invitations {
		if inv.InvitationToken == token {
			return inv, nil
		}
	}
	return nil, nil
}

func (m *MockUserInvitationRepository) GetByEmailAndOrg(email string, orgID int) (*models.UserInvitation, error) {
	for _, inv := range m.invitations {
		if inv.Email == email && inv.OrganizationID == orgID {
			return inv, nil
		}
	}
	return nil, nil
}

func (m *MockUserInvitationRepository) GetPendingByOrganization(orgID int) ([]*models.UserInvitation, error) {
	var invitations []*models.UserInvitation
	for _, inv := range m.invitations {
		if inv.OrganizationID == orgID && inv.AcceptedAt == nil && inv.ExpiresAt.After(time.Now()) {
			invitations = append(invitations, inv)
		}
	}
	return invitations, nil
}

func (m *MockUserInvitationRepository) GetByOrganization(orgID int) ([]*models.UserInvitation, error) {
	var invitations []*models.UserInvitation
	for _, inv := range m.invitations {
		if inv.OrganizationID == orgID {
			invitations = append(invitations, inv)
		}
	}
	return invitations, nil
}

func (m *MockUserInvitationRepository) AcceptInvitation(invitationID int, userID int) error {
	inv, exists := m.invitations[invitationID]
	if !exists {
		return nil
	}
	now := time.Now()
	inv.AcceptedAt = &now
	inv.AcceptedByUserID = &userID
	return nil
}

func (m *MockUserInvitationRepository) Delete(id int) error {
	delete(m.invitations, id)
	return nil
}

func (m *MockUserInvitationRepository) CleanupExpiredInvitations() error {
	for id, inv := range m.invitations {
		if inv.AcceptedAt == nil && inv.ExpiresAt.Before(time.Now()) {
			delete(m.invitations, id)
		}
	}
	return nil
}

// MockOrgAuditService is a mock implementation for testing
type MockOrgAuditService struct {
	logs []map[string]interface{}
}

func NewMockAuditForOrg() *MockOrgAuditService {
	return &MockOrgAuditService{
		logs: make([]map[string]interface{}, 0),
	}
}

func (m *MockOrgAuditService) LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	m.logs = append(m.logs, map[string]interface{}{
		"user_id": userID,
		"action":  action,
		"table":   tableName,
		"record":  recordID,
		"old":     oldValues,
		"new":     newValues,
	})
	return nil
}

func (m *MockOrgAuditService) LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	m.logs = append(m.logs, map[string]interface{}{
		"action": action,
		"table":  tableName,
		"record": recordID,
		"old":    oldValues,
		"new":    newValues,
	})
	return nil
}

// Test cases

func TestCreateOrganization(t *testing.T) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockInvRepo := NewMockUserInvitationRepository()
	mockAudit := NewMockAuditForOrg()

	service := NewOrganizationService(mockOrgRepo, mockInvRepo, mockAudit)

	req := &models.CreateOrganizationRequest{
		Name:             "Test Org",
		Slug:             "test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         50,
	}

	org, err := service.CreateOrganization(req, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if org.Name != "Test Org" || org.Slug != "test-org" {
		t.Fatalf("organization data mismatch")
	}

	if org.ID != 1 {
		t.Fatalf("expected ID 1, got %d", org.ID)
	}
}

func TestGetOrganization(t *testing.T) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockInvRepo := NewMockUserInvitationRepository()
	mockAudit := NewMockAuditForOrg()

	// Create an organization first
	org := &models.Organization{
		Name:             "Test Org",
		Slug:             "test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         50,
		Active:           true,
	}
	_ = mockOrgRepo.Create(org)

	service := NewOrganizationService(mockOrgRepo, mockInvRepo, mockAudit)

	retrieved, err := service.GetOrganization(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Name != "Test Org" {
		t.Fatalf("expected name 'Test Org', got %s", retrieved.Name)
	}
}

func TestListOrganizations(t *testing.T) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockInvRepo := NewMockUserInvitationRepository()
	mockAudit := NewMockAuditForOrg()

	// Create test organizations
	org1 := &models.Organization{
		Name:             "Org 1",
		Slug:             "org-1",
		SubscriptionTier: models.TierBasic,
		Active:           true,
	}
	org2 := &models.Organization{
		Name:             "Org 2",
		Slug:             "org-2",
		SubscriptionTier: models.TierProfessional,
		Active:           false,
	}
	_ = mockOrgRepo.Create(org1)
	_ = mockOrgRepo.Create(org2)

	service := NewOrganizationService(mockOrgRepo, mockInvRepo, mockAudit)

	orgs, err := service.ListOrganizations(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgs) != 2 {
		t.Fatalf("expected 2 organizations, got %d", len(orgs))
	}

	orgsActive, err := service.ListOrganizations(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orgsActive) != 1 {
		t.Fatalf("expected 1 active organization, got %d", len(orgsActive))
	}
}

func TestUpdateOrganization(t *testing.T) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockInvRepo := NewMockUserInvitationRepository()
	mockAudit := NewMockAuditForOrg()

	org := &models.Organization{
		Name:             "Test Org",
		Slug:             "test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         50,
		Active:           true,
	}
	_ = mockOrgRepo.Create(org)

	service := NewOrganizationService(mockOrgRepo, mockInvRepo, mockAudit)

	updateReq := &models.UpdateOrganizationRequest{
		Name:             "Updated Org",
		SubscriptionTier: models.TierEnterprise,
	}

	err := service.UpdateOrganization(1, updateReq, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetOrganization(1)
	if updated.Name != "Updated Org" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
}

func TestInviteUserToOrganization(t *testing.T) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockInvRepo := NewMockUserInvitationRepository()
	mockAudit := NewMockAuditForOrg()

	org := &models.Organization{
		Name:             "Test Org",
		Slug:             "test-org",
		SubscriptionTier: models.TierBasic,
		MaxUsers:         50,
		Active:           true,
	}
	_ = mockOrgRepo.Create(org)

	service := NewOrganizationService(mockOrgRepo, mockInvRepo, mockAudit)

	inviteReq := &models.InviteUserRequest{
		Email: "user@example.com",
		Role:  "Admin",
	}

	invitation, err := service.InviteUserToOrganization(1, inviteReq, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invitation.Email != "user@example.com" || invitation.Role != "Admin" {
		t.Fatalf("invitation data mismatch")
	}

	if invitation.InvitationToken == "" {
		t.Fatalf("invitation token should not be empty")
	}
}
