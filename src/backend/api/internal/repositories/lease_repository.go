package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// LeaseRepository handles lease data operations
type LeaseRepository struct {
	db *sqlx.DB
}

// NewLeaseRepository creates a new lease repository
func NewLeaseRepository(db *sqlx.DB) *LeaseRepository {
	return &LeaseRepository{db: db}
}

// Create creates a new lease
func (r *LeaseRepository) Create(req *models.CreateLeaseRequest) (*models.Lease, error) {
	lease := &models.Lease{
		UnitID:          req.UnitID,
		TenantID:        req.TenantID,
		LeaseType:       req.LeaseType,
		DurationMonths:  req.DurationMonths,
		MonthlyRent:     req.MonthlyRent,
		SecurityDeposit: req.SecurityDeposit,
		Active:          true,
		OrganizationID:  req.OrganizationID,
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	lease.StartDate = startDate

	// Calculate end date if not provided
	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
		lease.EndDate = endDate
	} else {
		// Calculate end date from duration
		lease.EndDate = startDate.AddDate(0, req.DurationMonths, 0)
	}

	query := `
		INSERT INTO leases (unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		query,
		lease.UnitID,
		lease.TenantID,
		lease.LeaseType,
		lease.StartDate,
		lease.EndDate,
		lease.DurationMonths,
		lease.MonthlyRent,
		lease.SecurityDeposit,
		lease.Active,
		lease.OrganizationID,
		time.Now(),
		time.Now(),
	).Scan(&lease.ID, &lease.CreatedAt, &lease.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create lease: %w", err)
	}

	return lease, nil
}

// GetByID retrieves a lease by ID
func (r *LeaseRepository) GetByID(id int) (*models.Lease, error) {
	query := `
		SELECT id, unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, created_at, updated_at
		FROM leases
		WHERE id = $1
	`

	lease := &models.Lease{}
	err := r.db.QueryRow(query, id).Scan(
		&lease.ID,
		&lease.UnitID,
		&lease.TenantID,
		&lease.LeaseType,
		&lease.StartDate,
		&lease.EndDate,
		&lease.DurationMonths,
		&lease.MonthlyRent,
		&lease.SecurityDeposit,
		&lease.Active,
		&lease.OrganizationID,
		&lease.CreatedAt,
		&lease.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lease: %w", err)
	}

	return lease, nil
}

// GetByIDWithDetails retrieves a lease with full details
func (r *LeaseRepository) GetByIDWithDetails(id int) (*models.LeaseWithDetails, error) {
	query := `
		SELECT
			l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
			l.duration_months, l.monthly_rent, l.security_deposit, l.active,
			l.organization_id, l.created_at, l.updated_at,
			p.property_name, b.building_name, b.building_code,
			u.unit_number, u.unit_type,
			t.name as tenant_name, t.phone_number as tenant_phone
		FROM leases l
		JOIN units u ON l.unit_id = u.id
		JOIN buildings b ON u.building_id = b.id
		JOIN properties p ON b.property_id = p.id
		JOIN tenants t ON l.tenant_id = t.id
		WHERE l.id = $1
	`

	lease := &models.LeaseWithDetails{}
	err := r.db.QueryRow(query, id).Scan(
		&lease.ID,
		&lease.UnitID,
		&lease.TenantID,
		&lease.LeaseType,
		&lease.StartDate,
		&lease.EndDate,
		&lease.DurationMonths,
		&lease.MonthlyRent,
		&lease.SecurityDeposit,
		&lease.Active,
		&lease.OrganizationID,
		&lease.CreatedAt,
		&lease.UpdatedAt,
		&lease.PropertyName,
		&lease.BuildingName,
		&lease.BuildingCode,
		&lease.UnitNumber,
		&lease.UnitType,
		&lease.TenantName,
		&lease.TenantPhone,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lease details: %w", err)
	}

	// Calculate expiration status
	lease.IsExpired = time.Now().After(lease.EndDate)
	lease.DaysRemaining = int(time.Until(lease.EndDate).Hours() / 24)

	return lease, nil
}

// GetAll retrieves leases with pagination
func (r *LeaseRepository) GetAll(page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM leases l
		WHERE l.organization_id = $1
	`
	err := r.db.QueryRow(countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leases: %w", err)
	}

	// Get leases
	query := `
		SELECT
			l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
			l.duration_months, l.monthly_rent, l.security_deposit, l.active,
			l.organization_id, l.created_at, l.updated_at,
			p.property_name, b.building_name, b.building_code,
			u.unit_number, u.unit_type,
			t.name as tenant_name, t.phone_number as tenant_phone
		FROM leases l
		JOIN units u ON l.unit_id = u.id
		JOIN buildings b ON u.building_id = b.id
		JOIN properties p ON b.property_id = p.id
		JOIN tenants t ON l.tenant_id = t.id
		WHERE l.organization_id = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	leases := []*models.LeaseWithDetails{}
	for rows.Next() {
		lease := &models.LeaseWithDetails{}
		err := rows.Scan(
			&lease.ID,
			&lease.UnitID,
			&lease.TenantID,
			&lease.LeaseType,
			&lease.StartDate,
			&lease.EndDate,
			&lease.DurationMonths,
			&lease.MonthlyRent,
			&lease.SecurityDeposit,
			&lease.Active,
			&lease.OrganizationID,
			&lease.CreatedAt,
			&lease.UpdatedAt,
			&lease.PropertyName,
			&lease.BuildingName,
			&lease.BuildingCode,
			&lease.UnitNumber,
			&lease.UnitType,
			&lease.TenantName,
			&lease.TenantPhone,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan lease: %w", err)
		}

		// Calculate expiration status
		lease.IsExpired = time.Now().After(lease.EndDate)
		lease.DaysRemaining = int(time.Until(lease.EndDate).Hours() / 24)

		leases = append(leases, lease)
	}

	return leases, total, nil
}

// GetByUnitID retrieves leases for a specific unit
func (r *LeaseRepository) GetByUnitID(unitID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM leases l
		WHERE l.unit_id = $1 AND l.organization_id = $2
	`
	err := r.db.QueryRow(countQuery, unitID, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leases: %w", err)
	}

	// Get leases
	query := `
		SELECT
			l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
			l.duration_months, l.monthly_rent, l.security_deposit, l.active,
			l.organization_id, l.created_at, l.updated_at,
			p.property_name, b.building_name, b.building_code,
			u.unit_number, u.unit_type,
			t.name as tenant_name, t.phone_number as tenant_phone
		FROM leases l
		JOIN units u ON l.unit_id = u.id
		JOIN buildings b ON u.building_id = b.id
		JOIN properties p ON b.property_id = p.id
		JOIN tenants t ON l.tenant_id = t.id
		WHERE l.unit_id = $1 AND l.organization_id = $2
		ORDER BY l.start_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(query, unitID, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	leases := []*models.LeaseWithDetails{}
	for rows.Next() {
		lease := &models.LeaseWithDetails{}
		err := rows.Scan(
			&lease.ID,
			&lease.UnitID,
			&lease.TenantID,
			&lease.LeaseType,
			&lease.StartDate,
			&lease.EndDate,
			&lease.DurationMonths,
			&lease.MonthlyRent,
			&lease.SecurityDeposit,
			&lease.Active,
			&lease.OrganizationID,
			&lease.CreatedAt,
			&lease.UpdatedAt,
			&lease.PropertyName,
			&lease.BuildingName,
			&lease.BuildingCode,
			&lease.UnitNumber,
			&lease.UnitType,
			&lease.TenantName,
			&lease.TenantPhone,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan lease: %w", err)
		}

		// Calculate expiration status
		lease.IsExpired = time.Now().After(lease.EndDate)
		lease.DaysRemaining = int(time.Until(lease.EndDate).Hours() / 24)

		leases = append(leases, lease)
	}

	return leases, total, nil
}

// GetByTenantID retrieves leases for a specific tenant
func (r *LeaseRepository) GetByTenantID(tenantID int, page, pageSize, orgID int) ([]*models.LeaseWithDetails, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM leases l
		WHERE l.tenant_id = $1 AND l.organization_id = $2
	`
	err := r.db.QueryRow(countQuery, tenantID, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leases: %w", err)
	}

	// Get leases
	query := `
		SELECT
			l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
			l.duration_months, l.monthly_rent, l.security_deposit, l.active,
			l.organization_id, l.created_at, l.updated_at,
			p.property_name, b.building_name, b.building_code,
			u.unit_number, u.unit_type,
			t.name as tenant_name, t.phone_number as tenant_phone
		FROM leases l
		JOIN units u ON l.unit_id = u.id
		JOIN buildings b ON u.building_id = b.id
		JOIN properties p ON b.property_id = p.id
		JOIN tenants t ON l.tenant_id = t.id
		WHERE l.tenant_id = $1 AND l.organization_id = $2
		ORDER BY l.start_date DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(query, tenantID, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get leases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	leases := []*models.LeaseWithDetails{}
	for rows.Next() {
		lease := &models.LeaseWithDetails{}
		err := rows.Scan(
			&lease.ID,
			&lease.UnitID,
			&lease.TenantID,
			&lease.LeaseType,
			&lease.StartDate,
			&lease.EndDate,
			&lease.DurationMonths,
			&lease.MonthlyRent,
			&lease.SecurityDeposit,
			&lease.Active,
			&lease.OrganizationID,
			&lease.CreatedAt,
			&lease.UpdatedAt,
			&lease.PropertyName,
			&lease.BuildingName,
			&lease.BuildingCode,
			&lease.UnitNumber,
			&lease.UnitType,
			&lease.TenantName,
			&lease.TenantPhone,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan lease: %w", err)
		}

		// Calculate expiration status
		lease.IsExpired = time.Now().After(lease.EndDate)
		lease.DaysRemaining = int(time.Until(lease.EndDate).Hours() / 24)

		leases = append(leases, lease)
	}

	return leases, total, nil
}

// Update updates a lease
func (r *LeaseRepository) Update(id int, req *models.UpdateLeaseRequest) (*models.Lease, error) {
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.LeaseType != nil {
		updates = append(updates, fmt.Sprintf("lease_type = $%d", argPos))
		args = append(args, *req.LeaseType)
		argPos++
	}

	startDateArgPos := -1
	if req.StartDate != nil {
		updates = append(updates, fmt.Sprintf("start_date = $%d", argPos))
		args = append(args, *req.StartDate)
		startDateArgPos = argPos
		argPos++
	}

	if req.DurationMonths != nil {
		updates = append(updates, fmt.Sprintf("duration_months = $%d", argPos))
		args = append(args, *req.DurationMonths)
		argPos++
		// Recalculate end_date from start_date + duration
		if startDateArgPos >= 0 {
			updates = append(updates, fmt.Sprintf("end_date = $%d::date + ($%d * INTERVAL '1 month')", startDateArgPos, argPos))
		} else {
			updates = append(updates, fmt.Sprintf("end_date = start_date + ($%d * INTERVAL '1 month')", argPos))
		}
		args = append(args, *req.DurationMonths)
		argPos++
	}

	if req.MonthlyRent != nil {
		updates = append(updates, fmt.Sprintf("monthly_rent = $%d", argPos))
		args = append(args, *req.MonthlyRent)
		argPos++
	}

	if req.SecurityDeposit != nil {
		updates = append(updates, fmt.Sprintf("security_deposit = $%d", argPos))
		args = append(args, *req.SecurityDeposit)
		argPos++
	}

	if req.Active != nil {
		updates = append(updates, fmt.Sprintf("active = $%d", argPos))
		args = append(args, *req.Active)
		argPos++
	}

	if len(updates) == 0 {
		return r.GetByID(id)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE leases
		SET %s
		WHERE id = $%d
		RETURNING id, unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, created_at, updated_at
	`, strings.Join(updates, ", "), argPos)

	lease := &models.Lease{}
	err := r.db.QueryRow(query, args...).Scan(
		&lease.ID,
		&lease.UnitID,
		&lease.TenantID,
		&lease.LeaseType,
		&lease.StartDate,
		&lease.EndDate,
		&lease.DurationMonths,
		&lease.MonthlyRent,
		&lease.SecurityDeposit,
		&lease.Active,
		&lease.OrganizationID,
		&lease.CreatedAt,
		&lease.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update lease: %w", err)
	}

	return lease, nil
}

