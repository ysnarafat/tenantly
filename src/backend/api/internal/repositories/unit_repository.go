package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/models/columns"
)

// UnitRepository implements the UnitRepositoryInterface
type UnitRepository struct {
	db *sql.DB
}

// NewUnitRepository creates a new UnitRepository
func NewUnitRepository(db *sql.DB) *UnitRepository {
	return &UnitRepository{db: db}
}

// Create creates a new unit
func (r *UnitRepository) Create(req *models.CreateUnitRequest, organizationID int) (*models.Unit, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			%s, %s, %s, %s,
			%s, %s, %s, %s, %s, %s
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true)
		RETURNING %s, %s, %s`,
		columns.UnitTable,
		columns.UnitBuildingID, columns.UnitPropertyID, columns.UnitNumber, columns.UnitName,
		columns.UnitFloor, columns.UnitSection, columns.UnitType, columns.UnitMetadata, columns.UnitOrganizationID, columns.UnitActive,
		columns.UnitID, columns.UnitCreatedAt, columns.UnitUpdatedAt)

	unit := &models.Unit{
		BuildingID:     req.BuildingID,
		PropertyID:     req.PropertyID,
		UnitNumber:     req.UnitNumber,
		UnitName:       req.UnitName,
		Floor:          req.Floor,
		Section:        req.Section,
		UnitType:       req.UnitType,
		Metadata:       req.Metadata,
		OrganizationID: organizationID,
		Active:         true,
	}

	err := r.db.QueryRow(
		query,
		unit.BuildingID,
		unit.PropertyID,
		unit.UnitNumber,
		unit.UnitName,
		unit.Floor,
		unit.Section,
		unit.UnitType,
		unit.Metadata,
		unit.OrganizationID,
	).Scan(&unit.ID, &unit.CreatedAt, &unit.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}

	return unit, nil
}

// GetByID retrieves a unit by ID
func (r *UnitRepository) GetByID(id int) (*models.Unit, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = $1 AND %s = true`,
		columns.UnitAllColumns(),
		columns.UnitTable,
		columns.UnitID, columns.UnitActive)

	unit := &models.Unit{}
	err := r.db.QueryRow(query, id).Scan(
		&unit.ID,
		&unit.BuildingID,
		&unit.PropertyID,
		&unit.OrganizationID,
		&unit.UnitNumber,
		&unit.UnitName,
		&unit.Floor,
		&unit.Section,
		&unit.UnitType,
		&unit.Metadata,
		&unit.Active,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found")
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	return unit, nil
}

// GetByIDWithDetails retrieves a unit by ID with property and building details
func (r *UnitRepository) GetByIDWithDetails(id int) (*models.UnitWithDetails, error) {
	query := `
		SELECT u.id, u.building_id, u.property_id, u.unit_number, u.unit_name,
			   u.floor, u.section, u.unit_type, u.metadata, u.active,
			   u.created_at, u.updated_at,
			   p.property_name as property_name,
			   b.building_name, b.building_code,
			   COALESCE(t.name, '') as tenant_name,
			   CASE WHEN l.id IS NOT NULL THEN true ELSE false END as lease_active
		FROM units u
		JOIN properties p ON u.property_id = p.id
		JOIN buildings b ON u.building_id = b.id
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN tenants t ON l.tenant_id = t.id
		WHERE u.id = $1 AND u.active = true`

	unit := &models.UnitWithDetails{}
	err := r.db.QueryRow(query, id).Scan(
		&unit.ID,
		&unit.BuildingID,
		&unit.PropertyID,
		&unit.UnitNumber,
		&unit.UnitName,
		&unit.Floor,
		&unit.Section,
		&unit.UnitType,
		&unit.Metadata,
		&unit.Active,
		&unit.CreatedAt,
		&unit.UpdatedAt,
		&unit.PropertyName,
		&unit.BuildingName,
		&unit.BuildingCode,
		&unit.TenantName,
		&unit.LeaseActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found")
		}
		return nil, fmt.Errorf("failed to get unit with details: %w", err)
	}

	return unit, nil
}

