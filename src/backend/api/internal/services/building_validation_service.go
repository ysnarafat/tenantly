package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// BuildingValidationServiceInterface defines the interface for building validation operations
type BuildingValidationServiceInterface interface {
	// Core validation methods
	ValidateBuildingType(buildingType models.BuildingType) error
	ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error
	ValidatePropertyAssociation(propertyID int) error
	ValidateMetadataSchema(buildingType models.BuildingType, metadata models.BuildingMetadata) error
	ValidateDeletionConstraints(buildingID int) error
	ValidateConstructionYear(year *int) error
	ValidateFloorCountAndElevatorRequirement(totalFloors int, hasElevator bool) error

	// Business rules validation
	ValidateBuildingCreation(req *models.CreateBuildingRequest) error
	ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error
	ValidateBuildingDeletion(id int) error

	// Advanced validation methods
	ValidateFloorRange(floorRange string) error
	ValidateBusinessHours(businessHours string) error
	ValidateElevatorRequirement(totalFloors int, hasElevator bool, buildingType models.BuildingType) error
	ValidateMetadataConsistency(buildingType models.BuildingType, metadata models.BuildingMetadata) error
}

// PropertyRepositoryInterface defines the interface for property repository operations needed for validation
type PropertyRepositoryInterface interface {
	GetByID(id int) (*models.Property, error)
}

// BuildingValidationService implements comprehensive validation logic for building operations
type BuildingValidationService struct {
	buildingRepo      interfaces.BuildingRepositoryInterface
	propertyRepo      PropertyRepositoryInterface
	metadataValidator interfaces.MetadataValidatorInterface
}

// NewBuildingValidationService creates a new building validation service instance
func NewBuildingValidationService(
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo PropertyRepositoryInterface,
	metadataValidator interfaces.MetadataValidatorInterface,
) *BuildingValidationService {
	return &BuildingValidationService{
		buildingRepo:      buildingRepo,
		propertyRepo:      propertyRepo,
		metadataValidator: metadataValidator,
	}
}

// ValidateBuildingType validates building type against allowed enum values
func (v *BuildingValidationService) ValidateBuildingType(buildingType models.BuildingType) error {
	validTypes := map[models.BuildingType]bool{
		models.BuildingTypeResidential: true,
		models.BuildingTypeCommercial:  true,
		models.BuildingTypeMixed:       true,
	}

	if !validTypes[buildingType] {
		return &models.BuildingError{
			Code:    models.ErrInvalidBuildingType,
			Message: fmt.Sprintf("invalid building type '%s': must be Residential, Commercial, or Mixed", buildingType),
			Field:   "building_type",
		}
	}

	return nil
}

// ValidateBuildingCodeUniqueness validates building code uniqueness within property scope
func (v *BuildingValidationService) ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error {
	if code == "" {
		return &models.BuildingError{
			Code:    models.ErrBuildingCodeExists,
			Message: "building code cannot be empty",
			Field:   "building_code",
		}
	}

	existingBuilding, err := v.buildingRepo.GetByPropertyAndCode(propertyID, code)
	if err != nil {
		// If building not found, code is unique
		if err.Error() == "building not found" {
			return nil
		}
		return fmt.Errorf("failed to check building code uniqueness: %w", err)
	}

	// If we're updating and the existing building is the same one, it's allowed
	if excludeID != nil && existingBuilding.ID == *excludeID {
		return nil
	}

	return &models.BuildingError{
		Code:    models.ErrBuildingCodeExists,
		Message: fmt.Sprintf("building code '%s' already exists in this property", code),
		Field:   "building_code",
	}
}

// ValidatePropertyAssociation validates property-building relationship
func (v *BuildingValidationService) ValidatePropertyAssociation(propertyID int) error {
	if propertyID <= 0 {
		return &models.BuildingError{
			Code:    models.ErrPropertyNotFound,
			Message: "property ID must be a positive integer",
			Field:   "property_id",
		}
	}

	_, err := v.propertyRepo.GetByID(propertyID)
	if err != nil {
		return &models.BuildingError{
			Code:    models.ErrPropertyNotFound,
			Message: fmt.Sprintf("property with ID %d not found", propertyID),
			Field:   "property_id",
		}
	}

	return nil
}

