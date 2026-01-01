package models

import (
	"time"
)

// BuildingAnalytics represents building performance metrics
type BuildingAnalytics struct {
	BuildingID     int     `json:"building_id"`
	UnitCount      int     `json:"unit_count"`
	OccupiedUnits  int     `json:"occupied_units"`
	VacantUnits    int     `json:"vacant_units"`
	OccupancyRate  float64 `json:"occupancy_rate"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
	AverageRent    float64 `json:"average_rent"`
	TotalArea      float64 `json:"total_area"`
}

// BuildingMetrics represents comprehensive building performance metrics
type BuildingMetrics struct {
	BuildingID       int                  `json:"building_id"`
	BuildingName     string               `json:"building_name"`
	BuildingCode     string               `json:"building_code"`
	UnitCount        int                  `json:"unit_count"`
	OccupiedUnits    int                  `json:"occupied_units"`
	VacantUnits      int                  `json:"vacant_units"`
	OccupancyRate    float64              `json:"occupancy_rate"`
	MonthlyRevenue   float64              `json:"monthly_revenue"`
	AverageRent      float64              `json:"average_rent"`
	TotalArea        float64              `json:"total_area"`
	PerformanceScore float64              `json:"performance_score"`
	Trends           *BuildingTrends      `json:"trends,omitempty"`
	Comparisons      *BuildingComparisons `json:"comparisons,omitempty"`
}

// BuildingTrends represents trend analysis for building performance
type BuildingTrends struct {
	OccupancyTrend   *TrendData `json:"occupancy_trend"`
	RevenueTrend     *TrendData `json:"revenue_trend"`
	RentTrend        *TrendData `json:"rent_trend"`
	PeriodComparison string     `json:"period_comparison"` // "month", "quarter", "year"
}

// TrendData represents trend information
type TrendData struct {
	Current        float64 `json:"current"`
	Previous       float64 `json:"previous"`
	ChangePercent  float64 `json:"change_percent"`
	ChangeAbsolute float64 `json:"change_absolute"`
	Direction      string  `json:"direction"` // "up", "down", "stable"
}

// BuildingComparisons represents building performance comparisons
type BuildingComparisons struct {
	PropertyAverage *ComparisonData `json:"property_average"`
	TypeAverage     *ComparisonData `json:"type_average"`
	TopPerformer    *ComparisonData `json:"top_performer"`
	BottomPerformer *ComparisonData `json:"bottom_performer"`
}

// ComparisonData represents comparison metrics
type ComparisonData struct {
	OccupancyRate  float64 `json:"occupancy_rate"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
	AverageRent    float64 `json:"average_rent"`
	Difference     float64 `json:"difference"`
	PerformsBetter bool    `json:"performs_better"`
}

// OccupancyAnalytics represents vacancy tracking and utilization analysis
type OccupancyAnalytics struct {
	BuildingID          int                  `json:"building_id"`
	CurrentOccupancy    *OccupancySnapshot   `json:"current_occupancy"`
	HistoricalOccupancy []*OccupancySnapshot `json:"historical_occupancy"`
	VacancyAnalysis     *VacancyAnalysis     `json:"vacancy_analysis"`
	UtilizationMetrics  *UtilizationMetrics  `json:"utilization_metrics"`
	Forecasting         *OccupancyForecast   `json:"forecasting,omitempty"`
}

// OccupancySnapshot represents occupancy at a specific point in time
type OccupancySnapshot struct {
	Date          time.Time `json:"date"`
	TotalUnits    int       `json:"total_units"`
	OccupiedUnits int       `json:"occupied_units"`
	VacantUnits   int       `json:"vacant_units"`
	OccupancyRate float64   `json:"occupancy_rate"`
}

// VacancyAnalysis represents detailed vacancy analysis
type VacancyAnalysis struct {
	AverageVacancyDuration float64            `json:"average_vacancy_duration_days"`
	VacancyTurnoverRate    float64            `json:"vacancy_turnover_rate"`
	SeasonalPatterns       []*SeasonalVacancy `json:"seasonal_patterns"`
	VacantUnitsByType      map[string]int     `json:"vacant_units_by_type"`
	VacantUnitsByFloor     map[int]int        `json:"vacant_units_by_floor"`
}

// SeasonalVacancy represents seasonal vacancy patterns
type SeasonalVacancy struct {
	Month        int     `json:"month"`
	MonthName    string  `json:"month_name"`
	VacancyRate  float64 `json:"vacancy_rate"`
	AverageUnits float64 `json:"average_vacant_units"`
}

// UtilizationMetrics represents building utilization metrics
type UtilizationMetrics struct {
	SpaceUtilization   float64            `json:"space_utilization_percent"`
	RevenueUtilization float64            `json:"revenue_utilization_percent"`
	EfficiencyScore    float64            `json:"efficiency_score"`
	UtilizationByType  map[string]float64 `json:"utilization_by_type"`
	UtilizationByFloor map[int]float64    `json:"utilization_by_floor"`
}

// OccupancyForecast represents occupancy forecasting
type OccupancyForecast struct {
	NextMonth   *ForecastData `json:"next_month"`
	NextQuarter *ForecastData `json:"next_quarter"`
	NextYear    *ForecastData `json:"next_year"`
	Confidence  float64       `json:"confidence_level"`
	ModelUsed   string        `json:"model_used"`
}

