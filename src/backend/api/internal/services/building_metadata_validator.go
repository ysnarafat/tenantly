package services

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ysnarafat/tenantly/internal/models"
)

// MetadataValidatorInterface defines the interface for building metadata validation
type MetadataValidatorInterface interface {
	ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error
	ValidateResidentialMetadata(metadata models.BuildingMetadata) error
	ValidateCommercialMetadata(metadata models.BuildingMetadata) error
	ValidateMixedMetadata(metadata models.BuildingMetadata) error
	GetMetadataSchema(buildingType models.BuildingType) map[string]interface{}
}

// BuildingMetadataValidator implements metadata validation for different building types
type BuildingMetadataValidator struct{}

// NewBuildingMetadataValidator creates a new instance of BuildingMetadataValidator
func NewBuildingMetadataValidator() *BuildingMetadataValidator {
	return &BuildingMetadataValidator{}
}

// ValidateMetadata validates metadata based on building type
func (v *BuildingMetadataValidator) ValidateMetadata(buildingType models.BuildingType, metadata models.BuildingMetadata) error {
	if metadata == nil {
		return nil // Empty metadata is allowed
	}

	switch buildingType {
	case models.BuildingTypeResidential:
		return v.ValidateResidentialMetadata(metadata)
	case models.BuildingTypeCommercial:
		return v.ValidateCommercialMetadata(metadata)
	case models.BuildingTypeMixed:
		return v.ValidateMixedMetadata(metadata)
	default:
		return fmt.Errorf("invalid building type: %s", buildingType)
	}
}

// ValidateResidentialMetadata validates metadata for residential buildings
func (v *BuildingMetadataValidator) ValidateResidentialMetadata(metadata models.BuildingMetadata) error {
	// Validate amenities
	if amenities, exists := metadata["amenities"]; exists {
		if err := v.validateAmenities(amenities); err != nil {
			return fmt.Errorf("amenities validation failed: %w", err)
		}
	}

	// Validate security_type
	if securityType, exists := metadata["security_type"]; exists {
		if err := v.validateSecurityType(securityType); err != nil {
			return fmt.Errorf("security_type validation failed: %w", err)
		}
	}

	// Validate maintenance_staff_count
	if staffCount, exists := metadata["maintenance_staff_count"]; exists {
		if err := v.validateMaintenanceStaffCount(staffCount); err != nil {
			return fmt.Errorf("maintenance_staff_count validation failed: %w", err)
		}
	}

	// Validate parking_spaces
	if parkingSpaces, exists := metadata["parking_spaces"]; exists {
		if err := v.validateParkingSpaces(parkingSpaces); err != nil {
			return fmt.Errorf("parking_spaces validation failed: %w", err)
		}
	}

	// Validate utilities
	if utilities, exists := metadata["utilities"]; exists {
		if err := v.validateUtilities(utilities); err != nil {
			return fmt.Errorf("utilities validation failed: %w", err)
		}
	}

	return nil
}

// ValidateCommercialMetadata validates metadata for commercial buildings
func (v *BuildingMetadataValidator) ValidateCommercialMetadata(metadata models.BuildingMetadata) error {
	// Validate parking_spaces (commercial format)
	if parkingSpaces, exists := metadata["parking_spaces"]; exists {
		if err := v.validateCommercialParkingSpaces(parkingSpaces); err != nil {
			return fmt.Errorf("parking_spaces validation failed: %w", err)
		}
	}

	// Validate loading_docks
	if loadingDocks, exists := metadata["loading_docks"]; exists {
		if err := v.validateLoadingDocks(loadingDocks); err != nil {
			return fmt.Errorf("loading_docks validation failed: %w", err)
		}
	}

	// Validate security_system
	if securitySystem, exists := metadata["security_system"]; exists {
		if err := v.validateSecuritySystem(securitySystem); err != nil {
			return fmt.Errorf("security_system validation failed: %w", err)
		}
	}

	// Validate business_hours
	if businessHours, exists := metadata["business_hours"]; exists {
		if err := v.validateBusinessHours(businessHours); err != nil {
			return fmt.Errorf("business_hours validation failed: %w", err)
		}
	}

	// Validate facilities
	if facilities, exists := metadata["facilities"]; exists {
		if err := v.validateFacilities(facilities); err != nil {
			return fmt.Errorf("facilities validation failed: %w", err)
		}
	}

	return nil
}

