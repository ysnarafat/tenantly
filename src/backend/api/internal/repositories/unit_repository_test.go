package repositories

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
	"github.com/ysnarafat/tenantly/internal/testutil"

	_ "github.com/lib/pq"
)

func setupUnitRepository(t *testing.T) (*UnitRepository, *sqlx.DB, func()) {
	db, cleanup := testutil.SetupTestDB(t)
	repo := NewUnitRepository(db)
	return repo, db, cleanup
}

func TestUnitRepository_BulkCreate(t *testing.T) {
	repo, db, cleanup := setupUnitRepository(t)
	defer cleanup()
	propertyID := testutil.CreateTestProperty(t, db)
	buildingID := testutil.CreateTestBuilding(t, db, propertyID, 1)

	units := []*models.Unit{
		{BuildingID: buildingID, PropertyID: propertyID, OrganizationID: 1, UnitNumber: "BULK-101", UnitType: models.UnitTypeOffice, Active: true},
		{BuildingID: buildingID, PropertyID: propertyID, OrganizationID: 1, UnitNumber: "BULK-102", UnitType: models.UnitTypeShop, Active: true},
	}

	if err := repo.BulkCreate(units); err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	for _, u := range units {
		if u.ID == 0 {
			t.Errorf("Expected unit ID to be set")
		}
	}

	// Empty slice should not error.
	if err := repo.BulkCreate([]*models.Unit{}); err != nil {
		t.Errorf("Expected no error for empty slice but got: %v", err)
	}
}

func TestUnitRepository_BulkCreate_RollsBackOnConflict(t *testing.T) {
	repo, db, cleanup := setupUnitRepository(t)
	defer cleanup()
	propertyID := testutil.CreateTestProperty(t, db)
	buildingID := testutil.CreateTestBuilding(t, db, propertyID, 1)

	// Pre-existing unit that the batch's second item will collide with
	// (units_building_id_unit_number_key).
	if _, err := db.Exec(`
		INSERT INTO units (building_id, property_id, unit_number, unit_type, organization_id, active)
		VALUES ($1, $2, $3, $4, $5, true)`,
		buildingID, propertyID, "CONFLICT-1", models.UnitTypeOffice, 1); err != nil {
		t.Fatalf("Failed to seed conflicting unit: %v", err)
	}

	units := []*models.Unit{
		{BuildingID: buildingID, PropertyID: propertyID, OrganizationID: 1, UnitNumber: "ROLLBACK-1", UnitType: models.UnitTypeOffice, Active: true},
		{BuildingID: buildingID, PropertyID: propertyID, OrganizationID: 1, UnitNumber: "CONFLICT-1", UnitType: models.UnitTypeShop, Active: true},
	}

	if err := repo.BulkCreate(units); err == nil {
		t.Fatal("Expected an error from the unique constraint violation, got nil")
	}

	// The first item in the batch must not have been left behind by a
	// partial commit — the whole batch is one transaction.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM units WHERE unit_number = $1`, "ROLLBACK-1").Scan(&count); err != nil {
		t.Fatalf("Failed to query units: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected the earlier-in-batch unit to be rolled back, but found %d row(s)", count)
	}
}
