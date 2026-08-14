package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// LeaseService implements the LeaseServiceInterface
type LeaseService struct {
	leaseRepo    interfaces.LeaseRepositoryInterface
	tenantRepo   interfaces.TenantRepositoryInterface
	unitRepo     interfaces.UnitRepositoryInterface
	auditService interfaces.AuditServiceInterface
}

// NewLeaseService creates a new lease service
func NewLeaseService(
	leaseRepo interfaces.LeaseRepositoryInterface,
	tenantRepo interfaces.TenantRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *LeaseService {
	return &LeaseService{
		leaseRepo:    leaseRepo,
		tenantRepo:   tenantRepo,
		unitRepo:     unitRepo,
		auditService: auditService,
	}
}

// CreateLease creates a new lease with validation
func (s *LeaseService) CreateLease(req *models.CreateLeaseRequest, userID int) (*models.LeaseWithDetails, error) {
	// Validate the unit and tenant belong to the caller's organization — prevents
	// creating a lease that references another organization's unit/tenant (IDOR).
	unit, err := s.unitRepo.GetByID(req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}
	if unit.OrganizationID != req.OrganizationID {
		return nil, fmt.Errorf("unit not found")
	}

	tenant, err := s.tenantRepo.GetByID(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	if tenant.OrganizationID != req.OrganizationID {
		return nil, fmt.Errorf("tenant not found")
	}

	// Validate unit doesn't have active lease
	hasActiveLease, err := s.leaseRepo.HasActiveLeaseOnUnit(req.UnitID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check unit availability: %w", err)
	}
	if hasActiveLease {
		return nil, fmt.Errorf("unit already has an active lease")
	}

	// Create the lease
	lease, err := s.leaseRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease: %w", err)
	}

	// Get lease with details
	leaseDetails, err := s.leaseRepo.GetByIDWithDetails(lease.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease details: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "create", "leases", &lease.ID, nil, leaseDetails)
	}

	return leaseDetails, nil
}

// GetLeaseByID retrieves a lease by ID
func (s *LeaseService) GetLeaseByID(id int, orgID int) (*models.LeaseWithDetails, error) {
	lease, err := s.leaseRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease: %w", err)
	}

	// Verify organization ownership
	if lease.OrganizationID != orgID {
		return nil, fmt.Errorf("lease not found")
	}

	return lease, nil
}

// GetAllLeases retrieves all leases with pagination
func (s *LeaseService) GetAllLeases(page, pageSize, orgID int) (*models.LeaseListResponse, error) {
	leases, total, err := s.leaseRepo.GetAll(page, pageSize, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get leases: %w", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	response := &models.LeaseListResponse{
		Leases: leases,
		Pagination: &models.PaginationInfo{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
	}

	return response, nil
}

// UpdateLease updates a lease (financials and dates only, not tenant/unit)
func (s *LeaseService) UpdateLease(id int, req *models.UpdateLeaseRequest, userID, orgID int) (*models.LeaseWithDetails, error) {
	// Get existing lease
	existingLease, err := s.leaseRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease: %w", err)
	}

	// Verify organization ownership
	if existingLease.OrganizationID != orgID {
		return nil, fmt.Errorf("lease not found")
	}

	// Store old values for audit
	oldLease := *existingLease

	// Update lease
	updatedLease, err := s.leaseRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update lease: %w", err)
	}

	// Get lease with details
	leaseDetails, err := s.leaseRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease details: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "update", "leases", &id, oldLease, updatedLease)
	}

	return leaseDetails, nil
}

// DeleteLease hard deletes a lease (only for draft/future leases)
func (s *LeaseService) DeleteLease(id int, userID, orgID int) error {
	// Get existing lease
	existingLease, err := s.leaseRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get lease: %w", err)
	}

	// Verify organization ownership
	if existingLease.OrganizationID != orgID {
		return fmt.Errorf("lease not found")
	}

	// Only allow deletion for future leases
	now := time.Now()
	if existingLease.StartDate.Before(now) && existingLease.Active {
		return fmt.Errorf("cannot delete active or past leases, use terminate instead")
	}

	// Delete lease
	if err := s.leaseRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete lease: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		_ = s.auditService.LogUserAction(userID, "delete", "leases", &id, existingLease, nil)
	}

	return nil
}

// TerminateLease soft deletes a lease (sets active to false with termination date)
func (s *LeaseService) TerminateLease(id int, userID, orgID int, terminationDateStr string) error {
	// Get existing lease
	existingLease, err := s.leaseRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get lease: %w", err)
	}

	// Verify organization ownership
	if existingLease.OrganizationID != orgID {
		return fmt.Errorf("lease not found")
	}

	// Only active leases can be terminated
	if !existingLease.Active {
		return fmt.Errorf("lease is not active")
	}

	// Parse termination date if provided
	var terminationDate *time.Time
	if terminationDateStr != "" {
		t, err := time.Parse("2006-01-02", terminationDateStr)
		if err != nil {
			return fmt.Errorf("invalid termination date format: %w", err)
		}
		terminationDate = &t
	} else {
		now := time.Now()
		terminationDate = &now
	}

	// Termination date cannot be before lease start date
	if terminationDate.Before(existingLease.StartDate) {
		return fmt.Errorf("termination date cannot be before lease start date")
	}

	// Soft delete lease
	if err := s.leaseRepo.SoftDelete(id); err != nil {
		return fmt.Errorf("failed to terminate lease: %w", err)
	}

	// Log audit
	if s.auditService != nil {
		oldLease := *existingLease
		newLease := *existingLease
		newLease.Active = false
		newLease.EndDate = *terminationDate
		_ = s.auditService.LogUserAction(userID, "terminate", "leases", &id, oldLease, newLease)
	}

	return nil
}

// GetLeasesByUnit retrieves leases for a specific unit
func (s *LeaseService) GetLeasesByUnit(unitID int, page, pageSize, orgID int) (*models.LeaseListResponse, error) {
	leases, total, err := s.leaseRepo.GetByUnitID(unitID, page, pageSize, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit leases: %w", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	response := &models.LeaseListResponse{
		Leases: leases,
		Pagination: &models.PaginationInfo{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
	}

	return response, nil
}

// GetLeasesByTenant retrieves leases for a specific tenant
func (s *LeaseService) GetLeasesByTenant(tenantID int, page, pageSize, orgID int) (*models.LeaseListResponse, error) {
	leases, total, err := s.leaseRepo.GetByTenantID(tenantID, page, pageSize, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant leases: %w", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	response := &models.LeaseListResponse{
		Leases: leases,
		Pagination: &models.PaginationInfo{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
	}

	return response, nil
}

// GetLeasesDue retrieves all leases with unpaid rent for the current month
func (s *LeaseService) GetLeasesDue(orgID int) ([]models.LeaseDue, error) {
	leasesDue, err := s.leaseRepo.GetLeasesDueForMonth(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get leases due: %w", err)
	}
	return leasesDue, nil
}

// GetDueSummary retrieves summary statistics for unpaid rent
func (s *LeaseService) GetDueSummary(orgID int) (*models.DueSummary, error) {
	summary, err := s.leaseRepo.GetDueSummary(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get due summary: %w", err)
	}
	return summary, nil
}
