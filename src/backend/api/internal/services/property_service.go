package services

import (
	"fmt"
	"strings"

	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/repositories"
)

type PropertyService struct {
	propertyRepo *repositories.PropertyRepository
	auditService *database.AuditService
}

func NewPropertyService(propertyRepo *repositories.PropertyRepository, auditService *database.AuditService) *PropertyService {
	return &PropertyService{
		propertyRepo: propertyRepo,
		auditService: auditService,
	}
}

// CreateProperty creates a new property with validation
func (s *PropertyService) CreateProperty(req *models.CreatePropertyRequest, userID int) (*models.Property, error) {
	// Validate property type
	if err := s.validatePropertyType(string(req.PropertyType)); err != nil {
		return nil, err
	}

	// Check property name uniqueness
	exists, err := s.propertyRepo.CheckPropertyNameExists(req.PropertyName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check property name uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("property name already exists")
	}

	// Check property code uniqueness
	exists, err = s.propertyRepo.CheckPropertyCodeExists(req.PropertyCode, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check property code uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("property code already exists")
	}

	// Create property
	property, err := s.propertyRepo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create property: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "CREATE", "properties", &property.ID, nil, property)

	return property, nil
}

// GetProperty retrieves a property by ID
func (s *PropertyService) GetProperty(id, orgID int) (*models.Property, error) {
	property, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	if property.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}
	return property, nil
}

// GetPropertyWithStats retrieves a property with aggregated statistics including building context
func (s *PropertyService) GetPropertyWithStats(id, orgID int) (*models.PropertyWithStats, error) {
	property, err := s.propertyRepo.GetByIDWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property with stats: %w", err)
	}
	if property.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}
	return property, nil
}

// ListProperties retrieves properties with filtering and pagination
func (s *PropertyService) ListProperties(filters map[string]interface{}, page, pageSize int) ([]*models.Property, int, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Validate property type filter if provided
	if propertyType, ok := filters["property_type"]; ok && propertyType != "" {
		if err := s.validatePropertyType(propertyType.(string)); err != nil {
			return nil, 0, err
		}
	}

	properties, total, err := s.propertyRepo.List(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list properties: %w", err)
	}

	return properties, total, nil
}

// UpdateProperty updates a property with validation
func (s *PropertyService) UpdateProperty(id int, req *models.UpdatePropertyRequest, userID, orgID int) (*models.Property, error) {
	// Get existing property for audit logging
	existingProperty, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing property: %w", err)
	}
	if existingProperty.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}

	// Validate property type if provided
	if req.PropertyType != nil {
		if err := s.validatePropertyType(string(*req.PropertyType)); err != nil {
			return nil, err
		}
	}

	// Check property name uniqueness if provided
	if req.PropertyName != nil {
		exists, err := s.propertyRepo.CheckPropertyNameExists(*req.PropertyName, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check property name uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("property name already exists")
		}
	}

	// Update property
	updatedProperty, err := s.propertyRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update property: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "UPDATE", "properties", &id, existingProperty, updatedProperty)

	return updatedProperty, nil
}

// DeleteProperty soft deletes a property with cascade validation
func (s *PropertyService) DeleteProperty(id int, userID, orgID int) error {
	// Get existing property for audit logging
	existingProperty, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get existing property: %w", err)
	}
	if existingProperty.OrganizationID != orgID {
		return fmt.Errorf("property not found")
	}

	// Check if property has active buildings
	hasBuildings, err := s.propertyRepo.HasActiveBuildings(id)
	if err != nil {
		return fmt.Errorf("failed to check active buildings: %w", err)
	}
	if hasBuildings {
		return fmt.Errorf("cannot delete property: has active buildings")
	}

	// Check if property has active units
	hasUnits, err := s.propertyRepo.HasActiveUnits(id)
	if err != nil {
		return fmt.Errorf("failed to check active units: %w", err)
	}
	if hasUnits {
		return fmt.Errorf("cannot delete property: has active units")
	}

	// Soft delete property
	err = s.propertyRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete property: %w", err)
	}

	// Log audit
	s.auditService.LogUserAction(userID, "DELETE", "properties", &id, existingProperty, nil)

	return nil
}

// validatePropertyType validates the property type
func (s *PropertyService) validatePropertyType(propertyType string) error {
	validTypes := map[string]bool{
		"Residential": true,
		"Commercial":  true,
		"Mixed":       true,
	}

	if !validTypes[propertyType] {
		return fmt.Errorf("invalid property type: must be Residential, Commercial, or Mixed")
	}

	return nil
}

