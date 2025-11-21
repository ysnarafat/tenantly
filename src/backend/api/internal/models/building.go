package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// BuildingType represents the type of building
type BuildingType string

const (
	BuildingTypeResidential BuildingType = "Residential"
	BuildingTypeCommercial  BuildingType = "Commercial"
	BuildingTypeMixed       BuildingType = "Mixed"
)

// BuildingMetadata stores building-specific attributes
type BuildingMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m BuildingMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *BuildingMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(BuildingMetadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Building represents a physical structure within a property
type Building struct {
	ID               int              `json:"id" db:"id"`
	PropertyID       int              `json:"property_id" db:"property_id"`
	BuildingName     string           `json:"building_name" db:"building_name"`
	BuildingCode     string           `json:"building_code" db:"building_code"`
	BuildingType     BuildingType     `json:"building_type" db:"building_type"`
	TotalFloors      int              `json:"total_floors" db:"total_floors"`
	HasElevator      bool             `json:"has_elevator" db:"has_elevator"`
	ConstructionYear *int             `json:"construction_year" db:"construction_year"`
	Metadata         BuildingMetadata `json:"metadata" db:"metadata"`
	ActiveStatus     bool             `json:"active_status" db:"active_status"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateBuildingRequest represents the request to create a building
type CreateBuildingRequest struct {
	PropertyID       int              `json:"property_id" binding:"required"`
	BuildingName     string           `json:"building_name" binding:"required,max=100"`
	BuildingCode     string           `json:"building_code" binding:"required,max=50"`
	BuildingType     BuildingType     `json:"building_type" binding:"required,oneof=Residential Commercial Mixed"`
	TotalFloors      int              `json:"total_floors" binding:"required,gte=1"`
	HasElevator      bool             `json:"has_elevator"`
	ConstructionYear *int             `json:"construction_year" binding:"omitempty,gte=1800"`
	Metadata         BuildingMetadata `json:"metadata" binding:"omitempty"`
	ActiveStatus     bool             `json:"active_status"`
}

// UpdateBuildingRequest represents the request to update a building
type UpdateBuildingRequest struct {
	BuildingName     *string           `json:"building_name" binding:"omitempty,max=100"`
	BuildingType     *BuildingType     `json:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	TotalFloors      *int              `json:"total_floors" binding:"omitempty,gte=1"`
	HasElevator      *bool             `json:"has_elevator"`
	ConstructionYear *int              `json:"construction_year" binding:"omitempty,gte=1800"`
	Metadata         *BuildingMetadata `json:"metadata" binding:"omitempty"`
	ActiveStatus     *bool             `json:"active_status"`
}

// BuildingWithStats includes building with aggregated statistics
type BuildingWithStats struct {
	Building
	PropertyName  string  `json:"property_name"`
	UnitCount     int     `json:"unit_count"`
	OccupiedUnits int     `json:"occupied_units"`
	TotalRevenue  float64 `json:"total_revenue"`
	OccupancyRate float64 `json:"occupancy_rate"`
}

// PaginationInfo represents common pagination metadata

// BulkCreateBuildingsRequest represents the request to create multiple buildings
type BulkCreateBuildingsRequest struct {
	PropertyID int                     `json:"property_id" binding:"required"`
	Buildings  []CreateBuildingRequest `json:"buildings" binding:"required,min=1"`
}

// BuildingSearchFilters represents filters for building search
type BuildingSearchFilters struct {
	PropertyID   *int          `json:"property_id"`
	BuildingType *BuildingType `json:"building_type"`
	ActiveStatus *bool         `json:"active_status"`
	HasElevator  *bool         `json:"has_elevator"`
	MinFloors    *int          `json:"min_floors"`
	MaxFloors    *int          `json:"max_floors"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
}

// PropertyStatistics represents aggregated statistics for a property

// PropertyWithBuildings represents a property with its buildings

// BuildingAnalytics represents detailed analytics for a building
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

// Building validation error constants
const (
	ErrBuildingNotFound          = "BUILDING_NOT_FOUND"
	ErrBuildingCodeExists        = "BUILDING_CODE_EXISTS"
	ErrInvalidBuildingType       = "INVALID_BUILDING_TYPE"
	ErrInvalidMetadata           = "INVALID_METADATA"
	ErrPropertyNotFound          = "PROPERTY_NOT_FOUND"
	ErrBuildingHasActiveUnits    = "BUILDING_HAS_ACTIVE_UNITS"
	ErrInvalidFloorCount         = "INVALID_FLOOR_COUNT"
	ErrInvalidConstructionYear   = "INVALID_CONSTRUCTION_YEAR"
	ErrMetadataValidationFailed  = "METADATA_VALIDATION_FAILED"
	ErrInvalidAmenities          = "INVALID_AMENITIES"
	ErrInvalidSecurityType       = "INVALID_SECURITY_TYPE"
	ErrInvalidParkingSpaces      = "INVALID_PARKING_SPACES"
	ErrInvalidBusinessHours      = "INVALID_BUSINESS_HOURS"
	ErrInvalidFacilities         = "INVALID_FACILITIES"
	ErrInvalidResidentialSection = "INVALID_RESIDENTIAL_SECTION"
	ErrInvalidCommercialSection  = "INVALID_COMMERCIAL_SECTION"
	ErrInvalidSharedFacilities   = "INVALID_SHARED_FACILITIES"
)

// BuildingError represents a building-specific error
type BuildingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *BuildingError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewBuildingError creates a new BuildingError
func NewBuildingError(code, message, field string) *BuildingError {
	return &BuildingError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// BuildingSearchRequest represents advanced search request for buildings
type BuildingSearchRequest struct {
	PropertyID       *int   `form:"property_id"`
	BuildingType     string `form:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	ActiveStatus     *bool  `form:"active_status"`
	HasElevator      *bool  `form:"has_elevator"`
	MinFloors        *int   `form:"min_floors" binding:"omitempty,gte=1"`
	MaxFloors        *int   `form:"max_floors" binding:"omitempty,gte=1"`
	ConstructionYear *int   `form:"construction_year" binding:"omitempty,gte=1800"`
	SearchTerm       string `form:"search_term"`
	MetadataQuery    string `form:"metadata_query"`
	Page             int    `form:"page" binding:"omitempty,gte=1"`
	PageSize         int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	SortBy           string `form:"sort_by" binding:"omitempty,oneof=building_name building_code created_at updated_at total_floors construction_year"`
	SortOrder        string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	IncludeStats     bool   `form:"include_stats"`
}

// BuildingExportRequest represents request for building data export
type BuildingExportRequest struct {
	PropertyID   *int   `form:"property_id"`
	BuildingType string `form:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	ActiveStatus *bool  `form:"active_status"`
	Format       string `form:"format" binding:"omitempty,oneof=csv json xlsx"`
	IncludeStats bool   `form:"include_stats"`
}

// BuildingStatusRequest represents request for building status management
type BuildingStatusRequest struct {
	ActiveStatus bool   `json:"active_status" binding:"required"`
	Reason       string `json:"reason" binding:"omitempty,max=255"`
}

// MetadataSchemaResponse represents metadata schema information for building types
type MetadataSchemaResponse struct {
	BuildingType string                 `json:"building_type"`
	Schema       map[string]interface{} `json:"schema"`
	Examples     map[string]interface{} `json:"examples"`
	Description  string                 `json:"description"`
}

// BuildingUnitSummary represents unit summary for a building
type BuildingUnitSummary struct {
	UnitID      int     `json:"unit_id"`
	UnitNumber  string  `json:"unit_number"`
	UnitName    string  `json:"unit_name"`
	Floor       int     `json:"floor"`
	Section     string  `json:"section"`
	UnitType    string  `json:"unit_type"`
	MonthlyRent float64 `json:"monthly_rent"`
	Active      bool    `json:"active"`
	TenantName  string  `json:"tenant_name,omitempty"`
	LeaseActive bool    `json:"lease_active"`
}

// BuildingUnitsResponse represents response for building units endpoint
type BuildingUnitsResponse struct {
	BuildingID   int                    `json:"building_id"`
	BuildingName string                 `json:"building_name"`
	BuildingCode string                 `json:"building_code"`
	Units        []*BuildingUnitSummary `json:"units"`
	Summary      struct {
		TotalUnits    int     `json:"total_units"`
		OccupiedUnits int     `json:"occupied_units"`
		VacantUnits   int     `json:"vacant_units"`
		TotalRevenue  float64 `json:"total_revenue"`
	} `json:"summary"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
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
