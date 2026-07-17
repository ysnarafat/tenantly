package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID        int             `json:"id" db:"id"`
	UserID    *int            `json:"user_id" db:"user_id"`
	Action    string          `json:"action" db:"action"`
	TableName string          `json:"table_name" db:"table_name"`
	RecordID  *int            `json:"record_id" db:"record_id"`
	OldValues json.RawMessage `json:"old_values" db:"old_values"`
	NewValues json.RawMessage `json:"new_values" db:"new_values"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type CreateAuditLogRequest struct {
	UserID    *int   `json:"user_id"`
	Action    string `json:"action" binding:"required"`
	TableName string `json:"table_name" binding:"required"`
	RecordID  *int   `json:"record_id"`
	OldValues any    `json:"old_values"`
	NewValues any    `json:"new_values"`
}

// AuditAction constants for consistent audit logging
const (
	AuditActionCreate = "CREATE"
	AuditActionUpdate = "UPDATE"
	AuditActionDelete = "DELETE"
	AuditActionLogin  = "LOGIN"
	AuditActionLogout = "LOGOUT"
)

// TableName constants for audit logging
const (
	TableUsers         = "users"
	TableTenants       = "tenants"
	TableLeases        = "leases"
	TablePayments      = "payments"
	TableNotifications = "notification_queue"
)
