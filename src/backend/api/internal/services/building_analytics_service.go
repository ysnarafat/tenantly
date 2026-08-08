package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// BuildingAnalyticsService implements comprehensive building performance analysis
type BuildingAnalyticsService struct {
	buildingRepo interfaces.BuildingRepositoryInterface
	db           *sql.DB
	cache        map[int]*models.BuildingMetrics
	cacheMutex   sync.RWMutex
	cacheExpiry  map[int]time.Time
}

// NewBuildingAnalyticsService creates a new building analytics service instance
func NewBuildingAnalyticsService(
	buildingRepo interfaces.BuildingRepositoryInterface,
	db *sql.DB,
) *BuildingAnalyticsService {
	return &BuildingAnalyticsService{
		buildingRepo: buildingRepo,
		db:           db,
		cache:        make(map[int]*models.BuildingMetrics),
		cacheExpiry:  make(map[int]time.Time),
	}
}

// GetBuildingMetrics retrieves comprehensive building performance metrics
func (s *BuildingAnalyticsService) GetBuildingMetrics(buildingID int) (*models.BuildingMetrics, error) {
	// Check cache first
	if metrics, found := s.GetCachedMetrics(buildingID); found {
		return metrics, nil
	}

	// Get building information
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	// Get basic analytics
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	// Calculate performance score
	performanceScore, err := s.CalculatePerformanceScore(buildingID)
	if err != nil {
		performanceScore = 0 // Default if calculation fails
	}

	// Generate trend analysis
	trends, err := s.GenerateTrendAnalysis(buildingID, "month")
	if err != nil {
		trends = nil // Optional component
	}

	// Generate comparisons
	comparisons, err := s.CompareWithPropertyAverage(buildingID)
	if err != nil {
		comparisons = nil // Optional component
	}

	metrics := &models.BuildingMetrics{
		BuildingID:       buildingID,
		BuildingName:     building.BuildingName,
		BuildingCode:     building.BuildingCode,
		UnitCount:        analytics.UnitCount,
		OccupiedUnits:    analytics.OccupiedUnits,
		VacantUnits:      analytics.VacantUnits,
		OccupancyRate:    analytics.OccupancyRate,
		MonthlyRevenue:   analytics.MonthlyRevenue,
		AverageRent:      analytics.AverageRent,
		TotalArea:        analytics.TotalArea,
		PerformanceScore: performanceScore,
		Trends:           trends,
		Comparisons:      comparisons,
	}

	// Cache the result
	s.cacheMetrics(buildingID, metrics)

	return metrics, nil
}

// GetOccupancyAnalytics provides vacancy tracking and utilization analysis
func (s *BuildingAnalyticsService) GetOccupancyAnalytics(buildingID int) (*models.OccupancyAnalytics, error) {
	// Get current occupancy snapshot
	currentOccupancy, err := s.getCurrentOccupancySnapshot(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current occupancy: %w", err)
	}

	// Get historical occupancy data (last 12 months)
	historicalOccupancy, err := s.getHistoricalOccupancy(buildingID, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical occupancy: %w", err)
	}

	// Generate vacancy analysis
	vacancyAnalysis, err := s.generateVacancyAnalysis(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate vacancy analysis: %w", err)
	}

	// Calculate utilization metrics
	utilizationMetrics, err := s.calculateUtilizationMetrics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate utilization metrics: %w", err)
	}

	// Generate occupancy forecast
	forecasting, err := s.GenerateOccupancyForecast(buildingID)
	if err != nil {
		forecasting = nil // Optional component
	}

	return &models.OccupancyAnalytics{
		BuildingID:          buildingID,
		CurrentOccupancy:    currentOccupancy,
		HistoricalOccupancy: historicalOccupancy,
		VacancyAnalysis:     vacancyAnalysis,
		UtilizationMetrics:  utilizationMetrics,
		Forecasting:         forecasting,
	}, nil
}

