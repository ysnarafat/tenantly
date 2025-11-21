package services

import (
	"testing"

	"github.com/ysnarafat/tenantly/internal/models"
)

func TestBuildingMetadataValidator_ValidateResidentialMetadata(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name        string
		metadata    models.BuildingMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty metadata should be valid",
			metadata:    models.BuildingMetadata{},
			expectError: false,
		},
		{
			name: "Valid residential metadata",
			metadata: models.BuildingMetadata{
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
			},
			expectError: false,
		},
		{
			name: "Invalid amenity",
			metadata: models.BuildingMetadata{
				"amenities": []interface{}{"gym", "invalid_amenity"},
			},
			expectError: true,
			errorMsg:    "invalid amenity",
		},
		{
			name: "Invalid security type",
			metadata: models.BuildingMetadata{
				"security_type": "invalid_security",
			},
			expectError: true,
			errorMsg:    "invalid security_type",
		},
		{
			name: "Negative maintenance staff count",
			metadata: models.BuildingMetadata{
				"maintenance_staff_count": -1,
			},
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name: "Invalid parking spaces format",
			metadata: models.BuildingMetadata{
				"parking_spaces": "invalid",
			},
			expectError: true,
			errorMsg:    "must be an object",
		},
		{
			name: "Invalid water supply type",
			metadata: models.BuildingMetadata{
				"utilities": map[string]interface{}{
					"water_supply": "invalid_supply",
				},
			},
			expectError: true,
			errorMsg:    "invalid water_supply",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateResidentialMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_ValidateCommercialMetadata(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name        string
		metadata    models.BuildingMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty metadata should be valid",
			metadata:    models.BuildingMetadata{},
			expectError: false,
		},
		{
			name: "Valid commercial metadata",
			metadata: models.BuildingMetadata{
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
			},
			expectError: false,
		},
		{
			name: "Invalid business hours format",
			metadata: models.BuildingMetadata{
				"business_hours": map[string]interface{}{
					"weekdays": "25:00-30:00",
				},
			},
			expectError: true,
			errorMsg:    "invalid business_hours.weekdays format",
		},
		{
			name: "Invalid security system type",
			metadata: models.BuildingMetadata{
				"security_system": map[string]interface{}{
					"type": "invalid_security",
				},
			},
			expectError: true,
			errorMsg:    "invalid security_system.type",
		},
		{
			name: "Negative loading docks",
			metadata: models.BuildingMetadata{
				"loading_docks": -1,
			},
			expectError: true,
			errorMsg:    "cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCommercialMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_ValidateMixedMetadata(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name        string
		metadata    models.BuildingMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty metadata should be valid",
			metadata:    models.BuildingMetadata{},
			expectError: false,
		},
		{
			name: "Valid mixed metadata",
			metadata: models.BuildingMetadata{
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
			},
			expectError: false,
		},
		{
			name: "Invalid floor range format",
			metadata: models.BuildingMetadata{
				"residential_section": map[string]interface{}{
					"floors": "invalid-range",
				},
			},
			expectError: true,
			errorMsg:    "invalid floor range format",
		},
		{
			name: "Invalid floor range order",
			metadata: models.BuildingMetadata{
				"commercial_section": map[string]interface{}{
					"floors": "5-3",
				},
			},
			expectError: true,
			errorMsg:    "start floor must be less than end floor",
		},
		{
			name: "Invalid business hours in commercial section",
			metadata: models.BuildingMetadata{
				"commercial_section": map[string]interface{}{
					"business_hours": "25:00-30:00",
				},
			},
			expectError: true,
			errorMsg:    "invalid commercial_section.business_hours format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMixedMetadata(tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_ValidateMetadata(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name         string
		buildingType models.BuildingType
		metadata     models.BuildingMetadata
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "Valid residential building",
			buildingType: models.BuildingTypeResidential,
			metadata: models.BuildingMetadata{
				"amenities":     []interface{}{"gym", "swimming_pool"},
				"security_type": "24_hour_guard",
			},
			expectError: false,
		},
		{
			name:         "Valid commercial building",
			buildingType: models.BuildingTypeCommercial,
			metadata: models.BuildingMetadata{
				"loading_docks": 2,
				"business_hours": map[string]interface{}{
					"weekdays": "09:00-18:00",
				},
			},
			expectError: false,
		},
		{
			name:         "Valid mixed building",
			buildingType: models.BuildingTypeMixed,
			metadata: models.BuildingMetadata{
				"residential_section": map[string]interface{}{
					"floors": "3-10",
				},
				"commercial_section": map[string]interface{}{
					"floors": "1-2",
				},
			},
			expectError: false,
		},
		{
			name:         "Invalid building type",
			buildingType: "InvalidType",
			metadata:     models.BuildingMetadata{},
			expectError:  true,
			errorMsg:     "invalid building type",
		},
		{
			name:         "Nil metadata should be valid",
			buildingType: models.BuildingTypeResidential,
			metadata:     nil,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMetadata(tt.buildingType, tt.metadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_GetMetadataSchema(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name         string
		buildingType models.BuildingType
		expectNil    bool
	}{
		{
			name:         "Residential schema",
			buildingType: models.BuildingTypeResidential,
			expectNil:    false,
		},
		{
			name:         "Commercial schema",
			buildingType: models.BuildingTypeCommercial,
			expectNil:    false,
		},
		{
			name:         "Mixed schema",
			buildingType: models.BuildingTypeMixed,
			expectNil:    false,
		},
		{
			name:         "Invalid building type",
			buildingType: "InvalidType",
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := validator.GetMetadataSchema(tt.buildingType)

			if tt.expectNil {
				if schema != nil {
					t.Errorf("Expected nil schema but got: %v", schema)
				}
			} else {
				if schema == nil {
					t.Errorf("Expected schema but got nil")
				} else {
					// Verify schema has required structure
					if schema["type"] != "object" {
						t.Errorf("Expected schema type to be 'object', got: %v", schema["type"])
					}
					if _, exists := schema["properties"]; !exists {
						t.Errorf("Expected schema to have 'properties' field")
					}
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_ValidatePositiveInteger(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name        string
		value       interface{}
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid integer",
			value:       5,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Valid float64",
			value:       5.0,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Valid string integer",
			value:       "5",
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Zero value",
			value:       0,
			fieldName:   "test_field",
			expectError: false,
		},
		{
			name:        "Negative integer",
			value:       -1,
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name:        "Invalid string",
			value:       "invalid",
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "must be a valid integer",
		},
		{
			name:        "Invalid type",
			value:       []string{"invalid"},
			fieldName:   "test_field",
			expectError: true,
			errorMsg:    "must be an integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePositiveInteger(tt.value, tt.fieldName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestBuildingMetadataValidator_ValidateFloorRange(t *testing.T) {
	validator := NewBuildingMetadataValidator()

	tests := []struct {
		name        string
		floorRange  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid floor range",
			floorRange:  "1-5",
			expectError: false,
		},
		{
			name:        "Valid multi-digit floor range",
			floorRange:  "3-15",
			expectError: false,
		},
		{
			name:        "Invalid format - no dash",
			floorRange:  "15",
			expectError: true,
			errorMsg:    "invalid floor range format",
		},
		{
			name:        "Invalid format - multiple dashes",
			floorRange:  "1-5-10",
			expectError: true,
			errorMsg:    "invalid floor range format",
		},
		{
			name:        "Invalid range - start equals end",
			floorRange:  "5-5",
			expectError: true,
			errorMsg:    "start floor must be less than end floor",
		},
		{
			name:        "Invalid range - start greater than end",
			floorRange:  "10-5",
			expectError: true,
			errorMsg:    "start floor must be less than end floor",
		},
		{
			name:        "Invalid range - zero start",
			floorRange:  "0-5",
			expectError: true,
			errorMsg:    "floor numbers must be positive",
		},
		{
			name:        "Invalid format - letters",
			floorRange:  "a-b",
			expectError: true,
			errorMsg:    "invalid floor range format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFloorRange(tt.floorRange)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
