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
}

func NewPaymentService(
	paymentRepo interfaces.PaymentRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo interfaces.PropertyRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *PaymentService {
	return &PaymentService{
		paymentRepo:  paymentRepo,
		unitRepo:     unitRepo,
		buildingRepo: buildingRepo,
		propertyRepo: propertyRepo,
		auditService: auditService,
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
func (s *PaymentService) GetPayment(id int) (*models.PaymentWithDetails, error) {
	payment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment with details: %w", err)
	}
	return payment, nil
}

// UpdatePayment updates a payment record with building context logging
func (s *PaymentService) UpdatePayment(id int, req *models.UpdatePaymentRequest, userID int) (*models.Payment, error) {
	// Get existing payment for audit and building context
	existingPayment, err := s.paymentRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing payment: %w", err)
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
func (s *PaymentService) GenerateBuildingPaymentReport(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentReport, error) {
	// Validate building exists and get details
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
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
func (s *PaymentService) GeneratePropertyPaymentReport(propertyID int, startDate, endDate time.Time) (*models.PropertyPaymentReport, error) {
	// Validate property exists and get details
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
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

// GetDashboardSummaryWithBuildingContext returns dashboard summary with building-level context
func (s *PaymentService) GetDashboardSummaryWithBuildingContext() (*models.DashboardSummary, error) {
	summary, err := s.paymentRepo.GetDashboardSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}

	// Enhance with building context
	buildingStats, err := s.paymentRepo.GetBuildingLevelSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get building-level summary: %w", err)
	}

	// Add building context to summary
	summary.BuildingCount = buildingStats["total_buildings"].(int)
	// Note: Additional building context fields would need to be added to DashboardSummary model
	// For now, we'll store them in a separate structure or extend the model

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
