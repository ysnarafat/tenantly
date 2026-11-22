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

// Create creates a new tenant and its isolated PII row in tenant_pii.
func (r *TenantRepository) Create(req *models.CreateTenantRequest) (*models.Tenant, error) {
	encrypted, err := r.nid.Encrypt(req.NIDNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt NID: %w", err)
	}
	lastFour := appcrypto.LastFour(req.NIDNumber)
	nidHash := r.nid.Hash(req.NIDNumber)

	encryptedPhone, err := r.nid.Encrypt(req.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt phone: %w", err)
	}
	phoneHash := r.nid.Hash(req.PhoneNumber)

	encryptedEmail, err := r.nid.Encrypt(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt email: %w", err)
	}
	emailHash := r.nid.Hash(req.Email)

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin tenant create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var tenantID int
	err = tx.QueryRow(
		fmt.Sprintf("INSERT INTO %s (name, tenant_type, address, active, organization_id) VALUES ($1, $2, $3, $4, $5) RETURNING id",
			columns.TenantTable),
		req.Name, req.TenantType, req.Address, true, req.OrganizationID,
	).Scan(&tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO tenant_pii (
			tenant_id, phone_number, phone_encrypted, phone_hash,
			email, email_encrypted, email_hash,
			nid_number, nid_encrypted, nid_last_four, nid_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		tenantID, req.PhoneNumber, encryptedPhone, phoneHash, req.Email, encryptedEmail, emailHash,
		req.NIDNumber, encrypted, lastFour, nidHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant pii: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tenant create: %w", err)
	}

	tenant := &models.Tenant{
		ID:             tenantID,
		Name:           req.Name,
		TenantType:     req.TenantType,
		PhoneNumber:    req.PhoneNumber,
		Email:          req.Email,
		NIDNumber:      req.NIDNumber,
		NIDLastFour:    lastFour,
		Address:        req.Address,
		Active:         true,
		OrganizationID: req.OrganizationID,
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
			WHERE %s = $1 AND %s != $2 AND %s IS NOT NULL
		)`,
		columns.TenantPiiTable,
		columns.TenantPiiEmailHash, columns.TenantPiiTenantID, columns.TenantPiiEmailHash,
	)

	var exists bool
	err := r.db.QueryRow(query, r.nid.Hash(email), excludeID).Scan(&exists)
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
			WHERE %s = $1 AND %s != $2 AND %s IS NOT NULL
		)`,
		columns.TenantPiiTable,
		columns.TenantPiiNIDHash, columns.TenantPiiTenantID, columns.TenantPiiNIDHash,
	)

	var exists bool
	err := r.db.QueryRow(query, r.nid.Hash(nid), excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check NID existence: %w", err)
	}

	return exists, nil
}