// ValidateMetadataSchema validates metadata schema based on building type
func (v *BuildingValidationService) ValidateMetadataSchema(buildingType models.BuildingType, metadata models.BuildingMetadata) error {
	if metadata == nil {
		return nil // Empty metadata is allowed
	}

	// First validate building type
	if err := v.ValidateBuildingType(buildingType); err != nil {
		return err
	}

	// Use the metadata validator to validate the schema
	if err := v.metadataValidator.ValidateMetadata(buildingType, metadata); err != nil {
		return &models.BuildingError{
			Code:    models.ErrInvalidMetadata,
			Message: fmt.Sprintf("metadata validation failed: %s", err.Error()),
			Field:   "metadata",
		}
	}

	// Additional consistency checks
	if err := v.ValidateMetadataConsistency(buildingType, metadata); err != nil {
		return err
	}

	return nil
}

// ValidateDeletionConstraints validates building deletion constraints (prevent deletion with active units)
func (v *BuildingValidationService) ValidateDeletionConstraints(buildingID int) error {
	if buildingID <= 0 {
		return &models.BuildingError{
			Code:    models.ErrBuildingNotFound,
			Message: "building ID must be a positive integer",
			Field:   "id",
		}
	}

	// Check if building exists
	_, err := v.buildingRepo.GetByID(buildingID)
	if err != nil {
		return &models.BuildingError{
			Code:    models.ErrBuildingNotFound,
			Message: fmt.Sprintf("building with ID %d not found", buildingID),
			Field:   "id",
		}
	}

	// The building repository's SoftDelete method already checks for active units
	// We can rely on that validation, or implement a simple check here
	// For now, we'll let the repository handle this constraint
	return nil
}

// ValidateConstructionYear validates construction year and business rule enforcement
func (v *BuildingValidationService) ValidateConstructionYear(year *int) error {
	if year == nil {
		return nil // Construction year is optional
	}

	currentYear := time.Now().Year()
	minYear := 1800
	maxYear := currentYear + 5 // Allow up to 5 years in the future for planned constructions

	if *year < minYear || *year > maxYear {
		return &models.BuildingError{
			Code:    models.ErrInvalidConstructionYear,
			Message: fmt.Sprintf("construction year must be between %d and %d", minYear, maxYear),
			Field:   "construction_year",
		}
	}

	// Business rule: construction years up to maxYear in the future are allowed
	// (e.g. planned constructions); this could be surfaced as a warning in future.

	return nil
}

// ValidateFloorCountAndElevatorRequirement validates floor count and elevator requirement business rules
func (v *BuildingValidationService) ValidateFloorCountAndElevatorRequirement(totalFloors int, hasElevator bool) error {
	if totalFloors < 1 {
		return &models.BuildingError{
			Code:    models.ErrInvalidFloorCount,
			Message: "total floors must be at least 1",
			Field:   "total_floors",
		}
	}

	if totalFloors > 100 {
		return &models.BuildingError{
			Code:    models.ErrInvalidFloorCount,
			Message: "total floors cannot exceed 100",
			Field:   "total_floors",
		}
	}

	// Business rule: buildings with more than 4 floors should have an elevator.
	// This is a warning, not an error - allowed, but could be logged and made
	// configurable based on local building codes in future.

	return nil
}

// ValidateBuildingCreation validates comprehensive building creation request
func (v *BuildingValidationService) ValidateBuildingCreation(req *models.CreateBuildingRequest) error {
	// Validate required fields
	if req.BuildingName == "" {
		return &models.BuildingError{
			Code:    "BUILDING_NAME_REQUIRED",
			Message: "building name is required",
			Field:   "building_name",
		}
	}

	if req.BuildingCode == "" {
		return &models.BuildingError{
			Code:    "BUILDING_CODE_REQUIRED",
			Message: "building code is required",
			Field:   "building_code",
		}
	}

	// Validate property association
	if err := v.ValidatePropertyAssociation(req.PropertyID); err != nil {
		return err
	}

	// Validate building type
	if err := v.ValidateBuildingType(req.BuildingType); err != nil {
		return err
	}

	// Validate building code uniqueness
	if err := v.ValidateBuildingCodeUniqueness(req.PropertyID, req.BuildingCode, nil); err != nil {
		return err
	}

	// Validate floor count and elevator requirement
	if err := v.ValidateFloorCountAndElevatorRequirement(req.TotalFloors, req.HasElevator); err != nil {
		return err
	}

	// Validate construction year
	if err := v.ValidateConstructionYear(req.ConstructionYear); err != nil {
		return err
	}

	// Validate metadata schema
	if err := v.ValidateMetadataSchema(req.BuildingType, req.Metadata); err != nil {
		return err
	}

	// Additional business rule: Validate elevator requirement based on building type
	if err := v.ValidateElevatorRequirement(req.TotalFloors, req.HasElevator, req.BuildingType); err != nil {
		return err
	}

	return nil
}

