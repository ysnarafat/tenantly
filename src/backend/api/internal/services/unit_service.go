package services

import (
	"fmt"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type UnitService struct {
	unitRepo     interfaces.UnitRepositoryInterface
	buildingRepo interfaces.BuildingRepositoryInterface
	propertyRepo interfaces.PropertyRepositoryInterface
	auditService interfaces.AuditServiceInterface
}

func NewUnitService(
	unitRepo interfaces.UnitRepositoryInterface,
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo interfaces.PropertyRepositoryInterface,
	auditService interfaces.AuditServiceInterface,
) *UnitService {
	return &UnitService{
		unitRepo:     unitRepo,
		buildingRepo: buildingRepo,
		propertyRepo: propertyRepo,
		auditService: auditService,
	}
}

// CreateUnit creates a new unit with building-unit relationship validation
func (s *UnitService) CreateUnit(req *models.CreateUnitRequest, userID, orgID int) (*models.Unit, error) {
	// Validate building-unit relationship and hierarchy integrity
	if err := s.ValidateBuildingUnitRelationship(req.BuildingID, req.PropertyID); err != nil {
		return nil, fmt.Errorf("building-unit relationship validation failed: %w", err)
	}

	// Validate unit number uniqueness within building
	exists, err := s.unitRepo.CheckUnitNumberExists(req.BuildingID, req.UnitNumber, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check unit number uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("unit number already exists in this building")
	}

	// Validate unit type against building type
	if err := s.ValidateUnitTypeForBuilding(req.BuildingID, req.UnitType); err != nil {
		return nil, fmt.Errorf("unit type validation failed: %w", err)
	}

	// Get building to retrieve organization_id
	building, err := s.buildingRepo.GetByID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	// Validate the referenced building actually belongs to the caller's
	// organization — prevents creating a unit inside another org's building (IDOR).
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("building not found")
	}

	// Create unit with organization_id
	unit, err := s.unitRepo.Create(req, building.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}

	// Log audit with building context
	_ = s.auditService.LogUserAction(userID, "CREATE", "units", &unit.ID, nil, map[string]interface{}{
		"unit_id":     unit.ID,
		"building_id": unit.BuildingID,
		"property_id": unit.PropertyID,
		"unit_number": unit.UnitNumber,
		"unit_type":   unit.UnitType,
	})

	return unit, nil
}

// GetUnit retrieves a unit with building and property context
func (s *UnitService) GetUnit(id, orgID int) (*models.UnitWithDetails, error) {
	existing, err := s.unitRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit with details: %w", err)
	}
	if existing.OrganizationID != orgID {
		return nil, fmt.Errorf("failed to get unit with details: unit not found")
	}
	unit, err := s.unitRepo.GetByIDWithDetails(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit with details: %w", err)
	}
	return unit, nil
}

// UpdateUnit updates a unit with building relationship validation
func (s *UnitService) UpdateUnit(id int, req *models.UpdateUnitRequest, userID, orgID int) (*models.Unit, error) {
	// Get existing unit for validation and audit
	existingUnit, err := s.unitRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing unit: %w", err)
	}
	if existingUnit.OrganizationID != orgID {
		return nil, fmt.Errorf("unit not found")
	}

	// Validate unit type change against building type if provided
	if req.UnitType != nil {
		if err := s.ValidateUnitTypeForBuilding(existingUnit.BuildingID, *req.UnitType); err != nil {
			return nil, fmt.Errorf("unit type validation failed: %w", err)
		}
	}

	// Update unit
	updatedUnit, err := s.unitRepo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update unit: %w", err)
	}

	// Log audit with building context
	_ = s.auditService.LogUserAction(userID, "UPDATE", "units", &id, existingUnit, updatedUnit)

	return updatedUnit, nil
}

// DeleteUnit soft deletes a unit with constraint validation
func (s *UnitService) DeleteUnit(id, userID, orgID int) error {
	// Get existing unit for audit
	existingUnit, err := s.unitRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get existing unit: %w", err)
	}
	if existingUnit.OrganizationID != orgID {
		return fmt.Errorf("unit not found")
	}

	// Check if unit has active leases
	hasActiveLeases, err := s.unitRepo.HasActiveLeases(id)
	if err != nil {
		return fmt.Errorf("failed to check active leases: %w", err)
	}
	if hasActiveLeases {
		return fmt.Errorf("cannot delete unit: has active leases")
	}

	// Soft delete unit
	err = s.unitRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete unit: %w", err)
	}

	// Log audit with building context
	_ = s.auditService.LogUserAction(userID, "DELETE", "units", &id, existingUnit, nil)

	return nil
}