// Delete hard deletes a lease
func (r *LeaseRepository) Delete(id int) error {
	query := `DELETE FROM leases WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lease: %w", err)
	}
	return nil
}

// SoftDelete soft deletes a lease (sets active to false)
func (r *LeaseRepository) SoftDelete(id int) error {
	query := `UPDATE leases SET active = false, updated_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete lease: %w", err)
	}
	return nil
}

// HasActiveLeaseOnUnit checks if a unit has an active lease
func (r *LeaseRepository) HasActiveLeaseOnUnit(unitID int, excludeLeaseID *int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM leases
			WHERE unit_id = $1 AND active = true
		)
	`

	args := []interface{}{unitID}
	if excludeLeaseID != nil {
		query = `
			SELECT EXISTS(
				SELECT 1 FROM leases
				WHERE unit_id = $1 AND active = true AND id != $2
			)
		`
		args = append(args, *excludeLeaseID)
	}

	var exists bool
	err := r.db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active lease on unit: %w", err)
	}

	return exists, nil
}

// HasActiveLeaseForTenant checks if a tenant has an active lease
func (r *LeaseRepository) HasActiveLeaseForTenant(tenantID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM leases
			WHERE tenant_id = $1 AND active = true
		)
	`

	var exists bool
	err := r.db.QueryRow(query, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active lease for tenant: %w", err)
	}

	return exists, nil
}

