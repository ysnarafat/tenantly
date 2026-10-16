package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// LeaseType represents the type of lease
type LeaseType string

const (
	LeaseTypeResidential LeaseType = "Residential"
	LeaseTypeCommercial  LeaseType = "Commercial"
)

// LeaseCustomFields stores arbitrary org-defined key/value pairs not covered
// by structured lease fields — same JSONB pattern as PropertyMetadata/
// BuildingMetadata/UnitMetadata.
type LeaseCustomFields map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m LeaseCustomFields) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *LeaseCustomFields) Scan(value interface{}) error {
	if value == nil {
		*m = make(LeaseCustomFields)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// LeaseEndReason records why a lease stopped being active, so the historical
// record stays legible without inferring intent from dates alone.
type LeaseEndReason string

const (
	LeaseEndReasonExpired    LeaseEndReason = "Expired"
	LeaseEndReasonTerminated LeaseEndReason = "Terminated"
	LeaseEndReasonRenewed    LeaseEndReason = "Renewed"
)

type Lease struct {
	ID                 int               `json:"id" db:"id"`
	UnitID             int               `json:"unit_id" db:"unit_id"`
	TenantID           int               `json:"tenant_id" db:"tenant_id"`
	LeaseType          LeaseType         `json:"lease_type" db:"lease_type"`
	StartDate          time.Time         `json:"start_date" db:"start_date"`
	EndDate            time.Time         `json:"end_date" db:"end_date"`
	DurationMonths     int               `json:"duration_months" db:"duration_months"`
	MonthlyRent        float64           `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit    float64           `json:"security_deposit" db:"security_deposit"`
	Active             bool              `json:"active" db:"active"`
	OrganizationID     int               `json:"organization_id" db:"organization_id"`
	EndReason          *LeaseEndReason   `json:"end_reason,omitempty" db:"end_reason"`
	RenewedFromLeaseID *int              `json:"renewed_from_lease_id,omitempty" db:"renewed_from_lease_id"`
	CustomFields       LeaseCustomFields `json:"custom_fields,omitempty" db:"custom_fields"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`
}

type CreateLeaseRequest struct {
	UnitID          int               `json:"unit_id" binding:"required"`
	TenantID        int               `json:"tenant_id" binding:"required"`
	LeaseType       LeaseType         `json:"lease_type" binding:"required,oneof=Residential Commercial"`
	StartDate       string            `json:"start_date" binding:"required"`
	EndDate         *string           `json:"end_date" binding:"omitempty"` // Optional for open-ended leases
	DurationMonths  int               `json:"duration_months" binding:"required,min=1"`
	MonthlyRent     float64           `json:"monthly_rent" binding:"required,gt=0"`
	SecurityDeposit float64           `json:"security_deposit" binding:"omitempty,gte=0"`
	CustomFields    LeaseCustomFields `json:"custom_fields" binding:"omitempty"`
	OrganizationID  int               `json:"-"`
}

type UpdateLeaseRequest struct {
	LeaseType      *LeaseType `json:"lease_type" binding:"omitempty,oneof=Residential Commercial"`
	StartDate      *string    `json:"start_date" binding:"omitempty"`
	DurationMonths *int       `json:"duration_months" binding:"omitempty,min=1"`
	// EndDate directly overrides the lease's end date — set independently of
	// DurationMonths so a correction isn't forced onto a whole-month
	// boundary. When provided, duration_months (if not itself also given) is
	// derived from it for display purposes only; it is not billing-relevant.
	EndDate         *string            `json:"end_date" binding:"omitempty"`
	MonthlyRent     *float64           `json:"monthly_rent" binding:"omitempty,gt=0"`
	SecurityDeposit *float64           `json:"security_deposit" binding:"omitempty,gte=0"`
	Active          *bool              `json:"active"`
	CustomFields    *LeaseCustomFields `json:"custom_fields" binding:"omitempty"`
}

// RenewLeaseRequest starts a new lease term for the same unit/tenant, closing
// out the lease being renewed rather than mutating it in place — so the
// original term's rent/duration/dates remain an accurate historical record.
// Fields left nil carry the corresponding value forward from the old lease.
type RenewLeaseRequest struct {
	// StartDate defaults to the renewed lease's end_date (back-to-back
	// coverage). Set explicitly for an early or late renewal.
	StartDate       *string    `json:"start_date" binding:"omitempty"`
	DurationMonths  int        `json:"duration_months" binding:"required,min=1"`
	MonthlyRent     *float64   `json:"monthly_rent" binding:"omitempty,gt=0"`
	SecurityDeposit *float64   `json:"security_deposit" binding:"omitempty,gte=0"`
	LeaseType       *LeaseType `json:"lease_type" binding:"omitempty,oneof=Residential Commercial"`
	// CarryForwardCharges defaults to true (nil) — the new lease inherits the
	// old lease's active charges unless the caller explicitly opts out.
	CarryForwardCharges *bool `json:"carry_forward_charges"`
}

// ReplaceTenantRequest describes a tenant turnover on an existing lease. The
// outgoing lease is closed at HandoverDate and a successor lease for
// NewTenantID opens on the same unit the same day, both in one transaction.
//
// Turnover is deliberately modelled as terminate-then-create rather than
// reassigning tenant_id in place: payments reference tenant_id directly, so
// mutating it would retroactively reattribute the outgoing tenant's payment
// history to the incoming one.
//
// Term fields are optional — omit one to carry it over from the outgoing lease.
type ReplaceTenantRequest struct {
	NewTenantID     int        `json:"new_tenant_id" binding:"required"`
	HandoverDate    string     `json:"handover_date" binding:"required"`
	LeaseType       *LeaseType `json:"lease_type" binding:"omitempty,oneof=Residential Commercial"`
	DurationMonths  *int       `json:"duration_months" binding:"omitempty,min=1,max=600"`
	MonthlyRent     *float64   `json:"monthly_rent" binding:"omitempty,gt=0"`
	SecurityDeposit *float64   `json:"security_deposit" binding:"omitempty,gte=0"`
}

// ReplaceTenantResponse returns both sides of a completed turnover so the
// caller can show what was closed and what was opened.
type ReplaceTenantResponse struct {
	PreviousLease *LeaseWithDetails `json:"previous_lease"`
	NewLease      *LeaseWithDetails `json:"new_lease"`
}

type LeaseWithDetails struct {
	Lease
	BuildingID    int            `json:"building_id" db:"building_id"`
	PropertyID    int            `json:"property_id" db:"property_id"`
	PropertyName  string         `json:"property_name" db:"property_name"`
	BuildingName  string         `json:"building_name" db:"building_name"`
	BuildingCode  string         `json:"building_code" db:"building_code"`
	UnitNumber    string         `json:"unit_number" db:"unit_number"`
	UnitType      string         `json:"unit_type" db:"unit_type"`
	TenantName    string         `json:"tenant_name" db:"tenant_name"`
	TenantPhone   string         `json:"tenant_phone" db:"tenant_phone"`
	IsExpired     bool           `json:"is_expired"`
	DaysRemaining int            `json:"days_remaining"`
	Charges       []*LeaseCharge `json:"charges,omitempty"`
}

// LeaseListResponse represents a paginated list of leases
type LeaseListResponse struct {
	Leases     []*LeaseWithDetails `json:"leases"`
	Pagination *PaginationInfo     `json:"pagination"`
}

// LeaseDetailResponse represents a lease with full details including payment history
type LeaseDetailResponse struct {
	LeaseWithDetails
	PaymentHistory  []*Payment `json:"payment_history,omitempty"`
	TotalPayments   float64    `json:"total_payments,omitempty"`
	Status          string     `json:"status"` // Active, Expired, Terminated
	TerminationDate *time.Time `json:"termination_date,omitempty"`
}
