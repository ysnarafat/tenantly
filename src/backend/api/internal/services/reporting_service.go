package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type ReportingService struct {
	propertyRepo     interfaces.PropertyRepositoryInterface
	buildingRepo     interfaces.BuildingRepositoryInterface
	unitRepo         interfaces.UnitRepositoryInterface
	paymentRepo      interfaces.PaymentRepositoryInterface
	notificationRepo interfaces.NotificationRepositoryInterface
	tenantRepo       interfaces.TenantRepositoryInterface
	auditService     interfaces.AuditServiceInterface
}

func NewReportingService(
	propertyRepo interfaces.PropertyRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	unitRepo interfaces.UnitRepositoryInterface,
	paymentRepo interfaces.PaymentRepositoryInterface,
	notificationRepo interfaces.NotificationRepositoryInterface,
	tenantRepo interfaces.TenantRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *ReportingService {
	return &ReportingService{
		propertyRepo:     propertyRepo,
		buildingRepo:     buildingRepo,
		unitRepo:         unitRepo,
		paymentRepo:      paymentRepo,
		notificationRepo: notificationRepo,
		tenantRepo:       tenantRepo,
		auditService:     auditService,
	}
}

// GenerateComprehensiveReport generates a comprehensive report with building-level breakdowns
func (s *ReportingService) GenerateComprehensiveReport(
	propertyID *int,
	buildingID *int,
	startDate, endDate time.Time,
	userID int,
) (*models.ComprehensiveReport, error) {
	var report *models.ComprehensiveReport
	var err error

	if buildingID != nil {
		// Generate building-specific comprehensive report
		report, err = s.generateBuildingComprehensiveReport(*buildingID, startDate, endDate)
	} else if propertyID != nil {
		// Generate property-wide comprehensive report with building breakdowns
		report, err = s.generatePropertyComprehensiveReport(*propertyID, startDate, endDate)
	} else {
		// Generate system-wide comprehensive report
		report, err = s.generateSystemComprehensiveReport(startDate, endDate)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate comprehensive report: %w", err)
	}

	// Log report generation audit
	s.auditService.LogUserAction(userID, "GENERATE_REPORT", "reports", nil, nil, map[string]interface{}{
		"report_type":  report.ReportType,
		"property_id":  propertyID,
		"building_id":  buildingID,
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Format("2006-01-02"),
		"generated_at": report.GeneratedAt,
	})

	return report, nil
}

// generateBuildingComprehensiveReport generates comprehensive report for a specific building
func (s *ReportingService) generateBuildingComprehensiveReport(buildingID int, startDate, endDate time.Time) (*models.ComprehensiveReport, error) {
	// Get building details
	building, err := s.buildingRepo.GetWithStats(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get property details
	property, err := s.propertyRepo.GetByID(building.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get building analytics
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	// Get payment statistics
	paymentStats, err := s.paymentRepo.GetBuildingPaymentStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	// Get occupancy statistics
	occupancyStats, err := s.unitRepo.GetBuildingOccupancyStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get occupancy stats: %w", err)
	}

	// Get notification statistics
	notificationStats, err := s.notificationRepo.GetBuildingNotificationStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	report := &models.ComprehensiveReport{
		ReportType:        "building",
		BuildingID:        &buildingID,
		BuildingName:      &building.BuildingName,
		BuildingCode:      &building.BuildingCode,
		PropertyID:        &building.PropertyID,
		PropertyName:      &property.PropertyName,
		StartDate:         startDate,
		EndDate:           endDate,
		BuildingAnalytics: analytics,
		PaymentStats:      paymentStats,
		OccupancyStats:    occupancyStats,
		NotificationStats: notificationStats,
		GeneratedAt:       time.Now(),
	}

	return report, nil
}

// generatePropertyComprehensiveReport generates comprehensive report for a property with building breakdowns
func (s *ReportingService) generatePropertyComprehensiveReport(propertyID int, startDate, endDate time.Time) (*models.ComprehensiveReport, error) {
	// Get property details
	property, err := s.propertyRepo.GetByIDWithStats(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get all buildings in the property
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property buildings: %w", err)
	}

	// Get overall property statistics
	propertyPaymentStats, err := s.paymentRepo.GetPropertyPaymentStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}

	propertyOccupancyStats, err := s.unitRepo.GetPropertyOccupancyStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property occupancy stats: %w", err)
	}

	propertyNotificationStats, err := s.notificationRepo.GetPropertyNotificationStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property notification stats: %w", err)
	}

	// Generate building-level breakdowns
	buildingBreakdowns := make([]*models.BuildingReportBreakdown, 0, len(buildings))
	for _, building := range buildings {
		breakdown, err := s.generateBuildingBreakdown(building.ID, startDate, endDate)
		if err != nil {
			continue // Skip buildings with errors but don't fail the entire report
		}
		buildingBreakdowns = append(buildingBreakdowns, breakdown)
	}

	report := &models.ComprehensiveReport{
		ReportType:         "property",
		PropertyID:         &propertyID,
		PropertyName:       &property.PropertyName,
		PropertyCode:       &property.PropertyCode,
		StartDate:          startDate,
		EndDate:            endDate,
		PaymentStats:       propertyPaymentStats,
		OccupancyStats:     propertyOccupancyStats,
		NotificationStats:  propertyNotificationStats,
		BuildingBreakdowns: buildingBreakdowns,
		PropertyStatistics: s.calculatePropertyStatistics(property, buildingBreakdowns),
		GeneratedAt:        time.Now(),
	}

	return report, nil
}

