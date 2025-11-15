package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ysnarafat/tenantly/internal/models"
)

type PropertyRepository struct {
	db *sql.DB
}

func NewPropertyRepository(db *sql.DB) *PropertyRepository {
	return &PropertyRepository{db: db}
}

// Create creates a new property
func (r *PropertyRepository) Create(property *models.CreatePropertyRequest) (*models.Property, error) {
	query := `
		INSERT INTO properties (property_name, property_code, address, city, postal_code, property_type, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, property_name, property_code, address, city, postal_code, property_type, 
		          total_buildings, metadata, active, created_at, updated_at`

	var result models.Property
	err := r.db.QueryRow(
		query,
		property.PropertyName,
		property.PropertyCode,
		property.Address,
		property.City,
		property.PostalCode,
		property.PropertyType,
		property.Metadata,
	).Scan(
		&result.ID,
		&result.PropertyName,
		&result.PropertyCode,
		&result.Address,
		&result.City,
		&result.PostalCode,
		&result.PropertyType,
		&result.TotalBuildings,
		&result.Metadata,
		&result.Active,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create property: %w", err)
	}

	return &result, nil
}

// GetByID retrieves a property by ID
func (r *PropertyRepository) GetByID(id int) (*models.Property, error) {
	query := `
		SELECT id, property_name, property_code, address, city, postal_code, property_type,
		       total_buildings, metadata, active, created_at, updated_at
		FROM properties
		WHERE id = $1`

	var property models.Property
	err := r.db.QueryRow(query, id).Scan(
		&property.ID,
		&property.PropertyName,
		&property.PropertyCode,
		&property.Address,
		&property.City,
		&property.PostalCode,
		&property.PropertyType,
		&property.TotalBuildings,
		&property.Metadata,
		&property.Active,
		&property.CreatedAt,
		&property.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("property not found")
		}
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	return &property, nil
}

// GetByIDWithStats retrieves a property with aggregated statistics
func (r *PropertyRepository) GetByIDWithStats(id int) (*models.PropertyWithStats, error) {
	query := `
		SELECT 
			p.id, p.property_name, p.property_code, p.address, p.city, p.postal_code, 
			p.property_type, p.total_buildings, p.metadata, p.active, p.created_at, p.updated_at,
			COALESCE(COUNT(DISTINCT b.id), 0) as building_count,
			COALESCE(COUNT(DISTINCT u.id), 0) as unit_count,
			COALESCE(COUNT(DISTINCT CASE WHEN l.active = true THEN u.id END), 0) as occupied_units,
			COALESCE(SUM(CASE WHEN pay.status = 'Paid' THEN pay.amount_paid ELSE 0 END), 0) as total_revenue
		FROM properties p
		LEFT JOIN buildings b ON p.id = b.property_id AND b.active = true
		LEFT JOIN units u ON p.id = u.property_id AND u.active = true
		LEFT JOIN leases l ON u.id = l.unit_id AND l.active = true
		LEFT JOIN payments pay ON u.id = pay.unit_id
		WHERE p.id = $1
		GROUP BY p.id, p.property_name, p.property_code, p.address, p.city, p.postal_code,
		         p.property_type, p.total_buildings, p.metadata, p.active, p.created_at, p.updated_at`

	var property models.PropertyWithStats
	err := r.db.QueryRow(query, id).Scan(
		&property.ID,
		&property.PropertyName,
		&property.PropertyCode,
		&property.Address,
		&property.City,
		&property.PostalCode,
		&property.PropertyType,
		&property.TotalBuildings,
		&property.Metadata,
		&property.Active,
		&property.CreatedAt,
		&property.UpdatedAt,
		&property.BuildingCount,
		&property.UnitCount,
		&property.OccupiedUnits,
		&property.TotalRevenue,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("property not found")
		}
		return nil, fmt.Errorf("failed to get property with stats: %w", err)
	}

	return &property, nil
}

