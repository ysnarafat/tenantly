package models

import "time"

// PaymentTransactionAttachment is a file (receipt photo, bKash/Nagad
// screenshot, etc.) evidencing one installment of a payment. File bytes are
// intentionally not part of this struct — they're only ever served through
// the dedicated download endpoint, never embedded in a list/get JSON response.
type PaymentTransactionAttachment struct {
	ID                   int       `json:"id" db:"id"`
	PaymentTransactionID int       `json:"payment_transaction_id" db:"payment_transaction_id"`
	FileName             string    `json:"file_name" db:"file_name"`
	ContentType          string    `json:"content_type" db:"content_type"`
	FileSize             int       `json:"file_size" db:"file_size"`
	UploadedBy           *int      `json:"uploaded_by,omitempty" db:"uploaded_by"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}
