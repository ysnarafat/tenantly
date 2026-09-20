package models

import "time"

// TenantType represents the type of tenant
type TenantType string

const (
	TenantTypeIndividual TenantType = "Individual"
	TenantTypeBusiness   TenantType = "Business"
)

type Tenant struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	TenantType  TenantType `json:"tenant_type" db:"tenant_type"`
	PhoneNumber string     `json:"phone_number" db:"phone_number"`
	Email       string     `json:"email" db:"email"`
	// NIDNumber holds the decrypted NID in memory only and is never serialized.
	// It is populated only on explicit, role-gated reveal — not on list/detail
	// reads, which carry only NIDLastFour.
	NIDNumber      string    `json:"-" db:"-"`
	NIDLastFour    string    `json:"nid_last_four" db:"nid_last_four"`
	Address        string    `json:"address" db:"address"`
	Active         bool      `json:"active" db:"active"`
	OrganizationID int       `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// TenantResponse represents the API response for a tenant. The full NID is never
// included here; only the non-sensitive last four digits are exposed. Use the
// dedicated role-gated reveal endpoint to obtain the full value.
type TenantResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	TenantType  TenantType `json:"tenant_type"`
	PhoneNumber string     `json:"phone_number"`
	Email       string     `json:"email"`
	NIDLastFour string     `json:"nid_last_four"`
	Address     string     `json:"address"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts a Tenant model to a TenantResponse DTO
func (t *Tenant) ToResponse() *TenantResponse {
	return &TenantResponse{
		ID:          t.ID,
		Name:        t.Name,
		TenantType:  t.TenantType,
		PhoneNumber: t.PhoneNumber,
		Email:       t.Email,
		NIDLastFour: t.NIDLastFour,
		Address:     t.Address,
		Active:      t.Active,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
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