// ValidateMixedMetadata validates metadata for mixed-use buildings
func (v *BuildingMetadataValidator) ValidateMixedMetadata(metadata models.BuildingMetadata) error {
	// Validate residential_section
	if residentialSection, exists := metadata["residential_section"]; exists {
		if err := v.validateResidentialSection(residentialSection); err != nil {
			return fmt.Errorf("residential_section validation failed: %w", err)
		}
	}

	// Validate commercial_section
	if commercialSection, exists := metadata["commercial_section"]; exists {
		if err := v.validateCommercialSection(commercialSection); err != nil {
			return fmt.Errorf("commercial_section validation failed: %w", err)
		}
	}

	// Validate shared_facilities
	if sharedFacilities, exists := metadata["shared_facilities"]; exists {
		if err := v.validateSharedFacilities(sharedFacilities); err != nil {
			return fmt.Errorf("shared_facilities validation failed: %w", err)
		}
	}

	return nil
}

// validateAmenities validates residential building amenities
func (v *BuildingMetadataValidator) validateAmenities(amenities interface{}) error {
	amenitiesList, ok := amenities.([]interface{})
	if !ok {
		return fmt.Errorf("amenities must be an array")
	}

	validAmenities := map[string]bool{
		"gym":            true,
		"swimming_pool":  true,
		"playground":     true,
		"community_hall": true,
		"rooftop_garden": true,
		"library":        true,
		"prayer_room":    true,
		"parking":        true,
		"security":       true,
		"elevator":       true,
		"generator":      true,
		"water_tank":     true,
	}

	for _, amenity := range amenitiesList {
		amenityStr, ok := amenity.(string)
		if !ok {
			return fmt.Errorf("each amenity must be a string")
		}
		if !validAmenities[amenityStr] {
			return fmt.Errorf("invalid amenity: %s", amenityStr)
		}
	}

	return nil
}

// validateSecurityType validates security type for residential buildings
func (v *BuildingMetadataValidator) validateSecurityType(securityType interface{}) error {
	securityTypeStr, ok := securityType.(string)
	if !ok {
		return fmt.Errorf("security_type must be a string")
	}

	validSecurityTypes := map[string]bool{
		"24_hour_guard": true,
		"cctv_only":     true,
		"card_access":   true,
		"basic":         true,
	}

	if !validSecurityTypes[securityTypeStr] {
		return fmt.Errorf("invalid security_type: %s", securityTypeStr)
	}

	return nil
}

// validateMaintenanceStaffCount validates maintenance staff count
func (v *BuildingMetadataValidator) validateMaintenanceStaffCount(staffCount interface{}) error {
	var count int
	switch v := staffCount.(type) {
	case float64:
		count = int(v)
	case int:
		count = v
	case string:
		var err error
		count, err = strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("maintenance_staff_count must be a valid integer")
		}
	default:
		return fmt.Errorf("maintenance_staff_count must be an integer")
	}

	if count < 0 {
		return fmt.Errorf("maintenance_staff_count cannot be negative")
	}

	if count > 100 {
		return fmt.Errorf("maintenance_staff_count cannot exceed 100")
	}

	return nil
}

// validateParkingSpaces validates parking spaces for residential buildings
func (v *BuildingMetadataValidator) validateParkingSpaces(parkingSpaces interface{}) error {
	parkingMap, ok := parkingSpaces.(map[string]interface{})
	if !ok {
		return fmt.Errorf("parking_spaces must be an object")
	}

	// Validate total parking spaces
	if total, exists := parkingMap["total"]; exists {
		if err := v.validatePositiveInteger(total, "total"); err != nil {
			return err
		}
	}

	// Validate covered parking spaces
	if covered, exists := parkingMap["covered"]; exists {
		if err := v.validatePositiveInteger(covered, "covered"); err != nil {
			return err
		}
	}

	// Validate visitor parking spaces
	if visitor, exists := parkingMap["visitor"]; exists {
		if err := v.validatePositiveInteger(visitor, "visitor"); err != nil {
			return err
		}
	}

	return nil
}

