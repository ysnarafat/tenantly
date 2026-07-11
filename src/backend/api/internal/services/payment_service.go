package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type PaymentService struct {
	paymentRepo  interfaces.PaymentRepositoryInterface
	unitRepo     interfaces.UnitRepositoryInterface
	buildingRepo interfaces.BuildingRepositoryInterface
	propertyRepo interfaces.PropertyRepositoryInterface
	auditService interfaces.AuditServiceInterface
	userRepo     interfaces.UserRepositoryInterface
}

func NewPaymentService(
	paymentRepo interfaces.PaymentRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo interfaces.PropertyRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
	userRepo interfaces.UserRepositoryInterface,
) *PaymentService {
	return &PaymentService{
		paymentRepo:  paymentRepo,
		unitRepo:     unitRepo,
		buildingRepo: buildingRepo,
		propertyRepo: propertyRepo,
		auditService: auditService,
		userRepo:     userRepo,
	}
}

// CreatePayment creates a new payment record with building information
func (s *PaymentService) CreatePayment(req *models.CreatePaymentRequest, userID int) (*models.Payment, error) {
	// Validate unit exists and get building context
	unit, err := s.unitRepo.GetByID(req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("unit not found: %w", err)
	}

	// Validate building and property IDs match unit's hierarchy
	if unit.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("building ID mismatch with unit's building")
	}
	if unit.PropertyID != req.PropertyID {
		return nil, fmt.Errorf("property ID mismatch with unit's property")
	}

	// Validate the unit actually belongs to the caller's organization — prevents
	// creating a payment record against another organization's unit (IDOR).
	if unit.OrganizationID != req.OrganizationID {
		return nil, fmt.Errorf("unit not found")
	}

	// Get building information for context
	building, err := s.buildingRepo.GetByID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get property information for context
	property, err := s.propertyRepo.GetByID(req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Create payment with building context
	payment, err := s.paymentRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Log audit with complete building context
	s.auditService.LogUserAction(userID, "CREATE", "payments", &payment.ID, nil, map[string]interface{}{
		"payment_id":    payment.ID,
		"unit_id":       payment.UnitID,
		"building_id":   payment.BuildingID,
		"property_id":   payment.PropertyID,
		"building_name": building.BuildingName,
		"building_code": building.BuildingCode,
		"property_name": property.PropertyName,
		"unit_number":   unit.UnitNumber,
		"amount_due":    payment.AmountDue,
		"month":         payment.Month,
		"year":          payment.Year,
	})

	return payment, nil
}

// GetPayment retrieves a payment with building and property context
func (s *PaymentService) GetPayment(id, orgID int) (*models.PaymentWithDetails, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment with details: %w", err)
	}
	if payment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}
	return payment, nil
}

// UpdatePayment updates a payment record with building context logging
func (s *PaymentService) UpdatePayment(id int, req *models.UpdatePaymentRequest, userID, orgID int) (*models.Payment, error) {
	// Get existing payment for audit and building context
	existingPayment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing payment: %w", err)
	}
	if existingPayment.OrganizationID != orgID {
		return nil, fmt.Errorf("payment not found")
	}

	// Update payment
	updatedPayment, err := s.paymentRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Log audit with building context
	s.auditService.LogUserAction(userID, "UPDATE", "payments", &id, map[string]interface{}{
		"payment_id":      existingPayment.ID,
		"building_name":   existingPayment.BuildingName,
		"building_code":   existingPayment.BuildingCode,
		"property_name":   existingPayment.PropertyName,
		"unit_number":     existingPayment.UnitNumber,
		"old_status":      existingPayment.Status,
		"old_amount_paid": existingPayment.AmountPaid,
	}, map[string]interface{}{
		"payment_id":      updatedPayment.ID,
		"building_id":     updatedPayment.BuildingID,
		"property_id":     updatedPayment.PropertyID,
		"new_status":      updatedPayment.Status,
		"new_amount_paid": updatedPayment.AmountPaid,
	})

	return updatedPayment, nil
}

// GetPayments retrieves payments scoped only by the provided filters (no entity validation)
func (s *PaymentService) GetPayments(page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments: %w", err)
	}
	return payments, total, nil
}

// GetPaymentsByBuilding retrieves payments for a specific building
func (s *PaymentService) GetPaymentsByBuilding(buildingID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	// Validate building exists
	_, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, 0, fmt.Errorf("building not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Add building filter
	filters["building_id"] = buildingID

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments by building: %w", err)
	}

	return payments, total, nil
}

// GetPaymentsByProperty retrieves payments for a specific property with building breakdown
func (s *PaymentService) GetPaymentsByProperty(propertyID int, page, pageSize int, filters map[string]interface{}) ([]*models.PaymentWithDetails, int, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, 0, fmt.Errorf("property not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Add property filter
	filters["property_id"] = propertyID

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get payments by property: %w", err)
	}

	return payments, total, nil
}

