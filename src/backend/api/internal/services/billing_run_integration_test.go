package services

import (
	"testing"

	"github.com/ysnarafat/tenantly/internal/repositories"
	"github.com/ysnarafat/tenantly/internal/testutil"
)

func TestBillingRunIsScopedToOneBuildingAndPrintsEveryActiveUnit(t *testing.T) {
	db, cleanup := testutil.SetupTestDBNamed(t, servicesTestDB)
	defer cleanup()

	orgID := testutil.CreateTestOrganization(t, db)
	propertyID := testutil.CreateTestProperty(t, db)
	if _, err := db.Exec(`UPDATE properties SET organization_id = $1 WHERE id = $2`, orgID, propertyID); err != nil {
		t.Fatalf("scope test property to organization: %v", err)
	}

	buildingID := testutil.CreateTestBuilding(t, db, propertyID, orgID)
	unitOneID := testutil.CreateTestUnit(t, db, buildingID, orgID)
	unitTwoID := testutil.CreateTestUnit(t, db, buildingID, orgID)
	otherBuildingID := testutil.CreateTestBuilding(t, db, propertyID, orgID)
	testutil.CreateTestUnit(t, db, otherBuildingID, orgID)

	repo := repositories.NewBillingRunRepository(db)
	userID := testutil.CreateTestUser(t, db)
	run, err := repo.Create(buildingID, orgID, 10, 2026, userID)
	if err != nil {
		t.Fatalf("create building electricity register: %v", err)
	}
	if run.BuildingID != buildingID || len(run.Items) != 2 {
		t.Fatalf("register should include only the selected building's two units: %+v", run)
	}
	if run.Items[0].ElectricityAmount != nil || run.Items[1].ElectricityAmount != nil {
		t.Fatal("new electricity amounts should be blank until entered")
	}

	var firstItemID, secondItemID int
	for _, item := range run.Items {
		switch item.UnitID {
		case unitOneID:
			firstItemID = item.ID
		case unitTwoID:
			secondItemID = item.ID
		}
	}
	if firstItemID == 0 || secondItemID == 0 {
		t.Fatalf("register omitted a unit: %+v", run.Items)
	}

	run, err = repo.UpdateElectricity(run.ID, firstItemID, orgID, "143833", 3460)
	if err != nil {
		t.Fatalf("enter first unit's bill: %v", err)
	}
	run, err = repo.UpdateElectricity(run.ID, secondItemID, orgID, "143834", 0)
	if err != nil {
		t.Fatalf("enter second unit's zero bill: %v", err)
	}
	amounts := make(map[int]float64, len(run.Items))
	for _, item := range run.Items {
		if item.ElectricityAmount != nil {
			amounts[item.UnitID] = *item.ElectricityAmount
		}
	}
	if amounts[unitOneID] != 3460 || amounts[unitTwoID] != 0 {
		t.Fatalf("saved amounts were not returned: %+v", amounts)
	}
	if _, err := repo.Finalize(run.ID, orgID); err != nil {
		t.Fatalf("finalize complete electricity register: %v", err)
	}
	if _, err := repo.UpdateElectricity(run.ID, firstItemID, orgID, "143833", 4000); err == nil {
		t.Fatal("expected finalized electricity register to be immutable")
	}

	nextRun, err := repo.Create(buildingID, orgID, 11, 2026, userID)
	if err != nil {
		t.Fatalf("create next month's register: %v", err)
	}
	if len(nextRun.Items) != 2 {
		t.Fatalf("next month's register should include the selected building's units: %+v", nextRun.Items)
	}
	meterNumbers := make(map[int]string, len(nextRun.Items))
	for _, item := range nextRun.Items {
		meterNumbers[item.UnitID] = item.MeterNumber
	}
	if meterNumbers[unitOneID] != "143833" || meterNumbers[unitTwoID] != "143834" {
		t.Fatalf("meter numbers should carry over from the previous period: %+v", meterNumbers)
	}

	otherRun, err := repo.Create(otherBuildingID, orgID, 10, 2026, userID)
	if err != nil {
		t.Fatalf("create same-period register for another building: %v", err)
	}
	if otherRun.BuildingID != otherBuildingID || len(otherRun.Items) != 1 {
		t.Fatalf("other building's register has unexpected units: %+v", otherRun)
	}

	var generatedPayments int
	if err := db.Get(&generatedPayments, `SELECT COUNT(*) FROM payments WHERE unit_id IN ($1, $2)`, unitOneID, unitTwoID); err != nil {
		t.Fatalf("check that electricity register did not create payments: %v", err)
	}
	if generatedPayments != 0 {
		t.Fatalf("electricity register unexpectedly generated %d payments", generatedPayments)
	}

	incompleteRun, err := repo.Create(buildingID, orgID, 12, 2026, userID)
	if err != nil {
		t.Fatalf("create incomplete register for validation: %v", err)
	}
	if _, err := repo.Finalize(incompleteRun.ID, orgID); err == nil {
		t.Fatal("expected an incomplete register to reject finalization")
	}
}
