package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentRepository implements PaymentRepositoryInterface using sqlx.
type PaymentRepository struct {
	db *sqlx.DB
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// paymentCols is the canonical SELECT column list for the payments table.
// COALESCE ensures nullable columns scan cleanly into non-pointer Go types.
const paymentCols = `
	id, unit_id, tenant_id, building_id, property_id, organization_id,
	month, year, amount_due,
	COALESCE(amount_paid, 0)     AS amount_paid,
	COALESCE(status, 'Due')      AS status,
	COALESCE(payment_method, '') AS payment_method,
	COALESCE(notes, '')          AS notes,
	COALESCE(receipt_number, '') AS receipt_number,
	payment_date, due_date, created_at, updated_at`

const paymentWithDetailCols = `
	p.id, p.unit_id, p.tenant_id, p.building_id, p.property_id, p.organization_id,
	p.month, p.year, p.amount_due,
	COALESCE(p.amount_paid, 0)     AS amount_paid,
	COALESCE(p.status, 'Due')      AS status,
	COALESCE(p.payment_method, '') AS payment_method,
	COALESCE(p.notes, '')          AS notes,
	COALESCE(p.receipt_number, '') AS receipt_number,
	p.payment_date, p.due_date, p.created_at, p.updated_at,
	COALESCE(pr.property_name, '') AS property_name,
	COALESCE(b.building_name, '')  AS building_name,
	COALESCE(b.building_code, '')  AS building_code,
	COALESCE(u.unit_number, '')    AS unit_number,
	COALESCE(u.unit_type::text, '') AS unit_type,
	COALESCE(t.name, '')           AS tenant_name`

const paymentDetailJoinsQ = `
	LEFT JOIN units u       ON p.unit_id = u.id
	LEFT JOIN tenants t     ON p.tenant_id = t.id
	LEFT JOIN buildings b   ON p.building_id = b.id
	LEFT JOIN properties pr ON p.property_id = pr.id`

// statsRow scans aggregate payment statistics.
type statsRow struct {
	TotalRecords int64   `db:"total_records"`
	TotalDue     float64 `db:"total_due"`
	TotalPaid    float64 `db:"total_paid"`
	TotalOverdue float64 `db:"total_overdue"`
	PaidCount    int64   `db:"paid_count"`
	DueCount     int64   `db:"due_count"`
	PartialCount int64   `db:"partial_count"`
	OverdueCount int64   `db:"overdue_count"`
}

type dashboardRow struct {
	TotalDue      float64 `db:"total_due"`
	TotalPaid     float64 `db:"total_paid"`
	TotalPending  float64 `db:"total_pending"`
	TotalOverdue  float64 `db:"total_overdue"`
	PropertyCount int64   `db:"property_count"`
	BuildingCount int64   `db:"building_count"`
	UnitCount     int64   `db:"unit_count"`
	TenantCount   int64   `db:"tenant_count"`
}

type buildingLevelRow struct {
	TotalBuildings int64   `db:"total_buildings"`
	TotalDue       float64 `db:"total_due"`
	TotalPaid      float64 `db:"total_paid"`
}

type buildingAnalyticsRow struct {
	TotalRevenue   float64 `db:"total_revenue"`
	CollectionRate float64 `db:"collection_rate"`
	AvgPaymentDays float64 `db:"avg_payment_days"`
	OverdueCount   int64   `db:"overdue_count"`
}

type trendPoint struct {
	Year  int     `db:"year"  json:"year"`
	Month int     `db:"month" json:"month"`
	Due   float64 `db:"due"   json:"due"`
	Paid  float64 `db:"paid"  json:"paid"`
}

// Create inserts a new payment record.
func (r *PaymentRepository) Create(req *models.CreatePaymentRequest) (*models.Payment, error) {
	var dueDate sql.NullTime
	if req.DueDate != "" {
		t, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format (expected YYYY-MM-DD): %w", err)
		}
		dueDate = sql.NullTime{Time: t, Valid: true}
	}

	amountPaid := 0.0
	if req.AmountPaid != nil {
		amountPaid = *req.AmountPaid
	}

	var paymentDate sql.NullTime
	if req.PaymentDate != nil && *req.PaymentDate != "" {
		t, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return nil, fmt.Errorf("invalid payment_date format (expected YYYY-MM-DD): %w", err)
		}
		paymentDate = sql.NullTime{Time: t, Valid: true}
	} else if amountPaid > 0 {
		// A payment is being recorded but no explicit date was given —
		// default to today rather than leaving payment_date null.
		paymentDate = sql.NullTime{Time: time.Now().UTC().Truncate(24 * time.Hour), Valid: true}
	}

	// Status is always derived server-side from amount_paid vs amount_due (and
	// due_date, for Overdue) — req.Status is ignored on create so a client
	// can't misreport payment state.
	status := string(models.PaymentStatusDue)
	if amountPaid > 0 {
		if amountPaid >= req.AmountDue {
			status = string(models.PaymentStatusPaid)
		} else {
			status = string(models.PaymentStatusPartial)
		}
	} else if dueDate.Valid && dueDate.Time.Before(time.Now().UTC().Truncate(24*time.Hour)) {
		status = string(models.PaymentStatusOverdue)
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

	q := `
		INSERT INTO payments (
			unit_id, tenant_id, building_id, property_id, organization_id,
			month, year, amount_due, amount_paid, status,
			payment_method, payment_date, receipt_number, notes, due_date
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING ` + paymentCols

	var p models.Payment
	err := r.db.GetContext(context.Background(), &p, q,
		req.UnitID, req.TenantID, req.BuildingID, req.PropertyID, req.OrganizationID,
		req.Month, req.Year, req.AmountDue, amountPaid, status,
		paymentMethod, paymentDate, receiptNumber, notes, dueDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}
	return &p, nil
}

