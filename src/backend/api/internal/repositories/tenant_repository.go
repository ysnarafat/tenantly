package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/models/columns"
)

// TenantRepository implements the TenantRepositoryInterface
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new TenantRepository
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s, %s, %s, %s, organization_id
		) VALUES ($1, $2, $3, $4, $5, $6, true, $7)
		RETURNING %s, %s, %s`,
		columns.TenantTable,
		columns.TenantName, columns.TenantType, columns.TenantPhoneNumber,
		columns.TenantEmail, columns.TenantNIDNumber, columns.TenantAddress, columns.TenantActive,
		columns.TenantID, columns.TenantCreatedAt, columns.TenantUpdatedAt)

	tenant := &models.Tenant{
		Name:        req.Name,
		TenantType:  req.TenantType,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		NIDNumber:   req.NIDNumber,
		Address:     req.Address,
		Active:      true,
	}

	err := r.db.QueryRow(
		query,
		tenant.Name,
		tenant.TenantType,
		tenant.PhoneNumber,
		tenant.Email,
		tenant.NIDNumber,
		tenant.Address,
		req.OrganizationID,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// CheckEmailExists checks if an email already exists for another tenant
func (r *TenantRepository) CheckEmailExists(email string, excludeID int) (bool, error) {
	if email == "" {
		return false, nil
	}

	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s 
			WHERE %s = $1 AND %s != $2 AND %s = true
		)`,
		columns.TenantTable,
		columns.TenantEmail, columns.TenantID, columns.TenantActive)

	var exists bool
	err := r.db.QueryRow(query, email, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// CheckNIDExists checks if a NID number already exists for another tenant
func (r *TenantRepository) CheckNIDExists(nid string, excludeID int) (bool, error) {
	if nid == "" {
		return false, nil
	}

	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s 
			WHERE %s = $1 AND %s != $2 AND %s = true
		)`,
		columns.TenantTable,
		columns.TenantNIDNumber, columns.TenantID, columns.TenantActive)

	var exists bool
	err := r.db.QueryRow(query, nid, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check NID existence: %w", err)
	}

	return exists, nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(id int) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1 AND %s = true`,
		columns.TenantAllColumns(),
		columns.TenantTable,
		columns.TenantID, columns.TenantActive)

	tenant := &models.Tenant{}
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.TenantType,
		&tenant.PhoneNumber,
		&tenant.Email,
		&tenant.NIDNumber,
		&tenant.Address,
		&tenant.Active,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetByUnitID retrieves a tenant by unit ID (will be implemented with lease functionality)
func (r *TenantRepository) GetByUnitID(unitID int) (*models.Tenant, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// GetAll retrieves all active tenants with pagination
func (r *TenantRepository) GetAll(page, pageSize, orgID int) ([]*models.Tenant, int, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Get total count scoped to org
	var totalCount int
	err := r.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = true AND organization_id = $1", columns.TenantTable, columns.TenantActive),
		orgID,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = true AND organization_id = $1
		ORDER BY %s DESC
		LIMIT $2 OFFSET $3`,
		columns.TenantAllColumns(),
		columns.TenantTable,
		columns.TenantActive,
		columns.TenantCreatedAt)

	rows, err := r.db.Query(query, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.TenantType,
			&tenant.PhoneNumber,
			&tenant.Email,
			&tenant.NIDNumber,
			&tenant.Address,
			&tenant.Active,
			&tenant.OrganizationID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, totalCount, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build SET clause
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	for column, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argPos))
		args = append(args, value)
		argPos++
	}

	// Add updated_at timestamp
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now())
	argPos++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE tenants
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argPos)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	return nil
}
