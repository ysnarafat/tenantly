package repositories

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/ysnarafat/tenantly/internal/db"
	"github.com/ysnarafat/tenantly/internal/models"
)

// nullTimeToPtr converts sql.NullTime to *time.Time.
func nullTimeToPtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

// nullTimeToTime converts sql.NullTime to time.Time (zero value when invalid).
func nullTimeToTime(t sql.NullTime) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// nullStringToFloat64 parses a sql.NullString that holds a numeric value.
func nullStringToFloat64(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s.String, 64)
	return f
}

// nullStringToStr safely extracts the string value.
func nullStringToStr(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

// interfaceToFloat64 converts interface{} DB values (NUMERIC/COALESCE results) to float64.
func interfaceToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case []byte:
		f, _ := strconv.ParseFloat(string(val), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

// interfaceToStr converts interface{} DB values (enum casts) to string.
func interfaceToStr(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	}
	return ""
}

// rawPaymentFields holds the common payment columns shared across multiple query result types.
type rawPaymentFields struct {
	ID             int32
	UnitID         int32
	TenantID       int32
	BuildingID     int32
	PropertyID     int32
	OrganizationID int32
	Month          int32
	Year           int32
	AmountDue      float64
	AmountPaid     sql.NullString
	Status         sql.NullString
	PaymentMethod  string
	Notes          string
	ReceiptNumber  string
	PaymentDate    sql.NullTime
	DueDate        sql.NullTime
	CreatedAt      sql.NullTime
	UpdatedAt      sql.NullTime
}

func rawToPayment(r rawPaymentFields) models.Payment {
	return models.Payment{
		ID:             int(r.ID),
		UnitID:         int(r.UnitID),
		TenantID:       int(r.TenantID),
		BuildingID:     int(r.BuildingID),
		PropertyID:     int(r.PropertyID),
		OrganizationID: int(r.OrganizationID),
		Month:          int(r.Month),
		Year:           int(r.Year),
		AmountDue:      r.AmountDue,
		AmountPaid:     nullStringToFloat64(r.AmountPaid),
		Status:         models.PaymentStatus(nullStringToStr(r.Status)),
		PaymentMethod:  r.PaymentMethod,
		Notes:          r.Notes,
		ReceiptNumber:  r.ReceiptNumber,
		PaymentDate:    nullTimeToPtr(r.PaymentDate),
		DueDate:        nullTimeToPtr(r.DueDate),
		CreatedAt:      nullTimeToTime(r.CreatedAt),
		UpdatedAt:      nullTimeToTime(r.UpdatedAt),
	}
}

// toPayment converts a GetPaymentByID row to a Payment model.
func toPayment(row db.GetPaymentByIDRow) *models.Payment {
	p := rawToPayment(rawPaymentFields{
		ID: row.ID, UnitID: row.UnitID, TenantID: row.TenantID,
		BuildingID: row.BuildingID, PropertyID: row.PropertyID,
		OrganizationID: row.OrganizationID, Month: row.Month, Year: row.Year,
		AmountDue: row.AmountDue, AmountPaid: row.AmountPaid, Status: row.Status,
		PaymentMethod: row.PaymentMethod, Notes: row.Notes, ReceiptNumber: row.ReceiptNumber,
		PaymentDate: row.PaymentDate, DueDate: row.DueDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	})
	return &p
}

// toPaymentFromCreate converts a CreatePayment result to a Payment model.
func toPaymentFromCreate(row db.CreatePaymentRow) *models.Payment {
	p := rawToPayment(rawPaymentFields{
		ID: row.ID, UnitID: row.UnitID, TenantID: row.TenantID,
		BuildingID: row.BuildingID, PropertyID: row.PropertyID,
		OrganizationID: row.OrganizationID, Month: row.Month, Year: row.Year,
		AmountDue: row.AmountDue, AmountPaid: row.AmountPaid, Status: row.Status,
		PaymentMethod: row.PaymentMethod, Notes: row.Notes, ReceiptNumber: row.ReceiptNumber,
		PaymentDate: row.PaymentDate, DueDate: row.DueDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	})
	return &p
}

// toPaymentWithDetails converts a GetPaymentByIDWithDetails row to a PaymentWithDetails model.
func toPaymentWithDetails(row db.GetPaymentByIDWithDetailsRow) *models.PaymentWithDetails {
	p := rawToPayment(rawPaymentFields{
		ID: row.ID, UnitID: row.UnitID, TenantID: row.TenantID,
		BuildingID: row.BuildingID, PropertyID: row.PropertyID,
		OrganizationID: row.OrganizationID, Month: row.Month, Year: row.Year,
		AmountDue: row.AmountDue, AmountPaid: row.AmountPaid, Status: row.Status,
		PaymentMethod: row.PaymentMethod, Notes: row.Notes, ReceiptNumber: row.ReceiptNumber,
		PaymentDate: row.PaymentDate, DueDate: row.DueDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	})
	return &models.PaymentWithDetails{
		Payment:      p,
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     interfaceToStr(row.UnitType),
		TenantName:   row.TenantName,
	}
}

// toPaymentWithDetailsFromPeriod converts a GetBuildingPaymentsInPeriod row.
func toPaymentWithDetailsFromPeriod(row db.GetBuildingPaymentsInPeriodRow) *models.PaymentWithDetails {
	p := rawToPayment(rawPaymentFields{
		ID: row.ID, UnitID: row.UnitID, TenantID: row.TenantID,
		BuildingID: row.BuildingID, PropertyID: row.PropertyID,
		OrganizationID: row.OrganizationID, Month: row.Month, Year: row.Year,
		AmountDue: row.AmountDue, AmountPaid: row.AmountPaid, Status: row.Status,
		PaymentMethod: row.PaymentMethod, Notes: row.Notes, ReceiptNumber: row.ReceiptNumber,
		PaymentDate: row.PaymentDate, DueDate: row.DueDate,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	})
	return &models.PaymentWithDetails{
		Payment:      p,
		PropertyName: row.PropertyName,
		BuildingName: row.BuildingName,
		BuildingCode: row.BuildingCode,
		UnitNumber:   row.UnitNumber,
		UnitType:     interfaceToStr(row.UnitType),
		TenantName:   row.TenantName,
	}
}

func buildPaymentStats(totalRecords, paidCount, dueCount, partialCount, overdueCount int64, totalDue, totalPaid, totalOverdue interface{}) *models.PaymentStats {
	due := interfaceToFloat64(totalDue)
	paid := interfaceToFloat64(totalPaid)
	overdue := interfaceToFloat64(totalOverdue)

	stats := &models.PaymentStats{
		TotalRecords: int(totalRecords),
		TotalDue:     due,
		TotalPaid:    paid,
		TotalOverdue: overdue,
		TotalPending: due - paid,
		PaidCount:    int(paidCount),
		DueCount:     int(dueCount),
		PartialCount: int(partialCount),
		OverdueCount: int(overdueCount),
	}
	if due > 0 {
		stats.CollectionRate = (paid / due) * 100
	}
	return stats
}

// toPaymentStats converts building stats row to PaymentStats model.
func toPaymentStats(row db.GetBuildingPaymentStatsRow) interface{} {
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue)
}

