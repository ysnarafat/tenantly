package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

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

// scanPaymentWithDetails scans a row into PaymentWithDetails
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

// Update applies a partial update to a payment and auto-calculates status when amount_paid changes
// This is kept as raw SQL because it builds a dynamic SET clause and auto-derives status from amount_paid
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
// This is kept as raw SQL because it supports 6 optional dynamic filters
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

// GetActiveLeasesForPeriod returns all active leases whose coverage overlaps with the given month/year.
// Optionally scoped to a single building when buildingID is non-nil.
func (r *PaymentRepository) GetActiveLeasesForPeriod(orgID, month, year int, buildingID *int) ([]*models.LeaseSearchResult, error) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	args := []interface{}{orgID, firstDay, lastDay}
	where := `l.organization_id = $1
		AND l.active = true
		AND l.start_date <= $3
		AND (l.end_date IS NULL OR l.end_date >= $2)`

	if buildingID != nil {
		args = append(args, *buildingID)
		where += fmt.Sprintf(" AND u.building_id = $%d", len(args))
	}

	query := fmt.Sprintf(`
		SELECT
			l.id AS lease_id,
			COALESCE(t.id, 0) AS tenant_id,
			COALESCE(t.name, '') AS tenant_name,
			COALESCE(t.phone_number, '') AS tenant_phone,
			COALESCE(u.property_id, 0) AS property_id,
			COALESCE(pr.property_name, '') AS property_name,
			COALESCE(u.building_id, 0) AS building_id,
			COALESCE(b.building_name, '') AS building_name,
			COALESCE(b.building_code, '') AS building_code,
			l.unit_id,
			COALESCE(u.unit_number, '') AS unit_number,
			COALESCE(u.unit_type::text, '') AS unit_type,
			l.start_date AS lease_start_date,
			l.end_date AS lease_end_date,
			l.monthly_rent,
			l.active
		FROM leases l
		LEFT JOIN tenants t ON l.tenant_id = t.id
		LEFT JOIN units u ON l.unit_id = u.id
		LEFT JOIN buildings b ON u.building_id = b.id
		LEFT JOIN properties pr ON u.property_id = pr.id
		WHERE %s
		ORDER BY b.building_name, u.unit_number`, where)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active leases for period: %w", err)
	}
	defer rows.Close()

	var results []*models.LeaseSearchResult
	for rows.Next() {
		lr := &models.LeaseSearchResult{}
		var endDate sql.NullTime
		err := rows.Scan(
			&lr.LeaseID, &lr.TenantID, &lr.TenantName, &lr.TenantPhone,
			&lr.PropertyID, &lr.PropertyName,
			&lr.BuildingID, &lr.BuildingName, &lr.BuildingCode,
			&lr.UnitID, &lr.UnitNumber, &lr.UnitType,
			&lr.LeaseStartDate, &endDate,
			&lr.MonthlyRent, &lr.Active,
		)
		if err != nil {
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

// CheckPaymentExists returns true when a payment record already exists for the given unit/month/year.
func (r *PaymentRepository) CheckPaymentExists(unitID, month, year int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM payments WHERE unit_id = $1 AND month = $2 AND year = $3)`,
		unitID, month, year,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check payment existence: %w", err)
	}
	return exists, nil
}
