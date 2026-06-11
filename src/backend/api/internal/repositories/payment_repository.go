package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/db"
	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentRepository implements PaymentRepositoryInterface
type PaymentRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(sqlDB *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		db:      sqlDB,
		queries: db.New(sqlDB),
	}
}

// Create inserts a new payment record
func (r *PaymentRepository) Create(req *models.CreatePaymentRequest) (*models.Payment, error) {
	var dueDate sql.NullTime
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format (expected YYYY-MM-DD): %w", err)
		}
		dueDate = sql.NullTime{Time: t, Valid: true}
	}

	var paymentDate sql.NullTime
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		t, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return nil, fmt.Errorf("invalid payment_date format (expected YYYY-MM-DD): %w", err)
		}
		paymentDate = sql.NullTime{Time: t, Valid: true}
	}

	amountPaid := 0.0
	if req.AmountPaid != nil {
		amountPaid = *req.AmountPaid
	}

	status := string(models.PaymentStatusDue)
	if req.Status != nil {
		status = string(*req.Status)
	} else if amountPaid > 0 {
		if amountPaid >= req.AmountDue {
			status = string(models.PaymentStatusPaid)
		} else {
			status = string(models.PaymentStatusPartial)
		}
	}

	var paymentMethod sql.NullString
	if req.PaymentMethod != nil && *req.PaymentMethod != "" {
		paymentMethod = sql.NullString{String: *req.PaymentMethod, Valid: true}
	}

	var receiptNumber sql.NullString
	if req.ReceiptNumber != nil && *req.ReceiptNumber != "" {
		receiptNumber = sql.NullString{String: *req.ReceiptNumber, Valid: true}
	}

	var notes sql.NullString
	if req.Notes != nil && *req.Notes != "" {
		notes = sql.NullString{String: *req.Notes, Valid: true}
	}

	row, err := r.queries.CreatePayment(context.Background(), db.CreatePaymentParams{
		UnitID:         int32(req.UnitID),
		TenantID:       int32(req.TenantID),
		BuildingID:     int32(req.BuildingID),
		PropertyID:     int32(req.PropertyID),
		OrganizationID: int32(req.OrganizationID),
		Month:          int32(req.Month),
		Year:           int32(req.Year),
		AmountDue:      req.AmountDue,
		AmountPaid:     amountPaid,
		Status:         status,
		PaymentMethod:  paymentMethod,
		PaymentDate:    paymentDate,
		ReceiptNumber:  receiptNumber,
		Notes:          notes,
		DueDate:        dueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return toPaymentFromCreate(row), nil
}

// GetByID retrieves a payment by its ID
func (r *PaymentRepository) GetByID(id int) (*models.Payment, error) {
	row, err := r.queries.GetPaymentByID(context.Background(), int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return toPayment(row), nil
}

// GetByIDWithDetails retrieves a payment with related entity details
func (r *PaymentRepository) GetByIDWithDetails(id int) (*models.PaymentWithDetails, error) {
	row, err := r.queries.GetPaymentByIDWithDetails(context.Background(), int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment details: %w", err)
	}
	return toPaymentWithDetails(row), nil
}

// GetBuildingPaymentsInPeriod returns paginated payments for a building within a date range
func (r *PaymentRepository) GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	count, err := r.queries.CountBuildingPaymentsInPeriod(context.Background(), db.CountBuildingPaymentsInPeriodParams{
		BuildingID:  int32(buildingID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count building payments: %w", err)
	}

	rows, err := r.queries.GetBuildingPaymentsInPeriod(context.Background(), db.GetBuildingPaymentsInPeriodParams{
		BuildingID:  int32(buildingID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query building payments: %w", err)
	}

	payments := make([]*models.PaymentWithDetails, len(rows))
	for i, row := range rows {
		payments[i] = toPaymentWithDetailsFromPeriod(row)
	}

	return payments, int(count), nil
}

// GetBuildingPaymentStats returns aggregate payment stats for a building in a date range
func (r *PaymentRepository) GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	row, err := r.queries.GetBuildingPaymentStats(context.Background(), db.GetBuildingPaymentStatsParams{
		BuildingID:  int32(buildingID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment stats: %w", err)
	}
	return toPaymentStats(row), nil
}

// GetPropertyPaymentStats returns aggregate payment stats for a property in a date range
func (r *PaymentRepository) GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	row, err := r.queries.GetPropertyPaymentStats(context.Background(), db.GetPropertyPaymentStatsParams{
		PropertyID:  int32(propertyID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}
	return toPropertyPaymentStats(row), nil
}

// GetSystemPaymentStats returns system-wide aggregate payment stats for a date range
func (r *PaymentRepository) GetSystemPaymentStats(startDate, endDate time.Time) (interface{}, error) {
	row, err := r.queries.GetSystemPaymentStats(context.Background(), db.GetSystemPaymentStatsParams{
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get system payment stats: %w", err)
	}
	return toSystemPaymentStats(row), nil
}

// GetDashboardSummary returns payment summary scoped to the given organisation.
func (r *PaymentRepository) GetDashboardSummary(orgID int) (*models.DashboardSummary, error) {
	row, err := r.queries.GetDashboardSummary(context.Background(), int32(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	return toDashboardSummary(row), nil
}

// GetBuildingLevelSummary returns a map with building-level aggregate totals
func (r *PaymentRepository) GetBuildingLevelSummary() (map[string]interface{}, error) {
	row, err := r.queries.GetBuildingLevelSummary(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get building level summary: %w", err)
	}
	return toBuildingLevelSummary(row), nil
}

// SearchLeases searches for active leases across tenants, properties, buildings, and units
func (r *PaymentRepository) SearchLeases(orgID int, query string) ([]*models.LeaseSearchResult, error) {
	searchPattern := "%" + query + "%"

	rows, err := r.queries.SearchLeases(context.Background(), db.SearchLeasesParams{
		OrganizationID: int32(orgID),
		Name:           searchPattern,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search leases: %w", err)
	}

	results := make([]*models.LeaseSearchResult, len(rows))
	for i, row := range rows {
		results[i] = toLeaseSearchResult(row)
	}

	return results, nil
}

// GetBuildingPaymentAnalytics returns detailed analytics for a building over a date range
func (r *PaymentRepository) GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error) {
	analyticsRow, err := r.queries.GetBuildingPaymentAnalytics(context.Background(), db.GetBuildingPaymentAnalyticsParams{
		BuildingID:  int32(buildingID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment analytics: %w", err)
	}

	analytics := toBuildingAnalytics(analyticsRow, buildingID, startDate, endDate)

	// Fetch monthly trend
	trendRows, err := r.queries.GetBuildingPaymentTrend(context.Background(), db.GetBuildingPaymentTrendParams{
		BuildingID:  int32(buildingID),
		CreatedAt:   sql.NullTime{Time: startDate, Valid: true},
		CreatedAt_2: sql.NullTime{Time: endDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment trend: %w", err)
	}

	type trendPoint struct {
		Year  int     `json:"year"`
		Month int     `json:"month"`
		Due   float64 `json:"due"`
		Paid  float64 `json:"paid"`
	}
	trend := make([]trendPoint, len(trendRows))
	for i, row := range trendRows {
		trend[i] = trendPoint{
			Year:  int(row.Year),
			Month: int(row.Month),
			Due:   interfaceToFloat64(row.Due),
			Paid:  interfaceToFloat64(row.Paid),
		}
	}
	analytics.TrendAnalysis = trend

	return analytics, nil
}
