package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentRepository implements PaymentRepositoryInterface
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentWithDetailsCols = `
	p.id, p.unit_id, p.tenant_id, p.building_id, p.property_id, p.organization_id,
	p.month, p.year, p.amount_due, p.amount_paid, p.status,
	COALESCE(p.payment_method, ''), COALESCE(p.notes, ''), COALESCE(p.receipt_number, ''),
	p.payment_date, p.due_date, p.created_at, p.updated_at,
	COALESCE(pr.property_name, ''), COALESCE(b.building_name, ''),
	COALESCE(b.building_code, ''), COALESCE(u.unit_number, ''),
	COALESCE(u.unit_type::text, ''), COALESCE(t.name, '')`

const paymentDetailJoins = `
	LEFT JOIN units u ON p.unit_id = u.id
	LEFT JOIN tenants t ON p.tenant_id = t.id
	LEFT JOIN buildings b ON p.building_id = b.id
	LEFT JOIN properties pr ON p.property_id = pr.id`

func scanPaymentWithDetails(row interface {
	Scan(...interface{}) error
}) (*models.PaymentWithDetails, error) {
	p := &models.PaymentWithDetails{}
	err := row.Scan(
		&p.ID, &p.UnitID, &p.TenantID, &p.BuildingID, &p.PropertyID, &p.OrganizationID,
		&p.Month, &p.Year, &p.AmountDue, &p.AmountPaid, &p.Status,
		&p.PaymentMethod, &p.Notes, &p.ReceiptNumber,
		&p.PaymentDate, &p.DueDate, &p.CreatedAt, &p.UpdatedAt,
		&p.PropertyName, &p.BuildingName, &p.BuildingCode,
		&p.UnitNumber, &p.UnitType, &p.TenantName,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Create inserts a new payment record
func (r *PaymentRepository) Create(req *models.CreatePaymentRequest) (*models.Payment, error) {
	var dueDate *time.Time
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format (expected YYYY-MM-DD): %w", err)
		}
		dueDate = &t
	}

	payment := &models.Payment{
		UnitID:         req.UnitID,
		TenantID:       req.TenantID,
		BuildingID:     req.BuildingID,
		PropertyID:     req.PropertyID,
		OrganizationID: req.OrganizationID,
		Month:          req.Month,
		Year:           req.Year,
		AmountDue:      req.AmountDue,
		AmountPaid:     0,
		Status:         models.PaymentStatusDue,
		DueDate:        dueDate,
	}

	query := `
		INSERT INTO payments (
			unit_id, tenant_id, building_id, property_id, organization_id,
			month, year, amount_due, amount_paid, status, due_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 'Due', $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query,
		payment.UnitID, payment.TenantID, payment.BuildingID, payment.PropertyID,
		payment.OrganizationID, payment.Month, payment.Year, payment.AmountDue, dueDate,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// GetByID retrieves a payment by its ID
func (r *PaymentRepository) GetByID(id int) (*models.Payment, error) {
	query := `
		SELECT id, unit_id, tenant_id, building_id, property_id, organization_id,
			month, year, amount_due, amount_paid, status,
			COALESCE(payment_method, ''), COALESCE(notes, ''), COALESCE(receipt_number, ''),
			payment_date, due_date, created_at, updated_at
		FROM payments
		WHERE id = $1`

	p := &models.Payment{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.UnitID, &p.TenantID, &p.BuildingID, &p.PropertyID, &p.OrganizationID,
		&p.Month, &p.Year, &p.AmountDue, &p.AmountPaid, &p.Status,
		&p.PaymentMethod, &p.Notes, &p.ReceiptNumber,
		&p.PaymentDate, &p.DueDate, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return p, nil
}

// GetByIDWithDetails retrieves a payment with related entity details
func (r *PaymentRepository) GetByIDWithDetails(id int) (*models.PaymentWithDetails, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM payments p
		%s
		WHERE p.id = $1`,
		paymentWithDetailsCols, paymentDetailJoins)

	p, err := scanPaymentWithDetails(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment details: %w", err)
	}
	return p, nil
}

