package services

import (
	"errors"
	"testing"

	"github.com/ysnarafat/tenantly/internal/models"
)

func TestValidationErrorHandler_HandleMetadataValidationError(t *testing.T) {
	handler := NewValidationErrorHandler()

	tests := []struct {
		name          string
		err           error
		buildingType  models.BuildingType
		expectedCode  string
		expectedField string
	}{
		{
			name:         "Nil error should return nil",
			err:          nil,
			buildingType: models.BuildingTypeResidential,
			expectedCode: "",
		},
		{
			name:          "Amenities validation error",
			err:           errors.New("amenities validation failed: invalid amenity"),
			buildingType:  models.BuildingTypeResidential,
			expectedCode:  models.ErrInvalidAmenities,
			expectedField: "amenities",
		},
		{
			name:          "Security type validation error",
			err:           errors.New("security_type validation failed: invalid type"),
			buildingType:  models.BuildingTypeResidential,
			expectedCode:  models.ErrInvalidSecurityType,
			expectedField: "security_type",
		},
		{
			name:          "Parking spaces validation error",
			err:           errors.New("parking_spaces validation failed: invalid format"),
			buildingType:  models.BuildingTypeCommercial,
			expectedCode:  models.ErrInvalidParkingSpaces,
			expectedField: "parking_spaces",
		},
		{
			name:          "Business hours validation error",
			err:           errors.New("business_hours validation failed: invalid format"),
			buildingType:  models.BuildingTypeCommercial,
			expectedCode:  models.ErrInvalidBusinessHours,
			expectedField: "business_hours",
		},
		{
			name:          "Facilities validation error",
			err:           errors.New("facilities validation failed: invalid data"),
			buildingType:  models.BuildingTypeCommercial,
			expectedCode:  models.ErrInvalidFacilities,
			expectedField: "facilities",
		},
		{
			name:          "Residential section validation error",
			err:           errors.New("residential_section validation failed: invalid floors"),
			buildingType:  models.BuildingTypeMixed,
			expectedCode:  models.ErrInvalidResidentialSection,
			expectedField: "residential_section",
		},
		{
			name:          "Commercial section validation error",
			err:           errors.New("commercial_section validation failed: invalid hours"),
			buildingType:  models.BuildingTypeMixed,
			expectedCode:  models.ErrInvalidCommercialSection,
			expectedField: "commercial_section",
		},
		{
			name:          "Shared facilities validation error",
			err:           errors.New("shared_facilities validation failed: invalid data"),
			buildingType:  models.BuildingTypeMixed,
			expectedCode:  models.ErrInvalidSharedFacilities,
			expectedField: "shared_facilities",
		},
		{
			name:          "Generic metadata validation error",
			err:           errors.New("some other validation error"),
			buildingType:  models.BuildingTypeResidential,
			expectedCode:  models.ErrMetadataValidationFailed,
			expectedField: "metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleMetadataValidationError(tt.err, tt.buildingType)

			if tt.expectedCode == "" {
				if result != nil {
					t.Errorf("Expected nil result but got: %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected BuildingError but got nil")
					return
				}

				if result.Code != tt.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, result.Code)
				}

				if result.Field != tt.expectedField {
					t.Errorf("Expected field '%s', got '%s'", tt.expectedField, result.Field)
				}

				if result.Message == "" {
					t.Errorf("Expected non-empty message")
				}
			}
		})
	}
}

