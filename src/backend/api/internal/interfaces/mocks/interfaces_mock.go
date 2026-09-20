package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/models"
)

// TenantRepositoryInterface is a mock of TenantRepositoryInterface
type TenantRepositoryInterface struct {
	mock.Mock
}

func (m *TenantRepositoryInterface) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *TenantRepositoryInterface) CheckEmailExists(email string, excludeID int) (bool, error) {
	args := m.Called(email, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *TenantRepositoryInterface) CheckNIDExists(nid string, excludeID int) (bool, error) {
	args := m.Called(nid, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *TenantRepositoryInterface) GetByID(id int) (*models.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *TenantRepositoryInterface) GetByIDIncludingInactive(id int) (*models.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *TenantRepositoryInterface) GetDecryptedNID(id int) (string, int, error) {
	args := m.Called(id)
	return args.String(0), args.Int(1), args.Error(2)
}

func (m *TenantRepositoryInterface) GetByUnitID(unitID int) (*models.Tenant, error) {
	args := m.Called(unitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tenant), args.Error(1)
}

func (m *TenantRepositoryInterface) GetAll(page, pageSize, orgID int) ([]*models.Tenant, int, error) {
	args := m.Called(page, pageSize, orgID)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Tenant), args.Int(1), args.Error(2)
}

func (m *TenantRepositoryInterface) Update(id int, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

// AuditServiceInterface is a mock of AuditServiceInterface
type AuditServiceInterface struct {
	mock.Mock
}

func (m *AuditServiceInterface) LogSystemAction(action, entity string, entityID *int, oldData, newData interface{}) error {
	args := m.Called(action, entity, entityID, oldData, newData)
	return args.Error(0)
}

func (m *AuditServiceInterface) LogUserAction(userID int, action, entity string, entityID *int, oldData, newData interface{}) error {
	args := m.Called(userID, action, entity, entityID, oldData, newData)
	return args.Error(0)
}
