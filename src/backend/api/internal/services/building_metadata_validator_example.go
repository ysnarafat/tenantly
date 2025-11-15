package services

import (
	"fmt"

	"github.com/ysnarafat/tenantly/internal/models"
)

// BuildingMetadataValidationService demonstrates how to use the metadata validation system
type BuildingMetadataValidationService struct {
	validator    MetadataValidatorInterface
	errorHandler *ValidationErrorHandler
}

// NewBuildingMetadataValidationService creates a new service instance
func NewBuildingMetadataValidationService() *BuildingMetadataValidationService {
	return &BuildingMetadataValidationService{
		validator:    NewBuildingMetadataValidator(),
		errorHandler: NewValidationErrorHandler(),
	}
}

// ValidateBuildingCreation validates building metadata during creation
func (s *BuildingMetadataValidationService) ValidateBuildingCreation(req *models.CreateBuildingRequest) *models.BuildingError {
	if req.Metadata == nil {
		return nil // Empty metadata is allowed
	}

	return s.errorHandler.ValidateAndFormatMetadataErrors(
		s.validator,
		req.BuildingType,
		req.Metadata,
	)
}

// ValidateBuildingUpdate validates building metadata during updates
func (s *BuildingMetadataValidationService) ValidateBuildingUpdate(buildingType models.BuildingType, req *models.UpdateBuildingRequest) *models.BuildingError {
	// If building type is being updated, validate metadata against new type
	targetBuildingType := buildingType
	if req.BuildingType != nil {
		targetBuildingType = *req.BuildingType
	}

	// If metadata is being updated, validate it
	if req.Metadata != nil {
		return s.errorHandler.ValidateAndFormatMetadataErrors(
			s.validator,
			targetBuildingType,
			*req.Metadata,
		)
	}

	return nil
}

// GetMetadataSchemaForType returns the metadata schema for a building type
func (s *BuildingMetadataValidationService) GetMetadataSchemaForType(buildingType models.BuildingType) (map[string]interface{}, error) {
	schema := s.validator.GetMetadataSchema(buildingType)
	if schema == nil {
		return nil, fmt.Errorf("invalid building type: %s", buildingType)
	}
	return schema, nil
}

// ValidateMetadataField validates a specific metadata field
func (s *BuildingMetadataValidationService) ValidateMetadataField(buildingType models.BuildingType, fieldName string, fieldValue interface{}) *models.BuildingError {
	// Create a temporary metadata object with just the field to validate
	tempMetadata := models.BuildingMetadata{
		fieldName: fieldValue,
	}

	return s.errorHandler.ValidateAndFormatMetadataErrors(
		s.validator,
		buildingType,
		tempMetadata,
	)
}

// Example usage functions for demonstration

// ExampleResidentialBuildingValidation demonstrates residential building validation
func ExampleResidentialBuildingValidation() {
	service := NewBuildingMetadataValidationService()

	// Valid residential building metadata
	validMetadata := models.BuildingMetadata{
		"amenities":               []interface{}{"gym", "swimming_pool", "playground"},
		"security_type":           "24_hour_guard",
		"maintenance_staff_count": 3,
		"parking_spaces": map[string]interface{}{
			"total":   50,
			"covered": 30,
			"visitor": 10,
		},
		"utilities": map[string]interface{}{
			"backup_generator": true,
			"water_supply":     "24_hour",
			"internet_ready":   true,
		},
	}

	req := &models.CreateBuildingRequest{
		PropertyID:       1,
		BuildingName:     "Residential Tower A",
		BuildingCode:     "RTA-001",
		BuildingType:     models.BuildingTypeResidential,
		TotalFloors:      10,
		HasElevator:      true,
		ConstructionYear: &[]int{2020}[0],
		Metadata:         validMetadata,
	}

	if err := service.ValidateBuildingCreation(req); err != nil {
		fmt.Printf("Validation failed: %s\n", err.Error())
	} else {
		fmt.Println("Residential building validation passed!")
	}
}

