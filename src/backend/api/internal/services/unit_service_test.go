package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/ysnarafat/tenantly/internal/models"
)

// TestUnitRepositoryForBulk implements interfaces.UnitRepositoryInterface with
// configurable behavior for BulkCreateUnits tests (MockPaymentUnitRepo in
// payment_service_test.go hardcodes CheckUnitNumberExists/BulkCreate, which
// doesn't support the exists/fail scenarios these tests need).
type TestUnitRepositoryForBulk struct {
	existingUnitNumbers  map[string]bool
	shouldFailBulkCreate bool
	created              []*models.Unit
}

func newTestUnitRepositoryForBulk() *TestUnitRepositoryForBulk {
	return &TestUnitRepositoryForBulk{existingUnitNumbers: make(map[string]bool)}
}

func (m *TestUnitRepositoryForBulk) addExistingUnitNumber(buildingID int, unitNumber string) {
	m.existingUnitNumbers[fmt.Sprintf("%d:%s", buildingID, unitNumber)] = true
}

func (m *TestUnitRepositoryForBulk) CheckUnitNumberExists(buildingID int, unitNumber string, excludeID int) (bool, error) {
	return m.existingUnitNumbers[fmt.Sprintf("%d:%s", buildingID, unitNumber)], nil
}

func (m *TestUnitRepositoryForBulk) BulkCreate(units []*models.Unit) error {
	if m.shouldFailBulkCreate {
		return fmt.Errorf("bulk insert failed")
	}
	for i, u := range units {
		u.ID = i + 1
	}
	m.created = append(m.created, units...)
	return nil
}

func (m *TestUnitRepositoryForBulk) Create(req *models.CreateUnitRequest, organizationID int) (*models.Unit, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) GetByID(id int) (*models.Unit, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *TestUnitRepositoryForBulk) GetByIDWithDetails(id int) (*models.UnitWithDetails, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) Update(id int, req *models.UpdateUnitRequest) (*models.Unit, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) Delete(id int) error                      { return nil }
func (m *TestUnitRepositoryForBulk) HasActiveLeases(unitID int) (bool, error) { return false, nil }
func (m *TestUnitRepositoryForBulk) GetByBuildingWithDetails(buildingID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	return nil, 0, nil
}
func (m *TestUnitRepositoryForBulk) GetByPropertyWithDetails(propertyID int, limit, offset, orgID int) ([]*models.UnitWithDetails, int, error) {
	return nil, 0, nil
}
func (m *TestUnitRepositoryForBulk) GetBuildingOccupancyStats(buildingID int, startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) GetPropertyOccupancyStats(propertyID int, startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) GetSystemOccupancyStats(startDate, endDate time.Time) (interface{}, error) {
	return nil, nil
}
func (m *TestUnitRepositoryForBulk) GetBuildingUnitTypeDistribution(buildingID int) (interface{}, error) {
	return nil, nil
}

func newTestBuilding(id, propertyID, orgID int, buildingType models.BuildingType, active bool) *models.Building {
	return &models.Building{
		ID:             id,
		PropertyID:     propertyID,
		OrganizationID: orgID,
		BuildingName:   "Test Building",
		BuildingCode:   "BLD-TEST",
		BuildingType:   buildingType,
		ActiveStatus:   active,
	}
}

func newTestPropertyForUnits(id, orgID int, active bool) *models.Property {
	return &models.Property{
		ID:             id,
		OrganizationID: orgID,
		PropertyName:   "Test Property",
		PropertyCode:   "PROP-TEST",
		Active:         active,
	}
}

func newBulkUnitsRequest(buildingID int, items ...models.BulkCreateUnitItem) *models.BulkCreateUnitsRequest {
	return &models.BulkCreateUnitsRequest{BuildingID: buildingID, Units: items}
}

