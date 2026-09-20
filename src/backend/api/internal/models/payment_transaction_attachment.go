package models

import "time"

// MaxAttachmentFileSize is the hard cap on any single attachment upload.
const MaxAttachmentFileSize = 10 * 1024 * 1024 // 10MB

// AllowedAttachmentContentTypes are the only MIME types accepted for payment
// transaction attachments — images and PDFs. Callers must match this against
// content sniffed from the file bytes (http.DetectContentType), never the
// client-supplied filename or Content-Type header, both of which are
// trivially spoofable.
var AllowedAttachmentContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
}

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
