package models

import "time"

// ChargeType is the controlled set of recurring charges a lease can carry,
// kept as an enum (rather than freeform text) so charges stay reportable
// across leases (e.g. "total utility revenue this month").
type ChargeType string

const (
	ChargeTypeUtility       ChargeType = "Utility"
	ChargeTypeServiceCharge ChargeType = "ServiceCharge"
	ChargeTypeMaintenance   ChargeType = "Maintenance"
	ChargeTypeParking       ChargeType = "Parking"
	ChargeTypeOther         ChargeType = "Other"
)

// LeaseCharge is a recurring amount added on top of a lease's monthly_rent
// when generating payments (see PaymentService.GenerateMonthlyPayments).
type LeaseCharge struct {
	ID         int        `json:"id" db:"id"`
	LeaseID    int        `json:"lease_id" db:"lease_id"`
	ChargeType ChargeType `json:"charge_type" db:"charge_type"`
	Label      string     `json:"label" db:"label"`
	Amount     float64    `json:"amount" db:"amount"`
	Active     bool       `json:"active" db:"active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateLeaseChargeRequest struct {
	ChargeType ChargeType `json:"charge_type" binding:"required,oneof=Utility ServiceCharge Maintenance Parking Other"`
	Label      string     `json:"label" binding:"required,max=100"`
	Amount     float64    `json:"amount" binding:"required,gte=0"`
}

type UpdateLeaseChargeRequest struct {
	ChargeType *ChargeType `json:"charge_type" binding:"omitempty,oneof=Utility ServiceCharge Maintenance Parking Other"`
	Label      *string     `json:"label" binding:"omitempty,max=100"`
	Amount     *float64    `json:"amount" binding:"omitempty,gte=0"`
	Active     *bool       `json:"active"`
}