// GetByID retrieves a payment by its ID.
func (r *PaymentRepository) GetByID(id int) (*models.Payment, error) {
	q := `SELECT ` + paymentCols + ` FROM payments WHERE id = $1`
	var p models.Payment
	err := r.db.GetContext(context.Background(), &p, q, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &p, nil
}

// GetByIDWithDetails retrieves a payment with related entity details.
func (r *PaymentRepository) GetByIDWithDetails(id int) (*models.PaymentWithDetails, error) {
	q := `SELECT ` + paymentWithDetailCols + `
		FROM payments p` + paymentDetailJoinsQ + `
		WHERE p.id = $1`
	var p models.PaymentWithDetails
	err := r.db.GetContext(context.Background(), &p, q, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment details: %w", err)
	}
	return &p, nil
}

// GetBuildingPaymentsInPeriod returns paginated payments for a building within a date range.
func (r *PaymentRepository) GetBuildingPaymentsInPeriod(buildingID int, startDate, endDate time.Time, limit, offset int) ([]*models.PaymentWithDetails, int, error) {
	var count int64
	err := r.db.GetContext(context.Background(), &count,
		`SELECT COUNT(*) FROM payments p`+paymentDetailJoinsQ+`
		 WHERE p.building_id = $1 AND p.created_at >= $2 AND p.created_at <= $3`,
		buildingID, startDate, endDate)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count building payments: %w", err)
	}

	q := `SELECT ` + paymentWithDetailCols + `
		FROM payments p` + paymentDetailJoinsQ + `
		WHERE p.building_id = $1 AND p.created_at >= $2 AND p.created_at <= $3
		ORDER BY p.year DESC, p.month DESC
		LIMIT $4 OFFSET $5`

	var payments []models.PaymentWithDetails
	if err := r.db.SelectContext(context.Background(), &payments, q, buildingID, startDate, endDate, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to query building payments: %w", err)
	}

	result := make([]*models.PaymentWithDetails, len(payments))
	for i := range payments {
		result[i] = &payments[i]
	}
	return result, int(count), nil
}

// GetBuildingPaymentStats returns aggregate payment stats for a building in a date range.
func (r *PaymentRepository) GetBuildingPaymentStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	const q = `
		SELECT
			COUNT(*) AS total_records,
			COALESCE(SUM(amount_due), 0)  AS total_due,
			COALESCE(SUM(amount_paid), 0) AS total_paid,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			COUNT(CASE WHEN status='Paid'    THEN 1 END) AS paid_count,
			COUNT(CASE WHEN status='Due'     THEN 1 END) AS due_count,
			COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
			COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
		FROM payments
		WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3`
	var row statsRow
	if err := r.db.GetContext(context.Background(), &row, q, buildingID, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get building payment stats: %w", err)
	}
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue), nil
}

