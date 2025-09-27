package models

import "time"

type Payment struct {
	ID            int        `json:"id" db:"id"`
	ShopID        int        `json:"shop_id" db:"shop_id"`
	TenantID      int        `json:"tenant_id" db:"tenant_id"`
	Month         int        `json:"month" db:"month"`
	Year          int        `json:"year" db:"year"`
	AmountDue     float64    `json:"amount_due" db:"amount_due"`
	AmountPaid    float64    `json:"amount_paid" db:"amount_paid"`
	Status        string     `json:"status" db:"status"`
	PaymentMethod string     `json:"payment_method" db:"payment_method"`
	PaymentDate   *time.Time `json:"payment_date" db:"payment_date"`
	Notes         string     `json:"notes" db:"notes"`
	ReceiptNumber string     `json:"receipt_number" db:"receipt_number"`
	DueDate       *time.Time `json:"due_date" db:"due_date"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type CreatePaymentRequest struct {
	ShopID    int     `json:"shop_id" binding:"required"`
	TenantID  int     `json:"tenant_id" binding:"required"`
	Month     int     `json:"month" binding:"required,min=1,max=12"`
	Year      int     `json:"year" binding:"required,min=2020"`
	AmountDue float64 `json:"amount_due" binding:"required,gt=0"`
	DueDate   string  `json:"due_date" binding:"omitempty"`
}

type UpdatePaymentRequest struct {
	AmountPaid    *float64 `json:"amount_paid" binding:"omitempty,gte=0"`
	Status        string   `json:"status" binding:"omitempty,oneof=Paid Due Partial"`
	PaymentMethod string   `json:"payment_method" binding:"omitempty"`
	PaymentDate   string   `json:"payment_date" binding:"omitempty"`
	Notes         string   `json:"notes" binding:"omitempty"`
	ReceiptNumber string   `json:"receipt_number" binding:"omitempty"`
}

type PaymentWithDetails struct {
	Payment
	ShopName   string `json:"shop_name" db:"shop_name"`
	TenantName string `json:"tenant_name" db:"tenant_name"`
}
