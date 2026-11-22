package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

type BillingRunRepository struct {
	db *sqlx.DB
}

func NewBillingRunRepository(db *sqlx.DB) *BillingRunRepository {
	return &BillingRunRepository{db: db}
}

const billingRunSelect = `
	SELECT br.id, br.building_id, br.organization_id, p.property_name,
		b.building_name, b.building_code, br.month, br.year, br.status,
		br.created_by, br.finalized_at, br.created_at
	FROM electricity_bill_runs br
	JOIN buildings b ON b.id = br.building_id
	JOIN properties p ON p.id = b.property_id`

func (r *BillingRunRepository) Create(buildingID, orgID, month, year, userID int) (*models.BillingRun, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin electricity register: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var id int
	err = tx.QueryRow(`
		INSERT INTO electricity_bill_runs (building_id, organization_id, month, year, created_by)
		SELECT b.id, p.organization_id, $3, $4, $5
		FROM buildings b
		JOIN properties p ON p.id = b.property_id
		WHERE b.id = $1 AND p.organization_id = $2 AND p.active = true AND b.active_status = true
		RETURNING id`, buildingID, orgID, month, year, userID).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("building not found")
	}
	if err != nil {
		return nil, fmt.Errorf("create electricity register: %w", err)
	}

	period := year*12 + month
	_, err = tx.Exec(`
		INSERT INTO electricity_bill_run_items
			(billing_run_id, unit_id, unit_number, unit_type, floor, meter_number)
		SELECT $1, u.id, u.unit_number, u.unit_type, COALESCE(u.floor, 0),
			COALESCE((
				SELECT previous_item.meter_number
				FROM electricity_bill_run_items previous_item
				JOIN electricity_bill_runs previous_run ON previous_run.id = previous_item.billing_run_id
				WHERE previous_item.unit_id = u.id
					AND previous_run.building_id = $2
					AND previous_run.year * 12 + previous_run.month < $3
				ORDER BY previous_run.year DESC, previous_run.month DESC
				LIMIT 1
			), '')
		FROM units u
		WHERE u.building_id = $2 AND u.active = true
		ORDER BY COALESCE(u.floor, 0), u.unit_number`,
		id, buildingID, period)
	if err != nil {
		return nil, fmt.Errorf("snapshot active building units: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit electricity register: %w", err)
	}
	return r.Get(id, orgID)
}

func (r *BillingRunRepository) Get(id, orgID int) (*models.BillingRun, error) {
	run := &models.BillingRun{}
	err := r.db.QueryRowx(billingRunSelect+` WHERE br.id = $1 AND br.organization_id = $2`, id, orgID).StructScan(run)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("billing run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get electricity register: %w", err)
	}
	run.Items, err = r.getItems(id)
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (r *BillingRunRepository) GetByPeriod(buildingID, orgID, month, year int) (*models.BillingRun, error) {
	run := &models.BillingRun{}
	err := r.db.QueryRowx(billingRunSelect+`
		WHERE br.building_id = $1 AND br.organization_id = $2 AND br.month = $3 AND br.year = $4`,
		buildingID, orgID, month, year).StructScan(run)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get electricity register by period: %w", err)
	}
	run.Items, err = r.getItems(run.ID)
	return run, err
}

func (r *BillingRunRepository) getItems(runID int) ([]*models.BillingRunItem, error) {
	items := make([]*models.BillingRunItem, 0)
	err := r.db.Select(&items, `
		SELECT id, unit_id, unit_number, unit_type, floor, meter_number, electricity_amount
		FROM electricity_bill_run_items
		WHERE billing_run_id = $1
		ORDER BY floor, unit_number`, runID)
	if err != nil {
		return nil, fmt.Errorf("get electricity register items: %w", err)
	}
	return items, nil
}

func (r *BillingRunRepository) UpdateElectricity(runID, itemID, orgID int, meterNumber string, amount float64) (*models.BillingRun, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin electricity item update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRow(`
		SELECT status FROM electricity_bill_runs WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		runID, orgID).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("billing run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("load electricity register for update: %w", err)
	}
	if status != string(models.BillingRunDraft) {
		return nil, fmt.Errorf("billing run is already finalized")
	}

	result, err := tx.Exec(`
		UPDATE electricity_bill_run_items
		SET meter_number = $1, electricity_amount = $2
		WHERE billing_run_id = $3 AND id = $4`,
		meterNumber, amount, runID, itemID)
	if err != nil {
		return nil, fmt.Errorf("update electricity bill item: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("check electricity bill item update: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("electricity bill item not found")
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit electricity bill item update: %w", err)
	}
	return r.Get(runID, orgID)
}

func (r *BillingRunRepository) Finalize(id, orgID int) (*models.BillingRun, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin electricity register finalization: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRow(`
		SELECT status FROM electricity_bill_runs WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		id, orgID).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("billing run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("load electricity register for finalization: %w", err)
	}
	if status != string(models.BillingRunDraft) {
		return nil, fmt.Errorf("billing run is already finalized")
	}

	var itemCount, missingAmountCount int
	err = tx.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (WHERE electricity_amount IS NULL)
		FROM electricity_bill_run_items WHERE billing_run_id = $1`, id).Scan(&itemCount, &missingAmountCount)
	if err != nil {
		return nil, fmt.Errorf("validate electricity register items: %w", err)
	}
	if itemCount == 0 {
		return nil, fmt.Errorf("cannot finalize a register with no active units")
	}
	if missingAmountCount > 0 {
		return nil, fmt.Errorf("enter an electricity amount for every unit before finalizing")
	}

	if _, err := tx.Exec(`
		UPDATE electricity_bill_runs SET status = 'Finalized', finalized_at = $1 WHERE id = $2`,
		time.Now().UTC(), id); err != nil {
		return nil, fmt.Errorf("finalize electricity register: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit electricity register finalization: %w", err)
	}
	return r.Get(id, orgID)
}
