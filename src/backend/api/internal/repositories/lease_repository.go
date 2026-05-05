package repositories

import (
	"database/sql"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

type LeaseRepository struct {
	db *sql.DB
}

func NewLeaseRepository(db *sql.DB) *LeaseRepository {
	return &LeaseRepository{db: db}
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
	defer rows.Close()

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
