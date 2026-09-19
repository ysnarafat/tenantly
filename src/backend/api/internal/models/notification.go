package models

import "time"

// NotificationQueue is a row in the notification_queue table, consumed by
// the .NET notification-service (SMS/Email delivery worker). Column shapes
// here must match migrations/000001_initial_schema.up.sql exactly.
type NotificationQueue struct {
	ID               int        `json:"id" db:"id"`
	TenantID         int        `json:"tenant_id" db:"tenant_id"`
	UnitID           int        `json:"unit_id" db:"unit_id"`
	PropertyID       *int       `json:"property_id,omitempty" db:"property_id"`
	BuildingID       *int       `json:"building_id,omitempty" db:"building_id"`
	Message          string     `json:"message" db:"message"`
	NotificationType string     `json:"notification_type" db:"notification_type"`
	Recipient        string     `json:"recipient" db:"recipient"`
	Status           string     `json:"status" db:"status"`
	RetryCount       int        `json:"retry_count" db:"retry_count"`
	ErrorMessage     string     `json:"error_message" db:"error_message"`
	SentAt           *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateNotificationRequest queues a new notification for the .NET
// notification-service to pick up and deliver.
type CreateNotificationRequest struct {
	TenantID         int    `json:"tenant_id" binding:"required"`
	UnitID           int    `json:"unit_id" binding:"required"`
	PropertyID       *int   `json:"property_id" binding:"omitempty"`
	BuildingID       *int   `json:"building_id" binding:"omitempty"`
	Message          string `json:"message" binding:"required"`
	NotificationType string `json:"notification_type" binding:"required,oneof=SMS Email Reminder Acknowledgment"`
	Recipient        string `json:"recipient" binding:"required"`
}

// Notification type constants
const (
	NotificationTypeSMS            = "SMS"
	NotificationTypeEmail          = "Email"
	NotificationTypeReminder       = "Reminder"
	NotificationTypeAcknowledgment = "Acknowledgment"
)

// Notification status constants
const (
	NotificationStatusPending = "Pending"
	NotificationStatusSent    = "Sent"
	NotificationStatusFailed  = "Failed"
)
