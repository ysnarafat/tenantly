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
	ConstructionYear *int             `json:"construction_year" db:"construction_year"`
	Metadata         BuildingMetadata `json:"metadata" db:"metadata"`
	ActiveStatus     bool             `json:"active_status" db:"active_status"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateBuildingRequest represents the request to create a building
type CreateBuildingRequest struct {
	PropertyID       int              `json:"property_id" binding:"required"`
	BuildingName     string           `json:"building_name" binding:"required,max=100"`
	BuildingCode     string           `json:"building_code" binding:"required,max=50"`
	BuildingType     BuildingType     `json:"building_type" binding:"required,oneof=Residential Commercial Mixed"`
	TotalFloors      int              `json:"total_floors" binding:"required,gte=1"`
	HasElevator      bool             `json:"has_elevator"`
	ConstructionYear *int             `json:"construction_year" binding:"omitempty,gte=1800"`
	Metadata         BuildingMetadata `json:"metadata" binding:"omitempty"`
	ActiveStatus     bool             `json:"active_status"`
}

// UpdateBuildingRequest represents the request to update a building
type UpdateBuildingRequest struct {
	BuildingName     *string           `json:"building_name" binding:"omitempty,max=100"`
	BuildingType     *BuildingType     `json:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	TotalFloors      *int              `json:"total_floors" binding:"omitempty,gte=1"`
	HasElevator      *bool             `json:"has_elevator"`
	ConstructionYear *int              `json:"construction_year" binding:"omitempty,gte=1800"`
	Metadata         *BuildingMetadata `json:"metadata" binding:"omitempty"`
	ActiveStatus     *bool             `json:"active_status"`
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

// PaginationInfo represents common pagination metadata
type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	PageSize    int  `json:"page_size"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// BulkCreateBuildingsRequest represents the request to create multiple buildings
type BulkCreateBuildingsRequest struct {
	PropertyID int                     `json:"property_id" binding:"required"`
	Buildings  []CreateBuildingRequest `json:"buildings" binding:"required,min=1"`
}

// BuildingSearchFilters represents filters for building search
type BuildingSearchFilters struct {
	PropertyID   *int          `json:"property_id"`
	BuildingType *BuildingType `json:"building_type"`
	ActiveStatus *bool         `json:"active_status"`
	HasElevator  *bool         `json:"has_elevator"`
	MinFloors    *int          `json:"min_floors"`
	MaxFloors    *int          `json:"max_floors"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
}

// BuildingSearchRequest represents advanced search request
type BuildingSearchRequest struct {
	PropertyID       *int   `json:"property_id"`
	BuildingType     string `json:"building_type"`
	ActiveStatus     *bool  `json:"active_status"`
	HasElevator      *bool  `json:"has_elevator"`
	MinFloors        *int   `json:"min_floors"`
	MaxFloors        *int   `json:"max_floors"`
	ConstructionYear *int   `json:"construction_year"`
	SearchTerm       string `json:"search_term"`
	MetadataQuery    string `json:"metadata_query"`
	SortBy           string `json:"sort_by"`
	SortOrder        string `json:"sort_order"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
	IncludeStats     bool   `json:"include_stats"`
}

// BuildingUnitSummary represents a summary of a unit in a building
type BuildingUnitSummary struct {
	UnitID      int      `json:"unit_id"`
	UnitNumber  string   `json:"unit_number"`
	UnitName    *string  `json:"unit_name"`
	Floor       int      `json:"floor"`
	Section     *string  `json:"section"`
	UnitType    UnitType `json:"unit_type"`
	MonthlyRent float64  `json:"monthly_rent"`
	Active      bool     `json:"active"`
	TenantName  string   `json:"tenant_name"`
	LeaseActive bool     `json:"lease_active"`
}

// BuildingListResponse represents the response for a list of buildings
type BuildingListResponse struct {
	Buildings  []*BuildingWithStats `json:"buildings"`
	Pagination *PaginationInfo      `json:"pagination"`
	Statistics *PropertyStatistics  `json:"statistics,omitempty"`
}

// PropertyStatistics represents aggregated statistics for a property
type PropertyStatistics struct {
	TotalUnits     int     `json:"total_units"`
	OccupiedUnits  int     `json:"occupied_units"`
	VacantUnits    int     `json:"vacant_units"`
	TotalRevenue   float64 `json:"total_revenue"`
	OccupancyRate  float64 `json:"occupancy_rate"`
	AverageRevenue float64 `json:"average_revenue"`
}

// PropertyWithBuildings represents a property with its buildings
type PropertyWithBuildings struct {
	Property      Property            `json:"property"`
	Buildings     []*Building         `json:"buildings"`
	BuildingCount int                 `json:"building_count"`
	Statistics    *PropertyStatistics `json:"statistics,omitempty"`
}

// BuildingUnitsResponse represents the response for building units
type BuildingUnitsResponse struct {
	BuildingID   int                    `json:"building_id"`
	BuildingName string                 `json:"building_name"`
	BuildingCode string                 `json:"building_code"`
	Units        []*BuildingUnitSummary `json:"units"`
	Pagination   *PaginationInfo        `json:"pagination"`
	Summary      struct {
		TotalUnits    int     `json:"total_units"`
		OccupiedUnits int     `json:"occupied_units"`
		VacantUnits   int     `json:"vacant_units"`
		TotalRevenue  float64 `json:"total_revenue"`
	} `json:"summary"`
}

// MetadataSchemaResponse represents the schema for building metadata
type MetadataSchemaResponse struct {
	BuildingType string                 `json:"building_type"`
	Schema       map[string]interface{} `json:"schema"`
	Examples     map[string]interface{} `json:"examples"`
	Description  string                 `json:"description"`
}

// BuildingAnalytics represents detailed analytics for a building
type BuildingAnalytics struct {
	BuildingID     int     `json:"building_id"`
	UnitCount      int     `json:"unit_count"`
	OccupiedUnits  int     `json:"occupied_units"`
	VacantUnits    int     `json:"vacant_units"`
	OccupancyRate  float64 `json:"occupancy_rate"`
	MonthlyRevenue float64 `json:"monthly_revenue"`
	AverageRent    float64 `json:"average_rent"`
	TotalArea      float64 `json:"total_area"`
}
