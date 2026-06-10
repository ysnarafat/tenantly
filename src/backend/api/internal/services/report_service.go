package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type ReportService struct {
	paymentRepo  interfaces.PaymentRepositoryInterface
	propertyRepo interfaces.PropertyRepositoryInterface
}

func NewReportService(paymentRepo interfaces.PaymentRepositoryInterface, propertyRepo interfaces.PropertyRepositoryInterface) *ReportService {
	return &ReportService{
		paymentRepo:  paymentRepo,
		propertyRepo: propertyRepo,
	}
}

// FinancialLedgerReport returns complete transaction history with balances
func (s *ReportService) FinancialLedgerReport(orgID int, filters map[string]interface{}, limit, offset int) (*models.FinancialLedgerReport, error) {
	filters["organization_id"] = orgID

	payments, total, err := s.paymentRepo.GetWithDetailsAndFilters(filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment ledger: %w", err)
	}

	summary, err := s.paymentRepo.GetDashboardSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	return &models.FinancialLedgerReport{
		OrganizationID: orgID,
		Payments:       payments,
		Total:          total,
		TotalDue:       int64(summary.TotalDue),
		TotalPaid:      int64(summary.TotalPaid),
		TotalPending:   int64(summary.TotalPending),
		TotalOverdue:   int64(summary.TotalOverdue),
		CollectionRate: summary.CollectionRate,
		GeneratedAt:    time.Now(),
	}, nil
}

// CollectionSummaryReport returns collection rates and payment summary
func (s *ReportService) CollectionSummaryReport(orgID int, startDate, endDate time.Time) (*models.CollectionSummaryReport, error) {
	summary, err := s.paymentRepo.GetDashboardSummary()
	if err != nil {
		return nil, fmt.Errorf("failed to get collection summary: %w", err)
	}

	agingBuckets, err := s.paymentRepo.GetAgingBuckets(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get aging buckets: %w", err)
	}

	monthlyTrend, err := s.paymentRepo.GetMonthlyCollectionTrend(orgID, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly trend: %w", err)
	}

	return &models.CollectionSummaryReport{
		OrganizationID: orgID,
		CollectionRate: summary.CollectionRate,
		TotalDue:       int64(summary.TotalDue),
		TotalCollected: int64(summary.TotalPaid),
		TotalPending:   int64(summary.TotalPending),
		TotalOverdue:   int64(summary.TotalOverdue),
		AgingBuckets:   agingBuckets,
		MonthlyTrend:   monthlyTrend,
		ReportPeriod:   fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt:    time.Now(),
	}, nil
}

// TenantSummaryReport returns all tenants with their lease and payment totals
func (s *ReportService) TenantSummaryReport(orgID int) (*models.TenantSummaryReport, error) {
	entries, err := s.paymentRepo.GetTenantPaymentSummary(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant summary: %w", err)
	}

	active := 0
	for _, e := range entries {
		if e.LeaseActive {
			active++
		}
	}

	return &models.TenantSummaryReport{
		OrganizationID: orgID,
		Tenants:        entries,
		Total:          len(entries),
		ActiveTenants:  active,
		GeneratedAt:    time.Now(),
	}, nil
}

// PropertyAnalyticsReport returns per-property payment analytics for the organisation
func (s *ReportService) PropertyAnalyticsReport(orgID int, startDate, endDate time.Time) (*models.PropertyAnalyticsReport, error) {
	properties, _, err := s.propertyRepo.List(map[string]interface{}{
		"organization_id": orgID,
		"active":          true,
	}, 200, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list properties: %w", err)
	}

	entries := make([]*models.PropertyAnalyticsEntry, 0, len(properties))
	for _, p := range properties {
		stats, err := s.paymentRepo.GetPropertyPaymentStats(p.ID, startDate, endDate)
		if err != nil {
			stats = nil
		}
		entries = append(entries, &models.PropertyAnalyticsEntry{
			PropertyID:   p.ID,
			PropertyName: p.PropertyName,
			PropertyCode: p.PropertyCode,
			PropertyType: string(p.PropertyType),
			PaymentStats: stats,
		})
	}

	return &models.PropertyAnalyticsReport{
		OrganizationID: orgID,
		Properties:     entries,
		Total:          len(entries),
		ReportPeriod:   fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt:    time.Now(),
	}, nil
}

// PaymentAnalysisReport returns payment methods and status distribution
func (s *ReportService) PaymentAnalysisReport(orgID int, startDate, endDate time.Time) (*models.PaymentAnalysisReport, error) {
	filters := map[string]interface{}{
		"organization_id": orgID,
	}

	payments, _, err := s.paymentRepo.GetWithDetailsAndFilters(filters, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payments: %w", err)
	}

	methodCounts := make(map[string]int)
	statusCounts := make(map[string]int64)
	dailyTrend := make(map[string]int64)

	if payments != nil {
		for _, p := range payments {
			if p.CreatedAt.After(startDate) && p.CreatedAt.Before(endDate) {
				method := "cash"
				if p.PaymentMethod != "" {
					method = p.PaymentMethod
				}
				methodCounts[method]++
				statusCounts[string(p.Status)]++
				day := p.CreatedAt.Format("2006-01-02")
				dailyTrend[day]++
			}
		}
	}

	return &models.PaymentAnalysisReport{
		OrganizationID:     orgID,
		PaymentMethods:     methodCounts,
		StatusDistribution: statusCounts,
		DailyTrend:         dailyTrend,
		TotalPayments:      int64(len(payments)),
		ReportPeriod:       fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt:        time.Now(),
	}, nil
}