func TestValidationErrorHandler_HandleBuildingValidationError(t *testing.T) {
	handler := NewValidationErrorHandler()

	tests := []struct {
		name          string
		err           error
		expectedCode  string
		expectedField string
	}{
		{
			name:         "Nil error should return nil",
			err:          nil,
			expectedCode: "",
		},
		{
			name:          "Building not found error",
			err:           errors.New("building not found"),
			expectedCode:  models.ErrBuildingNotFound,
			expectedField: "",
		},
		{
			name:          "Building code exists error",
			err:           errors.New("building code already exists"),
			expectedCode:  models.ErrBuildingCodeExists,
			expectedField: "building_code",
		},
		{
			name:          "Invalid building type error",
			err:           errors.New("invalid building type"),
			expectedCode:  models.ErrInvalidBuildingType,
			expectedField: "building_type",
		},
		{
			name:          "Property not found error",
			err:           errors.New("property not found"),
			expectedCode:  models.ErrPropertyNotFound,
			expectedField: "property_id",
		},
		{
			name:          "Building has active units error",
			err:           errors.New("cannot delete building with active units"),
			expectedCode:  models.ErrBuildingHasActiveUnits,
			expectedField: "",
		},
		{
			name:          "Invalid floor count error",
			err:           errors.New("invalid floor count"),
			expectedCode:  models.ErrInvalidFloorCount,
			expectedField: "total_floors",
		},
		{
			name:          "Invalid construction year error",
			err:           errors.New("invalid construction year"),
			expectedCode:  models.ErrInvalidConstructionYear,
			expectedField: "construction_year",
		},
		{
			name:          "Generic validation error",
			err:           errors.New("some other validation error"),
			expectedCode:  models.ErrInvalidMetadata,
			expectedField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleBuildingValidationError(tt.err)

			if tt.expectedCode == "" {
				if result != nil {
					t.Errorf("Expected nil result but got: %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected BuildingError but got nil")
					return
				}

				if result.Code != tt.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, result.Code)
				}

				if result.Field != tt.expectedField {
					t.Errorf("Expected field '%s', got '%s'", tt.expectedField, result.Field)
				}
			}
		})
	}
}

func TestValidationErrorHandler_HandleConstraintViolationError(t *testing.T) {
	handler := NewValidationErrorHandler()

	tests := []struct {
		name          string
		err           error
		expectedCode  string
		expectedField string
	}{
		{
			name:         "Nil error should return nil",
			err:          nil,
			expectedCode: "",
		},
		{
			name:          "Unique building code constraint violation",
			err:           errors.New("unique_building_code_per_property constraint violation"),
			expectedCode:  models.ErrBuildingCodeExists,
			expectedField: "building_code",
		},
		{
			name:          "Property foreign key constraint violation",
			err:           errors.New("foreign key constraint violation on property_id"),
			expectedCode:  models.ErrPropertyNotFound,
			expectedField: "property_id",
		},
		{
			name:          "Generic foreign key constraint violation",
			err:           errors.New("foreign key constraint violation"),
			expectedCode:  models.ErrInvalidMetadata,
			expectedField: "",
		},
		{
			name:          "Valid floors check constraint violation",
			err:           errors.New("check constraint violation: valid_floors"),
			expectedCode:  models.ErrInvalidFloorCount,
			expectedField: "total_floors",
		},
		{
			name:          "Valid construction year check constraint violation",
			err:           errors.New("check constraint violation: valid_construction_year"),
			expectedCode:  models.ErrInvalidConstructionYear,
			expectedField: "construction_year",
		},
		{
			name:          "Generic check constraint violation",
			err:           errors.New("check constraint violation"),
			expectedCode:  models.ErrInvalidMetadata,
			expectedField: "",
		},
		{
			name:          "Generic database constraint violation",
			err:           errors.New("some database constraint violation"),
			expectedCode:  models.ErrInvalidMetadata,
			expectedField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.HandleConstraintViolationError(tt.err)

			if tt.expectedCode == "" {
				if result != nil {
					t.Errorf("Expected nil result but got: %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected BuildingError but got nil")
					return
				}

				if result.Code != tt.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, result.Code)
				}

				if result.Field != tt.expectedField {
					t.Errorf("Expected field '%s', got '%s'", tt.expectedField, result.Field)
				}
			}
		})
	}
}

