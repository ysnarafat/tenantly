package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// PaymentTransactionRepository handles CRUD for individual payment
// transactions — the source of truth for how much has actually been paid
// against a payments row.
type PaymentTransactionRepository struct {
	db *sqlx.DB
}

// NewPaymentTransactionRepository creates a new payment transaction repository
func NewPaymentTransactionRepository(db *sqlx.DB) *PaymentTransactionRepository {
	return &PaymentTransactionRepository{db: db}
}

const paymentTransactionCols = `id, payment_id, amount, COALESCE(payment_method, ''), payment_date, COALESCE(receipt_number, ''), COALESCE(notes, ''), created_at`

func scanPaymentTransaction(row interface {
	Scan(...interface{}) error
}) (*models.PaymentTransaction, error) {
	txn := &models.PaymentTransaction{}
	err := row.Scan(
		&txn.ID, &txn.PaymentID, &txn.Amount, &txn.PaymentMethod,
		&txn.PaymentDate, &txn.ReceiptNumber, &txn.Notes, &txn.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

// Create records a new amount received against a payment.
func (r *PaymentTransactionRepository) Create(paymentID int, amount float64, paymentMethod, receiptNumber, notes string, paymentDate time.Time) (*models.PaymentTransaction, error) {
	var method, receipt, note sql.NullString
	if paymentMethod != "" {
		method = sql.NullString{String: paymentMethod, Valid: true}
	}
	if receiptNumber != "" {
		receipt = sql.NullString{String: receiptNumber, Valid: true}
	}
	if notes != "" {
		note = sql.NullString{String: notes, Valid: true}
	}

	query := `
		INSERT INTO payment_transactions (payment_id, amount, payment_method, payment_date, receipt_number, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + paymentTransactionCols

	row := r.db.QueryRow(query, paymentID, amount, method, paymentDate, receipt, note)
	txn, err := scanPaymentTransaction(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}
	return txn, nil
}

// GetByID retrieves a single transaction by ID.
func (r *PaymentTransactionRepository) GetByID(id int) (*models.PaymentTransaction, error) {
	query := `SELECT ` + paymentTransactionCols + ` FROM payment_transactions WHERE id = $1`
	txn, err := scanPaymentTransaction(r.db.QueryRow(query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transaction: %w", err)
	}
	return txn, nil
}

// GetByPaymentID returns every transaction recorded against a payment,
// earliest first — the ledger view of how it was settled.
func (r *PaymentTransactionRepository) GetByPaymentID(paymentID int) ([]*models.PaymentTransaction, error) {
	query := `SELECT ` + paymentTransactionCols + ` FROM payment_transactions WHERE payment_id = $1 ORDER BY payment_date, id`
	rows, err := r.db.Query(query, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment transactions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	txns := make([]*models.PaymentTransaction, 0)
	for rows.Next() {
		txn, err := scanPaymentTransaction(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment transaction: %w", err)
		}
		txns = append(txns, txn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment transaction rows error: %w", err)
	}
	return txns, nil
}

// Delete removes a mistakenly-recorded transaction outright — there is no
// history-of-history to preserve for an entry that should never have existed.
func (r *PaymentTransactionRepository) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM payment_transactions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment transaction: %w", err)
	}
	return nil
}

// SumByPaymentID totals every transaction recorded against a payment — the
// authoritative amount_paid, recomputed whenever a transaction is added or
// removed rather than trusted from client input.
func (r *PaymentTransactionRepository) SumByPaymentID(paymentID int) (float64, error) {
	var total float64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM payment_transactions WHERE payment_id = $1`,
		paymentID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to sum payment transactions: %w", err)
	}
	return total, nil
}
