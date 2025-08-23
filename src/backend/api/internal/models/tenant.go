package models

import "time"

type Tenant struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Email       string    `json:"email" db:"email"`
	NIDNumber   string    `json:"nid_number" db:"nid_number"`
	Address     string    `json:"address" db:"address"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Email       string `json:"email" binding:"omitempty,email"`
	NIDNumber   string `json:"nid_number" binding:"omitempty,max=20"`
	Address     string `json:"address" binding:"omitempty"`
}

type UpdateTenantRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Email       string `json:"email" binding:"omitempty,email"`
	NIDNumber   string `json:"nid_number" binding:"omitempty,max=20"`
	Address     string `json:"address" binding:"omitempty"`
	Active      *bool  `json:"active"`
}
