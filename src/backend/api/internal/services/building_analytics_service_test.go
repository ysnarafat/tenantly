package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// MockAnalyticsBuildingRepository implements BuildingRepositoryInterface for testing
type MockAnalyticsBuildingRepository struct {
	buildings map[int]*models.Building
	analytics map[int]*models.BuildingAnalytics
}

func NewMockAnalyticsBuildingRepository() *MockAnalyticsBuildingRepository {
	return &MockAnalyticsBuildingRepository{
		buildings: make(map[int]*models.Building),
		analytics: make(map[int]*models.BuildingAnalytics),
	}
}

func (m *MockAnalyticsBuildingRepository) Create(building *models.Building) error {
	building.ID = len(m.buildings) + 1
	building.CreatedAt = time.Now()
	building.UpdatedAt = time.Now()
	m.buildings[building.ID] = building
	return nil
}

func (m *MockAnalyticsBuildingRepository) GetByID(id int) (*models.Building, error) {
	if building, exists := m.buildings[id]; exists {
		return building, nil
	}
	return nil, sql.ErrNoRows
}

func (m *MockAnalyticsBuildingRepository) GetByPropertyID(propertyID int) ([]*models.Building, error) {
	var buildings []*models.Building
	for _, building := range m.buildings {
		if building.PropertyID == propertyID {
			buildings = append(buildings, building)
		}
	}
	return buildings, nil
}

func (m *MockAnalyticsBuildingRepository) GetByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	for _, building := range m.buildings {
		if building.PropertyID == propertyID && building.BuildingCode == code {
			return building, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *MockAnalyticsBuildingRepository) Update(id int, updates map[string]interface{}) error {
	if building, exists := m.buildings[id]; exists {
		building.UpdatedAt = time.Now()
		return nil
	}
	return sql.ErrNoRows
}

func (m *MockAnalyticsBuildingRepository) SoftDelete(id int) error {
	if building, exists := m.buildings[id]; exists {
		building.ActiveStatus = false
		building.UpdatedAt = time.Now()
		return nil
	}
	return sql.ErrNoRows
}

func (m *MockAnalyticsBuildingRepository) GetWithStats(id int) (*models.BuildingWithStats, error) {
	building, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}

	analytics := m.analytics[id]
	if analytics == nil {
		analytics = &models.BuildingAnalytics{
			BuildingID:     id,
			UnitCount:      10,
			OccupiedUnits:  8,
			VacantUnits:    2,
			OccupancyRate:  80.0,
			MonthlyRevenue: 50000,
			AverageRent:    2500,
			TotalArea:      1000,
		}
	}

	return &models.BuildingWithStats{
		Building:      *building,
		PropertyName:  "Test Property",
		UnitCount:     analytics.UnitCount,
		OccupiedUnits: analytics.OccupiedUnits,
		TotalRevenue:  analytics.MonthlyRevenue,
		OccupancyRate: analytics.OccupancyRate,
	}, nil
}

func (m *MockAnalyticsBuildingRepository) BulkCreate(buildings []*models.Building) error {
	for _, building := range buildings {
		_ = m.Create(building)
	}
	return nil
}

func (m *MockAnalyticsBuildingRepository) Search(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	var results []*models.Building
	for _, building := range m.buildings {
		if filters.PropertyID != nil && building.PropertyID != *filters.PropertyID {
			continue
		}
		if filters.BuildingType != nil && building.BuildingType != *filters.BuildingType {
			continue
		}
		if filters.ActiveStatus != nil && building.ActiveStatus != *filters.ActiveStatus {
			continue
		}
		results = append(results, building)
	}
	return results, nil
}

func (m *MockAnalyticsBuildingRepository) GetAnalytics(id int) (*models.BuildingAnalytics, error) {
	// Check if building exists first
	if _, exists := m.buildings[id]; !exists {
		return nil, sql.ErrNoRows
	}

	if analytics, exists := m.analytics[id]; exists {
		return analytics, nil
	}

	// Return default analytics if not set but building exists
	return &models.BuildingAnalytics{
		BuildingID:     id,
		UnitCount:      10,
		OccupiedUnits:  8,
		VacantUnits:    2,
		OccupancyRate:  80.0,
		MonthlyRevenue: 50000,
		AverageRent:    2500,
		TotalArea:      1000,
	}, nil
}

