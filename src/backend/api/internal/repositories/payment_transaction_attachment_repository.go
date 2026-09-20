package repositories

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentTransactionAttachmentRepository handles CRUD for files attached to
// individual payment transactions (receipt photos, mobile-banking
// screenshots, etc.). File bytes are stored in Postgres (bytea) rather than
// on local disk — the API container has no persistent volume for uploads,
// and attachments are capped small enough (10MB) for this to be practical.
type PaymentTransactionAttachmentRepository struct {
	db *sqlx.DB
}

// NewPaymentTransactionAttachmentRepository creates a new payment transaction attachment repository
func NewPaymentTransactionAttachmentRepository(db *sqlx.DB) *PaymentTransactionAttachmentRepository {
	return &PaymentTransactionAttachmentRepository{db: db}
}

const paymentTransactionAttachmentCols = `id, payment_transaction_id, file_name, content_type, file_size, uploaded_by, created_at`

func scanPaymentTransactionAttachment(row interface {
	Scan(...interface{}) error
}) (*models.PaymentTransactionAttachment, error) {
	att := &models.PaymentTransactionAttachment{}
	var uploadedBy sql.NullInt64
	err := row.Scan(
		&att.ID, &att.PaymentTransactionID, &att.FileName, &att.ContentType, &att.FileSize,
		&uploadedBy, &att.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if uploadedBy.Valid {
		id := int(uploadedBy.Int64)
		att.UploadedBy = &id
	}
	return att, nil
}

// Create stores a new attachment's metadata and bytes against a transaction.
func (r *PaymentTransactionAttachmentRepository) Create(transactionID int, fileName, contentType string, fileSize int, data []byte, uploadedBy *int) (*models.PaymentTransactionAttachment, error) {
	var uploader sql.NullInt64
	if uploadedBy != nil {
		uploader = sql.NullInt64{Int64: int64(*uploadedBy), Valid: true}
	}

	query := `
		INSERT INTO payment_transaction_attachments (payment_transaction_id, file_name, content_type, file_size, file_data, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + paymentTransactionAttachmentCols

	row := r.db.QueryRow(query, transactionID, fileName, contentType, fileSize, data, uploader)
	att, err := scanPaymentTransactionAttachment(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment transaction attachment: %w", err)
	}
	return att, nil
}

// GetByID retrieves an attachment's metadata (not its file bytes) by ID.
func (r *PaymentTransactionAttachmentRepository) GetByID(id int) (*models.PaymentTransactionAttachment, error) {
	query := `SELECT ` + paymentTransactionAttachmentCols + ` FROM payment_transaction_attachments WHERE id = $1`
	att, err := scanPaymentTransactionAttachment(r.db.QueryRow(query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment transaction attachment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction attachment: %w", err)
	}
	return att, nil
}

// GetByTransactionID lists every attachment recorded against a transaction, oldest first.
func (r *PaymentTransactionAttachmentRepository) GetByTransactionID(transactionID int) ([]*models.PaymentTransactionAttachment, error) {
	query := `SELECT ` + paymentTransactionAttachmentCols + ` FROM payment_transaction_attachments WHERE payment_transaction_id = $1 ORDER BY created_at, id`
	rows, err := r.db.Query(query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	atts := make([]*models.PaymentTransactionAttachment, 0)
	for rows.Next() {
		att, err := scanPaymentTransactionAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment transaction attachment: %w", err)
		}
		atts = append(atts, att)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment transaction attachment rows error: %w", err)
	}
	return atts, nil
}

// GetFileData returns an attachment's raw bytes for download, along with its
// file name and content type.
func (r *PaymentTransactionAttachmentRepository) GetFileData(id int) ([]byte, string, string, error) {
	var data []byte
	var fileName, contentType string
	err := r.db.QueryRow(
		`SELECT file_data, file_name, content_type FROM payment_transaction_attachments WHERE id = $1`,
		id,
	).Scan(&data, &fileName, &contentType)
	if err == sql.ErrNoRows {
		return nil, "", "", fmt.Errorf("payment transaction attachment not found")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get payment transaction attachment file: %w", err)
	}
	return data, fileName, contentType, nil
}

// Delete removes an attachment outright.
func (r *PaymentTransactionAttachmentRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM payment_transaction_attachments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment transaction attachment: %w", err)
	}
	return nil
}
