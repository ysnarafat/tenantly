package models

import "time"

type NotificationQueue struct {
	ID               int       `json:"id" db:"id"`
	TenantID         int       `json:"tenant_id" db:"tenant_id"`
	ShopID           int       `json:"shop_id" db:"shop_id"`
	Message          string    `json:"message" db:"message"`
	NotificationType string    `json:"notification_type" db:"notification_type"`
	Recipient        string    `json:"recipient" db:"recipient"`
	Status           string    `json:"status" db:"status"`
	RetryCount       int       `json:"retry_count" db:"retry_count"`
	ErrorMessage     string    `json:"error_message" db:"error_message"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type CreateNotificationRequest struct {
	TenantID         int    `json:"tenant_id" binding:"required"`
	ShopID           int    `json:"shop_id" binding:"required"`
	Message          string `json:"message" binding:"required"`
	NotificationType string `json:"notification_type" binding:"required,oneof=SMS Email Reminder"`
	Recipient        string `json:"recipient" binding:"required"`
}

type UpdateNotificationRequest struct {
	Status       string `json:"status" binding:"omitempty,oneof=Pending Sent Failed"`
	RetryCount   *int   `json:"retry_count" binding:"omitempty,gte=0"`
	ErrorMessage string `json:"error_message" binding:"omitempty"`
}

type NotificationWithDetails struct {
	NotificationQueue
	TenantName string `json:"tenant_name" db:"tenant_name"`
	ShopName   string `json:"shop_name" db:"shop_name"`
}

// Notification type constants
const (
	NotificationTypeSMS      = "SMS"
	NotificationTypeEmail    = "Email"
	NotificationTypeReminder = "Reminder"
)

// Notification status constants
const (
	NotificationStatusPending = "Pending"
	NotificationStatusSent    = "Sent"
	NotificationStatusFailed  = "Failed"
)