// GetPropertyPaymentStats returns aggregate payment stats for a property in a date range.
func (r *PaymentRepository) GetPropertyPaymentStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	const q = `
		SELECT
			COUNT(*) AS total_records,
			COALESCE(SUM(amount_due), 0)  AS total_due,
			COALESCE(SUM(amount_paid), 0) AS total_paid,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			COUNT(CASE WHEN status='Paid'    THEN 1 END) AS paid_count,
			COUNT(CASE WHEN status='Due'     THEN 1 END) AS due_count,
			COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
			COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
		FROM payments
		WHERE property_id = $1 AND created_at >= $2 AND created_at <= $3`
	var row statsRow
	if err := r.db.GetContext(context.Background(), &row, q, propertyID, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get property payment stats: %w", err)
	}
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue), nil
}

// GetBatchPropertyPaymentStats returns aggregate payment stats for multiple properties in a single query.
func (r *PaymentRepository) GetBatchPropertyPaymentStats(propertyIDs []int, startDate, endDate time.Time) (map[int]any, error) {
	if len(propertyIDs) == 0 {
		return map[int]any{}, nil
	}

	placeholders := make([]string, len(propertyIDs))
	args := make([]interface{}, len(propertyIDs)+2)
	for i, id := range propertyIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(propertyIDs)] = startDate
	args[len(propertyIDs)+1] = endDate

	q := fmt.Sprintf(`
		SELECT
			property_id,
			COUNT(*)                                                                         AS total_records,
			COALESCE(SUM(amount_due), 0)                                                     AS total_due,
			COALESCE(SUM(amount_paid), 0)                                                    AS total_paid,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			COUNT(CASE WHEN status='Paid'    THEN 1 END)                                     AS paid_count,
			COUNT(CASE WHEN status='Due'     THEN 1 END)                                     AS due_count,
			COUNT(CASE WHEN status='Partial' THEN 1 END)                                     AS partial_count,
			COUNT(CASE WHEN status='Overdue' THEN 1 END)                                     AS overdue_count
		FROM payments
		WHERE property_id IN (%s)
		  AND created_at >= $%d AND created_at <= $%d
		GROUP BY property_id`,
		strings.Join(placeholders, ","), len(propertyIDs)+1, len(propertyIDs)+2)

	rows, err := r.db.QueryContext(context.Background(), q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch property payment stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int]any, len(propertyIDs))
	for rows.Next() {
		var propID int
		var row statsRow
		if err := rows.Scan(
			&propID, &row.TotalRecords, &row.TotalDue, &row.TotalPaid, &row.TotalOverdue,
			&row.PaidCount, &row.DueCount, &row.PartialCount, &row.OverdueCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan batch stats row: %w", err)
		}
		result[propID] = buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
			row.TotalDue, row.TotalPaid, row.TotalOverdue)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("batch property stats rows error: %w", err)
	}
	return result, nil
}

// GetSystemPaymentStats returns system-wide aggregate payment stats for a date range.
func (r *PaymentRepository) GetSystemPaymentStats(startDate, endDate time.Time) (interface{}, error) {
	const q = `
		SELECT
			COUNT(*) AS total_records,
			COALESCE(SUM(amount_due), 0)  AS total_due,
			COALESCE(SUM(amount_paid), 0) AS total_paid,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			COUNT(CASE WHEN status='Paid'    THEN 1 END) AS paid_count,
			COUNT(CASE WHEN status='Due'     THEN 1 END) AS due_count,
			COUNT(CASE WHEN status='Partial' THEN 1 END) AS partial_count,
			COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
		FROM payments
		WHERE created_at >= $1 AND created_at <= $2`
	var row statsRow
	if err := r.db.GetContext(context.Background(), &row, q, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get system payment stats: %w", err)
	}
	return buildPaymentStats(row.TotalRecords, row.PaidCount, row.DueCount, row.PartialCount, row.OverdueCount,
		row.TotalDue, row.TotalPaid, row.TotalOverdue), nil
}

