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
func (s *PropertyService) GetProperty(id int) (*models.Property, error) {
	property, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	return property, nil
}

// GetPropertyWithStats retrieves a property with aggregated statistics
func (s *PropertyService) GetPropertyWithStats(id int) (*models.PropertyWithStats, error) {
	property, err := s.propertyRepo.GetByIDWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property with stats: %w", err)
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
func (s *PropertyService) UpdateProperty(id int, req *models.UpdatePropertyRequest, userID int) (*models.Property, error) {
	// Get existing property for audit logging
	existingProperty, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing property: %w", err)
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
func (s *PropertyService) DeleteProperty(id int, userID int) error {
	// Get existing property for audit logging
	existingProperty, err := s.propertyRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get existing property: %w", err)
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

// GetPropertyAggregations returns property-level aggregations
func (s *PropertyService) GetPropertyAggregations(id int) (map[string]interface{}, error) {
	property, err := s.propertyRepo.GetByIDWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property aggregations: %w", err)
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

// SearchProperties performs advanced search on properties
func (s *PropertyService) SearchProperties(searchTerm string, filters map[string]interface{}, page, pageSize int) ([]*models.Property, int, error) {
	// Add search term to filters
	if strings.TrimSpace(searchTerm) != "" {
		filters["search"] = strings.TrimSpace(searchTerm)
	}

	return s.ListProperties(filters, page, pageSize)
}
