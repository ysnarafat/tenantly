package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

// propertyExistenceChecker is a minimal interface for the subset of property repo used here.
type propertyExistenceChecker interface {
	GetByID(id int) (*models.Property, error)
}

// BuildingService implements the BuildingServiceInterface
type BuildingService struct {
	buildingRepo      interfaces.BuildingRepositoryInterface
	propertyRepo      propertyExistenceChecker
	auditService      interfaces.AuditServiceInterface
	metadataValidator interfaces.MetadataValidatorInterface
}

// NewBuildingService creates a new building service instance
func NewBuildingService(
	buildingRepo interfaces.BuildingRepositoryInterface,
	propertyRepo propertyExistenceChecker,
	auditService interfaces.AuditServiceInterface,
	metadataValidator interfaces.MetadataValidatorInterface,
) *BuildingService {
	return &BuildingService{
		buildingRepo:      buildingRepo,
		propertyRepo:      propertyRepo,
		auditService:      auditService,
		metadataValidator: metadataValidator,
	}
}

// CreateBuilding creates a new building with property validation and building code uniqueness checks
func (s *BuildingService) CreateBuilding(req *models.CreateBuildingRequest) (*models.Building, error) {
	// Validate building creation request
	if err := s.ValidateBuildingCreation(req); err != nil {
		return nil, err
	}

	// Validate property exists
	_, err := s.propertyRepo.GetByID(req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	// Check building code uniqueness within property
	if err := s.ValidateBuildingCodeUniqueness(req.PropertyID, req.BuildingCode, nil); err != nil {
		return nil, err
	}

	// Validate metadata based on building type
	if req.Metadata != nil {
		if err := s.metadataValidator.ValidateMetadata(req.BuildingType, req.Metadata); err != nil {
			return nil, fmt.Errorf("metadata validation failed: %w", err)
		}
	}

	// Create building entity
	building := &models.Building{
		PropertyID:       req.PropertyID,
		OrganizationID:   req.OrganizationID,
		BuildingName:     req.BuildingName,
		BuildingCode:     req.BuildingCode,
		BuildingType:     req.BuildingType,
		TotalFloors:      req.TotalFloors,
		HasElevator:      req.HasElevator,
		ConstructionYear: req.ConstructionYear,
		Metadata:         req.Metadata,
		ActiveStatus:     true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Create building in repository
	if err := s.buildingRepo.Create(building); err != nil {
		return nil, fmt.Errorf("failed to create building: %w", err)
	}

	// Log audit action
	_ = s.auditService.LogSystemAction("CREATE", "buildings", &building.ID, nil, building)

	return building, nil
}

// GetBuilding retrieves a building by ID with proper error handling
func (s *BuildingService) GetBuilding(id, orgID int) (*models.Building, error) {
	building, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("failed to get building: building not found")
	}
	return building, nil
}

// GetBuildingWithStats retrieves a building with aggregated statistics and data enrichment
func (s *BuildingService) GetBuildingWithStats(id, orgID int) (*models.BuildingWithStats, error) {
	building, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get building with stats: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("failed to get building with stats: building not found")
	}
	buildingStats, err := s.buildingRepo.GetWithStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get building with stats: %w", err)
	}
	return buildingStats, nil
}

// UpdateBuilding updates a building with metadata validation and audit logging
func (s *BuildingService) UpdateBuilding(id int, req *models.UpdateBuildingRequest, orgID int) (*models.Building, error) {
	// Get existing building for audit logging
	existingBuilding, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing building: %w", err)
	}
	if existingBuilding.OrganizationID != orgID {
		return nil, fmt.Errorf("building not found")
	}

	// Validate building update request
	if err := s.ValidateBuildingUpdate(id, req); err != nil {
		return nil, err
	}

	// Validate metadata if provided
	if req.Metadata != nil {
		buildingType := existingBuilding.BuildingType
		if req.BuildingType != nil {
			buildingType = *req.BuildingType
		}
		if err := s.metadataValidator.ValidateMetadata(buildingType, *req.Metadata); err != nil {
			return nil, fmt.Errorf("metadata validation failed: %w", err)
		}
	}

	// Prepare updates map
	updates := make(map[string]interface{})
	if req.BuildingName != nil {
		updates["building_name"] = *req.BuildingName
	}
	if req.BuildingType != nil {
		updates["building_type"] = *req.BuildingType
	}
	if req.TotalFloors != nil {
		updates["total_floors"] = *req.TotalFloors
	}
	if req.HasElevator != nil {
		updates["has_elevator"] = *req.HasElevator
	}
	if req.ConstructionYear != nil {
		updates["construction_year"] = *req.ConstructionYear
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}
	if req.ActiveStatus != nil {
		updates["active_status"] = *req.ActiveStatus
	}

	// Update building in repository
	if err := s.buildingRepo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to update building: %w", err)
	}

	// Get updated building
	updatedBuilding, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated building: %w", err)
	}

	// Log audit action
	_ = s.auditService.LogSystemAction("UPDATE", "buildings", &id, existingBuilding, updatedBuilding)

	return updatedBuilding, nil
}

