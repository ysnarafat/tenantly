package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ysnarafat/tenantly/internal/models"
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
	query := `
		INSERT INTO buildings (property_id, building_name, building_code, building_type, 
		                      total_floors, has_elevator, construction_year, metadata, active_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		building.PropertyID,
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
	query := `
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE id = $1`

	var building models.Building
	err := r.db.QueryRow(query, id).Scan(
		&building.ID,
		&building.PropertyID,
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
	query := `
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE property_id = $1 AND active_status = true
		ORDER BY building_name`

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
	query := `
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE property_id = $1 AND building_code = $2`

	var building models.Building
	err := r.db.QueryRow(query, propertyID, code).Scan(
		&building.ID,
		&building.PropertyID,
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
		case "building_name", "building_type", "total_floors", "has_elevator",
			"construction_year", "metadata", "active_status":
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
		UPDATE buildings 
		SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "), argIndex)

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

	query := `UPDATE buildings SET active_status = false, updated_at = NOW() AT TIME ZONE 'UTC' WHERE id = $1`
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

// GetWithStats retrieves building with analytics and unit counts
func (r *BuildingRepository) GetWithStats(id int) (*models.BuildingWithStats, error) {
	query := `
		SELECT 
			b.id, b.property_id, b.building_name, b.building_code, b.building_type,
			b.total_floors, b.has_elevator, b.construction_year, b.metadata, b.active_status,
			b.created_at, b.updated_at,
			p.property_name,
			COALESCE(COUNT(DISTINCT u.id), 0) as unit_count,
			COALESCE(COUNT(DISTINCT CASE WHEN l.active = true THEN u.id END), 0) as occupied_units,
			COALESCE(SUM(CASE WHEN pay.status = 'Paid' THEN pay.amount_paid ELSE 0 END), 0) as total_revenue
		FROM buildings b
		LEFT JOIN properties p ON b.property_id = p.id
		LEFT JOIN units u ON b.id = u.building_id AND u.active = true
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN payments pay ON u.id = pay.unit_id
		WHERE b.id = $1
		GROUP BY b.id, b.property_id, b.building_name, b.building_code, b.building_type,
		         b.total_floors, b.has_elevator, b.construction_year, b.metadata, b.active_status,
		         b.created_at, b.updated_at, p.property_name`

	var stats models.BuildingWithStats
	err := r.db.QueryRow(query, id).Scan(
		&stats.ID,
		&stats.PropertyID,
		&stats.BuildingName,
		&stats.BuildingCode,
		&stats.BuildingType,
		&stats.TotalFloors,
		&stats.HasElevator,
		&stats.ConstructionYear,
		&stats.Metadata,
		&stats.ActiveStatus,
		&stats.CreatedAt,
		&stats.UpdatedAt,
		&stats.PropertyName,
		&stats.UnitCount,
		&stats.OccupiedUnits,
		&stats.TotalRevenue,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building with stats: %w", err)
	}

	// Calculate occupancy rate
	if stats.UnitCount > 0 {
		stats.OccupancyRate = float64(stats.OccupiedUnits) / float64(stats.UnitCount) * 100
	}

	return &stats, nil
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

	query := `
		INSERT INTO buildings (property_id, building_name, building_code, building_type, 
		                      total_floors, has_elevator, construction_year, metadata, active_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	for _, building := range buildings {
		err := tx.QueryRow(
			query,
			building.PropertyID,
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

	if filters.PropertyID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("property_id = $%d", argIndex))
		args = append(args, *filters.PropertyID)
		argIndex++
	}

	if filters.BuildingType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("building_type = $%d", argIndex))
		args = append(args, *filters.BuildingType)
		argIndex++
	}

	if filters.ActiveStatus != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("active_status = $%d", argIndex))
		args = append(args, *filters.ActiveStatus)
		argIndex++
	}

	if filters.HasElevator != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("has_elevator = $%d", argIndex))
		args = append(args, *filters.HasElevator)
		argIndex++
	}

	if filters.MinFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("total_floors >= $%d", argIndex))
		args = append(args, *filters.MinFloors)
		argIndex++
	}

	if filters.MaxFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("total_floors <= $%d", argIndex))
		args = append(args, *filters.MaxFloors)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE %s
		ORDER BY building_name
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

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