// validateCommercialParkingSpaces validates parking spaces for commercial buildings
func (v *BuildingMetadataValidator) validateCommercialParkingSpaces(parkingSpaces interface{}) error {
	parkingMap, ok := parkingSpaces.(map[string]interface{})
	if !ok {
		return fmt.Errorf("parking_spaces must be an object")
	}

	// Validate total parking spaces
	if total, exists := parkingMap["total"]; exists {
		if err := v.validatePositiveInteger(total, "total"); err != nil {
			return err
		}
	}

	// Validate customer parking spaces
	if customer, exists := parkingMap["customer"]; exists {
		if err := v.validatePositiveInteger(customer, "customer"); err != nil {
			return err
		}
	}

	// Validate staff parking spaces
	if staff, exists := parkingMap["staff"]; exists {
		if err := v.validatePositiveInteger(staff, "staff"); err != nil {
			return err
		}
	}

	return nil
}

// validateUtilities validates utilities for residential buildings
func (v *BuildingMetadataValidator) validateUtilities(utilities interface{}) error {
	utilitiesMap, ok := utilities.(map[string]interface{})
	if !ok {
		return fmt.Errorf("utilities must be an object")
	}

	// Validate backup_generator
	if generator, exists := utilitiesMap["backup_generator"]; exists {
		if _, ok := generator.(bool); !ok {
			return fmt.Errorf("backup_generator must be a boolean")
		}
	}

	// Validate water_supply
	if waterSupply, exists := utilitiesMap["water_supply"]; exists {
		waterSupplyStr, ok := waterSupply.(string)
		if !ok {
			return fmt.Errorf("water_supply must be a string")
		}
		validWaterSupply := map[string]bool{
			"24_hour":   true,
			"scheduled": true,
			"limited":   true,
		}
		if !validWaterSupply[waterSupplyStr] {
			return fmt.Errorf("invalid water_supply: %s", waterSupplyStr)
		}
	}

	// Validate internet_ready
	if internetReady, exists := utilitiesMap["internet_ready"]; exists {
		if _, ok := internetReady.(bool); !ok {
			return fmt.Errorf("internet_ready must be a boolean")
		}
	}

	return nil
}

// validateLoadingDocks validates loading docks for commercial buildings
func (v *BuildingMetadataValidator) validateLoadingDocks(loadingDocks interface{}) error {
	return v.validatePositiveInteger(loadingDocks, "loading_docks")
}

// validateSecuritySystem validates security system for commercial buildings
func (v *BuildingMetadataValidator) validateSecuritySystem(securitySystem interface{}) error {
	securityMap, ok := securitySystem.(map[string]interface{})
	if !ok {
		return fmt.Errorf("security_system must be an object")
	}

	// Validate type
	if securityType, exists := securityMap["type"]; exists {
		securityTypeStr, ok := securityType.(string)
		if !ok {
			return fmt.Errorf("security_system.type must be a string")
		}
		validTypes := map[string]bool{
			"basic_cctv":    true,
			"advanced_cctv": true,
			"full_security": true,
		}
		if !validTypes[securityTypeStr] {
			return fmt.Errorf("invalid security_system.type: %s", securityTypeStr)
		}
	}

	// Validate access_control
	if accessControl, exists := securityMap["access_control"]; exists {
		if _, ok := accessControl.(bool); !ok {
			return fmt.Errorf("security_system.access_control must be a boolean")
		}
	}

	// Validate fire_safety
	if fireSafety, exists := securityMap["fire_safety"]; exists {
		fireSafetyStr, ok := fireSafety.(string)
		if !ok {
			return fmt.Errorf("security_system.fire_safety must be a string")
		}
		validFireSafety := map[string]bool{
			"basic":            true,
			"sprinkler_system": true,
			"full_system":      true,
		}
		if !validFireSafety[fireSafetyStr] {
			return fmt.Errorf("invalid security_system.fire_safety: %s", fireSafetyStr)
		}
	}

	return nil
}