// GetDashboardSummary returns payment summary scoped to the given organisation.
func (r *PaymentRepository) GetDashboardSummary(orgID int) (*models.DashboardSummary, error) {
	const q = `
		SELECT
			COALESCE(SUM(amount_due), 0)  AS total_due,
			COALESCE(SUM(amount_paid), 0) AS total_paid,
			COALESCE(SUM(CASE WHEN status != 'Paid' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_pending,
			COALESCE(SUM(CASE WHEN status='Overdue' THEN amount_due - amount_paid ELSE 0 END), 0) AS total_overdue,
			(SELECT COUNT(DISTINCT id) FROM properties WHERE organization_id = $1) AS property_count,
			(SELECT COUNT(DISTINCT id) FROM buildings WHERE organization_id = $1) AS building_count,
			(SELECT COUNT(DISTINCT id) FROM units     WHERE active = true AND organization_id = $1) AS unit_count,
			(SELECT COUNT(DISTINCT id) FROM tenants   WHERE active = true AND organization_id = $1) AS tenant_count
		FROM payments
		WHERE organization_id = $1`
	var row dashboardRow
	if err := r.db.GetContext(context.Background(), &row, q, orgID); err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	s := &models.DashboardSummary{
		TotalDue:      row.TotalDue,
		TotalPaid:     row.TotalPaid,
		TotalPending:  row.TotalPending,
		TotalOverdue:  row.TotalOverdue,
		PropertyCount: int(row.PropertyCount),
		BuildingCount: int(row.BuildingCount),
		UnitCount:     int(row.UnitCount),
		TenantCount:   int(row.TenantCount),
	}
	if row.TotalDue > 0 {
		s.CollectionRate = (row.TotalPaid / row.TotalDue) * 100
	}
	return s, nil
}

// GetBuildingLevelSummary returns a map with building-level aggregate totals.
func (r *PaymentRepository) GetBuildingLevelSummary() (map[string]interface{}, error) {
	const q = `
		SELECT
			COUNT(DISTINCT building_id)   AS total_buildings,
			COALESCE(SUM(amount_due), 0)  AS total_due,
			COALESCE(SUM(amount_paid), 0) AS total_paid
		FROM payments`
	var row buildingLevelRow
	if err := r.db.GetContext(context.Background(), &row, q); err != nil {
		return nil, fmt.Errorf("failed to get building level summary: %w", err)
	}
	return map[string]interface{}{
		"total_buildings": row.TotalBuildings,
		"total_due":       row.TotalDue,
		"total_paid":      row.TotalPaid,
	}, nil
}

// SearchLeases searches for active leases across tenants, properties, buildings, and units.
// Uses positional scan because LeaseSearchResult.LeaseEndDate is time.Time but the DB column is nullable.
func (r *PaymentRepository) SearchLeases(orgID int, query string) ([]*models.LeaseSearchResult, error) {
	searchPattern := "%" + query + "%"
	const q = `
		SELECT
			l.id                           AS lease_id,
			COALESCE(t.id, 0)              AS tenant_id,
			COALESCE(t.name, '')           AS tenant_name,
			COALESCE(t.phone_number, '')   AS tenant_phone,
			COALESCE(p.id, 0)             AS property_id,
			COALESCE(p.property_name, '') AS property_name,
			COALESCE(b.id, 0)             AS building_id,
			COALESCE(b.building_name, '') AS building_name,
			COALESCE(b.building_code, '') AS building_code,
			COALESCE(u.id, 0)             AS unit_id,
			COALESCE(u.unit_number, '')   AS unit_number,
			COALESCE(u.unit_type::text, '') AS unit_type,
			l.start_date                   AS lease_start_date,
			l.end_date                     AS lease_end_date,
			l.monthly_rent,
			l.active,
			COALESCE((
				SELECT SUM(pay.amount_due - pay.amount_paid)
				FROM payments pay
				WHERE pay.unit_id = u.id AND pay.status IN ('Due', 'Partial', 'Overdue')
			), 0) AS outstanding_balance
		FROM leases l
		LEFT JOIN tenants t    ON l.tenant_id = t.id
		LEFT JOIN units u      ON l.unit_id = u.id
		LEFT JOIN buildings b  ON u.building_id = b.id
		LEFT JOIN properties p ON u.property_id = p.id
		WHERE l.organization_id = $1
			AND l.active = true
			AND (
				LOWER(t.name) LIKE $2
				OR LOWER(p.property_name) LIKE $2
				OR LOWER(b.building_name) LIKE $2
				OR LOWER(b.building_code) LIKE $2
				OR LOWER(u.unit_number) LIKE $2
				OR COALESCE(LOWER(t.phone_number), '') LIKE $2
				OR CAST(l.id AS TEXT) LIKE $2
			)
		ORDER BY t.name, l.start_date DESC
		LIMIT 50`

	rows, err := r.db.QueryContext(context.Background(), q, orgID, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search leases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*models.LeaseSearchResult
	for rows.Next() {
		lr := &models.LeaseSearchResult{}
		var endDate sql.NullTime
		if err := rows.Scan(
			&lr.LeaseID, &lr.TenantID, &lr.TenantName, &lr.TenantPhone,
			&lr.PropertyID, &lr.PropertyName,
			&lr.BuildingID, &lr.BuildingName, &lr.BuildingCode,
			&lr.UnitID, &lr.UnitNumber, &lr.UnitType,
			&lr.LeaseStartDate, &endDate,
			&lr.MonthlyRent, &lr.Active,
			&lr.OutstandingBalance,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lease row: %w", err)
		}
		if endDate.Valid {
			lr.LeaseEndDate = endDate.Time
		}
		results = append(results, lr)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("lease rows error: %w", err)
	}
	return results, nil
}

// NextReceiptNumber atomically issues the next sequential receipt number for
// an organization/period, formatted as ORG<id>-<yearMonth>-<4-digit seq>.
// The sequence resets to 1 whenever yearMonth changes from the stored value.
func (r *PaymentRepository) NextReceiptNumber(orgID int, yearMonth string) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", fmt.Errorf("failed to begin receipt sequence transaction: %w", err)
	}
	// Rollback is a no-op once the transaction commits below.
	defer func() { _ = tx.Rollback() }()

	var storedYearMonth string
	var nextSeq int
	err = tx.QueryRow(
		`INSERT INTO payment_receipt_sequences (organization_id, year_month, next_seq)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (organization_id) DO UPDATE SET organization_id = payment_receipt_sequences.organization_id
		 RETURNING year_month, next_seq`,
		orgID, yearMonth,
	).Scan(&storedYearMonth, &nextSeq)
	if err != nil {
		return "", fmt.Errorf("failed to load receipt sequence: %w", err)
	}

	seq := nextSeq
	if storedYearMonth != yearMonth {
		seq = 1
	}

	if _, err := tx.Exec(
		`UPDATE payment_receipt_sequences SET year_month = $2, next_seq = $3 WHERE organization_id = $1`,
		orgID, yearMonth, seq+1,
	); err != nil {
		return "", fmt.Errorf("failed to advance receipt sequence: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit receipt sequence: %w", err)
	}

	return fmt.Sprintf("ORG%d-%s-%04d", orgID, yearMonth, seq), nil
}