// Update updates a unit
func (r *UnitRepository) Update(id int, req *models.UpdateUnitRequest) (*models.Unit, error) {
	// Build dynamic update query
	query := "UPDATE units SET updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.UnitName != nil {
		query += fmt.Sprintf(", unit_name = $%d", argIdx)
		args = append(args, *req.UnitName)
		argIdx++
	}
	if req.Floor != nil {
		query += fmt.Sprintf(", floor = $%d", argIdx)
		args = append(args, *req.Floor)
		argIdx++
	}
	if req.Section != nil {
		query += fmt.Sprintf(", section = $%d", argIdx)
		args = append(args, *req.Section)
		argIdx++
	}
	if req.UnitType != nil {
		query += fmt.Sprintf(", unit_type = $%d", argIdx)
		args = append(args, *req.UnitType)
		argIdx++
	}
	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		query += fmt.Sprintf(", metadata = $%d", argIdx)
		args = append(args, metadataJSON)
		argIdx++
	}
	if req.Active != nil {
		query += fmt.Sprintf(", active = $%d", argIdx)
		args = append(args, *req.Active)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, building_id, property_id, unit_number, unit_name, floor, section, unit_type, metadata, active, created_at, updated_at", argIdx)
	args = append(args, id)

	unit := &models.Unit{}
	err := r.db.QueryRow(query, args...).Scan(
		&unit.ID,
		&unit.BuildingID,
		&unit.PropertyID,
		&unit.UnitNumber,
		&unit.UnitName,
		&unit.Floor,
		&unit.Section,
		&unit.UnitType,
		&unit.Metadata,
		&unit.Active,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unit not found")
		}
		return nil, fmt.Errorf("failed to update unit: %w", err)
	}

	return unit, nil
}

// Delete soft deletes a unit
func (r *UnitRepository) Delete(id int) error {
	query := fmt.Sprintf("UPDATE %s SET %s = false, %s = NOW() WHERE %s = $1",
		columns.UnitTable, columns.UnitActive, columns.UnitUpdatedAt, columns.UnitID)
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete unit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("unit not found")
	}

	return nil
}

// CheckUnitNumberExists checks if a unit number exists in a building
func (r *UnitRepository) CheckUnitNumberExists(buildingID int, unitNumber string, excludeID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM units 
			WHERE building_id = $1 AND unit_number = $2 AND id != $3 AND active = true
		)`

	var exists bool
	err := r.db.QueryRow(query, buildingID, unitNumber, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check unit number existence: %w", err)
	}

	return exists, nil
}

// HasActiveLeases checks if a unit has any active leases
func (r *UnitRepository) HasActiveLeases(unitID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM leases 
			WHERE unit_id = $1 AND status = 'Active'
		)`

	var exists bool
	err := r.db.QueryRow(query, unitID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active leases: %w", err)
	}

	return exists, nil
}

