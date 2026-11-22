package models

import (
	"time"

	appcrypto "github.com/ysnarafat/tenantly/internal/crypto"
)

// TenantType represents the type of tenant
type TenantType string

const (
	TenantTypeIndividual TenantType = "Individual"
	TenantTypeBusiness   TenantType = "Business"
)

// Tenant represents a tenant with PII joined from tenant_pii.
// Non-PII fields come from `tenants`; PII fields come from `tenant_pii`.
type Tenant struct {
	ID             int        `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	TenantType     TenantType `json:"tenant_type" db:"tenant_type"`
	PhoneNumber    string     `json:"phone_number" db:"phone_number"` // plaintext from tenant_pii
	PhoneEncrypted string     `json:"-" db:"phone_encrypted"`
	PhoneHash      string     `json:"-" db:"phone_hash"`
	Email          string     `json:"email" db:"email"` // plaintext from tenant_pii
	EmailEncrypted string     `json:"-" db:"email_encrypted"`
	EmailHash      string     `json:"-" db:"email_hash"`
	Address        string     `json:"address" db:"address"`
	Active         bool       `json:"active" db:"active"`
	OrganizationID int        `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// NID fields — stored encrypted with hash and last-four projection in tenant_pii.
	// Plaintext nid_number is available from tenant_pii for backfill/SMS but should
	// not be serialized in API responses (use NIDLastFour instead).
	NIDNumber    string `json:"-" db:"nid_number"`                // plaintext (memory-only, not serialized)
	NIDEncrypted string `json:"-" db:"nid_encrypted"`             // ciphertext
	NIDLastFour  string `json:"nid_last_four" db:"nid_last_four"` // non-sensitive projection
	NIDHash      string `json:"-" db:"nid_hash"`                  // deterministic hash for lookups
}

// TenantResponse represents the API response for a tenant. Phone/Email are
// returned masked; the full values are only recoverable via role-gated
// decryption. The full NID is never included here; only the non-sensitive
// last-four digits are exposed.
type TenantResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	TenantType  TenantType `json:"tenant_type"`
	PhoneNumber string     `json:"phone_number"` // masked
	Email       string     `json:"email"`        // masked
	NIDLastFour string     `json:"nid_last_four"`
	Address     string     `json:"address"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts a Tenant model to a TenantResponse DTO. Phone and email
// are masked for display; the full values are only recoverable via role-gated
// decryption (RevealNID endpoint).
func (t *Tenant) ToResponse() *TenantResponse {
	return &TenantResponse{
		ID:          t.ID,
		Name:        t.Name,
		TenantType:  t.TenantType,
		PhoneNumber: appcrypto.MaskPhone(t.PhoneNumber),
		Email:       appcrypto.MaskEmail(t.Email),
		NIDLastFour: t.NIDLastFour,
		Address:     t.Address,
		Active:      t.Active,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// MaskedPhoneNumber returns the phone number masked for non-sensitive display
// (e.g. "••••••6789").
func (t *Tenant) MaskedPhoneNumber() string {
	return appcrypto.MaskPhone(t.PhoneNumber)
}

// MaskedEmail returns the email masked for non-sensitive display
// (e.g. "j***@example.com").
func (t *Tenant) MaskedEmail() string {
	return appcrypto.MaskEmail(t.Email)
}

// TenantListResponse represents a paginated list of tenants
type TenantListResponse struct {
	Tenants    []*TenantResponse `json:"tenants"`
	Pagination *PaginationInfo   `json:"pagination"`
}

type CreateTenantRequest struct {
	Name           string     `json:"name" binding:"required,max=100"`
	TenantType     TenantType `json:"tenant_type" binding:"required,oneof=Individual Business"`
	PhoneNumber    string     `json:"phone_number" binding:"required,max=20"`
	Email          string     `json:"email" binding:"omitempty,email"`
	NIDNumber      string     `json:"nid_number" binding:"required,max=20"`
	Address        string     `json:"address" binding:"omitempty"`
	OrganizationID int        `json:"-"`
}

type UpdateTenantRequest struct {
	Name        *string     `json:"name" binding:"omitempty,max=100"`
	TenantType  *TenantType `json:"tenant_type" binding:"omitempty,oneof=Individual Business"`
	PhoneNumber *string     `json:"phone_number" binding:"omitempty,max=20"`
	Email       *string     `json:"email" binding:"omitempty,email"`
	NIDNumber   *string     `json:"nid_number" binding:"omitempty,max=20"`
	Address     *string     `json:"address" binding:"omitempty"`
	Active      *bool       `json:"active"`
}

// TenantWithLeases includes tenant with their active leases
type TenantWithLeases struct {
	Tenant
	ActiveLeases int `json:"active_leases"`
	TotalUnits   int `json:"total_units"`
}