// GetRevenueAnalytics provides financial performance analysis by time periods
func (s *BuildingAnalyticsService) GetRevenueAnalytics(buildingID int, period string) (*models.RevenueAnalytics, error) {
	// Get current period revenue
	currentPeriod, err := s.getRevenuePeriod(buildingID, period, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period revenue: %w", err)
	}

	// Get previous period revenue
	previousPeriod, err := s.getRevenuePeriod(buildingID, period, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous period revenue: %w", err)
	}

	// Get year-to-date revenue
	yearToDate, err := s.getYearToDateRevenue(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get year-to-date revenue: %w", err)
	}

	// Get historical revenue (last 12 periods)
	historicalRevenue, err := s.getHistoricalRevenue(buildingID, period, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical revenue: %w", err)
	}

	// Generate revenue breakdown
	revenueBreakdown, err := s.generateRevenueBreakdown(buildingID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to generate revenue breakdown: %w", err)
	}

	// Calculate performance metrics
	performanceMetrics, err := s.calculateRevenuePerformance(buildingID, currentPeriod, previousPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	// Generate revenue projections
	projections, err := s.GenerateRevenueProjections(buildingID)
	if err != nil {
		projections = nil // Optional component
	}

	return &models.RevenueAnalytics{
		BuildingID:         buildingID,
		Period:             period,
		CurrentPeriod:      currentPeriod,
		PreviousPeriod:     previousPeriod,
		YearToDate:         yearToDate,
		HistoricalRevenue:  historicalRevenue,
		RevenueBreakdown:   revenueBreakdown,
		PerformanceMetrics: performanceMetrics,
		Projections:        projections,
	}, nil
}

// CompareBuildingPerformance provides property-level building comparisons
func (s *BuildingAnalyticsService) CompareBuildingPerformance(propertyID int) (*models.BuildingPerformanceComparison, error) {
	// Get all buildings in the property
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	if len(buildings) == 0 {
		return nil, fmt.Errorf("no buildings found for property")
	}

	// Get property name (assuming we have access to property repo)
	propertyName := fmt.Sprintf("Property %d", propertyID) // Placeholder

	// Generate rankings
	rankings, err := s.GetPropertyBuildingRankings(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building rankings: %w", err)
	}

	// Calculate property averages
	averages, err := s.calculatePropertyAverages(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate property averages: %w", err)
	}

	// Identify top and under performers
	topPerformers, underPerformers := s.identifyPerformers(rankings)

	// Generate insights
	insights := s.generatePerformanceInsights(rankings, averages)

	return &models.BuildingPerformanceComparison{
		PropertyID:       propertyID,
		PropertyName:     propertyName,
		TotalBuildings:   len(buildings),
		ComparisonPeriod: "current_month",
		Rankings:         rankings,
		Averages:         averages,
		TopPerformers:    topPerformers,
		UnderPerformers:  underPerformers,
		Insights:         insights,
	}, nil
}

// CalculatePerformanceScore calculates a comprehensive performance score for a building
func (s *BuildingAnalyticsService) CalculatePerformanceScore(buildingID int) (float64, error) {
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return 0, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Performance score calculation (0-100)
	// Weights: Occupancy Rate (40%), Revenue Performance (30%), Efficiency (30%)

	occupancyScore := analytics.OccupancyRate // Already in percentage

	// Revenue performance (compare to property average)
	revenueScore := 50.0 // Default middle score
	if analytics.UnitCount > 0 {
		revenuePerUnit := analytics.MonthlyRevenue / float64(analytics.UnitCount)
		// Normalize to 0-100 scale (this is simplified)
		revenueScore = math.Min(100, (revenuePerUnit/1000)*100)
	}

	// Efficiency score (based on area utilization if available)
	efficiencyScore := 75.0 // Default score
	if analytics.TotalArea > 0 {
		areaUtilization := float64(analytics.OccupiedUnits) / float64(analytics.UnitCount) * 100
		efficiencyScore = areaUtilization
	}

	// Weighted average
	performanceScore := (occupancyScore*0.4 + revenueScore*0.3 + efficiencyScore*0.3)

	return math.Round(performanceScore*100) / 100, nil
}

// GenerateTrendAnalysis generates trend analysis for building performance
func (s *BuildingAnalyticsService) GenerateTrendAnalysis(buildingID int, period string) (*models.BuildingTrends, error) {
	// Get current and previous period data
	currentAnalytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current analytics: %w", err)
	}

	// Get previous period analytics (simplified - would need historical data)
	previousOccupancy := currentAnalytics.OccupancyRate * 0.95 // Simulated
	previousRevenue := currentAnalytics.MonthlyRevenue * 0.98  // Simulated
	previousRent := currentAnalytics.AverageRent * 0.97        // Simulated

	occupancyTrend := &models.TrendData{
		Current:        currentAnalytics.OccupancyRate,
		Previous:       previousOccupancy,
		ChangePercent:  ((currentAnalytics.OccupancyRate - previousOccupancy) / previousOccupancy) * 100,
		ChangeAbsolute: currentAnalytics.OccupancyRate - previousOccupancy,
		Direction:      s.getTrendDirection(currentAnalytics.OccupancyRate, previousOccupancy),
	}

	revenueTrend := &models.TrendData{
		Current:        currentAnalytics.MonthlyRevenue,
		Previous:       previousRevenue,
		ChangePercent:  ((currentAnalytics.MonthlyRevenue - previousRevenue) / previousRevenue) * 100,
		ChangeAbsolute: currentAnalytics.MonthlyRevenue - previousRevenue,
		Direction:      s.getTrendDirection(currentAnalytics.MonthlyRevenue, previousRevenue),
	}

	rentTrend := &models.TrendData{
		Current:        currentAnalytics.AverageRent,
		Previous:       previousRent,
		ChangePercent:  ((currentAnalytics.AverageRent - previousRent) / previousRent) * 100,
		ChangeAbsolute: currentAnalytics.AverageRent - previousRent,
		Direction:      s.getTrendDirection(currentAnalytics.AverageRent, previousRent),
	}

	return &models.BuildingTrends{
		OccupancyTrend:   occupancyTrend,
		RevenueTrend:     revenueTrend,
		RentTrend:        rentTrend,
		PeriodComparison: period,
	}, nil
}