// Update applies a partial update to a payment and auto-calculates status when amount_paid changes
func (r *PaymentRepository) Update(id int, req *models.UpdatePaymentRequest) (*models.Payment, error) {
	setClauses := []string{"updated_at = NOW() AT TIME ZONE 'UTC'"}
	args := []interface{}{}
	argIdx := 1

	if req.AmountPaid != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount_paid = $%d", argIdx))
		args = append(args, *req.AmountPaid)
		argIdx++

		// Auto-derive status from amount when caller does not override it
		if req.Status == nil {
			setClauses = append(setClauses, fmt.Sprintf(
				"status = CASE WHEN $%d >= amount_due THEN 'Paid' WHEN $%d > 0 THEN 'Partial' ELSE status END",
				argIdx, argIdx))
			args = append(args, *req.AmountPaid)
			argIdx++
		}
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*req.Status))
		argIdx++
	}
	if req.PaymentMethod != nil {
		setClauses = append(setClauses, fmt.Sprintf("payment_method = $%d", argIdx))
		args = append(args, *req.PaymentMethod)
		argIdx++
	}
	if req.PaymentDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("payment_date = $%d", argIdx))
		args = append(args, *req.PaymentDate)
		argIdx++
	}
	if req.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}
	if req.ReceiptNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("receipt_number = $%d", argIdx))
		args = append(args, *req.ReceiptNumber)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE payments SET %s
		WHERE id = $%d
		RETURNING id, unit_id, tenant_id, building_id, property_id, organization_id,
			month, year, amount_due, amount_paid, status,
			COALESCE(payment_method, ''), COALESCE(notes, ''), COALESCE(receipt_number, ''),
			payment_date, due_date, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx)

	p := &models.Payment{}
	err := r.db.QueryRow(query, args...).Scan(
		&p.ID, &p.UnitID, &p.TenantID, &p.BuildingID, &p.PropertyID, &p.OrganizationID,
		&p.Month, &p.Year, &p.AmountDue, &p.AmountPaid, &p.Status,
		&p.PaymentMethod, &p.Notes, &p.ReceiptNumber,
		&p.PaymentDate, &p.DueDate, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}
	return p, nil
}

// GetWithDetailsAndFilters returns a paginated, filtered list of payments with related details
func (r *PaymentRepository) GetWithDetailsAndFilters(filters map[string]interface{}, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	conditions := []string{}
	args := []interface{}{}
	idx := 1

	filterKeys := []string{"building_id", "property_id", "status", "month", "year", "organization_id"}
	colMap := map[string]string{
		"building_id":     "p.building_id",
		"property_id":     "p.property_id",
		"status":          "p.status",
		"month":           "p.month",
		"year":            "p.year",
		"organization_id": "p.organization_id",
	}
	for _, k := range filterKeys {
		if v, ok := filters[k]; ok {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", colMap[k], idx))
			args = append(args, v)
			idx++
		}
	}

	whereClause := "1=1"
	if len(conditions) > 0 {
		whereClause = strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM payments p %s
		WHERE %s`, paymentDetailJoins, whereClause)

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM payments p %s
		WHERE %s
		ORDER BY p.year DESC, p.month DESC, p.created_at DESC
		LIMIT $%d OFFSET $%d`,
		paymentWithDetailsCols, paymentDetailJoins, whereClause, idx, idx+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	payments := make([]*models.PaymentWithDetails, 0, limit)
	for rows.Next() {
		p, err := scanPaymentWithDetails(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment row: %w", err)
		}
		payments = append(payments, p)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("payment rows error: %w", err)
	}

	return payments, total, nil
}

// GetBuildingPaymentsInPeriod returns paginated payments for a building within a date range
func (r *PaymentRepository) GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	where := `p.building_id = $1 AND p.created_at >= $2 AND p.created_at <= $3`

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM payments p %s WHERE %s`, paymentDetailJoins, where)
	var total int
	if err := r.db.QueryRow(countQuery, buildingID, startDate, endDate).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count building payments: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM payments p %s
		WHERE %s
		ORDER BY p.year DESC, p.month DESC
		LIMIT $4 OFFSET $5`,
		paymentWithDetailsCols, paymentDetailJoins, where)

	rows, err := r.db.Query(dataQuery, buildingID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query building payments: %w", err)
	}
	defer rows.Close()

	payments := make([]*models.PaymentWithDetails, 0, limit)
	for rows.Next() {
		p, err := scanPaymentWithDetails(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan building payment row: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, rows.Err()
}

func scanPaymentStats(db *sql.DB, query string, args ...interface{}) (interface{}, error) {
	stats := &models.PaymentStats{}
	var totalDue, totalPaid, totalOverdue float64
	var totalRecords, paidCount, dueCount, partialCount, overdueCount int

	err := db.QueryRow(query, args...).Scan(
		&totalRecords, &totalDue, &totalPaid, &totalOverdue,
		&paidCount, &dueCount, &partialCount, &overdueCount,
	)
	if err != nil {
		return nil, err
	}

	stats.TotalRecords = totalRecords
	stats.TotalDue = totalDue
	stats.TotalPaid = totalPaid
	stats.TotalOverdue = totalOverdue
	stats.TotalPending = totalDue - totalPaid
	stats.PaidCount = paidCount
	stats.DueCount = dueCount
	stats.PartialCount = partialCount
	stats.OverdueCount = overdueCount
	if totalDue > 0 {
		stats.CollectionRate = (totalPaid / totalDue) * 100
	}

	return stats, nil
}

const paymentStatsQuery = `
	SELECT
		COUNT(*)                                         AS total_records,
		COALESCE(SUM(amount_due), 0)                    AS total_due,
		COALESCE(SUM(amount_paid), 0)                   AS total_paid,
		COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
		COUNT(CASE WHEN status='Paid'     THEN 1 END)   AS paid_count,
		COUNT(CASE WHEN status='Due'      THEN 1 END)   AS due_count,
		COUNT(CASE WHEN status='Partial'  THEN 1 END)   AS partial_count,
		COUNT(CASE WHEN status='Overdue'  THEN 1 END)   AS overdue_count
	FROM payments
	WHERE %s`

// GetBuildingPaymentStats returns aggregate payment stats for a building in a date range
func (r *PaymentRepository) GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	query := fmt.Sprintf(paymentStatsQuery, "building_id = $1 AND created_at >= $2 AND created_at <= $3")
	stats, err := scanPaymentStats(r.db, query, buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment stats: %w", err)
	}
	return stats, nil
}

