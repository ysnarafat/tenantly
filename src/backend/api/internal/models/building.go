package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	OrganizationID   int              `json:"organization_id" db:"organization_id"`
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
	OrganizationID   int              `json:"-"`
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

// BulkCreateBuildingsRequest represents the request to create multiple buildings
type BulkCreateBuildingsRequest struct {
	PropertyID     int                     `json:"property_id" binding:"required"`
	Buildings      []CreateBuildingRequest `json:"buildings" binding:"required,min=1"`
	OrganizationID int                     `json:"-"`
}

// BuildingSearchFilters represents filters for building search
type BuildingSearchFilters struct {
	OrganizationID *int          `json:"organization_id"`
	PropertyID     *int          `json:"property_id"`
	BuildingType   *BuildingType `json:"building_type"`
	ActiveStatus   *bool         `json:"active_status"`
	HasElevator    *bool         `json:"has_elevator"`
	MinFloors      *int          `json:"min_floors"`
	MaxFloors      *int          `json:"max_floors"`
	Limit          int           `json:"limit"`
	Offset         int           `json:"offset"`
}

// Building validation error constants
const (
	ErrBuildingNotFound          = "BUILDING_NOT_FOUND"
	ErrBuildingCodeExists        = "BUILDING_CODE_EXISTS"
	ErrInvalidBuildingType       = "INVALID_BUILDING_TYPE"
	ErrInvalidMetadata           = "INVALID_METADATA"
	ErrPropertyNotFound          = "PROPERTY_NOT_FOUND"
	ErrBuildingHasActiveUnits    = "BUILDING_HAS_ACTIVE_UNITS"
	ErrInvalidFloorCount         = "INVALID_FLOOR_COUNT"
	ErrInvalidConstructionYear   = "INVALID_CONSTRUCTION_YEAR"
	ErrMetadataValidationFailed  = "METADATA_VALIDATION_FAILED"
	ErrInvalidAmenities          = "INVALID_AMENITIES"
	ErrInvalidSecurityType       = "INVALID_SECURITY_TYPE"
	ErrInvalidParkingSpaces      = "INVALID_PARKING_SPACES"
	ErrInvalidBusinessHours      = "INVALID_BUSINESS_HOURS"
	ErrInvalidFacilities         = "INVALID_FACILITIES"
	ErrInvalidResidentialSection = "INVALID_RESIDENTIAL_SECTION"
	ErrInvalidCommercialSection  = "INVALID_COMMERCIAL_SECTION"
	ErrInvalidSharedFacilities   = "INVALID_SHARED_FACILITIES"
)

// BuildingError represents a building-specific error
type BuildingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *BuildingError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewBuildingError creates a new BuildingError
func NewBuildingError(code, message, field string) *BuildingError {
	return &BuildingError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// BuildingSearchRequest represents advanced search request for buildings
type BuildingSearchRequest struct {
	PropertyID       *int   `form:"property_id"`
	BuildingType     string `form:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	ActiveStatus     *bool  `form:"active_status"`
	HasElevator      *bool  `form:"has_elevator"`
	MinFloors        *int   `form:"min_floors" binding:"omitempty,gte=1"`
	MaxFloors        *int   `form:"max_floors" binding:"omitempty,gte=1"`
	ConstructionYear *int   `form:"construction_year" binding:"omitempty,gte=1800"`
	SearchTerm       string `form:"search_term"`
	MetadataQuery    string `form:"metadata_query"`
	Page             int    `form:"page" binding:"omitempty,gte=1"`
	PageSize         int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	SortBy           string `form:"sort_by" binding:"omitempty,oneof=building_name building_code created_at updated_at total_floors construction_year"`
	SortOrder        string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	IncludeStats     bool   `form:"include_stats"`
	OrganizationID   *int   `form:"-"`
}

// BuildingExportRequest represents request for building data export
type BuildingExportRequest struct {
	PropertyID   *int   `form:"property_id"`
	BuildingType string `form:"building_type" binding:"omitempty,oneof=Residential Commercial Mixed"`
	ActiveStatus *bool  `form:"active_status"`
	Format       string `form:"format" binding:"omitempty,oneof=csv json xlsx"`
	IncludeStats bool   `form:"include_stats"`
}

// BuildingStatusRequest represents request for building status management
type BuildingStatusRequest struct {
	ActiveStatus bool   `json:"active_status" binding:"required"`
	Reason       string `json:"reason" binding:"omitempty,max=255"`
}

// MetadataSchemaResponse represents metadata schema information for building types
type MetadataSchemaResponse struct {
	BuildingType string                 `json:"building_type"`
	Schema       map[string]interface{} `json:"schema"`
	Examples     map[string]interface{} `json:"examples"`
	Description  string                 `json:"description"`
}

// BuildingUnitSummary represents unit summary for a building
type BuildingUnitSummary struct {
	UnitID      int    `json:"unit_id"`
	UnitNumber  string `json:"unit_number"`
	UnitName    string `json:"unit_name"`
	Floor       int    `json:"floor"`
	Section     string `json:"section"`
	UnitType    string `json:"unit_type"`
	Active      bool   `json:"active"`
	TenantName  string `json:"tenant_name,omitempty"`
	LeaseActive bool   `json:"lease_active"`
}

// BuildingUnitsResponse represents response for building units endpoint
type BuildingUnitsResponse struct {
	BuildingID   int                    `json:"building_id"`
	BuildingName string                 `json:"building_name"`
	BuildingCode string                 `json:"building_code"`
	Units        []*BuildingUnitSummary `json:"units"`
	Summary      struct {
		TotalUnits    int     `json:"total_units"`
		OccupiedUnits int     `json:"occupied_units"`
		VacantUnits   int     `json:"vacant_units"`
		TotalRevenue  float64 `json:"total_revenue"`
	} `json:"summary"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}
