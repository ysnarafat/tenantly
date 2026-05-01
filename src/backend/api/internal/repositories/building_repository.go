package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/models/columns"
)

// BuildingRepository implements the BuildingRepositoryInterface
type BuildingRepository struct {
	db *sql.DB
}

// NewBuildingRepository creates a new building repository instance
func NewBuildingRepository(db *sql.DB) *BuildingRepository {
	return &BuildingRepository{db: db}
}

// Create creates a new building with validation and constraint handling
func (r *BuildingRepository) Create(building *models.Building) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s, %s, %s,
		                      %s, %s, %s, %s, %s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING %s, %s, %s`,
		columns.BuildingTable,
		columns.BuildingPropertyID, columns.BuildingOrganizationID, columns.BuildingName, columns.BuildingCode, columns.BuildingType,
		columns.BuildingTotalFloors, columns.BuildingHasElevator, columns.BuildingConstructionYear,
		columns.BuildingMetadata, columns.BuildingActiveStatus,
		columns.BuildingID, columns.BuildingCreatedAt, columns.BuildingUpdatedAt)

	err := r.db.QueryRow(
		query,
		building.PropertyID,
		building.OrganizationID,
		building.BuildingName,
		building.BuildingCode,
		building.BuildingType,
		building.TotalFloors,
		building.HasElevator,
		building.ConstructionYear,
		building.Metadata,
		building.ActiveStatus,
	).Scan(&building.ID, &building.CreatedAt, &building.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create building: %w", err)
	}

	return nil
}

// GetByID retrieves a building by its ID
func (r *BuildingRepository) GetByID(id int) (*models.Building, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		columns.BuildingID)

	var building models.Building
	err := r.db.QueryRow(query, id).Scan(
		&building.ID,
		&building.PropertyID,
		&building.OrganizationID,
		&building.BuildingName,
		&building.BuildingCode,
		&building.BuildingType,
		&building.TotalFloors,
		&building.HasElevator,
		&building.ConstructionYear,
		&building.Metadata,
		&building.ActiveStatus,
		&building.CreatedAt,
		&building.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building: %w", err)
	}

	return &building, nil
}

// GetByPropertyID retrieves all buildings for a specific property
func (r *BuildingRepository) GetByPropertyID(propertyID int) ([]*models.Building, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1 AND %s = true
		ORDER BY %s`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		columns.BuildingPropertyID, columns.BuildingActiveStatus,
		columns.BuildingName)

	rows, err := r.db.Query(query, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings by property: %w", err)
	}
	defer rows.Close()

	var buildings []*models.Building
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.PropertyID,
			&building.OrganizationID,
			&building.BuildingName,
			&building.BuildingCode,
			&building.BuildingType,
			&building.TotalFloors,
			&building.HasElevator,
			&building.ConstructionYear,
			&building.Metadata,
			&building.ActiveStatus,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan building: %w", err)
		}
		buildings = append(buildings, &building)
	}

	return buildings, nil
}

