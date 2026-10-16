package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ysnarafat/tenantly/internal/models"
)

// isUniqueViolation reports whether err is a Postgres unique-constraint
// violation (SQLSTATE 23505) — used to translate the idx_leases_unit_active_unique
// index's rejection into the same friendly error LeaseService.CreateLease's
// own (non-atomic, TOCTOU-able) HasActiveLeaseOnUnit check normally produces,
// for the rare case where a concurrent request wins that race.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

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
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("unit already has an active lease")
		}
		return nil, fmt.Errorf("failed to create lease: %w", err)
	}

	return lease, nil
}

// GetByID retrieves a lease by ID
func (r *LeaseRepository) GetByID(id int) (*models.Lease, error) {
	query := `
		SELECT id, unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, end_reason, renewed_from_lease_id, created_at, updated_at
		FROM leases
		WHERE id = $1
	`

	lease := &models.Lease{}
	var endReason sql.NullString
	var renewedFromLeaseID sql.NullInt64
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
		&endReason,
		&renewedFromLeaseID,
		&lease.CreatedAt,
		&lease.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lease: %w", err)
	}
	applyLeaseEndMetadata(lease, endReason, renewedFromLeaseID)

	return lease, nil
}

// applyLeaseEndMetadata assigns the nullable end_reason/renewed_from_lease_id
// columns onto a Lease, leaving the fields nil when the column was NULL.
func applyLeaseEndMetadata(lease *models.Lease, endReason sql.NullString, renewedFromLeaseID sql.NullInt64) {
	if endReason.Valid {
		reason := models.LeaseEndReason(endReason.String)
		lease.EndReason = &reason
	}
	if renewedFromLeaseID.Valid {
		id := int(renewedFromLeaseID.Int64)
		lease.RenewedFromLeaseID = &id
	}
}

// GetByIDWithDetails retrieves a lease with full details
func (r *LeaseRepository) GetByIDWithDetails(id int) (*models.LeaseWithDetails, error) {
	query := `
		SELECT
			l.id, l.unit_id, l.tenant_id, l.lease_type, l.start_date, l.end_date,
			l.duration_months, l.monthly_rent, l.security_deposit, l.active,
			l.organization_id, l.end_reason, l.renewed_from_lease_id, l.created_at, l.updated_at,
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
	var endReason sql.NullString
	var renewedFromLeaseID sql.NullInt64
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
		&endReason,
		&renewedFromLeaseID,
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
	applyLeaseEndMetadata(&lease.Lease, endReason, renewedFromLeaseID)

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

	startDateChanged := req.StartDate != nil
	if startDateChanged {
		updates = append(updates, fmt.Sprintf("start_date = $%d", argPos))
		args = append(args, *req.StartDate)
		argPos++
	}

	durationChanged := req.DurationMonths != nil
	if durationChanged {
		updates = append(updates, fmt.Sprintf("duration_months = $%d", argPos))
		args = append(args, *req.DurationMonths)
		argPos++
	}

	if req.EndDate != nil {
		// An explicit end_date always wins — it's the one way to correct a
		// lease onto a date that isn't a whole-month offset from start_date,
		// which duration_months (an integer column) can never exactly express.
		updates = append(updates, fmt.Sprintf("end_date = $%d::date", argPos))
		args = append(args, *req.EndDate)
		argPos++
	} else if startDateChanged || durationChanged {
		// Otherwise end_date must be recalculated whenever EITHER start_date
		// or duration_months changes — not just when both change together.
		// Previously this only fired inside the DurationMonths branch, so
		// correcting just the start_date (e.g. fixing a data-entry mistake)
		// silently left end_date stale/inconsistent with the new start_date.
		//
		// Each value gets its own fresh placeholder here rather than reusing
		// the one bound above (even though it's the same value) — Postgres
		// infers a single type per placeholder across the whole query, and
		// reusing $N for both an `integer` column assignment and an
		// `integer * INTERVAL` expression trips "inconsistent types deduced"
		// (42P08).
		startExpr := "start_date"
		if startDateChanged {
			startExpr = fmt.Sprintf("$%d::date", argPos)
			args = append(args, *req.StartDate)
			argPos++
		}
		durationExpr := "duration_months"
		if durationChanged {
			durationExpr = fmt.Sprintf("$%d", argPos)
			args = append(args, *req.DurationMonths)
			argPos++
		}
		updates = append(updates, fmt.Sprintf("end_date = %s + (%s * INTERVAL '1 month')", startExpr, durationExpr))
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

// SoftDelete ends a lease early (tenant moved out before the natural end
// date) — sets active to false and records the real move-out date and reason,
// so the lease stays an accurate historical record instead of silently
// keeping its original, now-inaccurate end_date.
func (r *LeaseRepository) SoftDelete(id int, endDate time.Time, reason models.LeaseEndReason) error {
	query := `UPDATE leases SET active = false, end_date = $1, end_reason = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.Exec(query, endDate, reason, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete lease: %w", err)
	}
	return nil
}

