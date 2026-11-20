package models

import "time"

type BillingRunStatus string

const (
	BillingRunDraft     BillingRunStatus = "Draft"
	BillingRunFinalized BillingRunStatus = "Finalized"
)

type BillingRun struct {
	ID             int               `json:"id" db:"id"`
	BuildingID     int               `json:"building_id" db:"building_id"`
	OrganizationID int               `json:"organization_id" db:"organization_id"`
	PropertyName   string            `json:"property_name" db:"property_name"`
	BuildingName   string            `json:"building_name" db:"building_name"`
	BuildingCode   string            `json:"building_code" db:"building_code"`
	Month          int               `json:"month" db:"month"`
	Year           int               `json:"year" db:"year"`
	Status         BillingRunStatus  `json:"status" db:"status"`
	CreatedBy      *int              `json:"created_by" db:"created_by"`
	FinalizedAt    *time.Time        `json:"finalized_at" db:"finalized_at"`
	CreatedAt      time.Time         `json:"created_at" db:"created_at"`
	Items          []*BillingRunItem `json:"items"`
}

type BillingRunItem struct {
	ID                int      `json:"id" db:"id"`
	UnitID            int      `json:"unit_id" db:"unit_id"`
	UnitNumber        string   `json:"unit_number" db:"unit_number"`
	UnitType          string   `json:"unit_type" db:"unit_type"`
	Floor             int      `json:"floor" db:"floor"`
	MeterNumber       string   `json:"meter_number" db:"meter_number"`
	ElectricityAmount *float64 `json:"electricity_amount" db:"electricity_amount"`
}

type CreateBillingRunRequest struct {
	BuildingID int `json:"building_id" binding:"required,gt=0"`
	Month      int `json:"month" binding:"required,min=1,max=12"`
	Year       int `json:"year" binding:"required,min=2020"`
}

type UpdateBillingRunItemRequest struct {
	MeterNumber       string  `json:"meter_number" binding:"max=50"`
	ElectricityAmount float64 `json:"electricity_amount" binding:"required,gte=0"`
}