// GetByPropertyAndCode retrieves a building by property ID and building code
func (r *BuildingRepository) GetByPropertyAndCode(propertyID int, code string) (*models.Building, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1 AND %s = $2`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		columns.BuildingPropertyID, columns.BuildingCode)

	var building models.Building
	err := r.db.QueryRow(query, propertyID, code).Scan(
		&building.ID,
		&building.PropertyID,
		&building.OrganizationID,
		&building.BuildingName,
		&building.BuildingCode,
		&building.BuildingType,
		&building.TotalFloors,
		&building.HasElevator,
		&building.ConstructionYear,
		&building.Metadata,
		&building.ActiveStatus,
		&building.CreatedAt,
		&building.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building by property and code: %w", err)
	}

	return &building, nil
}

// Update updates a building with partial updates and metadata handling
func (r *BuildingRepository) Update(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case columns.BuildingName, columns.BuildingType, columns.BuildingTotalFloors, columns.BuildingHasElevator,
			columns.BuildingConstructionYear, columns.BuildingMetadata, columns.BuildingActiveStatus:
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW() AT TIME ZONE 'UTC'")

	query := fmt.Sprintf(`
		UPDATE %s 
		SET %s
		WHERE %s = $%d`,
		columns.BuildingTable,
		strings.Join(setParts, ", "),
		columns.BuildingID, argIndex)

	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update building: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("building not found")
	}

	return nil
}

// SoftDelete performs soft delete maintaining referential integrity
func (r *BuildingRepository) SoftDelete(id int) error {
	// Check if building has active units
	hasActiveUnits, err := r.hasActiveUnits(id)
	if err != nil {
		return fmt.Errorf("failed to check active units: %w", err)
	}

	if hasActiveUnits {
		return fmt.Errorf("cannot delete building with active units")
	}

	query := fmt.Sprintf(`UPDATE %s SET %s = false, %s = NOW() AT TIME ZONE 'UTC' WHERE %s = $1`,
		columns.BuildingTable, columns.BuildingActiveStatus, columns.BuildingUpdatedAt, columns.BuildingID)
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete building: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("building not found")
	}

	return nil
}

// BulkCreate creates multiple buildings efficiently
func (r *BuildingRepository) BulkCreate(buildings []*models.Building) error {
	if len(buildings) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s, %s, %s,
		                      %s, %s, %s, %s, %s)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING %s, %s, %s`,
		columns.BuildingTable,
		columns.BuildingPropertyID, columns.BuildingOrganizationID, columns.BuildingName, columns.BuildingCode, columns.BuildingType,
		columns.BuildingTotalFloors, columns.BuildingHasElevator, columns.BuildingConstructionYear,
		columns.BuildingMetadata, columns.BuildingActiveStatus,
		columns.BuildingID, columns.BuildingCreatedAt, columns.BuildingUpdatedAt)

	for _, building := range buildings {
		err := tx.QueryRow(
			query,
			building.PropertyID,
			building.OrganizationID,
			building.BuildingName,
			building.BuildingCode,
			building.BuildingType,
			building.TotalFloors,
			building.HasElevator,
			building.ConstructionYear,
			building.Metadata,
			building.ActiveStatus,
		).Scan(&building.ID, &building.CreatedAt, &building.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to create building %s: %w", building.BuildingName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Search searches buildings with filtering by property, type, and metadata attributes
func (r *BuildingRepository) Search(filters *models.BuildingSearchFilters) ([]*models.Building, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filters.OrganizationID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingOrganizationID, argIndex))
		args = append(args, *filters.OrganizationID)
		argIndex++
	}

	if filters.PropertyID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingPropertyID, argIndex))
		args = append(args, *filters.PropertyID)
		argIndex++
	}

	if filters.BuildingType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingType, argIndex))
		args = append(args, *filters.BuildingType)
		argIndex++
	}

	if filters.ActiveStatus != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingActiveStatus, argIndex))
		args = append(args, *filters.ActiveStatus)
		argIndex++
	}

	if filters.HasElevator != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingHasElevator, argIndex))
		args = append(args, *filters.HasElevator)
		argIndex++
	}

	if filters.MinFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s >= $%d", columns.BuildingTotalFloors, argIndex))
		args = append(args, *filters.MinFloors)
		argIndex++
	}

	if filters.MaxFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s <= $%d", columns.BuildingTotalFloors, argIndex))
		args = append(args, *filters.MaxFloors)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		whereClause,
		columns.BuildingName,
		argIndex, argIndex+1)

	limit := filters.Limit
	if limit <= 0 {
		limit = 50 // Default limit
	}

	args = append(args, limit, filters.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search buildings: %w", err)
	}
	defer rows.Close()

	var buildings []*models.Building
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.PropertyID,
			&building.OrganizationID,
			&building.BuildingName,
			&building.BuildingCode,
			&building.BuildingType,
			&building.TotalFloors,
			&building.HasElevator,
			&building.ConstructionYear,
			&building.Metadata,
			&building.ActiveStatus,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan building: %w", err)
		}
		buildings = append(buildings, &building)
	}

	return buildings, nil
}

// CountByProperty counts buildings for a property with optional filters
func (r *BuildingRepository) CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error) {
	whereConditions := []string{fmt.Sprintf("%s = $1", columns.BuildingPropertyID)}
	args := []interface{}{propertyID}
	argIndex := 2

	if filters != nil {
		if filters.BuildingType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("building_type = $%d", argIndex))
			args = append(args, *filters.BuildingType)
			argIndex++
		}

		if filters.ActiveStatus != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingActiveStatus, argIndex))
			args = append(args, *filters.ActiveStatus)
			argIndex++
		} else {
			// Default to active buildings only
			whereConditions = append(whereConditions, fmt.Sprintf("%s = true", columns.BuildingActiveStatus))
		}

		if filters.HasElevator != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingHasElevator, argIndex))
			args = append(args, *filters.HasElevator)
			argIndex++
		}

		if filters.MinFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s >= $%d", columns.BuildingTotalFloors, argIndex))
			args = append(args, *filters.MinFloors)
			argIndex++
		}

		if filters.MaxFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s <= $%d", columns.BuildingTotalFloors, argIndex))
			args = append(args, *filters.MaxFloors)
			argIndex++
		}
	} else {
		whereConditions = append(whereConditions, "active_status = true")
	}

	whereClause := strings.Join(whereConditions, " AND ")
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", columns.BuildingTable, whereClause)

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count buildings: %w", err)
	}

	return count, nil
}