// RenewLease starts a new lease term for the same unit/tenant and closes out
// the lease being renewed, atomically, so a renewal never leaves the unit
// with zero or two active leases if either half fails.
func (r *LeaseRepository) RenewLease(oldLeaseID int, req *models.RenewLeaseRequest) (*models.Lease, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin lease renewal transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var old models.Lease
	err = tx.QueryRow(
		`SELECT id, unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id
		 FROM leases WHERE id = $1 FOR UPDATE`,
		oldLeaseID,
	).Scan(
		&old.ID, &old.UnitID, &old.TenantID, &old.LeaseType, &old.StartDate, &old.EndDate,
		&old.DurationMonths, &old.MonthlyRent, &old.SecurityDeposit, &old.Active, &old.OrganizationID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load lease to renew: %w", err)
	}
	if !old.Active {
		return nil, fmt.Errorf("only an active lease can be renewed")
	}

	// New term starts where the old one's coverage ends, unless the caller
	// gives an explicit (early/late) renewal date.
	newStart := old.EndDate
	if req.StartDate != nil {
		newStart, err = time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
	}
	if newStart.Before(old.StartDate) {
		return nil, fmt.Errorf("renewal start date cannot be before the original lease's start date")
	}
	newEnd := newStart.AddDate(0, req.DurationMonths, 0)

	monthlyRent := old.MonthlyRent
	if req.MonthlyRent != nil {
		monthlyRent = *req.MonthlyRent
	}
	securityDeposit := old.SecurityDeposit
	if req.SecurityDeposit != nil {
		securityDeposit = *req.SecurityDeposit
	}
	leaseType := old.LeaseType
	if req.LeaseType != nil {
		leaseType = *req.LeaseType
	}

	// Close out the lease being renewed — its coverage now accurately ends
	// where the new term begins, whether that's on-time, early, or late.
	if _, err := tx.Exec(
		`UPDATE leases SET active = false, end_date = $1, end_reason = $2, updated_at = $3 WHERE id = $4`,
		newStart, models.LeaseEndReasonRenewed, time.Now(), oldLeaseID,
	); err != nil {
		return nil, fmt.Errorf("failed to close out renewed lease: %w", err)
	}

	newLease := &models.Lease{
		UnitID:             old.UnitID,
		TenantID:           old.TenantID,
		LeaseType:          leaseType,
		StartDate:          newStart,
		EndDate:            newEnd,
		DurationMonths:     req.DurationMonths,
		MonthlyRent:        monthlyRent,
		SecurityDeposit:    securityDeposit,
		Active:             true,
		OrganizationID:     old.OrganizationID,
		RenewedFromLeaseID: &oldLeaseID,
	}

	err = tx.QueryRow(
		`INSERT INTO leases (unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, renewed_from_lease_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, created_at, updated_at`,
		newLease.UnitID, newLease.TenantID, newLease.LeaseType, newLease.StartDate, newLease.EndDate,
		newLease.DurationMonths, newLease.MonthlyRent, newLease.SecurityDeposit, newLease.Active,
		newLease.OrganizationID, oldLeaseID, time.Now(), time.Now(),
	).Scan(&newLease.ID, &newLease.CreatedAt, &newLease.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create renewed lease: %w", err)
	}

	// Carry the old lease's active recurring charges forward by default —
	// a landlord renewing a lease almost always keeps the same utility/
	// service charges unless they explicitly change them.
	carryForwardCharges := req.CarryForwardCharges == nil || *req.CarryForwardCharges
	if carryForwardCharges {
		if _, err := tx.Exec(
			`INSERT INTO lease_charges (lease_id, charge_type, label, amount, active, created_at, updated_at)
			 SELECT $1, charge_type, label, amount, active, $2, $2
			 FROM lease_charges WHERE lease_id = $3 AND active = true`,
			newLease.ID, time.Now(), oldLeaseID,
		); err != nil {
			return nil, fmt.Errorf("failed to carry forward lease charges: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit lease renewal: %w", err)
	}

	return newLease, nil
}

// ReplaceTenant performs a tenant turnover on a unit in a single transaction:
// the outgoing lease is closed at handoverDate and the successor lease is
// inserted for the incoming tenant.
//
// Both statements must share one transaction. The service's "unit already has an
// active lease" guard would reject the successor while the outgoing lease is
// still active, so there is no valid intermediate state to expose — and a
// failure between the two would otherwise leave a unit either double-leased or
// silently vacant.
//
// The UPDATE is guarded on active = true so two concurrent turnovers cannot both
// succeed: the loser affects zero rows and rolls back.
func (r *LeaseRepository) ReplaceTenant(oldLeaseID int, handoverDate time.Time, successor *models.CreateLeaseRequest) (*models.Lease, error) {
	startDate, err := time.Parse("2006-01-02", successor.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	endDate := startDate.AddDate(0, successor.DurationMonths, 0)
	if successor.EndDate != nil {
		endDate, err = time.Parse("2006-01-02", *successor.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()

	result, err := tx.Exec(
		`UPDATE leases SET active = false, end_date = $1, end_reason = $2, updated_at = $3 WHERE id = $4 AND active = true`,
		handoverDate, models.LeaseEndReasonTerminated, now, oldLeaseID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to close outgoing lease: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to confirm outgoing lease closure: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("lease is no longer active")
	}

	lease := &models.Lease{
		UnitID:          successor.UnitID,
		TenantID:        successor.TenantID,
		LeaseType:       successor.LeaseType,
		StartDate:       startDate,
		EndDate:         endDate,
		DurationMonths:  successor.DurationMonths,
		MonthlyRent:     successor.MonthlyRent,
		SecurityDeposit: successor.SecurityDeposit,
		Active:          true,
		OrganizationID:  successor.OrganizationID,
	}

	err = tx.QueryRow(
		`INSERT INTO leases (unit_id, tenant_id, lease_type, start_date, end_date, duration_months, monthly_rent, security_deposit, active, organization_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, created_at, updated_at`,
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
		now,
		now,
	).Scan(&lease.ID, &lease.CreatedAt, &lease.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create successor lease: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tenant replacement: %w", err)
	}

	return lease, nil
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

// HasPayableLeaseForUnitAndTenant reports whether the given tenant currently
// holds an active lease on the given unit whose end_date has not yet passed.
// A lease that is active=true but past its end_date (i.e. staff has not yet
// terminated/renewed it) does NOT count — such a lease is expired for
// payment purposes even though the active flag hasn't been flipped. A lease
// that is merely "expiring soon" (active, end_date in the future) DOES count.
func (r *LeaseRepository) HasPayableLeaseForUnitAndTenant(unitID, tenantID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM leases
			WHERE unit_id = $1 AND tenant_id = $2 AND active = true AND end_date >= CURRENT_DATE
		)
	`

	var exists bool
	err := r.db.QueryRow(query, unitID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check payable lease for unit and tenant: %w", err)
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