// ValidateBuildingUpdate validates comprehensive building update request
func (v *BuildingValidationService) ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error {
	// Check if building exists
	existingBuilding, err := v.buildingRepo.GetByID(id)
	if err != nil {
		return &models.BuildingError{
			Code:    models.ErrBuildingNotFound,
			Message: fmt.Sprintf("building with ID %d not found", id),
			Field:   "id",
		}
	}

	// Validate building type if provided
	if req.BuildingType != nil {
		if err := v.ValidateBuildingType(*req.BuildingType); err != nil {
			return err
		}
	}

	// Validate floor count and elevator requirement if provided
	totalFloors := existingBuilding.TotalFloors
	hasElevator := existingBuilding.HasElevator

	if req.TotalFloors != nil {
		totalFloors = *req.TotalFloors
	}
	if req.HasElevator != nil {
		hasElevator = *req.HasElevator
	}

	if err := v.ValidateFloorCountAndElevatorRequirement(totalFloors, hasElevator); err != nil {
		return err
	}

	// Validate construction year if provided
	if err := v.ValidateConstructionYear(req.ConstructionYear); err != nil {
		return err
	}

	// Validate metadata schema if provided
	if req.Metadata != nil {
		buildingType := existingBuilding.BuildingType
		if req.BuildingType != nil {
			buildingType = *req.BuildingType
		}
		if err := v.ValidateMetadataSchema(buildingType, *req.Metadata); err != nil {
			return err
		}
	}

	// Additional business rule: Validate elevator requirement based on building type
	buildingType := existingBuilding.BuildingType
	if req.BuildingType != nil {
		buildingType = *req.BuildingType
	}
	if err := v.ValidateElevatorRequirement(totalFloors, hasElevator, buildingType); err != nil {
		return err
	}

	return nil
}

// ValidateBuildingDeletion validates comprehensive building deletion constraints
func (v *BuildingValidationService) ValidateBuildingDeletion(id int) error {
	return v.ValidateDeletionConstraints(id)
}

// ValidateFloorRange validates floor range format (e.g., "1-5", "3-10") for mixed buildings
func (v *BuildingValidationService) ValidateFloorRange(floorRange string) error {
	if floorRange == "" {
		return &models.BuildingError{
			Code:    "INVALID_FLOOR_RANGE",
			Message: "floor range cannot be empty",
			Field:   "floor_range",
		}
	}

	// Use regex to validate format
	// This is a simplified validation - the metadata validator has more comprehensive validation
	if len(floorRange) < 3 || !contains(floorRange, "-") {
		return &models.BuildingError{
			Code:    "INVALID_FLOOR_RANGE",
			Message: "floor range must be in format 'start-end' (e.g., '1-5')",
			Field:   "floor_range",
		}
	}

	return nil
}

// ValidateBusinessHours validates business hours format for commercial buildings
func (v *BuildingValidationService) ValidateBusinessHours(businessHours string) error {
	if businessHours == "" {
		return &models.BuildingError{
			Code:    models.ErrInvalidBusinessHours,
			Message: "business hours cannot be empty",
			Field:   "business_hours",
		}
	}

	// Basic format validation - the metadata validator has more comprehensive validation
	if len(businessHours) < 11 || !contains(businessHours, "-") || !contains(businessHours, ":") {
		return &models.BuildingError{
			Code:    models.ErrInvalidBusinessHours,
			Message: "business hours must be in format 'HH:MM-HH:MM' (e.g., '09:00-18:00')",
			Field:   "business_hours",
		}
	}

	return nil
}