// GetAnalytics retrieves building performance metrics
func (r *BuildingRepository) GetAnalytics(id int) (*models.BuildingAnalytics, error) {
	query := `
		SELECT 
			b.id,
			COALESCE(COUNT(DISTINCT u.id), 0) as unit_count,
			COALESCE(COUNT(DISTINCT CASE WHEN l.active = true THEN u.id END), 0) as occupied_units,
			COALESCE(COUNT(DISTINCT CASE WHEN l.active = false OR l.id IS NULL THEN u.id END), 0) as vacant_units,
			COALESCE(SUM(CASE WHEN pay.status = 'Paid' AND pay.payment_date >= DATE_TRUNC('month', CURRENT_DATE) THEN pay.amount_paid ELSE 0 END), 0) as monthly_revenue,
			COALESCE(AVG(CASE WHEN l.active = true THEN l.monthly_rent END), 0) as average_rent,
			COALESCE(SUM(u.area), 0) as total_area
		FROM buildings b
		LEFT JOIN units u ON b.id = u.building_id AND u.active = true
		LEFT JOIN leases l ON u.id = l.unit_id
		LEFT JOIN payments pay ON u.id = pay.unit_id
		WHERE b.id = $1
		GROUP BY b.id`

	var analytics models.BuildingAnalytics
	var unitCount, occupiedUnits, vacantUnits int
	var monthlyRevenue, averageRent, totalArea float64

	err := r.db.QueryRow(query, id).Scan(
		&analytics.BuildingID,
		&unitCount,
		&occupiedUnits,
		&vacantUnits,
		&monthlyRevenue,
		&averageRent,
		&totalArea,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("building not found")
		}
		return nil, fmt.Errorf("failed to get building analytics: %w", err)
	}

	analytics.UnitCount = unitCount
	analytics.OccupiedUnits = occupiedUnits
	analytics.VacantUnits = vacantUnits
	analytics.MonthlyRevenue = monthlyRevenue
	analytics.AverageRent = averageRent
	analytics.TotalArea = totalArea

	// Calculate occupancy rate
	if unitCount > 0 {
		analytics.OccupancyRate = float64(occupiedUnits) / float64(unitCount) * 100
	}

	return &analytics, nil
}

// CountByProperty counts buildings for a property with optional filters
func (r *BuildingRepository) CountByProperty(propertyID int, filters *models.BuildingSearchFilters) (int, error) {
	whereConditions := []string{"property_id = $1"}
	args := []interface{}{propertyID}
	argIndex := 2

	if filters != nil {
		if filters.BuildingType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("building_type = $%d", argIndex))
			args = append(args, *filters.BuildingType)
			argIndex++
		}

		if filters.ActiveStatus != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("active_status = $%d", argIndex))
			args = append(args, *filters.ActiveStatus)
			argIndex++
		} else {
			// Default to active buildings only
			whereConditions = append(whereConditions, "active_status = true")
		}

		if filters.HasElevator != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("has_elevator = $%d", argIndex))
			args = append(args, *filters.HasElevator)
			argIndex++
		}

		if filters.MinFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("total_floors >= $%d", argIndex))
			args = append(args, *filters.MinFloors)
			argIndex++
		}

		if filters.MaxFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("total_floors <= $%d", argIndex))
			args = append(args, *filters.MaxFloors)
			argIndex++
		}
	} else {
		whereConditions = append(whereConditions, "active_status = true")
	}

	whereClause := strings.Join(whereConditions, " AND ")
	query := fmt.Sprintf("SELECT COUNT(*) FROM buildings WHERE %s", whereClause)

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count buildings: %w", err)
	}

	return count, nil
}