// ValidateBuildingUnitRelationship validates building-unit relationships and hierarchy integrity
func (s *UnitService) ValidateBuildingUnitRelationship(buildingID, propertyID int) error {
	// Validate building exists and belongs to the specified property
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return fmt.Errorf("building not found: %w", err)
	}

	if building.PropertyID != propertyID {
		return fmt.Errorf("building does not belong to the specified property")
	}

	// Validate property exists and is active
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return fmt.Errorf("property not found: %w", err)
	}

	if !property.Active {
		return fmt.Errorf("cannot create unit in inactive property")
	}

	if !building.ActiveStatus {
		return fmt.Errorf("cannot create unit in inactive building")
	}

	return nil
}

// ValidateUnitTypeForBuilding validates unit type compatibility with building type
func (s *UnitService) ValidateUnitTypeForBuilding(buildingID int, unitType models.UnitType) error {
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return fmt.Errorf("building not found: %w", err)
	}

	// Define allowed unit types for each building type
	allowedTypes := map[models.BuildingType][]models.UnitType{
		models.BuildingTypeResidential: {
			models.UnitTypeApartment,
			models.UnitTypeParking,
			models.UnitTypeStorage,
		},
		models.BuildingTypeCommercial: {
			models.UnitTypeShop,
			models.UnitTypeOffice,
			models.UnitTypeParking,
			models.UnitTypeStorage,
		},
		models.BuildingTypeMixed: {
			models.UnitTypeShop,
			models.UnitTypeApartment,
			models.UnitTypeOffice,
			models.UnitTypeParking,
			models.UnitTypeStorage,
			models.UnitTypeOther,
		},
	}

	allowed := allowedTypes[building.BuildingType]
	for _, allowedType := range allowed {
		if unitType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("unit type %s is not allowed in %s building", unitType, building.BuildingType)
}

// GetUnitsByBuilding retrieves units for a specific building with pagination
func (s *UnitService) GetUnitsByBuilding(buildingID int, page, pageSize, orgID int) ([]*models.UnitWithDetails, int, error) {
	// Validate building exists
	_, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, 0, fmt.Errorf("building not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	units, total, err := s.unitRepo.GetByBuildingWithDetails(buildingID, pageSize, offset, orgID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get units by building: %w", err)
	}

	return units, total, nil
}

// GetUnitsByProperty retrieves units for a specific property with building context
func (s *UnitService) GetUnitsByProperty(propertyID int, page, pageSize, orgID int) ([]*models.UnitWithDetails, int, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, 0, fmt.Errorf("property not found: %w", err)
	}

	// Calculate offset
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	units, total, err := s.unitRepo.GetByPropertyWithDetails(propertyID, pageSize, offset, orgID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get units by property: %w", err)
	}

	return units, total, nil
}

// ValidateHierarchyIntegrity validates the complete property-building-unit hierarchy
func (s *UnitService) ValidateHierarchyIntegrity(unitID int) error {
	unit, err := s.unitRepo.GetByID(unitID)
	if err != nil {
		return fmt.Errorf("unit not found: %w", err)
	}

	// Validate building exists and is active
	building, err := s.buildingRepo.GetByID(unit.BuildingID)
	if err != nil {
		return fmt.Errorf("building not found for unit: %w", err)
	}

	// Validate property exists and is active
	property, err := s.propertyRepo.GetByID(unit.PropertyID)
	if err != nil {
		return fmt.Errorf("property not found for unit: %w", err)
	}

	// Validate hierarchy consistency
	if building.PropertyID != unit.PropertyID {
		return fmt.Errorf("hierarchy integrity violation: building property mismatch")
	}

	if !property.Active {
		return fmt.Errorf("hierarchy integrity violation: property is inactive")
	}

	if !building.ActiveStatus {
		return fmt.Errorf("hierarchy integrity violation: building is inactive")
	}

	return nil
}

// GetUnitHierarchyContext returns complete hierarchy context for a unit
func (s *UnitService) GetUnitHierarchyContext(unitID, orgID int) (map[string]interface{}, error) {
	unit, err := s.unitRepo.GetByIDWithDetails(unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit details: %w", err)
	}

	building, err := s.buildingRepo.GetByID(unit.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building details: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("unit not found")
	}

	property, err := s.propertyRepo.GetByID(unit.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property details: %w", err)
	}

	context := map[string]interface{}{
		"unit": map[string]interface{}{
			"id":          unit.ID,
			"unit_number": unit.UnitNumber,
			"unit_name":   unit.UnitName,
			"unit_type":   unit.UnitType,
			"floor":       unit.Floor,
			"section":     unit.Section,
			"active":      unit.Active,
		},
		"building": map[string]interface{}{
			"id":                building.ID,
			"building_name":     building.BuildingName,
			"building_code":     building.BuildingCode,
			"building_type":     building.BuildingType,
			"total_floors":      building.TotalFloors,
			"has_elevator":      building.HasElevator,
			"construction_year": building.ConstructionYear,
			"active_status":     building.ActiveStatus,
		},
		"property": map[string]interface{}{
			"id":            property.ID,
			"property_name": property.PropertyName,
			"property_code": property.PropertyCode,
			"property_type": property.PropertyType,
			"address":       property.Address,
			"city":          property.City,
			"active":        property.Active,
		},
	}

	return context, nil
}
