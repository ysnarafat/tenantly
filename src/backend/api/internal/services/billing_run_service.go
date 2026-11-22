package services

import (
	"fmt"
	"time"

	"github.com/ysnarafat/tenantly/internal/interfaces"
	"github.com/ysnarafat/tenantly/internal/models"
)

type BillingRunService struct {
	repo         interfaces.BillingRunRepositoryInterface
	auditService interfaces.AuditServiceInterface
}

func NewBillingRunService(repo interfaces.BillingRunRepositoryInterface, auditService interfaces.AuditServiceInterface) *BillingRunService {
	return &BillingRunService{repo: repo, auditService: auditService}
}

func (s *BillingRunService) Create(req *models.CreateBillingRunRequest, orgID, userID int) (*models.BillingRun, error) {
	if req.Month < 1 || req.Month > 12 || req.Year < 2020 {
		return nil, fmt.Errorf("invalid billing period")
	}
	now := time.Now().UTC()
	requested := req.Year*12 + req.Month
	current := now.Year()*12 + int(now.Month())
	if requested > current+1 || requested < current-24 {
		return nil, fmt.Errorf("billing period must be within the last 24 months and one month ahead")
	}
	run, err := s.repo.Create(req.BuildingID, orgID, req.Month, req.Year, userID)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "CREATE", "electricity_bill_runs", &run.ID, nil, map[string]interface{}{
		"building_id": run.BuildingID, "month": run.Month, "year": run.Year, "items": len(run.Items),
	})
	return run, nil
}

func (s *BillingRunService) Get(id, orgID int) (*models.BillingRun, error) {
	return s.repo.Get(id, orgID)
}

func (s *BillingRunService) GetByPeriod(buildingID, orgID, month, year int) (*models.BillingRun, error) {
	return s.repo.GetByPeriod(buildingID, orgID, month, year)
}

func (s *BillingRunService) UpdateElectricity(runID, itemID, orgID, userID int, meterNumber string, amount float64) (*models.BillingRun, error) {
	if amount < 0 {
		return nil, fmt.Errorf("electricity amount cannot be negative")
	}
	run, err := s.repo.UpdateElectricity(runID, itemID, orgID, meterNumber, amount)
	if err == nil {
		_ = s.auditService.LogUserAction(userID, "UPDATE_ELECTRICITY", "electricity_bill_runs", &runID, nil, map[string]interface{}{
			"item_id": itemID, "meter_number": meterNumber, "electricity_amount": amount,
		})
	}
	return run, err
}

func (s *BillingRunService) Finalize(id, orgID, userID int) (*models.BillingRun, error) {
	run, err := s.repo.Finalize(id, orgID)
	if err != nil {
		return nil, err
	}
	_ = s.auditService.LogUserAction(userID, "FINALIZE", "electricity_bill_runs", &id, nil, map[string]interface{}{
		"building_id": run.BuildingID, "month": run.Month, "year": run.Year, "items": len(run.Items),
	})
	return run, nil
}