// GenerateBuildingPaymentReport generates payment report for a building
func (s *PaymentService) GenerateBuildingPaymentReport(buildingID, orgID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
	// Validate building exists and get details
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("building not found")
	}

	// Get property details for context
	property, err := s.propertyRepo.GetByID(building.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get payment statistics for the building
	stats, err := s.paymentRepo.GetBuildingPaymentStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment stats: %w", err)
	}

	// Get detailed payment records
	payments, _, err := s.paymentRepo.GetBuildingPaymentsInPeriod(buildingID, startDate, endDate, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payments: %w", err)
	}

	report := &models.BuildingPaymentReport{
		BuildingID:   buildingID,
		BuildingName: building.BuildingName,
		BuildingCode: building.BuildingCode,
		PropertyID:   property.ID,
		PropertyName: property.PropertyName,
		ReportPeriod: fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Statistics:   stats,
		Payments:     payments,
		GeneratedAt:  time.Now(),
	}

	return report, nil
}

// GeneratePropertyPaymentReport generates payment report for a property with building breakdowns
func (s *PaymentService) GeneratePropertyPaymentReport(propertyID, orgID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
	// Validate property exists and get details
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}
	if property.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}

	// Get buildings in the property
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property buildings: %w", err)
	}

	// Get overall property payment statistics
	propertyStats, err := s.paymentRepo.GetPropertyPaymentStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}

	// Get building-level breakdowns
	buildingBreakdowns := make([]*models.BuildingPaymentBreakdown, 0, len(buildings))
	for _, building := range buildings {
		buildingStats, err := s.paymentRepo.GetBuildingPaymentStats(building.ID, startDate, endDate)
		if err != nil {
			continue // Skip buildings with errors but don't fail the entire report
		}

		breakdown := &models.BuildingPaymentBreakdown{
			BuildingID:   building.ID,
			BuildingName: building.BuildingName,
			BuildingCode: building.BuildingCode,
			BuildingType: string(building.BuildingType),
			Statistics:   buildingStats,
		}
		buildingBreakdowns = append(buildingBreakdowns, breakdown)
	}

	report := &models.PropertyPaymentReport{
		PropertyID:         propertyID,
		PropertyName:       property.PropertyName,
		PropertyCode:       property.PropertyCode,
		ReportPeriod:       fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		OverallStatistics:  propertyStats,
		BuildingBreakdowns: buildingBreakdowns,
		GeneratedAt:        time.Now(),
	}

	return report, nil
}

// GetDashboardSummaryWithBuildingContext returns dashboard summary scoped to the given org.
func (s *PaymentService) GetDashboardSummaryWithBuildingContext(orgID int) (*models.DashboardSummary, error) {
	summary, err := s.paymentRepo.GetDashboardSummary(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	buildingLevel, err := s.paymentRepo.GetBuildingLevelSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get building-level summary: %w", err)
	}
	if count, ok := buildingLevel["total_buildings"].(int); ok {
		summary.BuildingCount = count
	}
	return summary, nil
}

// ProcessBulkPayments processes multiple payments with building context validation
func (s *PaymentService) ProcessBulkPayments(requests []*models.CreatePaymentRequest, userID int) ([]*models.Payment, []error) {
	payments := make([]*models.Payment, 0, len(requests))
	errors := make([]error, 0)

	for i, req := range requests {
		payment, err := s.CreatePayment(req, userID)
		if err != nil {
			errors = append(errors, fmt.Errorf("payment %d: %w", i+1, err))
			continue
		}
		payments = append(payments, payment)
	}

	return payments, errors
}

// GetPaymentAnalyticsByBuilding returns payment analytics for a specific building
func (s *PaymentService) GetPaymentAnalyticsByBuilding(buildingID int, period string) (*models.BuildingPaymentAnalytics, error) {
	// Validate building exists
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Calculate date range based on period
	endDate := time.Now()
	var startDate time.Time
	switch period {
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "quarter":
		startDate = endDate.AddDate(0, -3, 0)
	case "year":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, -1, 0) // Default to month
	}

	// Get payment analytics
	analytics, err := s.paymentRepo.GetBuildingPaymentAnalytics(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment analytics: %w", err)
	}

	// Add building context
	analytics.BuildingName = building.BuildingName
	analytics.BuildingCode = building.BuildingCode
	analytics.BuildingType = string(building.BuildingType)

	return analytics, nil
}

// CanUserAccessPayment verifies if a user can access a specific payment based on their role and organization
func (s *PaymentService) CanUserAccessPayment(userID int, userRole string, payment *models.PaymentWithDetails, userOrgID int) bool {
	// Must belong to same organization
	if payment.OrganizationID != userOrgID {
		return false
	}

	switch userRole {
	case "SUPER_ADMIN":
		return true // Can access all payments in their organization
	case "ORG_ADMIN":
		return true // Can access all org payments
	case "Admin":
		return true // Can access all org payments
	case "PropertyManager":
		// Can only access payments for properties they manage
		return s.canPropertyManagerAccessPayment(userID, payment.PropertyID)
	case "Accountant":
		return true // Read-only access to all payments in org
	default:
		return false
	}
}