// DeleteBuilding performs soft delete with active unit constraint validation
func (s *BuildingService) DeleteBuilding(id, orgID int) error {
	// Get existing building for audit logging
	existingBuilding, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get existing building: %w", err)
	}
	if existingBuilding.OrganizationID != orgID {
		return fmt.Errorf("building not found")
	}

	// Validate building deletion constraints
	if err := s.ValidateBuildingDeletion(id); err != nil {
		return err
	}

	// Perform soft delete
	if err := s.buildingRepo.SoftDelete(id); err != nil {
		return fmt.Errorf("failed to delete building: %w", err)
	}

	// Log audit action
	_ = s.auditService.LogSystemAction("DELETE", "buildings", &id, existingBuilding, nil)

	return nil
}

// GetBuildingsByProperty retrieves all buildings for a specific property with filtering and sorting
func (s *BuildingService) GetBuildingsByProperty(propertyID int) ([]*models.Building, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings by property: %w", err)
	}

	return buildings, nil
}

// GetBuildingsByPropertyWithStats retrieves buildings for a property with statistics
func (s *BuildingService) GetBuildingsByPropertyWithStats(propertyID int) ([]*models.BuildingWithStats, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings by property: %w", err)
	}

	var buildingsWithStats []*models.BuildingWithStats
	for _, building := range buildings {
		stats, err := s.buildingRepo.GetWithStats(building.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for building %d: %w", building.ID, err)
		}
		buildingsWithStats = append(buildingsWithStats, stats)
	}

	return buildingsWithStats, nil
}

// GetBuildingByPropertyAndCode retrieves a building by property ID and building code
func (s *BuildingService) GetBuildingByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	building, err := s.buildingRepo.GetByPropertyAndCode(propertyID, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get building by property and code: %w", err)
	}

	return building, nil
}