// GetByBuildingWithDetails retrieves units for a building with details.
// Pass orgID=0 to skip organization filtering (internal use only).
func (r *UnitRepository) GetByBuildingWithDetails(buildingID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	orgFilter := ""
	if orgID > 0 {
		orgFilter = fmt.Sprintf(" AND b.organization_id = %d", orgID)
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM units u JOIN buildings b ON u.building_id = b.id WHERE u.building_id = $1 AND u.active = true%s`, orgFilter)
	var total int
	err := r.db.QueryRow(countQuery, buildingID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get unit count: %w", err)
	}

	// Get units
	query := fmt.Sprintf(`
		SELECT u.id, u.building_id, u.property_id, u.unit_number, u.unit_name,
			   u.floor, u.section, u.unit_type, u.metadata, u.active,
			   u.created_at, u.updated_at,
			   p.property_name as property_name,
			   b.building_name, b.building_code,
			   COALESCE(t.name, '') as tenant_name,
			   CASE WHEN l.id IS NOT NULL THEN true ELSE false END as lease_active
		FROM units u
		JOIN properties p ON u.property_id = p.id
		JOIN buildings b ON u.building_id = b.id
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN tenants t ON l.tenant_id = t.id
		WHERE u.building_id = $1 AND u.active = true%s
		ORDER BY u.floor, u.unit_number
		LIMIT $2 OFFSET $3`, orgFilter)

	rows, err := r.db.Query(query, buildingID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get units: %w", err)
	}
	defer rows.Close()

	var units []*models.UnitWithDetails
	for rows.Next() {
		unit := &models.UnitWithDetails{}
		err := rows.Scan(
			&unit.ID,
			&unit.BuildingID,
			&unit.PropertyID,
			&unit.UnitNumber,
			&unit.UnitName,
			&unit.Floor,
			&unit.Section,
			&unit.UnitType,
			&unit.Metadata,
			&unit.Active,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			&unit.PropertyName,
			&unit.BuildingName,
			&unit.BuildingCode,
			&unit.TenantName,
			&unit.LeaseActive,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, unit)
	}

	return units, total, nil
}

// GetByPropertyWithDetails retrieves units for a property with details
func (r *UnitRepository) GetByPropertyWithDetails(propertyID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM units u JOIN properties p ON u.property_id = p.id WHERE u.property_id = $1 AND u.active = true AND p.organization_id = $2`
	var total int
	err := r.db.QueryRow(countQuery, propertyID, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get unit count: %w", err)
	}

	// Get units
	query := `
		SELECT u.id, u.building_id, u.property_id, u.unit_number, u.unit_name,
			   u.floor, u.section, u.unit_type, u.metadata, u.active,
			   u.created_at, u.updated_at,
			   p.property_name as property_name,
			   b.building_name, b.building_code,
			   COALESCE(t.name, '') as tenant_name,
			   CASE WHEN l.id IS NOT NULL THEN true ELSE false END as lease_active
		FROM units u
		JOIN properties p ON u.property_id = p.id
		JOIN buildings b ON u.building_id = b.id
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN tenants t ON l.tenant_id = t.id
		WHERE u.property_id = $1 AND u.active = true AND p.organization_id = $2
		ORDER BY b.building_name, u.floor, u.unit_number
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(query, propertyID, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get units: %w", err)
	}
	defer rows.Close()

	var units []*models.UnitWithDetails
	for rows.Next() {
		unit := &models.UnitWithDetails{}
		err := rows.Scan(
			&unit.ID,
			&unit.BuildingID,
			&unit.PropertyID,
			&unit.UnitNumber,
			&unit.UnitName,
			&unit.Floor,
			&unit.Section,
			&unit.UnitType,
			&unit.Metadata,
			&unit.Active,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			&unit.PropertyName,
			&unit.BuildingName,
			&unit.BuildingCode,
			&unit.TenantName,
			&unit.LeaseActive,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, unit)
	}

	return units, total, nil
}

// GetBuildingOccupancyStats retrieves occupancy statistics for a building
func (r *UnitRepository) GetBuildingOccupancyStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	// Placeholder implementation - to be fully implemented with analytics module
	return map[string]interface{}{
		"building_id":    buildingID,
		"occupancy_rate": 0.0,
	}, nil
}

// GetPropertyOccupancyStats retrieves occupancy statistics for a property
func (r *UnitRepository) GetPropertyOccupancyStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	// Placeholder implementation - to be fully implemented with analytics module
	return map[string]interface{}{
		"property_id":    propertyID,
		"occupancy_rate": 0.0,
	}, nil
}

// GetSystemOccupancyStats retrieves system-wide occupancy statistics
func (r *UnitRepository) GetSystemOccupancyStats(startDate, endDate time.Time) (interface{}, error) {
	// Placeholder implementation - to be fully implemented with analytics module
	return map[string]interface{}{
		"occupancy_rate": 0.0,
	}, nil
}

// GetBuildingUnitTypeDistribution retrieves unit type distribution for a building
func (r *UnitRepository) GetBuildingUnitTypeDistribution(buildingID int) (interface{}, error) {
	query := `
		SELECT unit_type, COUNT(*) as count
		FROM units
		WHERE building_id = $1 AND active = true
		GROUP BY unit_type`

	rows, err := r.db.Query(query, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit type distribution: %w", err)
	}
	defer rows.Close()

	distribution := make(map[string]int)
	for rows.Next() {
		var unitType string
		var count int
		if err := rows.Scan(&unitType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		distribution[unitType] = count
	}

	return distribution, nil
}

// GetByOrganizationID retrieves all units for a specific organization
func (r *UnitRepository) GetByOrganizationID(orgID int) ([]*models.Unit, error) {
	query := `
		SELECT id, property_id, building_id, unit_number, unit_name, floor, section,
		       unit_type, metadata, active, created_at, updated_at
		FROM units
		WHERE organization_id = $1 AND active = true
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units by organization: %w", err)
	}
	defer rows.Close()

	var units []*models.Unit
	for rows.Next() {
		var unit models.Unit
		var metadataStr string
		err := rows.Scan(
			&unit.ID,
			&unit.PropertyID,
			&unit.BuildingID,
			&unit.UnitNumber,
			&unit.UnitName,
			&unit.Floor,
			&unit.Section,
			&unit.UnitType,
			&metadataStr,
			&unit.Active,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, &unit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating units: %w", err)
	}

	return units, nil
}