// ForecastData represents forecasted metrics
type ForecastData struct {
	PredictedOccupancyRate float64 `json:"predicted_occupancy_rate"`
	PredictedRevenue       float64 `json:"predicted_revenue"`
	ConfidenceInterval     struct {
		Lower float64 `json:"lower"`
		Upper float64 `json:"upper"`
	} `json:"confidence_interval"`
}

// RevenueAnalytics represents financial performance analysis by time periods
type RevenueAnalytics struct {
	BuildingID         int                 `json:"building_id"`
	Period             string              `json:"period"` // "month", "quarter", "year"
	CurrentPeriod      *RevenuePeriod      `json:"current_period"`
	PreviousPeriod     *RevenuePeriod      `json:"previous_period"`
	YearToDate         *RevenuePeriod      `json:"year_to_date"`
	HistoricalRevenue  []*RevenuePeriod    `json:"historical_revenue"`
	RevenueBreakdown   *RevenueBreakdown   `json:"revenue_breakdown"`
	PerformanceMetrics *RevenuePerformance `json:"performance_metrics"`
	Projections        *RevenueProjections `json:"projections,omitempty"`
}

// RevenuePeriod represents revenue for a specific period
type RevenuePeriod struct {
	Period         string    `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	TotalRevenue   float64   `json:"total_revenue"`
	PaidRevenue    float64   `json:"paid_revenue"`
	PendingRevenue float64   `json:"pending_revenue"`
	OverdueRevenue float64   `json:"overdue_revenue"`
	CollectionRate float64   `json:"collection_rate"`
	UnitCount      int       `json:"unit_count"`
	AverageRent    float64   `json:"average_rent"`
}

// RevenueBreakdown represents detailed revenue breakdown
type RevenueBreakdown struct {
	ByUnitType    map[string]float64 `json:"by_unit_type"`
	ByFloor       map[int]float64    `json:"by_floor"`
	BySection     map[string]float64 `json:"by_section"`
	ByPaymentType map[string]float64 `json:"by_payment_type"`
}

// RevenuePerformance represents revenue performance metrics
type RevenuePerformance struct {
	GrowthRate           float64 `json:"growth_rate_percent"`
	RevenuePerSqft       float64 `json:"revenue_per_sqft"`
	RevenuePerUnit       float64 `json:"revenue_per_unit"`
	CollectionEfficiency float64 `json:"collection_efficiency"`
	ProfitabilityScore   float64 `json:"profitability_score"`
}

// RevenueProjections represents revenue forecasting
type RevenueProjections struct {
	NextMonth   *RevenueProjection `json:"next_month"`
	NextQuarter *RevenueProjection `json:"next_quarter"`
	NextYear    *RevenueProjection `json:"next_year"`
	Assumptions []string           `json:"assumptions"`
}

// RevenueProjection represents projected revenue
type RevenueProjection struct {
	ProjectedRevenue    float64 `json:"projected_revenue"`
	ConservativeRevenue float64 `json:"conservative_revenue"`
	OptimisticRevenue   float64 `json:"optimistic_revenue"`
	Confidence          float64 `json:"confidence_level"`
}

// BuildingPerformanceComparison represents property-level building comparisons
type BuildingPerformanceComparison struct {
	PropertyID       int                           `json:"property_id"`
	PropertyName     string                        `json:"property_name"`
	TotalBuildings   int                           `json:"total_buildings"`
	ComparisonPeriod string                        `json:"comparison_period"`
	Rankings         []*BuildingRanking            `json:"rankings"`
	Averages         *PropertyAverages             `json:"averages"`
	TopPerformers    []*BuildingPerformanceSummary `json:"top_performers"`
	UnderPerformers  []*BuildingPerformanceSummary `json:"under_performers"`
	Insights         []string                      `json:"insights"`
}

// BuildingRanking represents building performance ranking
type BuildingRanking struct {
	Rank             int     `json:"rank"`
	BuildingID       int     `json:"building_id"`
	BuildingName     string  `json:"building_name"`
	BuildingCode     string  `json:"building_code"`
	PerformanceScore float64 `json:"performance_score"`
	OccupancyRate    float64 `json:"occupancy_rate"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
	RevenuePerUnit   float64 `json:"revenue_per_unit"`
}

// PropertyAverages represents property-level averages
type PropertyAverages struct {
	AverageOccupancyRate    float64 `json:"average_occupancy_rate"`
	AverageRevenue          float64 `json:"average_revenue"`
	AverageRevenuePerUnit   float64 `json:"average_revenue_per_unit"`
	AveragePerformanceScore float64 `json:"average_performance_score"`
}

// BuildingPerformanceSummary represents summary of building performance
type BuildingPerformanceSummary struct {
	BuildingID       int     `json:"building_id"`
	BuildingName     string  `json:"building_name"`
	BuildingCode     string  `json:"building_code"`
	BuildingType     string  `json:"building_type"`
	PerformanceScore float64 `json:"performance_score"`
	KeyMetrics       struct {
		OccupancyRate  float64 `json:"occupancy_rate"`
		MonthlyRevenue float64 `json:"monthly_revenue"`
		RevenuePerUnit float64 `json:"revenue_per_unit"`
		GrowthRate     float64 `json:"growth_rate"`
	} `json:"key_metrics"`
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
}