// GenerateOccupancyForecast generates occupancy forecasting
func (s *BuildingAnalyticsService) GenerateOccupancyForecast(buildingID int) (*models.OccupancyForecast, error) {
	// Get current analytics for baseline
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Simple forecasting model (in production, this would use more sophisticated algorithms)
	baseOccupancy := analytics.OccupancyRate
	baseRevenue := analytics.MonthlyRevenue

	// Apply seasonal and trend adjustments (simplified)
	nextMonth := &models.ForecastData{
		PredictedOccupancyRate: math.Min(100, baseOccupancy*1.02),
		PredictedRevenue:       baseRevenue * 1.01,
	}
	nextMonth.ConfidenceInterval.Lower = nextMonth.PredictedOccupancyRate * 0.95
	nextMonth.ConfidenceInterval.Upper = nextMonth.PredictedOccupancyRate * 1.05

	nextQuarter := &models.ForecastData{
		PredictedOccupancyRate: math.Min(100, baseOccupancy*1.05),
		PredictedRevenue:       baseRevenue * 1.03,
	}
	nextQuarter.ConfidenceInterval.Lower = nextQuarter.PredictedOccupancyRate * 0.90
	nextQuarter.ConfidenceInterval.Upper = nextQuarter.PredictedOccupancyRate * 1.10

	nextYear := &models.ForecastData{
		PredictedOccupancyRate: math.Min(100, baseOccupancy*1.08),
		PredictedRevenue:       baseRevenue * 1.06,
	}
	nextYear.ConfidenceInterval.Lower = nextYear.PredictedOccupancyRate * 0.85
	nextYear.ConfidenceInterval.Upper = nextYear.PredictedOccupancyRate * 1.15

	return &models.OccupancyForecast{
		NextMonth:   nextMonth,
		NextQuarter: nextQuarter,
		NextYear:    nextYear,
		Confidence:  75.0,
		ModelUsed:   "linear_trend_with_seasonal_adjustment",
	}, nil
}