// BulkCreateBuildings creates multiple buildings with transaction handling and rollback
func (s *BuildingService) BulkCreateBuildings(req *models.BulkCreateBuildingsRequest) ([]*models.Building, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	// Validate all building requests
	buildingCodes := make(map[string]bool)
	for i, buildingReq := range req.Buildings {
		// Set property ID for each building
		buildingReq.PropertyID = req.PropertyID

		// Validate building creation
		if err := s.ValidateBuildingCreation(&buildingReq); err != nil {
			return nil, fmt.Errorf("validation failed for building %d: %w", i+1, err)
		}

		// Check for duplicate building codes within the request
		if buildingCodes[buildingReq.BuildingCode] {
			return nil, fmt.Errorf("duplicate building code '%s' in request", buildingReq.BuildingCode)
		}
		buildingCodes[buildingReq.BuildingCode] = true

		// Check building code uniqueness in database
		if err := s.ValidateBuildingCodeUniqueness(req.PropertyID, buildingReq.BuildingCode, nil); err != nil {
			return nil, fmt.Errorf("building code validation failed for building %d: %w", i+1, err)
		}

		// Validate metadata
		if buildingReq.Metadata != nil {
			if err := s.metadataValidator.ValidateMetadata(buildingReq.BuildingType, buildingReq.Metadata); err != nil {
				return nil, fmt.Errorf("metadata validation failed for building %d: %w", i+1, err)
			}
		}
	}

	// Create building entities
	var buildings []*models.Building
	for _, buildingReq := range req.Buildings {
		building := &models.Building{
			PropertyID:       buildingReq.PropertyID,
			OrganizationID:   req.OrganizationID,
			BuildingName:     buildingReq.BuildingName,
			BuildingCode:     buildingReq.BuildingCode,
			BuildingType:     buildingReq.BuildingType,
			TotalFloors:      buildingReq.TotalFloors,
			HasElevator:      buildingReq.HasElevator,
			ConstructionYear: buildingReq.ConstructionYear,
			Metadata:         buildingReq.Metadata,
			ActiveStatus:     true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		buildings = append(buildings, building)
	}

	// Bulk create buildings with transaction handling
	if err := s.buildingRepo.BulkCreate(buildings); err != nil {
		return nil, fmt.Errorf("failed to bulk create buildings: %w", err)
	}

	// Log audit actions for each building
	for _, building := range buildings {
		_ = s.auditService.LogSystemAction("CREATE", "buildings", &building.ID, nil, building)
	}

	return buildings, nil
}

// SearchBuildings performs advanced search with filtering
func (s *BuildingService) SearchBuildings(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	// Validate property ID if provided
	if filters.PropertyID != nil {
		_, err := s.propertyRepo.GetByID(*filters.PropertyID)
		if err != nil {
			return nil, fmt.Errorf("property validation failed: %w", err)
		}
	}

	// Set default limit if not provided
	if filters.Limit <= 0 {
		filters.Limit = 50
	}

	buildings, err := s.buildingRepo.Search(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search buildings: %w", err)
	}

	return buildings, nil
}

// GetBuildingAnalytics retrieves building performance metrics
func (s *BuildingService) GetBuildingAnalytics(id, orgID int) (*models.BuildingAnalytics, error) {
	building, err := s.buildingRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("building validation failed: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("building validation failed: building not found")
	}

	analytics, err := s.buildingRepo.GetAnalytics(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	return analytics, nil
}

// GetPropertyBuildingAnalytics retrieves analytics for all buildings in a property
func (s *BuildingService) GetPropertyBuildingAnalytics(propertyID int) ([]*models.BuildingAnalytics, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings by property: %w", err)
	}

	var analyticsResults []*models.BuildingAnalytics
	for _, building := range buildings {
		analytics, err := s.buildingRepo.GetAnalytics(building.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get analytics for building %d: %w", building.ID, err)
		}
		analyticsResults = append(analyticsResults, analytics)
	}

	return analyticsResults, nil
}

// CalculateBuildingOccupancyRate calculates the occupancy rate for a building
func (s *BuildingService) CalculateBuildingOccupancyRate(id int) (float64, error) {
	analytics, err := s.buildingRepo.GetAnalytics(id)
	if err != nil {
		return 0, fmt.Errorf("building validation failed: %w", err)
	}

	return analytics.OccupancyRate, nil
}

// CalculateBuildingRevenue calculates the monthly revenue for a building
func (s *BuildingService) CalculateBuildingRevenue(id int) (float64, error) {
	analytics, err := s.buildingRepo.GetAnalytics(id)
	if err != nil {
		return 0, fmt.Errorf("building validation failed: %w", err)
	}

	return analytics.MonthlyRevenue, nil
}

// GetBuildingUnitCounts returns unit counts for a building
func (s *BuildingService) GetBuildingUnitCounts(id int) (total int, occupied int, vacant int, err error) {
	analytics, err := s.buildingRepo.GetAnalytics(id)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("building validation failed: %w", err)
	}

	return analytics.UnitCount, analytics.OccupiedUnits, analytics.VacantUnits, nil
}

// ValidateBuildingCreation validates building creation request
func (s *BuildingService) ValidateBuildingCreation(req *models.CreateBuildingRequest) error {
	// Validate building type
	if err := s.validateBuildingType(req.BuildingType); err != nil {
		return err
	}

	// Validate floor count
	if req.TotalFloors < 1 {
		return fmt.Errorf("total floors must be at least 1")
	}

	// Validate construction year if provided
	if req.ConstructionYear != nil {
		currentYear := time.Now().Year()
		if *req.ConstructionYear < 1800 || *req.ConstructionYear > currentYear+5 {
			return fmt.Errorf("construction year must be between 1800 and %d", currentYear+5)
		}
	}

	// Validate building name and code
	if req.BuildingName == "" {
		return fmt.Errorf("building name is required")
	}
	if req.BuildingCode == "" {
		return fmt.Errorf("building code is required")
	}

	return nil
}

// ValidateBuildingUpdate validates building update request
func (s *BuildingService) ValidateBuildingUpdate(id int, req *models.UpdateBuildingRequest) error {
	// Validate building type if provided
	if req.BuildingType != nil {
		if err := s.validateBuildingType(*req.BuildingType); err != nil {
			return err
		}
	}

	// Validate floor count if provided
	if req.TotalFloors != nil && *req.TotalFloors < 1 {
		return fmt.Errorf("total floors must be at least 1")
	}

	// Validate construction year if provided
	if req.ConstructionYear != nil {
		currentYear := time.Now().Year()
		if *req.ConstructionYear < 1800 || *req.ConstructionYear > currentYear+5 {
			return fmt.Errorf("construction year must be between 1800 and %d", currentYear+5)
		}
	}

	return nil
}