// List retrieves properties with filtering and pagination
func (r *PropertyRepository) List(filters map[string]interface{}, limit, offset int) ([]*models.Property, int, error) {
	// Build WHERE clause
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if propertyType, ok := filters["property_type"]; ok && propertyType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("property_type = $%d", argIndex))
		args = append(args, propertyType)
		argIndex++
	}

	if active, ok := filters["active"]; ok {
		whereConditions = append(whereConditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, active)
		argIndex++
	}

	if search, ok := filters["search"]; ok && search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		whereConditions = append(whereConditions, fmt.Sprintf("(property_name ILIKE $%d OR property_code ILIKE $%d OR address ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM properties WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count properties: %w", err)
	}

	// Get properties with pagination
	query := fmt.Sprintf(`
		SELECT id, property_name, property_code, address, city, postal_code, property_type,
		       total_buildings, metadata, active, created_at, updated_at
		FROM properties
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list properties: %w", err)
	}
	defer rows.Close()

	var properties []*models.Property
	for rows.Next() {
		var property models.Property
		err := rows.Scan(
			&property.ID,
			&property.PropertyName,
			&property.PropertyCode,
			&property.Address,
			&property.City,
			&property.PostalCode,
			&property.PropertyType,
			&property.TotalBuildings,
			&property.Metadata,
			&property.Active,
			&property.CreatedAt,
			&property.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan property: %w", err)
		}
		properties = append(properties, &property)
	}

	return properties, total, nil
}

// Update updates a property
func (r *PropertyRepository) Update(id int, updates *models.UpdatePropertyRequest) (*models.Property, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.PropertyName != nil {
		setParts = append(setParts, fmt.Sprintf("property_name = $%d", argIndex))
		args = append(args, *updates.PropertyName)
		argIndex++
	}

	if updates.Address != nil {
		setParts = append(setParts, fmt.Sprintf("address = $%d", argIndex))
		args = append(args, *updates.Address)
		argIndex++
	}

	if updates.City != nil {
		setParts = append(setParts, fmt.Sprintf("city = $%d", argIndex))
		args = append(args, *updates.City)
		argIndex++
	}

	if updates.PostalCode != nil {
		setParts = append(setParts, fmt.Sprintf("postal_code = $%d", argIndex))
		args = append(args, *updates.PostalCode)
		argIndex++
	}

	if updates.PropertyType != nil {
		setParts = append(setParts, fmt.Sprintf("property_type = $%d", argIndex))
		args = append(args, *updates.PropertyType)
		argIndex++
	}

	if updates.TotalBuildings != nil {
		setParts = append(setParts, fmt.Sprintf("total_buildings = $%d", argIndex))
		args = append(args, *updates.TotalBuildings)
		argIndex++
	}

	if updates.Metadata != nil {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, *updates.Metadata)
		argIndex++
	}

	if updates.Active != nil {
		setParts = append(setParts, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *updates.Active)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(id)
	}

	setParts = append(setParts, "updated_at = NOW() AT TIME ZONE 'UTC'")

	query := fmt.Sprintf(`
		UPDATE properties 
		SET %s
		WHERE id = $%d
		RETURNING id, property_name, property_code, address, city, postal_code, property_type,
		          total_buildings, metadata, active, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	var property models.Property
	err := r.db.QueryRow(query, args...).Scan(
		&property.ID,
		&property.PropertyName,
		&property.PropertyCode,
		&property.Address,
		&property.City,
		&property.PostalCode,
		&property.PropertyType,
		&property.TotalBuildings,
		&property.Metadata,
		&property.Active,
		&property.CreatedAt,
		&property.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("property not found")
		}
		return nil, fmt.Errorf("failed to update property: %w", err)
	}

	return &property, nil
}

