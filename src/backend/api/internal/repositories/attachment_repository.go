package repositories

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// AttachmentRepository handles CRUD for files attached to tenants and leases.
// File bytes are stored in Postgres (bytea) rather than on local disk.
type AttachmentRepository struct {
	db *sqlx.DB
}

// NewAttachmentRepository creates a new attachment repository
func NewAttachmentRepository(db *sqlx.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

const tenantAttachmentCols = `id, tenant_id, file_name, content_type, file_size, blob, uploaded_at`
const leaseAttachmentCols = `id, lease_id, file_name, content_type, file_size, blob, uploaded_at`

func scanAttachment(row interface {
	Scan(...interface{}) error
}) (*models.Attachment, error) {
	att := &models.Attachment{}
	err := row.Scan(
		&att.ID, &att.TenantID, &att.LeaseID, &att.Filename, &att.ContentType,
		&att.FileSize, &att.Blob, &att.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	return att, nil
}

// CreateTenantAttachment stores a new attachment's metadata and bytes against a tenant.
func (r *AttachmentRepository) CreateTenantAttachment(tenantID int, fileName, contentType string, fileSize int, data []byte) (*models.Attachment, error) {
	query := `
		INSERT INTO tenant_attachments (tenant_id, file_name, content_type, file_size, blob)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + tenantAttachmentCols

	row := r.db.QueryRow(query, tenantID, fileName, contentType, fileSize, data)
	att, err := scanAttachment(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant attachment: %w", err)
	}
	return att, nil
}

// CreateLeaseAttachment stores a new attachment's metadata and bytes against a lease.
func (r *AttachmentRepository) CreateLeaseAttachment(leaseID int, fileName, contentType string, fileSize int, data []byte) (*models.Attachment, error) {
	query := `
		INSERT INTO lease_attachments (lease_id, file_name, content_type, file_size, blob)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + leaseAttachmentCols

	row := r.db.QueryRow(query, leaseID, fileName, contentType, fileSize, data)
	att, err := scanAttachment(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease attachment: %w", err)
	}
	return att, nil
}

// GetTenantAttachmentByID retrieves a tenant attachment's metadata and blob by ID.
func (r *AttachmentRepository) GetTenantAttachmentByID(id int) (*models.Attachment, error) {
	query := `SELECT ` + tenantAttachmentCols + ` FROM tenant_attachments WHERE id = $1`
	att, err := scanAttachment(r.db.QueryRow(query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant attachment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant attachment: %w", err)
	}
	return att, nil
}

// GetLeaseAttachmentByID retrieves a lease attachment's metadata and blob by ID.
func (r *AttachmentRepository) GetLeaseAttachmentByID(id int) (*models.Attachment, error) {
	query := `SELECT ` + leaseAttachmentCols + ` FROM lease_attachments WHERE id = $1`
	att, err := scanAttachment(r.db.QueryRow(query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lease attachment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lease attachment: %w", err)
	}
	return att, nil
}

// GetAttachmentsByTenantID lists every attachment for a tenant, oldest first.
func (r *AttachmentRepository) GetAttachmentsByTenantID(tenantID int) ([]*models.Attachment, error) {
	query := `SELECT ` + tenantAttachmentCols + ` FROM tenant_attachments WHERE tenant_id = $1 ORDER BY uploaded_at, id`
	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	atts := make([]*models.Attachment, 0)
	for rows.Next() {
		att, err := scanAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant attachment: %w", err)
		}
		atts = append(atts, att)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tenant attachment rows error: %w", err)
	}
	return atts, nil
}

// GetAttachmentsByLeaseID lists every attachment for a lease, oldest first.
func (r *AttachmentRepository) GetAttachmentsByLeaseID(leaseID int) ([]*models.Attachment, error) {
	query := `SELECT ` + leaseAttachmentCols + ` FROM lease_attachments WHERE lease_id = $1 ORDER BY uploaded_at, id`
	rows, err := r.db.Query(query, leaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	atts := make([]*models.Attachment, 0)
	for rows.Next() {
		att, err := scanAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lease attachment: %w", err)
		}
		atts = append(atts, att)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lease attachment rows error: %w", err)
	}
	return atts, nil
}

// GetTenantAttachmentFile returns a tenant attachment's raw bytes for download.
func (r *AttachmentRepository) GetTenantAttachmentFile(id int) ([]byte, string, string, error) {
	var data []byte
	var fileName, contentType string
	err := r.db.QueryRow(
		`SELECT blob, file_name, content_type FROM tenant_attachments WHERE id = $1`,
		id,
	).Scan(&data, &fileName, &contentType)
	if err == sql.ErrNoRows {
		return nil, "", "", fmt.Errorf("tenant attachment not found")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get tenant attachment file: %w", err)
	}
	return data, fileName, contentType, nil
}

// GetLeaseAttachmentFile returns a lease attachment's raw bytes for download.
func (r *AttachmentRepository) GetLeaseAttachmentFile(id int) ([]byte, string, string, error) {
	var data []byte
	var fileName, contentType string
	err := r.db.QueryRow(
		`SELECT blob, file_name, content_type FROM lease_attachments WHERE id = $1`,
		id,
	).Scan(&data, &fileName, &contentType)
	if err == sql.ErrNoRows {
		return nil, "", "", fmt.Errorf("lease attachment not found")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get lease attachment file: %w", err)
	}
	return data, fileName, contentType, nil
}

// DeleteTenantAttachment removes a tenant attachment.
func (r *AttachmentRepository) DeleteTenantAttachment(id int) error {
	_, err := r.db.Exec(`DELETE FROM tenant_attachments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant attachment: %w", err)
	}
	return nil
}

// DeleteLeaseAttachment removes a lease attachment.
func (r *AttachmentRepository) DeleteLeaseAttachment(id int) error {
	_, err := r.db.Exec(`DELETE FROM lease_attachments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete lease attachment: %w", err)
	}
	return nil
}
