package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/models/columns"
)

// TenantRepository implements the TenantRepositoryInterface
type TenantRepository struct {
	db  *sqlx.DB
	nid *appcrypto.NIDProtector
}

// NewTenantRepository creates a new TenantRepository. The NID protector encrypts
// the NID at rest and derives its lookup hash and last-four projection.
func NewTenantRepository(db *sqlx.DB, nid *appcrypto.NIDProtector) *TenantRepository {
	return &TenantRepository{db: db, nid: nid}
}

// Create creates a new tenant, storing the NID as ciphertext plus a deterministic
// hash and a last-four projection — never as plaintext.
func (r *TenantRepository) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	encrypted, err := r.nid.Encrypt(req.NIDNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt NID: %w", err)
	}
	lastFour := appcrypto.LastFour(req.NIDNumber)
	nidHash := r.nid.Hash(req.NIDNumber)

	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s, %s, %s, %s, %s, %s, organization_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9)
		RETURNING %s, %s, %s`,
		columns.TenantTable,
		columns.TenantName, columns.TenantType, columns.TenantPhoneNumber,
		columns.TenantEmail, columns.TenantNIDEncrypted, columns.TenantNIDLastFour,
		columns.TenantNIDHash, columns.TenantAddress, columns.TenantActive,
		columns.TenantID, columns.TenantCreatedAt, columns.TenantUpdatedAt)

	tenant := &models.Tenant{
		Name:        req.Name,
		TenantType:  req.TenantType,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		NIDNumber:   req.NIDNumber,
		NIDLastFour: lastFour,
		Address:     req.Address,
		Active:      true,
	}

	err = r.db.QueryRow(
		query,
		tenant.Name,
		tenant.TenantType,
		tenant.PhoneNumber,
		tenant.Email,
		encrypted,
		lastFour,
		nidHash,
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

// CheckNIDExists checks if a NID number already exists for another tenant. The
// lookup runs against the deterministic hash column, so no plaintext NID is
// stored or compared.
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
		columns.TenantNIDHash, columns.TenantID, columns.TenantActive)

	var exists bool
	err := r.db.QueryRow(query, r.nid.Hash(nid), excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check NID existence: %w", err)
	}

	return exists, nil
}

// GetByID retrieves a tenant by ID. The returned tenant carries only the NID
// last-four; the full value is never decrypted here.
func (r *TenantRepository) GetByID(id int) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1 AND %s = true`,
		columns.TenantAllColumns(),
		columns.TenantTable,
		columns.TenantID, columns.TenantActive)

	tenant := &models.Tenant{}
	var lastFour sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.TenantType,
		&tenant.PhoneNumber,
		&tenant.Email,
		&lastFour,
		&tenant.Address,
		&tenant.Active,
		&tenant.OrganizationID,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	tenant.NIDLastFour = lastFour.String

	return tenant, nil
}

// GetDecryptedNID returns the full, decrypted NID for a tenant along with its
// organization ID (for ownership verification). This is the only read path that
// materializes the plaintext NID and must be gated by the caller.
func (r *TenantRepository) GetDecryptedNID(id int) (string, int, error) {
	query := fmt.Sprintf(`
		SELECT %s, organization_id
		FROM %s
		WHERE %s = $1 AND %s = true`,
		columns.TenantNIDEncrypted,
		columns.TenantTable,
		columns.TenantID, columns.TenantActive)

	var encrypted sql.NullString
	var orgID int
	err := r.db.QueryRow(query, id).Scan(&encrypted, &orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("tenant not found")
		}
		return "", 0, fmt.Errorf("failed to get tenant NID: %w", err)
	}

	plaintext, err := r.nid.Decrypt(encrypted.String)
	if err != nil {
		return "", 0, fmt.Errorf("failed to decrypt NID: %w", err)
	}

	return plaintext, orgID, nil
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
	defer func() { _ = rows.Close() }()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		var lastFour sql.NullString
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.TenantType,
			&tenant.PhoneNumber,
			&tenant.Email,
			&lastFour,
			&tenant.Address,
			&tenant.Active,
			&tenant.OrganizationID,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenant.NIDLastFour = lastFour.String
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating tenants: %w", err)
	}

	return tenants, totalCount, nil
}

// Update updates a tenant. A "nid_number" key in the update map is transparently
// re-mapped to the encrypted/hash/last-four columns so callers never write
// plaintext NID to the database.
func (r *TenantRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build SET clause
	setClauses := []string{}
	args := []interface{}{}
	argPos := 1

	for column, value := range updates {
		if column == columns.TenantNIDNumber {
			nid, _ := value.(string)
			encrypted, err := r.nid.Encrypt(nid)
			if err != nil {
				return fmt.Errorf("failed to encrypt NID: %w", err)
			}
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columns.TenantNIDEncrypted, argPos))
			args = append(args, encrypted)
			argPos++
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columns.TenantNIDLastFour, argPos))
			args = append(args, appcrypto.LastFour(nid))
			argPos++
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", columns.TenantNIDHash, argPos))
			args = append(args, r.nid.Hash(nid))
			argPos++
			continue
		}
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