// toPropertyPaymentStats converts property stats row to PaymentStats model.
func toPropertyPaymentStats(row db.GetPropertyPaymentStatsRow) interface{} {
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue)
}

// toSystemPaymentStats converts system stats row to PaymentStats model.
func toSystemPaymentStats(row db.GetSystemPaymentStatsRow) interface{} {
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue)
}

// toDashboardSummary converts dashboard query row to DashboardSummary.
func toDashboardSummary(row db.GetDashboardSummaryRow) *models.DashboardSummary {
	totalDue := interfaceToFloat64(row.TotalDue)
	totalPaid := interfaceToFloat64(row.TotalPaid)

	s := &models.DashboardSummary{
		TotalDue:      totalDue,
		TotalPaid:     totalPaid,
		TotalPending:  interfaceToFloat64(row.TotalPending),
		TotalOverdue:  interfaceToFloat64(row.TotalOverdue),
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

// toBuildingLevelSummary converts to map format.
func toBuildingLevelSummary(row db.GetBuildingLevelSummaryRow) map[string]interface{} {
	return map[string]interface{}{
		"total_buildings": row.TotalBuildings,
		"total_due":       row.TotalDue,
		"total_paid":      row.TotalPaid,
	}
}

// toBuildingAnalytics converts analytics row.
func toBuildingAnalytics(row db.GetBuildingPaymentAnalyticsRow, buildingID int, startDate, endDate time.Time) *models.BuildingPaymentAnalytics {
	return &models.BuildingPaymentAnalytics{
		BuildingID:         buildingID,
		Period:             startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalRevenue:       interfaceToFloat64(row.TotalRevenue),
		CollectionRate:     float64(row.CollectionRate),
		AveragePaymentTime: interfaceToFloat64(row.AvgPaymentDays),
		OverduePayments:    int(row.OverdueCount),
	}
}

// toLeaseSearchResult converts a search result row.
func toLeaseSearchResult(row db.SearchLeasesRow) *models.LeaseSearchResult {
	leaseEnd := time.Time{}
	if row.LeaseEndDate.Valid {
		leaseEnd = row.LeaseEndDate.Time
	}
	return &models.LeaseSearchResult{
		LeaseID:        int(row.LeaseID),
		TenantID:       int(row.TenantID.Int32),
		TenantName:     nullStringToStr(row.TenantName),
		TenantPhone:    row.TenantPhone,
		PropertyID:     int(row.PropertyID.Int32),
		PropertyName:   nullStringToStr(row.PropertyName),
		BuildingID:     int(row.BuildingID.Int32),
		BuildingName:   nullStringToStr(row.BuildingName),
		BuildingCode:   nullStringToStr(row.BuildingCode),
		UnitID:         int(row.UnitID.Int32),
		UnitNumber:     nullStringToStr(row.UnitNumber),
		UnitType:       interfaceToStr(row.UnitType),
		LeaseStartDate: row.LeaseStartDate,
		LeaseEndDate:   leaseEnd,
		MonthlyRent:    row.MonthlyRent,
		Active:         row.Active.Bool,
	}
}