// ValidateBuildingDeletion validates building deletion constraints
func (s *BuildingService) ValidateBuildingDeletion(id int) error {
	// The repository's SoftDelete method already checks for active units
	// This method can be extended for additional business rules
	return nil
}

// ValidateBuildingCodeUniqueness validates building code uniqueness within property
func (s *BuildingService) ValidateBuildingCodeUniqueness(propertyID int, code string, excludeID *int) error {
	existingBuilding, err := s.buildingRepo.GetByPropertyAndCode(propertyID, code)
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

	return fmt.Errorf("building code '%s' already exists in this property", code)
}

// GetPropertyBuildingsWithPagination retrieves buildings for a property with enhanced pagination, sorting, and filtering
func (s *BuildingService) GetPropertyBuildingsWithPagination(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string, includeStats bool) (*models.BuildingListResponse, error) {
	// Validate property exists (already done by middleware, but keeping for service integrity)
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	// Get total count for pagination
	totalCount, err := s.buildingRepo.CountByProperty(propertyID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count buildings: %w", err)
	}

	// Get buildings with sorting
	buildings, err := s.buildingRepo.GetByPropertyWithSorting(propertyID, filters, sortBy, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	// Calculate pagination info
	totalPages := (totalCount + filters.Limit - 1) / filters.Limit
	currentPage := (filters.Offset / filters.Limit) + 1
	hasNext := currentPage < totalPages
	hasPrev := currentPage > 1

	pagination := &models.PaginationInfo{
		CurrentPage: currentPage,
		PageSize:    filters.Limit,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}

	response := &models.BuildingListResponse{
		Pagination: pagination,
	}

	// Get property statistics
	propertyStats, err := s.GetPropertyStatistics(propertyID)
	if err == nil {
		response.Statistics = propertyStats
	}

	if includeStats {
		// Convert to buildings with stats
		var buildingsWithStats []*models.BuildingWithStats
		for _, building := range buildings {
			stats, err := s.buildingRepo.GetWithStats(building.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get stats for building %d: %w", building.ID, err)
			}
			buildingsWithStats = append(buildingsWithStats, stats)
		}
		response.Buildings = buildingsWithStats
	} else {
		response.Buildings = buildings
	}

	return response, nil
}

// GetPropertyStatistics calculates comprehensive statistics for a property
func (s *BuildingService) GetPropertyStatistics(propertyID int) (*models.PropertyStatistics, error) {
	// Validate property exists
	_, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	// Get property analytics from all buildings
	buildingAnalytics, err := s.GetPropertyBuildingAnalytics(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property building analytics: %w", err)
	}

	// Aggregate statistics
	stats := &models.PropertyStatistics{}
	totalRevenue := 0.0

	for _, analytics := range buildingAnalytics {
		stats.TotalUnits += analytics.UnitCount
		stats.OccupiedUnits += analytics.OccupiedUnits
		stats.VacantUnits += analytics.VacantUnits
		totalRevenue += analytics.MonthlyRevenue
	}

	stats.TotalRevenue = totalRevenue

	// Calculate occupancy rate
	if stats.TotalUnits > 0 {
		stats.OccupancyRate = float64(stats.OccupiedUnits) / float64(stats.TotalUnits) * 100
	}

	// Calculate average revenue per unit
	if stats.TotalUnits > 0 {
		stats.AverageRevenue = totalRevenue / float64(stats.TotalUnits)
	}

	return stats, nil
}

// GetPropertyWithBuildingsHierarchy returns property with buildings in hierarchical structure
func (s *BuildingService) GetPropertyWithBuildingsHierarchy(propertyID int, includeStats bool) (*models.PropertyWithBuildings, error) {
	// Get property
	property, err := s.propertyRepo.GetByID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("property validation failed: %w", err)
	}

	// Get buildings
	buildings, err := s.buildingRepo.GetByPropertyID(propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	// Get property statistics
	propertyStats, err := s.GetPropertyStatistics(propertyID)
	if err != nil {
		propertyStats = nil // Don't fail if stats can't be calculated
	}

	response := &models.PropertyWithBuildings{
		Property:      *property,
		Buildings:     buildings,
		BuildingCount: len(buildings),
		Statistics:    propertyStats,
	}

	return response, nil
}

// validateBuildingType validates the building type
func (s *BuildingService) validateBuildingType(buildingType models.BuildingType) error {
	validTypes := map[models.BuildingType]bool{
		models.BuildingTypeResidential: true,
		models.BuildingTypeCommercial:  true,
		models.BuildingTypeMixed:       true,
	}

	if !validTypes[buildingType] {
		return fmt.Errorf("invalid building type: must be Residential, Commercial, or Mixed")
	}

	return nil
}

// AdvancedSearchBuildings performs advanced building search with metadata queries and enhanced filtering
func (s *BuildingService) AdvancedSearchBuildings(req *models.BuildingSearchRequest) (*models.BuildingListResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Validate property ID if provided
	if req.PropertyID != nil {
		_, err := s.propertyRepo.GetByID(*req.PropertyID)
		if err != nil {
			return nil, fmt.Errorf("property validation failed: %w", err)
		}
	}

	// Validate building type if provided
	if req.BuildingType != "" {
		buildingType := models.BuildingType(req.BuildingType)
		if err := s.validateBuildingType(buildingType); err != nil {
			return nil, fmt.Errorf("invalid building type: %w", err)
		}
	}

	// Perform advanced search
	buildings, totalCount, err := s.buildingRepo.AdvancedSearch(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform advanced search: %w", err)
	}

	// Calculate pagination info
	totalPages := (totalCount + req.PageSize - 1) / req.PageSize
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	pagination := &models.PaginationInfo{
		CurrentPage: req.Page,
		PageSize:    req.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}

	response := &models.BuildingListResponse{
		Pagination: pagination,
	}

	if req.IncludeStats {
		// Convert to buildings with stats
		var buildingsWithStats []*models.BuildingWithStats
		for _, building := range buildings {
			stats, err := s.buildingRepo.GetWithStats(building.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get stats for building %d: %w", building.ID, err)
			}
			buildingsWithStats = append(buildingsWithStats, stats)
		}
		response.Buildings = buildingsWithStats
	} else {
		response.Buildings = buildings
	}

	return response, nil
}

// GetBuildingUnits retrieves units for a specific building with pagination
func (s *BuildingService) GetBuildingUnits(buildingID, orgID int, page, pageSize int) (*models.BuildingUnitsResponse, error) {
	// Validate building exists and belongs to the caller's organization (IDOR guard)
	building, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("building validation failed: %w", err)
	}
	if building.OrganizationID != orgID {
		return nil, fmt.Errorf("building validation failed: building not found")
	}

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get building units
	units, totalCount, err := s.buildingRepo.GetBuildingUnits(buildingID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get building units: %w", err)
	}

	// Calculate summary statistics
	totalUnits := 0
	occupiedUnits := 0

	for _, unit := range units {
		totalUnits++
		if unit.LeaseActive {
			occupiedUnits++
		}
	}

	// Calculate pagination info
	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := &models.PaginationInfo{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
	}

	response := &models.BuildingUnitsResponse{
		BuildingID:   building.ID,
		BuildingName: building.BuildingName,
		BuildingCode: building.BuildingCode,
		Units:        units,
		Pagination:   pagination,
	}

	response.Summary.TotalUnits = totalUnits
	response.Summary.OccupiedUnits = occupiedUnits
	response.Summary.VacantUnits = totalUnits - occupiedUnits
	response.Summary.TotalRevenue = 0

	return response, nil
}