func (m *MockAnalyticsBuildingRepository) CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error) {
	count := 0
	for _, building := range m.buildings {
		if building.PropertyID == propertyID {
			count++
		}
	}
	return count, nil
}

func (m *MockAnalyticsBuildingRepository) GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error) {
	return m.GetByPropertyID(propertyID)
}

func (m *MockAnalyticsBuildingRepository) AdvancedSearch(req *models.BuildingSearchRequest) ([]*models.Building, int, error) {
	var results []*models.Building
	for _, building := range m.buildings {
		results = append(results, building)
	}
	return results, len(results), nil
}

func (m *MockAnalyticsBuildingRepository) GetBuildingUnits(buildingID int, offset, limit int) ([]*models.BuildingUnitSummary, int, error) {
	units := []*models.BuildingUnitSummary{
		{
			UnitID:      1,
			UnitNumber:  "101",
			UnitName:    "Shop 101",
			Floor:       1,
			Section:     "A",
			UnitType:    "Shop",
			Active:      true,
			TenantName:  "John Doe",
			LeaseActive: true,
		},
	}
	return units, len(units), nil
}

// Helper method to set analytics for testing
func (m *MockAnalyticsBuildingRepository) SetAnalytics(buildingID int, analytics *models.BuildingAnalytics) {
	m.analytics[buildingID] = analytics
}

func setupTestAnalyticsService() (*BuildingAnalyticsService, *MockAnalyticsBuildingRepository) {
	mockRepo := NewMockAnalyticsBuildingRepository()
	service := NewBuildingAnalyticsService(mockRepo, nil)
	return service, mockRepo
}

func TestBuildingAnalyticsService_GetBuildingMetrics(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeCommercial,
		TotalFloors:  5,
		HasElevator:  true,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Set test analytics
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

	// Test GetBuildingMetrics
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if metrics.BuildingID != building.ID {
		t.Errorf("Expected BuildingID %d, got %d", building.ID, metrics.BuildingID)
	}

	if metrics.BuildingName != building.BuildingName {
		t.Errorf("Expected BuildingName %s, got %s", building.BuildingName, metrics.BuildingName)
	}

	if metrics.UnitCount != analytics.UnitCount {
		t.Errorf("Expected UnitCount %d, got %d", analytics.UnitCount, metrics.UnitCount)
	}

	if metrics.OccupancyRate != analytics.OccupancyRate {
		t.Errorf("Expected OccupancyRate %f, got %f", analytics.OccupancyRate, metrics.OccupancyRate)
	}

	if metrics.PerformanceScore <= 0 {
		t.Errorf("Expected positive PerformanceScore, got %f", metrics.PerformanceScore)
	}
}

