package repositories

import (
	"strconv"
	"time"

	"github.com/ysnarafat/tenantly/internal/db"
	"github.com/ysnarafat/tenantly/internal/models"
)

// toPayment converts a database row to a Payment model
func toPayment(row db.GetPaymentByIDRow) *models.Payment {
	amountPaid, _ := strconv.ParseFloat(row.AmountPaid.String, 64)
	paymentDate := (*time.Time)(nil)
	if row.PaymentDate.Valid {
		paymentDate = &row.PaymentDate.Time
	}
	dueDate := (*time.Time)(nil)
	if row.DueDate.Valid {
		dueDate = &row.DueDate.Time
	}

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
		AmountPaid:     amountPaid,
		Status:         models.PaymentStatus(row.Status.String),
		PaymentMethod:  row.PaymentMethod,
		Notes:          row.Notes,
		ReceiptNumber:  row.ReceiptNumber,
		PaymentDate:    paymentDate,
		DueDate:        dueDate,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}

// toPaymentFromCreate converts a CreatePayment result to a Payment model
func toPaymentFromCreate(row db.CreatePaymentRow) *models.Payment {
	amountPaid, _ := strconv.ParseFloat(row.AmountPaid.String, 64)

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
		AmountPaid:     amountPaid,
		Status:         models.PaymentStatus(row.Status.String),
		PaymentMethod:  row.PaymentMethod,
		Notes:          row.Notes,
		ReceiptNumber:  row.ReceiptNumber,
		PaymentDate:    nil,
		DueDate:        nil,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}

// toPaymentWithDetails converts a database row to a PaymentWithDetails model
func toPaymentWithDetails(row db.GetPaymentByIDWithDetailsRow) *models.PaymentWithDetails {
	amountPaid, _ := strconv.ParseFloat(row.AmountPaid.String, 64)
	paymentDate := (*time.Time)(nil)
	if row.PaymentDate.Valid {
		paymentDate = &row.PaymentDate.Time
	}
	dueDate := (*time.Time)(nil)
	if row.DueDate.Valid {
		dueDate = &row.DueDate.Time
	}

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
			AmountPaid:     amountPaid,
			Status:         models.PaymentStatus(row.Status.String),
			PaymentMethod:  row.PaymentMethod,
			Notes:          row.Notes,
			ReceiptNumber:  row.ReceiptNumber,
			PaymentDate:    paymentDate,
			DueDate:        dueDate,
			CreatedAt:      row.CreatedAt.Time,
			UpdatedAt:      row.UpdatedAt.Time,
		},
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     row.UnitType.(string),
		TenantName:   row.TenantName,
	}
}

// toPaymentWithDetailsFromPeriod converts a building payment period row
func toPaymentWithDetailsFromPeriod(row db.GetBuildingPaymentsInPeriodRow) *models.PaymentWithDetails {
	amountPaid, _ := strconv.ParseFloat(row.AmountPaid.String, 64)
	paymentDate := (*time.Time)(nil)
	if row.PaymentDate.Valid {
		paymentDate = &row.PaymentDate.Time
	}
	dueDate := (*time.Time)(nil)
	if row.DueDate.Valid {
		dueDate = &row.DueDate.Time
	}

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
			AmountPaid:     amountPaid,
			Status:         models.PaymentStatus(row.Status.String),
			PaymentMethod:  row.PaymentMethod,
			Notes:          row.Notes,
			ReceiptNumber:  row.ReceiptNumber,
			PaymentDate:    paymentDate,
			DueDate:        dueDate,
			CreatedAt:      row.CreatedAt.Time,
			UpdatedAt:      row.UpdatedAt.Time,
		},
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     row.UnitType.(string),
		TenantName:   row.TenantName,
	}
}

// toPaymentStats converts database stats row to PaymentStats model
func toPaymentStats(row db.GetBuildingPaymentStatsRow) interface{} {
	totalDue, _ := row.TotalDue.(float64)
	totalPaid, _ := row.TotalPaid.(float64)
	totalOverdue, _ := row.TotalOverdue.(float64)

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

// toPropertyPaymentStats converts property payment stats row to PaymentStats model
func toPropertyPaymentStats(row db.GetPropertyPaymentStatsRow) interface{} {
	totalDue, _ := row.TotalDue.(float64)
	totalPaid, _ := row.TotalPaid.(float64)
	totalOverdue, _ := row.TotalOverdue.(float64)

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

// toSystemPaymentStats converts system payment stats row to PaymentStats model
func toSystemPaymentStats(row db.GetSystemPaymentStatsRow) interface{} {
	totalDue, _ := row.TotalDue.(float64)
	totalPaid, _ := row.TotalPaid.(float64)
	totalOverdue, _ := row.TotalOverdue.(float64)

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
	totalDue, _ := row.TotalDue.(float64)
	totalPaid, _ := row.TotalPaid.(float64)
	totalPending, _ := row.TotalPending.(float64)
	totalOverdue, _ := row.TotalOverdue.(float64)

	s := &models.DashboardSummary{
		TotalDue:      totalDue,
		TotalPaid:     totalPaid,
		TotalPending:  totalPending,
		TotalOverdue:  totalOverdue,
		PropertyCount: int(row.PropertyCount),
		BuildingCount: int(row.BuildingCount),
		UnitCount:     int(row.UnitCount),
		TenantCount:   int(row.TenantCount),
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
	totalRevenue, _ := row.TotalRevenue.(float64)
	avgPaymentDays, _ := row.AvgPaymentDays.(float64)

	return &models.BuildingPaymentAnalytics{
		BuildingID:         buildingID,
		Period:             startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalRevenue:       totalRevenue,
		CollectionRate:     float64(row.CollectionRate),
		AveragePaymentTime: avgPaymentDays,
		OverduePayments:    int(row.OverdueCount),
	}
}

// toLeaseSearchResult converts a search result row
func toLeaseSearchResult(row db.SearchLeasesRow) *models.LeaseSearchResult {
	leaseEndDate := time.Time{}
	if row.LeaseEndDate.Valid {
		leaseEndDate = row.LeaseEndDate.Time
	}
	active := false
	if row.Active.Valid {
		active = row.Active.Bool
	}
	unitType := ""
	if row.UnitType != nil {
		unitType = row.UnitType.(string)
	}

	return &models.LeaseSearchResult{
		LeaseID:        int(row.LeaseID),
		TenantID:       int(row.TenantID.Int32),
		TenantName:     row.TenantName.String,
		TenantPhone:    row.TenantPhone,
		PropertyID:     int(row.PropertyID.Int32),
		PropertyName:   row.PropertyName.String,
		BuildingID:     int(row.BuildingID.Int32),
		BuildingName:   row.BuildingName.String,
		BuildingCode:   row.BuildingCode.String,
		UnitID:         int(row.UnitID.Int32),
		UnitNumber:     row.UnitNumber.String,
		UnitType:       unitType,
		LeaseStartDate: row.LeaseStartDate,
		LeaseEndDate:   leaseEndDate,
		MonthlyRent:    row.MonthlyRent,
		Active:         active,
	}
}