// GetBuildingMetadataSchema returns metadata schema information for a building type
func (s *BuildingService) GetBuildingMetadataSchema(buildingType models.BuildingType) (*models.MetadataSchemaResponse, error) {
	// Validate building type
	if err := s.validateBuildingType(buildingType); err != nil {
		return nil, fmt.Errorf("invalid building type: %w", err)
	}

	// Get schema from metadata validator
	schema := s.metadataValidator.GetMetadataSchema(buildingType)

	// Create examples based on building type
	var examples map[string]interface{}
	var description string

	switch buildingType {
	case models.BuildingTypeResidential:
		examples = map[string]interface{}{
			"amenities":               []string{"gym", "swimming_pool", "playground"},
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
		description = "Metadata schema for residential buildings including amenities, security, and utilities information"

	case models.BuildingTypeCommercial:
		examples = map[string]interface{}{
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
		description = "Metadata schema for commercial buildings including parking, security, business hours, and facilities"

	case models.BuildingTypeMixed:
		examples = map[string]interface{}{
			"residential_section": map[string]interface{}{
				"floors":        "3-10",
				"amenities":     []string{"gym", "rooftop_garden"},
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
		description = "Metadata schema for mixed-use buildings with separate residential and commercial sections"
	}

	response := &models.MetadataSchemaResponse{
		BuildingType: string(buildingType),
		Schema:       schema,
		Examples:     examples,
		Description:  description,
	}

	return response, nil
}

// ExportBuildingData exports building data in various formats for reporting and analysis
func (s *BuildingService) ExportBuildingData(req *models.BuildingExportRequest) ([]byte, string, error) {
	// Set default format
	if req.Format == "" {
		req.Format = "json"
	}

	// Validate property ID if provided
	if req.PropertyID != nil {
		_, err := s.propertyRepo.GetByID(*req.PropertyID)
		if err != nil {
			return nil, "", fmt.Errorf("property validation failed: %w", err)
		}
	}

	// Build search filters
	filters := &models.BuildingSearchFilters{
		PropertyID:   req.PropertyID,
		ActiveStatus: req.ActiveStatus,
		Limit:        1000, // Large limit for export
		Offset:       0,
	}

	if req.BuildingType != "" {
		buildingType := models.BuildingType(req.BuildingType)
		filters.BuildingType = &buildingType
	}

	// Get buildings
	var buildings interface{}
	if req.IncludeStats {
		buildingList, err := s.buildingRepo.Search(filters)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get buildings: %w", err)
		}

		var buildingsWithStats []*models.BuildingWithStats
		for _, building := range buildingList {
			stats, err := s.buildingRepo.GetWithStats(building.ID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to get stats for building %d: %w", building.ID, err)
			}
			buildingsWithStats = append(buildingsWithStats, stats)
		}
		buildings = buildingsWithStats
	} else {
		buildingList, err := s.buildingRepo.Search(filters)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get buildings: %w", err)
		}
		buildings = buildingList
	}

	// Export data based on format
	switch req.Format {
	case "json":
		data, err := json.Marshal(buildings)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return data, "application/json", nil

	case "csv":
		// For CSV, we'll create a simplified format
		csvData := "ID,Property ID,Building Name,Building Code,Building Type,Total Floors,Has Elevator,Active Status,Created At\n"

		var buildingList []*models.Building
		if req.IncludeStats {
			statsBuildings := buildings.([]*models.BuildingWithStats)
			for _, stats := range statsBuildings {
				buildingList = append(buildingList, &stats.Building)
			}
		} else {
			buildingList = buildings.([]*models.Building)
		}

		for _, building := range buildingList {
			csvData += fmt.Sprintf("%d,%d,%s,%s,%s,%d,%t,%t,%s\n",
				building.ID,
				building.PropertyID,
				building.BuildingName,
				building.BuildingCode,
				building.BuildingType,
				building.TotalFloors,
				building.HasElevator,
				building.ActiveStatus,
				building.CreatedAt.Format("2006-01-02 15:04:05"),
			)
		}
		return []byte(csvData), "text/csv", nil

	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", req.Format)
	}
}

// UpdateBuildingStatus updates building activation/deactivation status
func (s *BuildingService) UpdateBuildingStatus(buildingID, orgID int, req *models.BuildingStatusRequest) (*models.Building, error) {
	// Get existing building for audit logging
	existingBuilding, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing building: %w", err)
	}
	if existingBuilding.OrganizationID != orgID {
		return nil, fmt.Errorf("building not found")
	}

	// If deactivating, check for active units
	if !req.ActiveStatus {
		if err := s.ValidateBuildingDeletion(buildingID); err != nil {
			return nil, fmt.Errorf("cannot deactivate building: %w", err)
		}
	}

	// Prepare updates
	updates := map[string]interface{}{
		"active_status": req.ActiveStatus,
		"updated_at":    time.Now(),
	}

	// Update building status
	if err := s.buildingRepo.Update(buildingID, updates); err != nil {
		return nil, fmt.Errorf("failed to update building status: %w", err)
	}

	// Get updated building
	updatedBuilding, err := s.buildingRepo.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated building: %w", err)
	}

	// Log audit action
	action := "ACTIVATE"
	if !req.ActiveStatus {
		action = "DEACTIVATE"
	}
	_ = s.auditService.LogSystemAction(action, "buildings", &buildingID, existingBuilding, updatedBuilding)

	return updatedBuilding, nil
}