// GenerateRevenueProjections generates revenue forecasting
func (s *BuildingAnalyticsService) GenerateRevenueProjections(buildingID int) (*models.RevenueProjections, error) {
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	baseRevenue := analytics.MonthlyRevenue

	nextMonth := &models.RevenueProjection{
		ProjectedRevenue:    baseRevenue * 1.01,
		ConservativeRevenue: baseRevenue * 0.98,
		OptimisticRevenue:   baseRevenue * 1.05,
		Confidence:          80.0,
	}

	nextQuarter := &models.RevenueProjection{
		ProjectedRevenue:    baseRevenue * 3 * 1.02,
		ConservativeRevenue: baseRevenue * 3 * 0.95,
		OptimisticRevenue:   baseRevenue * 3 * 1.08,
		Confidence:          70.0,
	}

	nextYear := &models.RevenueProjection{
		ProjectedRevenue:    baseRevenue * 12 * 1.05,
		ConservativeRevenue: baseRevenue * 12 * 0.90,
		OptimisticRevenue:   baseRevenue * 12 * 1.15,
		Confidence:          60.0,
	}

	assumptions := []string{
		"Current occupancy rates maintained",
		"No major market disruptions",
		"Seasonal patterns continue",
		"Rent increases follow market trends",
	}

	return &models.RevenueProjections{
		NextMonth:   nextMonth,
		NextQuarter: nextQuarter,
		NextYear:    nextYear,
		Assumptions: assumptions,
	}, nil
}

// CompareWithPropertyAverage compares building with property average
func (s *BuildingAnalyticsService) CompareWithPropertyAverage(buildingID int) (*models.BuildingComparisons, error) {
	// Get building info to find property
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	// Get building analytics
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	// Calculate property averages
	propertyAverages, err := s.calculatePropertyAverages(building.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate property averages: %w", err)
	}

	propertyAverage := &models.ComparisonData{
		OccupancyRate:  propertyAverages.AverageOccupancyRate,
		MonthlyRevenue: propertyAverages.AverageRevenue,
		AverageRent:    propertyAverages.AverageRevenuePerUnit,
		Difference:     analytics.OccupancyRate - propertyAverages.AverageOccupancyRate,
		PerformsBetter: analytics.OccupancyRate > propertyAverages.AverageOccupancyRate,
	}

	// Get type average (simplified - would need more complex calculation)
	typeAverage := &models.ComparisonData{
		OccupancyRate:  85.0,  // Placeholder
		MonthlyRevenue: 50000, // Placeholder
		AverageRent:    2500,  // Placeholder
		Difference:     analytics.OccupancyRate - 85.0,
		PerformsBetter: analytics.OccupancyRate > 85.0,
	}

	return &models.BuildingComparisons{
		PropertyAverage: propertyAverage,
		TypeAverage:     typeAverage,
		TopPerformer:    nil, // Would be populated with actual top performer data
		BottomPerformer: nil, // Would be populated with actual bottom performer data
	}, nil
}

// CompareWithTypeAverage compares building with type average
func (s *BuildingAnalyticsService) CompareWithTypeAverage(buildingID, propertyID int, buildingType models.BuildingType) (*models.BuildingComparisons, error) {
	// This would implement comparison with buildings of the same type
	// For now, returning a simplified implementation
	return s.CompareWithPropertyAverage(buildingID)
}

