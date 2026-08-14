package services

import (
	"testing"

	"github.com/ysnarafat/tenantly/internal/models"
)

// TestBuildingAnalyticsIntegration tests integration between building analytics and other services
func TestBuildingAnalyticsIntegration_WithPropertyService(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test property and buildings
	propertyID := 1
	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "Building A",
			BuildingCode: "BA001",
			BuildingType: models.BuildingTypeCommercial,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Building B",
			BuildingCode: "BB001",
			BuildingType: models.BuildingTypeResidential,
			ActiveStatus: true,
		},
	}

	for _, building := range buildings {
		_ = mockRepo.Create(building)
		// Set different analytics for each building
		analytics := &models.BuildingAnalytics{
			BuildingID:     building.ID,
			UnitCount:      10 + building.ID*5,
			OccupiedUnits:  8 + building.ID*3,
			VacantUnits:    2 + building.ID*2,
			OccupancyRate:  80.0 + float64(building.ID)*5,
			MonthlyRevenue: 50000 + float64(building.ID)*10000,
			AverageRent:    2500 + float64(building.ID)*500,
			TotalArea:      1000 + float64(building.ID)*200,
		}
		mockRepo.SetAnalytics(building.ID, analytics)
	}

	// Test property-level building comparison
	comparison, err := service.CompareBuildingPerformance(propertyID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify integration results
	if comparison.PropertyID != propertyID {
		t.Errorf("Expected PropertyID %d, got %d", propertyID, comparison.PropertyID)
	}

	if comparison.TotalBuildings != len(buildings) {
		t.Errorf("Expected TotalBuildings %d, got %d", len(buildings), comparison.TotalBuildings)
	}

	if len(comparison.Rankings) != len(buildings) {
		t.Errorf("Expected %d rankings, got %d", len(buildings), len(comparison.Rankings))
	}

	// Verify rankings are properly calculated and sorted
	for i := 1; i < len(comparison.Rankings); i++ {
		if comparison.Rankings[i-1].PerformanceScore < comparison.Rankings[i].PerformanceScore {
			t.Error("Expected rankings to be sorted by performance score in descending order")
		}
	}

	// Test property averages calculation
	if comparison.Averages == nil {
		t.Error("Expected property averages to be calculated")
	}

	if comparison.Averages.AverageOccupancyRate <= 0 {
		t.Error("Expected positive average occupancy rate")
	}

	if comparison.Averages.AverageRevenue <= 0 {
		t.Error("Expected positive average revenue")
	}
}