// canPropertyManagerAccessPayment checks if a PropertyManager manages the specified property
func (s *PaymentService) canPropertyManagerAccessPayment(userID int, propertyID int) bool {
	// Get user to verify they manage this property
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}

	// PropertyManagers should have property assignment info
	// This assumes user model has a field indicating which properties they manage
	// For now, we'll check if they're assigned to this property via metadata or a relationship
	if user == nil {
		return false
	}

	// TODO: Implement property manager assignment lookup
	// For now, allow access if user is PropertyManager (assumes assignment validation elsewhere)
	return true
}

// LogPaymentAccess logs access to payment data for audit trail
func (s *PaymentService) LogPaymentAccess(userID int, action string, paymentID int, allowed bool) {
	accessLog := map[string]interface{}{
		"payment_id": paymentID,
		"allowed":    allowed,
		"action":     action,
	}

	if allowed {
		s.auditService.LogUserAction(userID, action, "payments", &paymentID, nil, accessLog)
	} else {
		s.auditService.LogUserAction(userID, fmt.Sprintf("%s_DENIED", action), "payments", &paymentID, nil, accessLog)
	}
}

// GenerateMonthlyPayments creates Due payment records for every active lease in the given month/year.
// Leases that already have a payment record for that unit/month/year are skipped.
func (s *PaymentService) GenerateMonthlyPayments(req *models.GenerateMonthlyPaymentsRequest, orgID, userID int) (*models.GenerateMonthlyPaymentsResult, error) {
	now := time.Now().UTC()
	currentYear, currentMonth := now.Year(), int(now.Month())

	// Reject dates more than 1 month ahead of today.
	reqMonths := req.Year*12 + req.Month
	currentMonths := currentYear*12 + currentMonth
	if reqMonths > currentMonths+1 {
		return nil, fmt.Errorf("cannot generate payments more than 1 month in the future (requested %04d-%02d)", req.Year, req.Month)
	}

	// Reject dates older than 24 months to prevent mass back-generation.
	if reqMonths < currentMonths-24 {
		return nil, fmt.Errorf("cannot generate payments more than 24 months in the past (requested %04d-%02d)", req.Year, req.Month)
	}

	leases, err := s.paymentRepo.GetActiveLeasesForPeriod(orgID, req.Month, req.Year, req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active leases: %w", err)
	}

	dueDay := req.DueDayOfMonth
	if dueDay < 1 || dueDay > 28 {
		dueDay = 7
	}
	lastDay := time.Date(req.Year, time.Month(req.Month+1), 0, 0, 0, 0, 0, time.UTC).Day()
	if dueDay > lastDay {
		dueDay = lastDay
	}
	dueDateStr := fmt.Sprintf("%04d-%02d-%02d", req.Year, req.Month, dueDay)

	result := &models.GenerateMonthlyPaymentsResult{Errors: []string{}}

	for _, lease := range leases {
		exists, err := s.paymentRepo.CheckPaymentExists(lease.UnitID, req.Month, req.Year)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("unit %d: existence check failed: %v", lease.UnitID, err))
			continue
		}
		if exists {
			result.Skipped++
			continue
		}

		createReq := &models.CreatePaymentRequest{
			UnitID:         lease.UnitID,
			TenantID:       lease.TenantID,
			BuildingID:     lease.BuildingID,
			PropertyID:     lease.PropertyID,
			OrganizationID: orgID,
			Month:          req.Month,
			Year:           req.Year,
			AmountDue:      lease.MonthlyRent,
			DueDate:        dueDateStr,
		}

		if _, err := s.paymentRepo.Create(createReq); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("unit %d: %v", lease.UnitID, err))
			continue
		}
		result.Generated++
	}

	s.auditService.LogUserAction(userID, "GENERATE_MONTHLY", "payments", nil, nil, map[string]interface{}{
		"month":       req.Month,
		"year":        req.Year,
		"org_id":      orgID,
		"generated":   result.Generated,
		"skipped":     result.Skipped,
		"failed":      result.Failed,
		"building_id": req.BuildingID,
	})

	return result, nil
}

// SearchLeases searches for active leases by tenant name, property, building, unit, or lease ID
func (s *PaymentService) SearchLeases(orgID int, query string) (*models.LeaseSearchResponse, error) {
	if query == "" {
		return &models.LeaseSearchResponse{Results: make([]*models.LeaseSearchResult, 0), Total: 0}, nil
	}

	results, err := s.paymentRepo.SearchLeases(orgID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search leases: %w", err)
	}

	return &models.LeaseSearchResponse{
		Results: results,
		Total:   len(results),
	}, nil
}