func TestUnitService_BulkCreateUnits_Success(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	auditService := newMockPaymentAuditService()

	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))

	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, auditService)

	req := newBulkUnitsRequest(1,
		models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment},
		models.BulkCreateUnitItem{UnitNumber: "102", UnitType: models.UnitTypeApartment},
		models.BulkCreateUnitItem{UnitNumber: "103", UnitType: models.UnitTypeApartment},
	)

	units, err := service.BulkCreateUnits(req, 5, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(units) != 3 {
		t.Fatalf("expected 3 units, got %d", len(units))
	}
	for _, u := range units {
		if u.ID == 0 {
			t.Errorf("expected unit to have an assigned ID, got 0")
		}
		if u.PropertyID != 10 {
			t.Errorf("expected property_id derived from building (10), got %d", u.PropertyID)
		}
		if u.OrganizationID != 100 {
			t.Errorf("expected organization_id from building (100), got %d", u.OrganizationID)
		}
	}
	if auditService.userActionCalls != 3 {
		t.Errorf("expected 3 audit log calls, got %d", auditService.userActionCalls)
	}
}

func TestUnitService_BulkCreateUnits_BuildingNotFound(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(999, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment})

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for nonexistent building, got nil")
	}
}

func TestUnitService_BulkCreateUnits_WrongOrganization(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment})

	// Building belongs to org 100, but the caller is org 999 — must be rejected (IDOR).
	if _, err := service.BulkCreateUnits(req, 5, 999); err == nil {
		t.Fatal("expected error for cross-organization access, got nil")
	}
}

func TestUnitService_BulkCreateUnits_InactiveBuilding(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, false))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment})

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for inactive building, got nil")
	}
}

func TestUnitService_BulkCreateUnits_InactiveProperty(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, false))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment})

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for inactive property, got nil")
	}
}

func TestUnitService_BulkCreateUnits_DuplicateUnitNumberInRequest(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1,
		models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment},
		models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment},
	)

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for duplicate unit number within the request, got nil")
	}
	if len(unitRepo.created) != 0 {
		t.Errorf("expected nothing persisted on validation failure, got %d", len(unitRepo.created))
	}
}

func TestUnitService_BulkCreateUnits_DuplicateUnitNumberInDatabase(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	unitRepo.addExistingUnitNumber(1, "101")
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment})

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for unit number already existing in the building, got nil")
	}
	if len(unitRepo.created) != 0 {
		t.Errorf("expected nothing persisted on validation failure, got %d", len(unitRepo.created))
	}
}

func TestUnitService_BulkCreateUnits_UnitTypeNotAllowedForBuildingType(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	// Residential building — Shop is not an allowed unit type here.
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, newMockPaymentAuditService())

	req := newBulkUnitsRequest(1, models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeShop})

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error for disallowed unit type, got nil")
	}
	if len(unitRepo.created) != 0 {
		t.Errorf("expected nothing persisted on validation failure, got %d", len(unitRepo.created))
	}
}

func TestUnitService_BulkCreateUnits_RepositoryFailure(t *testing.T) {
	unitRepo := newTestUnitRepositoryForBulk()
	unitRepo.shouldFailBulkCreate = true
	buildingRepo := newMockPaymentBuildingRepo()
	propertyRepo := newMockPaymentPropertyRepo()
	buildingRepo.addBuilding(newTestBuilding(1, 10, 100, models.BuildingTypeResidential, true))
	propertyRepo.addProperty(newTestPropertyForUnits(10, 100, true))
	auditService := newMockPaymentAuditService()
	service := NewUnitService(unitRepo, buildingRepo, propertyRepo, auditService)

	req := newBulkUnitsRequest(1,
		models.BulkCreateUnitItem{UnitNumber: "101", UnitType: models.UnitTypeApartment},
		models.BulkCreateUnitItem{UnitNumber: "102", UnitType: models.UnitTypeApartment},
	)

	if _, err := service.BulkCreateUnits(req, 5, 100); err == nil {
		t.Fatal("expected error when the repository fails, got nil")
	}
	if len(unitRepo.created) != 0 {
		t.Errorf("expected nothing persisted when the repository call fails, got %d", len(unitRepo.created))
	}
	if auditService.userActionCalls != 0 {
		t.Errorf("expected no audit logging when nothing was created, got %d calls", auditService.userActionCalls)
	}
}