// GetPropertyBuildingRankings gets building rankings for a property
func (s *BuildingAnalyticsService) GetPropertyBuildingRankings(propertyID int) ([]*models.BuildingRanking, error) {
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	var rankings []*models.BuildingRanking
	for _, building := range buildings {
		analytics, err := s.buildingRepo.GetAnalytics(building.ID)
		if err != nil {
			continue // Skip buildings with analytics errors
		}

		performanceScore, _ := s.CalculatePerformanceScore(building.ID)
		revenuePerUnit := 0.0
		if analytics.UnitCount > 0 {
			revenuePerUnit = analytics.MonthlyRevenue / float64(analytics.UnitCount)
		}

		ranking := &models.BuildingRanking{
			BuildingID:       building.ID,
			BuildingName:     building.BuildingName,
			BuildingCode:     building.BuildingCode,
			PerformanceScore: performanceScore,
			OccupancyRate:    analytics.OccupancyRate,
			MonthlyRevenue:   analytics.MonthlyRevenue,
			RevenuePerUnit:   revenuePerUnit,
		}
		rankings = append(rankings, ranking)
	}

	// Sort by performance score (descending)
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].PerformanceScore > rankings[j].PerformanceScore
	})

	// Assign ranks
	for i, ranking := range rankings {
		ranking.Rank = i + 1
	}

	return rankings, nil
}

// RefreshAnalyticsCache refreshes the cache for a specific building
func (s *BuildingAnalyticsService) RefreshAnalyticsCache(buildingID int) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	delete(s.cache, buildingID)
	delete(s.cacheExpiry, buildingID)

	return nil
}

// GetCachedMetrics retrieves cached metrics if available and not expired
func (s *BuildingAnalyticsService) GetCachedMetrics(buildingID int) (*models.BuildingMetrics, bool) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	metrics, exists := s.cache[buildingID]
	if !exists {
		return nil, false
	}

	expiry, hasExpiry := s.cacheExpiry[buildingID]
	if hasExpiry && time.Now().After(expiry) {
		return nil, false
	}

	return metrics, true
}

// InvalidateCache invalidates the cache for a specific building
func (s *BuildingAnalyticsService) InvalidateCache(buildingID int) error {
	return s.RefreshAnalyticsCache(buildingID)
}

// Helper methods

func (s *BuildingAnalyticsService) cacheMetrics(buildingID int, metrics *models.BuildingMetrics) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache[buildingID] = metrics
	s.cacheExpiry[buildingID] = time.Now().Add(15 * time.Minute) // Cache for 15 minutes
}

func (s *BuildingAnalyticsService) getTrendDirection(current, previous float64) string {
	diff := current - previous
	if math.Abs(diff) < 0.01 { // Less than 1% change
		return "stable"
	}
	if diff > 0 {
		return "up"
	}
	return "down"
}

func (s *BuildingAnalyticsService) getCurrentOccupancySnapshot(buildingID int) (*models.OccupancySnapshot, error) {
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	return &models.OccupancySnapshot{
		Date:          time.Now(),
		TotalUnits:    analytics.UnitCount,
		OccupiedUnits: analytics.OccupiedUnits,
		VacantUnits:   analytics.VacantUnits,
		OccupancyRate: analytics.OccupancyRate,
	}, nil
}

