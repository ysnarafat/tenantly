package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// BuildingType represents the type of building
type BuildingType string

const (
	BuildingTypeResidential BuildingType = "Residential"
	BuildingTypeCommercial  BuildingType = "Commercial"
	BuildingTypeMixed       BuildingType = "Mixed"
)

// BuildingMetadata stores building-specific attributes
type BuildingMetadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m BuildingMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *BuildingMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(BuildingMetadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Building represents a physical structure within a property
type Building struct {
	ID               int              `json:"id" db:"id"`
	PropertyID       int              `json:"property_id" db:"property_id"`
	BuildingName     string           `json:"building_name" db:"building_name"`
	BuildingCode     string           `json:"building_code" db:"building_code"`
	BuildingType     BuildingType     `json:"building_type" db:"building_type"`
	TotalFloors      int              `json:"total_floors" db:"total_floors"`
	HasElevator      bool             `json:"has_elevator" db:"has_elevator"`
	ConstructionYear int              `json:"construction_year" db:"construction_year"`
	Metadata         BuildingMetadata `json:"metadata" db:"metadata"`
	Active           bool             `json:"active" db:"active"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateBuildingRequest represents the request to create a building
type CreateBuildingRequest struct {
	PropertyID       int              `json:"property_id" binding:"required"`
	BuildingName     string           `json:"building_name" binding:"required,max=100"`
	BuildingCode     string           `json:"building_code" binding:"required,max=20"`
	BuildingType     BuildingType     `json:"building_type" binding:"required,oneof=Residential Commercial Mixed"`
	TotalFloors      int              `json:"total_floors" binding:"omitempty,gte=0"`
	HasElevator      bool             `json:"has_elevator"`
	ConstructionYear int              `json:"construction_year" binding:"omitempty,gte=1900,lte=2100"`
	Metadata         BuildingMetadata `json:"metadata" binding:"omitempty"`
}

// UpdateBuildingRequest represents the request to update a building
type UpdateBuildingRequest struct {
	BuildingName     *string           `json:"building_name" binding:"omitempty,max=100"`
	BuildingType     *BuildingType     `json:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	TotalFloors      *int              `json:"total_floors" binding:"omitempty,gte=0"`
	HasElevator      *bool             `json:"has_elevator"`
	ConstructionYear *int              `json:"construction_year" binding:"omitempty,gte=1900,lte=2100"`
	Metadata         *BuildingMetadata `json:"metadata" binding:"omitempty"`
	Active           *bool             `json:"active"`
}

// BuildingWithStats includes building with aggregated statistics
type BuildingWithStats struct {
	Building
	PropertyName  string  `json:"property_name"`
	UnitCount     int     `json:"unit_count"`
	OccupiedUnits int     `json:"occupied_units"`
	TotalRevenue  float64 `json:"total_revenue"`
	OccupancyRate float64 `json:"occupancy_rate"`
}