// GetPropertyPaymentStats returns aggregate payment stats for a property in a date range
func (r *PaymentRepository) GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	query := fmt.Sprintf(paymentStatsQuery, "property_id = $1 AND created_at >= $2 AND created_at <= $3")
	stats, err := scanPaymentStats(r.db, query, propertyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}
	return stats, nil
}

// GetSystemPaymentStats returns system-wide aggregate payment stats for a date range
func (r *PaymentRepository) GetSystemPaymentStats(startDate, endDate time.Time) (interface{}, error) {
	query := fmt.Sprintf(paymentStatsQuery, "created_at >= $1 AND created_at <= $2")
	stats, err := scanPaymentStats(r.db, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get system payment stats: %w", err)
	}
	return stats, nil
}

// GetDashboardSummary returns overall payment summary across the system
func (r *PaymentRepository) GetDashboardSummary() (*models.DashboardSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(amount_due), 0)                                            AS total_due,
			COALESCE(SUM(amount_paid), 0)                                           AS total_paid,
			COALESCE(SUM(CASE WHEN status != 'Paid' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_pending,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			(SELECT COUNT(DISTINCT id)   FROM properties)                           AS property_count,
			(SELECT COUNT(DISTINCT id)   FROM buildings)                            AS building_count,
			(SELECT COUNT(DISTINCT id)   FROM units WHERE active = true)            AS unit_count,
			(SELECT COUNT(DISTINCT id)   FROM tenants WHERE active = true)          AS tenant_count
		FROM payments`

	s := &models.DashboardSummary{}
	err := r.db.QueryRow(query).Scan(
		&s.TotalDue, &s.TotalPaid, &s.TotalPending, &s.TotalOverdue,
		&s.PropertyCount, &s.BuildingCount, &s.UnitCount, &s.TenantCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	if s.TotalDue > 0 {
		s.CollectionRate = (s.TotalPaid / s.TotalDue) * 100
	}
	return s, nil
}

// GetBuildingLevelSummary returns a map with building-level aggregate totals
func (r *PaymentRepository) GetBuildingLevelSummary() (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(DISTINCT building_id)                  AS total_buildings,
			COALESCE(SUM(amount_due), 0)                 AS total_due,
			COALESCE(SUM(amount_paid), 0)                AS total_paid
		FROM payments`

	var totalBuildings int
	var totalDue, totalPaid float64
	err := r.db.QueryRow(query).Scan(&totalBuildings, &totalDue, &totalPaid)
	if err != nil {
		return nil, fmt.Errorf("failed to get building level summary: %w", err)
	}

	return map[string]interface{}{
		"total_buildings": totalBuildings,
		"total_due":       totalDue,
		"total_paid":      totalPaid,
	}, nil
}

// SearchLeases searches for active leases across tenants, properties, buildings, and units
func (r *PaymentRepository) SearchLeases(orgID int, query string) ([]*models.LeaseSearchResult, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"

	sqlQuery := `
		SELECT
			l.id AS lease_id,
			t.id AS tenant_id,
			t.name AS tenant_name,
			COALESCE(t.phone, '') AS tenant_phone,
			p.id AS property_id,
			p.property_name,
			b.id AS building_id,
			b.building_name,
			b.building_code,
			u.id AS unit_id,
			u.unit_number,
			COALESCE(u.unit_type::text, '') AS unit_type,
			l.start_date AS lease_start_date,
			l.end_date AS lease_end_date,
			l.monthly_rent,
			l.active
		FROM leases l
		LEFT JOIN tenants t ON l.tenant_id = t.id
		LEFT JOIN units u ON l.unit_id = u.id
		LEFT JOIN buildings b ON u.building_id = b.id
		LEFT JOIN properties p ON u.property_id = p.id
		WHERE l.organization_id = $1
			AND l.active = true
			AND (
				LOWER(t.name) LIKE $2
				OR LOWER(p.property_name) LIKE $2
				OR LOWER(b.building_name) LIKE $2
				OR LOWER(b.building_code) LIKE $2
				OR LOWER(u.unit_number) LIKE $2
				OR LOWER(t.phone) LIKE $2
				OR CAST(l.id AS TEXT) LIKE $2
			)
		ORDER BY t.name, l.start_date DESC
		LIMIT 50`

	rows, err := r.db.Query(sqlQuery, orgID, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search leases: %w", err)
	}
	defer rows.Close()

	results := make([]*models.LeaseSearchResult, 0)
	for rows.Next() {
		result := &models.LeaseSearchResult{}
		err := rows.Scan(
			&result.LeaseID, &result.TenantID, &result.TenantName, &result.TenantPhone,
			&result.PropertyID, &result.PropertyName,
			&result.BuildingID, &result.BuildingName, &result.BuildingCode,
			&result.UnitID, &result.UnitNumber, &result.UnitType,
			&result.LeaseStartDate, &result.LeaseEndDate,
			&result.MonthlyRent, &result.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lease search result: %w", err)
		}
		results = append(results, result)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("lease search rows error: %w", err)
	}

	return results, nil
}

// GetBuildingPaymentAnalytics returns detailed analytics for a building over a date range
func (r *PaymentRepository) GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error) {
	query := `
		SELECT
			COALESCE(SUM(amount_paid), 0)                                        AS total_revenue,
			CASE WHEN COALESCE(SUM(amount_due), 0) > 0
				THEN (COALESCE(SUM(amount_paid), 0) / COALESCE(SUM(amount_due), 0)) * 100
				ELSE 0 END                                                       AS collection_rate,
			COALESCE(AVG(
				CASE WHEN payment_date IS NOT NULL AND due_date IS NOT NULL
					THEN EXTRACT(EPOCH FROM (payment_date - due_date)) / 86400
					ELSE NULL END
			), 0)                                                                AS avg_payment_days,
			COUNT(CASE WHEN status='Overdue' THEN 1 END)                         AS overdue_count
		FROM payments
		WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3`

	analytics := &models.BuildingPaymentAnalytics{
		BuildingID: buildingID,
		Period:     fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}

	err := r.db.QueryRow(query, buildingID, startDate, endDate).Scan(
		&analytics.TotalRevenue,
		&analytics.CollectionRate,
		&analytics.AveragePaymentTime,
		&analytics.OverduePayments,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment analytics: %w", err)
	}

	// Monthly trend for the period
	trendQuery := `
		SELECT year, month,
			COALESCE(SUM(amount_due), 0)  AS due,
			COALESCE(SUM(amount_paid), 0) AS paid
		FROM payments
		WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY year, month
		ORDER BY year, month`

	rows, err := r.db.Query(trendQuery, buildingID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get building payment trend: %w", err)
	}
	defer rows.Close()

	type trendPoint struct {
		Year  int     `json:"year"`
		Month int     `json:"month"`
		Due   float64 `json:"due"`
		Paid  float64 `json:"paid"`
	}
	trend := make([]trendPoint, 0)
	for rows.Next() {
		var tp trendPoint
		if err := rows.Scan(&tp.Year, &tp.Month, &tp.Due, &tp.Paid); err != nil {
			return nil, fmt.Errorf("failed to scan trend row: %w", err)
		}
		trend = append(trend, tp)
	}
	analytics.TrendAnalysis = trend

	return analytics, nil
}