func TestBuildingAnalyticsIntegration_WithUnitService(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Set analytics with unit-related data
	analytics := &models.BuildingAnalytics{
		BuildingID:     building.ID,
		UnitCount:      20,
		OccupiedUnits:  16,
		VacantUnits:    4,
		OccupancyRate:  80.0,
		MonthlyRevenue: 100000,
		AverageRent:    5000,
		TotalArea:      2000,
	}
	mockRepo.SetAnalytics(building.ID, analytics)

	// Test building metrics calculation
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify unit-related metrics are properly integrated
	if metrics.UnitCount != analytics.UnitCount {
		t.Errorf("Expected UnitCount %d, got %d", analytics.UnitCount, metrics.UnitCount)
	}

	if metrics.OccupiedUnits != analytics.OccupiedUnits {
		t.Errorf("Expected OccupiedUnits %d, got %d", analytics.OccupiedUnits, metrics.OccupiedUnits)
	}

	if metrics.VacantUnits != analytics.VacantUnits {
		t.Errorf("Expected VacantUnits %d, got %d", analytics.VacantUnits, metrics.VacantUnits)
	}

	// Test occupancy analytics integration
	occupancyAnalytics, err := service.GetOccupancyAnalytics(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if occupancyAnalytics.CurrentOccupancy.TotalUnits != analytics.UnitCount {
		t.Errorf("Expected TotalUnits %d, got %d", analytics.UnitCount, occupancyAnalytics.CurrentOccupancy.TotalUnits)
	}

	if occupancyAnalytics.CurrentOccupancy.OccupiedUnits != analytics.OccupiedUnits {
		t.Errorf("Expected OccupiedUnits %d, got %d", analytics.OccupiedUnits, occupancyAnalytics.CurrentOccupancy.OccupiedUnits)
	}
}

func TestBuildingAnalyticsIntegration_WithPaymentService(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Revenue Building",
		BuildingCode: "RB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Set analytics with revenue data
	analytics := &models.BuildingAnalytics{
		BuildingID:     building.ID,
		UnitCount:      15,
		OccupiedUnits:  12,
		VacantUnits:    3,
		OccupancyRate:  80.0,
		MonthlyRevenue: 75000,
		AverageRent:    6250,
		TotalArea:      1500,
	}
	mockRepo.SetAnalytics(building.ID, analytics)

	// Test revenue analytics integration
	revenueAnalytics, err := service.GetRevenueAnalytics(building.ID, "month")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify revenue integration
	if revenueAnalytics.CurrentPeriod.TotalRevenue != analytics.MonthlyRevenue {
		t.Errorf("Expected TotalRevenue %f, got %f", analytics.MonthlyRevenue, revenueAnalytics.CurrentPeriod.TotalRevenue)
	}

	if revenueAnalytics.CurrentPeriod.AverageRent != analytics.AverageRent {
		t.Errorf("Expected AverageRent %f, got %f", analytics.AverageRent, revenueAnalytics.CurrentPeriod.AverageRent)
	}

	// Test revenue breakdown integration
	if revenueAnalytics.RevenueBreakdown == nil {
		t.Error("Expected revenue breakdown to be calculated")
	}

	if len(revenueAnalytics.RevenueBreakdown.ByUnitType) == 0 {
		t.Error("Expected revenue breakdown by unit type")
	}

	if len(revenueAnalytics.RevenueBreakdown.ByFloor) == 0 {
		t.Error("Expected revenue breakdown by floor")
	}

	// Test performance metrics integration
	if revenueAnalytics.PerformanceMetrics == nil {
		t.Error("Expected performance metrics to be calculated")
	}

	if revenueAnalytics.PerformanceMetrics.RevenuePerUnit <= 0 {
		t.Error("Expected positive revenue per unit")
	}

	if revenueAnalytics.PerformanceMetrics.CollectionEfficiency <= 0 {
		t.Error("Expected positive collection efficiency")
	}
}

func TestBuildingAnalyticsIntegration_WithReportingService(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	propertyID := 1

	// Create multiple buildings for comprehensive reporting
	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "Commercial Tower",
			BuildingCode: "CT001",
			BuildingType: models.BuildingTypeCommercial,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Residential Complex",
			BuildingCode: "RC001",
			BuildingType: models.BuildingTypeResidential,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Mixed Use Building",
			BuildingCode: "MU001",
			BuildingType: models.BuildingTypeMixed,
			ActiveStatus: true,
		},
	}

	// Use wide variance so performance spread > 30 and insights are generated.
	analyticsData := []*models.BuildingAnalytics{
		{OccupancyRate: 40, OccupiedUnits: 8, UnitCount: 20, VacantUnits: 12, MonthlyRevenue: 4000, AverageRent: 200, TotalArea: 1200},
		{OccupancyRate: 75, OccupiedUnits: 23, UnitCount: 30, VacantUnits: 7, MonthlyRevenue: 80000, AverageRent: 2666, TotalArea: 1600},
		{OccupancyRate: 95, OccupiedUnits: 38, UnitCount: 40, VacantUnits: 2, MonthlyRevenue: 100000, AverageRent: 2500, TotalArea: 2000},
	}
	for i, building := range buildings {
		_ = mockRepo.Create(building)
		a := analyticsData[i]
		a.BuildingID = building.ID
		mockRepo.SetAnalytics(building.ID, a)
	}

	// Test comprehensive property reporting
	comparison, err := service.CompareBuildingPerformance(propertyID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify reporting integration
	if len(comparison.Rankings) != len(buildings) {
		t.Errorf("Expected %d buildings in report, got %d", len(buildings), len(comparison.Rankings))
	}

	// Test top and under performers identification
	if len(comparison.TopPerformers) == 0 {
		t.Error("Expected top performers to be identified")
	}

	if len(comparison.UnderPerformers) == 0 {
		t.Error("Expected under performers to be identified")
	}

	// Test insights generation
	if len(comparison.Insights) == 0 {
		t.Error("Expected insights to be generated for reporting")
	}

	// Verify performance spread analysis
	if len(comparison.Rankings) > 1 {
		spread := comparison.Rankings[0].PerformanceScore - comparison.Rankings[len(comparison.Rankings)-1].PerformanceScore
		if spread <= 0 {
			t.Error("Expected performance spread between buildings")
		}
	}

	// Test individual building metrics for reporting
	for _, building := range buildings {
		metrics, err := service.GetBuildingMetrics(building.ID)
		if err != nil {
			t.Errorf("Failed to get metrics for building %d: %v", building.ID, err)
			continue
		}

		if metrics.PerformanceScore <= 0 {
			t.Errorf("Expected positive performance score for building %d", building.ID)
		}

		if metrics.Trends == nil {
			t.Errorf("Expected trends data for building %d reporting", building.ID)
		}

		if metrics.Comparisons == nil {
			t.Errorf("Expected comparisons data for building %d reporting", building.ID)
		}
	}
}

