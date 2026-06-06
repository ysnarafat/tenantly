package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type ReportService struct {
	paymentRepo interfaces.PaymentRepositoryInterface
}

func NewReportService(paymentRepo interfaces.PaymentRepositoryInterface) *ReportService {
	return &ReportService{
		paymentRepo: paymentRepo,
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

	agingBuckets := map[string]int64{
		"current": 0,
		"30d":     0,
		"60d":     0,
		"90d+":    0,
	}

	monthlyTrend := make([]*models.MonthlyCollectionTrend, 0, 6)
	now := time.Now()
	for i := 5; i >= 0; i-- {
		month := now.AddDate(0, -i, 0)
		monthlyTrend = append(monthlyTrend, &models.MonthlyCollectionTrend{
			Month:            month,
			CollectionRate:   summary.CollectionRate,
			AmountDue:        int64(summary.TotalDue),
			AmountCollected:  int64(summary.TotalPaid),
		})
	}

	return &models.CollectionSummaryReport{
		OrganizationID:    orgID,
		CollectionRate:    summary.CollectionRate,
		TotalDue:          int64(summary.TotalDue),
		TotalCollected:    int64(summary.TotalPaid),
		TotalPending:      int64(summary.TotalPending),
		TotalOverdue:      int64(summary.TotalOverdue),
		AgingBuckets:      agingBuckets,
		MonthlyTrend:      monthlyTrend,
		ReportPeriod:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt:       time.Now(),
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