func TestBuildingAnalyticsService_GetOccupancyAnalytics(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Test GetOccupancyAnalytics
	occupancyAnalytics, err := service.GetOccupancyAnalytics(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if occupancyAnalytics.BuildingID != building.ID {
		t.Errorf("Expected BuildingID %d, got %d", building.ID, occupancyAnalytics.BuildingID)
	}

	if occupancyAnalytics.CurrentOccupancy == nil {
		t.Error("Expected CurrentOccupancy to be set")
	}

	if len(occupancyAnalytics.HistoricalOccupancy) == 0 {
		t.Error("Expected HistoricalOccupancy to have data")
	}

	if occupancyAnalytics.VacancyAnalysis == nil {
		t.Error("Expected VacancyAnalysis to be set")
	}

	if occupancyAnalytics.UtilizationMetrics == nil {
		t.Error("Expected UtilizationMetrics to be set")
	}
}

func TestBuildingAnalyticsService_GetRevenueAnalytics(t *testing.T) {
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

	// Test GetRevenueAnalytics for different periods
	periods := []string{"month", "quarter", "year"}

	for _, period := range periods {
		revenueAnalytics, err := service.GetRevenueAnalytics(building.ID, period)
		if err != nil {
			t.Fatalf("Expected no error for period %s, got %v", period, err)
		}

		if revenueAnalytics.BuildingID != building.ID {
			t.Errorf("Expected BuildingID %d, got %d", building.ID, revenueAnalytics.BuildingID)
		}

		if revenueAnalytics.Period != period {
			t.Errorf("Expected Period %s, got %s", period, revenueAnalytics.Period)
		}

		if revenueAnalytics.CurrentPeriod == nil {
			t.Error("Expected CurrentPeriod to be set")
		}

		if revenueAnalytics.PreviousPeriod == nil {
			t.Error("Expected PreviousPeriod to be set")
		}

		if revenueAnalytics.YearToDate == nil {
			t.Error("Expected YearToDate to be set")
		}

		if revenueAnalytics.RevenueBreakdown == nil {
			t.Error("Expected RevenueBreakdown to be set")
		}

		if revenueAnalytics.PerformanceMetrics == nil {
			t.Error("Expected PerformanceMetrics to be set")
		}
	}
}

func TestBuildingAnalyticsService_CompareBuildingPerformance(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	propertyID := 1

	// Create multiple test buildings
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
		{
			PropertyID:   propertyID,
			BuildingName: "Building C",
			BuildingCode: "BC001",
			BuildingType: models.BuildingTypeMixed,
			ActiveStatus: true,
		},
	}

	for _, building := range buildings {
		_ = mockRepo.Create(building)
	}

	// Test CompareBuildingPerformance
	comparison, err := service.CompareBuildingPerformance(propertyID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comparison.PropertyID != propertyID {
		t.Errorf("Expected PropertyID %d, got %d", propertyID, comparison.PropertyID)
	}

	if comparison.TotalBuildings != len(buildings) {
		t.Errorf("Expected TotalBuildings %d, got %d", len(buildings), comparison.TotalBuildings)
	}

	if len(comparison.Rankings) != len(buildings) {
		t.Errorf("Expected %d rankings, got %d", len(buildings), len(comparison.Rankings))
	}

	if comparison.Averages == nil {
		t.Error("Expected Averages to be set")
	}

	// Verify rankings are sorted by performance score
	for i := 1; i < len(comparison.Rankings); i++ {
		if comparison.Rankings[i-1].PerformanceScore < comparison.Rankings[i].PerformanceScore {
			t.Error("Expected rankings to be sorted by performance score in descending order")
		}
	}

	// Verify rank assignment
	for i, ranking := range comparison.Rankings {
		if ranking.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d", i+1, ranking.Rank)
		}
	}
}

func TestBuildingAnalyticsService_CalculatePerformanceScore(t *testing.T) {
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

	// Set test analytics with known values
	analytics := &models.BuildingAnalytics{
		BuildingID:     building.ID,
		UnitCount:      10,
		OccupiedUnits:  9,
		VacantUnits:    1,
		OccupancyRate:  90.0,
		MonthlyRevenue: 45000,
		AverageRent:    5000,
		TotalArea:      1000,
	}
	mockRepo.SetAnalytics(building.ID, analytics)

	// Test CalculatePerformanceScore
	score, err := service.CalculatePerformanceScore(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("Expected performance score between 0 and 100, got %f", score)
	}

	// Score should be reasonable for good occupancy and revenue
	if score < 50 {
		t.Errorf("Expected higher performance score for good metrics, got %f", score)
	}
}

func TestBuildingAnalyticsService_GenerateTrendAnalysis(t *testing.T) {
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

	// Test GenerateTrendAnalysis
	trends, err := service.GenerateTrendAnalysis(building.ID, "month")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if trends.OccupancyTrend == nil {
		t.Error("Expected OccupancyTrend to be set")
	}

	if trends.RevenueTrend == nil {
		t.Error("Expected RevenueTrend to be set")
	}

	if trends.RentTrend == nil {
		t.Error("Expected RentTrend to be set")
	}

	if trends.PeriodComparison != "month" {
		t.Errorf("Expected PeriodComparison 'month', got %s", trends.PeriodComparison)
	}

	// Verify trend data structure
	if trends.OccupancyTrend.Direction == "" {
		t.Error("Expected OccupancyTrend Direction to be set")
	}

	validDirections := map[string]bool{"up": true, "down": true, "stable": true}
	if !validDirections[trends.OccupancyTrend.Direction] {
		t.Errorf("Expected valid direction, got %s", trends.OccupancyTrend.Direction)
	}
}