// GetLeasesDueForMonth returns all leases with unpaid rent for the current month
func (r *LeaseRepository) GetLeasesDueForMonth(orgID int) ([]models.LeaseDue, error) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	firstDayOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, now.Location())

	query := `
		SELECT
			l.id as lease_id,
			l.tenant_id,
			COALESCE(t.name, '') as tenant_name,
			l.unit_id,
			u.unit_number,
			u.unit_type,
			u.building_id,
			b.building_name,
			b.building_code,
			b.property_id,
			p.property_name,
			l.monthly_rent,
			CASE
				WHEN l.start_date > $1 THEN
					EXTRACT(DAY FROM AGE(CURRENT_DATE, l.start_date))::int
				ELSE
					EXTRACT(DAY FROM AGE(CURRENT_DATE, $1))::int
			END as days_overdue,
			l.organization_id
		FROM leases l
		INNER JOIN units u ON l.unit_id = u.id
		INNER JOIN buildings b ON u.building_id = b.id
		INNER JOIN properties p ON b.property_id = p.id
		INNER JOIN tenants t ON l.tenant_id = t.id
		WHERE l.active = true
			AND l.organization_id = $2
			AND l.start_date <= CURRENT_DATE
			AND l.end_date >= $1
			AND NOT EXISTS (
				SELECT 1 FROM payments pay
				WHERE pay.unit_id = l.unit_id
					AND pay.tenant_id = l.tenant_id
					AND pay.month = $3
					AND pay.year = $4
					AND pay.status IN ('Paid', 'Partial')
			)
		ORDER BY days_overdue DESC, tenant_name ASC
	`

	rows, err := r.db.Query(query, firstDayOfMonth, orgID, currentMonth, currentYear)
	if err != nil {
		return nil, fmt.Errorf("failed to query due leases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var leasesDue []models.LeaseDue
	for rows.Next() {
		var leaseDue models.LeaseDue
		err := rows.Scan(
			&leaseDue.LeaseID,
			&leaseDue.TenantID,
			&leaseDue.TenantName,
			&leaseDue.UnitID,
			&leaseDue.UnitNumber,
			&leaseDue.UnitType,
			&leaseDue.BuildingID,
			&leaseDue.BuildingName,
			&leaseDue.BuildingCode,
			&leaseDue.PropertyID,
			&leaseDue.PropertyName,
			&leaseDue.MonthlyRent,
			&leaseDue.DaysOverdue,
			&leaseDue.OrganizationID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lease due row: %w", err)
		}
		leasesDue = append(leasesDue, leaseDue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lease due rows: %w", err)
	}

	return leasesDue, nil
}

// GetDueSummary calculates summary statistics for unpaid rent
func (r *LeaseRepository) GetDueSummary(orgID int) (*models.DueSummary, error) {
	leasesDue, err := r.GetLeasesDueForMonth(orgID)
	if err != nil {
		return nil, err
	}

	summary := &models.DueSummary{
		TotalDueAmount:  0,
		TotalTenantsDue: len(leasesDue),
	}

	for _, lease := range leasesDue {
		summary.TotalDueAmount += lease.MonthlyRent
	}

	return summary, nil
}

const leaseWithDetailsCols = `
	l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
	l.duration_months, l.monthly_rent, l.security_deposit, l.active, l.organization_id,
	l.created_at, l.updated_at,
	u.building_id, u.property_id,
	COALESCE(pr.property_name, '') AS property_name,
	COALESCE(b.building_name, '')  AS building_name,
	COALESCE(b.building_code, '')  AS building_code,
	COALESCE(u.unit_number, '')    AS unit_number,
	COALESCE(u.unit_type::text, '') AS unit_type,
	COALESCE(t.name, '')           AS tenant_name,
	COALESCE(t.phone_number, '')   AS tenant_phone`

const leaseDetailJoins = `
	LEFT JOIN units      u  ON l.unit_id   = u.id
	LEFT JOIN tenants    t  ON l.tenant_id  = t.id
	LEFT JOIN buildings  b  ON u.building_id = b.id
	LEFT JOIN properties pr ON u.property_id = pr.id`

func scanLeaseWithDetails(row interface {
	Scan(...interface{}) error
}) (*models.LeaseWithDetails, error) {
	l := &models.LeaseWithDetails{}
	err := row.Scan(
		&l.ID, &l.UnitID, &l.TenantID, &l.LeaseType,
		&l.StartDate, &l.EndDate, &l.DurationMonths, &l.MonthlyRent,
		&l.SecurityDeposit, &l.Active, &l.OrganizationID,
		&l.CreatedAt, &l.UpdatedAt,
		&l.BuildingID, &l.PropertyID,
		&l.PropertyName, &l.BuildingName, &l.BuildingCode,
		&l.UnitNumber, &l.UnitType, &l.TenantName, &l.TenantPhone,
	)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	l.IsExpired = l.EndDate.Before(now)
	if !l.IsExpired {
		l.DaysRemaining = int(time.Until(l.EndDate).Hours() / 24)
	}
	return l, nil
}

func (r *LeaseRepository) GetActiveLeases(orgID int) ([]*models.LeaseWithDetails, error) {
	query := `SELECT ` + leaseWithDetailsCols + `
		FROM leases l` + leaseDetailJoins + `
		WHERE l.active = true AND l.organization_id = $1
		ORDER BY t.name, u.unit_number`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var leases []*models.LeaseWithDetails
	for rows.Next() {
		l, err := scanLeaseWithDetails(rows)
		if err != nil {
			return nil, err
		}
		leases = append(leases, l)
	}
	return leases, rows.Err()
}
