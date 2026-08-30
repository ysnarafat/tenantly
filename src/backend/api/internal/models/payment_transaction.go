package models

import "time"

// PaymentTransaction is one amount actually received against a payments row.
// A payment can be settled across several of these (partial/installment
// payments) — see PaymentService.RecordPaymentTransaction.
type PaymentTransaction struct {
	ID            int       `json:"id" db:"id"`
	PaymentID     int       `json:"payment_id" db:"payment_id"`
	Amount        float64   `json:"amount" db:"amount"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	PaymentDate   time.Time `json:"payment_date" db:"payment_date"`
	ReceiptNumber string    `json:"receipt_number" db:"receipt_number"`
	Notes         string    `json:"notes" db:"notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// CreatePaymentTransactionRequest records a new amount received. PaymentDate
// defaults to today when omitted; a receipt number is always server-generated
// (see PaymentRepository.NextReceiptNumber), matching the invariant already
// established for the payments table itself.
type CreatePaymentTransactionRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod *string `json:"payment_method" binding:"omitempty"`
	PaymentDate   *string `json:"payment_date" binding:"omitempty"`
	Notes         *string `json:"notes" binding:"omitempty"`
}