func TestBuildingAnalyticsService_GenerateOccupancyForecast(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	// Create test building
	building := &models.Building{
		PropertyID:   1,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeResidential,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Test GenerateOccupancyForecast
	forecast, err := service.GenerateOccupancyForecast(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if forecast.NextMonth == nil {
		t.Error("Expected NextMonth forecast to be set")
	}

	if forecast.NextQuarter == nil {
		t.Error("Expected NextQuarter forecast to be set")
	}

	if forecast.NextYear == nil {
		t.Error("Expected NextYear forecast to be set")
	}

	if forecast.Confidence <= 0 || forecast.Confidence > 100 {
		t.Errorf("Expected confidence between 0 and 100, got %f", forecast.Confidence)
	}

	if forecast.ModelUsed == "" {
		t.Error("Expected ModelUsed to be set")
	}

	// Verify forecast data structure
	if forecast.NextMonth.PredictedOccupancyRate < 0 || forecast.NextMonth.PredictedOccupancyRate > 100 {
		t.Errorf("Expected predicted occupancy rate between 0 and 100, got %f", forecast.NextMonth.PredictedOccupancyRate)
	}

	if forecast.NextMonth.ConfidenceInterval.Lower >= forecast.NextMonth.ConfidenceInterval.Upper {
		t.Error("Expected confidence interval lower bound to be less than upper bound")
	}
}

func TestBuildingAnalyticsService_GenerateRevenueProjections(t *testing.T) {
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

	// Test GenerateRevenueProjections
	projections, err := service.GenerateRevenueProjections(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if projections.NextMonth == nil {
		t.Error("Expected NextMonth projection to be set")
	}

	if projections.NextQuarter == nil {
		t.Error("Expected NextQuarter projection to be set")
	}

	if projections.NextYear == nil {
		t.Error("Expected NextYear projection to be set")
	}

	if len(projections.Assumptions) == 0 {
		t.Error("Expected Assumptions to be provided")
	}

	// Verify projection data structure
	if projections.NextMonth.ConservativeRevenue >= projections.NextMonth.ProjectedRevenue {
		t.Error("Expected conservative revenue to be less than projected revenue")
	}

	if projections.NextMonth.ProjectedRevenue >= projections.NextMonth.OptimisticRevenue {
		t.Error("Expected projected revenue to be less than optimistic revenue")
	}

	if projections.NextMonth.Confidence <= 0 || projections.NextMonth.Confidence > 100 {
		t.Errorf("Expected confidence between 0 and 100, got %f", projections.NextMonth.Confidence)
	}
}

func TestBuildingAnalyticsService_GetPropertyBuildingRankings(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	propertyID := 1

	// Create test buildings with different performance characteristics
	buildings := []*models.Building{
		{
			PropertyID:   propertyID,
			BuildingName: "High Performer",
			BuildingCode: "HP001",
			BuildingType: models.BuildingTypeCommercial,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Average Performer",
			BuildingCode: "AP001",
			BuildingType: models.BuildingTypeResidential,
			ActiveStatus: true,
		},
		{
			PropertyID:   propertyID,
			BuildingName: "Low Performer",
			BuildingCode: "LP001",
			BuildingType: models.BuildingTypeMixed,
			ActiveStatus: true,
		},
	}

	for _, building := range buildings {
		_ = mockRepo.Create(building)
	}

	// Test GetPropertyBuildingRankings
	rankings, err := service.GetPropertyBuildingRankings(propertyID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rankings) != len(buildings) {
		t.Errorf("Expected %d rankings, got %d", len(buildings), len(rankings))
	}

	// Verify rankings are properly structured
	for i, ranking := range rankings {
		if ranking.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d", i+1, ranking.Rank)
		}

		if ranking.BuildingID <= 0 {
			t.Error("Expected valid BuildingID")
		}

		if ranking.BuildingName == "" {
			t.Error("Expected BuildingName to be set")
		}

		if ranking.BuildingCode == "" {
			t.Error("Expected BuildingCode to be set")
		}

		if ranking.PerformanceScore < 0 || ranking.PerformanceScore > 100 {
			t.Errorf("Expected performance score between 0 and 100, got %f", ranking.PerformanceScore)
		}
	}

	// Verify rankings are sorted by performance score (descending)
	for i := 1; i < len(rankings); i++ {
		if rankings[i-1].PerformanceScore < rankings[i].PerformanceScore {
			t.Error("Expected rankings to be sorted by performance score in descending order")
		}
	}
}

func TestBuildingAnalyticsService_CacheOperations(t *testing.T) {
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

	// Test cache miss
	_, found := service.GetCachedMetrics(building.ID)
	if found {
		t.Error("Expected cache miss for new building")
	}

	// Get metrics to populate cache
	metrics, err := service.GetBuildingMetrics(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test cache hit
	cachedMetrics, found := service.GetCachedMetrics(building.ID)
	if !found {
		t.Error("Expected cache hit after getting metrics")
	}

	if cachedMetrics.BuildingID != metrics.BuildingID {
		t.Error("Expected cached metrics to match original metrics")
	}

	// Test cache invalidation
	err = service.InvalidateCache(building.ID)
	if err != nil {
		t.Fatalf("Expected no error invalidating cache, got %v", err)
	}

	_, found = service.GetCachedMetrics(building.ID)
	if found {
		t.Error("Expected cache miss after invalidation")
	}

	// Test cache refresh
	err = service.RefreshAnalyticsCache(building.ID)
	if err != nil {
		t.Fatalf("Expected no error refreshing cache, got %v", err)
	}

	_, found = service.GetCachedMetrics(building.ID)
	if found {
		t.Error("Expected cache miss after refresh")
	}
}

func TestBuildingAnalyticsService_ErrorHandling(t *testing.T) {
	service, _ := setupTestAnalyticsService()

	// Test with non-existent building
	_, err := service.GetBuildingMetrics(999)
	if err == nil {
		t.Error("Expected error for non-existent building")
	}

	_, err = service.GetOccupancyAnalytics(999)
	if err == nil {
		t.Error("Expected error for non-existent building")
	}

	_, err = service.GetRevenueAnalytics(999, "month")
	if err == nil {
		t.Error("Expected error for non-existent building")
	}

	_, err = service.CompareBuildingPerformance(999)
	if err == nil {
		t.Error("Expected error for non-existent property")
	}

	// Test with invalid period
	_, err = service.GetRevenueAnalytics(1, "invalid_period")
	if err == nil {
		t.Error("Expected error for invalid period")
	}
}

func TestBuildingAnalyticsService_CompareWithPropertyAverage(t *testing.T) {
	service, mockRepo := setupTestAnalyticsService()

	propertyID := 1

	// Create test building
	building := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Test Building",
		BuildingCode: "TB001",
		BuildingType: models.BuildingTypeCommercial,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building)

	// Create additional buildings for property average calculation
	building2 := &models.Building{
		PropertyID:   propertyID,
		BuildingName: "Building 2",
		BuildingCode: "TB002",
		BuildingType: models.BuildingTypeResidential,
		ActiveStatus: true,
	}
	_ = mockRepo.Create(building2)

	// Test CompareWithPropertyAverage
	comparisons, err := service.CompareWithPropertyAverage(building.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if comparisons.PropertyAverage == nil {
		t.Error("Expected PropertyAverage to be set")
	}

	if comparisons.TypeAverage == nil {
		t.Error("Expected TypeAverage to be set")
	}

	// Verify comparison data structure
	if comparisons.PropertyAverage.OccupancyRate < 0 {
		t.Error("Expected non-negative occupancy rate")
	}

	if comparisons.PropertyAverage.MonthlyRevenue < 0 {
		t.Error("Expected non-negative monthly revenue")
	}
}
