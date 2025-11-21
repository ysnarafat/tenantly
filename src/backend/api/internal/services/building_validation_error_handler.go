package services

import (
	"fmt"
	"strings"

	"github.com/ysnarafat/tenantly/internal/models"
)

// ValidationErrorHandler handles building validation errors
type ValidationErrorHandler struct{}

// NewValidationErrorHandler creates a new ValidationErrorHandler
func NewValidationErrorHandler() *ValidationErrorHandler {
	return &ValidationErrorHandler{}
}

// HandleMetadataValidationError converts metadata validation errors to BuildingError
func (h *ValidationErrorHandler) HandleMetadataValidationError(err error, buildingType models.BuildingType) *models.BuildingError {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()

	// Parse the error message to determine the specific validation failure
	switch {
	case strings.Contains(errorMessage, "amenities"):
		return h.createMetadataError(models.ErrInvalidAmenities, errorMessage, "amenities")
	case strings.Contains(errorMessage, "security_type"):
		return h.createMetadataError(models.ErrInvalidSecurityType, errorMessage, "security_type")
	case strings.Contains(errorMessage, "parking_spaces"):
		return h.createMetadataError(models.ErrInvalidParkingSpaces, errorMessage, "parking_spaces")
	case strings.Contains(errorMessage, "business_hours"):
		return h.createMetadataError(models.ErrInvalidBusinessHours, errorMessage, "business_hours")
	case strings.Contains(errorMessage, "shared_facilities"):
		return h.createMetadataError(models.ErrInvalidSharedFacilities, errorMessage, "shared_facilities")
	case strings.Contains(errorMessage, "facilities"):
		return h.createMetadataError(models.ErrInvalidFacilities, errorMessage, "facilities")
	case strings.Contains(errorMessage, "residential_section"):
		return h.createMetadataError(models.ErrInvalidResidentialSection, errorMessage, "residential_section")
	case strings.Contains(errorMessage, "commercial_section"):
		return h.createMetadataError(models.ErrInvalidCommercialSection, errorMessage, "commercial_section")
	default:
		return h.createMetadataError(models.ErrMetadataValidationFailed, errorMessage, "metadata")
	}
}

// HandleBuildingValidationError converts general building validation errors to BuildingError
func (h *ValidationErrorHandler) HandleBuildingValidationError(err error) *models.BuildingError {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()

	switch {
	case strings.Contains(errorMessage, "building not found"):
		return models.NewBuildingError(models.ErrBuildingNotFound, errorMessage, "")
	case strings.Contains(errorMessage, "building code") && strings.Contains(errorMessage, "exists"):
		return models.NewBuildingError(models.ErrBuildingCodeExists, errorMessage, "building_code")
	case strings.Contains(errorMessage, "invalid building type"):
		return models.NewBuildingError(models.ErrInvalidBuildingType, errorMessage, "building_type")
	case strings.Contains(errorMessage, "property not found"):
		return models.NewBuildingError(models.ErrPropertyNotFound, errorMessage, "property_id")
	case strings.Contains(errorMessage, "active units"):
		return models.NewBuildingError(models.ErrBuildingHasActiveUnits, errorMessage, "")
	case strings.Contains(errorMessage, "floor"):
		return models.NewBuildingError(models.ErrInvalidFloorCount, errorMessage, "total_floors")
	case strings.Contains(errorMessage, "construction year"):
		return models.NewBuildingError(models.ErrInvalidConstructionYear, errorMessage, "construction_year")
	default:
		return models.NewBuildingError(models.ErrInvalidMetadata, errorMessage, "")
	}
}

// HandleConstraintViolationError handles database constraint violations
func (h *ValidationErrorHandler) HandleConstraintViolationError(err error) *models.BuildingError {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()

	switch {
	case strings.Contains(errorMessage, "unique_building_code_per_property"):
		return models.NewBuildingError(
			models.ErrBuildingCodeExists,
			"Building code already exists within this property",
			"building_code",
		)
	case strings.Contains(errorMessage, "foreign key constraint"):
		if strings.Contains(errorMessage, "property_id") {
			return models.NewBuildingError(
				models.ErrPropertyNotFound,
				"Referenced property does not exist",
				"property_id",
			)
		}
		return models.NewBuildingError(
			models.ErrInvalidMetadata,
			"Foreign key constraint violation",
			"",
		)
	case strings.Contains(errorMessage, "check constraint"):
		if strings.Contains(errorMessage, "valid_floors") {
			return models.NewBuildingError(
				models.ErrInvalidFloorCount,
				"Total floors must be greater than 0",
				"total_floors",
			)
		}
		if strings.Contains(errorMessage, "valid_construction_year") {
			return models.NewBuildingError(
				models.ErrInvalidConstructionYear,
				"Construction year must be between 1800 and current year + 5",
				"construction_year",
			)
		}
		return models.NewBuildingError(
			models.ErrInvalidMetadata,
			"Check constraint violation",
			"",
		)
	default:
		return models.NewBuildingError(
			models.ErrInvalidMetadata,
			fmt.Sprintf("Database constraint violation: %s", errorMessage),
			"",
		)
	}
}

// createMetadataError creates a metadata-specific BuildingError
func (h *ValidationErrorHandler) createMetadataError(code, message, field string) *models.BuildingError {
	return models.NewBuildingError(code, message, field)
}

// GetValidationErrorResponse creates a standardized error response for API
func (h *ValidationErrorHandler) GetValidationErrorResponse(buildingError *models.BuildingError) map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    buildingError.Code,
			"message": buildingError.Message,
		},
	}

	if buildingError.Field != "" {
		response["error"].(map[string]interface{})["field"] = buildingError.Field
	}

	return response
}

// GetMultipleValidationErrorsResponse creates a response for multiple validation errors
func (h *ValidationErrorHandler) GetMultipleValidationErrorsResponse(errors []*models.BuildingError) map[string]interface{} {
	errorList := make([]map[string]interface{}, len(errors))

	for i, err := range errors {
		errorItem := map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
		}
		if err.Field != "" {
			errorItem["field"] = err.Field
		}
		errorList[i] = errorItem
	}

	return map[string]interface{}{
		"errors": errorList,
	}
}

// ValidateAndFormatMetadataErrors validates metadata and returns formatted errors
func (h *ValidationErrorHandler) ValidateAndFormatMetadataErrors(
	validator MetadataValidatorInterface,
	buildingType models.BuildingType,
	metadata models.BuildingMetadata,
) *models.BuildingError {
	if err := validator.ValidateMetadata(buildingType, metadata); err != nil {
		return h.HandleMetadataValidationError(err, buildingType)
	}
	return nil
}