// generateSystemComprehensiveReport generates system-wide comprehensive report
func (s *ReportingService) generateSystemComprehensiveReport(startDate, endDate time.Time) (*models.ComprehensiveReport, error) {
	// Get system-wide statistics
	systemPaymentStats, err := s.paymentRepo.GetSystemPaymentStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get system payment stats: %w", err)
	}

	systemOccupancyStats, err := s.unitRepo.GetSystemOccupancyStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get system occupancy stats: %w", err)
	}

	systemNotificationStats, err := s.notificationRepo.GetSystemNotificationStats(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get system notification stats: %w", err)
	}

	// Get property-level summaries
	propertySummaries, err := s.generatePropertySummaries(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property summaries: %w", err)
	}

	report := &models.ComprehensiveReport{
		ReportType:        "system",
		StartDate:         startDate,
		EndDate:           endDate,
		PaymentStats:      systemPaymentStats,
		OccupancyStats:    systemOccupancyStats,
		NotificationStats: systemNotificationStats,
		PropertySummaries: propertySummaries,
		SystemStatistics:  s.calculateSystemStatistics(propertySummaries),
		GeneratedAt:       time.Now(),
	}

	return report, nil
}

// generateBuildingBreakdown generates detailed breakdown for a specific building
func (s *ReportingService) generateBuildingBreakdown(buildingID int, startDate, endDate time.Time) (*models.BuildingReportBreakdown, error) {
	// Get building details
	building, err := s.buildingRepo.GetWithStats(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building not found: %w", err)
	}

	// Get building analytics
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	// Get payment statistics
	paymentStats, err := s.paymentRepo.GetBuildingPaymentStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	// Get occupancy statistics
	occupancyStats, err := s.unitRepo.GetBuildingOccupancyStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get occupancy stats: %w", err)
	}

	// Get notification statistics
	notificationStats, err := s.notificationRepo.GetBuildingNotificationStats(buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	// Get unit type distribution
	unitTypeDistribution, err := s.unitRepo.GetBuildingUnitTypeDistribution(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit type distribution: %w", err)
	}

	breakdown := &models.BuildingReportBreakdown{
		BuildingID:           buildingID,
		BuildingName:         building.BuildingName,
		BuildingCode:         building.BuildingCode,
		BuildingType:         string(building.BuildingType),
		TotalFloors:          building.TotalFloors,
		HasElevator:          building.HasElevator,
		ConstructionYear:     building.ConstructionYear,
		Analytics:            analytics,
		PaymentStats:         paymentStats,
		OccupancyStats:       occupancyStats,
		NotificationStats:    notificationStats,
		UnitTypeDistribution: unitTypeDistribution,
		PerformanceScore:     s.calculateBuildingPerformanceScore(paymentStats, occupancyStats),
	}

	return breakdown, nil
}