// GetByID retrieves a tenant by ID, joining tenant_pii for PII. It filters on
// the tenants.active flag.
func (r *TenantRepository) GetByID(id int) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT t.id, t.name, t.tenant_type, t.address, t.active, t.organization_id, t.created_at, t.updated_at,
		       p.phone_number, p.phone_encrypted, p.phone_hash, p.email, p.email_encrypted, p.email_hash,
		       p.nid_number, p.nid_encrypted, p.nid_last_four, p.nid_hash
		FROM %s t
		LEFT JOIN %s p ON p.tenant_id = t.id
		WHERE t.id = $1 AND t.active = true`,
		columns.TenantTable, columns.TenantPiiTable)

	tenant := &models.Tenant{}
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.TenantType, &tenant.Address, &tenant.Active,
		&tenant.OrganizationID, &tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.PhoneNumber, &tenant.PhoneEncrypted, &tenant.PhoneHash, &tenant.Email,
		&tenant.EmailEncrypted, &tenant.EmailHash, &tenant.NIDNumber, &tenant.NIDEncrypted,
		&tenant.NIDLastFour, &tenant.NIDHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// GetByIDIncludingInactive fetches a tenant by ID without filtering on active
// status — used to read back a tenant immediately after an update that may have
// just deactivated it, where GetByID's active-only filter would otherwise report
// a legitimately-updated tenant as "not found".
func (r *TenantRepository) GetByIDIncludingInactive(id int) (*models.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT t.id, t.name, t.tenant_type, t.address, t.active, t.organization_id, t.created_at, t.updated_at,
		       p.phone_number, p.phone_encrypted, p.phone_hash, p.email, p.email_encrypted, p.email_hash,
		       p.nid_number, p.nid_encrypted, p.nid_last_four, p.nid_hash
		FROM %s t
		LEFT JOIN %s p ON p.tenant_id = t.id
		WHERE t.id = $1`,
		columns.TenantTable, columns.TenantPiiTable)

	tenant := &models.Tenant{}
	err := r.db.QueryRow(query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.TenantType, &tenant.Address, &tenant.Active,
		&tenant.OrganizationID, &tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.PhoneNumber, &tenant.PhoneEncrypted, &tenant.PhoneHash, &tenant.Email,
		&tenant.EmailEncrypted, &tenant.EmailHash, &tenant.NIDNumber, &tenant.NIDEncrypted,
		&tenant.NIDLastFour, &tenant.NIDHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// GetDecryptedNID returns the full, decrypted NID for a tenant along with its
// organization ID (for ownership verification). This is the only read path that
// materializes the plaintext NID and must be gated by the caller.
func (r *TenantRepository) GetDecryptedNID(id int) (string, int, error) {
	query := fmt.Sprintf(`
		SELECT t.organization_id, p.nid_encrypted
		FROM %s t
		JOIN %s p ON p.tenant_id = t.id
		WHERE t.id = $1`,
		columns.TenantTable, columns.TenantPiiTable)

	var orgID int
	var encrypted sql.NullString
	err := r.db.QueryRow(query, id).Scan(&orgID, &encrypted)
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
	return nil, fmt.Errorf("GetByUnitID not yet implemented")
}

// GetAll retrieves all active tenants with pagination, joining tenant_pii for PII.
func (r *TenantRepository) GetAll(page, pageSize, orgID int) ([]*models.Tenant, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int
	err := r.db.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s t WHERE t.active = true AND t.organization_id = $1",
			columns.TenantTable),
		orgID,
	).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.name, t.tenant_type, t.address, t.active, t.organization_id, t.created_at, t.updated_at,
		       p.phone_number, p.phone_encrypted, p.phone_hash, p.email, p.email_encrypted, p.email_hash,
		       p.nid_number, p.nid_encrypted, p.nid_last_four, p.nid_hash
		FROM %s t
		LEFT JOIN %s p ON p.tenant_id = t.id
		WHERE t.active = true AND t.organization_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3`,
		columns.TenantTable, columns.TenantPiiTable)

	rows, err := r.db.Query(query, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all tenants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tenants []*models.Tenant
	for rows.Next() {
		tenant := &models.Tenant{}
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.TenantType, &tenant.Address, &tenant.Active,
			&tenant.OrganizationID, &tenant.CreatedAt, &tenant.UpdatedAt,
			&tenant.PhoneNumber, &tenant.PhoneEncrypted, &tenant.PhoneHash, &tenant.Email,
			&tenant.EmailEncrypted, &tenant.EmailHash, &tenant.NIDNumber, &tenant.NIDEncrypted,
			&tenant.NIDLastFour, &tenant.NIDHash,
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

// Update updates a tenant. Tenant fields go to tenants; PII fields go to
// tenant_pii. A "nid_number" key in the update map is transparently re-mapped to
// the tenant_pii encrypted/hash/last-four columns; a "phone_number"/"email" key
// updates the plaintext copy plus its encrypted/hash pair.
func (r *TenantRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin tenant update transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tenantArgs := []interface{}{}
	tenantCols := []string{}
	piiArgs := []interface{}{}
	piiCols := []string{}
	argPos := 1

	addTenant := func(col string, val interface{}) {
		tenantCols = append(tenantCols, fmt.Sprintf("%s = $%d", col, argPos))
		tenantArgs = append(tenantArgs, val)
		argPos++
	}
	addPii := func(col string, val interface{}) {
		piiCols = append(piiCols, fmt.Sprintf("%s = $%d", col, argPos))
		piiArgs = append(piiArgs, val)
		argPos++
	}

	for column, value := range updates {
		switch column {
		case columns.TenantName, columns.TenantType, columns.TenantAddress, columns.TenantActive:
			addTenant(column, value)
		case columns.TenantPiiPhoneNumber, columns.TenantPiiPhoneEncrypted, columns.TenantPiiPhoneHash:
			phone, _ := value.(string)
			encrypted, err := r.nid.Encrypt(phone)
			if err != nil {
				return fmt.Errorf("failed to encrypt phone: %w", err)
			}
			addPii(columns.TenantPiiPhoneNumber, phone)
			addPii(columns.TenantPiiPhoneEncrypted, encrypted)
			addPii(columns.TenantPiiPhoneHash, r.nid.Hash(phone))
		case columns.TenantPiiEmail, columns.TenantPiiEmailEncrypted, columns.TenantPiiEmailHash:
			email, _ := value.(string)
			encrypted, err := r.nid.Encrypt(email)
			if err != nil {
				return fmt.Errorf("failed to encrypt email: %w", err)
			}
			addPii(columns.TenantPiiEmail, email)
			addPii(columns.TenantPiiEmailEncrypted, encrypted)
			addPii(columns.TenantPiiEmailHash, r.nid.Hash(email))
		case columns.TenantPiiNIDNumber, columns.TenantPiiNIDEncrypted, columns.TenantPiiNIDHash, columns.TenantPiiNIDLastFour:
			nid, _ := value.(string)
			encrypted, err := r.nid.Encrypt(nid)
			if err != nil {
				return fmt.Errorf("failed to encrypt NID: %w", err)
			}
			addPii(columns.TenantPiiNIDEncrypted, encrypted)
			addPii(columns.TenantPiiNIDHash, r.nid.Hash(nid))
			addPii(columns.TenantPiiNIDLastFour, appcrypto.LastFour(nid))
		default:
			addTenant(column, value)
		}
	}

	if len(tenantCols) > 0 {
		setSql := strings.Join(tenantCols, ", ")
		tenantArgs = append(tenantArgs, id)
		if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", columns.TenantTable, setSql, argPos), tenantArgs...); err != nil {
			return fmt.Errorf("failed to update tenant: %w", err)
		}
	}
	if len(piiCols) > 0 {
		setSql := strings.Join(piiCols, ", ")
		piiArgs = append(piiArgs, id)
		if _, err := tx.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE tenant_id = $%d", columns.TenantPiiTable, setSql, argPos), piiArgs...); err != nil {
			return fmt.Errorf("failed to update tenant pii: %w", err)
		}
	}

	if len(tenantCols) > 0 || len(piiCols) > 0 {
		if _, err := tx.Exec(
			fmt.Sprintf("UPDATE %s SET updated_at = $1 WHERE id = $2", columns.TenantTable),
			time.Now(), id,
		); err != nil {
			return fmt.Errorf("failed to update tenant timestamp: %w", err)
		}
	}

	return tx.Commit()
}
