package models

import (
	"time"

	"github.com/lib/pq"
)

// Attachment represents a document attachment uploaded by users
type Attachment struct {
	ID        int64     `json:"id"`
	TenantID  *int64    `json:"tenant_id,omitempty"`      // Nullable - either TenantID or LeaseID will be set
	LeaseID   *int64    `json:"lease_id,omitempty"`       // Nullable - either TenantID or LeaseID will be set
	Filename  string    `json:"file_name"`
	ContentType string   `json:"content_type"`
	FileSize  int64     `json:"file_size"`
	Blob      []byte    `json:"-"` // Binary data, not serialized to JSON
	UploadedAt time.Time `json:"uploaded_at"`
}

// Constants for attachment validation
const (
	MaxAttachmentFileSize = 10 * 1024 * 1024 // 10 MB
)

// AllowedAttachmentContentTypes contains the MIME types that are allowed for uploads
var AllowedAttachmentContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"application/pdf",
}