// generatePropertySummaries generates summaries for all properties
func (s *ReportingService) generatePropertySummaries(startDate, endDate time.Time) ([]*models.PropertyReportSummary, error) {
	// Get all active properties
	properties, _, err := s.propertyRepo.List(map[string]interface{}{"active": true}, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties: %w", err)
	}

	summaries := make([]*models.PropertyReportSummary, 0, len(properties))
	for _, property := range properties {
		summary, err := s.generatePropertySummary(property.ID, startDate, endDate)
		if err != nil {
			continue // Skip properties with errors
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// generatePropertySummary generates summary for a specific property
func (s *ReportingService) generatePropertySummary(propertyID int, startDate, endDate time.Time) (*models.PropertyReportSummary, error) {
	// Get property details
	property, err := s.propertyRepo.GetByIDWithStats(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property not found: %w", err)
	}

	// Get building count and types (would be implemented in repository)
	// For now, we'll use basic building count from property
	buildingStats := map[string]interface{}{
		"total_buildings": property.BuildingCount,
		"building_types":  []string{}, // Would be populated from repository
	}

	// Get payment statistics
	paymentStats, err := s.paymentRepo.GetPropertyPaymentStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	// Get occupancy statistics
	occupancyStats, err := s.unitRepo.GetPropertyOccupancyStats(propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get occupancy stats: %w", err)
	}

	summary := &models.PropertyReportSummary{
		PropertyID:       propertyID,
		PropertyName:     property.PropertyName,
		PropertyCode:     property.PropertyCode,
		PropertyType:     string(property.PropertyType),
		BuildingCount:    property.TotalBuildings, // Use TotalBuildings field
		UnitCount:        0,                       // Would need to be calculated from units or added to property model
		BuildingStats:    buildingStats,
		PaymentStats:     paymentStats,
		OccupancyStats:   occupancyStats,
		PerformanceScore: s.calculatePropertyPerformanceScore(paymentStats, occupancyStats),
	}

	return summary, nil
}

// GenerateDashboardReport generates dashboard report with building filtering and grouping
func (s *ReportingService) GenerateDashboardReport(
	filters map[string]interface{},
	groupBy string,
	userID int,
) (*models.DashboardReport, error) {
	// Get overall system statistics
	systemStats, err := s.getSystemStatistics(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get system statistics: %w", err)
	}

	// Generate groupings based on groupBy parameter
	var groupings interface{}
	switch groupBy {
	case "property":
		groupings, err = s.generatePropertyGroupings(filters)
	case "building":
		groupings, err = s.generateBuildingGroupings(filters)
	case "building_type":
		groupings, err = s.generateBuildingTypeGroupings(filters)
	default:
		groupings, err = s.generatePropertyGroupings(filters) // Default to property grouping
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate groupings: %w", err)
	}

	report := &models.DashboardReport{
		SystemStatistics: systemStats,
		Groupings:        groupings,
		GroupBy:          groupBy,
		Filters:          filters,
		GeneratedAt:      time.Now(),
	}

	// Log dashboard report generation
	s.auditService.LogUserAction(userID, "GENERATE_DASHBOARD", "reports", nil, nil, map[string]interface{}{
		"group_by":     groupBy,
		"filters":      filters,
		"generated_at": report.GeneratedAt,
	})

	return report, nil
}

// generatePropertyGroupings generates property-level groupings with building context
func (s *ReportingService) generatePropertyGroupings(filters map[string]interface{}) ([]*models.PropertyGrouping, error) {
	// Get properties based on filters
	properties, _, err := s.propertyRepo.List(filters, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties: %w", err)
	}

	groupings := make([]*models.PropertyGrouping, 0, len(properties))
	for _, property := range properties {
		// Get building statistics for the property (simplified for now)
		buildingStats := map[string]interface{}{
			"total_buildings": property.TotalBuildings,
			"building_types":  []string{}, // Would be populated from repository
		}

		// Get payment and occupancy statistics
		paymentStats, _ := s.paymentRepo.GetPropertyPaymentStats(property.ID, time.Now().AddDate(0, -1, 0), time.Now())
		occupancyStats, _ := s.unitRepo.GetPropertyOccupancyStats(property.ID, time.Now().AddDate(0, -1, 0), time.Now())

		grouping := &models.PropertyGrouping{
			PropertyID:     property.ID,
			PropertyName:   property.PropertyName,
			PropertyCode:   property.PropertyCode,
			PropertyType:   string(property.PropertyType),
			BuildingCount:  property.TotalBuildings, // Use TotalBuildings field
			UnitCount:      0,                       // Would need to be calculated or added to property model
			BuildingStats:  buildingStats,
			PaymentStats:   paymentStats,
			OccupancyStats: occupancyStats,
		}
		groupings = append(groupings, grouping)
	}

	return groupings, nil
}

// generateBuildingGroupings generates building-level groupings
func (s *ReportingService) generateBuildingGroupings(filters map[string]interface{}) ([]*models.BuildingGrouping, error) {
	// Get buildings based on filters (simplified search)
	searchFilters := &models.BuildingSearchFilters{
		// Convert generic filters to BuildingSearchFilters
		// This would need proper implementation based on the actual BuildingSearchFilters structure
	}
	buildings, err := s.buildingRepo.Search(searchFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	groupings := make([]*models.BuildingGrouping, 0, len(buildings))
	for _, building := range buildings {
		// Get building analytics
		analytics, err := s.buildingRepo.GetAnalytics(building.ID)
		if err != nil {
			continue // Skip buildings with errors
		}

		// Get payment and occupancy statistics
		paymentStats, _ := s.paymentRepo.GetBuildingPaymentStats(building.ID, time.Now().AddDate(0, -1, 0), time.Now())
		occupancyStats, _ := s.unitRepo.GetBuildingOccupancyStats(building.ID, time.Now().AddDate(0, -1, 0), time.Now())

		grouping := &models.BuildingGrouping{
			BuildingID:     building.ID,
			BuildingName:   building.BuildingName,
			BuildingCode:   building.BuildingCode,
			BuildingType:   string(building.BuildingType),
			PropertyID:     building.PropertyID,
			Analytics:      analytics,
			PaymentStats:   paymentStats,
			OccupancyStats: occupancyStats,
		}
		groupings = append(groupings, grouping)
	}

	return groupings, nil
}

// generateBuildingTypeGroupings generates building type-based groupings
func (s *ReportingService) generateBuildingTypeGroupings(filters map[string]interface{}) ([]*models.BuildingTypeGrouping, error) {
	buildingTypes := []string{"Residential", "Commercial", "Mixed"}
	groupings := make([]*models.BuildingTypeGrouping, 0, len(buildingTypes))

	for _, buildingType := range buildingTypes {
		// Get statistics for this building type
		typeFilters := make(map[string]interface{})
		for k, v := range filters {
			typeFilters[k] = v
		}
		typeFilters["building_type"] = buildingType

		stats, err := s.getBuildingTypeStatistics(typeFilters)
		if err != nil {
			continue // Skip building types with errors
		}

		grouping := &models.BuildingTypeGrouping{
			BuildingType: buildingType,
			Statistics:   stats,
		}
		groupings = append(groupings, grouping)
	}

	return groupings, nil
}

// Helper methods for calculations
func (s *ReportingService) calculateBuildingPerformanceScore(paymentStats, occupancyStats interface{}) float64 {
	// Implement performance score calculation logic
	// This is a simplified example - you would implement more sophisticated scoring
	return 85.0 // Placeholder
}

func (s *ReportingService) calculatePropertyPerformanceScore(paymentStats, occupancyStats interface{}) float64 {
	// Implement property performance score calculation logic
	return 82.0 // Placeholder
}

func (s *ReportingService) calculatePropertyStatistics(property *models.PropertyWithStats, breakdowns []*models.BuildingReportBreakdown) interface{} {
	// Implement property statistics calculation
	return map[string]interface{}{
		"total_buildings": len(breakdowns),
		"total_units":     property.UnitCount,
		"occupancy_rate":  85.5, // Placeholder
	}
}

func (s *ReportingService) calculateSystemStatistics(summaries []*models.PropertyReportSummary) interface{} {
	// Implement system statistics calculation
	totalProperties := len(summaries)
	totalBuildings := 0
	totalUnits := 0

	for _, summary := range summaries {
		totalBuildings += summary.BuildingCount
		totalUnits += summary.UnitCount
	}

	return map[string]interface{}{
		"total_properties": totalProperties,
		"total_buildings":  totalBuildings,
		"total_units":      totalUnits,
		"avg_occupancy":    87.2, // Placeholder
	}
}

func (s *ReportingService) getSystemStatistics(filters map[string]interface{}) (interface{}, error) {
	// Implement system statistics retrieval
	return map[string]interface{}{
		"total_revenue":    125000.0,
		"collection_rate":  92.5,
		"occupancy_rate":   87.2,
		"active_buildings": 45,
	}, nil
}

func (s *ReportingService) getBuildingTypeStatistics(filters map[string]interface{}) (interface{}, error) {
	// Implement building type statistics retrieval
	return map[string]interface{}{
		"building_count": 15,
		"unit_count":     120,
		"avg_occupancy":  85.0,
		"total_revenue":  45000.0,
	}, nil
}
