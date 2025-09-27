package models

import "time"

type Shop struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	ShopNumber  string    `json:"shop_number" db:"shop_number"`
	Floor       string    `json:"floor" db:"floor"`
	Section     string    `json:"section" db:"section"`
	MonthlyRent float64   `json:"monthly_rent" db:"monthly_rent"`
	Active      bool      `json:"active" db:"active"`
	PropertyID  int       `json:"property_id" db:"property_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateShopRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	ShopNumber  string  `json:"shop_number" binding:"required,max=20"`
	Floor       string  `json:"floor" binding:"omitempty,max=10"`
	Section     string  `json:"section" binding:"omitempty,max=50"`
	MonthlyRent float64 `json:"monthly_rent" binding:"required,gt=0"`
	PropertyID  int     `json:"property_id" binding:"omitempty"`
}

type UpdateShopRequest struct {
	Name        string   `json:"name" binding:"omitempty,max=100"`
	ShopNumber  string   `json:"shop_number" binding:"omitempty,max=20"`
	Floor       string   `json:"floor" binding:"omitempty,max=10"`
	Section     string   `json:"section" binding:"omitempty,max=50"`
	MonthlyRent *float64 `json:"monthly_rent" binding:"omitempty,gt=0"`
	Active      *bool    `json:"active"`
}