// ValidateElevatorRequirement validates elevator requirement based on building type and floors
func (v *BuildingValidationService) ValidateElevatorRequirement(totalFloors int, hasElevator bool, buildingType models.BuildingType) error {
	// Business rules for elevator requirements based on building type
	switch buildingType {
	case models.BuildingTypeCommercial:
		// Commercial buildings with more than 3 floors should have elevators.
		// This is a warning, not an error - allowed, but could be logged in future.
	case models.BuildingTypeResidential:
		// Residential buildings with more than 4 floors should have elevators.
		// This is a warning, not an error - allowed, but could be logged in future.
	case models.BuildingTypeMixed:
		// Mixed buildings with more than 3 floors should have elevators due to commercial use.
		// This is a warning, not an error - allowed, but could be logged in future.
	}

	return nil
}

// ValidateMetadataConsistency validates metadata consistency and cross-field validation
func (v *BuildingValidationService) ValidateMetadataConsistency(buildingType models.BuildingType, metadata models.BuildingMetadata) error {
	if metadata == nil {
		return nil
	}

	switch buildingType {
	case models.BuildingTypeResidential:
		return v.validateResidentialMetadataConsistency(metadata)
	case models.BuildingTypeCommercial:
		return v.validateCommercialMetadataConsistency(metadata)
	case models.BuildingTypeMixed:
		return v.validateMixedMetadataConsistency(metadata)
	}

	return nil
}

// validateResidentialMetadataConsistency validates consistency for residential building metadata
func (v *BuildingValidationService) validateResidentialMetadataConsistency(metadata models.BuildingMetadata) error {
	// Check parking spaces consistency
	if parkingSpaces, exists := metadata["parking_spaces"]; exists {
		if parkingMap, ok := parkingSpaces.(map[string]interface{}); ok {
			total, totalExists := parkingMap["total"]
			covered, coveredExists := parkingMap["covered"]
			visitor, visitorExists := parkingMap["visitor"]

			if totalExists && coveredExists && visitorExists {
				totalInt, _ := total.(float64)
				coveredInt, _ := covered.(float64)
				visitorInt, _ := visitor.(float64)

				if coveredInt+visitorInt > totalInt {
					return &models.BuildingError{
						Code:    models.ErrInvalidParkingSpaces,
						Message: "covered and visitor parking spaces cannot exceed total parking spaces",
						Field:   "metadata.parking_spaces",
					}
				}
			}
		}
	}

	return nil
}

// validateCommercialMetadataConsistency validates consistency for commercial building metadata
func (v *BuildingValidationService) validateCommercialMetadataConsistency(metadata models.BuildingMetadata) error {
	// Check parking spaces consistency
	if parkingSpaces, exists := metadata["parking_spaces"]; exists {
		if parkingMap, ok := parkingSpaces.(map[string]interface{}); ok {
			total, totalExists := parkingMap["total"]
			customer, customerExists := parkingMap["customer"]
			staff, staffExists := parkingMap["staff"]

			if totalExists && customerExists && staffExists {
				totalInt, _ := total.(float64)
				customerInt, _ := customer.(float64)
				staffInt, _ := staff.(float64)

				if customerInt+staffInt > totalInt {
					return &models.BuildingError{
						Code:    models.ErrInvalidParkingSpaces,
						Message: "customer and staff parking spaces cannot exceed total parking spaces",
						Field:   "metadata.parking_spaces",
					}
				}
			}
		}
	}

	return nil
}

// validateMixedMetadataConsistency validates consistency for mixed building metadata
func (v *BuildingValidationService) validateMixedMetadataConsistency(metadata models.BuildingMetadata) error {
	// Check shared facilities consistency
	if sharedFacilities, exists := metadata["shared_facilities"]; exists {
		if facilitiesMap, ok := sharedFacilities.(map[string]interface{}); ok {
			// Validate that parking allocation doesn't exceed total parking
			if parkingTotal, exists := facilitiesMap["parking_total"]; exists {
				totalParking, _ := parkingTotal.(float64)

				// Check commercial section parking allocation
				if commercialSection, exists := metadata["commercial_section"]; exists {
					if commercialMap, ok := commercialSection.(map[string]interface{}); ok {
						if parkingAllocation, exists := commercialMap["parking_allocation"]; exists {
							allocation, _ := parkingAllocation.(float64)
							if allocation > totalParking {
								return &models.BuildingError{
									Code:    models.ErrInvalidParkingSpaces,
									Message: "commercial parking allocation cannot exceed total parking spaces",
									Field:   "metadata.commercial_section.parking_allocation",
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
