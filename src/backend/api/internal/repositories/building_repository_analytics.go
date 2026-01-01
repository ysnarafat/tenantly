package repositories

import (
	"database/sql"
	"fmt"

	"github.com/ysnarafat/tenantly/internal/models"
)

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