func TestBuildingAnalyticsIntegration_CrossServiceDataConsistency(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Consistency Test Building",
		BuildingCode: "CTB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Set consistent analytics data
	analytics := &models.BuildingAnalytics{
		BuildingID:     building.ID,
		UnitCount:      25,
		OccupiedUnits:  20,
		VacantUnits:    5,
		OccupancyRate:  80.0,
		MonthlyRevenue: 125000,
		AverageRent:    6250,
		TotalArea:      2500,
	}
	mockRepo.SetAnalytics(building.ID, analytics)

	// Get data from different service methods
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Fatalf("Failed to get building metrics: %v", err)
	}

	occupancyAnalytics, err := service.GetOccupancyAnalytics(building.ID)
	if err != nil {
		t.Fatalf("Failed to get occupancy analytics: %v", err)
	}

	revenueAnalytics, err := service.GetRevenueAnalytics(building.ID, "month")
	if err != nil {
		t.Fatalf("Failed to get revenue analytics: %v", err)
	}

	// Verify data consistency across services
	if metrics.UnitCount != occupancyAnalytics.CurrentOccupancy.TotalUnits {
		t.Errorf("Unit count inconsistency: metrics=%d, occupancy=%d",
			metrics.UnitCount, occupancyAnalytics.CurrentOccupancy.TotalUnits)
	}

	if metrics.OccupiedUnits != occupancyAnalytics.CurrentOccupancy.OccupiedUnits {
		t.Errorf("Occupied units inconsistency: metrics=%d, occupancy=%d",
			metrics.OccupiedUnits, occupancyAnalytics.CurrentOccupancy.OccupiedUnits)
	}

	if metrics.MonthlyRevenue != revenueAnalytics.CurrentPeriod.TotalRevenue {
		t.Errorf("Revenue inconsistency: metrics=%f, revenue=%f",
			metrics.MonthlyRevenue, revenueAnalytics.CurrentPeriod.TotalRevenue)
	}

	if metrics.AverageRent != revenueAnalytics.CurrentPeriod.AverageRent {
		t.Errorf("Average rent inconsistency: metrics=%f, revenue=%f",
			metrics.AverageRent, revenueAnalytics.CurrentPeriod.AverageRent)
	}

	// Verify calculated fields consistency
	expectedOccupancyRate := float64(analytics.OccupiedUnits) / float64(analytics.UnitCount) * 100
	if metrics.OccupancyRate != expectedOccupancyRate {
		t.Errorf("Occupancy rate calculation inconsistency: expected=%f, got=%f",
			expectedOccupancyRate, metrics.OccupancyRate)
	}

	if occupancyAnalytics.CurrentOccupancy.OccupancyRate != expectedOccupancyRate {
		t.Errorf("Occupancy analytics rate inconsistency: expected=%f, got=%f",
			expectedOccupancyRate, occupancyAnalytics.CurrentOccupancy.OccupancyRate)
	}
}

func TestBuildingAnalyticsIntegration_ErrorPropagation(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Test error propagation from repository layer
	nonExistentBuildingID := 999

	// Test metrics error propagation
	_, err := service.GetBuildingMetrics(nonExistentBuildingID)
	if err == nil {
		t.Error("Expected error for non-existent building in metrics")
	}

	// Test occupancy analytics error propagation
	_, err = service.GetOccupancyAnalytics(nonExistentBuildingID)
	if err == nil {
		t.Error("Expected error for non-existent building in occupancy analytics")
	}

	// Test revenue analytics error propagation
	_, err = service.GetRevenueAnalytics(nonExistentBuildingID, "month")
	if err == nil {
		t.Error("Expected error for non-existent building in revenue analytics")
	}

	// Test comparison error propagation
	_, err = service.CompareBuildingPerformance(999)
	if err == nil {
		t.Error("Expected error for non-existent property in comparison")
	}

	// Test with building that exists but has no analytics
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "No Analytics Building",
		BuildingCode: "NAB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)
	// Don't set analytics for this building

	// Should still work with default analytics
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Errorf("Expected default analytics to be used, got error: %v", err)
	}

	if metrics == nil {
		t.Error("Expected metrics with default values")
	}
}

func TestBuildingAnalyticsIntegration_ServiceDependencyInjection(t *testing.T) {
	// Test that the analytics service properly integrates with injected dependencies
	mockRepo := NewMockAnalyticsBuildingRepository()

	// Test with nil database (should not cause panic)
	service := NewBuildingAnalyticsService(mockRepo, nil)
	if service == nil {
		t.Error("Expected service to be created even with nil database")
	}

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Dependency Test Building",
		BuildingCode: "DTB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Test that service methods work with injected repository
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Fatalf("Expected service to work with injected repository: %v", err)
	}

	if metrics.BuildingID != building.ID {
		t.Errorf("Expected BuildingID %d, got %d", building.ID, metrics.BuildingID)
	}

	// Test cache functionality with injected dependencies
	cachedMetrics, found := service.GetCachedMetrics(building.ID)
	if !found {
		t.Error("Expected metrics to be cached after first call")
	}

	if cachedMetrics.BuildingID != metrics.BuildingID {
		t.Error("Expected cached metrics to match original metrics")
	}
}