// GetByPropertyWithSorting retrieves buildings for a property with sorting and filtering
func (r *BuildingRepository) GetByPropertyWithSorting(propertyID int, filters *models.BuildingSearchFilters, sortBy, sortOrder string) ([]*models.Building, error) {
	whereConditions := []string{"property_id = $1"}
	args := []interface{}{propertyID}
	argIndex := 2

	if filters != nil {
		if filters.BuildingType != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("building_type = $%d", argIndex))
			args = append(args, *filters.BuildingType)
			argIndex++
		}

		if filters.ActiveStatus != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("active_status = $%d", argIndex))
			args = append(args, *filters.ActiveStatus)
			argIndex++
		} else {
			// Default to active buildings only
			whereConditions = append(whereConditions, "active_status = true")
		}

		if filters.HasElevator != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("has_elevator = $%d", argIndex))
			args = append(args, *filters.HasElevator)
			argIndex++
		}

		if filters.MinFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("total_floors >= $%d", argIndex))
			args = append(args, *filters.MinFloors)
			argIndex++
		}

		if filters.MaxFloors != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("total_floors <= $%d", argIndex))
			args = append(args, *filters.MaxFloors)
			argIndex++
		}
	} else {
		whereConditions = append(whereConditions, "active_status = true")
	}

	// Validate and set sort parameters
	validSortFields := map[string]bool{
		"building_name":     true,
		"building_code":     true,
		"building_type":     true,
		"total_floors":      true,
		"construction_year": true,
		"created_at":        true,
	}

	if !validSortFields[sortBy] {
		sortBy = "building_name" // Default sort field
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // Default sort order
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, whereClause, sortBy, strings.ToUpper(sortOrder), argIndex, argIndex+1)

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
	query := `SELECT EXISTS(SELECT 1 FROM units WHERE building_id = $1 AND active = true)`
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

	// Property filter
	if req.PropertyID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("property_id = $%d", argIndex))
		args = append(args, *req.PropertyID)
		argIndex++
	}

	// Building type filter
	if req.BuildingType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("building_type = $%d", argIndex))
		args = append(args, req.BuildingType)
		argIndex++
	}

	// Active status filter
	if req.ActiveStatus != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("active_status = $%d", argIndex))
		args = append(args, *req.ActiveStatus)
		argIndex++
	}

	// Elevator filter
	if req.HasElevator != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("has_elevator = $%d", argIndex))
		args = append(args, *req.HasElevator)
		argIndex++
	}

	// Floor range filters
	if req.MinFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("total_floors >= $%d", argIndex))
		args = append(args, *req.MinFloors)
		argIndex++
	}

	if req.MaxFloors != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("total_floors <= $%d", argIndex))
		args = append(args, *req.MaxFloors)
		argIndex++
	}

	// Construction year filter
	if req.ConstructionYear != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("construction_year = $%d", argIndex))
		args = append(args, *req.ConstructionYear)
		argIndex++
	}

	// Search term filter (searches in building name and code)
	if req.SearchTerm != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(building_name ILIKE $%d OR building_code ILIKE $%d)", argIndex, argIndex))
		searchPattern := "%" + req.SearchTerm + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	// Metadata query filter (basic JSONB query support)
	if req.MetadataQuery != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("metadata::text ILIKE $%d", argIndex))
		metadataPattern := "%" + req.MetadataQuery + "%"
		args = append(args, metadataPattern)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count first
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM buildings WHERE %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count buildings: %w", err)
	}

	// Validate and set sort parameters
	validSortFields := map[string]bool{
		"building_name":     true,
		"building_code":     true,
		"building_type":     true,
		"total_floors":      true,
		"construction_year": true,
		"created_at":        true,
		"updated_at":        true,
	}

	sortBy := req.SortBy
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	sortOrder := req.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT id, property_id, building_name, building_code, building_type,
		       total_floors, has_elevator, construction_year, metadata, active_status,
		       created_at, updated_at
		FROM buildings
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`, whereClause, sortBy, strings.ToUpper(sortOrder), argIndex, argIndex+1)

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
	countQuery := `SELECT COUNT(*) FROM units WHERE building_id = $1`
	var totalCount int
	err := r.db.QueryRow(countQuery, buildingID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count units: %w", err)
	}

	// Get units with lease information
	query := `
		SELECT 
			u.id, u.unit_number, u.unit_name, u.floor, u.section, u.unit_type, u.monthly_rent, u.active,
			COALESCE(t.full_name, '') as tenant_name,
			COALESCE(l.active, false) as lease_active
		FROM units u
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN tenants t ON l.tenant_id = t.id
		WHERE u.building_id = $1
		ORDER BY u.floor, u.section, u.unit_number
		LIMIT $2 OFFSET $3`

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
