package repositories

import (
	"database/sql"
	"fmt"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/models/columns"
)

// TenantRepository implements the TenantRepositoryInterface
type TenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new TenantRepository
func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant
func (r *TenantRepository) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s, %s, %s, %s
		) VALUES ($1, $2, $3, $4, $5, $6, true)
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