func (s *BuildingAnalyticsService) getHistoricalOccupancy(buildingID int, months int) ([]*models.OccupancySnapshot, error) {
	// This would query historical data from the database
	// For now, returning simulated data
	var snapshots []*models.OccupancySnapshot

	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	for i := months; i > 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		// Simulate some variation in occupancy
		variation := 1.0 + (float64(i%3)-1)*0.05

		snapshot := &models.OccupancySnapshot{
			Date:          date,
			TotalUnits:    analytics.UnitCount,
			OccupiedUnits: int(float64(analytics.OccupiedUnits) * variation),
			VacantUnits:   analytics.UnitCount - int(float64(analytics.OccupiedUnits)*variation),
			OccupancyRate: float64(int(float64(analytics.OccupiedUnits)*variation)) / float64(analytics.UnitCount) * 100,
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (s *BuildingAnalyticsService) generateVacancyAnalysis(buildingID int) (*models.VacancyAnalysis, error) {
	// This would analyze vacancy patterns from historical data
	// For now, returning simulated analysis

	vacantUnitsByType := map[string]int{
		"Shop":      2,
		"Apartment": 1,
		"Office":    0,
	}

	vacantUnitsByFloor := map[int]int{
		1: 1,
		2: 1,
		3: 1,
	}

	seasonalPatterns := []*models.SeasonalVacancy{
		{Month: 1, MonthName: "January", VacancyRate: 15.0, AverageUnits: 3.0},
		{Month: 2, MonthName: "February", VacancyRate: 12.0, AverageUnits: 2.4},
		{Month: 3, MonthName: "March", VacancyRate: 10.0, AverageUnits: 2.0},
		// ... would include all 12 months
	}

	return &models.VacancyAnalysis{
		AverageVacancyDuration: 45.5,
		VacancyTurnoverRate:    8.5,
		SeasonalPatterns:       seasonalPatterns,
		VacantUnitsByType:      vacantUnitsByType,
		VacantUnitsByFloor:     vacantUnitsByFloor,
	}, nil
}

func (s *BuildingAnalyticsService) calculateUtilizationMetrics(buildingID int) (*models.UtilizationMetrics, error) {
	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	spaceUtilization := analytics.OccupancyRate
	revenueUtilization := 85.0 // Placeholder calculation
	efficiencyScore := (spaceUtilization + revenueUtilization) / 2

	utilizationByType := map[string]float64{
		"Shop":      90.0,
		"Apartment": 85.0,
		"Office":    80.0,
	}

	utilizationByFloor := map[int]float64{
		1: 95.0,
		2: 85.0,
		3: 80.0,
	}

	return &models.UtilizationMetrics{
		SpaceUtilization:   spaceUtilization,
		RevenueUtilization: revenueUtilization,
		EfficiencyScore:    efficiencyScore,
		UtilizationByType:  utilizationByType,
		UtilizationByFloor: utilizationByFloor,
	}, nil
}

func (s *BuildingAnalyticsService) getRevenuePeriod(buildingID int, period string, offset int) (*models.RevenuePeriod, error) {
	// This would query actual payment data from the database
	// For now, returning simulated data

	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -offset, 0)
		endDate = startDate.AddDate(0, 1, -1)
	case "quarter":
		quarter := ((int(now.Month()) - 1) / 3) + 1
		startDate = time.Date(now.Year(), time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -offset*3, 0)
		endDate = startDate.AddDate(0, 3, -1)
	case "year":
		startDate = time.Date(now.Year()-offset, 1, 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(1, 0, -1)
	default:
		return nil, fmt.Errorf("unsupported period: %s", period)
	}

	totalRevenue := analytics.MonthlyRevenue
	switch period {
	case "quarter":
		totalRevenue *= 3
	case "year":
		totalRevenue *= 12
	}

	return &models.RevenuePeriod{
		Period:         fmt.Sprintf("%s_%d", period, offset),
		StartDate:      startDate,
		EndDate:        endDate,
		TotalRevenue:   totalRevenue,
		PaidRevenue:    totalRevenue * 0.95,
		PendingRevenue: totalRevenue * 0.03,
		OverdueRevenue: totalRevenue * 0.02,
		CollectionRate: 95.0,
		UnitCount:      analytics.UnitCount,
		AverageRent:    analytics.AverageRent,
	}, nil
}

func (s *BuildingAnalyticsService) getYearToDateRevenue(buildingID int) (*models.RevenuePeriod, error) {
	now := time.Now()
	startDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	monthsElapsed := int(now.Month())
	totalRevenue := analytics.MonthlyRevenue * float64(monthsElapsed)

	return &models.RevenuePeriod{
		Period:         "year_to_date",
		StartDate:      startDate,
		EndDate:        now,
		TotalRevenue:   totalRevenue,
		PaidRevenue:    totalRevenue * 0.95,
		PendingRevenue: totalRevenue * 0.03,
		OverdueRevenue: totalRevenue * 0.02,
		CollectionRate: 95.0,
		UnitCount:      analytics.UnitCount,
		AverageRent:    analytics.AverageRent,
	}, nil
}

func (s *BuildingAnalyticsService) getHistoricalRevenue(buildingID int, period string, periods int) ([]*models.RevenuePeriod, error) {
	var historical []*models.RevenuePeriod

	for i := 0; i < periods; i++ {
		revenuePeriod, err := s.getRevenuePeriod(buildingID, period, i)
		if err != nil {
			continue
		}
		historical = append(historical, revenuePeriod)
	}

	return historical, nil
}

func (s *BuildingAnalyticsService) generateRevenueBreakdown(buildingID int, period string) (*models.RevenueBreakdown, error) {
	// This would analyze actual payment data
	// For now, returning simulated breakdown

	byUnitType := map[string]float64{
		"Shop":      25000,
		"Apartment": 15000,
		"Office":    10000,
	}

	byFloor := map[int]float64{
		1: 20000,
		2: 18000,
		3: 12000,
	}

	bySection := map[string]float64{
		"A": 25000,
		"B": 15000,
		"C": 10000,
	}

	byPaymentType := map[string]float64{
		"Bank Transfer": 35000,
		"Cash":          10000,
		"Check":         5000,
	}

	return &models.RevenueBreakdown{
		ByUnitType:    byUnitType,
		ByFloor:       byFloor,
		BySection:     bySection,
		ByPaymentType: byPaymentType,
	}, nil
}

func (s *BuildingAnalyticsService) calculateRevenuePerformance(buildingID int, current, previous *models.RevenuePeriod) (*models.RevenuePerformance, error) {
	growthRate := 0.0
	if previous != nil && previous.TotalRevenue > 0 {
		growthRate = ((current.TotalRevenue - previous.TotalRevenue) / previous.TotalRevenue) * 100
	}

	analytics, err := s.buildingRepo.GetAnalytics(buildingID)
	if err != nil {
		return nil, err
	}

	revenuePerUnit := 0.0
	if analytics.UnitCount > 0 {
		revenuePerUnit = current.TotalRevenue / float64(analytics.UnitCount)
	}

	revenuePerSqft := 0.0
	if analytics.TotalArea > 0 {
		revenuePerSqft = current.TotalRevenue / analytics.TotalArea
	}

	return &models.RevenuePerformance{
		GrowthRate:           growthRate,
		RevenuePerSqft:       revenuePerSqft,
		RevenuePerUnit:       revenuePerUnit,
		CollectionEfficiency: current.CollectionRate,
		ProfitabilityScore:   75.0, // Placeholder calculation
	}, nil
}

func (s *BuildingAnalyticsService) calculatePropertyAverages(propertyID int) (*models.PropertyAverages, error) {
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, err
	}

	if len(buildings) == 0 {
		return &models.PropertyAverages{}, nil
	}

	var totalOccupancy, totalRevenue, totalRevenuePerUnit, totalPerformance float64
	validBuildings := 0

	for _, building := range buildings {
		analytics, err := s.buildingRepo.GetAnalytics(building.ID)
		if err != nil {
			continue
		}

		totalOccupancy += analytics.OccupancyRate
		totalRevenue += analytics.MonthlyRevenue

		if analytics.UnitCount > 0 {
			totalRevenuePerUnit += analytics.MonthlyRevenue / float64(analytics.UnitCount)
		}

		performanceScore, err := s.CalculatePerformanceScore(building.ID)
		if err == nil {
			totalPerformance += performanceScore
		}

		validBuildings++
	}

	if validBuildings == 0 {
		return &models.PropertyAverages{}, nil
	}

	return &models.PropertyAverages{
		AverageOccupancyRate:    totalOccupancy / float64(validBuildings),
		AverageRevenue:          totalRevenue / float64(validBuildings),
		AverageRevenuePerUnit:   totalRevenuePerUnit / float64(validBuildings),
		AveragePerformanceScore: totalPerformance / float64(validBuildings),
	}, nil
}

