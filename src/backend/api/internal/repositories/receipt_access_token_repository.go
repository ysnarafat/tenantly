package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// ReceiptAccessTokenRepository manages the opaque tokens that let a
// "payment recorded" SMS link to a receipt PDF without any tenant login.
type ReceiptAccessTokenRepository struct {
	db *sqlx.DB
}

// NewReceiptAccessTokenRepository creates a new receipt access token repository.
func NewReceiptAccessTokenRepository(db *sqlx.DB) *ReceiptAccessTokenRepository {
	return &ReceiptAccessTokenRepository{db: db}
}

const receiptAccessTokenCols = `id, token, payment_id, expires_at, created_at`

func scanReceiptAccessToken(row interface {
	Scan(...interface{}) error
}) (*models.ReceiptAccessToken, error) {
	t := &models.ReceiptAccessToken{}
	if err := row.Scan(&t.ID, &t.Token, &t.PaymentID, &t.ExpiresAt, &t.CreatedAt); err != nil {
		return nil, err
	}
	return t, nil
}

// Create stores a new receipt access token for a payment.
func (r *ReceiptAccessTokenRepository) Create(paymentID int, token string, expiresAt time.Time) (*models.ReceiptAccessToken, error) {
	query := `
		INSERT INTO receipt_access_tokens (token, payment_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING ` + receiptAccessTokenCols

	t, err := scanReceiptAccessToken(r.db.QueryRow(query, token, paymentID, expiresAt))
	if err != nil {
		return nil, fmt.Errorf("failed to create receipt access token: %w", err)
	}
	return t, nil
}

// GetValidByPaymentID returns the most recent non-expired token for a
// payment, if one exists — reused across installments so the same SMS link
// keeps working rather than issuing a new one every time.
func (r *ReceiptAccessTokenRepository) GetValidByPaymentID(paymentID int) (*models.ReceiptAccessToken, error) {
	query := `SELECT ` + receiptAccessTokenCols + `
		FROM receipt_access_tokens
		WHERE payment_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	t, err := scanReceiptAccessToken(r.db.QueryRow(query, paymentID))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no valid receipt access token for payment")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt access token: %w", err)
	}
	return t, nil
}

// GetByToken looks up a token regardless of expiry — callers check
// ExpiresAt themselves so they can distinguish "not found" from "expired".
func (r *ReceiptAccessTokenRepository) GetByToken(token string) (*models.ReceiptAccessToken, error) {
	query := `SELECT ` + receiptAccessTokenCols + ` FROM receipt_access_tokens WHERE token = $1`

	t, err := scanReceiptAccessToken(r.db.QueryRow(query, token))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("receipt access token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt access token: %w", err)
	}
	return t, nil
}
