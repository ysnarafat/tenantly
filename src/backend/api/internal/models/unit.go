package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// UnitType represents the type of rentable unit
type UnitType string

const (
	UnitTypeShop      UnitType = "Shop"
	UnitTypeApartment UnitType = "Apartment"
	UnitTypeOffice    UnitType = "Office"
	UnitTypeParking   UnitType = "Parking"
	UnitTypeStorage   UnitType = "Storage"
	UnitTypeOther     UnitType = "Other"
)

// UnitMetadata stores unit type-specific attributes
type UnitMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m UnitMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *UnitMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(UnitMetadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Unit represents a rentable space (shop, apartment, office, etc.)
type Unit struct {
	ID          int          `json:"id" db:"id"`
	BuildingID     int          `json:"building_id" db:"building_id"`
	PropertyID     int          `json:"property_id" db:"property_id"` // Denormalized for performance
	OrganizationID int          `json:"organization_id" db:"organization_id"`
	UnitNumber     string       `json:"unit_number" db:"unit_number"`
	UnitName    string       `json:"unit_name" db:"unit_name"`
	Floor       int          `json:"floor" db:"floor"`
	Section     string       `json:"section" db:"section"`
	UnitType    UnitType     `json:"unit_type" db:"unit_type"`
	MonthlyRent float64      `json:"monthly_rent" db:"monthly_rent"`
	Metadata    UnitMetadata `json:"metadata" db:"metadata"`
	Active      bool         `json:"active" db:"active"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// CreateUnitRequest represents the request to create a unit
type CreateUnitRequest struct {
	BuildingID  int          `json:"building_id" binding:"required"`
	PropertyID  int          `json:"property_id" binding:"required"`
	UnitNumber  string       `json:"unit_number" binding:"required,max=50"`
	UnitName    string       `json:"unit_name" binding:"omitempty,max=100"`
	Floor       int          `json:"floor" binding:"omitempty"`
	Section     string       `json:"section" binding:"omitempty,max=50"`
	UnitType    UnitType     `json:"unit_type" binding:"required,oneof=Shop Apartment Office Parking Storage Other"`
	MonthlyRent float64      `json:"monthly_rent" binding:"required,gt=0"`
	Metadata    UnitMetadata `json:"metadata" binding:"omitempty"`
}

// UpdateUnitRequest represents the request to update a unit
type UpdateUnitRequest struct {
	UnitName    *string       `json:"unit_name" binding:"omitempty,max=100"`
	Floor       *int          `json:"floor" binding:"omitempty"`
	Section     *string       `json:"section" binding:"omitempty,max=50"`
	UnitType    *UnitType     `json:"unit_type" binding:"omitempty,oneof=Shop Apartment Office Parking Storage Other"`
	MonthlyRent *float64      `json:"monthly_rent" binding:"omitempty,gt=0"`
	Metadata    *UnitMetadata `json:"metadata" binding:"omitempty"`
	Active      *bool         `json:"active"`
}

// UnitWithDetails includes unit with property and building information
type UnitWithDetails struct {
	Unit
	PropertyName string `json:"property_name"`
	BuildingName string `json:"building_name"`
	BuildingCode string `json:"building_code"`
	TenantName   string `json:"tenant_name,omitempty"`
	LeaseActive  bool   `json:"lease_active"`
}

// ShopMetadata represents shop-specific attributes
type ShopMetadata struct {
	AreaSqft     float64 `json:"area_sqft,omitempty"`
	Category     string  `json:"category,omitempty"` // retail, food, service
	HasUtilities bool    `json:"has_utilities,omitempty"`
	FrontageFeet float64 `json:"frontage_feet,omitempty"`
}

// ApartmentMetadata represents apartment-specific attributes
type ApartmentMetadata struct {
	Bedrooms        int     `json:"bedrooms,omitempty"`
	Bathrooms       int     `json:"bathrooms,omitempty"`
	AreaSqft        float64 `json:"area_sqft,omitempty"`
	Furnished       bool    `json:"furnished,omitempty"`
	BalconyCount    int     `json:"balcony_count,omitempty"`
	ParkingIncluded bool    `json:"parking_included,omitempty"`
}

// OfficeMetadata represents office-specific attributes
type OfficeMetadata struct {
	AreaSqft         float64 `json:"area_sqft,omitempty"`
	CabinCount       int     `json:"cabin_count,omitempty"`
	Workstations     int     `json:"workstations,omitempty"`
	ConferenceRoom   bool    `json:"conference_room,omitempty"`
	InternetIncluded bool    `json:"internet_included,omitempty"`
}

// ParkingMetadata represents parking-specific attributes
type ParkingMetadata struct {
	SlotNumber  string `json:"slot_number,omitempty"`
	VehicleType string `json:"vehicle_type,omitempty"` // car, bike, both
	Covered     bool   `json:"covered,omitempty"`
	EVCharging  bool   `json:"ev_charging,omitempty"`
}

// StorageMetadata represents storage-specific attributes
type StorageMetadata struct {
	AreaSqft          float64 `json:"area_sqft,omitempty"`
	ClimateControlled bool    `json:"climate_controlled,omitempty"`
	AccessHours       string  `json:"access_hours,omitempty"`
	SecurityLevel     string  `json:"security_level,omitempty"`
}
