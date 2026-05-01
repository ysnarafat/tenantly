package models

import "time"

// LeaseType represents the type of lease
type LeaseType string

const (
	LeaseTypeResidential LeaseType = "Residential"
	LeaseTypeCommercial  LeaseType = "Commercial"
)

type Lease struct {
	ID              int       `json:"id" db:"id"`
	UnitID          int       `json:"unit_id" db:"unit_id"`
	TenantID        int       `json:"tenant_id" db:"tenant_id"`
	LeaseType       LeaseType `json:"lease_type" db:"lease_type"`
	StartDate       time.Time `json:"start_date" db:"start_date"`
	EndDate         time.Time `json:"end_date" db:"end_date"`
	DurationMonths  int       `json:"duration_months" db:"duration_months"`
	MonthlyRent     float64   `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit" db:"security_deposit"`
	Active          bool      `json:"active" db:"active"`
	OrganizationID  int       `json:"organization_id" db:"organization_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateLeaseRequest struct {
	UnitID          int       `json:"unit_id" binding:"required"`
	TenantID        int       `json:"tenant_id" binding:"required"`
	LeaseType       LeaseType `json:"lease_type" binding:"required,oneof=Residential Commercial"`
	StartDate       string    `json:"start_date" binding:"required"`
	EndDate         *string   `json:"end_date" binding:"omitempty"` // Optional for open-ended leases
	DurationMonths  int       `json:"duration_months" binding:"required,min=1"`
	MonthlyRent     float64   `json:"monthly_rent" binding:"required,gt=0"`
	SecurityDeposit float64   `json:"security_deposit" binding:"omitempty,gte=0"`
	OrganizationID  int       `json:"-"`
}

type UpdateLeaseRequest struct {
	LeaseType       *LeaseType `json:"lease_type" binding:"omitempty,oneof=Residential Commercial"`
	StartDate       *string    `json:"start_date" binding:"omitempty"`
	DurationMonths  *int       `json:"duration_months" binding:"omitempty,min=1"`
	MonthlyRent     *float64   `json:"monthly_rent" binding:"omitempty,gt=0"`
	SecurityDeposit *float64   `json:"security_deposit" binding:"omitempty,gte=0"`
	Active          *bool      `json:"active"`
}

type LeaseWithDetails struct {
	Lease
	PropertyName  string `json:"property_name" db:"property_name"`
	BuildingName  string `json:"building_name" db:"building_name"`
	BuildingCode  string `json:"building_code" db:"building_code"`
	UnitNumber    string `json:"unit_number" db:"unit_number"`
	UnitType      string `json:"unit_type" db:"unit_type"`
	TenantName    string `json:"tenant_name" db:"tenant_name"`
	TenantPhone   string `json:"tenant_phone" db:"tenant_phone"`
	IsExpired     bool   `json:"is_expired"`
	DaysRemaining int    `json:"days_remaining"`
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