// GetBuildingPaymentAnalytics returns detailed analytics for a building over a date range.
func (r *PaymentRepository) GetBuildingPaymentAnalytics(buildingID int, startDate, endDate time.Time) (*models.BuildingPaymentAnalytics, error) {
	const analyticsQ = `
		SELECT
			COALESCE(SUM(amount_paid), 0) AS total_revenue,
			CASE WHEN COALESCE(SUM(amount_due), 0) > 0
				THEN (COALESCE(SUM(amount_paid), 0) / COALESCE(SUM(amount_due), 0)) * 100
				ELSE 0 END AS collection_rate,
			COALESCE(AVG(
				CASE WHEN payment_date IS NOT NULL AND due_date IS NOT NULL
					THEN EXTRACT(EPOCH FROM (payment_date - due_date)) / 86400
					ELSE NULL END
			), 0) AS avg_payment_days,
			COUNT(CASE WHEN status='Overdue' THEN 1 END) AS overdue_count
		FROM payments
		WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3`

	var ar buildingAnalyticsRow
	if err := r.db.GetContext(context.Background(), &ar, analyticsQ, buildingID, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get building payment analytics: %w", err)
	}

	analytics := &models.BuildingPaymentAnalytics{
		BuildingID:         buildingID,
		Period:             startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalRevenue:       ar.TotalRevenue,
		CollectionRate:     ar.CollectionRate,
		AveragePaymentTime: ar.AvgPaymentDays,
		OverduePayments:    int(ar.OverdueCount),
	}

	const trendQ = `
		SELECT year, month,
			COALESCE(SUM(amount_due), 0)   AS due,
			COALESCE(SUM(amount_paid), 0)  AS paid
		FROM payments
		WHERE building_id = $1 AND created_at >= $2 AND created_at <= $3
		GROUP BY year, month
		ORDER BY year, month`

	var trend []trendPoint
	if err := r.db.SelectContext(context.Background(), &trend, trendQ, buildingID, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get building payment trend: %w", err)
	}
	analytics.TrendAnalysis = trend
	return analytics, nil
}
