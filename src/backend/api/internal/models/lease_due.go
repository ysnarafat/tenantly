package models

// LeaseDue represents a lease with unpaid rent for the current month
type LeaseDue struct {
	LeaseID        int     `json:"lease_id" db:"lease_id"`
	TenantID       int     `json:"tenant_id" db:"tenant_id"`
	TenantName     string  `json:"tenant_name" db:"tenant_name"`
	UnitID         int     `json:"unit_id" db:"unit_id"`
	UnitNumber     string  `json:"unit_number" db:"unit_number"`
	UnitType       string  `json:"unit_type" db:"unit_type"`
	BuildingID     int     `json:"building_id" db:"building_id"`
	BuildingName   string  `json:"building_name" db:"building_name"`
	BuildingCode   string  `json:"building_code" db:"building_code"`
	PropertyID     int     `json:"property_id" db:"property_id"`
	PropertyName   string  `json:"property_name" db:"property_name"`
	MonthlyRent    float64 `json:"monthly_rent" db:"monthly_rent"`
	DaysOverdue    int     `json:"days_overdue" db:"days_overdue"`
	OrganizationID int     `json:"organization_id" db:"organization_id"`
}

// DueSummary represents summary statistics for unpaid rent
type DueSummary struct {
	TotalDueAmount  float64 `json:"total_due_amount"`
	TotalTenantsDue int     `json:"total_tenants_due"`
}