func TestValidationErrorHandler_GetValidationErrorResponse(t *testing.T) {
	handler := NewValidationErrorHandler()

	tests := []struct {
		name           string
		buildingError  *models.BuildingError
		expectedFields []string
	}{
		{
			name: "Error with field",
			buildingError: models.NewBuildingError(
				models.ErrInvalidAmenities,
				"Invalid amenity provided",
				"amenities",
			),
			expectedFields: []string{"error", "error.code", "error.message", "error.field"},
		},
		{
			name: "Error without field",
			buildingError: models.NewBuildingError(
				models.ErrBuildingNotFound,
				"Building not found",
				"",
			),
			expectedFields: []string{"error", "error.code", "error.message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := handler.GetValidationErrorResponse(tt.buildingError)

			// Check that response has error object
			errorObj, exists := response["error"]
			if !exists {
				t.Errorf("Expected 'error' field in response")
				return
			}

			errorMap, ok := errorObj.(map[string]interface{})
			if !ok {
				t.Errorf("Expected 'error' to be a map")
				return
			}

			// Check required fields
			if errorMap["code"] != tt.buildingError.Code {
				t.Errorf("Expected code '%s', got '%s'", tt.buildingError.Code, errorMap["code"])
			}

			if errorMap["message"] != tt.buildingError.Message {
				t.Errorf("Expected message '%s', got '%s'", tt.buildingError.Message, errorMap["message"])
			}

			// Check field presence based on whether it should exist
			if tt.buildingError.Field != "" {
				if errorMap["field"] != tt.buildingError.Field {
					t.Errorf("Expected field '%s', got '%s'", tt.buildingError.Field, errorMap["field"])
				}
			} else {
				if _, exists := errorMap["field"]; exists {
					t.Errorf("Expected no 'field' in response but found one")
				}
			}
		})
	}
}

func TestValidationErrorHandler_GetMultipleValidationErrorsResponse(t *testing.T) {
	handler := NewValidationErrorHandler()

	errors := []*models.BuildingError{
		models.NewBuildingError(models.ErrInvalidAmenities, "Invalid amenity", "amenities"),
		models.NewBuildingError(models.ErrInvalidSecurityType, "Invalid security type", "security_type"),
		models.NewBuildingError(models.ErrBuildingNotFound, "Building not found", ""),
	}

	response := handler.GetMultipleValidationErrorsResponse(errors)

	// Check that response has errors array
	errorsArray, exists := response["errors"]
	if !exists {
		t.Errorf("Expected 'errors' field in response")
		return
	}

	errorsSlice, ok := errorsArray.([]map[string]interface{})
	if !ok {
		t.Errorf("Expected 'errors' to be a slice of maps")
		return
	}

	if len(errorsSlice) != len(errors) {
		t.Errorf("Expected %d errors, got %d", len(errors), len(errorsSlice))
		return
	}

	// Check each error in the response
	for i, expectedError := range errors {
		errorMap := errorsSlice[i]

		if errorMap["code"] != expectedError.Code {
			t.Errorf("Error %d: Expected code '%s', got '%s'", i, expectedError.Code, errorMap["code"])
		}

		if errorMap["message"] != expectedError.Message {
			t.Errorf("Error %d: Expected message '%s', got '%s'", i, expectedError.Message, errorMap["message"])
		}

		if expectedError.Field != "" {
			if errorMap["field"] != expectedError.Field {
				t.Errorf("Error %d: Expected field '%s', got '%s'", i, expectedError.Field, errorMap["field"])
			}
		} else {
			if _, exists := errorMap["field"]; exists {
				t.Errorf("Error %d: Expected no 'field' but found one", i)
			}
		}
	}
}

func TestValidationErrorHandler_ValidateAndFormatMetadataErrors(t *testing.T) {
	handler := NewValidationErrorHandler()
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name         string
		buildingType models.BuildingType
		metadata     models.BuildingMetadata
		expectError  bool
		expectedCode string
	}{
		{
			name:         "Valid metadata should return nil",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities": []interface{}{"gym", "swimming_pool"},
			},
			expectError: false,
		},
		{
			name:         "Invalid metadata should return formatted error",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities": []interface{}{"invalid_amenity"},
			},
			expectError:  true,
			expectedCode: models.ErrInvalidAmenities,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.ValidateAndFormatMetadataErrors(validator, tt.buildingType, tt.metadata)

			if tt.expectError {
				if result == nil {
					t.Errorf("Expected BuildingError but got nil")
				} else if result.Code != tt.expectedCode {
					t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, result.Code)
				}
			} else {
				if result != nil {
					t.Errorf("Expected nil but got: %v", result)
				}
			}
		})
	}
}
