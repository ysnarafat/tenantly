package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ysnarafat/tenantly/internal/interfaces/mocks"
	"github.com/ysnarafat/tenantly/internal/models"
)

func TestTenantService_CreateTenant(t *testing.T) {
	mockRepo := new(mocks.TenantRepositoryInterface)
	mockAudit := new(mocks.AuditServiceInterface)
	service := NewTenantService(mockRepo, mockAudit)

	req := &models.CreateTenantRequest{
		Name:        "John Doe",
		Email:       "john@example.com",
		NIDNumber:   "NID123",
		TenantType:  models.TenantTypeIndividual,
		PhoneNumber: "1234567890",
		Address:     "123 Main St",
	}

	userID := 1

	t.Run("success", func(t *testing.T) {
		createdTenant := &models.Tenant{
			ID:          1,
			Name:        req.Name,
			Email:       req.Email,
			NIDNumber:   req.NIDNumber,
			TenantType:  req.TenantType,
			PhoneNumber: req.PhoneNumber,
			Address:     req.Address,
			Active:      true,
		}

		mockRepo.On("CheckEmailExists", req.Email, 0).Return(false, nil).Once()
		mockRepo.On("CheckNIDExists", req.NIDNumber, 0).Return(false, nil).Once()
		mockRepo.On("Create", req).Return(createdTenant, nil).Once()
		mockAudit.On("LogUserAction", userID, "CREATE", "tenants", mock.AnythingOfType("*int"), mock.Anything, mock.Anything).Return(nil).Once()

		resp, err := service.CreateTenant(req, userID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, createdTenant.ID, resp.ID)
		assert.Equal(t, createdTenant.Name, resp.Name)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockRepo.On("CheckEmailExists", req.Email, 0).Return(true, nil).Once()

		resp, err := service.CreateTenant(req, userID)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("nid already exists", func(t *testing.T) {
		mockRepo.On("CheckEmailExists", req.Email, 0).Return(false, nil).Once()
		mockRepo.On("CheckNIDExists", req.NIDNumber, 0).Return(true, nil).Once()

		resp, err := service.CreateTenant(req, userID)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "NID number already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("CheckEmailExists", req.Email, 0).Return(false, nil).Once()
		mockRepo.On("CheckNIDExists", req.NIDNumber, 0).Return(false, nil).Once()
		mockRepo.On("Create", req).Return((*models.Tenant)(nil), fmt.Errorf("db error")).Once()

		resp, err := service.CreateTenant(req, userID)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create tenant")
		mockRepo.AssertExpectations(t)
	})
}
