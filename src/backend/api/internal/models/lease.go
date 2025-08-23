package models

import "time"

type Lease struct {
	ID              int       `json:"id" db:"id"`
	ShopID          int       `json:"shop_id" db:"shop_id"`
	TenantID        int       `json:"tenant_id" db:"tenant_id"`
	StartDate       time.Time `json:"start_date" db:"start_date"`
	DurationMonths  int       `json:"duration_months" db:"duration_months"`
	MonthlyRent     float64   `json:"monthly_rent" db:"monthly_rent"`
	SecurityDeposit float64   `json:"security_deposit" db:"security_deposit"`
	Active          bool      `json:"active" db:"active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateLeaseRequest struct {
	ShopID          int     `json:"shop_id" binding:"required"`
	TenantID        int     `json:"tenant_id" binding:"required"`
	StartDate       string  `json:"start_date" binding:"required"`
	DurationMonths  int     `json:"duration_months" binding:"required,min=1"`
	MonthlyRent     float64 `json:"monthly_rent" binding:"required,gt=0"`
	SecurityDeposit float64 `json:"security_deposit" binding:"omitempty,gte=0"`
}

type UpdateLeaseRequest struct {
	StartDate       string   `json:"start_date" binding:"omitempty"`
	DurationMonths  *int     `json:"duration_months" binding:"omitempty,min=1"`
	MonthlyRent     *float64 `json:"monthly_rent" binding:"omitempty,gt=0"`
	SecurityDeposit *float64 `json:"security_deposit" binding:"omitempty,gte=0"`
	Active          *bool    `json:"active"`
}

type LeaseWithDetails struct {
	Lease
	ShopName      string    `json:"shop_name" db:"shop_name"`
	ShopNumber    string    `json:"shop_number" db:"shop_number"`
	TenantName    string    `json:"tenant_name" db:"tenant_name"`
	TenantPhone   string    `json:"tenant_phone" db:"tenant_phone"`
	EndDate       time.Time `json:"end_date"`
	IsExpired     bool      `json:"is_expired"`
	DaysRemaining int       `json:"days_remaining"`
}
