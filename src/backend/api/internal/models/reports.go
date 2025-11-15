package models

import "time"

// ComprehensiveReport represents a comprehensive report with building-level breakdowns
type ComprehensiveReport struct {
	ReportType         string                     `json:"report_type"` // "building", "property", "system"
	BuildingID         *int                       `json:"building_id,omitempty"`
	BuildingName       *string                    `json:"building_name,omitempty"`
	BuildingCode       *string                    `json:"building_code,omitempty"`
	PropertyID         *int                       `json:"property_id,omitempty"`
	PropertyName       *string                    `json:"property_name,omitempty"`
	PropertyCode       *string                    `json:"property_code,omitempty"`
	StartDate          time.Time                  `json:"start_date"`
	EndDate            time.Time                  `json:"end_date"`
	BuildingAnalytics  *BuildingAnalytics         `json:"building_analytics,omitempty"`
	PaymentStats       interface{}                `json:"payment_stats"`
	OccupancyStats     interface{}                `json:"occupancy_stats"`
	NotificationStats  interface{}                `json:"notification_stats"`
	BuildingBreakdowns []*BuildingReportBreakdown `json:"building_breakdowns,omitempty"`
	PropertySummaries  []*PropertyReportSummary   `json:"property_summaries,omitempty"`
	PropertyStatistics interface{}                `json:"property_statistics,omitempty"`
	SystemStatistics   interface{}                `json:"system_statistics,omitempty"`
	GeneratedAt        time.Time                  `json:"generated_at"`
}

// BuildingReportBreakdown represents detailed breakdown for a specific building
type BuildingReportBreakdown struct {
	BuildingID           int                `json:"building_id"`
	BuildingName         string             `json:"building_name"`
	BuildingCode         string             `json:"building_code"`
	BuildingType         string             `json:"building_type"`
	TotalFloors          int                `json:"total_floors"`
	HasElevator          bool               `json:"has_elevator"`
	ConstructionYear     *int               `json:"construction_year"`
	Analytics            *BuildingAnalytics `json:"analytics"`
	PaymentStats         interface{}        `json:"payment_stats"`
	OccupancyStats       interface{}        `json:"occupancy_stats"`
	NotificationStats    interface{}        `json:"notification_stats"`
	UnitTypeDistribution interface{}        `json:"unit_type_distribution"`
	PerformanceScore     float64            `json:"performance_score"`
}

// PropertyReportSummary represents summary for a specific property
type PropertyReportSummary struct {
	PropertyID       int         `json:"property_id"`
	PropertyName     string      `json:"property_name"`
	PropertyCode     string      `json:"property_code"`
	PropertyType     string      `json:"property_type"`
	BuildingCount    int         `json:"building_count"`
	UnitCount        int         `json:"unit_count"`
	BuildingStats    interface{} `json:"building_stats"`
	PaymentStats     interface{} `json:"payment_stats"`
	OccupancyStats   interface{} `json:"occupancy_stats"`
	PerformanceScore float64     `json:"performance_score"`
}

