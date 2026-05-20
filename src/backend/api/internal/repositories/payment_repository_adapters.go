package repositories

import (
	"time"

	"github.com/ysnarafat/tenantly/internal/db"
	"github.com/ysnarafat/tenantly/internal/models"
)

// toPayment converts a database row to a Payment model
func toPayment(row db.GetPaymentByIDRow) *models.Payment {
	return &models.Payment{
		ID:             int(row.ID),
		UnitID:         int(row.UnitID),
		TenantID:       int(row.TenantID),
		BuildingID:     int(row.BuildingID),
		PropertyID:     int(row.PropertyID),
		OrganizationID: int(row.OrganizationID),
		Month:          int(row.Month),
		Year:           int(row.Year),
		AmountDue:      row.AmountDue,
		AmountPaid:     row.AmountPaid,
		Status:         models.PaymentStatus(row.Status),
		PaymentMethod:  row.PaymentMethod,
		Notes:          row.Notes,
		ReceiptNumber:  row.ReceiptNumber,
		PaymentDate:    row.PaymentDate,
		DueDate:        row.DueDate,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// toPaymentFromCreate converts a CreatePayment result to a Payment model
func toPaymentFromCreate(row db.CreatePaymentRow) *models.Payment {
	return &models.Payment{
		ID:             int(row.ID),
		UnitID:         int(row.UnitID),
		TenantID:       int(row.TenantID),
		BuildingID:     int(row.BuildingID),
		PropertyID:     int(row.PropertyID),
		OrganizationID: int(row.OrganizationID),
		Month:          int(row.Month),
		Year:           int(row.Year),
		AmountDue:      row.AmountDue,
		AmountPaid:     row.AmountPaid,
		Status:         models.PaymentStatus(row.Status),
		PaymentMethod:  row.PaymentMethod,
		Notes:          row.Notes,
		ReceiptNumber:  row.ReceiptNumber,
		PaymentDate:    row.PaymentDate,
		DueDate:        row.DueDate,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// toPaymentWithDetails converts a database row to a PaymentWithDetails model
func toPaymentWithDetails(row db.GetPaymentByIDWithDetailsRow) *models.PaymentWithDetails {
	return &models.PaymentWithDetails{
		Payment: models.Payment{
			ID:             int(row.ID),
			UnitID:         int(row.UnitID),
			TenantID:       int(row.TenantID),
			BuildingID:     int(row.BuildingID),
			PropertyID:     int(row.PropertyID),
			OrganizationID: int(row.OrganizationID),
			Month:          int(row.Month),
			Year:           int(row.Year),
			AmountDue:      row.AmountDue,
			AmountPaid:     row.AmountPaid,
			Status:         models.PaymentStatus(row.Status),
			PaymentMethod:  row.PaymentMethod,
			Notes:          row.Notes,
			ReceiptNumber:  row.ReceiptNumber,
			PaymentDate:    row.PaymentDate,
			DueDate:        row.DueDate,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		},
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     row.UnitType,
		TenantName:   row.TenantName,
	}
}

// toPaymentWithDetailsFromPeriod converts a building payment period row
func toPaymentWithDetailsFromPeriod(row db.GetBuildingPaymentsInPeriodRow) *models.PaymentWithDetails {
	return &models.PaymentWithDetails{
		Payment: models.Payment{
			ID:             int(row.ID),
			UnitID:         int(row.UnitID),
			TenantID:       int(row.TenantID),
			BuildingID:     int(row.BuildingID),
			PropertyID:     int(row.PropertyID),
			OrganizationID: int(row.OrganizationID),
			Month:          int(row.Month),
			Year:           int(row.Year),
			AmountDue:      row.AmountDue,
			AmountPaid:     row.AmountPaid,
			Status:         models.PaymentStatus(row.Status),
			PaymentMethod:  row.PaymentMethod,
			Notes:          row.Notes,
			ReceiptNumber:  row.ReceiptNumber,
			PaymentDate:    row.PaymentDate,
			DueDate:        row.DueDate,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		},
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     row.UnitType,
		TenantName:   row.TenantName,
	}
}

// toPaymentStats converts database stats row to PaymentStats model
func toPaymentStats(row db.GetBuildingPaymentStatsRow) interface{} {
	totalDue := float64(row.TotalDue)
	totalPaid := float64(row.TotalPaid)
	totalOverdue := float64(row.TotalOverdue)

	stats := &models.PaymentStats{
		TotalRecords: int(row.TotalRecords),
		TotalDue:     totalDue,
		TotalPaid:    totalPaid,
		TotalOverdue: totalOverdue,
		TotalPending: totalDue - totalPaid,
		PaidCount:    int(row.PaidCount),
		DueCount:     int(row.DueCount),
		PartialCount: int(row.PartialCount),
		OverdueCount: int(row.OverdueCount),
	}

	if totalDue > 0 {
		stats.CollectionRate = (totalPaid / totalDue) * 100
	}

	return stats
}

// toDashboardSummary converts dashboard query row to DashboardSummary
func toDashboardSummary(row db.GetDashboardSummaryRow) *models.DashboardSummary {
	totalDue := float64(row.TotalDue)
	totalPaid := float64(row.TotalPaid)

	s := &models.DashboardSummary{
		TotalDue:       totalDue,
		TotalPaid:      totalPaid,
		TotalPending:   float64(row.TotalPending),
		TotalOverdue:   float64(row.TotalOverdue),
		PropertyCount:  int(row.PropertyCount),
		BuildingCount:  int(row.BuildingCount),
		UnitCount:      int(row.UnitCount),
		TenantCount:    int(row.TenantCount),
	}

	if totalDue > 0 {
		s.CollectionRate = (totalPaid / totalDue) * 100
	}

	return s
}

// toBuildingLevelSummary converts to map format
func toBuildingLevelSummary(row db.GetBuildingLevelSummaryRow) map[string]interface{} {
	return map[string]interface{}{
		"total_buildings": row.TotalBuildings,
		"total_due":       row.TotalDue,
		"total_paid":      row.TotalPaid,
	}
}

// toBuildingPaymentAnalytics converts analytics row
func toBuildingAnalytics(row db.GetBuildingPaymentAnalyticsRow, buildingID int, startDate, endDate time.Time) *models.BuildingPaymentAnalytics {
	return &models.BuildingPaymentAnalytics{
		BuildingID:        buildingID,
		Period:            startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalRevenue:      float64(row.TotalRevenue),
		CollectionRate:    row.CollectionRate,
		AveragePaymentTime: row.AvgPaymentDays,
		OverduePayments:   int(row.OverdueCount),
	}
}

// toLeaseSearchResult converts a search result row
func toLeaseSearchResult(row db.SearchLeasesRow) *models.LeaseSearchResult {
	return &models.LeaseSearchResult{
		LeaseID:         int(row.LeaseID),
		TenantID:        int(row.TenantID),
		TenantName:      row.TenantName,
		TenantPhone:     row.TenantPhone,
		PropertyID:      int(row.PropertyID),
		PropertyName:    row.PropertyName,
		BuildingID:      int(row.BuildingID),
		BuildingName:    row.BuildingName,
		BuildingCode:    row.BuildingCode,
		UnitID:          int(row.UnitID),
		UnitNumber:      row.UnitNumber,
		UnitType:        row.UnitType,
		LeaseStartDate:  row.LeaseStartDate,
		LeaseEndDate:    row.LeaseEndDate,
		MonthlyRent:     row.MonthlyRent,
		Active:          row.Active,
	}
}
