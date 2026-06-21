package models

import "time"

// BuildingPaymentReport represents payment report for a building
type BuildingPaymentReport struct {
	BuildingID   int                   `json:"building_id"`
	BuildingName string                `json:"building_name"`
	BuildingCode string                `json:"building_code"`
	PropertyID   int                   `json:"property_id"`
	PropertyName string                `json:"property_name"`
	ReportPeriod string                `json:"report_period"`
	Statistics   any                   `json:"statistics"`
	Payments     []*PaymentWithDetails `json:"payments"`
	GeneratedAt  time.Time             `json:"generated_at"`
}

// PropertyPaymentReport represents payment report for a property with building breakdowns
type PropertyPaymentReport struct {
	PropertyID         int                         `json:"property_id"`
	PropertyName       string                      `json:"property_name"`
	PropertyCode       string                      `json:"property_code"`
	ReportPeriod       string                      `json:"report_period"`
	OverallStatistics  any                         `json:"overall_statistics"`
	BuildingBreakdowns []*BuildingPaymentBreakdown `json:"building_breakdowns"`
	GeneratedAt        time.Time                   `json:"generated_at"`
}

// BuildingPaymentBreakdown represents payment breakdown for a building
type BuildingPaymentBreakdown struct {
	BuildingID   int         `json:"building_id"`
	BuildingName string      `json:"building_name"`
	BuildingCode string      `json:"building_code"`
	BuildingType string      `json:"building_type"`
	Statistics   any `json:"statistics"`
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
	TrendAnalysis      any `json:"trend_analysis"`
	ComparisonMetrics  any `json:"comparison_metrics"`
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
	Statistics         any                      `json:"statistics"`
	BuildingBreakdowns []*BuildingNotificationBreakdown `json:"building_breakdowns,omitempty"`
	GeneratedAt        time.Time                        `json:"generated_at"`
}

// BuildingNotificationBreakdown represents notification breakdown for a building
type BuildingNotificationBreakdown struct {
	BuildingID   int         `json:"building_id"`
	BuildingName string      `json:"building_name"`
	BuildingCode string      `json:"building_code"`
	Statistics   any `json:"statistics"`
}

// Enhanced DashboardSummary with building context
type EnhancedDashboardSummary struct {
	DashboardSummary
	AvgRevenuePerBuilding   float64                `json:"avg_revenue_per_building"`
	TopPerformingBuilding   map[string]any `json:"top_performing_building"`
	BuildingTypeBreakdown   map[string]any `json:"building_type_breakdown"`
	OccupancyByBuildingType map[string]float64     `json:"occupancy_by_building_type"`
}