// GetByPropertyWithSorting retrieves buildings for a property with sorting and filtering
func (r *BuildingRepository) GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error) {
	whereConditions := []string{fmt.Sprintf("%s = $1", columns.BuildingPropertyID)}
	args := []interface{}{propertyID}
	argIndex := 2

	if filters != nil {
		if filters.BuildingType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingType, argIndex))
			args = append(args, *filters.BuildingType)
			argIndex++
		}

		if filters.ActiveStatus != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingActiveStatus, argIndex))
			args = append(args, *filters.ActiveStatus)
			argIndex++
		} else {
			// Default to active buildings only
			whereConditions = append(whereConditions, fmt.Sprintf("%s = true", columns.BuildingActiveStatus))
		}

		if filters.HasElevator != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingHasElevator, argIndex))
			args = append(args, *filters.HasElevator)
			argIndex++
		}

		if filters.MinFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s >= $%d", columns.BuildingTotalFloors, argIndex))
			args = append(args, *filters.MinFloors)
			argIndex++
		}

		if filters.MaxFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("%s <= $%d", columns.BuildingTotalFloors, argIndex))
			args = append(args, *filters.MaxFloors)
			argIndex++
		}
	} else {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = true", columns.BuildingActiveStatus))
	}

	// Validate and set sort parameters
	validSortFields := map[string]bool{
		columns.BuildingName:             true,
		columns.BuildingCode:             true,
		columns.BuildingType:             true,
		columns.BuildingTotalFloors:      true,
		columns.BuildingConstructionYear: true,
		columns.BuildingCreatedAt:        true,
	}

	if !validSortFields[sortBy] {
		sortBy = columns.BuildingName // Default sort field
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // Default sort order
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		whereClause,
		sortBy, strings.ToUpper(sortOrder), argIndex, argIndex+1)

	limit := 20 // Default limit
	offset := 0 // Default offset

	if filters != nil {
		if filters.Limit > 0 {
			limit = filters.Limit
		}
		offset = filters.Offset
	}

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings with sorting: %w", err)
	}
	defer rows.Close()

	var buildings []*models.Building
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.PropertyID,
			&building.OrganizationID,
			&building.BuildingName,
			&building.BuildingCode,
			&building.BuildingType,
			&building.TotalFloors,
			&building.HasElevator,
			&building.ConstructionYear,
			&building.Metadata,
			&building.ActiveStatus,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan building: %w", err)
		}
		buildings = append(buildings, &building)
	}

	return buildings, nil
}

