package models

import "time"

// TenantReportEntry represents one tenant's summary for the tenant report
type TenantReportEntry struct {
	TenantID     int     `json:"tenant_id"`
	TenantName   string  `json:"tenant_name"`
	PhoneNumber  string  `json:"phone_number"`
	Email        string  `json:"email"`
	UnitNumber   string  `json:"unit_number"`
	BuildingName string  `json:"building_name"`
	PropertyName string  `json:"property_name"`
	LeaseStart   *string `json:"lease_start"`
	LeaseEnd     *string `json:"lease_end"`
	MonthlyRent  float64 `json:"monthly_rent"`
	LeaseActive  bool    `json:"lease_active"`
	TotalDue     float64 `json:"total_due"`
	TotalPaid    float64 `json:"total_paid"`
	BalanceDue   float64 `json:"balance_due"`
}

// TenantSummaryReport aggregates all tenants for an organisation
type TenantSummaryReport struct {
	OrganizationID int                  `json:"organization_id"`
	Tenants        []*TenantReportEntry `json:"tenants"`
	Total          int                  `json:"total"`
	ActiveTenants  int                  `json:"active_tenants"`
	GeneratedAt    time.Time            `json:"generated_at"`
}

// PropertyAnalyticsEntry holds per-property stats in the analytics report
type PropertyAnalyticsEntry struct {
	PropertyID   int         `json:"property_id"`
	PropertyName string      `json:"property_name"`
	PropertyCode string      `json:"property_code"`
	PropertyType string      `json:"property_type"`
	PaymentStats interface{} `json:"payment_stats"`
}

// PropertyAnalyticsReport groups property-level payment analytics for an org
type PropertyAnalyticsReport struct {
	OrganizationID int                       `json:"organization_id"`
	Properties     []*PropertyAnalyticsEntry `json:"properties"`
	Total          int                       `json:"total"`
	ReportPeriod   string                    `json:"report_period"`
	GeneratedAt    time.Time                 `json:"generated_at"`
}

// PaymentAnalyticsResult holds DB-aggregated counts for PaymentAnalysisReport.
// Produced by a repository query; consumed by ReportService.
type PaymentAnalyticsResult struct {
	MethodCounts  map[string]int64
	StatusCounts  map[string]int64
	DailyTrend    map[string]int64
	TotalPayments int64
}

// FinancialLedgerReport complete transaction history with balances
type FinancialLedgerReport struct {
	OrganizationID int                   `json:"organization_id"`
	Payments       []*PaymentWithDetails `json:"payments"`
	Total          int                   `json:"total"`
	TotalDue       int64                 `json:"total_due"`
	TotalPaid      int64                 `json:"total_paid"`
	TotalPending   int64                 `json:"total_pending"`
	TotalOverdue   int64                 `json:"total_overdue"`
	CollectionRate float64               `json:"collection_rate"`
	GeneratedAt    time.Time             `json:"generated_at"`
}

// CollectionSummaryReport returns collection rates, aging analysis, trends
type CollectionSummaryReport struct {
	OrganizationID int                       `json:"organization_id"`
	CollectionRate float64                   `json:"collection_rate"`
	TotalDue       int64                     `json:"total_due"`
	TotalCollected int64                     `json:"total_collected"`
	TotalPending   int64                     `json:"total_pending"`
	TotalOverdue   int64                     `json:"total_overdue"`
	AgingBuckets   map[string]int64          `json:"aging_buckets"`
	MonthlyTrend   []*MonthlyCollectionTrend `json:"monthly_trend"`
	ReportPeriod   string                    `json:"report_period"`
	GeneratedAt    time.Time                 `json:"generated_at"`
}

// MonthlyCollectionTrend represents collection data for a month
type MonthlyCollectionTrend struct {
	Month           time.Time `json:"month"`
	CollectionRate  float64   `json:"collection_rate"`
	AmountDue       int64     `json:"amount_due"`
	AmountCollected int64     `json:"amount_collected"`
}

// PaymentAnalysisReport returns payment analysis data
type PaymentAnalysisReport struct {
	OrganizationID     int              `json:"organization_id"`
	PaymentMethods     map[string]int64 `json:"payment_methods"`
	StatusDistribution map[string]int64 `json:"status_distribution"`
	DailyTrend         map[string]int64 `json:"daily_trend"`
	TotalPayments      int64            `json:"total_payments"`
	ReportPeriod       string           `json:"report_period"`
	GeneratedAt        time.Time        `json:"generated_at"`
}
