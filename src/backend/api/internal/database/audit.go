package database

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

type AuditService struct {
	db *sqlx.DB
}

func NewAuditService(db *sqlx.DB) *AuditService {
	return &AuditService{db: db}
}

// LogUserAction logs user-specific actions like login, logout, etc.
func (a *AuditService) LogUserAction(userID int, action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	return a.logAudit(&userID, action, tableName, recordID, oldValues, newValues)
}

// LogSystemAction logs system actions without a specific user
func (a *AuditService) LogSystemAction(action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	return a.logAudit(nil, action, tableName, recordID, oldValues, newValues)
}

func (a *AuditService) logAudit(userID *int, action, tableName string, recordID *int, oldValues, newValues interface{}) error {
	var oldJSON, newJSON []byte
	var err error

	if oldValues != nil {
		oldJSON, err = json.Marshal(oldValues)
		if err != nil {
			return fmt.Errorf("failed to marshal old values: %w", err)
		}
	}

	if newValues != nil {
		newJSON, err = json.Marshal(newValues)
		if err != nil {
			return fmt.Errorf("failed to marshal new values: %w", err)
		}
	}

	query := `
		INSERT INTO audit_log (user_id, action, table_name, record_id, old_values, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err = a.db.Exec(query, userID, action, tableName, recordID, oldJSON, newJSON)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (a *AuditService) GetAuditLogs(userID *int, tableName, action string, limit, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, action, table_name, record_id, old_values, new_values, created_at
		FROM audit_log
		WHERE ($1::INTEGER IS NULL OR user_id = $1)
		AND ($2::TEXT IS NULL OR table_name = $2)
		AND ($3::TEXT IS NULL OR action = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := a.db.Query(query, userID, tableName, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.TableName,
			&log.RecordID,
			&log.OldValues,
			&log.NewValues,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogsByRecord retrieves audit logs for a specific record
func (a *AuditService) GetAuditLogsByRecord(tableName string, recordID int) ([]models.AuditLog, error) {
	query := `
		SELECT id, user_id, action, table_name, record_id, old_values, new_values, created_at
		FROM audit_log
		WHERE table_name = $1 AND record_id = $2
		ORDER BY created_at DESC
	`

	rows, err := a.db.Query(query, tableName, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by record: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.TableName,
			&log.RecordID,
			&log.OldValues,
			&log.NewValues,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log row: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit log rows: %w", err)
	}

	return logs, nil
}
