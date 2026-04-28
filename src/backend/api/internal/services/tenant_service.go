package services

import (
	"fmt"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type TenantService struct {
	tenantRepo   interfaces.TenantRepositoryInterface
	auditService interfaces.AuditServiceInterface
}

func NewTenantService(
	tenantRepo interfaces.TenantRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *TenantService {
	return &TenantService{
		tenantRepo:   tenantRepo,
		auditService: auditService,
	}
}

// CreateTenant creates a new tenant with validation
func (s *TenantService) CreateTenant(req *models.CreateTenantRequest, userID int) (*models.TenantResponse, error) {
	// Validate email uniqueness if provided
	if req.Email != "" {
		exists, err := s.tenantRepo.CheckEmailExists(req.Email, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Validate NID uniqueness if provided
	exists, err := s.tenantRepo.CheckNIDExists(req.NIDNumber, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check NID uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("NID number already exists")
	}

	// Create tenant
	tenant, err := s.tenantRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "CREATE", "tenants", &tenant.ID, nil, map[string]interface{}{
		"tenant_id":   tenant.ID,
		"name":        tenant.Name,
		"tenant_type": tenant.TenantType,
		"email":       tenant.Email,
		"nid_number":  tenant.NIDNumber,
	})

	return tenant.ToResponse(), nil
}

// GetAllTenants retrieves all active tenants with pagination
func (s *TenantService) GetAllTenants(page, pageSize, orgID int) (*models.TenantListResponse, error) {
	// Defaults if 0
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	tenants, totalCount, err := s.tenantRepo.GetAll(page, pageSize, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}

	var responses []*models.TenantResponse
	for _, tenant := range tenants {
		responses = append(responses, tenant.ToResponse())
	}

	// Calculate pagination info
	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := &models.PaginationInfo{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}

	return &models.TenantListResponse{
		Tenants:    responses,
		Pagination: pagination,
	}, nil
}