func (s *BuildingAnalyticsService) identifyPerformers(rankings []*models.BuildingRanking) ([]*models.BuildingPerformanceSummary, []*models.BuildingPerformanceSummary) {
	var topPerformers, underPerformers []*models.BuildingPerformanceSummary

	// Top 20% are top performers
	topCount := int(math.Max(1, float64(len(rankings))*0.2))
	// Bottom 20% are under performers
	bottomCount := int(math.Max(1, float64(len(rankings))*0.2))

	for i := 0; i < topCount && i < len(rankings); i++ {
		ranking := rankings[i]
		summary := &models.BuildingPerformanceSummary{
			BuildingID:       ranking.BuildingID,
			BuildingName:     ranking.BuildingName,
			BuildingCode:     ranking.BuildingCode,
			BuildingType:     "Mixed", // Placeholder
			PerformanceScore: ranking.PerformanceScore,
		}
		summary.KeyMetrics.OccupancyRate = ranking.OccupancyRate
		summary.KeyMetrics.MonthlyRevenue = ranking.MonthlyRevenue
		summary.KeyMetrics.RevenuePerUnit = ranking.RevenuePerUnit
		summary.KeyMetrics.GrowthRate = 5.0 // Placeholder
		summary.Strengths = []string{"High occupancy", "Strong revenue"}
		summary.Weaknesses = []string{}

		topPerformers = append(topPerformers, summary)
	}

	startIndex := len(rankings) - bottomCount
	for i := startIndex; i < len(rankings); i++ {
		ranking := rankings[i]
		summary := &models.BuildingPerformanceSummary{
			BuildingID:       ranking.BuildingID,
			BuildingName:     ranking.BuildingName,
			BuildingCode:     ranking.BuildingCode,
			BuildingType:     "Mixed", // Placeholder
			PerformanceScore: ranking.PerformanceScore,
		}
		summary.KeyMetrics.OccupancyRate = ranking.OccupancyRate
		summary.KeyMetrics.MonthlyRevenue = ranking.MonthlyRevenue
		summary.KeyMetrics.RevenuePerUnit = ranking.RevenuePerUnit
		summary.KeyMetrics.GrowthRate = -2.0 // Placeholder
		summary.Strengths = []string{}
		summary.Weaknesses = []string{"Low occupancy", "Below average revenue"}

		underPerformers = append(underPerformers, summary)
	}

	return topPerformers, underPerformers
}

func (s *BuildingAnalyticsService) generatePerformanceInsights(rankings []*models.BuildingRanking, averages *models.PropertyAverages) []string {
	var insights []string

	if len(rankings) == 0 {
		return insights
	}

	// Performance spread analysis
	if len(rankings) > 1 {
		spread := rankings[0].PerformanceScore - rankings[len(rankings)-1].PerformanceScore
		if spread > 30 {
			insights = append(insights, "High performance variation across buildings suggests optimization opportunities")
		}
	}

	// Occupancy insights
	if averages.AverageOccupancyRate < 80 {
		insights = append(insights, "Property occupancy below optimal levels - consider marketing initiatives")
	} else if averages.AverageOccupancyRate > 95 {
		insights = append(insights, "Excellent occupancy rates - consider rent optimization")
	}

	// Revenue insights
	highRevenue := 0
	for _, ranking := range rankings {
		if ranking.MonthlyRevenue > averages.AverageRevenue*1.2 {
			highRevenue++
		}
	}

	if highRevenue > len(rankings)/2 {
		insights = append(insights, "Strong revenue performance across majority of buildings")
	}

	return insights
}
