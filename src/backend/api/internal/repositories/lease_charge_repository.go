package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// LeaseChargeRepository handles CRUD for recurring per-lease charges.
type LeaseChargeRepository struct {
	db *sqlx.DB
}

// NewLeaseChargeRepository creates a new lease charge repository
func NewLeaseChargeRepository(db *sqlx.DB) *LeaseChargeRepository {
	return &LeaseChargeRepository{db: db}
}

const leaseChargeCols = `id, lease_id, charge_type, label, amount, active, created_at, updated_at`

// Create adds a new recurring charge to a lease.
func (r *LeaseChargeRepository) Create(leaseID int, req *models.CreateLeaseChargeRequest) (*models.LeaseCharge, error) {
	charge := &models.LeaseCharge{}
	query := `
		INSERT INTO lease_charges (lease_id, charge_type, label, amount, active)
		VALUES ($1, $2, $3, $4, true)
		RETURNING ` + leaseChargeCols

	err := r.db.QueryRow(query, leaseID, req.ChargeType, req.Label, req.Amount).Scan(
		&charge.ID, &charge.LeaseID, &charge.ChargeType, &charge.Label, &charge.Amount,
		&charge.Active, &charge.CreatedAt, &charge.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease charge: %w", err)
	}
	return charge, nil
}

// GetByID retrieves a single charge by ID.
func (r *LeaseChargeRepository) GetByID(id int) (*models.LeaseCharge, error) {
	charge := &models.LeaseCharge{}
	query := `SELECT ` + leaseChargeCols + ` FROM lease_charges WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&charge.ID, &charge.LeaseID, &charge.ChargeType, &charge.Label, &charge.Amount,
		&charge.Active, &charge.CreatedAt, &charge.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease charge not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lease charge: %w", err)
	}
	return charge, nil
}

// GetByLeaseID returns every charge (active or not) for a lease, most
// recently created first.
func (r *LeaseChargeRepository) GetByLeaseID(leaseID int) ([]*models.LeaseCharge, error) {
	query := `SELECT ` + leaseChargeCols + ` FROM lease_charges WHERE lease_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease charges: %w", err)
	}
	defer func() { _ = rows.Close() }()

	charges := make([]*models.LeaseCharge, 0)
	for rows.Next() {
		charge := &models.LeaseCharge{}
		if err := rows.Scan(
			&charge.ID, &charge.LeaseID, &charge.ChargeType, &charge.Label, &charge.Amount,
			&charge.Active, &charge.CreatedAt, &charge.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lease charge: %w", err)
		}
		charges = append(charges, charge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lease charge rows error: %w", err)
	}
	return charges, nil
}

// Update applies a partial update to a charge.
func (r *LeaseChargeRepository) Update(id int, req *models.UpdateLeaseChargeRequest) (*models.LeaseCharge, error) {
	setClauses := []string{"updated_at = NOW() AT TIME ZONE 'UTC'"}
	args := []interface{}{}
	argIdx := 1

	if req.ChargeType != nil {
		setClauses = append(setClauses, fmt.Sprintf("charge_type = $%d", argIdx))
		args = append(args, *req.ChargeType)
		argIdx++
	}
	if req.Label != nil {
		setClauses = append(setClauses, fmt.Sprintf("label = $%d", argIdx))
		args = append(args, *req.Label)
		argIdx++
	}
	if req.Amount != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *req.Amount)
		argIdx++
	}
	if req.Active != nil {
		setClauses = append(setClauses, fmt.Sprintf("active = $%d", argIdx))
		args = append(args, *req.Active)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE lease_charges SET %s WHERE id = $%d RETURNING `+leaseChargeCols,
		strings.Join(setClauses, ", "), argIdx,
	)

	charge := &models.LeaseCharge{}
	err := r.db.QueryRow(query, args...).Scan(
		&charge.ID, &charge.LeaseID, &charge.ChargeType, &charge.Label, &charge.Amount,
		&charge.Active, &charge.CreatedAt, &charge.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease charge not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update lease charge: %w", err)
	}
	return charge, nil
}

// Delete removes a charge outright — charges have no payment history of
// their own (unlike leases), so there's no historical record to preserve by
// soft-deleting instead.
func (r *LeaseChargeRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM lease_charges WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete lease charge: %w", err)
	}
	return nil
}

// SumActiveChargesByLeaseID totals a lease's active recurring charges.
func (r *LeaseChargeRepository) SumActiveChargesByLeaseID(leaseID int) (float64, error) {
	var total float64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM lease_charges WHERE lease_id = $1 AND active = true`,
		leaseID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to sum lease charges: %w", err)
	}
	return total, nil
}