// Delete soft deletes a property (marks as inactive)
func (r *PropertyRepository) Delete(id int) error {
	query := `UPDATE properties SET active = false, updated_at = NOW() AT TIME ZONE 'UTC' WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete property: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("property not found")
	}

	return nil
}

// CheckPropertyNameExists checks if a property name already exists
func (r *PropertyRepository) CheckPropertyNameExists(name string, excludeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM properties WHERE property_name = $1 AND id != $2 AND active = true)`
	var exists bool
	err := r.db.QueryRow(query, name, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check property name existence: %w", err)
	}
	return exists, nil
}

// CheckPropertyCodeExists checks if a property code already exists
func (r *PropertyRepository) CheckPropertyCodeExists(code string, excludeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM properties WHERE property_code = $1 AND id != $2 AND active = true)`
	var exists bool
	err := r.db.QueryRow(query, code, excludeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check property code existence: %w", err)
	}
	return exists, nil
}

// HasActiveBuildings checks if property has active buildings
func (r *PropertyRepository) HasActiveBuildings(id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM buildings WHERE property_id = $1 AND active = true)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active buildings: %w", err)
	}
	return exists, nil
}

// HasActiveUnits checks if property has active units
func (r *PropertyRepository) HasActiveUnits(id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM units WHERE property_id = $1 AND active = true)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active units: %w", err)
	}
	return exists, nil
}

// GetBuildingCount returns the count of active buildings for a property
func (r *PropertyRepository) GetBuildingCount(propertyID int) (int, error) {
	query := `SELECT COUNT(*) FROM buildings WHERE property_id = $1 AND active_status = true`
	var count int
	err := r.db.QueryRow(query, propertyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get building count: %w", err)
	}
	return count, nil
}

// GetBuildingSummary returns building summary statistics for a property
func (r *PropertyRepository) GetBuildingSummary(propertyID int) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_buildings,
			COUNT(CASE WHEN building_type = 'Residential' THEN 1 END) as residential_buildings,
			COUNT(CASE WHEN building_type = 'Commercial' THEN 1 END) as commercial_buildings,
			COUNT(CASE WHEN building_type = 'Mixed' THEN 1 END) as mixed_buildings,
			COUNT(CASE WHEN has_elevator = true THEN 1 END) as buildings_with_elevator,
			AVG(total_floors) as average_floors,
			MIN(total_floors) as min_floors,
			MAX(total_floors) as max_floors,
			COUNT(CASE WHEN construction_year IS NOT NULL THEN 1 END) as buildings_with_construction_year,
			AVG(CASE WHEN construction_year IS NOT NULL THEN construction_year END) as average_construction_year
		FROM buildings 
		WHERE property_id = $1 AND active_status = true`

	var totalBuildings, residentialBuildings, commercialBuildings, mixedBuildings, buildingsWithElevator int
	var buildingsWithConstructionYear int
	var averageFloors, minFloors, maxFloors, averageConstructionYear sql.NullFloat64

	err := r.db.QueryRow(query, propertyID).Scan(
		&totalBuildings,
		&residentialBuildings,
		&commercialBuildings,
		&mixedBuildings,
		&buildingsWithElevator,
		&averageFloors,
		&minFloors,
		&maxFloors,
		&buildingsWithConstructionYear,
		&averageConstructionYear,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get building summary: %w", err)
	}

	summary := map[string]interface{}{
		"total_buildings":                  totalBuildings,
		"residential_buildings":            residentialBuildings,
		"commercial_buildings":             commercialBuildings,
		"mixed_buildings":                  mixedBuildings,
		"buildings_with_elevator":          buildingsWithElevator,
		"buildings_with_construction_year": buildingsWithConstructionYear,
	}

	if averageFloors.Valid {
		summary["average_floors"] = averageFloors.Float64
	}
	if minFloors.Valid {
		summary["min_floors"] = int(minFloors.Float64)
	}
	if maxFloors.Valid {
		summary["max_floors"] = int(maxFloors.Float64)
	}
	if averageConstructionYear.Valid {
		summary["average_construction_year"] = int(averageConstructionYear.Float64)
	}

	return summary, nil
}