// GetPropertyAggregations returns property-level aggregations with building breakdowns
func (s *PropertyService) GetPropertyAggregations(id, orgID int) (map[string]interface{}, error) {
	property, err := s.propertyRepo.GetByIDWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property aggregations: %w", err)
	}
	if property.OrganizationID != orgID {
		return nil, fmt.Errorf("property not found")
	}

	occupancyRate := 0.0
	if property.UnitCount > 0 {
		occupancyRate = float64(property.OccupiedUnits) / float64(property.UnitCount) * 100
	}

	aggregations := map[string]interface{}{
		"total_buildings": property.BuildingCount,
		"total_units":     property.UnitCount,
		"occupied_units":  property.OccupiedUnits,
		"vacant_units":    property.UnitCount - property.OccupiedUnits,
		"occupancy_rate":  occupancyRate,
		"total_revenue":   property.TotalRevenue,
		"property_type":   property.PropertyType,
		"active_status":   property.Active,
	}

	// Building breakdowns and type distribution would be implemented in repository layer
	// For now, we provide basic aggregations

	return aggregations, nil
}

// ValidatePropertyAccess validates if a user has access to a property based on their role
func (s *PropertyService) ValidatePropertyAccess(userRole string, propertyID int) error {
	// Admin has access to all properties
	if userRole == "Admin" {
		return nil
	}

	// PropertyManager has access to all properties
	if userRole == "PropertyManager" {
		return nil
	}

	// Accountant has read-only access to all properties
	if userRole == "Accountant" {
		return nil
	}

	return fmt.Errorf("insufficient permissions to access property")
}

// GetPropertyBuildingCount returns the count of active buildings for a property
func (s *PropertyService) GetPropertyBuildingCount(propertyID int) (int, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return 0, fmt.Errorf("property validation failed: %w", err)
	}

	count, err := s.propertyRepo.GetBuildingCount(propertyID)
	if err != nil {
		return 0, fmt.Errorf("failed to get building count: %w", err)
	}

	return count, nil
}

// GetPropertyBuildingSummary returns building summary statistics for a property
func (s *PropertyService) GetPropertyBuildingSummary(propertyID int) (map[string]interface{}, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	summary, err := s.propertyRepo.GetBuildingSummary(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building summary: %w", err)
	}

	return summary, nil
}

// SearchProperties performs advanced search on properties
func (s *PropertyService) SearchProperties(searchTerm string, filters map[string]interface{}, page, pageSize int) ([]*models.Property, int, error) {
	// Add search term to filters
	if strings.TrimSpace(searchTerm) != "" {
		filters["search"] = strings.TrimSpace(searchTerm)
	}

	return s.ListProperties(filters, page, pageSize)
}

// GetPropertyWithBuildingContext returns property with enhanced building context
func (s *PropertyService) GetPropertyWithBuildingContext(id int) (*models.PropertyWithBuildings, error) {
	// Get property with stats
	property, err := s.propertyRepo.GetByIDWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	// Get buildings for the property (this would need to be implemented in repository)
	// For now, return basic structure
	result := &models.PropertyWithBuildings{
		Property: models.Property{
			ID:             property.ID,
			PropertyName:   property.PropertyName,
			PropertyCode:   property.PropertyCode,
			Address:        property.Address,
			City:           property.City,
			PostalCode:     property.PostalCode,
			PropertyType:   property.PropertyType,
			TotalBuildings: property.TotalBuildings,
			Metadata:       property.Metadata,
			Active:         property.Active,
			CreatedAt:      property.CreatedAt,
			UpdatedAt:      property.UpdatedAt,
		},
		BuildingCount: property.BuildingCount,
		Statistics: &models.PropertyStatistics{
			TotalUnits:     property.UnitCount,
			OccupiedUnits:  property.OccupiedUnits,
			VacantUnits:    property.UnitCount - property.OccupiedUnits,
			OccupancyRate:  float64(property.OccupiedUnits) / float64(property.UnitCount) * 100,
			TotalRevenue:   property.TotalRevenue,
			AverageRevenue: property.TotalRevenue / float64(property.UnitCount),
		},
	}

	return result, nil
}

// GetPropertiesWithBuildingStats returns properties with building-level statistics
func (s *PropertyService) GetPropertiesWithBuildingStats(filters map[string]interface{}, page, pageSize int) ([]*models.PropertyWithStats, int, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get properties with enhanced building context
	properties, total, err := s.propertyRepo.List(filters, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list properties: %w", err)
	}

	// Enhance each property with building statistics
	enhancedProperties := make([]*models.PropertyWithStats, 0, len(properties))
	for _, property := range properties {
		enhanced, err := s.propertyRepo.GetByIDWithStats(property.ID)
		if err != nil {
			continue // Skip properties with errors but don't fail the entire request
		}
		enhancedProperties = append(enhancedProperties, enhanced)
	}

	return enhancedProperties, total, nil
}
