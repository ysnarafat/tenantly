package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PropertyType represents the type of property
type PropertyType string

const (
	PropertyTypeResidential PropertyType = "Residential"
	PropertyTypeCommercial  PropertyType = "Commercial"
	PropertyTypeMixed       PropertyType = "Mixed"
)

// PropertyMetadata stores property-specific attributes
type PropertyMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m PropertyMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *PropertyMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(PropertyMetadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Property represents a top-level real estate asset
type Property struct {
	ID             int              `json:"id" db:"id"`
	PropertyName   string           `json:"property_name" db:"property_name"`
	PropertyCode   string           `json:"property_code" db:"property_code"`
	Address        string           `json:"address" db:"address"`
	City           *string          `json:"city" db:"city"`
	PostalCode     *string          `json:"postal_code" db:"postal_code"`
	PropertyType   PropertyType     `json:"property_type" db:"property_type"`
	TotalBuildings int              `json:"total_buildings" db:"total_buildings"`
	Metadata       PropertyMetadata `json:"metadata" db:"metadata"`
	Active         bool             `json:"active" db:"active"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

// CreatePropertyRequest represents the request to create a property
type CreatePropertyRequest struct {
	PropertyName string           `json:"property_name" binding:"required,max=200"`
	PropertyCode string           `json:"property_code" binding:"required,max=50"`
	Address      string           `json:"address" binding:"required"`
	City         string           `json:"city" binding:"omitempty,max=100"`
	PostalCode   string           `json:"postal_code" binding:"omitempty,max=20"`
	PropertyType PropertyType     `json:"property_type" binding:"required,oneof=Residential Commercial Mixed"`
	Metadata     PropertyMetadata `json:"metadata" binding:"omitempty"`
}

// UpdatePropertyRequest represents the request to update a property
type UpdatePropertyRequest struct {
	PropertyName   *string           `json:"property_name" binding:"omitempty,max=200"`
	Address        *string           `json:"address" binding:"omitempty"`
	City           *string           `json:"city" binding:"omitempty,max=100"`
	PostalCode     *string           `json:"postal_code" binding:"omitempty,max=20"`
	PropertyType   *PropertyType     `json:"property_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	TotalBuildings *int              `json:"total_buildings" binding:"omitempty,gte=0"`
	Metadata       *PropertyMetadata `json:"metadata" binding:"omitempty"`
	Active         *bool             `json:"active"`
}

// PropertyWithStats includes property with aggregated statistics
type PropertyWithStats struct {
	Property
	BuildingCount int     `json:"building_count"`
	UnitCount     int     `json:"unit_count"`
	OccupiedUnits int     `json:"occupied_units"`
	TotalRevenue  float64 `json:"total_revenue"`
}
