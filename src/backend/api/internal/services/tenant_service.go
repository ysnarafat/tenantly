package services

import (
	"fmt"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type TenantService struct {
	tenantRepo   interfaces.TenantRepositoryInterface
	leaseRepo    interfaces.LeaseRepositoryInterface
	auditService interfaces.AuditServiceInterface
}

func NewTenantService(
	tenantRepo interfaces.TenantRepositoryInterface,
	leaseRepo interfaces.LeaseRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *TenantService {
	return &TenantService{
		tenantRepo:   tenantRepo,
		leaseRepo:    leaseRepo,
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

	// Log audit — record only the non-sensitive NID last-four, never the full value.
	s.auditService.LogUserAction(userID, "CREATE", "tenants", &tenant.ID, nil, map[string]interface{}{
		"tenant_id":     tenant.ID,
		"name":          tenant.Name,
		"tenant_type":   tenant.TenantType,
		"email":         tenant.Email,
		"nid_last_four": tenant.NIDLastFour,
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

// GetTenantByID retrieves a tenant by ID with lease information
func (s *TenantService) GetTenantByID(id int, orgID int) (*models.TenantWithLeases, error) {
	tenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Verify organization ownership
	if tenant.OrganizationID != orgID {
		return nil, fmt.Errorf("tenant not found")
	}

	// Get active lease count
	hasActiveLease, err := s.leaseRepo.HasActiveLeaseForTenant(id)
	if err != nil {
		return nil, fmt.Errorf("failed to check active leases: %w", err)
	}

	activeLeases := 0
	if hasActiveLease {
		activeLeases = 1
	}

	return &models.TenantWithLeases{
		Tenant:       *tenant,
		ActiveLeases: activeLeases,
		TotalUnits:   activeLeases, // Each active lease corresponds to a unit
	}, nil
}

// UpdateTenant updates a tenant
func (s *TenantService) UpdateTenant(id int, req *models.UpdateTenantRequest, userID, orgID int) (*models.TenantResponse, error) {
	// Get existing tenant
	existingTenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Verify organization ownership
	if existingTenant.OrganizationID != orgID {
		return nil, fmt.Errorf("tenant not found")
	}

	// Validate email uniqueness if being updated
	if req.Email != nil && *req.Email != existingTenant.Email {
		exists, err := s.tenantRepo.CheckEmailExists(*req.Email, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email already exists")
		}
	}

	// Validate NID uniqueness if being updated. The existing plaintext is no
	// longer read back (NID is stored encrypted), so we always run the hashed
	// uniqueness check excluding this tenant's own record.
	if req.NIDNumber != nil {
		exists, err := s.tenantRepo.CheckNIDExists(*req.NIDNumber, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check NID uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("NID number already exists")
		}
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.TenantType != nil {
		updates["tenant_type"] = *req.TenantType
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.NIDNumber != nil {
		updates["nid_number"] = *req.NIDNumber
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	// Update tenant
	if err := s.tenantRepo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Get updated tenant
	updatedTenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated tenant: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "update", "tenants", &id, existingTenant, updatedTenant)
	}

	return updatedTenant.ToResponse(), nil
}

// DeleteTenant soft deletes a tenant (guards against active leases)
func (s *TenantService) DeleteTenant(id int, userID, orgID int) error {
	// Get existing tenant
	existingTenant, err := s.tenantRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Verify organization ownership
	if existingTenant.OrganizationID != orgID {
		return fmt.Errorf("tenant not found")
	}

	// Check for active leases
	hasActiveLease, err := s.leaseRepo.HasActiveLeaseForTenant(id)
	if err != nil {
		return fmt.Errorf("failed to check active leases: %w", err)
	}
	if hasActiveLease {
		return fmt.Errorf("cannot delete tenant with active leases")
	}

	// Soft delete tenant
	updates := map[string]interface{}{"active": false}
	if err := s.tenantRepo.Update(id, updates); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "delete", "tenants", &id, existingTenant, nil)
	}

	return nil
}

// RevealNID returns the full decrypted NID for a tenant after verifying the
// caller's organization owns it, and records the reveal in the audit log (with
// only the last four digits, never the full value).
func (s *TenantService) RevealNID(id, orgID, userID int) (string, error) {
	nid, tenantOrgID, err := s.tenantRepo.GetDecryptedNID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant NID: %w", err)
	}

	// Verify organization ownership
	if tenantOrgID != orgID {
		return "", fmt.Errorf("tenant not found")
	}

	// Audit the reveal — sensitive access to full PII must be traceable.
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "REVEAL_NID", "tenants", &id, nil, map[string]interface{}{
			"tenant_id":     id,
			"nid_last_four": nid[max(0, len(nid)-4):],
		})
	}

	return nid, nil
}
