package repositories

import "github.com/ysnarafat/tenantly/internal/models"

// buildPaymentStats assembles a PaymentStats from aggregate query results.
func buildPaymentStats(totalRecords, paidCount, dueCount, partialCount, overdueCount int64, totalDue, totalPaid, totalOverdue float64) *models.PaymentStats {
	stats := &models.PaymentStats{
		TotalRecords: int(totalRecords),
		TotalDue:     totalDue,
		TotalPaid:    totalPaid,
		TotalOverdue: totalOverdue,
		TotalPending: totalDue - totalPaid,
		PaidCount:    int(paidCount),
		DueCount:     int(dueCount),
		PartialCount: int(partialCount),
		OverdueCount: int(overdueCount),
	}
	if totalDue > 0 {
		stats.CollectionRate = (totalPaid / totalDue) * 100
	}
	return stats
}
