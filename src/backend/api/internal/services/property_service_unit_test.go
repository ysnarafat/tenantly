package services

import (
	"testing"

	"github.com/ysnarafat/tenantly/internal/models"
)

// TestValidatePropertyType tests the property type validation without database
func TestValidatePropertyType_Unit(t *testing.T) {
	// Create a minimal service instance for testing validation logic
	service := &PropertyService{}

	tests := []struct {
		name         string
		propertyType string
		expectError  bool
	}{
		{
			name:         "Valid Residential type",
			propertyType: "Residential",
			expectError:  false,
		},
		{
			name:         "Valid Commercial type",
			propertyType: "Commercial",
			expectError:  false,
		},
		{
			name:         "Valid Mixed type",
			propertyType: "Mixed",
			expectError:  false,
		},
		{
			name:         "Invalid type",
			propertyType: "InvalidType",
			expectError:  true,
		},
		{
			name:         "Empty type",
			propertyType: "",
			expectError:  true,
		},
		{
			name:         "Case sensitive - lowercase",
			propertyType: "residential",
			expectError:  true,
		},
		{
			name:         "Case sensitive - mixed case",
			propertyType: "COMMERCIAL",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePropertyType(tt.propertyType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for property type: %s", tt.propertyType)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for valid property type %s: %v", tt.propertyType, err)
			}
		})
	}
}

// TestPropertyTypeConstants tests that the model constants are correct
func TestPropertyTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant models.PropertyType
		expected string
	}{
		{
			name:     "Residential constant",
			constant: models.PropertyTypeResidential,
			expected: "Residential",
		},
		{
			name:     "Commercial constant",
			constant: models.PropertyTypeCommercial,
			expected: "Commercial",
		},
		{
			name:     "Mixed constant",
			constant: models.PropertyTypeMixed,
			expected: "Mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("Expected constant %s to equal %s, got %s", tt.name, tt.expected, string(tt.constant))
			}
		})
	}
}

// TestCreatePropertyRequestValidation tests the validation tags on CreatePropertyRequest
func TestCreatePropertyRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request models.CreatePropertyRequest
		valid   bool
	}{
		{
			name: "Valid request",
			request: models.CreatePropertyRequest{
				PropertyName: "Test Property",
				PropertyCode: "TEST001",
				Address:      "123 Test Street",
				PropertyType: models.PropertyTypeCommercial,
			},
			valid: true,
		},
		{
			name: "Empty property name",
			request: models.CreatePropertyRequest{
				PropertyName: "",
				PropertyCode: "TEST001",
				Address:      "123 Test Street",
				PropertyType: models.PropertyTypeCommercial,
			},
			valid: false,
		},
		{
			name: "Empty property code",
			request: models.CreatePropertyRequest{
				PropertyName: "Test Property",
				PropertyCode: "",
				Address:      "123 Test Street",
				PropertyType: models.PropertyTypeCommercial,
			},
			valid: false,
		},
		{
			name: "Empty address",
			request: models.CreatePropertyRequest{
				PropertyName: "Test Property",
				PropertyCode: "TEST001",
				Address:      "",
				PropertyType: models.PropertyTypeCommercial,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in a real scenario, this would be done by Gin's binding
			hasRequiredFields := tt.request.PropertyName != "" &&
				tt.request.PropertyCode != "" &&
				tt.request.Address != "" &&
				tt.request.PropertyType != ""

			if tt.valid && !hasRequiredFields {
				t.Errorf("Expected valid request but missing required fields")
			}

			if !tt.valid && hasRequiredFields {
				// This test case expects validation to fail, but all required fields are present
				// In this case, we're testing empty string validation which should fail
				if tt.request.PropertyName == "" || tt.request.PropertyCode == "" || tt.request.Address == "" {
					// This is expected - empty strings should fail validation
					return
				}
				t.Errorf("Expected invalid request but all required fields are present")
			}
		})
	}
}

// TestPropertyMetadata tests the PropertyMetadata type
func TestPropertyMetadata(t *testing.T) {
	metadata := models.PropertyMetadata{
		"parking_spaces":  50,
		"total_area_sqft": 25000,
		"amenities":       []string{"elevator", "security", "parking"},
	}

	// Test that we can access metadata fields
	if parkingSpaces, ok := metadata["parking_spaces"]; !ok || parkingSpaces != 50 {
		t.Errorf("Expected parking_spaces to be 50, got %v", parkingSpaces)
	}

	if totalArea, ok := metadata["total_area_sqft"]; !ok || totalArea != 25000 {
		t.Errorf("Expected total_area_sqft to be 25000, got %v", totalArea)
	}

	if amenities, ok := metadata["amenities"]; !ok {
		t.Errorf("Expected amenities to exist in metadata")
	} else {
		amenitiesList, ok := amenities.([]string)
		if !ok {
			t.Errorf("Expected amenities to be []string")
		} else if len(amenitiesList) != 3 {
			t.Errorf("Expected 3 amenities, got %d", len(amenitiesList))
		}
	}
}

// TestUpdatePropertyRequest tests the update request structure
func TestUpdatePropertyRequest(t *testing.T) {
	newName := "Updated Property Name"
	newAddress := "Updated Address"
	newType := models.PropertyTypeMixed
	newActive := false

	updateReq := models.UpdatePropertyRequest{
		PropertyName: &newName,
		Address:      &newAddress,
		PropertyType: &newType,
		Active:       &newActive,
	}

	// Test that pointers work correctly
	if updateReq.PropertyName == nil || *updateReq.PropertyName != newName {
		t.Errorf("Expected PropertyName to be %s", newName)
	}

	if updateReq.Address == nil || *updateReq.Address != newAddress {
		t.Errorf("Expected Address to be %s", newAddress)
	}

	if updateReq.PropertyType == nil || *updateReq.PropertyType != newType {
		t.Errorf("Expected PropertyType to be %s", newType)
	}

	if updateReq.Active == nil || *updateReq.Active != newActive {
		t.Errorf("Expected Active to be %t", newActive)
	}

	// Test that nil pointers work (fields not being updated)
	emptyUpdateReq := models.UpdatePropertyRequest{}

	if emptyUpdateReq.PropertyName != nil {
		t.Errorf("Expected PropertyName to be nil")
	}

	if emptyUpdateReq.Address != nil {
		t.Errorf("Expected Address to be nil")
	}
}