// validateBusinessHours validates business hours for commercial buildings
func (v *BuildingMetadataValidator) validateBusinessHours(businessHours interface{}) error {
	hoursMap, ok := businessHours.(map[string]interface{})
	if !ok {
		return fmt.Errorf("business_hours must be an object")
	}

	timePattern := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

	// Validate weekdays
	if weekdays, exists := hoursMap["weekdays"]; exists {
		weekdaysStr, ok := weekdays.(string)
		if !ok {
			return fmt.Errorf("business_hours.weekdays must be a string")
		}
		if !timePattern.MatchString(weekdaysStr) {
			return fmt.Errorf("invalid business_hours.weekdays format, expected HH:MM-HH:MM")
		}
	}

	// Validate weekends
	if weekends, exists := hoursMap["weekends"]; exists {
		weekendsStr, ok := weekends.(string)
		if !ok {
			return fmt.Errorf("business_hours.weekends must be a string")
		}
		if !timePattern.MatchString(weekendsStr) {
			return fmt.Errorf("invalid business_hours.weekends format, expected HH:MM-HH:MM")
		}
	}

	// Validate holidays
	if holidays, exists := hoursMap["holidays"]; exists {
		holidaysStr, ok := holidays.(string)
		if !ok {
			return fmt.Errorf("business_hours.holidays must be a string")
		}
		if !timePattern.MatchString(holidaysStr) {
			return fmt.Errorf("invalid business_hours.holidays format, expected HH:MM-HH:MM")
		}
	}

	return nil
}

// validateFacilities validates facilities for commercial buildings
func (v *BuildingMetadataValidator) validateFacilities(facilities interface{}) error {
	facilitiesMap, ok := facilities.(map[string]interface{})
	if !ok {
		return fmt.Errorf("facilities must be an object")
	}

	// Validate elevators
	if elevators, exists := facilitiesMap["elevators"]; exists {
		if err := v.validatePositiveInteger(elevators, "elevators"); err != nil {
			return err
		}
	}

	// Validate escalators
	if escalators, exists := facilitiesMap["escalators"]; exists {
		if err := v.validatePositiveInteger(escalators, "escalators"); err != nil {
			return err
		}
	}

	// Validate food_court
	if foodCourt, exists := facilitiesMap["food_court"]; exists {
		if _, ok := foodCourt.(bool); !ok {
			return fmt.Errorf("facilities.food_court must be a boolean")
		}
	}

	// Validate atm
	if atm, exists := facilitiesMap["atm"]; exists {
		if _, ok := atm.(bool); !ok {
			return fmt.Errorf("facilities.atm must be a boolean")
		}
	}

	return nil
}

// validateResidentialSection validates residential section for mixed buildings
func (v *BuildingMetadataValidator) validateResidentialSection(residentialSection interface{}) error {
	sectionMap, ok := residentialSection.(map[string]interface{})
	if !ok {
		return fmt.Errorf("residential_section must be an object")
	}

	// Validate floors
	if floors, exists := sectionMap["floors"]; exists {
		floorsStr, ok := floors.(string)
		if !ok {
			return fmt.Errorf("residential_section.floors must be a string")
		}
		if err := v.validateFloorRange(floorsStr); err != nil {
			return fmt.Errorf("residential_section.floors validation failed: %w", err)
		}
	}

	// Validate amenities
	if amenities, exists := sectionMap["amenities"]; exists {
		if err := v.validateAmenities(amenities); err != nil {
			return fmt.Errorf("residential_section.amenities validation failed: %w", err)
		}
	}

	// Validate security_type
	if securityType, exists := sectionMap["security_type"]; exists {
		if err := v.validateSecurityType(securityType); err != nil {
			return fmt.Errorf("residential_section.security_type validation failed: %w", err)
		}
	}

	return nil
}