// ExampleCommercialBuildingValidation demonstrates commercial building validation
func ExampleCommercialBuildingValidation() {
	service := NewBuildingMetadataValidationService()

	// Valid commercial building metadata
	validMetadata := models.BuildingMetadata{
		"parking_spaces": map[string]interface{}{
			"total":    100,
			"customer": 80,
			"staff":    20,
		},
		"loading_docks": 3,
		"security_system": map[string]interface{}{
			"type":           "advanced_cctv",
			"access_control": true,
			"fire_safety":    "sprinkler_system",
		},
		"business_hours": map[string]interface{}{
			"weekdays": "09:00-22:00",
			"weekends": "10:00-23:00",
			"holidays": "10:00-20:00",
		},
		"facilities": map[string]interface{}{
			"elevators":  4,
			"escalators": 2,
			"food_court": true,
			"atm":        true,
		},
	}

	req := &models.CreateBuildingRequest{
		PropertyID:       1,
		BuildingName:     "Shopping Mall",
		BuildingCode:     "SM-001",
		BuildingType:     models.BuildingTypeCommercial,
		TotalFloors:      5,
		HasElevator:      true,
		ConstructionYear: &[]int{2019}[0],
		Metadata:         validMetadata,
	}

	if err := service.ValidateBuildingCreation(req); err != nil {
		fmt.Printf("Validation failed: %s\n", err.Error())
	} else {
		fmt.Println("Commercial building validation passed!")
	}
}

// ExampleMixedBuildingValidation demonstrates mixed building validation
func ExampleMixedBuildingValidation() {
	service := NewBuildingMetadataValidationService()

	// Valid mixed building metadata
	validMetadata := models.BuildingMetadata{
		"residential_section": map[string]interface{}{
			"floors":        "3-10",
			"amenities":     []interface{}{"gym", "rooftop_garden"},
			"security_type": "card_access",
		},
		"commercial_section": map[string]interface{}{
			"floors":             "1-2",
			"business_hours":     "09:00-22:00",
			"parking_allocation": 60,
		},
		"shared_facilities": map[string]interface{}{
			"elevators":        3,
			"parking_total":    120,
			"backup_generator": true,
		},
	}

	req := &models.CreateBuildingRequest{
		PropertyID:       1,
		BuildingName:     "Mixed Use Complex",
		BuildingCode:     "MUC-001",
		BuildingType:     models.BuildingTypeMixed,
		TotalFloors:      10,
		HasElevator:      true,
		ConstructionYear: &[]int{2021}[0],
		Metadata:         validMetadata,
	}

	if err := service.ValidateBuildingCreation(req); err != nil {
		fmt.Printf("Validation failed: %s\n", err.Error())
	} else {
		fmt.Println("Mixed building validation passed!")
	}
}

// ExampleInvalidMetadataValidation demonstrates validation error handling
func ExampleInvalidMetadataValidation() {
	service := NewBuildingMetadataValidationService()

	// Invalid residential building metadata
	invalidMetadata := models.BuildingMetadata{
		"amenities":               []interface{}{"gym", "invalid_amenity"}, // Invalid amenity
		"security_type":           "invalid_security",                      // Invalid security type
		"maintenance_staff_count": -1,                                      // Negative staff count
	}

	req := &models.CreateBuildingRequest{
		PropertyID:   1,
		BuildingName: "Invalid Building",
		BuildingCode: "INV-001",
		BuildingType: models.BuildingTypeResidential,
		TotalFloors:  5,
		Metadata:     invalidMetadata,
	}

	if err := service.ValidateBuildingCreation(req); err != nil {
		fmt.Printf("Expected validation error: %s\n", err.Error())

		// Get formatted error response for API
		errorResponse := service.errorHandler.GetValidationErrorResponse(err)
		fmt.Printf("API Error Response: %+v\n", errorResponse)
	}
}

// ExampleGetMetadataSchema demonstrates schema retrieval
func ExampleGetMetadataSchema() {
	service := NewBuildingMetadataValidationService()

	// Get schema for residential buildings
	schema, err := service.GetMetadataSchemaForType(models.BuildingTypeResidential)
	if err != nil {
		fmt.Printf("Error getting schema: %s\n", err.Error())
		return
	}

	fmt.Printf("Residential building metadata schema: %+v\n", schema)
}