// hasActiveUnits checks if building has active units (helper method)
func (r *BuildingRepository) hasActiveUnits(buildingID int) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1 AND %s = true)`,
		columns.UnitTable, columns.UnitBuildingID, columns.UnitActive)
	var exists bool
	err := r.db.QueryRow(query, buildingID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active units: %w", err)
	}
	return exists, nil
}

// AdvancedSearch performs advanced building search with metadata queries and enhanced filtering
func (r *BuildingRepository) AdvancedSearch(req *models.BuildingSearchRequest) ([]*models.Building, int, error) {
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	// Organization filter (mandatory for org-scoped callers)
	if req.OrganizationID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingOrganizationID, argIndex))
		args = append(args, *req.OrganizationID)
		argIndex++
	}

	// Property filter
	if req.PropertyID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingPropertyID, argIndex))
		args = append(args, *req.PropertyID)
		argIndex++
	}

	// Building type filter
	if req.BuildingType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingType, argIndex))
		args = append(args, req.BuildingType)
		argIndex++
	}

	// Active status filter
	if req.ActiveStatus != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingActiveStatus, argIndex))
		args = append(args, *req.ActiveStatus)
		argIndex++
	}

	// Elevator filter
	if req.HasElevator != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingHasElevator, argIndex))
		args = append(args, *req.HasElevator)
		argIndex++
	}

	// Floor range filters
	if req.MinFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s >= $%d", columns.BuildingTotalFloors, argIndex))
		args = append(args, *req.MinFloors)
		argIndex++
	}

	if req.MaxFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s <= $%d", columns.BuildingTotalFloors, argIndex))
		args = append(args, *req.MaxFloors)
		argIndex++
	}

	// Construction year filter
	if req.ConstructionYear != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", columns.BuildingConstructionYear, argIndex))
		args = append(args, *req.ConstructionYear)
		argIndex++
	}

	// Search term filter (searches in building name and code)
	if req.SearchTerm != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(%s ILIKE $%d OR %s ILIKE $%d)", columns.BuildingName, argIndex, columns.BuildingCode, argIndex))
		searchPattern := "%" + req.SearchTerm + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	// Metadata query filter (basic JSONB query support)
	if req.MetadataQuery != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("%s::text ILIKE $%d", columns.BuildingMetadata, argIndex))
		metadataPattern := "%" + req.MetadataQuery + "%"
		args = append(args, metadataPattern)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", columns.BuildingTable, whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count buildings: %w", err)
	}

	// Validate and set sort parameters
	validSortFields := map[string]bool{
		columns.BuildingName:             true,
		columns.BuildingCode:             true,
		columns.BuildingType:             true,
		columns.BuildingTotalFloors:      true,
		columns.BuildingConstructionYear: true,
		columns.BuildingCreatedAt:        true,
		columns.BuildingUpdatedAt:        true,
	}

	sortBy := req.SortBy
	if !validSortFields[sortBy] {
		sortBy = columns.BuildingCreatedAt
	}

	sortOrder := req.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		columns.BuildingAllColumns(),
		columns.BuildingTable,
		whereClause, sortBy, strings.ToUpper(sortOrder), argIndex, argIndex+1)

	args = append(args, req.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search buildings: %w", err)
	}
	defer rows.Close()

	var buildings []*models.Building
	for rows.Next() {
		var building models.Building
		err := rows.Scan(
			&building.ID,
			&building.PropertyID,
			&building.OrganizationID,
			&building.BuildingName,
			&building.BuildingCode,
			&building.BuildingType,
			&building.TotalFloors,
			&building.HasElevator,
			&building.ConstructionYear,
			&building.Metadata,
			&building.ActiveStatus,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan building: %w", err)
		}
		buildings = append(buildings, &building)
	}

	return buildings, totalCount, nil
}

// GetBuildingUnits retrieves units for a specific building with pagination
func (r *BuildingRepository) GetBuildingUnits(buildingID int, offset, limit int) ([]*models.BuildingUnitSummary, int, error) {
	// Get total count first
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = $1`, columns.UnitTable, columns.UnitBuildingID)
	var totalCount int
	err := r.db.QueryRow(countQuery, buildingID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count units: %w", err)
	}

	// Get units with lease information
	query := fmt.Sprintf(`
		SELECT 
			u.%s, u.%s, u.%s, u.%s, u.%s, u.%s, u.%s, u.%s,
			COALESCE(t.name, '') as tenant_name,
			COALESCE(l.active, false) as lease_active
		FROM %s u
		LEFT JOIN leases l ON u.%s = l.unit_id AND l.active = true
		LEFT JOIN tenants t ON l.tenant_id = t.id
		WHERE u.%s = $1
		ORDER BY u.%s, u.%s, u.%s
		LIMIT $2 OFFSET $3`,
		columns.UnitID, columns.UnitNumber, columns.UnitName, columns.UnitFloor, columns.UnitSection,
		columns.UnitType, columns.UnitMonthlyRent, columns.UnitActive,
		columns.UnitTable,
		columns.UnitID,
		columns.UnitBuildingID,
		columns.UnitFloor, columns.UnitSection, columns.UnitNumber)

	rows, err := r.db.Query(query, buildingID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get building units: %w", err)
	}
	defer rows.Close()

	var units []*models.BuildingUnitSummary
	for rows.Next() {
		var unit models.BuildingUnitSummary
		err := rows.Scan(
			&unit.UnitID,
			&unit.UnitNumber,
			&unit.UnitName,
			&unit.Floor,
			&unit.Section,
			&unit.UnitType,
			&unit.MonthlyRent,
			&unit.Active,
			&unit.TenantName,
			&unit.LeaseActive,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, &unit)
	}

	return units, totalCount, nil
}

// GetByOrganizationID retrieves all buildings for a specific organization
func (r *BuildingRepository) GetByOrganizationID(orgID int) ([]*models.Building, error) {
	query := `
		SELECT id, property_id, building_code, building_name, building_type, total_floors,
		       has_elevator, construction_year, metadata, active_status, created_at, updated_at
		FROM buildings
		WHERE organization_id = $1 AND active_status = true
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings by organization: %w", err)
	}
	defer rows.Close()

	var buildings []*models.Building
	for rows.Next() {
		var building models.Building
		var metadata sql.NullString
		err := rows.Scan(
			&building.ID,
			&building.PropertyID,
			&building.BuildingCode,
			&building.BuildingName,
			&building.BuildingType,
			&building.TotalFloors,
			&building.HasElevator,
			&building.ConstructionYear,
			&metadata,
			&building.ActiveStatus,
			&building.CreatedAt,
			&building.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan building: %w", err)
		}

		if metadata.Valid {
			if err := json.Unmarshal([]byte(metadata.String), &building.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		buildings = append(buildings, &building)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating buildings: %w", err)
	}

	return buildings, nil
}