// validateCommercialSection validates commercial section for mixed buildings
func (v *BuildingMetadataValidator) validateCommercialSection(commercialSection interface{}) error {
	sectionMap, ok := commercialSection.(map[string]interface{})
	if !ok {
		return fmt.Errorf("commercial_section must be an object")
	}

	// Validate floors
	if floors, exists := sectionMap["floors"]; exists {
		floorsStr, ok := floors.(string)
		if !ok {
			return fmt.Errorf("commercial_section.floors must be a string")
		}
		if err := v.validateFloorRange(floorsStr); err != nil {
			return fmt.Errorf("commercial_section.floors validation failed: %w", err)
		}
	}

	// Validate business_hours
	if businessHours, exists := sectionMap["business_hours"]; exists {
		businessHoursStr, ok := businessHours.(string)
		if !ok {
			return fmt.Errorf("commercial_section.business_hours must be a string")
		}
		timePattern := regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$`)
		if !timePattern.MatchString(businessHoursStr) {
			return fmt.Errorf("invalid commercial_section.business_hours format, expected HH:MM-HH:MM")
		}
	}

	// Validate parking_allocation
	if parkingAllocation, exists := sectionMap["parking_allocation"]; exists {
		if err := v.validatePositiveInteger(parkingAllocation, "parking_allocation"); err != nil {
			return err
		}
	}

	return nil
}

// validateSharedFacilities validates shared facilities for mixed buildings
func (v *BuildingMetadataValidator) validateSharedFacilities(sharedFacilities interface{}) error {
	facilitiesMap, ok := sharedFacilities.(map[string]interface{})
	if !ok {
		return fmt.Errorf("shared_facilities must be an object")
	}

	// Validate elevators
	if elevators, exists := facilitiesMap["elevators"]; exists {
		if err := v.validatePositiveInteger(elevators, "elevators"); err != nil {
			return err
		}
	}

	// Validate parking_total
	if parkingTotal, exists := facilitiesMap["parking_total"]; exists {
		if err := v.validatePositiveInteger(parkingTotal, "parking_total"); err != nil {
			return err
		}
	}

	// Validate backup_generator
	if generator, exists := facilitiesMap["backup_generator"]; exists {
		if _, ok := generator.(bool); !ok {
			return fmt.Errorf("shared_facilities.backup_generator must be a boolean")
		}
	}

	return nil
}

// validatePositiveInteger validates that a value is a positive integer
func (v *BuildingMetadataValidator) validatePositiveInteger(value interface{}, fieldName string) error {
	var intValue int
	switch v := value.(type) {
	case float64:
		intValue = int(v)
	case int:
		intValue = v
	case string:
		var err error
		intValue, err = strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s must be a valid integer", fieldName)
		}
	default:
		return fmt.Errorf("%s must be an integer", fieldName)
	}

	if intValue < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}

	return nil
}

// validateFloorRange validates floor range format (e.g., "1-5", "3-10")
func (v *BuildingMetadataValidator) validateFloorRange(floorRange string) error {
	floorPattern := regexp.MustCompile(`^(\d+)-(\d+)$`)
	matches := floorPattern.FindStringSubmatch(floorRange)
	if len(matches) != 3 {
		return fmt.Errorf("invalid floor range format, expected 'start-end' (e.g., '1-5')")
	}

	startFloor, _ := strconv.Atoi(matches[1])
	endFloor, _ := strconv.Atoi(matches[2])

	if startFloor >= endFloor {
		return fmt.Errorf("start floor must be less than end floor")
	}

	if startFloor < 1 {
		return fmt.Errorf("floor numbers must be positive")
	}

	return nil
}

// GetMetadataSchema returns the metadata schema for a specific building type
func (v *BuildingMetadataValidator) GetMetadataSchema(buildingType models.BuildingType) map[string]interface{} {
	switch buildingType {
	case models.BuildingTypeResidential:
		return v.getResidentialSchema()
	case models.BuildingTypeCommercial:
		return v.getCommercialSchema()
	case models.BuildingTypeMixed:
		return v.getMixedSchema()
	default:
		return nil
	}
}

// getResidentialSchema returns the schema for residential building metadata
func (v *BuildingMetadataValidator) getResidentialSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"amenities": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"gym", "swimming_pool", "playground", "community_hall", "rooftop_garden", "library", "prayer_room", "parking", "security", "elevator", "generator", "water_tank"},
				},
			},
			"security_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"24_hour_guard", "cctv_only", "card_access", "basic"},
			},
			"maintenance_staff_count": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
				"maximum": 100,
			},
			"parking_spaces": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total":   map[string]interface{}{"type": "integer", "minimum": 0},
					"covered": map[string]interface{}{"type": "integer", "minimum": 0},
					"visitor": map[string]interface{}{"type": "integer", "minimum": 0},
				},
			},
			"utilities": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"backup_generator": map[string]interface{}{"type": "boolean"},
					"water_supply":     map[string]interface{}{"type": "string", "enum": []string{"24_hour", "scheduled", "limited"}},
					"internet_ready":   map[string]interface{}{"type": "boolean"},
				},
			},
		},
	}
}

// getCommercialSchema returns the schema for commercial building metadata
func (v *BuildingMetadataValidator) getCommercialSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"parking_spaces": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total":    map[string]interface{}{"type": "integer", "minimum": 0},
					"customer": map[string]interface{}{"type": "integer", "minimum": 0},
					"staff":    map[string]interface{}{"type": "integer", "minimum": 0},
				},
			},
			"loading_docks": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
			"security_system": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type":           map[string]interface{}{"type": "string", "enum": []string{"basic_cctv", "advanced_cctv", "full_security"}},
					"access_control": map[string]interface{}{"type": "boolean"},
					"fire_safety":    map[string]interface{}{"type": "string", "enum": []string{"basic", "sprinkler_system", "full_system"}},
				},
			},
			"business_hours": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"weekdays": map[string]interface{}{"type": "string", "pattern": "^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$"},
					"weekends": map[string]interface{}{"type": "string", "pattern": "^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$"},
					"holidays": map[string]interface{}{"type": "string", "pattern": "^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$"},
				},
			},
			"facilities": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"elevators":  map[string]interface{}{"type": "integer", "minimum": 0},
					"escalators": map[string]interface{}{"type": "integer", "minimum": 0},
					"food_court": map[string]interface{}{"type": "boolean"},
					"atm":        map[string]interface{}{"type": "boolean"},
				},
			},
		},
	}
}

// getMixedSchema returns the schema for mixed building metadata
func (v *BuildingMetadataValidator) getMixedSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"residential_section": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"floors": map[string]interface{}{
						"type":    "string",
						"pattern": "^\\d+-\\d+$",
					},
					"amenities": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"gym", "swimming_pool", "playground", "community_hall", "rooftop_garden", "library", "prayer_room"},
						},
					},
					"security_type": map[string]interface{}{
						"type": "string",
						"enum": []string{"24_hour_guard", "cctv_only", "card_access", "basic"},
					},
				},
			},
			"commercial_section": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"floors": map[string]interface{}{
						"type":    "string",
						"pattern": "^\\d+-\\d+$",
					},
					"business_hours": map[string]interface{}{
						"type":    "string",
						"pattern": "^([01]?[0-9]|2[0-3]):[0-5][0-9]-([01]?[0-9]|2[0-3]):[0-5][0-9]$",
					},
					"parking_allocation": map[string]interface{}{
						"type":    "integer",
						"minimum": 0,
					},
				},
			},
			"shared_facilities": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"elevators":        map[string]interface{}{"type": "integer", "minimum": 0},
					"parking_total":    map[string]interface{}{"type": "integer", "minimum": 0},
					"backup_generator": map[string]interface{}{"type": "boolean"},
				},
			},
		},
	}
}
