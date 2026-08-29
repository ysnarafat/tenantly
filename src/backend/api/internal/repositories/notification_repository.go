package repositories

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ysnarafat/tenantly/internal/models"
)

// NotificationRepository queues rows for the .NET notification-service to
// pick up and deliver (SMS/Email) — this Go API only ever writes Pending
// rows; delivery, retries, and status transitions are owned entirely by the
// notification-service.
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create queues a new Pending notification.
func (r *NotificationRepository) Create(req *models.CreateNotificationRequest) (*models.NotificationQueue, error) {
	var propertyID, buildingID sql.NullInt64
	if req.PropertyID != nil {
		propertyID = sql.NullInt64{Int64: int64(*req.PropertyID), Valid: true}
	}
	if req.BuildingID != nil {
		buildingID = sql.NullInt64{Int64: int64(*req.BuildingID), Valid: true}
	}

	query := `
		INSERT INTO notification_queue (tenant_id, unit_id, property_id, building_id, message, notification_type, recipient, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'Pending')
		RETURNING id, tenant_id, unit_id, property_id, building_id, message, notification_type, recipient, status,
			retry_count, COALESCE(error_message, ''), sent_at, created_at, updated_at`

	n := &models.NotificationQueue{}
	var propID, bldgID sql.NullInt64
	err := r.db.QueryRow(query,
		req.TenantID, req.UnitID, propertyID, buildingID, req.Message, req.NotificationType, req.Recipient,
	).Scan(
		&n.ID, &n.TenantID, &n.UnitID, &propID, &bldgID, &n.Message, &n.NotificationType, &n.Recipient, &n.Status,
		&n.RetryCount, &n.ErrorMessage, &n.SentAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to queue notification: %w", err)
	}
	if propID.Valid {
		id := int(propID.Int64)
		n.PropertyID = &id
	}
	if bldgID.Valid {
		id := int(bldgID.Int64)
		n.BuildingID = &id
	}
	return n, nil
}
