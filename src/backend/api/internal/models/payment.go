package models

import "time"

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPaid    PaymentStatus = "Paid"
	PaymentStatusDue     PaymentStatus = "Due"
	PaymentStatusPartial PaymentStatus = "Partial"
	PaymentStatusOverdue PaymentStatus = "Overdue"
)

type Payment struct {
	ID            int           `json:"id" db:"id"`
	UnitID        int           `json:"unit_id" db:"unit_id"`
	TenantID      int           `json:"tenant_id" db:"tenant_id"`
	BuildingID     int           `json:"building_id" db:"building_id"` // Denormalized for reporting
	PropertyID     int           `json:"property_id" db:"property_id"` // Denormalized for reporting
	OrganizationID int           `json:"organization_id" db:"organization_id"`
	Month          int           `json:"month" db:"month"`
	Year          int           `json:"year" db:"year"`
	AmountDue     float64       `json:"amount_due" db:"amount_due"`
	AmountPaid    float64       `json:"amount_paid" db:"amount_paid"`
	Status        PaymentStatus `json:"status" db:"status"`
	PaymentMethod string        `json:"payment_method" db:"payment_method"`
	PaymentDate   *time.Time    `json:"payment_date" db:"payment_date"`
	Notes         string        `json:"notes" db:"notes"`
	ReceiptNumber string        `json:"receipt_number" db:"receipt_number"`
	DueDate       *time.Time    `json:"due_date" db:"due_date"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

type CreatePaymentRequest struct {
	UnitID     int     `json:"unit_id" binding:"required"`
	TenantID   int     `json:"tenant_id" binding:"required"`
	BuildingID int     `json:"building_id" binding:"required"`
	PropertyID int     `json:"property_id" binding:"required"`
	Month      int     `json:"month" binding:"required,min=1,max=12"`
	Year       int     `json:"year" binding:"required,min=2020"`
	AmountDue  float64 `json:"amount_due" binding:"required,gt=0"`
	DueDate    string  `json:"due_date" binding:"omitempty"`
}

type UpdatePaymentRequest struct {
	AmountPaid    *float64       `json:"amount_paid" binding:"omitempty,gte=0"`
	Status        *PaymentStatus `json:"status" binding:"omitempty,oneof=Paid Due Partial Overdue"`
	PaymentMethod *string        `json:"payment_method" binding:"omitempty"`
	PaymentDate   *string        `json:"payment_date" binding:"omitempty"`
	Notes         *string        `json:"notes" binding:"omitempty"`
	ReceiptNumber *string        `json:"receipt_number" binding:"omitempty"`
}

type PaymentWithDetails struct {
	Payment
	PropertyName string `json:"property_name" db:"property_name"`
	BuildingName string `json:"building_name" db:"building_name"`
	BuildingCode string `json:"building_code" db:"building_code"`
	UnitNumber   string `json:"unit_number" db:"unit_number"`
	UnitType     string `json:"unit_type" db:"unit_type"`
	TenantName   string `json:"tenant_name" db:"tenant_name"`
}

// DashboardSummary represents aggregated payment statistics
type DashboardSummary struct {
	TotalDue       float64 `json:"total_due"`
	TotalPaid      float64 `json:"total_paid"`
	TotalPending   float64 `json:"total_pending"`
	TotalOverdue   float64 `json:"total_overdue"`
	CollectionRate float64 `json:"collection_rate"`
	PropertyCount  int     `json:"property_count"`
	BuildingCount  int     `json:"building_count"`
	UnitCount      int     `json:"unit_count"`
	TenantCount    int     `json:"tenant_count"`
}