// DashboardReport represents dashboard report with building filtering and grouping
type DashboardReport struct {
	SystemStatistics interface{}            `json:"system_statistics"`
	Groupings        interface{}            `json:"groupings"`
	GroupBy          string                 `json:"group_by"`
	Filters          map[string]interface{} `json:"filters"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// PropertyGrouping represents property-level grouping with building context
type PropertyGrouping struct {
	PropertyID     int         `json:"property_id"`
	PropertyName   string      `json:"property_name"`
	PropertyCode   string      `json:"property_code"`
	PropertyType   string      `json:"property_type"`
	BuildingCount  int         `json:"building_count"`
	UnitCount      int         `json:"unit_count"`
	BuildingStats  interface{} `json:"building_stats"`
	PaymentStats   interface{} `json:"payment_stats"`
	OccupancyStats interface{} `json:"occupancy_stats"`
}

// BuildingGrouping represents building-level grouping
type BuildingGrouping struct {
	BuildingID     int                `json:"building_id"`
	BuildingName   string             `json:"building_name"`
	BuildingCode   string             `json:"building_code"`
	BuildingType   string             `json:"building_type"`
	PropertyID     int                `json:"property_id"`
	Analytics      *BuildingAnalytics `json:"analytics"`
	PaymentStats   interface{}        `json:"payment_stats"`
	OccupancyStats interface{}        `json:"occupancy_stats"`
}

// BuildingTypeGrouping represents building type-based grouping
type BuildingTypeGrouping struct {
	BuildingType string      `json:"building_type"`
	Statistics   interface{} `json:"statistics"`
}

// BuildingPaymentReport represents payment report for a building
type BuildingPaymentReport struct {
	BuildingID   int                   `json:"building_id"`
	BuildingName string                `json:"building_name"`
	BuildingCode string                `json:"building_code"`
	PropertyID   int                   `json:"property_id"`
	PropertyName string                `json:"property_name"`
	ReportPeriod string                `json:"report_period"`
	Statistics   interface{}           `json:"statistics"`
	Payments     []*PaymentWithDetails `json:"payments"`
	GeneratedAt  time.Time             `json:"generated_at"`
}

// PropertyPaymentReport represents payment report for a property with building breakdowns
type PropertyPaymentReport struct {
	PropertyID         int                         `json:"property_id"`
	PropertyName       string                      `json:"property_name"`
	PropertyCode       string                      `json:"property_code"`
	ReportPeriod       string                      `json:"report_period"`
	OverallStatistics  interface{}                 `json:"overall_statistics"`
	BuildingBreakdowns []*BuildingPaymentBreakdown `json:"building_breakdowns"`
	GeneratedAt        time.Time                   `json:"generated_at"`
}

// BuildingPaymentBreakdown represents payment breakdown for a building
type BuildingPaymentBreakdown struct {
	BuildingID   int         `json:"building_id"`
	BuildingName string      `json:"building_name"`
	BuildingCode string      `json:"building_code"`
	BuildingType string      `json:"building_type"`
	Statistics   interface{} `json:"statistics"`
}

// BuildingPaymentAnalytics represents payment analytics for a building
type BuildingPaymentAnalytics struct {
	BuildingID         int         `json:"building_id"`
	BuildingName       string      `json:"building_name"`
	BuildingCode       string      `json:"building_code"`
	BuildingType       string      `json:"building_type"`
	Period             string      `json:"period"`
	TotalRevenue       float64     `json:"total_revenue"`
	CollectionRate     float64     `json:"collection_rate"`
	AveragePaymentTime float64     `json:"average_payment_time"`
	OverduePayments    int         `json:"overdue_payments"`
	TrendAnalysis      interface{} `json:"trend_analysis"`
	ComparisonMetrics  interface{} `json:"comparison_metrics"`
}

// NotificationReport represents notification report with building context
type NotificationReport struct {
	ReportType         string                           `json:"report_type"` // "building", "property", "system"
	BuildingID         *int                             `json:"building_id,omitempty"`
	BuildingName       *string                          `json:"building_name,omitempty"`
	PropertyID         *int                             `json:"property_id,omitempty"`
	PropertyName       *string                          `json:"property_name,omitempty"`
	StartDate          time.Time                        `json:"start_date"`
	EndDate            time.Time                        `json:"end_date"`
	Statistics         interface{}                      `json:"statistics"`
	BuildingBreakdowns []*BuildingNotificationBreakdown `json:"building_breakdowns,omitempty"`
	GeneratedAt        time.Time                        `json:"generated_at"`
}

// BuildingNotificationBreakdown represents notification breakdown for a building
type BuildingNotificationBreakdown struct {
	BuildingID   int         `json:"building_id"`
	BuildingName string      `json:"building_name"`
	BuildingCode string      `json:"building_code"`
	Statistics   interface{} `json:"statistics"`
}

// Enhanced DashboardSummary with building context
type EnhancedDashboardSummary struct {
	DashboardSummary
	AvgRevenuePerBuilding   float64                `json:"avg_revenue_per_building"`
	TopPerformingBuilding   map[string]interface{} `json:"top_performing_building"`
	BuildingTypeBreakdown   map[string]interface{} `json:"building_type_breakdown"`
	OccupancyByBuildingType map[string]float64     `json:"occupancy_by_building_type"`
}
